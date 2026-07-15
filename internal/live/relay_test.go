package live

import (
	"context"
	"testing"

	"github.com/libra/monti-jarvis/internal/workforce"
)

func TestResolveAgentUsesTenantCatalog(t *testing.T) {
	r := &Relay{AgentResolver: func(_ context.Context, tenantID, agentID string) (workforce.Agent, bool) {
		if tenantID == "demo" && agentID == "mira" {
			return workforce.Agent{ID: "mira", Name: "Mira", Greeting: "Hi, I'm Mira."}, true
		}
		return workforce.Agent{}, false
	}}

	agent := r.resolveAgent(context.Background(), "demo", "mira")
	if agent.ID != "mira" || agent.Greeting != "Hi, I'm Mira." {
		t.Fatalf("resolveAgent(demo, mira) = %#v, want tenant catalog Mira", agent)
	}
}

func TestResolveAgentFallsBackToBuiltIn(t *testing.T) {
	r := &Relay{AgentResolver: func(context.Context, string, string) (workforce.Agent, bool) {
		return workforce.Agent{}, false
	}}

	if got := r.resolveAgent(context.Background(), "demo", "luna"); got.ID != "luna" {
		t.Fatalf("resolveAgent(demo, luna).ID = %q, want luna", got.ID)
	}
}

func TestCustomerEndConfirmation(t *testing.T) {
	for _, phrase := range []string{
		"ไม่มีแล้วครับ ขอบคุณครับ",
		"ไม่มีอะไรแล้วค่ะ",
		"หมดคำถามแล้วครับ",
		"No more questions, thank you",
	} {
		if !customerEndConfirmation(phrase) {
			t.Errorf("customerEndConfirmation(%q) = false, want true", phrase)
		}
	}

	if customerEndConfirmation("ขอถามอีกเรื่องครับ") {
		t.Fatal("customerEndConfirmation matched a non-closing caller message")
	}
}
