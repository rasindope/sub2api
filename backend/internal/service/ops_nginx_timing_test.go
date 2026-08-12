package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestReadOpsNginxTimingLogSeparatesGatewayAndKeyScope(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub2api_timing.json.log")
	base := time.Date(2026, 8, 7, 5, 0, 0, 0, time.UTC)
	completedAt := base.Add(time.Minute + 750*time.Millisecond)
	lines := []string{
		fmt.Sprintf(`{"timestamp":"2026-08-07T05:01:00Z","completed_at_ms":"%.3f","path":"/responses","gateway":"1","status":200,"request_time":"2.000","upstream_connect_time":"0.001","upstream_header_time":"1.500","upstream_response_time":"1.800","request_length":2048,"client_request_id":"key-a","gateway_received_at_ms":"%d"}`, float64(completedAt.UnixMilli())/1000, completedAt.Add(-1900*time.Millisecond).UnixMilli()),
		`{"timestamp":"2026-08-07T05:02:00Z","path":"/responses","gateway":"1","status":408,"request_time":"1800.000","upstream_connect_time":"-","upstream_header_time":"-","upstream_response_time":"-","request_length":4096,"client_request_id":""}`,
		`{"timestamp":"2026-08-07T05:03:00Z","path":"/responses","gateway":"1","status":499,"request_time":"40.000","upstream_connect_time":"-","upstream_header_time":"-","upstream_response_time":"-","request_length":1024,"client_request_id":""}`,
		`{"timestamp":"2026-08-07T05:04:00Z","path":"/responses","gateway":"1","upgrade":"websocket","status":101,"request_time":"60.000","upstream_connect_time":"0.001","upstream_header_time":"0.003","upstream_response_time":"60.000","request_length":512,"client_request_id":"key-a"}`,
		`{"timestamp":"2026-08-07T05:05:00Z","path":"/api/v1/admin/ops/dashboard/overview","gateway":"0","status":200,"request_time":"0.020","upstream_connect_time":"0.001","upstream_header_time":"0.010","upstream_response_time":"0.020","request_length":512,"client_request_id":"admin"}`,
		`{"timestamp":"2026-08-07T05:06:00Z","path":"/responses","gateway":"1","status":200,"request_time":"3.000","upstream_connect_time":"0.001","upstream_header_time":"2.500","upstream_response_time":"2.800","request_length":1024,"client_request_id":"key-b"}`,
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	filter := &OpsNginxTimingFilter{StartTime: base, EndTime: base.Add(10 * time.Minute)}
	all, err := readOpsNginxTimingLog(path, filter, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !all.Available || all.HTTPRequestCount != 4 || all.WebSocketSessionCount != 1 {
		t.Fatalf("unexpected all scope: %+v", all)
	}
	if all.ClientTimeout408Count != 1 || all.ClientClosed499Count != 1 || all.UpstreamUnreachedCount != 2 {
		t.Fatalf("unexpected error split: %+v", all)
	}
	if all.RequestTime.P90 == nil || *all.RequestTime.P90 != 1800000 {
		t.Fatalf("unexpected request p90: %+v", all.RequestTime)
	}
	if all.UpstreamHeader.P50 == nil || *all.UpstreamHeader.P50 != 1500 {
		t.Fatalf("missing upstream headers must not be counted as zero: %+v", all.UpstreamHeader)
	}
	filter.APIKeyIDs = []int64{7}
	selected, err := readOpsNginxTimingLog(path, filter, map[string]OpsNginxTimingRequestKey{
		"key-a": {APIKeyID: 7, KeyName: "Key A"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if selected.HTTPRequestCount != 1 || selected.WebSocketSessionCount != 1 || selected.MatchedRequestCount != 2 {
		t.Fatalf("unexpected key scope: %+v", selected)
	}
	if selected.UnattributedErrorCount != 2 || selected.ClientTimeout408Count != 0 || selected.ClientClosed499Count != 0 {
		t.Fatalf("unexpected unattributed split: %+v", selected)
	}

	details, err := readOpsNginxTimingKeyDetailsLog(path, &OpsNginxTimingFilter{StartTime: base, EndTime: base.Add(10 * time.Minute)}, map[string]OpsNginxTimingRequestKey{
		"key-a": {APIKeyID: 7, KeyName: "Key A"},
		"key-b": {APIKeyID: 8, KeyName: "Key B"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if !details.Available || details.MatchedRequestCount != 3 || details.UnattributedErrorCount != 2 || len(details.Items) != 2 {
		t.Fatalf("unexpected Key details: %+v", details)
	}
	if details.Items[0].APIKeyID != 7 || details.Items[0].HTTPRequestCount != 1 || details.Items[0].WebSocketSessionCount != 1 {
		t.Fatalf("unexpected Key A details: %+v", details.Items[0])
	}
	if details.Items[0].ClientUploadSampleCount != 1 || details.Items[0].ClientUpload.P99 == nil || *details.Items[0].ClientUpload.P99 != 100 {
		t.Fatalf("unexpected Key A upload details: %+v", details.Items[0])
	}
	if details.Items[0].ClientResponseReceiveSampleCount != 1 || details.Items[0].ClientResponseReceive.P99 == nil || *details.Items[0].ClientResponseReceive.P99 != 100 {
		t.Fatalf("unexpected Key A response receive details: %+v", details.Items[0])
	}
	if details.Items[1].APIKeyID != 8 || details.Items[1].RequestTime.P99 == nil || *details.Items[1].RequestTime.P99 != 3000 {
		t.Fatalf("unexpected Key B details: %+v", details.Items[1])
	}
}

func TestCollectNginxTimingEntryKeepsZeroResponseReceiveSamples(t *testing.T) {
	completedAt := time.Date(2026, 8, 7, 5, 1, 0, 0, time.UTC)
	overview := &OpsNginxTimingOverview{}
	acc := nginxTimingAccumulator{}
	entry := nginxTimingLogEntry{
		Status:               200,
		RequestTime:          "2.000",
		UpstreamResponseTime: "1.800",
		CompletedAtMS:        fmt.Sprintf("%.3f", float64(completedAt.UnixMilli())/1000),
		GatewayReceivedAtMS:  strconv.FormatInt(completedAt.Add(-1500*time.Millisecond).UnixMilli(), 10),
	}

	collectNginxTimingEntry(overview, &acc, entry, completedAt, true)

	if len(acc.clientResponseReceive) != 1 || acc.clientResponseReceive[0] != 0 {
		t.Fatalf("response receive samples = %v, want [0]", acc.clientResponseReceive)
	}
}

func TestReadOpsNginxTimingLogReadsLegacyRetention(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sub2api_timing.json.log")
	legacyPath := filepath.Join(dir, "sub2api_timing.legacy.log")
	t.Setenv("OPS_NGINX_TIMING_LEGACY_LOG_PATH", legacyPath)

	line := `47.82.254.30 [07/Aug/2026:13:01:00 +0800] "POST /responses HTTP/1.1" status=200 rt=2.000 uct=0.001 uht=1.500 urt=1.800 us=200 req_len=2048 bytes=1024 ua="Codex CLI" crid="key-a"`
	if err := os.WriteFile(legacyPath, []byte(line+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	base := time.Date(2026, 8, 7, 5, 0, 0, 0, time.UTC)
	result, err := readOpsNginxTimingLog(path, &OpsNginxTimingFilter{
		StartTime: base,
		EndTime:   base.Add(10 * time.Minute),
		APIKeyIDs: []int64{7},
	}, map[string]OpsNginxTimingRequestKey{"key-a": {APIKeyID: 7, KeyName: "Key A"}})
	if err != nil {
		t.Fatal(err)
	}
	if !result.Available || result.HTTPRequestCount != 1 || result.MatchedRequestCount != 1 {
		t.Fatalf("unexpected legacy result: %+v", result)
	}
	if result.RequestTime.P90 == nil || *result.RequestTime.P90 != 2000 {
		t.Fatalf("unexpected legacy request time: %+v", result.RequestTime)
	}
	if result.UpstreamHeader.P90 == nil || *result.UpstreamHeader.P90 != 1500 {
		t.Fatalf("unexpected legacy upstream header: %+v", result.UpstreamHeader)
	}
}
