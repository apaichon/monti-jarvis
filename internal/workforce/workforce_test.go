package workforce

import (
	"testing"

	"github.com/libra/monti-jarvis/internal/store"
)

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

func TestFromWorkforceAgent(t *testing.T) {
	agent := FromWorkforceAgent(store.WorkforceAgent{
		ID:              "ava",
		Name:            "Ava",
		Role:            "General Support",
		Trait:           "Warm",
		Color:           "#008cff",
		Voice:           "Aoede",
		VoiceProviderID: "voice-gemini-live",
		VoiceID:         "gemini-model",
		Image:           "/images/ava.jpg",
		Greeting:        "Hello",
		Popular:         true,
		Skin:            "#f0bd9b",
		Hair:            "#5a3428",
	})
	if agent.ID != "ava" || agent.VoiceProviderID != "voice-gemini-live" || agent.Image != "/images/ava.jpg" || !agent.Popular {
		t.Fatalf("FromWorkforceAgent() = %#v, want mapped workforce agent", agent)
	}
}

func TestFindAssignedResolvesCustomAvatar(t *testing.T) {
	assigned := []store.WorkforceAgent{{
		ID:       "mira",
		Name:     "Mira",
		Role:     "Customer Care",
		Voice:    "Kore",
		Greeting: "Hi! Welcome. I'm Mira.",
	}}

	agent, ok := FindAssigned(" MIRA ", assigned)
	if !ok {
		t.Fatal("FindAssigned(mira) = false, want true")
	}
	if agent.ID != "mira" || agent.Name != "Mira" || agent.Greeting != "Hi! Welcome. I'm Mira." {
		t.Fatalf("FindAssigned(mira) = %#v, want Mira catalog data", agent)
	}
	if _, ok := FindAssigned("ava", assigned); ok {
		t.Fatal("FindAssigned(ava) = true, want false for unassigned avatar")
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
