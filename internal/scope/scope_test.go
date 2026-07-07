package scope

import "testing"

func TestResolveBillingAgent(t *testing.T) {
	got := Resolve("max", "billing")
	if len(got) != 1 || got[0] != "billing" {
		t.Fatalf("Resolve(max,billing) = %v, want [billing]", got)
	}
}

func TestResolveAvaOnTechnicalTab(t *testing.T) {
	got := Resolve("ava", "technical")
	if len(got) != 1 || got[0] != "general" {
		t.Fatalf("Resolve(ava,technical) = %v, want [general]", got)
	}
}

func TestResolveNeoTriage(t *testing.T) {
	got := Resolve("neo", "billing")
	if len(got) != 1 || got[0] != "billing" {
		t.Fatalf("Resolve(neo,billing) = %v, want [billing]", got)
	}
}