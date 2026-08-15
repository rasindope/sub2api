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
	CompletedAtMS        string `json:"completed_at_ms"`
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
	GatewayReceivedAtMS  string `json:"gateway_received_at_ms"`
}

type nginxTimingAccumulator struct {
	requestTime           []int
	upstreamConnect       []int
	upstreamHeader        []int
	upstreamResponse      []int
	clientOverhead        []int
	clientUpload          []int
	clientResponseReceive []int
	requestBytes          int64
	requestByteCount      int64
	maxRequestBytes       int64
}

type nginxTimingKeyAccumulator struct {
	key      OpsNginxTimingRequestKey
	overview OpsNginxTimingOverview
	acc      nginxTimingAccumulator
}

// GetNginxTimingOverview reads the JSON access log Nginx writes into the
// existing application data volume. It never writes raw Nginx data to Postgres.
func (s *OpsService) GetNginxTimingOverview(ctx context.Context, filter *OpsNginxTimingFilter) (*OpsNginxTimingOverview, error) {
	if err := s.RequireMonitoringEnabled(ctx); err != nil {
		return nil, err
	}
	effective, windowClamped, err := normalizeOpsNginxTimingFilter(filter)
	if err != nil {
		return nil, err
	}

	var clientRequestKeys map[string]OpsNginxTimingRequestKey
	if len(effective.APIKeyIDs) > 0 {
		resolver, ok := s.opsRepo.(OpsNginxTimingKeyResolver)
		if !ok {
			return nil, infraerrors.ServiceUnavailable("OPS_NGINX_KEY_CORRELATION_UNAVAILABLE", "Nginx Key correlation is not available")
		}
		clientRequestKeys, err = resolver.ListNginxTimingRequestKeys(ctx, &effective)
		if err != nil {
			return nil, err
		}
	}

	overview, err := readOpsNginxTimingLog(opsNginxTimingLogPath(), &effective, clientRequestKeys)
	if err != nil {
		return nil, infraerrors.InternalServer("OPS_NGINX_TIMING_READ_FAILED", "Failed to read Nginx timing log").WithCause(err)
	}
	overview.WindowClamped = windowClamped
	return overview, nil
}

// GetNginxTimingKeyDetails reads the Nginx log only when an administrator
// opens a card's Key-level detail view.
func (s *OpsService) GetNginxTimingKeyDetails(ctx context.Context, filter *OpsNginxTimingFilter) (*OpsNginxTimingKeyDetails, error) {
	if err := s.RequireMonitoringEnabled(ctx); err != nil {
		return nil, err
	}
	effective, windowClamped, err := normalizeOpsNginxTimingFilter(filter)
	if err != nil {
		return nil, err
	}

	resolver, ok := s.opsRepo.(OpsNginxTimingKeyResolver)
	if !ok {
		return nil, infraerrors.ServiceUnavailable("OPS_NGINX_KEY_CORRELATION_UNAVAILABLE", "Nginx Key correlation is not available")
	}
	clientRequestKeys, err := resolver.ListNginxTimingRequestKeys(ctx, &effective)
	if err != nil {
		return nil, err
	}

	details, err := readOpsNginxTimingKeyDetailsLog(opsNginxTimingLogPath(), &effective, clientRequestKeys)
	if err != nil {
		return nil, infraerrors.InternalServer("OPS_NGINX_TIMING_READ_FAILED", "Failed to read Nginx timing log").WithCause(err)
	}
	details.WindowClamped = windowClamped
	return details, nil
}

