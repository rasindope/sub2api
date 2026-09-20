package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

// Scenario: model_capabilities 段落从 YAML 载入后可按模型 ID 查到能力声明。
func TestLoadModelCapabilities(t *testing.T) {
	resetViperWithJWTSecret(t)

	configPath := filepath.Join(t.TempDir(), "config.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(`
model_capabilities:
  models:
    - id: gpt-5.6-sol
      context_length: 400000
      reasoning_levels: [low, medium, high, xhigh]
      default_reasoning_level: high
    - id: claude-fable-5-1
      context_length: 1000000
`), 0o600))
	t.Setenv("CONFIG_FILE", configPath)

	cfg, err := Load()
	require.NoError(t, err)
	require.Len(t, cfg.ModelCapabilities.Models, 2)

	capability, ok := cfg.ModelCapability("gpt-5.6-sol")
	require.True(t, ok)
	require.Equal(t, int64(400_000), capability.ContextLength)
	require.Equal(t, []string{"low", "medium", "high", "xhigh"}, capability.ReasoningLevels)
	require.Equal(t, "high", capability.DefaultReasoningLevel)

	_, ok = cfg.ModelCapability("unlisted-model")
	require.False(t, ok)

	var nilConfig *Config
	_, ok = nilConfig.ModelCapability("gpt-5.6-sol")
	require.False(t, ok)
}
