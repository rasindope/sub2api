package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type CodexInviteResetHandler struct {
	service *service.CodexInviteResetService
}

func NewCodexInviteResetHandler(service *service.CodexInviteResetService) *CodexInviteResetHandler {
	return &CodexInviteResetHandler{service: service}
}

type codexInviteResetInviteRequest struct {
	Emails []string `json:"emails" binding:"required"`
}

func (h *CodexInviteResetHandler) GetStatus(c *gin.Context) {
	accountID, ok := parseCodexInviteResetAccountID(c)
	if !ok {
		return
	}
	result, err := h.service.GetStatus(c.Request.Context(), accountID)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, result)
}

func (h *CodexInviteResetHandler) SendInvite(c *gin.Context) {
	accountID, ok := parseCodexInviteResetAccountID(c)
	if !ok {
		return
	}
	var req codexInviteResetInviteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.service.SendInvite(c.Request.Context(), accountID, req.Emails)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, result)
}

func parseCodexInviteResetAccountID(c *gin.Context) (int64, bool) {
	accountID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.BadRequest(c, "Invalid account ID")
		return 0, false
	}
	return accountID, true
}
