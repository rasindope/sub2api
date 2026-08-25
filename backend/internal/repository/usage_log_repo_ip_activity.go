package repository

import (
	"container/heap"
	"context"
	"encoding/json"
	"sort"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/usagestats"
)

func (r *usageLogRepository) GetAPIKeyIPActivity(ctx context.Context, startTime, endTime, now time.Time, limit int) (result *usagestats.APIKeyIPActivityResponse, err error) {
	if limit <= 0 {
		limit = 100
	}

	rows, err := r.sql.QueryContext(ctx, `
			WITH history AS (
				SELECT api_key_id, BTRIM(ip_address) AS ip_address, created_at
				FROM usage_logs
				WHERE created_at >= $1::timestamptz AND created_at < $2::timestamptz
					AND ip_address IS NOT NULL AND BTRIM(ip_address) <> ''
			), per_ip_history AS (
				SELECT api_key_id, ip_address, COUNT(*) AS requests, MIN(created_at) AS first_seen_at,
					MAX(created_at) AS last_seen_at
				FROM history GROUP BY api_key_id, ip_address
			), ranked_ips AS (
				SELECT *, ROW_NUMBER() OVER (PARTITION BY api_key_id ORDER BY requests DESC, last_seen_at DESC) AS ip_rank
				FROM per_ip_history
			), ip_json AS (
				SELECT api_key_id, JSONB_AGG(JSONB_BUILD_OBJECT(
					'ip_address', ip_address, 'requests', requests,
					'first_seen_at', TO_CHAR(first_seen_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'),
					'last_seen_at', TO_CHAR(last_seen_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"')
				) ORDER BY requests DESC, last_seen_at DESC) FILTER (WHERE ip_rank <= 20) AS ip_usages
				FROM ranked_ips GROUP BY api_key_id
			), key_history AS (
				SELECT api_key_id, COUNT(*) AS requests, COUNT(DISTINCT ip_address) AS distinct_ip_count
				FROM history GROUP BY api_key_id
			)
			SELECT kh.api_key_id, COALESCE(k.name, ''), kh.requests, kh.distinct_ip_count,
				COALESCE(ij.ip_usages, '[]'::jsonb)
			FROM key_history kh
			LEFT JOIN api_keys k ON k.id = kh.api_key_id
			LEFT JOIN ip_json ij ON ij.api_key_id = kh.api_key_id
			ORDER BY kh.requests DESC
		`, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			result = nil
		}
	}()

	result = &usagestats.APIKeyIPActivityResponse{Items: make([]usagestats.APIKeyIPActivityItem, 0), GeneratedAt: now.Format(time.RFC3339)}
	for rows.Next() {
		var item usagestats.APIKeyIPActivityItem
		var ipUsages []byte
		if err = rows.Scan(&item.APIKeyID, &item.KeyName, &item.Requests, &item.DistinctIPCount, &ipUsages); err != nil {
			return nil, err
		}
		if err = json.Unmarshal(ipUsages, &item.IPUsages); err != nil {
			return nil, err
		}
		result.Items = append(result.Items, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	if err = rows.Close(); err != nil {
		return nil, err
	}
	if err = r.applyAPIKeyIPRangeOverlaps(ctx, startTime, endTime, result); err != nil {
		return nil, err
	}
	sort.SliceStable(result.Items, func(i, j int) bool {
		left, right := apiKeyIPRiskRank(result.Items[i].RiskLevel), apiKeyIPRiskRank(result.Items[j].RiskLevel)
		if left != right {
			return left < right
		}
		return result.Items[i].Requests > result.Items[j].Requests
	})
	result.ActiveKeys = int64(len(result.Items))
	for _, item := range result.Items {
		switch item.RiskLevel {
		case "watch":
			result.WatchKeys++
		case "high":
			result.HighRiskKeys++
		}
	}
	if len(result.Items) > limit {
		result.Items = result.Items[:limit]
	}
	return result, nil
}

type apiKeyIPOverlapStat struct {
	count int64
	max   float64
	last  time.Time
}

func apiKeyIPRiskRank(risk string) int {
	switch risk {
	case "high":
		return 0
	case "watch":
		return 1
	default:
		return 2
	}
}

func (r *usageLogRepository) applyAPIKeyIPRangeOverlaps(ctx context.Context, startTime, endTime time.Time, result *usagestats.APIKeyIPActivityResponse) (err error) {
	if result == nil || len(result.Items) == 0 {
		return nil
	}
	items := make(map[int64]*usagestats.APIKeyIPActivityItem, len(result.Items))
	for i := range result.Items {
		items[result.Items[i].APIKeyID] = &result.Items[i]
	}
	rows, err := r.sql.QueryContext(ctx, `
		SELECT id, api_key_id, BTRIM(ip_address), created_at, duration_ms
		FROM usage_logs
		WHERE created_at >= $1 AND created_at < $2
			AND duration_ms > 0 AND ip_address IS NOT NULL AND BTRIM(ip_address) <> ''
		ORDER BY api_key_id, created_at - duration_ms * INTERVAL '1 millisecond', id
	`, startTime, endTime)
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()

	var currentKeyID int64
	requests := make([]apiKeyIPRequestInterval, 0)
	flush := func() {
		item := items[currentKeyID]
		if item == nil || len(requests) == 0 {
			requests = requests[:0]
			return
		}
		overlapIPs := make(map[string]struct{})
		perIP := make(map[string]*apiKeyIPOverlapStat)
		var lastOverlap time.Time
		forEachAPIKeyIPOverlap(requests, func(event apiKeyIPOverlapEvent) {
			seconds := event.endAt.Sub(event.startAt).Seconds()
			item.OverlapCount15m++
			item.TotalOverlapSeconds15m += seconds
			if seconds > item.MaxOverlapSeconds15m {
				item.MaxOverlapSeconds15m = seconds
			}
			if event.endAt.After(lastOverlap) {
				lastOverlap = event.endAt
			}
			for _, ip := range []string{event.ipA, event.ipB} {
				overlapIPs[ip] = struct{}{}
				stat := perIP[ip]
				if stat == nil {
					stat = &apiKeyIPOverlapStat{}
					perIP[ip] = stat
				}
				stat.count++
				if seconds > stat.max {
					stat.max = seconds
				}
				if event.endAt.After(stat.last) {
					stat.last = event.endAt
				}
			}
		})
		if !lastOverlap.IsZero() {
			item.LastOverlapAt = lastOverlap.UTC().Format(time.RFC3339Nano)
		}
		item.OverlapIPCount15m = int64(len(overlapIPs))
		for i := range item.IPUsages {
			if stat := perIP[item.IPUsages[i].IPAddress]; stat != nil {
				item.IPUsages[i].OverlapCount15m = stat.count
				item.IPUsages[i].MaxOverlapSeconds15m = stat.max
				item.IPUsages[i].LastOverlapAt = stat.last.UTC().Format(time.RFC3339Nano)
			}
		}
		if item.OverlapIPCount15m >= 3 || (item.OverlapCount15m >= 3 && item.TotalOverlapSeconds15m >= 30) {
			item.RiskLevel = "high"
		} else if item.OverlapCount15m > 0 {
			item.RiskLevel = "watch"
		} else {
			item.RiskLevel = "normal"
		}
		requests = requests[:0]
	}
	for rows.Next() {
		var request apiKeyIPRequestInterval
		var apiKeyID, durationMS int64
		if err = rows.Scan(&request.id, &apiKeyID, &request.ip, &request.endAt, &durationMS); err != nil {
			return err
		}
		if currentKeyID != 0 && apiKeyID != currentKeyID {
			flush()
		}
		currentKeyID = apiKeyID
		request.startAt = request.endAt.Add(-time.Duration(durationMS) * time.Millisecond)
		requests = append(requests, request)
	}
	if err = rows.Err(); err != nil {
		return err
	}
	flush()
	return nil
}

type apiKeyIPRequestInterval struct {
	id      int64
	ip      string
	startAt time.Time
	endAt   time.Time
}

type apiKeyIPActiveHeap []apiKeyIPRequestInterval

func (h apiKeyIPActiveHeap) Len() int           { return len(h) }
func (h apiKeyIPActiveHeap) Less(i, j int) bool { return h[i].endAt.Before(h[j].endAt) }
func (h apiKeyIPActiveHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }
func (h *apiKeyIPActiveHeap) Push(value any)    { *h = append(*h, value.(apiKeyIPRequestInterval)) }
func (h *apiKeyIPActiveHeap) Pop() any {
	old := *h
	last := old[len(old)-1]
	*h = old[:len(old)-1]
	return last
}

type apiKeyIPOverlapEvent struct {
	ipA, ipB       string
	startAt, endAt time.Time
}

type apiKeyIPOverlapHeap []apiKeyIPOverlapEvent

func (h apiKeyIPOverlapHeap) Len() int      { return len(h) }
func (h apiKeyIPOverlapHeap) Swap(i, j int) { h[i], h[j] = h[j], h[i] }
func (h apiKeyIPOverlapHeap) Less(i, j int) bool {
	if h[i].endAt.Equal(h[j].endAt) {
		return h[i].startAt.Before(h[j].startAt)
	}
	return h[i].endAt.Before(h[j].endAt)
}
func (h *apiKeyIPOverlapHeap) Push(value any) { *h = append(*h, value.(apiKeyIPOverlapEvent)) }
func (h *apiKeyIPOverlapHeap) Pop() any {
	old := *h
	last := old[len(old)-1]
	*h = old[:len(old)-1]
	return last
}

func forEachAPIKeyIPOverlap(requests []apiKeyIPRequestInterval, visit func(apiKeyIPOverlapEvent)) {
	active := &apiKeyIPActiveHeap{}
	heap.Init(active)
	for _, request := range requests {
		for active.Len() > 0 && !(*active)[0].endAt.After(request.startAt) {
			heap.Pop(active)
		}
		for _, previous := range *active {
			if previous.ip == request.ip {
				continue
			}
			endAt := request.endAt
			if previous.endAt.Before(endAt) {
				endAt = previous.endAt
			}
			if endAt.Sub(request.startAt) >= 5*time.Second {
				visit(apiKeyIPOverlapEvent{ipA: previous.ip, ipB: request.ip, startAt: request.startAt, endAt: endAt})
			}
		}
		heap.Push(active, request)
	}
}

func (r *usageLogRepository) GetAPIKeyIPOverlaps(ctx context.Context, apiKeyID int64, startTime, endTime time.Time, limit int) (result *usagestats.APIKeyIPOverlapResponse, err error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	rows, err := r.sql.QueryContext(ctx, `
		SELECT id, BTRIM(ip_address), created_at, duration_ms
		FROM usage_logs
		WHERE api_key_id = $1 AND created_at >= $2 AND created_at < $3
			AND duration_ms > 0 AND ip_address IS NOT NULL AND BTRIM(ip_address) <> ''
		ORDER BY created_at ASC, id ASC
	`, apiKeyID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil && err == nil {
			err = closeErr
			result = nil
		}
	}()

	requests := make([]apiKeyIPRequestInterval, 0)
	for rows.Next() {
		var request apiKeyIPRequestInterval
		var durationMS int64
		if err = rows.Scan(&request.id, &request.ip, &request.endAt, &durationMS); err != nil {
			return nil, err
		}
		request.startAt = request.endAt.Add(-time.Duration(durationMS) * time.Millisecond)
		requests = append(requests, request)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	sort.Slice(requests, func(i, j int) bool {
		if requests[i].startAt.Equal(requests[j].startAt) {
			return requests[i].id < requests[j].id
		}
		return requests[i].startAt.Before(requests[j].startAt)
	})
	recent := &apiKeyIPOverlapHeap{}
	heap.Init(recent)
	forEachAPIKeyIPOverlap(requests, func(event apiKeyIPOverlapEvent) {
		if recent.Len() < limit {
			heap.Push(recent, event)
		} else if oldest := (*recent)[0]; event.endAt.After(oldest.endAt) ||
			(event.endAt.Equal(oldest.endAt) && event.startAt.After(oldest.startAt)) {
			heap.Pop(recent)
			heap.Push(recent, event)
		}
	})

	events := make([]apiKeyIPOverlapEvent, recent.Len())
	for i := len(events) - 1; i >= 0; i-- {
		events[i] = heap.Pop(recent).(apiKeyIPOverlapEvent)
	}
	result = &usagestats.APIKeyIPOverlapResponse{Items: make([]usagestats.APIKeyIPOverlap, 0, len(events))}
	for _, event := range events {
		result.Items = append(result.Items, usagestats.APIKeyIPOverlap{
			IPA: event.ipA, IPB: event.ipB,
			OverlapStartAt: event.startAt.UTC().Format(time.RFC3339Nano),
			OverlapEndAt:   event.endAt.UTC().Format(time.RFC3339Nano),
			OverlapSeconds: event.endAt.Sub(event.startAt).Seconds(),
		})
	}
	return result, nil
}
