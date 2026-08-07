package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

// ListNginxTimingClientRequestIDs maps Nginx's response correlation header to
// selected Key IDs without ever persisting Key material in the Nginx log.
func (r *opsRepository) ListNginxTimingClientRequestIDs(ctx context.Context, filter *service.OpsNginxTimingFilter) (map[string]struct{}, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil ops repository")
	}
	if filter == nil || len(filter.APIKeyIDs) == 0 {
		return map[string]struct{}{}, nil
	}

	rows, err := r.db.QueryContext(ctx, `
SELECT client_request_id
FROM (
  SELECT substring(request_id FROM 8) AS client_request_id
  FROM usage_logs
  WHERE api_key_id = ANY($1)
    AND created_at >= $2
    AND created_at <= $3
    AND request_id LIKE 'client:%'

  UNION

  SELECT client_request_id
  FROM ops_error_logs
  WHERE api_key_id = ANY($1)
    AND created_at >= $2
    AND created_at <= $3
    AND client_request_id IS NOT NULL
    AND client_request_id <> ''
) matched
WHERE client_request_id IS NOT NULL
  AND client_request_id <> ''`, pq.Array(filter.APIKeyIDs), filter.StartTime, filter.EndTime)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ids := make(map[string]struct{})
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		if id = strings.TrimSpace(id); id != "" {
			ids[id] = struct{}{}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ids, nil
}