func normalizeOpsNginxTimingFilter(filter *OpsNginxTimingFilter) (OpsNginxTimingFilter, bool, error) {
	if filter == nil || filter.StartTime.IsZero() || filter.EndTime.IsZero() {
		return OpsNginxTimingFilter{}, false, infraerrors.BadRequest("OPS_NGINX_TIME_RANGE_REQUIRED", "start_time/end_time are required")
	}
	if filter.StartTime.After(filter.EndTime) {
		return OpsNginxTimingFilter{}, false, infraerrors.BadRequest("OPS_NGINX_TIME_RANGE_INVALID", "start_time must be <= end_time")
	}

	effective := *filter
	effective.StartTime = effective.StartTime.UTC()
	effective.EndTime = effective.EndTime.UTC()
	if effective.EndTime.Sub(effective.StartTime) > maxOpsNginxTimingWindow {
		effective.StartTime = effective.EndTime.Add(-maxOpsNginxTimingWindow)
		return effective, true, nil
	}
	return effective, false, nil
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

func readOpsNginxTimingLog(path string, filter *OpsNginxTimingFilter, clientRequestKeys map[string]OpsNginxTimingRequestKey) (*OpsNginxTimingOverview, error) {
	if filter == nil {
		return nil, infraerrors.BadRequest("OPS_NGINX_FILTER_REQUIRED", "filter is required")
	}

	overview := &OpsNginxTimingOverview{
		StartTime:        filter.StartTime.UTC(),
		EndTime:          filter.EndTime.UTC(),
		KeyFilterApplied: len(filter.APIKeyIDs) > 0,
	}
	acc := nginxTimingAccumulator{}

	for _, candidate := range nginxTimingLogPaths(path) {
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
				if _, ok := clientRequestKeys[clientRequestID]; !ok {
					return
				}
				overview.MatchedRequestCount++
			}

			collectNginxTimingEntry(overview, &acc, entry, at, false)
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
	finishNginxTimingOverview(overview, &acc)
	return overview, nil
}

func readOpsNginxTimingKeyDetailsLog(path string, filter *OpsNginxTimingFilter, clientRequestKeys map[string]OpsNginxTimingRequestKey) (*OpsNginxTimingKeyDetails, error) {
	if filter == nil {
		return nil, infraerrors.BadRequest("OPS_NGINX_FILTER_REQUIRED", "filter is required")
	}

	details := &OpsNginxTimingKeyDetails{
		StartTime:        filter.StartTime.UTC(),
		EndTime:          filter.EndTime.UTC(),
		KeyFilterApplied: len(filter.APIKeyIDs) > 0,
		Items:            []OpsNginxTimingKeyMetric{},
	}
	accumulators := make(map[int64]*nginxTimingKeyAccumulator)

	for _, candidate := range nginxTimingLogPaths(path) {
		found, updatedAt, err := scanOpsNginxTimingLog(candidate, func(entry nginxTimingLogEntry, at time.Time) {
			if at.Before(filter.StartTime) || at.After(filter.EndTime) || !entry.isGatewayRequest() {
				return
			}

			requestKey, ok := clientRequestKeys[strings.TrimSpace(entry.ClientRequestID)]
			if !ok {
				if entry.Status >= 400 {
					details.UnattributedErrorCount++
				}
				return
			}
			details.MatchedRequestCount++

			item := accumulators[requestKey.APIKeyID]
			if item == nil {
				item = &nginxTimingKeyAccumulator{key: requestKey}
				accumulators[requestKey.APIKeyID] = item
			}
			collectNginxTimingEntry(&item.overview, &item.acc, entry, at, true)
		})
		if err != nil {
			return nil, err
		}
		if found {
			details.Available = true
			if details.SourceUpdatedAt == nil || (updatedAt != nil && updatedAt.After(*details.SourceUpdatedAt)) {
				details.SourceUpdatedAt = updatedAt
			}
		}
	}

	for _, item := range accumulators {
		finishNginxTimingOverview(&item.overview, &item.acc)
		details.Items = append(details.Items, OpsNginxTimingKeyMetric{
			APIKeyID:                         item.key.APIKeyID,
			KeyName:                          item.key.KeyName,
			HTTPRequestCount:                 item.overview.HTTPRequestCount,
			WebSocketSessionCount:            item.overview.WebSocketSessionCount,
			SuccessCount:                     item.overview.SuccessCount,
			ClientTimeout408Count:            item.overview.ClientTimeout408Count,
			ClientClosed499Count:             item.overview.ClientClosed499Count,
			ServerError5xxCount:              item.overview.ServerError5xxCount,
			UpstreamUnreachedCount:           item.overview.UpstreamUnreachedCount,
			RequestTime:                      item.overview.RequestTime,
			UpstreamConnect:                  item.overview.UpstreamConnect,
			UpstreamHeader:                   item.overview.UpstreamHeader,
			UpstreamResponse:                 item.overview.UpstreamResponse,
			ClientOverhead:                   item.overview.ClientOverhead,
			ClientOverheadSampleCount:        int64(len(item.acc.clientOverhead)),
			ClientUploadSampleCount:          int64(len(item.acc.clientUpload)),
			ClientUpload:                     opsNginxPercentiles(item.acc.clientUpload),
			ClientResponseReceiveSampleCount: int64(len(item.acc.clientResponseReceive)),
			ClientResponseReceive:            opsNginxPercentiles(item.acc.clientResponseReceive),
		})
	}
	sort.Slice(details.Items, func(i, j int) bool {
		if details.Items[i].KeyName == details.Items[j].KeyName {
			return details.Items[i].APIKeyID < details.Items[j].APIKeyID
		}
		return details.Items[i].KeyName < details.Items[j].KeyName
	})
	return details, nil
}

func nginxTimingLogPaths(path string) []string {
	return []string{
		opsNginxTimingLegacyLogPath() + ".1",
		opsNginxTimingLegacyLogPath(),
		path + ".1",
		path,
	}
}

func collectNginxTimingEntry(overview *OpsNginxTimingOverview, acc *nginxTimingAccumulator, entry nginxTimingLogEntry, completedAt time.Time, collectClientTiming bool) {
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
	if collectClientTiming {
		if gatewayReceivedAtMS, ok := nginxTimingUnixMS(entry.GatewayReceivedAtMS); ok {
			completedAtMS := nginxTimingCompletedAtMS(entry.CompletedAtMS, completedAt)
			if hasRequestTime {
				clientUploadMS := gatewayReceivedAtMS - (completedAtMS - int64(requestTimeMS))
				if clientUploadMS >= 0 {
					acc.clientUpload = append(acc.clientUpload, int(clientUploadMS))
				}
			}
			if hasUpstreamResponse {
				clientResponseReceiveMS := completedAtMS - gatewayReceivedAtMS - int64(upstreamResponseMS)
				if clientResponseReceiveMS < 0 {
					clientResponseReceiveMS = 0
				}
				acc.clientResponseReceive = append(acc.clientResponseReceive, int(clientResponseReceiveMS))
			}
		}
	}
	if entry.RequestLength > 0 {
		acc.requestBytes += entry.RequestLength
		acc.requestByteCount++
		if entry.RequestLength > acc.maxRequestBytes {
			acc.maxRequestBytes = entry.RequestLength
		}
	}
}

func finishNginxTimingOverview(overview *OpsNginxTimingOverview, acc *nginxTimingAccumulator) {
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
}

func scanOpsNginxTimingLog(path string, visit func(nginxTimingLogEntry, time.Time)) (bool, *time.Time, error) {
	file, err := os.Open(path)
	if os.IsNotExist(err) {
		return false, nil, nil
	}
	if err != nil {
		return false, nil, err
	}
	defer func() { _ = file.Close() }()

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

func nginxTimingUnixMS(raw string) (int64, bool) {
	value, err := strconv.ParseInt(strings.TrimSpace(raw), 10, 64)
	return value, err == nil && value > 0
}

func nginxTimingCompletedAtMS(raw string, fallback time.Time) int64 {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "-" {
		return fallback.UnixMilli()
	}
	if strings.Contains(raw, ".") {
		if seconds, err := strconv.ParseFloat(raw, 64); err == nil && seconds > 0 {
			return int64(math.Round(seconds * 1000))
		}
		return fallback.UnixMilli()
	}
	if value, ok := nginxTimingUnixMS(raw); ok {
		if value > 10_000_000_000 {
			return value
		}
		return value * 1000
	}
	return fallback.UnixMilli()
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
