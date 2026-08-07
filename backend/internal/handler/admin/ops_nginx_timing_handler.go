package admin

import (
	"net/http"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

// GetNginxTimingOverview returns client-facing Nginx timing data for the Ops dashboard.
// GET /api/v1/admin/ops/dashboard/nginx-timing
func (h *OpsHandler) GetNginxTimingOverview(c *gin.Context) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	if err := h.opsService.RequireMonitoringEnabled(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	filter, err := parseNginxTimingFilter(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	data, err := h.opsService.GetNginxTimingOverview(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, data)
}

// GetNginxTimingKeyDetails returns the per-Key metrics for one Nginx card.
// GET /api/v1/admin/ops/dashboard/nginx-timing/keys
func (h *OpsHandler) GetNginxTimingKeyDetails(c *gin.Context) {
	if h.opsService == nil {
		response.Error(c, http.StatusServiceUnavailable, "Ops service not available")
		return
	}
	if err := h.opsService.RequireMonitoringEnabled(c.Request.Context()); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	filter, err := parseNginxTimingFilter(c)
	if err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	data, err := h.opsService.GetNginxTimingKeyDetails(c.Request.Context(), filter)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, data)
}

func parseNginxTimingFilter(c *gin.Context) (*service.OpsNginxTimingFilter, error) {
	startTime, endTime, err := parseOpsTimeRange(c, "1h")
	if err != nil {
		return nil, err
	}
	apiKeyIDs, err := parseOpsAPIKeyIDs(c)
	if err != nil {
		return nil, err
	}
	return &service.OpsNginxTimingFilter{
		StartTime: startTime,
		EndTime:   endTime,
		APIKeyIDs: apiKeyIDs,
	}, nil
}
