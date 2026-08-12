package domain

// ReasoningEffortMapping rewrites one explicit OpenAI/Codex reasoning effort
// value to another before the group ceiling is applied.
type ReasoningEffortMapping struct {
	From string `json:"from"`
	To   string `json:"to"`
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
