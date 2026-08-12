package domain

const (
	ReasoningEffortMatchExact  = "exact"
	ReasoningEffortMatchPrefix = "prefix"
	ReasoningEffortMatchSuffix = "suffix"
)

// ReasoningEffortMapping rewrites one explicit OpenAI/Codex reasoning effort
// value to another before the group ceiling is applied. To may also be "deny"
// to reject the request when the matching source value is present.
//
// Model and MatchType optionally scope the rewrite to a request model:
// exact matches the full model id, prefix/suffix match a model-id affix.
// Empty MatchType and empty Model mean the mapping applies to every model.
type ReasoningEffortMapping struct {
	From      string `json:"from"`
	To        string `json:"to"`
	MatchType string `json:"match_type,omitempty"`
	Model     string `json:"model,omitempty"`
}

// ReasoningEffortModelPolicy overrides the group default for one exact model.
// Model matching uses the client-requested model name.
type ReasoningEffortModelPolicy struct {
	Model     string                   `json:"model"`
	MaxEffort string                   `json:"max_effort"`
	Mappings  []ReasoningEffortMapping `json:"mappings"`
	// ActiveDays uses ISO weekdays: 1 is Monday and 7 is Sunday. Empty means every day.
	ActiveDays []int  `json:"active_days,omitempty"`
	StartTime  string `json:"start_time,omitempty"`
	EndTime    string `json:"end_time,omitempty"`
}
