package service

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

const (
	defaultOpsNginxTimingLogPath       = "/app/data/nginx/sub2api_timing.json.log"
	defaultOpsNginxTimingLegacyLogPath = "/app/data/nginx/sub2api_timing.legacy.log"
	maxOpsNginxTimingWindow            = 24 * time.Hour
	maxOpsNginxTimingLogBytes          = int64(64 << 20)
	maxOpsNginxTimingLineBytes         = 1 << 20
)

var nginxLegacyTimingLogPattern = regexp.MustCompile(`^\S+ \[([^\]]+)\] "([^"]*)" status=(\d+) rt=([^ ]+) uct=([^ ]+) uht=([^ ]+) urt=([^ ]+) us=([^ ]+) req_len=(\d+) .* crid="([^"]*)"$`)

type nginxTimingLogEntry struct {
	Timestamp            string `json:"timestamp"`
	Path                 string `json:"path"`
	Gateway              string `json:"gateway"`
	Upgrade              string `json:"upgrade"`
	Status               int    `json:"status"`
	RequestTime          string `json:"request_time"`
	UpstreamConnectTime  string `json:"upstream_connect_time"`
	UpstreamHeaderTime   string `json:"upstream_header_time"`
	UpstreamResponseTime string `json:"upstream_response_time"`
	RequestLength        int64  `json:"request_length"`
	ClientRequestID      string `json:"client_request_id"`
}

type nginxTimingAccumulator struct {
	requestTime      []int
	upstreamConnect  []int
	upstreamHeader   []int
	upstreamResponse []int
	clientOverhead   []int
	requestBytes     int64
	requestByteCount int64
	maxRequestBytes  int64
}

// GetNginxTimingOverview reads the JSON access log Nginx writes into the
// existing application data volume. It never writes raw Nginx data to Postgres.
func (s *OpsService) GetNginxTimingOverview(ctx context.Context, filter *OpsNginxTimingFilter) (*OpsNginxTimingOverview, error) {
	if err := s.RequireMonitoringEnabled(ctx); err != nil {
		return nil, err
	}
	if filter == nil || filter.StartTime.IsZero() || filter.EndTime.IsZero() {
		return nil, infraerrors.BadRequest("OPS_NGINX_TIME_RANGE_REQUIRED", "start_time/end_time are required")
	}
	if filter.StartTime.After(filter.EndTime) {
		return nil, infraerrors.BadRequest("OPS_NGINX_TIME_RANGE_INVALID", "start_time must be <= end_time")
	}

	effective := *filter
	effective.StartTime = effective.StartTime.UTC()
	effective.EndTime = effective.EndTime.UTC()
	windowClamped := false
	if effective.EndTime.Sub(effective.StartTime) > maxOpsNginxTimingWindow {
		effective.StartTime = effective.EndTime.Add(-maxOpsNginxTimingWindow)
		windowClamped = true
	}

	var clientRequestIDs map[string]struct{}
	if len(effective.APIKeyIDs) > 0 {
		resolver, ok := s.opsRepo.(OpsNginxTimingKeyResolver)
		if !ok {
			return nil, infraerrors.ServiceUnavailable("OPS_NGINX_KEY_CORRELATION_UNAVAILABLE", "Nginx Key correlation is not available")
		}
		var err error
		clientRequestIDs, err = resolver.ListNginxTimingClientRequestIDs(ctx, &effective)
		if err != nil {
			return nil, err
		}
	}

	overview, err := readOpsNginxTimingLog(opsNginxTimingLogPath(), &effective, clientRequestIDs)
	if err != nil {
		return nil, infraerrors.InternalServer("OPS_NGINX_TIMING_READ_FAILED", "Failed to read Nginx timing log").WithCause(err)
	}
	overview.WindowClamped = windowClamped
	return overview, nil
}

func opsNginxTimingLogPath() string {
	if path := strings.TrimSpace(os.Getenv("OPS_NGINX_TIMING_LOG_PATH")); path != "" {
		return path
	}
	return defaultOpsNginxTimingLogPath
}

func opsNginxTimingLegacyLogPath() string {
	if path := strings.TrimSpace(os.Getenv("OPS_NGINX_TIMING_LEGACY_LOG_PATH")); path != "" {
		return path
	}
	return defaultOpsNginxTimingLegacyLogPath
}

