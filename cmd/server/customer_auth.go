package main

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/auditctx"
	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/store"
)

type customerAuthSettingsBody struct {
	Enabled           *bool    `json:"enabled"`
	AuthMode          string   `json:"auth_mode"`
	AllowedDomains    []string `json:"allowed_domains"`
	OTPTTLSeconds     int      `json:"otp_ttl_seconds"`
	SessionTTLSeconds int      `json:"session_ttl_seconds"`
}

type customerOTPRequest struct {
	TenantID    string `json:"tenant_id"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	Locale      string `json:"locale"`
}

type customerOTPVerifyRequest struct {
	TenantID    string `json:"tenant_id"`
	ChallengeID string `json:"challenge_id"`
	OTP         string `json:"otp"`
}

type customerRefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

func (s *server) getCustomerAuthSettings(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	row, err := s.store.GetCustomerAuthSettings(r.Context(), tenantID)
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (s *server) putCustomerAuthSettings(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body customerAuthSettingsBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	row, err := s.store.PutCustomerAuthSettings(r.Context(), tenantID, store.CustomerAuthSettingsInput{
		Enabled: body.Enabled, AuthMode: body.AuthMode, AllowedDomains: body.AllowedDomains,
		OTPTTLSeconds: body.OTPTTLSeconds, SessionTTLSeconds: body.SessionTTLSeconds,
	})
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (s *server) requestCustomerOTP(w http.ResponseWriter, r *http.Request) {
	var req customerOTPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	tenantID := s.publicCustomerTenantID(r)
	email, err := store.NormalizeCustomerEmail(req.Email)
	if err != nil || email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "valid email is required", "code": "validation_error"})
		return
	}
	settings, err := s.store.GetCustomerAuthSettings(r.Context(), tenantID)
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	if !settings.Enabled {
		writeCustomerAuthError(w, store.ErrCustomerAuthDisabled)
		return
	}
	if err := s.checkCustomerAuthDomain(r.Context(), tenantID, email, settings); err != nil {
		s.store.RecordCustomerAuthEvent(r.Context(), tenantID, "", email, "customer.auth.otp_denied", clientIP(r), r.UserAgent(), nil)
		writeCustomerAuthError(w, err)
		return
	}

	customer, matched, err := s.resolveOrCreateCustomerForOTP(r.Context(), tenantID, email, req.DisplayName, req.Locale)
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}

	code := newOTPCode()
	codeHash := s.customerOTPHash(tenantID, email, code)
	chal, err := s.store.CreateCustomerOTPChallenge(r.Context(), tenantID, email, customer.ID, codeHash, time.Duration(settings.OTPTTLSeconds)*time.Second, map[string]any{
		"ip": clientIP(r), "user_agent": r.UserAgent(),
	})
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	s.sendCustomerOTPEmail(r.Context(), email, code, settings.OTPTTLSeconds)
	s.store.RecordCustomerAuthEvent(r.Context(), tenantID, customer.ID, email, "customer.auth.otp_requested", clientIP(r), r.UserAgent(), nil)

	writeJSON(w, http.StatusAccepted, map[string]any{
		"challenge_id": chal.ID,
		"status":       "otp_sent",
		"delivery": map[string]any{
			"channel": "email",
			"to":      maskEmail(email),
		},
		"expires_in":   settings.OTPTTLSeconds,
		"resend_after": 60,
		"customer_hint": map[string]any{
			"matched_existing_customer":   matched,
			"requires_profile_completion": strings.TrimSpace(customer.DisplayName) == "",
			"email_domain_policy":         "allowed",
		},
	})
}

func (s *server) verifyCustomerOTP(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil || !s.auth.TokensEnabled() {
		writeError(w, http.StatusServiceUnavailable, "auth is not configured")
		return
	}
	var req customerOTPVerifyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	tenantID := s.publicCustomerTenantID(r)
	challengeID := strings.TrimSpace(req.ChallengeID)
	otp := normalizeOTP(req.OTP)
	if challengeID == "" || otp == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "challenge_id and otp are required", "code": "validation_error"})
		return
	}
	chal, err := s.store.GetCustomerOTPChallenge(r.Context(), tenantID, challengeID)
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	codeHash := s.customerOTPHash(tenantID, chal.Email, otp)
	chal, err = s.store.VerifyCustomerOTPChallenge(r.Context(), tenantID, challengeID, codeHash, 5)
	if err != nil {
		s.store.RecordCustomerAuthEvent(r.Context(), tenantID, "", chal.Email, "customer.auth.otp_failed", clientIP(r), r.UserAgent(), nil)
		writeCustomerAuthError(w, err)
		return
	}
	customer, err := s.store.GetCustomer(r.Context(), tenantID, chal.CustomerID)
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	if customer.Status != "active" {
		writeCustomerAuthError(w, store.ErrCustomerAuthForbidden)
		return
	}
	if err := s.store.UpsertCustomerAuthIdentity(r.Context(), tenantID, customer.ID, chal.Email); err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	rawRefresh, refreshHash, err := auth.NewRefreshToken()
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	settings, _ := s.store.GetCustomerAuthSettings(r.Context(), tenantID)
	sessionTTL := time.Duration(settings.SessionTTLSeconds) * time.Second
	session, err := s.store.CreateCustomerSession(auditctx.WithActor(r.Context(), customer.ID), tenantID, customer.ID, refreshHash, sessionTTL)
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	access, expiresIn, err := s.auth.IssueAccessForPrincipal(customer.ID, customer.Email, auth.RoleCustomer, tenantID)
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	s.store.RecordCustomerAuthEvent(r.Context(), tenantID, customer.ID, customer.Email, "customer.auth.logged_in", clientIP(r), r.UserAgent(), map[string]any{"session_id": session.ID})
	writeJSON(w, http.StatusOK, map[string]any{
		"status":             "authenticated",
		"access_token":       access,
		"refresh_token":      rawRefresh,
		"token_type":         "Bearer",
		"expires_in":         expiresIn,
		"refresh_expires_in": int(sessionTTL.Seconds()),
		"customer":           customerAuthProfile(*customer, tenantID),
	})
}

func (s *server) refreshCustomerAuth(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil || !s.auth.TokensEnabled() {
		writeError(w, http.StatusServiceUnavailable, "auth is not configured")
		return
	}
	var req customerRefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := auth.ValidateRefreshToken(req.RefreshToken); err != nil {
		writeCustomerAuthError(w, store.ErrCustomerSessionInvalid)
		return
	}
	hash := auth.HashRefreshToken(req.RefreshToken)
	session, err := s.store.GetCustomerSessionByRefreshHash(r.Context(), hash)
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	customer, err := s.store.GetCustomer(r.Context(), session.TenantID, session.CustomerID)
	if err != nil || customer.Status != "active" {
		writeCustomerAuthError(w, store.ErrCustomerSessionInvalid)
		return
	}
	_ = s.store.RevokeCustomerSessionByRefreshHash(r.Context(), hash)
	rawRefresh, refreshHash, err := auth.NewRefreshToken()
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	settings, _ := s.store.GetCustomerAuthSettings(r.Context(), session.TenantID)
	sessionTTL := time.Duration(settings.SessionTTLSeconds) * time.Second
	newSession, err := s.store.CreateCustomerSession(auditctx.WithActor(r.Context(), customer.ID), session.TenantID, customer.ID, refreshHash, sessionTTL)
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	access, expiresIn, err := s.auth.IssueAccessForPrincipal(customer.ID, customer.Email, auth.RoleCustomer, session.TenantID)
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	s.store.RecordCustomerAuthEvent(r.Context(), session.TenantID, customer.ID, customer.Email, "customer.auth.token_refreshed", clientIP(r), r.UserAgent(), map[string]any{"session_id": newSession.ID})
	writeJSON(w, http.StatusOK, map[string]any{
		"status":             "authenticated",
		"access_token":       access,
		"refresh_token":      rawRefresh,
		"token_type":         "Bearer",
		"expires_in":         expiresIn,
		"refresh_expires_in": int(sessionTTL.Seconds()),
		"customer":           customerAuthProfile(*customer, session.TenantID),
	})
}

func (s *server) logoutCustomerAuth(w http.ResponseWriter, r *http.Request) {
	var req customerRefreshRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	if strings.TrimSpace(req.RefreshToken) != "" {
		_ = s.store.RevokeCustomerSessionByRefreshHash(r.Context(), auth.HashRefreshToken(req.RefreshToken))
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *server) customerMe(w http.ResponseWriter, r *http.Request) {
	ac, err := s.parseCustomerBearer(r)
	if err != nil {
		writeAuthHandlerError(w, err)
		return
	}
	customer, err := s.store.GetCustomer(r.Context(), ac.TenantID, ac.UserID)
	if err != nil || customer.Status != "active" {
		writeAuthHandlerError(w, auth.ErrUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"customer": customerAuthProfile(*customer, ac.TenantID)})
}

func (s *server) parseCustomerBearer(r *http.Request) (auth.AuthContext, error) {
	if s.auth == nil || !s.auth.TokensEnabled() {
		return auth.AuthContext{}, auth.ErrNotConfigured
	}
	ac, err := s.auth.ParseBearer(r.Header.Get("Authorization"))
	if err != nil {
		return auth.AuthContext{}, err
	}
	if ac.Role != auth.RoleCustomer {
		return auth.AuthContext{}, auth.ErrForbidden
	}
	return ac, nil
}

func (s *server) publicCustomerTenantID(r *http.Request) string {
	return auth.ResolveTenant(r.Context(), r.Header.Get("X-Tenant-Id"), s.cfg.AuthDisabled, s.cfg.DemoTenantID)
}

func (s *server) checkCustomerAuthDomain(ctx context.Context, tenantID, email string, settings store.CustomerAuthSettings) error {
	at := strings.LastIndex(email, "@")
	if at < 0 {
		return store.ErrCustomerAuthForbidden
	}
	domain := email[at+1:]
	if rule, err := s.store.FindCustomerDomainRule(ctx, tenantID, domain); err == nil && rule.Active {
		if rule.Policy == "deny" {
			return store.ErrCustomerAuthForbidden
		}
	}
	if len(settings.AllowedDomains) == 0 {
		return nil
	}
	for _, allowed := range settings.AllowedDomains {
		if strings.EqualFold(allowed, domain) {
			return nil
		}
	}
	return store.ErrCustomerAuthForbidden
}

func (s *server) resolveOrCreateCustomerForOTP(ctx context.Context, tenantID, email, displayName, locale string) (*store.Customer, bool, error) {
	if customer, err := s.store.FindCustomerByEmail(ctx, tenantID, email); err == nil {
		return customer, true, nil
	} else if !errors.Is(err, store.ErrCustomerNotFound) {
		return nil, false, err
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = email
		if at := strings.Index(email, "@"); at > 0 {
			displayName = email[:at]
		}
	}
	result, err := s.store.UpsertCustomer(ctx, tenantID, store.CustomerInput{
		Email: email, DisplayName: displayName, Locale: locale, Source: "self_claim", Status: "active", Metadata: map[string]any{"claimed_by": "email_otp"},
	})
	if err != nil {
		return nil, false, err
	}
	return result.Customer, false, nil
}

func (s *server) customerOTPHash(tenantID, email, otp string) string {
	secret := strings.TrimSpace(s.cfg.JWTSecret)
	if secret == "" {
		secret = strings.TrimSpace(s.cfg.ResendAPIKey)
	}
	if secret == "" {
		secret = "monti-jarvis-dev-customer-otp-secret"
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(tenantID + "\x00" + strings.ToLower(strings.TrimSpace(email)) + "\x00" + normalizeOTP(otp)))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *server) sendCustomerOTPEmail(ctx context.Context, email, code string, ttlSeconds int) {
	subject := "Your Monti sign-in code"
	htmlBody := fmt.Sprintf(`<p>Your Monti sign-in code is:</p><p style="font-size:28px;font-weight:700;letter-spacing:6px">%s</p><p>This code expires in %d minutes.</p>`, html.EscapeString(code), max(1, ttlSeconds/60))
	if s.mailer == nil || !s.mailer.Enabled() {
		log.Printf("mailer warning: customer OTP email skipped for %s (resend disabled)", email)
		return
	}
	mailCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := s.mailer.Send(mailCtx, email, subject, htmlBody); err != nil {
		log.Printf("mailer warning: customer OTP email to %s: %v", email, err)
		return
	}
	log.Printf("mailer: customer OTP email sent to %s", email)
}

func customerAuthProfile(c store.Customer, tenantID string) map[string]any {
	return map[string]any{
		"id": c.ID, "tenant_id": tenantID, "display_name": c.DisplayName, "email": c.Email,
		"tier_id": c.TierID, "group_ids": c.GroupIDs, "locale": c.Locale, "role": "customer",
	}
}

func writeCustomerAuthError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrCustomerAuthDisabled):
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "customer auth disabled", "code": "customer_auth_disabled"})
	case errors.Is(err, store.ErrCustomerAuthForbidden):
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "email domain is not allowed", "code": "domain_forbidden"})
	case errors.Is(err, store.ErrOTPInvalid), errors.Is(err, store.ErrCustomerSessionInvalid):
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid credentials", "code": "invalid_credentials"})
	case errors.Is(err, store.ErrOTPExpired):
		writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "otp expired", "code": "otp_expired"})
	case errors.Is(err, store.ErrCustomerNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "customer not found", "code": "not_found"})
	default:
		if err != nil && (strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must be")) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "code": "validation_error"})
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
	}
}

func newOTPCode() string {
	var b [4]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%06d", time.Now().UnixNano()%1000000)
	}
	n := int(b[0])<<24 | int(b[1])<<16 | int(b[2])<<8 | int(b[3])
	if n < 0 {
		n = -n
	}
	return fmt.Sprintf("%06d", n%1000000)
}

func normalizeOTP(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "")
	value = strings.ReplaceAll(value, "-", "")
	return value
}

func maskEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) != 2 {
		return "***"
	}
	local := parts[0]
	if len(local) <= 2 {
		return local[:1] + "***@" + parts[1]
	}
	return local[:1] + "***" + local[len(local)-1:] + "@" + parts[1]
}

func clientIP(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		return strings.TrimSpace(strings.Split(forwarded, ",")[0])
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}
