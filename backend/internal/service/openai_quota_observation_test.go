package service

import "testing"

func TestBuildCodexQuotaObservationHistoryKeepsOnlyPercentTransitions(t *testing.T) {
	first, changed := buildCodexQuotaObservationHistory(nil, map[string]any{
		"codex_usage_updated_at": "2026-08-15T10:00:00Z",
		"codex_7d_used_percent":  42.5,
		"codex_7d_reset_at":      "2026-08-20T03:00:00Z",
	})
	if !changed || len(first) != 1 {
		t.Fatalf("first observation = %#v, changed=%v", first, changed)
	}

	unchanged, changed := buildCodexQuotaObservationHistory(first, map[string]any{
		"codex_usage_updated_at": "2026-08-15T10:01:00Z",
		"codex_7d_used_percent":  42.5,
		"codex_7d_reset_at":      "2026-08-20T03:00:00Z",
	})
	if changed || len(unchanged) != 1 {
		t.Fatalf("unchanged observation appended: %#v", unchanged)
	}

	second, changed := buildCodexQuotaObservationHistory(first, map[string]any{
		"codex_usage_updated_at": "2026-08-15T10:02:00Z",
		"codex_7d_used_percent":  43.0,
		"codex_7d_reset_at":      "2026-08-20T03:00:00Z",
	})
	if !changed || len(second) != 2 || *second[1].Used7dPercent != 43.0 {
		t.Fatalf("second observation = %#v, changed=%v", second, changed)
	}
}
