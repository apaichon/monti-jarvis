package main

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/libra/monti-jarvis/internal/auth"
)

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type logoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *server) login(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil || !s.auth.TokensEnabled() {
		writeError(w, http.StatusServiceUnavailable, "auth is not configured")
		return
	}
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	ctx := auth.WithRequestMeta(r.Context(), r)
	pair, err := s.auth.Login(ctx, req.Email, req.Password)
	if err != nil {
		writeAuthHandlerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pair)
}

func (s *server) refreshToken(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil || !s.auth.TokensEnabled() {
		writeError(w, http.StatusServiceUnavailable, "auth is not configured")
		return
	}
	var req refreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	ctx := auth.WithRequestMeta(r.Context(), r)
	pair, err := s.auth.RefreshWithAccess(ctx, req.RefreshToken, r.Header.Get("Authorization"))
	if err != nil {
		writeAuthHandlerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, pair)
}

func (s *server) logout(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil || !s.auth.TokensEnabled() {
		writeError(w, http.StatusServiceUnavailable, "auth is not configured")
		return
	}
	var req logoutRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	ctx := auth.WithRequestMeta(r.Context(), r)
	_ = s.auth.Logout(ctx, req.RefreshToken, r.Header.Get("Authorization"))
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) me(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil || !s.auth.TokensEnabled() {
		writeError(w, http.StatusServiceUnavailable, "auth is not configured")
		return
	}
	ac, err := s.auth.ParseBearer(r.Header.Get("Authorization"))
	if err != nil {
		writeAuthHandlerError(w, err)
		return
	}
	profile, err := s.auth.Me(r.Context(), ac.UserID)
	if err != nil {
		writeAuthHandlerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, profile)
}

func writeAuthHandlerError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, auth.ErrEmailNotVerified):
		writeError(w, http.StatusForbidden, "email not verified")
	case errors.Is(err, auth.ErrUnauthorized):
		writeError(w, http.StatusUnauthorized, "unauthorized")
	case errors.Is(err, auth.ErrForbidden):
		writeError(w, http.StatusForbidden, "forbidden")
	case errors.Is(err, auth.ErrNotConfigured):
		writeError(w, http.StatusServiceUnavailable, "auth is not configured")
	default:
		writeError(w, http.StatusBadGateway, err.Error())
	}
}