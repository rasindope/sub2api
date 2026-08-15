package repository

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestListRequestDetailsIncludesAPIKeyName(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer func() { _ = db.Close() }()

	start := time.Date(2026, 8, 13, 0, 0, 0, 0, time.UTC)
	end := start.Add(time.Hour)
	mock.ExpectQuery(`SELECT COUNT\(1\) FROM combined`).
		WithArgs(start, end).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))
	mock.ExpectQuery(`SELECT kind, created_at, request_id, platform, model, duration_ms, status_code, error_id, phase, severity, message, user_id, api_key_id, api_key_name, account_id, group_id, stream FROM combined`).
		WithArgs(start, end, 10, 0).
		WillReturnRows(sqlmock.NewRows([]string{
			"kind", "created_at", "request_id", "platform", "model", "duration_ms", "status_code",
			"error_id", "phase", "severity", "message", "user_id", "api_key_id", "api_key_name",
			"account_id", "group_id", "stream",
		}).AddRow("success", start, "req-1", "openai", "gpt-5", 1234, nil, nil, nil, nil, nil, 1, 2, "team-key", 3, 4, true))

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
