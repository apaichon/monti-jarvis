package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/store"
	"github.com/libra/monti-jarvis/internal/tenantoauth"
)

type oauthCompleteRequest struct {
	SessionID   string `json:"session_id"`
	CompanyName string `json:"company_name"`
	Slug        string `json:"slug"`
}

func (s *server) startTenantOAuth(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.TenantRegisterEnabled {
		writeError(w, http.StatusServiceUnavailable, "tenant registration is disabled")
		return
	}
	provider := strings.TrimSpace(r.PathValue("provider"))
	if s.tenantOAuth == nil || !s.tenantOAuth.ProviderEnabled(provider) {
		writeError(w, http.StatusServiceUnavailable, provider+" sign-in is not configured")
		return
	}
	startURL, err := s.tenantOAuth.StartURL(r.Context(), tenantoauth.StartParams{
		Provider:    provider,
		CompanyName: r.URL.Query().Get("company_name"),
		Slug:        r.URL.Query().Get("slug"),
		DisplayName: r.URL.Query().Get("display_name"),
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	http.Redirect(w, r, startURL, http.StatusFound)
}

func (s *server) tenantOAuthCallback(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.TenantRegisterEnabled {
		writeError(w, http.StatusServiceUnavailable, "tenant registration is disabled")
		return
	}
	provider := strings.TrimSpace(r.PathValue("provider"))
	code := strings.TrimSpace(r.URL.Query().Get("code"))
	state := strings.TrimSpace(r.URL.Query().Get("state"))
	if code == "" || state == "" {
		writeError(w, http.StatusBadRequest, "missing oauth code")
		return
	}
	if s.tenantOAuth == nil {
		writeError(w, http.StatusServiceUnavailable, "oauth is not configured")
		return
	}

	ctx := auth.WithRequestMeta(r.Context(), r)
	identity, payload, err := s.tenantOAuth.Exchange(ctx, provider, state, code)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if payload.DisplayName != "" {
		identity.DisplayName = payload.DisplayName
	}

	base := strings.TrimRight(s.cfg.PublicBaseURL, "/") + "/tenant"

	// Existing Google/GitHub account → sign in (login page or re-auth after KYC).
	if pair, ok, err := s.finishOAuthLogin(ctx, identity); err != nil {
		s.redirectOAuthError(w, r, base, err)
		return
	} else if ok {
		redirectSuccess(w, r, base+"/login", pair)
		return
	}

	if payload.CompanyName != "" && payload.Slug != "" {
		result, user, pair, err := s.finishOAuthRegistration(ctx, identity, payload.CompanyName, payload.Slug)
		if err != nil {
			s.redirectOAuthError(w, r, base, err)
			return
		}
		s.publishTenantRegistered(ctx, result, user)
		redirectSuccess(w, r, base+"/register/success", pair)
		return
	}

	// New identity without company/slug → collect workspace details.
	session, err := s.tenantOAuth.CreatePendingSession(ctx, identity)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	http.Redirect(w, r, base+"/register/complete?session="+url.QueryEscape(session.ID), http.StatusFound)
}

func (s *server) completeTenantOAuth(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.TenantRegisterEnabled {
		writeError(w, http.StatusServiceUnavailable, "tenant registration is disabled")
		return
	}
	var req oauthCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if s.tenantOAuth == nil {
		writeError(w, http.StatusServiceUnavailable, "oauth is not configured")
		return
	}
	ctx := auth.WithRequestMeta(r.Context(), r)
	session, err := s.tenantOAuth.LoadPendingSession(ctx, req.SessionID)
	if err != nil {
		if errors.Is(err, tenantoauth.ErrSessionExpired) {
			writeError(w, http.StatusBadRequest, "oauth session expired")
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	identity := tenantoauth.Identity{
		Provider:       session.Provider,
		ProviderUserID: session.ProviderUserID,
		Email:          session.Email,
		DisplayName:    session.DisplayName,
	}
	// Race: identity registered while pending session was open → sign in.
	if pair, ok, err := s.finishOAuthLogin(ctx, identity); err != nil {
		writeRegisterError(w, err)
		return
	} else if ok {
		s.tenantOAuth.DeletePendingSession(ctx, session.ID)
		writeJSON(w, http.StatusOK, registerTenantResponse{
			TenantID:     pair.User.TenantID,
			Slug:         pair.User.TenantID,
			AccessToken:  pair.AccessToken,
			RefreshToken: pair.RefreshToken,
			ExpiresIn:    pair.ExpiresIn,
			TokenType:    pair.TokenType,
			User:         pair.User,
		})
		return
	}
	result, user, pair, err := s.finishOAuthRegistration(ctx, identity, req.CompanyName, req.Slug)
	if err != nil {
		writeRegisterError(w, err)
		return
	}
	s.tenantOAuth.DeletePendingSession(ctx, session.ID)
	s.publishTenantRegistered(ctx, result, user)
	writeJSON(w, http.StatusCreated, registerTenantResponse{
		TenantID:       result.TenantID,
		Slug:           result.Slug,
		RegistrationID: result.RegistrationID,
		AccessToken:    pair.AccessToken,
		RefreshToken:   pair.RefreshToken,
		ExpiresIn:      pair.ExpiresIn,
		TokenType:      pair.TokenType,
		User:           pair.User,
	})
}

func (s *server) finishOAuthLogin(ctx context.Context, identity tenantoauth.Identity) (auth.TokenPair, bool, error) {
	if s.auth == nil || !s.auth.TokensEnabled() {
		return auth.TokenPair{}, false, auth.ErrNotConfigured
	}
	if identity.Provider == "" || identity.ProviderUserID == "" {
		return auth.TokenPair{}, false, nil
	}
	userID, err := s.store.OAuthIdentityUserID(ctx, identity.Provider, identity.ProviderUserID)
	if err != nil {
		return auth.TokenPair{}, false, err
	}
	if userID == "" {
		return auth.TokenPair{}, false, nil
	}
	user, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return auth.TokenPair{}, false, err
	}
	if user.Status != "active" {
		return auth.TokenPair{}, false, errOAuthUserInactive
	}
	if user.Role != string(auth.RoleTenantAdmin) && user.Role != string(auth.RolePlatformAdmin) {
		return auth.TokenPair{}, false, errOAuthUserInactive
	}
	pair, err := s.auth.IssueTokenPairForUser(ctx, user)
	if err != nil {
		return auth.TokenPair{}, false, err
	}
	return pair, true, nil
}

func (s *server) finishOAuthRegistration(ctx context.Context, identity tenantoauth.Identity, companyName, slug string) (*store.RegisterTenantResult, store.AuthUser, auth.TokenPair, error) {
	if s.auth == nil || !s.auth.TokensEnabled() {
		return nil, store.AuthUser{}, auth.TokenPair{}, auth.ErrNotConfigured
	}
	result, err := s.store.RegisterTenant(ctx, store.RegisterTenantInput{
		CompanyName:         companyName,
		Slug:                slug,
		AdminEmail:          identity.Email,
		AdminDisplayName:    identity.DisplayName,
		AuthProvider:        identity.Provider,
		EmailVerified:       true,
		OAuthProvider:       identity.Provider,
		OAuthProviderUserID: identity.ProviderUserID,
	})
	if err != nil {
		return nil, store.AuthUser{}, auth.TokenPair{}, err
	}
	user, err := s.store.GetUserByID(ctx, result.UserID)
	if err != nil {
		return nil, store.AuthUser{}, auth.TokenPair{}, err
	}
	pair, err := s.auth.IssueTokenPairForUser(ctx, user)
	if err != nil {
		return nil, store.AuthUser{}, auth.TokenPair{}, err
	}
	s.sendRegistrationCompleteEmail(ctx, user, result.TenantID)
	return result, user, pair, nil
}

var errOAuthUserInactive = errors.New("oauth user inactive")

func redirectSuccess(w http.ResponseWriter, r *http.Request, successPath string, pair auth.TokenPair) {
	q := url.Values{}
	q.Set("access_token", pair.AccessToken)
	q.Set("refresh_token", pair.RefreshToken)
	q.Set("tenant_id", pair.User.TenantID)
	q.Set("user_id", pair.User.ID)
	q.Set("email", pair.User.Email)
	q.Set("display_name", pair.User.DisplayName)
	q.Set("role", string(pair.User.Role))
	if pair.ExpiresIn > 0 {
		q.Set("expires_in", fmt.Sprintf("%d", pair.ExpiresIn))
	}
	http.Redirect(w, r, successPath+"?"+q.Encode(), http.StatusFound)
}

func (s *server) redirectOAuthError(w http.ResponseWriter, r *http.Request, base string, err error) {
	msg := "registration failed"
	dest := base + "/register"
	switch {
	case errors.Is(err, store.ErrTenantSlugTaken):
		msg = "slug already taken"
	case errors.Is(err, store.ErrTenantEmailRegistered):
		msg = "email already registered"
	case errors.Is(err, store.ErrOAuthIdentityInUse):
		msg = "account already linked"
	case errors.Is(err, errOAuthUserInactive):
		msg = "account is not active"
		dest = base + "/login"
	case errors.Is(err, auth.ErrNotConfigured):
		msg = "sign-in is not configured"
		dest = base + "/login"
	}
	http.Redirect(w, r, dest+"?error="+url.QueryEscape(msg), http.StatusFound)
}

func (s *server) oauthProviders(w http.ResponseWriter, _ *http.Request) {
	providers := []string{}
	if s.tenantOAuth != nil {
		if s.tenantOAuth.ProviderEnabled("google") {
			providers = append(providers, "google")
		}
		if s.tenantOAuth.ProviderEnabled("github") {
			providers = append(providers, "github")
		}
	}
	writeJSON(w, http.StatusOK, map[string]any{"providers": providers})
}