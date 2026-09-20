package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type modelCapabilitiesRequest struct {
	Models []service.ModelCapability `json:"models"`
}

// GetModelCapabilities 返回 /v1/models 当前生效的模型能力声明。
// source=settings 表示后台保存的覆盖，source=config 表示回落到 config.yaml。
// GET /api/v1/admin/settings/model-capabilities
func (h *SettingHandler) GetModelCapabilities(c *gin.Context) {
	response.Success(c, h.modelCapabilitiesPayload(c))
}

// UpdateModelCapabilities 整体替换后台保存的模型能力声明（空数组 = 清空覆盖）。
// PUT /api/v1/admin/settings/model-capabilities
func (h *SettingHandler) UpdateModelCapabilities(c *gin.Context) {
	var req modelCapabilitiesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	ctx := c.Request.Context()
	if err := h.settingService.SetModelCapabilities(ctx, req.Models); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	response.Success(c, h.modelCapabilitiesPayload(c))
}

func (h *SettingHandler) modelCapabilitiesPayload(c *gin.Context) gin.H {
	ctx := c.Request.Context()
	source := "config"
	if len(h.settingService.StoredModelCapabilities(ctx)) > 0 {
		source = "settings"
	}
	return gin.H{
		"models":           h.settingService.EffectiveModelCapabilities(ctx),
		"source":           source,
		"reasoning_levels": service.ModelCapabilityReasoningLevels,
	}
}
