package service

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

// SettingKeyModelCapabilities 保存后台页面配置的模型能力声明（JSON 数组）。
// 未配置时 /v1/models 回落到 config.yaml 的 model_capabilities。
const SettingKeyModelCapabilities = "model_capabilities"

// ModelCapabilityReasoningLevels 是 Codex 认识的思考等级，顺序即页面展示顺序。
var ModelCapabilityReasoningLevels = []string{
	"none", "minimal", "low", "medium", "high", "xhigh", "max", "ultra",
}

// ModelCapability 是 /v1/models 对外暴露的单个模型能力声明。
type ModelCapability struct {
	ID                    string   `json:"id"`
	ContextLength         int64    `json:"context_length,omitempty"`
	ReasoningLevels       []string `json:"reasoning_levels,omitempty"`
	DefaultReasoningLevel string   `json:"default_reasoning_level,omitempty"`
}

// NormalizeModelCapabilities 校验并清洗声明：模型 ID 必填且唯一，思考等级必须是
// Codex 认识的取值，默认等级必须落在声明的等级里。
func NormalizeModelCapabilities(items []ModelCapability) ([]ModelCapability, error) {
	out := make([]ModelCapability, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		id := strings.TrimSpace(item.ID)
		if id == "" {
			return nil, fmt.Errorf("model id is required")
		}
		if _, exists := seen[id]; exists {
			continue
		}
		if item.ContextLength < 0 {
			return nil, fmt.Errorf("model %s: context_length cannot be negative", id)
		}

		levels := make([]string, 0, len(item.ReasoningLevels))
		for _, level := range item.ReasoningLevels {
			level = strings.ToLower(strings.TrimSpace(level))
			if level == "" {
				continue
			}
			if !slices.Contains(ModelCapabilityReasoningLevels, level) {
				return nil, fmt.Errorf("model %s: unknown reasoning level %q", id, level)
			}
			if slices.Contains(levels, level) {
				continue
			}
			levels = append(levels, level)
		}

		defaultLevel := strings.ToLower(strings.TrimSpace(item.DefaultReasoningLevel))
		if defaultLevel != "" && !slices.Contains(levels, defaultLevel) {
			return nil, fmt.Errorf("model %s: default_reasoning_level %q is not declared in reasoning_levels", id, defaultLevel)
		}

		seen[id] = struct{}{}
		out = append(out, ModelCapability{
			ID:                    id,
			ContextLength:         item.ContextLength,
			ReasoningLevels:       levels,
			DefaultReasoningLevel: defaultLevel,
		})
	}
	return out, nil
}

// ModelCapabilitiesFromConfig 返回 config.yaml 里声明的模型能力。
func ModelCapabilitiesFromConfig(cfg *config.Config) []ModelCapability {
	if cfg == nil || len(cfg.ModelCapabilities.Models) == 0 {
		return nil
	}
	out := make([]ModelCapability, 0, len(cfg.ModelCapabilities.Models))
	for _, item := range cfg.ModelCapabilities.Models {
		out = append(out, ModelCapability{
			ID:                    strings.TrimSpace(item.ID),
			ContextLength:         item.ContextLength,
			ReasoningLevels:       item.ReasoningLevels,
			DefaultReasoningLevel: strings.TrimSpace(item.DefaultReasoningLevel),
		})
	}
	return out
}

// StoredModelCapabilities 返回后台页面保存的声明；未保存或内容非法时返回 nil。
func (s *SettingService) StoredModelCapabilities(ctx context.Context) []ModelCapability {
	if s == nil || s.settingRepo == nil {
		return nil
	}
	raw, err := s.settingRepo.GetValue(ctx, SettingKeyModelCapabilities)
	if err != nil || strings.TrimSpace(raw) == "" {
		return nil
	}
	var stored []ModelCapability
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		return nil
	}
	normalized, err := NormalizeModelCapabilities(stored)
	if err != nil {
		return nil
	}
	return normalized
}

// EffectiveModelCapabilities 返回生效的声明：后台保存的设置优先，未配置时回落
// 到 config.yaml 的 model_capabilities。
func (s *SettingService) EffectiveModelCapabilities(ctx context.Context) []ModelCapability {
	if stored := s.StoredModelCapabilities(ctx); len(stored) > 0 {
		return stored
	}
	if s == nil {
		return nil
	}
	return ModelCapabilitiesFromConfig(s.cfg)
}

// SetModelCapabilities 整体替换后台保存的声明（空列表 = 清空覆盖，回落 config.yaml）。
func (s *SettingService) SetModelCapabilities(ctx context.Context, items []ModelCapability) error {
	if s == nil || s.settingRepo == nil {
		return fmt.Errorf("setting service is not configured")
	}
	normalized, err := NormalizeModelCapabilities(items)
	if err != nil {
		return err
	}
	payload, err := json.Marshal(normalized)
	if err != nil {
		return err
	}
	return s.settingRepo.Set(ctx, SettingKeyModelCapabilities, string(payload))
}
