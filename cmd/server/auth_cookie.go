package main

import (
	"net/http"
	"strings"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/env"
)

const (
	tenantRefreshCookie   = "monti_tenant_refresh"
	customerRefreshCookie = "monti_customer_refresh"
)

func refreshCookieConfig(cfg env.Config, name, path string, value string, maxAge int) *http.Cookie {
	sameSite := http.SameSiteLaxMode
	if strings.EqualFold(strings.TrimSpace(cfg.CookieSameSite), "strict") {
		sameSite = http.SameSiteStrictMode
	}
	return &http.Cookie{
		Name:     name,
		Value:    value,
		Path:     path,
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   cfg.CookieSecure,
		SameSite: sameSite,
	}
}

func setRefreshCookie(w http.ResponseWriter, cfg env.Config, name, path, value string, maxAge int) {
	http.SetCookie(w, refreshCookieConfig(cfg, name, path, value, maxAge))
}

func clearRefreshCookie(w http.ResponseWriter, cfg env.Config, name, path string) {
	setRefreshCookie(w, cfg, name, path, "", -1)
}

func requestRefreshToken(r *http.Request, cookieName, bodyValue string) string {
	if cookie, err := r.Cookie(cookieName); err == nil && strings.TrimSpace(cookie.Value) != "" {
		return strings.TrimSpace(cookie.Value)
	}
	return strings.TrimSpace(bodyValue)
}

func browserTokenPair(pair auth.TokenPair) auth.TokenPair {
	// Browser clients use the HttpOnly refresh cookie. Keep the refresh token
	// only in the internal service result for mobile/legacy callers.
	pair.RefreshToken = ""
	return pair
}
