package service

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/pkg/tlsfingerprint"
	"github.com/stretchr/testify/require"
)

type codexInviteResetAdminServiceStub struct {
	AdminService
	account *Account
	proxy   *Proxy
}

func (s codexInviteResetAdminServiceStub) GetAccount(context.Context, int64) (*Account, error) {
	return s.account, nil
}

func (s codexInviteResetAdminServiceStub) GetProxy(context.Context, int64) (*Proxy, error) {
	return s.proxy, nil
}

type codexInviteResetHTTPUpstreamStub struct {
	responses []*http.Response
	requests  []*http.Request
	bodies    []string
}

func (s *codexInviteResetHTTPUpstreamStub) Do(req *http.Request, proxyURL string, accountID int64, accountConcurrency int) (*http.Response, error) {
	return s.DoWithTLS(req, proxyURL, accountID, accountConcurrency, nil)
}

func (s *codexInviteResetHTTPUpstreamStub) DoWithTLS(req *http.Request, _ string, _ int64, _ int, _ *tlsfingerprint.Profile) (*http.Response, error) {
	body := ""
	if req.Body != nil {
		payload, _ := io.ReadAll(req.Body)
		body = string(payload)
		req.Body = io.NopCloser(strings.NewReader(body))
	}
	s.requests = append(s.requests, req)
	s.bodies = append(s.bodies, body)
	if len(s.responses) == 0 {
		return codexInviteResetJSONResponse(`{}`), nil
	}
	resp := s.responses[0]
	s.responses = s.responses[1:]
	return resp, nil
}

func TestCodexInviteResetServiceGetStatusCallsDesktopEndpoints(t *testing.T) {
	account := &Account{
		ID:          42,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Concurrency: 3,
		Credentials: map[string]any{
			"access_token":       "oauth-token",
			"chatgpt_account_id": "chatgpt-acc",
		},
	}
	upstream := &codexInviteResetHTTPUpstreamStub{responses: []*http.Response{
		codexInviteResetJSONResponse(`{"requires_explicit_confirmation":true}`),
		codexInviteResetJSONResponse(`{"rules":[{"text":"friend must send first Codex message"}]}`),
	}}
	svc := NewCodexInviteResetService(codexInviteResetAdminServiceStub{account: account}, upstream, nil, nil)

	status, err := svc.GetStatus(context.Background(), account.ID)
	require.NoError(t, err)
	require.Equal(t, "friend must send first Codex message", status.EligibilityRules[0])
	require.Len(t, upstream.requests, 2)
	require.Equal(t, "/backend-api/referrals/invite/eligibility", upstream.requests[0].URL.Path)
	require.Equal(t, codexInviteResetReferralKey, upstream.requests[0].URL.Query().Get("referral_key"))
	require.Equal(t, "/backend-api/wham/referrals/eligibility_rules", upstream.requests[1].URL.Path)
	require.Equal(t, "Bearer oauth-token", upstream.requests[0].Header.Get("Authorization"))
	require.Equal(t, "Codex Desktop", upstream.requests[0].Header.Get("originator"))
	require.Equal(t, codexInviteResetDefaultUserAgent, upstream.requests[0].Header.Get("User-Agent"))
	require.Equal(t, "chatgpt-acc", upstream.requests[0].Header.Get("chatgpt-account-id"))
}

func TestCodexInviteResetServiceSendInviteNormalizesEmails(t *testing.T) {
	account := &Account{
		ID:          7,
		Platform:    PlatformOpenAI,
		Type:        AccountTypeOAuth,
		Credentials: map[string]any{"access_token": "oauth-token"},
	}
	upstream := &codexInviteResetHTTPUpstreamStub{responses: []*http.Response{
		codexInviteResetJSONResponse(`{"invites":[{"email":"a@example.com"}],"message":"ok"}`),
	}}
	svc := NewCodexInviteResetService(codexInviteResetAdminServiceStub{account: account}, upstream, nil, nil)

	result, err := svc.SendInvite(context.Background(), account.ID, []string{"a@example.com, b@example.com", "A@example.com"})
	require.NoError(t, err)
	require.Equal(t, "ok", result.Message)
	require.Equal(t, "/backend-api/wham/referrals/invite", upstream.requests[0].URL.Path)

	var payload map[string]any
	require.NoError(t, json.Unmarshal([]byte(upstream.bodies[0]), &payload))
	require.Equal(t, codexInviteResetReferralKey, payload["referral_key"])
	require.Equal(t, []any{"a@example.com", "b@example.com"}, payload["emails"])
}

func codexInviteResetJSONResponse(body string) *http.Response {
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}