func readOpsNginxTimingLog(path string, filter *OpsNginxTimingFilter, clientRequestIDs map[string]struct{}) (*OpsNginxTimingOverview, error) {
	if filter == nil {
		return nil, infraerrors.BadRequest("OPS_NGINX_FILTER_REQUIRED", "filter is required")
	}

	overview := &OpsNginxTimingOverview{
		StartTime:        filter.StartTime.UTC(),
		EndTime:          filter.EndTime.UTC(),
		KeyFilterApplied: len(filter.APIKeyIDs) > 0,
	}
	acc := nginxTimingAccumulator{}

	for _, candidate := range []string{
		opsNginxTimingLegacyLogPath() + ".1",
		opsNginxTimingLegacyLogPath(),
		path + ".1",
		path,
	} {
		found, updatedAt, err := scanOpsNginxTimingLog(candidate, func(entry nginxTimingLogEntry, at time.Time) {
			if at.Before(filter.StartTime) || at.After(filter.EndTime) || !entry.isGatewayRequest() {
				return
			}

			if overview.KeyFilterApplied {
				clientRequestID := strings.TrimSpace(entry.ClientRequestID)
				if clientRequestID == "" {
					if entry.Status >= 400 {
						overview.UnattributedErrorCount++
					}
					return
				}
				if _, ok := clientRequestIDs[clientRequestID]; !ok {
					return
				}
				overview.MatchedRequestCount++
			}

			if entry.isWebSocket() {
				overview.WebSocketSessionCount++
				return
			}

			overview.HTTPRequestCount++
			if entry.Status >= 200 && entry.Status < 400 {
				overview.SuccessCount++
			}
			switch {
			case entry.Status == 408:
				overview.ClientTimeout408Count++
			case entry.Status == 499:
				overview.ClientClosed499Count++
			case entry.Status >= 500 && entry.Status <= 599:
				overview.ServerError5xxCount++
			}

			requestTimeMS, hasRequestTime := nginxTimingDurationMS(entry.RequestTime)
			upstreamConnectMS, hasUpstreamConnect := nginxTimingDurationMS(entry.UpstreamConnectTime)
			upstreamHeaderMS, hasUpstreamHeader := nginxTimingDurationMS(entry.UpstreamHeaderTime)
			upstreamResponseMS, hasUpstreamResponse := nginxTimingDurationMS(entry.UpstreamResponseTime)
			if !hasUpstreamConnect && !hasUpstreamResponse {
				overview.UpstreamUnreachedCount++
			}
			if hasRequestTime {
				acc.requestTime = append(acc.requestTime, requestTimeMS)
			}
			if hasUpstreamConnect {
				acc.upstreamConnect = append(acc.upstreamConnect, upstreamConnectMS)
			}
			if hasUpstreamHeader {
				acc.upstreamHeader = append(acc.upstreamHeader, upstreamHeaderMS)
			}
			if hasUpstreamResponse {
				acc.upstreamResponse = append(acc.upstreamResponse, upstreamResponseMS)
			}
			if hasRequestTime && hasUpstreamResponse {
				acc.clientOverhead = append(acc.clientOverhead, max(requestTimeMS-upstreamResponseMS, 0))
			}
			if entry.RequestLength > 0 {
				acc.requestBytes += entry.RequestLength
				acc.requestByteCount++
				if entry.RequestLength > acc.maxRequestBytes {
					acc.maxRequestBytes = entry.RequestLength
				}
			}
		})
		if err != nil {
			return nil, err
		}
		if found {
			overview.Available = true
			if overview.SourceUpdatedAt == nil || (updatedAt != nil && updatedAt.After(*overview.SourceUpdatedAt)) {
				overview.SourceUpdatedAt = updatedAt
			}
		}
	}

	if !overview.Available {
		return overview, nil
	}
	overview.RequestTime = opsNginxPercentiles(acc.requestTime)
	overview.UpstreamConnect = opsNginxPercentiles(acc.upstreamConnect)
	overview.UpstreamHeader = opsNginxPercentiles(acc.upstreamHeader)
	overview.UpstreamResponse = opsNginxPercentiles(acc.upstreamResponse)
	overview.ClientOverhead = opsNginxPercentiles(acc.clientOverhead)
	if acc.requestByteCount > 0 {
		avg := acc.requestBytes / acc.requestByteCount
		maxBytes := acc.maxRequestBytes
		overview.AvgRequestBytes = &avg
		overview.MaxRequestBytes = &maxBytes
	}
	return overview, nil
}

