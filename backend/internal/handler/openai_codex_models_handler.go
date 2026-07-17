package handler

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
)

// CodexModels serves the Codex models manifest for Codex clients.
//
// Codex CLI and the Codex desktop app refresh their model picker from
// GET {base_url}/models?client_version=... (custom provider mode) or
// GET /backend-api/codex/models (chatgpt_base_url mode). Both routes land
// here. The manifest is proxied verbatim from the selected account's ChatGPT
// backend or custom API key upstream. API key manifests use a short-lived,
// asynchronously revalidated cache to tolerate canceled client requests.
func (h *OpenAIGatewayHandler) CodexModels(c *gin.Context) {
	if c.Request.Context().Err() != nil {
		return
	}
	apiKey, ok := middleware2.GetAPIKeyFromContext(c)
	if !ok || apiKey.Group == nil {
		h.errorResponse(c, http.StatusUnauthorized, "invalid_request_error", "API key group is required")
		return
	}
	if apiKey.Group.Platform != service.PlatformOpenAI {
		h.errorResponse(c, http.StatusNotFound, "not_found_error", "Codex models manifest is only available for OpenAI groups")
		return
	}
	filterModels := apiKey.Group.CustomModelsListEnabled()
	ifNoneMatch := c.GetHeader("If-None-Match")
	if filterModels {
		ifNoneMatch = ""
	}

	maxAccountSwitches := h.maxAccountSwitches
	if maxAccountSwitches <= 0 {
		maxAccountSwitches = 3
	}
	failedAccountIDs := make(map[int64]struct{})
	switchCount := 0
	var lastUpstreamErr error

	for {
		account, err := h.gatewayService.SelectAccountForModelWithExclusions(c.Request.Context(), apiKey.GroupID, "", "", failedAccountIDs)
		if err != nil {
			if c.Request.Context().Err() != nil {
				return
			}
			if lastUpstreamErr != nil {
				h.errorResponse(c, infraerrors.Code(lastUpstreamErr), "upstream_error", infraerrors.Message(lastUpstreamErr))
				return
			}
			h.errorResponse(c, http.StatusServiceUnavailable, "upstream_error", "No available OpenAI accounts")
			return
		}

		manifest, err := h.gatewayService.FetchCodexModelsManifest(c.Request.Context(), account, c.Query("client_version"), ifNoneMatch)
		if err != nil {
			if c.Request.Context().Err() != nil {
				return
			}
			if service.IsRetryableCodexModelsManifestError(err) && switchCount < maxAccountSwitches {
				failedAccountIDs[account.ID] = struct{}{}
				switchCount++
				lastUpstreamErr = err
				continue
			}
			h.errorResponse(c, infraerrors.Code(err), "upstream_error", infraerrors.Message(err))
			return
		}
		if c.Request.Context().Err() != nil {
			return
		}

		if !filterModels && manifest.ETag != "" {
			c.Header("ETag", manifest.ETag)
		}
		if !filterModels && manifest.NotModified {
			c.Status(http.StatusNotModified)
			return
		}
		body := manifest.Body
		if filterModels {
			body, err = filterCodexModelsManifest(body, apiKey.Group)
			if err != nil {
				h.errorResponse(c, http.StatusBadGateway, "upstream_error", "Invalid Codex models manifest")
				return
			}
		}
		c.Data(http.StatusOK, "application/json", body)
		return
	}
}

func filterCodexModelsManifest(body []byte, group *service.Group) ([]byte, error) {
	var manifest map[string]json.RawMessage
	if err := json.Unmarshal(body, &manifest); err != nil {
		return nil, err
	}

	var models []json.RawMessage
	if err := json.Unmarshal(manifest["models"], &models); err != nil {
		return nil, err
	}
	filtered := make([]json.RawMessage, 0, len(models))
	for _, raw := range models {
		var model struct {
			Slug string `json:"slug"`
		}
		if json.Unmarshal(raw, &model) == nil && group.AllowsOpenAIModel(model.Slug) {
			filtered = append(filtered, raw)
		}
	}
	manifest["models"], _ = json.Marshal(filtered)
	return json.Marshal(manifest)
}
