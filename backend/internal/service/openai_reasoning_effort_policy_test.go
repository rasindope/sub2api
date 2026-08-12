package service

import (
	"context"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/pkg/timezone"
	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestNormalizeMaxReasoningEffort(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "empty", in: "", want: ""},
		{name: "separator", in: "x-high", want: "xhigh"},
		{name: "max is distinct", in: "max", want: "max"},
		{name: "none is unsupported", in: "none", want: ""},
		{name: "invalid", in: "banana", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, NormalizeMaxReasoningEffort(tt.in))
		})
	}
}

func TestNormalizeReasoningEffortMappings(t *testing.T) {
	t.Run("canonicalizes fixed OpenAI values", func(t *testing.T) {
		for _, platform := range []string{PlatformOpenAI, PlatformComposite} {
			got, err := NormalizeReasoningEffortMappings(platform, []ReasoningEffortMapping{
				{From: " MAX ", To: " x-high "},
				{From: "minimal", To: "high"},
			})
			require.NoError(t, err)
			require.Equal(t, []ReasoningEffortMapping{
				{From: "max", To: "xhigh"},
				{From: "minimal", To: "high"},
			}, got)
		}
	})

	t.Run("rejects empty values", func(t *testing.T) {
		_, err := NormalizeReasoningEffortMappings(PlatformOpenAI, []ReasoningEffortMapping{{From: "max"}})
		require.ErrorContains(t, err, "empty or unknown")
	})

	t.Run("rejects duplicate sources case insensitively", func(t *testing.T) {
		_, err := NormalizeReasoningEffortMappings(PlatformOpenAI, []ReasoningEffortMapping{
			{From: "max", To: "xhigh"},
			{From: " MAX ", To: "high"},
		})
		require.ErrorContains(t, err, "duplicate")
	})

	t.Run("rejects mappings for non OpenAI platforms", func(t *testing.T) {
		for _, platform := range []string{PlatformAnthropic, PlatformGemini, PlatformAntigravity, PlatformGrok} {
			_, err := NormalizeReasoningEffortMappings(platform, []ReasoningEffortMapping{{From: "low", To: "high"}})
			require.ErrorContains(t, err, "only supported for platforms \"openai\" and \"composite\"")
		}

		_, err := NormalizeReasoningEffortMappings(PlatformOpenAI, []ReasoningEffortMapping{{From: "none", To: "low"}})
		require.ErrorContains(t, err, "empty or unknown")

		_, err = NormalizeReasoningEffortMappings(PlatformOpenAI, []ReasoningEffortMapping{{From: "ultra", To: "high"}})
		require.ErrorContains(t, err, "empty or unknown")
	})
}

func TestOpenAIReasoningEffortPolicyForModel(t *testing.T) {
	defaultMappings := []ReasoningEffortMapping{{From: "max", To: "high"}}
	policies := []ReasoningEffortModelPolicy{{
		Model:     "gpt-5.6-sol",
		MaxEffort: "medium",
		Mappings:  []ReasoningEffortMapping{{From: "max", To: "xhigh"}},
	}}

	maxEffort, mappings := OpenAIReasoningEffortPolicyForModel("GPT-5.6-SOL", "", defaultMappings, policies)
	require.Equal(t, "medium", maxEffort)
	require.Equal(t, policies[0].Mappings, mappings)

	maxEffort, mappings = OpenAIReasoningEffortPolicyForModel("gpt-5.6", "", defaultMappings, policies)
	require.Empty(t, maxEffort)
	require.Equal(t, defaultMappings, mappings)
}

func TestOpenAIReasoningEffortPolicyForModelAt(t *testing.T) {
	defaultMappings := []ReasoningEffortMapping{{From: "max", To: "high"}}
	policies := []ReasoningEffortModelPolicy{{
		Model:      "gpt-5.6-sol",
		MaxEffort:  "medium",
		ActiveDays: []int{1},
		StartTime:  "22:00",
		EndTime:    "02:00",
	}}

	for _, tt := range []struct {
		name    string
		at      time.Time
		wantMax string
	}{
		{name: "start day inside overnight window", at: time.Date(2026, time.January, 5, 23, 0, 0, 0, timezone.Location()), wantMax: "medium"},
		{name: "following day inside overnight window", at: time.Date(2026, time.January, 6, 1, 0, 0, 0, timezone.Location()), wantMax: "medium"},
		{name: "following day at window end", at: time.Date(2026, time.January, 6, 2, 0, 0, 0, timezone.Location()), wantMax: ""},
		{name: "start day before window", at: time.Date(2026, time.January, 5, 21, 59, 0, 0, timezone.Location()), wantMax: ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			maxEffort, mappings := OpenAIReasoningEffortPolicyForModelAt(tt.at, "gpt-5.6-sol", "", defaultMappings, policies)
			require.Equal(t, tt.wantMax, maxEffort)
			if tt.wantMax == "" {
				require.Equal(t, defaultMappings, mappings)
			}
		})
	}
}

