package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

type modelCapabilityItemForTest struct {
	ID                    string   `json:"id"`
	ContextLength         *int64   `json:"context_length"`
	ReasoningLevels       []string `json:"reasoning_levels"`
	DefaultReasoningLevel *string  `json:"default_reasoning_level"`
}

func modelCapabilitiesResponseForTest(t *testing.T, h *GatewayHandler, platform string, groupID int64) map[string]modelCapabilityItemForTest {
	t.Helper()

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
		Group: &service.Group{ID: groupID, Platform: platform},
	})

	h.Models(c)
	require.Equal(t, http.StatusOK, rec.Code)

	var got struct {
		Object string                       `json:"object"`
		Data   []modelCapabilityItemForTest `json:"data"`
	}
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &got))
	require.Equal(t, "list", got.Object)

	byID := make(map[string]modelCapabilityItemForTest, len(got.Data))
	for _, item := range got.Data {
		byID[item.ID] = item
	}
	return byID
}

func modelCapabilityAccountsForTest(groupID int64, platform string, modelIDs ...string) map[int64][]service.Account {
	mapping := make(map[string]any, len(modelIDs))
	for _, modelID := range modelIDs {
		mapping[modelID] = modelID
	}
	return map[int64][]service.Account{
		groupID: {{ID: 1, Platform: platform, Credentials: map[string]any{"model_mapping": mapping}}},
	}
}

// Scenario: model_capabilities 声明的上下文窗口与思考等级会并入 /v1/models，
// 未声明的模型以及未配置该模块的部署保持原响应。
func TestGatewayModels_ModelCapabilitiesAreExposed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const groupID int64 = 77
	cfg := &config.Config{ModelCapabilities: config.ModelCapabilitiesConfig{Models: []config.ModelCapabilityConfig{
		{
			ID:                    "gpt-future-model",
			ContextLength:         400_000,
			ReasoningLevels:       []string{"low", "medium", "high", "xhigh"},
			DefaultReasoningLevel: "high",
		},
		{
			ID:              "claude-fable-5-1",
			ContextLength:   1_000_000,
			ReasoningLevels: []string{"none", "high"},
		},
	}}}

	t.Run("openai shaped list", func(t *testing.T) {
		h := newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{
			byGroup: modelCapabilityAccountsForTest(groupID, service.PlatformOpenAI, "gpt-future-model", "gpt-5.6-sol"),
		})
		h.cfg = cfg

		byID := modelCapabilitiesResponseForTest(t, h, service.PlatformOpenAI, groupID)

		declared := byID["gpt-future-model"]
		require.NotNil(t, declared.ContextLength)
		require.Equal(t, int64(400_000), *declared.ContextLength)
		require.Equal(t, []string{"low", "medium", "high", "xhigh"}, declared.ReasoningLevels)
		require.NotNil(t, declared.DefaultReasoningLevel)
		require.Equal(t, "high", *declared.DefaultReasoningLevel)

		require.NotNil(t, byID["gpt-5.6-sol"])
		require.Nil(t, byID["gpt-5.6-sol"].ContextLength)
		require.Empty(t, byID["gpt-5.6-sol"].ReasoningLevels)
	})

	t.Run("claude shaped list", func(t *testing.T) {
		h := newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{
			byGroup: modelCapabilityAccountsForTest(groupID, service.PlatformAnthropic, "claude-fable-5-1"),
		})
		h.cfg = cfg

		byID := modelCapabilitiesResponseForTest(t, h, service.PlatformAnthropic, groupID)

		declared := byID["claude-fable-5-1"]
		require.NotNil(t, declared.ContextLength)
		require.Equal(t, int64(1_000_000), *declared.ContextLength)
		require.Equal(t, []string{"none", "high"}, declared.ReasoningLevels)
		require.Nil(t, declared.DefaultReasoningLevel)
	})

	t.Run("without configuration", func(t *testing.T) {
		h := newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{
			byGroup: modelCapabilityAccountsForTest(groupID, service.PlatformOpenAI, "gpt-future-model"),
		})

		item := modelCapabilitiesResponseForTest(t, h, service.PlatformOpenAI, groupID)["gpt-future-model"]
		require.Nil(t, item.ContextLength)
		require.Empty(t, item.ReasoningLevels)
		require.Nil(t, item.DefaultReasoningLevel)
	})
}

