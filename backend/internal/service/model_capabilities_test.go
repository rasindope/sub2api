package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/stretchr/testify/require"
)

type modelCapabilitiesRepoStub struct {
	values map[string]string
}

func (s *modelCapabilitiesRepoStub) Get(context.Context, string) (*Setting, error) {
	panic("unexpected Get call")
}

func (s *modelCapabilitiesRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", errors.New("setting not found")
}

func (s *modelCapabilitiesRepoStub) Set(_ context.Context, key, value string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[key] = value
	return nil
}

func (s *modelCapabilitiesRepoStub) GetMultiple(context.Context, []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *modelCapabilitiesRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *modelCapabilitiesRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *modelCapabilitiesRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

func TestNormalizeModelCapabilities(t *testing.T) {
	normalized, err := NormalizeModelCapabilities([]ModelCapability{
		{ID: " gpt-5.6-sol ", ContextLength: 922_000, ReasoningLevels: []string{"low", "HIGH", "low"}, DefaultReasoningLevel: "HIGH"},
		{ID: "gpt-5.6-sol", ContextLength: 1},
		{ID: "deepseek-v4-pro", ContextLength: 1_000_000},
	})
	require.NoError(t, err)
	require.Equal(t, []ModelCapability{
		{ID: "gpt-5.6-sol", ContextLength: 922_000, ReasoningLevels: []string{"low", "high"}, DefaultReasoningLevel: "high"},
		{ID: "deepseek-v4-pro", ContextLength: 1_000_000, ReasoningLevels: []string{}},
	}, normalized)

	_, err = NormalizeModelCapabilities([]ModelCapability{{ID: "m", ReasoningLevels: []string{"bogus"}}})
	require.Error(t, err)

	_, err = NormalizeModelCapabilities([]ModelCapability{{ID: "m", ReasoningLevels: []string{"low"}, DefaultReasoningLevel: "high"}})
	require.Error(t, err)

	_, err = NormalizeModelCapabilities([]ModelCapability{{ID: "  "}})
	require.Error(t, err)

	_, err = NormalizeModelCapabilities([]ModelCapability{{ID: "m", ContextLength: -1}})
	require.Error(t, err)
}

// Scenario: 后台保存的设置优先，未保存或内容损坏时回落到 config.yaml。
func TestSettingServiceModelCapabilities(t *testing.T) {
	repo := &modelCapabilitiesRepoStub{values: map[string]string{}}
	cfg := &config.Config{ModelCapabilities: config.ModelCapabilitiesConfig{Models: []config.ModelCapabilityConfig{
		{ID: "from-config", ContextLength: 1000},
	}}}
	svc := NewSettingService(repo, cfg)
	ctx := context.Background()
	fromConfig := []ModelCapability{{ID: "from-config", ContextLength: 1000}}

	require.Empty(t, svc.StoredModelCapabilities(ctx))
	require.Equal(t, fromConfig, svc.EffectiveModelCapabilities(ctx))

	require.NoError(t, svc.SetModelCapabilities(ctx, []ModelCapability{{
		ID: "from-page", ContextLength: 2000, ReasoningLevels: []string{"low", "high"}, DefaultReasoningLevel: "high",
	}}))
	require.Equal(t, []ModelCapability{{
		ID: "from-page", ContextLength: 2000, ReasoningLevels: []string{"low", "high"}, DefaultReasoningLevel: "high",
	}}, svc.EffectiveModelCapabilities(ctx))

	repo.values[SettingKeyModelCapabilities] = "{not json"
	require.Equal(t, fromConfig, svc.EffectiveModelCapabilities(ctx))

	require.NoError(t, svc.SetModelCapabilities(ctx, nil))
	require.Equal(t, fromConfig, svc.EffectiveModelCapabilities(ctx))
}

// Scenario: 没有 config.yaml 声明时不产生任何字段。
func TestModelCapabilitiesFromConfigEmpty(t *testing.T) {
	require.Empty(t, ModelCapabilitiesFromConfig(nil))
	require.Empty(t, ModelCapabilitiesFromConfig(&config.Config{}))
	require.Empty(t, (*SettingService)(nil).EffectiveModelCapabilities(context.Background()))
}
