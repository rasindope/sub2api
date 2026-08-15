package service

import (
	"context"
	"encoding/json"
	"fmt"
)

const (
	codexQuotaObservationsExtraKey = "codex_quota_observations"
	codexQuotaObservationLimit     = 1024
)

type codexQuotaObservation struct {
	ObservedAt    string   `json:"observed_at"`
	Used5hPercent *float64 `json:"used_5h_percent,omitempty"`
	Used7dPercent *float64 `json:"used_7d_percent,omitempty"`
	Reset5hAt     string   `json:"reset_5h_at,omitempty"`
	Reset7dAt     string   `json:"reset_7d_at,omitempty"`
}

func buildCodexQuotaObservationHistory(raw any, updates map[string]any) ([]codexQuotaObservation, bool) {
	used7d := numberPointer(updates["codex_7d_used_percent"])
	if used7d == nil {
		return nil, false
	}

	history := make([]codexQuotaObservation, 0)
	if raw != nil {
		if encoded, err := json.Marshal(raw); err == nil {
			_ = json.Unmarshal(encoded, &history)
		}
	}
	observation := codexQuotaObservation{
		ObservedAt:    quotaStringValue(updates["codex_usage_updated_at"]),
		Used5hPercent: numberPointer(updates["codex_5h_used_percent"]),
		Used7dPercent: used7d,
		Reset5hAt:     quotaStringValue(updates["codex_5h_reset_at"]),
		Reset7dAt:     quotaStringValue(updates["codex_7d_reset_at"]),
	}
	if observation.ObservedAt == "" {
		return nil, false
	}
	if len(history) > 0 {
		last := history[len(history)-1]
		if equalFloatPointers(last.Used7dPercent, observation.Used7dPercent) && last.Reset7dAt == observation.Reset7dAt {
			return history, false
		}
	}
	history = append(history, observation)
	if len(history) > codexQuotaObservationLimit {
		history = history[len(history)-codexQuotaObservationLimit:]
	}
	return history, true
}

func persistCodexQuotaObservation(ctx context.Context, repo AccountRepository, accountID int64, updates map[string]any) {
	if repo == nil || accountID <= 0 {
		return
	}
	account, err := repo.GetByID(ctx, accountID)
	if err != nil || account == nil {
		return
	}
	history, changed := buildCodexQuotaObservationHistory(account.Extra[codexQuotaObservationsExtraKey], updates)
	if !changed {
		return
	}
	// ponytail: one writer per account is expected; switch to an atomic JSON append if multi-instance contention becomes measurable.
	_ = repo.UpdateExtra(ctx, accountID, map[string]any{codexQuotaObservationsExtraKey: history})
}

func numberPointer(value any) *float64 {
	switch typed := value.(type) {
	case float64:
		return &typed
	case float32:
		converted := float64(typed)
		return &converted
	case int:
		converted := float64(typed)
		return &converted
	case int64:
		converted := float64(typed)
		return &converted
	default:
		return nil
	}
}

func quotaStringValue(value any) string {
	if value == nil {
		return ""
	}
	return fmt.Sprint(value)
}

func equalFloatPointers(left, right *float64) bool {
	return left != nil && right != nil && *left == *right
}
