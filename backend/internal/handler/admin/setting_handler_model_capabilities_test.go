package admin

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

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

func callModelCapabilitiesHandler(t *testing.T, h *SettingHandler, method, path, body string, handler gin.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(method, path, strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	handler(c)
	return rec
}

// Scenario: 后台读写模型能力声明，非法思考等级被拒绝。
func TestModelCapabilitiesHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := &SettingHandler{settingService: service.NewSettingService(&modelCapabilitiesSettingRepoStub{values: map[string]string{}}, nil)}

	empty := callModelCapabilitiesHandler(t, h, http.MethodGet, "/api/v1/admin/settings/model-capabilities", "", h.GetModelCapabilities)
	require.Equal(t, http.StatusOK, empty.Code)
	require.Contains(t, empty.Body.String(), `"source":"config"`)

	saved := callModelCapabilitiesHandler(t, h, http.MethodPut, "/api/v1/admin/settings/model-capabilities",
		`{"models":[{"id":"gpt-5.6-sol","context_length":922000,"reasoning_levels":["low","high"],"default_reasoning_level":"high"}]}`,
		h.UpdateModelCapabilities)
	require.Equal(t, http.StatusOK, saved.Code)
	require.Contains(t, saved.Body.String(), `"source":"settings"`)
	require.Contains(t, saved.Body.String(), `"context_length":922000`)

	rejected := callModelCapabilitiesHandler(t, h, http.MethodPut, "/api/v1/admin/settings/model-capabilities",
		`{"models":[{"id":"gpt-5.6-sol","reasoning_levels":["bogus"]}]}`,
		h.UpdateModelCapabilities)
	require.Equal(t, http.StatusBadRequest, rejected.Code)
}
