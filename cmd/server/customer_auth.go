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
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/auditctx"
	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/resend"
	"github.com/libra/monti-jarvis/internal/store"
)

type customerAuthSettingsBody struct {
	Enabled                       *bool    `json:"enabled"`
	AuthMode                      string   `json:"auth_mode"`
	AllowedDomains                []string `json:"allowed_domains"`
	OTPTTLSeconds                 int      `json:"otp_ttl_seconds"`
	SessionTTLSeconds             int      `json:"session_ttl_seconds"`
	RequireAuthForWorkforce       *bool    `json:"require_auth_for_workforce"`
	CustomerDailyCallSeconds      *int     `json:"customer_daily_call_seconds"`
	CustomerMaxCallSeconds        *int     `json:"customer_max_call_seconds"`
	AutoRegisterOnConversationOTP *bool    `json:"auto_register_on_conversation_otp"`
}

type customerOTPRequest struct {
	TenantID     string                       `json:"tenant_id"`
	Email        string                       `json:"email"`
	DisplayName  string                       `json:"display_name"`
	Locale       string                       `json:"locale"`
	Purpose      string                       `json:"purpose"` // optional: conversation | default
	Notification *customerOTPNotificationBody `json:"notification,omitempty"`
}

type customerOTPNotificationBody struct {
	Platform   string `json:"platform"`
	PushToken  string `json:"push_token"`
	AppVersion string `json:"app_version"`
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
		RequireAuthForWorkforce:       body.RequireAuthForWorkforce,
		CustomerDailyCallSeconds:      body.CustomerDailyCallSeconds,
		CustomerMaxCallSeconds:        body.CustomerMaxCallSeconds,
		AutoRegisterOnConversationOTP: body.AutoRegisterOnConversationOTP,
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

	// Auto-register only creates on verify when flag is on; request may attach
	// existing customer or leave customer_id empty for first-time auto-register.
	customer, matched, err := s.resolveCustomerForOTP(r.Context(), tenantID, email, req.DisplayName, req.Locale, settings.AutoRegisterOnConversationOTP)
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}

	code := newOTPCode()
	codeHash := s.customerOTPHash(tenantID, email, code)
	purpose := strings.ToLower(strings.TrimSpace(req.Purpose))
	if purpose == "" {
		purpose = "conversation"
	}
	metadata := map[string]any{
		"ip":         clientIP(r),
		"user_agent": r.UserAgent(),
		"purpose":    purpose,
		"auto_register": settings.AutoRegisterOnConversationOTP,
		"display_name":  strings.TrimSpace(req.DisplayName),
		"locale":        strings.TrimSpace(req.Locale),
	}
	notification := map[string]any{"status": "not_configured"}
	if req.Notification != nil {
		platform := strings.ToLower(strings.TrimSpace(req.Notification.Platform))
		pushToken := strings.TrimSpace(req.Notification.PushToken)
		if (platform != "ios" && platform != "android") || pushToken == "" || len(pushToken) > 4096 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "valid iOS or Android notification details are required", "code": "notification_invalid"})
			return
		}
		// Keep only a digest in the challenge metadata. Provider delivery is
		// intentionally disabled until APNs/FCM credentials are configured.
		digest := sha256.Sum256([]byte(pushToken))
		metadata["notification"] = map[string]any{"platform": platform, "token_sha256": hex.EncodeToString(digest[:]), "app_version": strings.TrimSpace(req.Notification.AppVersion)}
		notification = map[string]any{"status": "not_configured", "platform": platform}
	}
	customerID := ""
	requiresProfile := true
	if customer != nil {
		customerID = customer.ID
		requiresProfile = strings.TrimSpace(customer.DisplayName) == ""
	}
	chal, err := s.store.CreateCustomerOTPChallenge(r.Context(), tenantID, email, customerID, codeHash, time.Duration(settings.OTPTTLSeconds)*time.Second, metadata)
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	s.sendCustomerOTPEmail(r.Context(), email, code, settings.OTPTTLSeconds)
	s.store.RecordCustomerAuthEvent(r.Context(), tenantID, customerID, email, "customer.auth.otp_requested", clientIP(r), r.UserAgent(), nil)

	writeJSON(w, http.StatusAccepted, map[string]any{
		"challenge_id": chal.ID,
		"status":       "otp_sent",
		"delivery": map[string]any{
			"channel":      "email",
			"to":           maskEmail(email),
			"notification": notification,
		},
		"expires_in":   settings.OTPTTLSeconds,
		"resend_after": 60,
		"customer_hint": map[string]any{
			"matched_existing_customer":   matched,
			"requires_profile_completion": requiresProfile,
			"email_domain_policy":         "allowed",
			"will_auto_register":          !matched && settings.AutoRegisterOnConversationOTP,
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
	settings, _ := s.store.GetCustomerAuthSettings(r.Context(), tenantID)
	var customer *store.Customer
	if strings.TrimSpace(chal.CustomerID) != "" {
		customer, err = s.store.GetCustomer(r.Context(), tenantID, chal.CustomerID)
		if err != nil {
			writeCustomerAuthError(w, err)
			return
		}
	} else if settings.AutoRegisterOnConversationOTP {
		// First-time conversation auto-register: create customer only after OTP succeeds.
		displayName := ""
		locale := ""
		if chal.Metadata != nil {
			if v, ok := chal.Metadata["display_name"].(string); ok {
				displayName = v
			}
			if v, ok := chal.Metadata["locale"].(string); ok {
				locale = v
			}
		}
		customer, _, err = s.resolveCustomerForOTP(r.Context(), tenantID, chal.Email, displayName, locale, true)
		if err != nil {
			writeCustomerAuthError(w, err)
			return
		}
	} else {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "customer not found", "code": "customer_not_found"})
		return
	}
	if customer == nil || customer.Status != "active" {
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
	setRefreshCookie(w, s.cfg, customerRefreshCookie, "/api/customer/auth", rawRefresh, int(sessionTTL.Seconds()))
	response := map[string]any{
		"status":             "authenticated",
		"access_token":       access,
		"token_type":         "Bearer",
		"expires_in":         expiresIn,
		"refresh_expires_in": int(sessionTTL.Seconds()),
		"customer":           customerAuthProfile(*customer, tenantID),
	}
	if r.Header.Get("X-Monti-Client") == "mobile" {
		response["refresh_token"] = rawRefresh
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *server) refreshCustomerAuth(w http.ResponseWriter, r *http.Request) {
	if s.auth == nil || !s.auth.TokensEnabled() {
		writeError(w, http.StatusServiceUnavailable, "auth is not configured")
		return
	}
	var req customerRefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	refreshToken := requestRefreshToken(r, customerRefreshCookie, req.RefreshToken)
	if err := auth.ValidateRefreshToken(refreshToken); err != nil {
		writeCustomerAuthError(w, store.ErrCustomerSessionInvalid)
		return
	}
	hash := auth.HashRefreshToken(refreshToken)
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
	setRefreshCookie(w, s.cfg, customerRefreshCookie, "/api/customer/auth", rawRefresh, int(sessionTTL.Seconds()))
	response := map[string]any{
		"status":             "authenticated",
		"access_token":       access,
		"token_type":         "Bearer",
		"expires_in":         expiresIn,
		"refresh_expires_in": int(sessionTTL.Seconds()),
		"customer":           customerAuthProfile(*customer, session.TenantID),
	}
	if r.Header.Get("X-Monti-Client") == "mobile" {
		response["refresh_token"] = rawRefresh
	}
	writeJSON(w, http.StatusOK, response)
}

func (s *server) logoutCustomerAuth(w http.ResponseWriter, r *http.Request) {
	var req customerRefreshRequest
	_ = json.NewDecoder(r.Body).Decode(&req)
	refreshToken := requestRefreshToken(r, customerRefreshCookie, req.RefreshToken)
	if strings.TrimSpace(refreshToken) != "" {
		_ = s.store.RevokeCustomerSessionByRefreshHash(r.Context(), auth.HashRefreshToken(refreshToken))
	}
	clearRefreshCookie(w, s.cfg, customerRefreshCookie, "/api/customer/auth")
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
	if tenantID, _, err := s.requestTenantContext(r); err == nil && tenantID != "" {
		return tenantID
	}
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

// resolveCustomerForOTP returns an existing customer, or creates one only when
// autoRegister is true. When autoRegister is false and the email is unknown,
// returns ErrCustomerNotFound (no row created).
func (s *server) resolveCustomerForOTP(ctx context.Context, tenantID, email, displayName, locale string, autoRegister bool) (*store.Customer, bool, error) {
	if customer, err := s.store.FindCustomerByEmail(ctx, tenantID, email); err == nil {
		return customer, true, nil
	} else if !errors.Is(err, store.ErrCustomerNotFound) {
		return nil, false, err
	}
	if !autoRegister {
		return nil, false, store.ErrCustomerNotFound
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		displayName = email
		if at := strings.Index(email, "@"); at > 0 {
			displayName = email[:at]
		}
	}
	result, err := s.store.UpsertCustomer(ctx, tenantID, store.CustomerInput{
		Email: email, DisplayName: displayName, Locale: locale, Source: "self_claim", Status: "active",
		Metadata: map[string]any{"claimed_by": "email_otp", "auto_register": true},
	})
	if err != nil {
		return nil, false, err
	}
	return result.Customer, false, nil
}

// resolveOrCreateCustomerForOTP is kept for callers that always auto-create (legacy).
func (s *server) resolveOrCreateCustomerForOTP(ctx context.Context, tenantID, email, displayName, locale string) (*store.Customer, bool, error) {
	return s.resolveCustomerForOTP(ctx, tenantID, email, displayName, locale, true)
}

func (s *server) publicCustomerAuthPolicy(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	if tenantID == "" {
		tenantID = s.publicCustomerTenantID(r)
	}
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "tenant_id is required")
		return
	}
	settings, err := s.store.GetCustomerAuthSettings(r.Context(), tenantID)
	if err != nil {
		writeCustomerAuthError(w, err)
		return
	}
	// Safe public booleans only — no domain list.
	writeJSON(w, http.StatusOK, map[string]any{
		"tenant_id": tenantID,
		"enabled":   settings.Enabled,
		"auth_mode": settings.AuthMode,
		"require_auth_for_workforce":      settings.RequireAuthForWorkforce || settings.Enabled,
		"auto_register_on_conversation_otp": settings.AutoRegisterOnConversationOTP,
	})
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
	subject, htmlBody := resend.CustomerOTPEmail(code, ttlSeconds, resend.BrandLogoURL(s.cfg.PublicBaseURL))
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
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "customer not found", "code": "customer_not_found"})
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
