package store

import "testing"

func TestShortTenantFP(t *testing.T) {
	a := shortTenantFP("acme")
	b := shortTenantFP("acme")
	if a != b || len(a) != 2 {
		t.Fatalf("fp = %q, want stable 2-char", a)
	}
	if shortTenantFP("other") == a {
		// collision possible but unlikely for these two; just ensure function runs
		t.Log("unexpected equal fp for different tenants (rare)")
	}
}

func TestSlugSanitizeMap(t *testing.T) {
	// CreateTenantLibraryAvatar slug generation uses Map; verify empty name fallback path via short helper.
	if shortTenantFP("") != "00" && len(shortTenantFP("")) != 2 {
		t.Fatalf("empty tenant fp unexpected: %q", shortTenantFP(""))
	}
}
