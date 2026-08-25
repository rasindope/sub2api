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
			SELECT id, api_key_id, BTRIM(ip_address) AS ip_address, created_at, duration_ms
			FROM usage_logs
			WHERE created_at >= $1::timestamptz AND created_at < $2::timestamptz
				AND ip_address IS NOT NULL AND BTRIM(ip_address) <> ''
		), active AS (
			SELECT id, api_key_id, BTRIM(ip_address) AS ip_address, created_at, duration_ms,
				created_at - duration_ms * INTERVAL '1 millisecond' AS started_at
			FROM usage_logs
			WHERE created_at >= $3::timestamptz - INTERVAL '15 minutes' AND created_at <= $3::timestamptz
				AND duration_ms > 0 AND ip_address IS NOT NULL AND BTRIM(ip_address) <> ''
		), overlap_pairs AS (
			SELECT a.api_key_id, a.ip_address AS ip_a, b.ip_address AS ip_b,
				EXTRACT(EPOCH FROM LEAST(a.created_at, b.created_at) - GREATEST(a.started_at, b.started_at)) AS overlap_seconds,
				LEAST(a.created_at, b.created_at) AS overlap_at
			FROM active a
			JOIN active b ON b.api_key_id = a.api_key_id AND b.id > a.id AND b.ip_address <> a.ip_address
				AND a.started_at < b.created_at AND b.started_at < a.created_at
		), significant AS (
			SELECT * FROM overlap_pairs WHERE overlap_seconds >= 5
		), overlap_ips AS (
			SELECT api_key_id, ip_a AS ip_address, overlap_seconds, overlap_at FROM significant
			UNION ALL
			SELECT api_key_id, ip_b AS ip_address, overlap_seconds, overlap_at FROM significant
		), per_ip_history AS (
			SELECT api_key_id, ip_address, COUNT(*) AS requests, MIN(created_at) AS first_seen_at,
				MAX(created_at) AS last_seen_at
			FROM history GROUP BY api_key_id, ip_address
		), per_ip_overlap AS (
			SELECT api_key_id, ip_address, COUNT(*) AS overlap_count_15m,
				MAX(overlap_seconds) AS max_overlap_seconds_15m, MAX(overlap_at) AS last_overlap_at
			FROM overlap_ips GROUP BY api_key_id, ip_address
		), per_ip AS (
			SELECT h.*, EXISTS (SELECT 1 FROM active a WHERE a.api_key_id = h.api_key_id AND a.ip_address = h.ip_address) AS active_15m,
				COALESCE(o.overlap_count_15m, 0) AS overlap_count_15m,
				COALESCE(o.max_overlap_seconds_15m, 0) AS max_overlap_seconds_15m,
				o.last_overlap_at
			FROM per_ip_history h
			LEFT JOIN per_ip_overlap o ON o.api_key_id = h.api_key_id AND o.ip_address = h.ip_address
		), ranked_ips AS (
			SELECT *, ROW_NUMBER() OVER (PARTITION BY api_key_id ORDER BY requests DESC, last_seen_at DESC) AS ip_rank
			FROM per_ip
		), ip_json AS (
			SELECT api_key_id, JSONB_AGG(JSONB_BUILD_OBJECT(
				'ip_address', ip_address, 'requests', requests,
				'first_seen_at', TO_CHAR(first_seen_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'),
				'last_seen_at', TO_CHAR(last_seen_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'),
				'active_15m', active_15m, 'overlap_count_15m', overlap_count_15m,
				'max_overlap_seconds_15m', max_overlap_seconds_15m,
				'last_overlap_at', COALESCE(TO_CHAR(last_overlap_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), '')
			) ORDER BY requests DESC, last_seen_at DESC) FILTER (WHERE ip_rank <= 20) AS ip_usages
			FROM ranked_ips GROUP BY api_key_id
		), key_history AS (
			SELECT api_key_id, COUNT(*) AS requests, COUNT(DISTINCT ip_address) AS distinct_ip_count
			FROM history GROUP BY api_key_id
		), key_active AS (
			SELECT api_key_id, COUNT(DISTINCT ip_address) AS active_ip_count_15m FROM active GROUP BY api_key_id
		), overlap_ip_counts AS (
			SELECT api_key_id, COUNT(DISTINCT ip_address) AS overlap_ip_count_15m
			FROM overlap_ips GROUP BY api_key_id
		), key_overlap AS (
			SELECT s.api_key_id, COUNT(*) AS overlap_count_15m, MAX(oi.overlap_ip_count_15m) AS overlap_ip_count_15m,
				SUM(s.overlap_seconds) AS total_overlap_seconds_15m,
				MAX(s.overlap_seconds) AS max_overlap_seconds_15m, MAX(s.overlap_at) AS last_overlap_at
			FROM significant s
			JOIN overlap_ip_counts oi ON oi.api_key_id = s.api_key_id
			GROUP BY s.api_key_id
		), key_ids AS (
			SELECT api_key_id FROM key_history
			UNION SELECT api_key_id FROM key_active
		), activity AS (
			SELECT ids.api_key_id, COALESCE(k.name, '') AS key_name, COALESCE(kh.requests, 0) AS requests,
				COALESCE(kh.distinct_ip_count, 0) AS distinct_ip_count,
				COALESCE(ka.active_ip_count_15m, 0) AS active_ip_count_15m,
				COALESCE(ko.overlap_ip_count_15m, 0) AS overlap_ip_count_15m,
				COALESCE(ko.overlap_count_15m, 0) AS overlap_count_15m,
				COALESCE(ko.total_overlap_seconds_15m, 0) AS total_overlap_seconds_15m,
				COALESCE(ko.max_overlap_seconds_15m, 0) AS max_overlap_seconds_15m,
				ko.last_overlap_at, COALESCE(ij.ip_usages, '[]'::jsonb) AS ip_usages,
				CASE WHEN COALESCE(ko.overlap_ip_count_15m, 0) >= 3
					OR (COALESCE(ko.overlap_count_15m, 0) >= 3 AND COALESCE(ko.total_overlap_seconds_15m, 0) >= 30) THEN 'high'
					WHEN COALESCE(ko.overlap_count_15m, 0) > 0 THEN 'watch' ELSE 'normal' END AS risk_level
			FROM key_ids ids
			LEFT JOIN key_history kh ON kh.api_key_id = ids.api_key_id
			LEFT JOIN api_keys k ON k.id = ids.api_key_id
			LEFT JOIN key_active ka ON ka.api_key_id = ids.api_key_id
			LEFT JOIN key_overlap ko ON ko.api_key_id = ids.api_key_id
			LEFT JOIN ip_json ij ON ij.api_key_id = ids.api_key_id
		), limited AS (
			SELECT * FROM activity ORDER BY CASE risk_level WHEN 'high' THEN 0 WHEN 'watch' THEN 1 ELSE 2 END,
				active_ip_count_15m DESC, requests DESC LIMIT $4
		)
		SELECT api_key_id, key_name, requests, distinct_ip_count, active_ip_count_15m,
			overlap_ip_count_15m, overlap_count_15m, total_overlap_seconds_15m,
			max_overlap_seconds_15m,
			COALESCE(TO_CHAR(last_overlap_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), ''),
			risk_level, ip_usages,
			COUNT(*) FILTER (WHERE active_ip_count_15m > 0) OVER (),
			COUNT(*) FILTER (WHERE risk_level = 'watch') OVER (),
			COUNT(*) FILTER (WHERE risk_level = 'high') OVER ()
		FROM limited
		ORDER BY CASE risk_level WHEN 'high' THEN 0 WHEN 'watch' THEN 1 ELSE 2 END, active_ip_count_15m DESC, requests DESC
	`, startTime, endTime, now, limit)
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
		if err = rows.Scan(&item.APIKeyID, &item.KeyName, &item.Requests, &item.DistinctIPCount,
			&item.ActiveIPCount15m, &item.OverlapIPCount15m, &item.OverlapCount15m,
			&item.TotalOverlapSeconds15m, &item.MaxOverlapSeconds15m, &item.LastOverlapAt,
			&item.RiskLevel, &ipUsages, &result.ActiveKeys, &result.WatchKeys, &result.HighRiskKeys); err != nil {
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
	return result, nil
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
	active := &apiKeyIPActiveHeap{}
	recent := &apiKeyIPOverlapHeap{}
	heap.Init(active)
	heap.Init(recent)
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
			if endAt.Sub(request.startAt) < 5*time.Second {
				continue
			}
			event := apiKeyIPOverlapEvent{ipA: previous.ip, ipB: request.ip, startAt: request.startAt, endAt: endAt}
			if recent.Len() < limit {
				heap.Push(recent, event)
			} else if oldest := (*recent)[0]; event.endAt.After(oldest.endAt) ||
				(event.endAt.Equal(oldest.endAt) && event.startAt.After(oldest.startAt)) {
				heap.Pop(recent)
				heap.Push(recent, event)
			}
		}
		heap.Push(active, request)
	}

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
