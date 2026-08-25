//go:build unit

package service

import (
	"context"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

func TestAPIKeyIPRiskSettingsRoundTripAndValidation(t *testing.T) {
	repo := newMockSettingRepo()
	svc := NewSettingService(repo, &config.Config{})

	defaults, err := svc.GetAPIKeyIPRiskSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, DefaultAPIKeyIPRiskSettings(), defaults)

	want := APIKeyIPRiskSettings{MinimumOverlapSeconds: 8, HighSingleOverlapSecs: 90, HighOverlapCount: 12, HighTotalOverlapSecs: 180}
	require.NoError(t, svc.SetAPIKeyIPRiskSettings(context.Background(), want))
	got, err := svc.GetAPIKeyIPRiskSettings(context.Background())
	require.NoError(t, err)
	require.Equal(t, want, got)

	want.HighSingleOverlapSecs = 4
	require.Error(t, svc.SetAPIKeyIPRiskSettings(context.Background(), want))
}
