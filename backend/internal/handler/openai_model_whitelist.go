package handler

import (
	"fmt"
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func openAIModelNotAllowedMessage(model string) string {
	return fmt.Sprintf("Model %q is not allowed for this API key's group", model)
}

func logDisallowedOpenAIModel(reqLog *zap.Logger, apiKey *service.APIKey, model string) {
	reqLog.Warn("openai.model_access_denied",
		zap.Int64("api_key_id", apiKey.ID),
		zap.Int64("group_id", apiKey.Group.ID),
		zap.String("requested_model", model),
	)
}

func (h *OpenAIGatewayHandler) rejectDisallowedOpenAIModel(c *gin.Context, reqLog *zap.Logger, apiKey *service.APIKey, model string) bool {
	if apiKey.Group.AllowsOpenAIModel(model) {
		return false
	}

	logDisallowedOpenAIModel(reqLog, apiKey, model)
	h.errorResponse(c, http.StatusForbidden, "permission_error", openAIModelNotAllowedMessage(model))
	return true
}
