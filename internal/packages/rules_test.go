package packages

import (
	"testing"
)

func TestValidateRules(t *testing.T) {
	fields := []byte(`{
		"max_ai_employees": {"type":"int","min":0,"required":true},
		"voice_enabled": {"type":"bool","required":true}
	}`)
	rules := map[string]any{
		"max_ai_employees": 2,
		"voice_enabled":    true,
	}
	if err := ValidateRules(fields, rules); err != nil {
		t.Fatal(err)
	}
}

func TestValidateRulesUnknownKey(t *testing.T) {
	fields := []byte(`{"max_ai_employees":{"type":"int","min":0,"required":true}}`)
	rules := map[string]any{"max_ai_employees": 1, "extra": true}
	if err := ValidateRules(fields, rules); err == nil {
		t.Fatal("expected unknown field error")
	}
}