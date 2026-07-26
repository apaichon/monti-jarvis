package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/env"
)

func TestRefreshCookieIsHttpOnlyAndSecureInProduction(t *testing.T) {
	rr := httptest.NewRecorder()
	cfg := env.Config{CookieSecure: true, CookieSameSite: "strict"}
	setRefreshCookie(rr, cfg, tenantRefreshCookie, "/api/auth", "refresh-secret", 3600)
	cookie := rr.Result().Cookies()[0]
	if !cookie.HttpOnly || !cookie.Secure || cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("unsafe refresh cookie: %+v", cookie)
	}
	if cookie.Value != "refresh-secret" || cookie.Path != "/api/auth" {
		t.Fatalf("unexpected refresh cookie: %+v", cookie)
	}
}

func TestBrowserTokenPairDoesNotExposeRefreshCredential(t *testing.T) {
	pair := browserTokenPair(auth.TokenPair{AccessToken: "access", RefreshToken: "refresh"})
	if pair.RefreshToken != "" {
		t.Fatal("browser token pair retained refresh credential")
	}
}

func TestRequestRefreshTokenPrefersHttpOnlyCookie(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: tenantRefreshCookie, Value: "cookie-token"})
	if got := requestRefreshToken(req, tenantRefreshCookie, "body-token"); got != "cookie-token" {
		t.Fatalf("refresh token = %q", got)
	}
	if got := requestRefreshToken(httptest.NewRequest(http.MethodPost, "/api/auth/refresh", nil), tenantRefreshCookie, "body-token"); got != "body-token" {
		t.Fatalf("legacy refresh token = %q", got)
	}
}

func TestSecurityPostureIsMetadataOnly(t *testing.T) {
	s := &server{cfg: env.Config{JWTSecret: "do-not-return-this-secret", AuthDisabled: true}}
	rr := httptest.NewRecorder()
	s.securityPosture(rr, httptest.NewRequest(http.MethodGet, "/api/platform/security/posture", nil))
	body, _ := io.ReadAll(rr.Result().Body)
	if strings.Contains(string(body), "do-not-return-this-secret") {
		t.Fatal("security posture exposed a secret")
	}
	if !strings.Contains(string(body), `"status":"degraded"`) {
		t.Fatalf("expected degraded development posture: %s", body)
	}
}