// Scenario: 未映射的 OpenAI 分组走默认模型列表时同样叠加配置的能力字段。
func TestGatewayModels_ModelCapabilitiesApplyToDefaultFallback(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const groupID int64 = 78
	h := newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{
		byGroup: map[int64][]service.Account{groupID: {{ID: 1, Platform: service.PlatformOpenAI}}},
	})
	h.cfg = &config.Config{ModelCapabilities: config.ModelCapabilitiesConfig{Models: []config.ModelCapabilityConfig{{
		ID:              "gpt-5.6-sol",
		ContextLength:   1_000_000,
		ReasoningLevels: []string{"none", "low", "medium", "high", "xhigh"},
	}}}}

	byID := modelCapabilitiesResponseForTest(t, h, service.PlatformOpenAI, groupID)
	require.Len(t, byID, len(defaultModelIDsForPlatform(service.PlatformOpenAI)))

	item := byID["gpt-5.6-sol"]
	require.NotNil(t, item.ContextLength)
	require.Equal(t, int64(1_000_000), *item.ContextLength)
	require.Equal(t, []string{"none", "low", "medium", "high", "xhigh"}, item.ReasoningLevels)
}

type modelCapabilitiesSettingRepoStub struct {
	values map[string]string
}

func (s *modelCapabilitiesSettingRepoStub) Get(context.Context, string) (*service.Setting, error) {
	panic("unexpected Get call")
}

func (s *modelCapabilitiesSettingRepoStub) GetValue(_ context.Context, key string) (string, error) {
	if value, ok := s.values[key]; ok {
		return value, nil
	}
	return "", errors.New("setting not found")
}

func (s *modelCapabilitiesSettingRepoStub) Set(_ context.Context, key, value string) error {
	if s.values == nil {
		s.values = map[string]string{}
	}
	s.values[key] = value
	return nil
}

func (s *modelCapabilitiesSettingRepoStub) GetMultiple(context.Context, []string) (map[string]string, error) {
	panic("unexpected GetMultiple call")
}

func (s *modelCapabilitiesSettingRepoStub) SetMultiple(context.Context, map[string]string) error {
	panic("unexpected SetMultiple call")
}

func (s *modelCapabilitiesSettingRepoStub) GetAll(context.Context) (map[string]string, error) {
	panic("unexpected GetAll call")
}

func (s *modelCapabilitiesSettingRepoStub) Delete(context.Context, string) error {
	panic("unexpected Delete call")
}

// Scenario: 后台保存的能力声明覆盖 config.yaml，/v1/models 立即生效。
func TestGatewayModels_ModelCapabilitiesUseStoredOverride(t *testing.T) {
	gin.SetMode(gin.TestMode)

	const groupID int64 = 79
	h := newGatewayModelsHandlerForTest(&gatewayModelsAccountRepoStub{
		byGroup: modelCapabilityAccountsForTest(groupID, service.PlatformOpenAI, "gpt-future-model"),
	})
	h.cfg = &config.Config{ModelCapabilities: config.ModelCapabilitiesConfig{Models: []config.ModelCapabilityConfig{{
		ID: "gpt-future-model", ContextLength: 111,
	}}}}
	h.settingService = service.NewSettingService(&modelCapabilitiesSettingRepoStub{values: map[string]string{}}, h.cfg)
	require.NoError(t, h.settingService.SetModelCapabilities(context.Background(), []service.ModelCapability{{
		ID: "gpt-future-model", ContextLength: 222, ReasoningLevels: []string{"low", "high"}, DefaultReasoningLevel: "high",
	}}))

	item := modelCapabilitiesResponseForTest(t, h, service.PlatformOpenAI, groupID)["gpt-future-model"]
	require.NotNil(t, item.ContextLength)
	require.Equal(t, int64(222), *item.ContextLength)
	require.Equal(t, []string{"low", "high"}, item.ReasoningLevels)
	require.NotNil(t, item.DefaultReasoningLevel)
	require.Equal(t, "high", *item.DefaultReasoningLevel)
}