func scanOpsNginxTimingLog(path string, visit func(nginxTimingLogEntry, time.Time)) (bool, *time.Time, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return false, nil, err
	}
	updatedAt := info.ModTime().UTC()
	var lineReader io.Reader = file
	if info.Size() > maxOpsNginxTimingLogBytes {
		if _, err := file.Seek(-maxOpsNginxTimingLogBytes, io.SeekEnd); err != nil {
			return false, nil, err
		}
		// The first scanned line starts in the middle of an access-log entry.
		reader := bufio.NewReader(file)
		if _, err := reader.ReadString('\n'); err != nil && err != io.EOF {
			return false, nil, err
		}
		lineReader = reader
	}

	scanner := bufio.NewScanner(lineReader)
	scanner.Buffer(make([]byte, 64*1024), maxOpsNginxTimingLineBytes)
	for scanner.Scan() {
		var entry nginxTimingLogEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err == nil {
			at, err := time.Parse(time.RFC3339Nano, entry.Timestamp)
			if err == nil {
				visit(entry, at.UTC())
			}
			continue
		}
		entry, at, ok := parseOpsNginxLegacyTimingLogLine(string(scanner.Bytes()))
		if ok {
			visit(entry, at)
		}
	}
	if err := scanner.Err(); err != nil {
		return false, nil, err
	}
	return true, &updatedAt, nil
}

func parseOpsNginxLegacyTimingLogLine(line string) (nginxTimingLogEntry, time.Time, bool) {
	matches := nginxLegacyTimingLogPattern.FindStringSubmatch(line)
	if len(matches) == 0 {
		return nginxTimingLogEntry{}, time.Time{}, false
	}
	at, err := time.Parse("02/Jan/2006:15:04:05 -0700", matches[1])
	if err != nil {
		return nginxTimingLogEntry{}, time.Time{}, false
	}
	requestParts := strings.Fields(matches[2])
	if len(requestParts) < 2 {
		return nginxTimingLogEntry{}, time.Time{}, false
	}
	status, err := strconv.Atoi(matches[3])
	if err != nil {
		return nginxTimingLogEntry{}, time.Time{}, false
	}
	requestLength, _ := strconv.ParseInt(matches[9], 10, 64)
	entry := nginxTimingLogEntry{
		Timestamp:            at.UTC().Format(time.RFC3339Nano),
		Path:                 strings.SplitN(requestParts[1], "?", 2)[0],
		Status:               status,
		RequestTime:          matches[4],
		UpstreamConnectTime:  matches[5],
		UpstreamHeaderTime:   matches[6],
		UpstreamResponseTime: matches[7],
		RequestLength:        requestLength,
		ClientRequestID:      matches[10],
	}
	if status == 101 {
		entry.Upgrade = "websocket"
	}
	return entry, at.UTC(), true
}

func (entry nginxTimingLogEntry) isGatewayRequest() bool {
	if gateway := strings.TrimSpace(entry.Gateway); gateway != "" {
		return gateway == "1"
	}
	path := strings.TrimSpace(entry.Path)
	return path == "/responses" || path == "/chat/completions" || path == "/messages" ||
		strings.HasPrefix(path, "/v1/") || strings.HasPrefix(path, "/images/") ||
		strings.HasPrefix(path, "/audio/") || strings.HasPrefix(path, "/videos/")
}

func (entry nginxTimingLogEntry) isWebSocket() bool {
	return strings.EqualFold(strings.TrimSpace(entry.Upgrade), "websocket")
}

func nginxTimingDurationMS(raw string) (int, bool) {
	var total int
	found := false
	for _, value := range strings.Split(raw, ",") {
		value = strings.TrimSpace(value)
		if value == "" || value == "-" {
			continue
		}
		seconds, err := strconv.ParseFloat(value, 64)
		if err != nil || seconds < 0 {
			continue
		}
		total += int(math.Round(seconds * 1000))
		found = true
	}
	return total, found
}

func opsNginxPercentiles(values []int) OpsPercentiles {
	if len(values) == 0 {
		return OpsPercentiles{}
	}
	sort.Ints(values)
	var sum int64
	for _, value := range values {
		sum += int64(value)
	}

	pick := func(percent float64) *int {
		index := int(math.Ceil(percent*float64(len(values)))) - 1
		index = min(max(index, 0), len(values)-1)
		value := values[index]
		return &value
	}
	avg := int(math.Round(float64(sum) / float64(len(values))))
	maxValue := values[len(values)-1]
	return OpsPercentiles{
		P50: pick(0.50),
		P90: pick(0.90),
		P95: pick(0.95),
		P99: pick(0.99),
		Avg: &avg,
		Max: &maxValue,
	}
}
