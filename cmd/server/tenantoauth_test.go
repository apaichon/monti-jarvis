package main

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/libra/monti-jarvis/internal/auth"
)

func TestRedirectSuccessQuery(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/cb", nil)
	pair := auth.TokenPair{
		AccessToken:  "at",
		RefreshToken: "rt",
		ExpiresIn:    900,
		TokenType:    "Bearer",
		User: auth.UserProfile{
			ID:          "usr_acme_admin",
			Email:       "owner@acme.test",
			DisplayName: "Owner",
			Role:        auth.RoleTenantAdmin,
			TenantID:    "acme",
		},
	}
	redirectSuccess(rr, req, "http://localhost:8091/tenant/login", pair)
	if rr.Code != http.StatusFound {
		t.Fatalf("status %d", rr.Code)
	}
	loc := rr.Header().Get("Location")
	u, err := url.Parse(loc)
	if err != nil {
		t.Fatal(err)
	}
	q := u.Query()
	for _, key := range []string{"access_token", "refresh_token", "tenant_id", "user_id", "email", "display_name", "role"} {
		if q.Get(key) == "" {
			t.Fatalf("missing query %s in %s", key, loc)
		}
	}
	if q.Get("role") != "tenant_admin" || q.Get("tenant_id") != "acme" {
		t.Fatalf("unexpected profile in redirect: %s", loc)
	}
	if !strings.HasPrefix(loc, "http://localhost:8091/tenant/login?") {
		t.Fatalf("unexpected location %s", loc)
	}
}

func TestRedirectOAuthErrorLoginForInactive(t *testing.T) {
	s := &server{}
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/cb", nil)
	s.redirectOAuthError(rr, req, "http://localhost:8091/tenant", errOAuthUserInactive)
	loc := rr.Header().Get("Location")
	if !strings.Contains(loc, "/login?") {
		t.Fatalf("expected login redirect, got %s", loc)
	}
	if !strings.Contains(loc, "not+active") && !strings.Contains(loc, "not%20active") {
		t.Fatalf("expected inactive message, got %s", loc)
	}
}
