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

func TestBuildCodexQuotaObservationHistoryCompactsExistingDuplicates(t *testing.T) {
	percent := 42.5
	history, changed := buildCodexQuotaObservationHistory([]codexQuotaObservation{
		{ObservedAt: "2026-08-15T10:00:00Z", Used7dPercent: &percent, Reset7dAt: "2026-08-20T03:00:00Z"},
		{ObservedAt: "2026-08-15T10:01:00Z", Used7dPercent: &percent, Reset7dAt: "2026-08-20T03:00:00Z"},
	}, map[string]any{
		"codex_usage_updated_at": "2026-08-15T10:02:00Z",
		"codex_7d_used_percent":  percent,
		"codex_7d_reset_at":      "2026-08-20T03:00:00Z",
	})

	if !changed || len(history) != 1 || history[0].ObservedAt != "2026-08-15T10:00:00Z" {
		t.Fatalf("duplicates not compacted: %#v, changed=%v", history, changed)
	}
}
