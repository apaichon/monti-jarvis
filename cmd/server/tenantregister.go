package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/natsbus"
	"github.com/libra/monti-jarvis/internal/store"
	"github.com/libra/monti-jarvis/internal/tenantregister"
)

type registerTenantRequest struct {
	CompanyName      string `json:"company_name"`
	Slug             string `json:"slug"`
	AdminEmail       string `json:"admin_email"`
	AdminPassword    string `json:"admin_password"`
	AdminDisplayName string `json:"admin_display_name"`
}

type registerTenantResponse struct {
	TenantID             string           `json:"tenant_id,omitempty"`
	Slug                 string           `json:"slug,omitempty"`
	RegistrationID       string           `json:"registration_id,omitempty"`
	AccessToken          string           `json:"access_token,omitempty"`
	RefreshToken         string           `json:"refresh_token,omitempty"`
	ExpiresIn            int              `json:"expires_in,omitempty"`
	TokenType            string           `json:"token_type,omitempty"`
	User                 auth.UserProfile `json:"user,omitempty"`
	VerificationRequired bool             `json:"verification_required,omitempty"`
	Message              string           `json:"message,omitempty"`
}

func (s *server) registerTenant(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.TenantRegisterEnabled {
		writeError(w, http.StatusServiceUnavailable, "tenant registration is disabled")
		return
	}
	if s.store == nil || s.store.Health(r.Context()).Postgres != "ok" {
		writeError(w, http.StatusServiceUnavailable, "registration is unavailable")
		return
	}
	if s.auth == nil || !s.auth.TokensEnabled() {
		writeError(w, http.StatusServiceUnavailable, "auth is not configured")
		return
	}

	var req registerTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	ctx := auth.WithRequestMeta(r.Context(), r)
	if s.registerLimiter != nil {
		meta := auth.RequestMetaFrom(ctx)
		allowed, err := s.registerLimiter.Allow(ctx, meta.IP)
		if err != nil {
			log.Printf("register rate limit warning: %v", err)
		} else if !allowed {
			writeError(w, http.StatusTooManyRequests, "too many registration attempts")
			return
		}
	}

	hash, err := auth.HashPassword(req.AdminPassword)
	if err != nil {
		writeError(w, http.StatusBadRequest, "invalid password")
		return
	}

	result, err := s.store.RegisterTenant(ctx, store.RegisterTenantInput{
		CompanyName:      req.CompanyName,
		Slug:             req.Slug,
		AdminEmail:       req.AdminEmail,
		AdminPassword:    req.AdminPassword,
		AdminDisplayName: req.AdminDisplayName,
		PasswordHash:     hash,
		AuthProvider:     "email",
		EmailVerified:    false,
	})
	if err != nil {
		writeRegisterError(w, err)
		return
	}

	user, err := s.store.GetUserByID(ctx, result.UserID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	rawToken, err := s.store.CreateEmailVerificationToken(ctx, result.UserID, 24*time.Hour)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	s.sendVerificationEmail(ctx, user, rawToken)
	s.publishTenantRegistered(ctx, result, user)

	writeJSON(w, http.StatusCreated, registerTenantResponse{
		TenantID:             result.TenantID,
		Slug:                 result.Slug,
		RegistrationID:       result.RegistrationID,
		VerificationRequired: true,
		Message:              "Check your email to verify your account before signing in.",
	})
}

func (s *server) verifyTenantEmail(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		var body struct {
			Token string `json:"token"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		token = body.Token
	}
	token = strings.TrimSpace(token)
	if token == "" {
		writeError(w, http.StatusBadRequest, "token is required")
		return
	}
	if s.auth == nil || !s.auth.TokensEnabled() {
		writeError(w, http.StatusServiceUnavailable, "auth is not configured")
		return
	}

	ctx := r.Context()
	user, err := s.store.VerifyEmailToken(ctx, token)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrVerificationExpired):
			writeError(w, http.StatusBadRequest, "verification token expired")
		case errors.Is(err, store.ErrVerificationInvalid):
			writeError(w, http.StatusBadRequest, "verification token invalid")
		default:
			writeError(w, http.StatusBadGateway, err.Error())
		}
		return
	}

	pair, err := s.auth.IssueTokenPairForUser(ctx, user)
	if err != nil {
		writeAuthHandlerError(w, err)
		return
	}
	s.sendRegistrationCompleteEmail(ctx, user, user.TenantID)

	writeJSON(w, http.StatusOK, registerTenantResponse{
		TenantID:       user.TenantID,
		Slug:           user.TenantID,
		AccessToken:    pair.AccessToken,
		RefreshToken:   pair.RefreshToken,
		ExpiresIn:      pair.ExpiresIn,
		TokenType:      pair.TokenType,
		User:           pair.User,
		Message:        "Email verified. You can sign in and complete KYC in the tenant backoffice.",
	})
}

func (s *server) listPlatformTenants(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	items, total, err := s.store.ListTenants(r.Context(), status, limit, offset)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if limit <= 0 {
		limit = 50
	}

	tenants := make([]map[string]any, 0, len(items))
	for _, item := range items {
		tenants = append(tenants, map[string]any{
			"id":              item.ID,
			"slug":            item.Slug,
			"name":            item.Name,
			"status":          item.Status,
			"registration_id": item.RegistrationID,
			"admin_email":     item.AdminEmail,
			"created_at":      item.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tenants": tenants,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}

func writeRegisterError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, tenantregister.ErrInvalidInput):
		writeError(w, http.StatusBadRequest, "invalid registration input")
	case errors.Is(err, tenantregister.ErrReservedSlug):
		writeError(w, http.StatusBadRequest, "slug is reserved")
	case errors.Is(err, store.ErrTenantSlugTaken):
		writeError(w, http.StatusConflict, "slug already taken")
	case errors.Is(err, store.ErrTenantEmailRegistered):
		writeError(w, http.StatusConflict, "email already registered")
	case errors.Is(err, store.ErrOAuthIdentityInUse):
		writeError(w, http.StatusConflict, "account already linked")
	default:
		writeError(w, http.StatusBadGateway, err.Error())
	}
}

func (s *server) publishTenantRegistered(ctx context.Context, result *store.RegisterTenantResult, user store.AuthUser) {
	if s.bus == nil || !s.bus.Enabled() || !s.cfg.AuthEventsEnabled {
		return
	}
	meta := auth.RequestMetaFrom(ctx)
	event := natsbus.AuthEvent{
		Event:     "auth.tenant.registered",
		TenantID:  result.TenantID,
		UserID:    user.ID,
		Email:     user.Email,
		Role:      user.Role,
		IP:        meta.IP,
		UserAgent: meta.UserAgent,
		At:        time.Now().UTC(),
		Meta: map[string]any{
			"registration_id": result.RegistrationID,
			"status":          "pending_kyc",
			"auth_provider":   result.AuthProvider,
		},
	}
	if err := s.bus.PublishAuthEvent(ctx, event); err != nil {
		log.Printf("register event warning: %v", err)
	}
}