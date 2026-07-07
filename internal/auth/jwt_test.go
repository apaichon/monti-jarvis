package auth

import (
	"testing"
	"time"
)

func TestJWTRoundTrip(t *testing.T) {
	issuer, err := NewTokenIssuer("abcdefghijklmnopqrstuvwxyz012345", 15*time.Minute)
	if err != nil {
		t.Fatal(err)
	}
	token, jti, expiresIn, err := issuer.IssueAccess("usr_1", "admin@demo.local", RoleTenantAdmin, "demo")
	if err != nil {
		t.Fatal(err)
	}
	if expiresIn <= 0 || token == "" || jti == "" {
		t.Fatal("expected non-empty token and jti")
	}
	ac, _, err := issuer.ParseAccess(token)
	if err != nil {
		t.Fatal(err)
	}
	if ac.UserID != "usr_1" || ac.Email != "admin@demo.local" || ac.Role != RoleTenantAdmin || ac.TenantID != "demo" || ac.JTI != jti {
		t.Fatalf("unexpected claims: %+v", ac)
	}
}

func TestJWTRejectsShortSecret(t *testing.T) {
	_, err := NewTokenIssuer("short", time.Minute)
	if err == nil {
		t.Fatal("expected error for short secret")
	}
}