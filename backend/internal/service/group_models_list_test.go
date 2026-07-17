package service

import "testing"

func TestGroupAllowsOpenAIModel(t *testing.T) {
	tests := []struct {
		name  string
		group *Group
		model string
		want  bool
	}{
		{name: "nil group", want: true},
		{name: "non OpenAI group", group: &Group{Platform: PlatformAnthropic, ModelsListConfig: GroupModelsListConfig{Enabled: true, Models: []string{"gpt-5.6"}}}, model: "gpt-5.5", want: true},
		{name: "disabled", group: &Group{Platform: PlatformOpenAI, ModelsListConfig: GroupModelsListConfig{Models: []string{"gpt-5.6"}}}, model: "gpt-5.5", want: true},
		{name: "empty", group: &Group{Platform: PlatformOpenAI, ModelsListConfig: GroupModelsListConfig{Enabled: true}}, model: "gpt-5.5", want: true},
		{name: "exact match", group: &Group{Platform: PlatformOpenAI, ModelsListConfig: GroupModelsListConfig{Enabled: true, Models: []string{"gpt-5.6"}}}, model: "gpt-5.6", want: true},
		{name: "trim request", group: &Group{Platform: PlatformOpenAI, ModelsListConfig: GroupModelsListConfig{Enabled: true, Models: []string{"gpt-5.6"}}}, model: " gpt-5.6 ", want: true},
		{name: "denied", group: &Group{Platform: PlatformOpenAI, ModelsListConfig: GroupModelsListConfig{Enabled: true, Models: []string{"gpt-5.6"}}}, model: "gpt-5.5", want: false},
		{name: "case sensitive", group: &Group{Platform: PlatformOpenAI, ModelsListConfig: GroupModelsListConfig{Enabled: true, Models: []string{"gpt-5.6"}}}, model: "GPT-5.6", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.group.AllowsOpenAIModel(tt.model); got != tt.want {
				t.Fatalf("AllowsOpenAIModel(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}
