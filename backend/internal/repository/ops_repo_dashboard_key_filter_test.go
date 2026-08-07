package repository

import (
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestBuildUsageWhere_FiltersAPIKeys(t *testing.T) {
	_, where, args, _ := buildUsageWhere(
		&service.OpsDashboardFilter{APIKeyIDs: []int64{2, 8}},
		time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2026, 1, 1, 1, 0, 0, 0, time.UTC),
		1,
	)

	require.Contains(t, where, "ul.api_key_id = ANY($3)")
	require.Len(t, args, 3)
}
