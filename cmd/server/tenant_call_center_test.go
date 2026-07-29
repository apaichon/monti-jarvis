package main

import "testing"

func TestCallCenterTopicFromSummaryNormalizesTopic(t *testing.T) {
	tests := []struct {
		name    string
		summary map[string]any
		want    string
	}{
		{name: "missing", summary: nil, want: "unknown"},
		{name: "plain", summary: map[string]any{"topic": "Billing"}, want: "billing"},
		{name: "general", summary: map[string]any{"topic": "general"}, want: "general"},
		{name: "spaced", summary: map[string]any{"topic": "Technical Support"}, want: "technical_support"},
		{name: "invalid", summary: map[string]any{"topic": "billing!"}, want: "unknown"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := callCenterTopicFromSummary(tt.summary); got != tt.want {
				t.Fatalf("topic = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseCallCenterTopicFilter(t *testing.T) {
	if got, err := parseCallCenterTopicFilter("Billing Issues"); err != nil || got != "billing_issues" {
		t.Fatalf("topic filter = %q, %v; want billing_issues", got, err)
	}
	if got, err := parseCallCenterTopicFilter(""); err != nil || got != "" {
		t.Fatalf("empty topic filter = %q, %v; want empty", got, err)
	}
	if _, err := parseCallCenterTopicFilter("billing!"); err == nil {
		t.Fatal("expected invalid topic filter error")
	}
}
