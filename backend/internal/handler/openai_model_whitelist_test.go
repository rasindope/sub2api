package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	coderws "github.com/coder/websocket"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
	"go.uber.org/zap"
)

func TestRejectDisallowedOpenAIModel(t *testing.T) {
	groupID := int64(42)
	apiKey := &service.APIKey{
		ID:      7,
		GroupID: &groupID,
		Group: &service.Group{
			ID:       groupID,
			Platform: service.PlatformOpenAI,
			ModelsListConfig: service.GroupModelsListConfig{
				Enabled: true,
				Models:  []string{"gpt-5.6-sol"},
			},
		},
	}
	h := &OpenAIGatewayHandler{}

	t.Run("allowed", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)

		require.False(t, h.rejectDisallowedOpenAIModel(c, zap.NewNop(), apiKey, "gpt-5.6-sol"))
	})

	t.Run("denied", func(t *testing.T) {
		rec := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rec)

		require.True(t, h.rejectDisallowedOpenAIModel(c, zap.NewNop(), apiKey, "gpt-5.5"))
		require.Equal(t, http.StatusForbidden, rec.Code)
		require.Equal(t, "permission_error", gjson.GetBytes(rec.Body.Bytes(), "error.type").String())
		require.Contains(t, rec.Body.String(), "gpt-5.5")
	})
}

func TestResponsesWebSocketRejectsDisallowedFirstModel(t *testing.T) {
	h := newOpenAIHandlerForPreviousResponseIDValidation(t, nil)
	groupID := int64(42)
	apiKey := &service.APIKey{
		ID:      7,
		GroupID: &groupID,
		User:    &service.User{ID: 8},
		Group: &service.Group{
			ID:       groupID,
			Platform: service.PlatformOpenAI,
			ModelsListConfig: service.GroupModelsListConfig{
				Enabled: true,
				Models:  []string{"gpt-5.6-sol"},
			},
		},
	}
	router := gin.New()
	router.Use(func(c *gin.Context) {
		c.Set(string(middleware2.ContextKeyAPIKey), apiKey)
		c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 8, Concurrency: 1})
		c.Next()
	})
	router.GET("/openai/v1/responses", h.ResponsesWebSocket)
	server := httptest.NewServer(router)
	defer server.Close()

	dialCtx, cancelDial := context.WithTimeout(context.Background(), 3*time.Second)
	conn, _, err := coderws.Dial(dialCtx, "ws"+strings.TrimPrefix(server.URL, "http")+"/openai/v1/responses", nil)
	cancelDial()
	require.NoError(t, err)
	defer func() { _ = conn.CloseNow() }()

	writeCtx, cancelWrite := context.WithTimeout(context.Background(), 3*time.Second)
	err = conn.Write(writeCtx, coderws.MessageText, []byte(`{"type":"response.create","model":"gpt-5.5"}`))
	cancelWrite()
	require.NoError(t, err)

	readCtx, cancelRead := context.WithTimeout(context.Background(), 3*time.Second)
	_, _, err = conn.Read(readCtx)
	cancelRead()
	var closeErr coderws.CloseError
	require.ErrorAs(t, err, &closeErr)
	require.Equal(t, coderws.StatusPolicyViolation, closeErr.Code)
	require.Contains(t, closeErr.Reason, "gpt-5.5")
}

func TestOpenAIHandlersRejectDisallowedModelBeforeScheduling(t *testing.T) {
	tests := []struct {
		name string
		path string
		body string
		run  func(*OpenAIGatewayHandler, *gin.Context)
	}{
		{name: "responses", path: "/v1/responses", body: `{"model":"gpt-5.5","input":"hi"}`, run: func(h *OpenAIGatewayHandler, c *gin.Context) { h.Responses(c) }},
		{name: "chat completions", path: "/v1/chat/completions", body: `{"model":"gpt-5.5","messages":[{"role":"user","content":"hi"}]}`, run: func(h *OpenAIGatewayHandler, c *gin.Context) { h.ChatCompletions(c) }},
		{name: "embeddings", path: "/v1/embeddings", body: `{"model":"gpt-5.5","input":"hi"}`, run: func(h *OpenAIGatewayHandler, c *gin.Context) { h.Embeddings(c) }},
		{name: "images", path: "/v1/images/generations", body: `{"model":"gpt-image-2","prompt":"hi"}`, run: func(h *OpenAIGatewayHandler, c *gin.Context) { h.Images(c) }},
		{name: "alpha search", path: "/backend-api/codex/alpha/search", body: `{"model":"gpt-5.5","query":"hi"}`, run: func(h *OpenAIGatewayHandler, c *gin.Context) { h.AlphaSearch(c) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")
			groupID := int64(42)
			c.Set(string(middleware2.ContextKeyAPIKey), &service.APIKey{
				ID:      7,
				GroupID: &groupID,
				Group: &service.Group{
					ID:                   groupID,
					Platform:             service.PlatformOpenAI,
					AllowImageGeneration: true,
					ModelsListConfig: service.GroupModelsListConfig{
						Enabled: true,
						Models:  []string{"gpt-5.6-sol"},
					},
				},
				User: &service.User{ID: 8},
			})
			c.Set(string(middleware2.ContextKeyUser), middleware2.AuthSubject{UserID: 8, Concurrency: 1})
			h := &OpenAIGatewayHandler{
				gatewayService:      &service.OpenAIGatewayService{},
				billingCacheService: &service.BillingCacheService{},
				apiKeyService:       &service.APIKeyService{},
				concurrencyHelper:   &ConcurrencyHelper{concurrencyService: &service.ConcurrencyService{}},
			}

			tt.run(h, c)

			require.Equal(t, http.StatusForbidden, rec.Code, rec.Body.String())
			require.Equal(t, "permission_error", gjson.GetBytes(rec.Body.Bytes(), "error.type").String())
		})
	}
}
