package workforce

import "testing"

func TestGetKnownAgent(t *testing.T) {
	agent, ok := Get("luna")
	if !ok {
		t.Fatal("Get(luna) = false, want true")
	}
	if agent.Name != "Luna" || agent.Voice != "Kore" {
		t.Fatalf("agent = %#v, want Luna with Kore voice", agent)
	}
}

func TestResolveFallsBackToDefault(t *testing.T) {
	agent := Resolve("unknown")
	if agent.ID != "ava" {
		t.Fatalf("Resolve(unknown).ID = %q, want ava", agent.ID)
	}
}

func TestSystemPromptIncludesRole(t *testing.T) {
	agent, _ := Get("max")
	prompt := SystemPrompt(agent)
	if !containsAll(prompt, "Max", "Billing Specialist", agent.Greeting) {
		t.Fatalf("SystemPrompt() missing expected content: %q", prompt)
	}
}

func TestAllReturnsWorkforce(t *testing.T) {
	if got, want := len(All()), 4; got != want {
		t.Fatalf("len(All()) = %d, want %d", got, want)
	}
}

func containsAll(text string, parts ...string) bool {
	for _, part := range parts {
		if !contains(text, part) {
			return false
		}
	}
	return true
}

func contains(text, part string) bool {
	return len(part) == 0 || (len(text) >= len(part) && indexOf(text, part) >= 0)
}

func indexOf(text, part string) int {
	for i := 0; i+len(part) <= len(text); i++ {
		if text[i:i+len(part)] == part {
			return i
		}
	}
	return -1
}