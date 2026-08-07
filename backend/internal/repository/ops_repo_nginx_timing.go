package repository

import (
	"context"
	"fmt"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/lib/pq"
)

// ListNginxTimingRequestKeys maps Nginx's response correlation header to API
// Key metadata without ever persisting Key material in the Nginx log.
func (r *opsRepository) ListNginxTimingRequestKeys(ctx context.Context, filter *service.OpsNginxTimingFilter) (map[string]service.OpsNginxTimingRequestKey, error) {
	if r == nil || r.db == nil {
		return nil, fmt.Errorf("nil ops repository")
	}
	if filter == nil {
		return map[string]service.OpsNginxTimingRequestKey{}, nil
	}

	args := []any{filter.StartTime, filter.EndTime}
	keyWhere := ""
	if len(filter.APIKeyIDs) > 0 {
		keyWhere = " AND api_key_id = ANY($3)"
		args = append(args, pq.Array(filter.APIKeyIDs))
	}

	rows, err := r.db.QueryContext(ctx, `
WITH matched AS (
  SELECT substring(request_id FROM 8) AS client_request_id, api_key_id
  FROM usage_logs
  WHERE created_at >= $1
    AND created_at <= $2
    AND request_id LIKE 'client:%'`+keyWhere+`

  UNION

  SELECT client_request_id, api_key_id
  FROM ops_error_logs
  WHERE created_at >= $1
    AND created_at <= $2
    AND client_request_id IS NOT NULL
    AND client_request_id <> ''`+keyWhere+`
)
SELECT DISTINCT ON (matched.client_request_id)
  matched.client_request_id,
  matched.api_key_id,
  COALESCE(NULLIF(api_keys.name, ''), 'Key #' || matched.api_key_id::text)
FROM matched
LEFT JOIN api_keys ON api_keys.id = matched.api_key_id
WHERE matched.client_request_id IS NOT NULL
  AND matched.client_request_id <> ''
  AND matched.api_key_id IS NOT NULL
ORDER BY matched.client_request_id`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	keys := make(map[string]service.OpsNginxTimingRequestKey)
	for rows.Next() {
		var (
			clientRequestID string
			apiKeyID        int64
			keyName         string
		)
		if err := rows.Scan(&clientRequestID, &apiKeyID, &keyName); err != nil {
			return nil, err
		}
		if clientRequestID = strings.TrimSpace(clientRequestID); clientRequestID != "" {
			keys[clientRequestID] = service.OpsNginxTimingRequestKey{
				APIKeyID: apiKeyID,
				KeyName:  strings.TrimSpace(keyName),
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return keys, nil
}
