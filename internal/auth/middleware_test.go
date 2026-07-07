package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRequireKMWriteForbidden(t *testing.T) {
	issuer, _ := NewTokenIssuer("abcdefghijklmnopqrstuvwxyz012345", time.Minute)
	token, _, _, _ := issuer.IssueAccess("u1", "c@x.local", RoleCustomer, "demo")
	svc := &Service{issuer: issuer, authDisabled: false}
	guard := NewHTTPGuard(svc, false)

	called := false
	handler := guard.RequireKMWrite(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/km/agents/ava/documents", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rec.Code)
	}
	if called {
		t.Fatal("handler should not run")
	}
}

func TestRequireKMWriteAllowsTenantAdmin(t *testing.T) {
	issuer, _ := NewTokenIssuer("abcdefghijklmnopqrstuvwxyz012345", time.Minute)
	token, _, _, _ := issuer.IssueAccess("u1", "a@x.local", RoleTenantAdmin, "demo")
	svc := &Service{issuer: issuer, authDisabled: false}
	guard := NewHTTPGuard(svc, false)

	called := false
	handler := guard.RequireKMWrite(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodPost, "/api/km/agents/ava/documents", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if !called {
		t.Fatal("handler should run")
	}
}

func TestRequireKMWriteBypassWhenDisabled(t *testing.T) {
	guard := NewHTTPGuard(nil, true)
	called := false
	handler := guard.RequireKMWrite(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	req := httptest.NewRequest(http.MethodPost, "/api/km/seed", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if !called {
		t.Fatal("handler should run when auth disabled")
	}
}