func TestNormalizeReasoningEffortModelPolicies(t *testing.T) {
	policies, err := NormalizeReasoningEffortModelPolicies(PlatformOpenAI, []ReasoningEffortModelPolicy{{
		Model:      " gpt-5.6-sol ",
		MaxEffort:  " x-high ",
		Mappings:   []ReasoningEffortMapping{{From: "max", To: "high"}},
		ActiveDays: []int{5, 1, 5},
		StartTime:  "9:30",
		EndTime:    "02:00",
	}})
	require.NoError(t, err)
	require.Equal(t, []ReasoningEffortModelPolicy{{
		Model:      "gpt-5.6-sol",
		MaxEffort:  "xhigh",
		Mappings:   []ReasoningEffortMapping{{From: "max", To: "high"}},
		ActiveDays: []int{1, 5},
		StartTime:  "09:30",
		EndTime:    "02:00",
	}}, policies)

	_, err = NormalizeReasoningEffortModelPolicies(PlatformOpenAI, []ReasoningEffortModelPolicy{{Model: "gpt-5"}})
	require.ErrorContains(t, err, "must set a ceiling or mapping")

	_, err = NormalizeReasoningEffortModelPolicies(PlatformOpenAI, []ReasoningEffortModelPolicy{{
		Model: "gpt-5", MaxEffort: "medium", StartTime: "09:00",
	}})
	require.ErrorContains(t, err, "must be set together")
}

func TestNormalizeMaxReasoningEffortForPlatform(t *testing.T) {
	value, err := normalizeMaxReasoningEffortForPlatform(PlatformOpenAI, "max")
	require.NoError(t, err)
	require.Equal(t, "max", value)
	value, err = normalizeMaxReasoningEffortForPlatform(PlatformComposite, "max")
	require.NoError(t, err)
	require.Equal(t, "max", value)

	for _, platform := range []string{PlatformAnthropic, PlatformGemini, PlatformAntigravity, PlatformGrok} {
		_, err = normalizeMaxReasoningEffortForPlatform(platform, "low")
		require.ErrorContains(t, err, "only supported for platforms \"openai\" and \"composite\"")
	}

	_, err = normalizeMaxReasoningEffortForPlatform(PlatformOpenAI, "none")
	require.ErrorContains(t, err, "not supported")
}

func TestOpenAIReasoningEffortPolicyContext(t *testing.T) {
	body := []byte(`{"reasoning":{"effort":"max"}}`)

	unbound, changed := ApplyOpenAIReasoningEffortPolicyFromContext(context.Background(), body)
	require.False(t, changed)
	require.Equal(t, body, unbound)

	mappings := []ReasoningEffortMapping{{From: "max", To: "xhigh"}}
	ctx := WithOpenAIReasoningEffortPolicy(context.Background(), "medium", mappings)
	mappings[0].To = "low"
	got, changed := ApplyOpenAIReasoningEffortPolicyFromContext(ctx, body)
	require.True(t, changed)
	require.Equal(t, "medium", gjson.GetBytes(got, "reasoning.effort").String())
}

func TestApplyOpenAIReasoningEffortPolicy(t *testing.T) {
	tests := []struct {
		name     string
		body     string
		max      string
		mappings []ReasoningEffortMapping
		path     string
		want     string
		changed  bool
	}{
		{name: "nested caps high", body: `{"reasoning":{"effort":"xhigh"}}`, max: "medium", path: "reasoning.effort", want: "medium", changed: true},
		{name: "flat caps high", body: `{"reasoning_effort":"high"}`, max: "low", path: "reasoning_effort", want: "low", changed: true},
		{name: "does not raise omitted", body: `{"model":"gpt-5"}`, max: "low", path: "reasoning_effort", want: "", changed: false},
		{name: "keeps lower value", body: `{"reasoning_effort":"low"}`, max: "high", path: "reasoning_effort", want: "low", changed: false},
		{name: "normalizes request alias", body: `{"reasoning_effort":"x-high"}`, max: "xhigh", path: "reasoning_effort", want: "xhigh", changed: true},
		{name: "caps max below its distinct rank", body: `{"reasoning_effort":"max"}`, max: "xhigh", path: "reasoning_effort", want: "xhigh", changed: true},
		{name: "keeps xhigh below max", body: `{"reasoning_effort":"xhigh"}`, max: "max", path: "reasoning_effort", want: "xhigh", changed: false},
		{name: "ignores stale none ceiling", body: `{"reasoning_effort":"high"}`, max: "none", path: "reasoning_effort", want: "high", changed: false},
		{name: "caps both shapes", body: `{"reasoning":{"effort":"high"},"reasoning_effort":"xhigh"}`, max: "low", path: "reasoning.effort", want: "low", changed: true},
		{name: "maps before cap", body: `{"reasoning":{"effort":"MAX"}}`, max: "medium", mappings: []ReasoningEffortMapping{{From: "max", To: "xhigh"}}, path: "reasoning.effort", want: "medium", changed: true},
		{name: "does not chain mappings", body: `{"reasoning_effort":"max"}`, mappings: []ReasoningEffortMapping{{From: "max", To: "xhigh"}, {From: "xhigh", To: "low"}}, path: "reasoning_effort", want: "xhigh", changed: true},
		{name: "keeps unknown without mapping", body: `{"reasoning_effort":"future"}`, max: "low", path: "reasoning_effort", want: "future", changed: false},
		{name: "keeps non string value", body: `{"reasoning_effort":{"level":"high"}}`, max: "low", path: "reasoning_effort.level", want: "high", changed: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, changed := ApplyOpenAIReasoningEffortPolicy([]byte(tt.body), tt.max, tt.mappings)
			require.Equal(t, tt.changed, changed)
			if tt.path != "" {
				require.Equal(t, tt.want, gjson.GetBytes(got, tt.path).String())
			}
		})
	}
}
