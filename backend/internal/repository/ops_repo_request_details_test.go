package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

// ListRequestDetails 的最终列清单：自定义的 first_token_ms（TTFT）与上游的
// api_key_name 在同一份实现里共存，两个测试共用这份清单，避免再次错位。
var requestDetailTestColumns = []string{
	"kind", "created_at", "request_id", "platform", "model", "duration_ms", "first_token_ms",
	"status_code", "error_id", "phase", "severity", "message", "user_id", "api_key_id",
	"api_key_name", "account_id", "group_id", "stream",
}

func TestOpsRepositoryListRequestDetails_LatencySort(t *testing.T) {
	for _, tc := range []struct {
		name  string
		sort  string
		order string
	}{
		{name: "TTFT", sort: "ttft_desc", order: "first_token_ms DESC NULLS LAST, created_at DESC"},
		{name: "duration", sort: "duration_desc", order: "duration_ms DESC NULLS LAST, created_at DESC"},
		{name: "default", order: "created_at DESC"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db, mock := newSQLMock(t)
			repo := &opsRepository{db: db}
			start := time.Date(2026, 9, 10, 0, 0, 0, 0, time.UTC)
			end := start.Add(time.Hour)
			filter := &service.OpsRequestDetailFilter{
				StartTime: &start,
				EndTime:   &end,
				Sort:      tc.sort,
				Page:      2,
				PageSize:  10,
			}

			mock.ExpectQuery(`SELECT COUNT\(1\) FROM combined`).
				WithArgs(start, end).
				WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(13))
			rows := sqlmock.NewRows(requestDetailTestColumns).
				AddRow("success", start, "req-slow", "openai", "gpt-5.5", 12000, 800, nil, nil, nil, nil, nil, 1, 2, "key-a", 3, 4, true).
				AddRow("error", start, "req-zero", "openai", "gpt-5.5", 9000, 0, 502, 5, "upstream", "error", "failed", 1, 2, "key-b", 3, 4, true).
				AddRow("success", start, "req-missing", "openai", "gpt-5.5", 5000, nil, nil, nil, nil, nil, nil, 1, 2, "key-c", 3, 4, false)
			mock.ExpectQuery(`(?s)ul\.first_token_ms AS first_token_ms.*o\.time_to_first_token_ms AS first_token_ms.*SELECT.*duration_ms,\s+first_token_ms,.*ORDER BY `+tc.order+`\s+LIMIT \$3 OFFSET \$4`).
				WithArgs(start, end, 10, 10).
				WillReturnRows(rows)

			items, total, err := repo.ListRequestDetails(context.Background(), filter)
			require.NoError(t, err)
			require.EqualValues(t, 13, total)
			require.Len(t, items, 3)
			require.NotNil(t, items[0].FirstTokenMs)
			require.Equal(t, 800, *items[0].FirstTokenMs)
			require.Equal(t, 12000, *items[0].DurationMs)
			require.NotNil(t, items[1].FirstTokenMs)
			require.Zero(t, *items[1].FirstTokenMs)
			require.Nil(t, items[2].FirstTokenMs)
			require.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestListRequestDetailsIncludesAPIKeyName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	start := time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM combined`).
		WithArgs(start, end).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`(?s)SELECT\s+kind,\s+created_at,\s+request_id,\s+platform,\s+model,\s+duration_ms,\s+first_token_ms,\s+status_code,\s+error_id,\s+phase,\s+severity,\s+message,\s+user_id,\s+api_key_id,\s+api_key_name,\s+account_id,\s+group_id,\s+stream\s+FROM combined`).
		WithArgs(start, end, 10, 0).
		WillReturnRows(sqlmock.NewRows(requestDetailTestColumns).
			AddRow("success", start, "req-1", "openai", "gpt-5", 1234, nil, nil, nil, nil, nil, nil, 1, 2, "team-key", 3, 4, true))

	items, total, err := (&opsRepository{db: db}).ListRequestDetails(context.Background(), &service.OpsRequestDetailFilter{
		StartTime: &start,
		EndTime:   &end,
		Page:      1,
		PageSize:  10,
	})

	require.NoError(t, err)
	require.Equal(t, int64(1), total)
	require.Len(t, items, 1)
	require.Equal(t, "team-key", items[0].APIKeyName)
	require.NoError(t, mock.ExpectationsWereMet())
}
