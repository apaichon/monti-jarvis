package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/store"
	"github.com/libra/monti-jarvis/internal/workforce"
)

type mobileCreateCallRequest struct {
	AvatarID string `json:"avatar_id"`
	Locale   string `json:"locale"`
}

type mobileRatingRequest struct {
	Score  int    `json:"score"`
	Review string `json:"review"`
}

type mobileCacheEntry struct {
	ExpiresAt time.Time
	Body      []byte
}

var mobileIdempotency sync.Map

func (s *server) mobileBootstrap(w http.ResponseWriter, r *http.Request) {
	tenantID, customer, settings, ok := s.mobileContext(w, r, false)
	if !ok {
		return
	}
	tenantSettings, err := s.store.GetOrCreateTenantSettings(r.Context(), tenantID)
	if err != nil {
		writeMobileError(w, http.StatusBadGateway, "bootstrap_unavailable")
		return
	}
	limits, err := s.store.GetOrCreateTenantCallLimits(r.Context(), tenantID)
	if err != nil {
		writeMobileError(w, http.StatusBadGateway, "bootstrap_unavailable")
		return
	}
	agents, err := s.customerWorkforceAgents(r, tenantID)
	if err != nil {
		writeMobileError(w, http.StatusBadGateway, "bootstrap_unavailable")
		return
	}
	locale := ""
	if customer != nil {
		locale = customer.Locale
	}
	if strings.TrimSpace(locale) == "" {
		locale = tenantSettings.AIReplyLocale
	}
	if strings.TrimSpace(locale) == "" {
		locale = tenantSettings.Locale
	}
	locale, _ = store.NormalizeLocale(locale)
	maxCallSeconds := settings.CustomerMaxCallSeconds
	if maxCallSeconds == 0 && limits.MaxMinutesPerCall > 0 {
		maxCallSeconds = limits.MaxMinutesPerCall * 60
	}
	dailyLimitSeconds := settings.CustomerDailyCallSeconds
	if dailyLimitSeconds == 0 && limits.MaxCallMinutesPerDay > 0 {
		dailyLimitSeconds = limits.MaxCallMinutesPerDay * 60
	}
	var quotaSummary *store.CustomerUsageSummary
	if customer != nil {
		if summary, summaryErr := s.store.CustomerUsageSummary(r.Context(), tenantID, customer.ID, dailyLimitSeconds, maxCallSeconds, time.Now()); summaryErr == nil {
			quotaSummary = &summary
		}
	}
	resetAt := time.Now().UTC().Add(24 * time.Hour)
	if loc, locErr := time.LoadLocation(s.store.TenantTimezone(r.Context(), tenantID)); locErr == nil {
		now := time.Now().In(loc)
		resetAt = time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, loc).UTC()
	}
	selected := ""
	if len(agents) > 0 {
		selected = agents[0].ID
	}
	avatarRows := make([]map[string]any, 0, len(agents))
	for _, agent := range agents {
		avatarRows = append(avatarRows, mobileAvatar(agent))
	}
	limitsOut := map[string]any{
		"max_call_seconds":        maxCallSeconds,
		"daily_limit_seconds":     dailyLimitSeconds,
		"daily_remaining_seconds": nil,
		"warning_at_seconds":      10,
		"reset_at":                resetAt,
	}
	if quotaSummary != nil {
		limitsOut["daily_remaining_seconds"] = quotaSummary.DailyRemainingSeconds
		limitsOut["state"] = quotaSummary.State
	} else {
		limitsOut["state"] = "quota_available"
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"version": "v1",
		"tenant":  map[string]any{"id": tenantID, "display_name": tenantSettings.DisplayName, "slug": tenantID},
		"auth": map[string]any{
			"enabled": settings.Enabled, "mode": settings.AuthMode,
			"required_for_call": settings.Enabled && settings.RequireAuthForWorkforce,
			"otp_ttl_seconds":   settings.OTPTTLSeconds, "session_ttl_seconds": settings.SessionTTLSeconds,
		},
		"locale":  map[string]any{"language": locale, "timezone": s.store.TenantTimezone(r.Context(), tenantID)},
		"avatars": avatarRows, "default_avatar_id": selected, "limits": limitsOut,
	})
}

func (s *server) mobileCreateCall(w http.ResponseWriter, r *http.Request) {
	var req mobileCreateCallRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeMobileError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	tenantID, customer, settings, ok := s.mobileContext(w, r, true)
	if !ok {
		return
	}
	agents, err := s.customerWorkforceAgents(r, tenantID)
	if err != nil {
		writeMobileError(w, http.StatusBadGateway, "avatar_unavailable")
		return
	}
	agentID := strings.TrimSpace(req.AvatarID)
	if agentID == "" && len(agents) > 0 {
		agentID = agents[0].ID
	}
	if !containsAgent(agents, agentID) {
		writeMobileError(w, http.StatusForbidden, "avatar_unavailable")
		return
	}
	if settings.CustomerMaxCallSeconds > 0 && settings.CustomerDailyCallSeconds > 0 && settings.CustomerMaxCallSeconds > settings.CustomerDailyCallSeconds {
		writeMobileError(w, http.StatusForbidden, "call_duration_limit_exceeded")
		return
	}
	if customer != nil && settings.CustomerDailyCallSeconds > 0 {
		summary, summaryErr := s.store.CustomerUsageSummary(r.Context(), tenantID, customer.ID, settings.CustomerDailyCallSeconds, settings.CustomerMaxCallSeconds, time.Now())
		if summaryErr != nil {
			writeMobileError(w, http.StatusBadGateway, "quota_unavailable")
			return
		}
		if summary.State == "quota_exhausted" {
			writeMobileError(w, http.StatusTooManyRequests, "customer_quota_exhausted")
			return
		}
	}
	key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))
	if key == "" {
		writeMobileError(w, http.StatusBadRequest, "idempotency_key_required")
		return
	}
	subject := "anonymous"
	if customer != nil {
		subject = customer.ID
	}
	cacheKey := mobileIdempotencyKey(tenantID, subject, "create", key)
	if body, ok := mobileCached(cacheKey); ok {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write(body)
		return
	}
	if s.calls == nil {
		writeMobileError(w, http.StatusServiceUnavailable, "call_service_unavailable")
		return
	}
	sessionID := newID()
	session, err := s.calls.CreateForTenant(r.Context(), tenantID, sessionID, "monti-"+sessionID[:12])
	if err != nil {
		writeMobileError(w, http.StatusBadGateway, "call_unavailable")
		return
	}
	if s.store != nil {
		customerID := ""
		if customer != nil {
			customerID = customer.ID
		}
		_ = s.store.UpdateCallSessionContext(r.Context(), session.ID, customerID, agentID)
		if customer != nil {
			_ = s.store.RecordCustomerUsage(r.Context(), tenantID, customer.ID, session.ID, agentID, "voice", 0, "reserved", "")
		}
	}
	response := map[string]any{"call_id": session.ID, "status": session.Status, "avatar_id": agentID, "started_at": session.StartedAt}
	body, _ := json.Marshal(response)
	mobileStoreCache(cacheKey, body)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_, _ = w.Write(body)
}

func (s *server) mobileGetCall(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.mobileOwnedCall(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, mobileCallPayload(ctx))
}

func (s *server) mobileTranscript(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.mobileOwnedCall(w, r); !ok {
		return
	}
	turns, err := s.calls.ListTurns(r.Context(), r.PathValue("call_id"))
	if err != nil {
		writeMobileError(w, http.StatusNotFound, "call_not_found")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"turns": turns})
}

func (s *server) mobileEndCall(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.mobileOwnedCall(w, r)
	if !ok {
		return
	}
	if ctx.Session.Status == "ended" {
		writeJSON(w, http.StatusOK, mobileCallPayload(ctx))
		return
	}
	if s.calls == nil {
		writeMobileError(w, http.StatusServiceUnavailable, "call_service_unavailable")
		return
	}
	session, err := s.calls.End(r.Context(), ctx.Session.ID)
	if err != nil {
		writeMobileError(w, http.StatusConflict, "call_already_ended")
		return
	}
	if s.store != nil {
		turns, _ := s.calls.ListTurns(r.Context(), ctx.Session.ID)
		payload := map[string]any{"call": session, "turns": turns}
		if object, archiveErr := s.store.ArchiveConversationTranscriptForChannel(r.Context(), session.TenantID, session.ID, "voice", payload, ""); archiveErr == nil {
			s.projectCallCenterRecord(r.Context(), session.TenantID, object.ConversationRecordID)
		}
	}
	ctx.Session = session
	writeJSON(w, http.StatusOK, mobileCallPayload(ctx))
}

func (s *server) mobileRateCall(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.mobileOwnedCall(w, r)
	if !ok {
		return
	}
	var req mobileRatingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeMobileError(w, http.StatusBadRequest, "invalid_json")
		return
	}
	if req.Score < 1 || req.Score > 5 {
		writeMobileError(w, http.StatusBadRequest, "rating_score_invalid")
		return
	}
	if len([]rune(req.Review)) > 2000 {
		writeMobileError(w, http.StatusBadRequest, "rating_review_too_long")
		return
	}
	if err := s.store.SaveConversationRating(r.Context(), ctx.Session.TenantID, ctx.Session.ID, req.Score, req.Review); err != nil {
		writeMobileError(w, http.StatusBadGateway, "rating_unavailable")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"status": "saved", "score": req.Score})
}

func (s *server) mobileVoiceWS(w http.ResponseWriter, r *http.Request) {
	ctx, ok := s.mobileOwnedCall(w, r)
	if !ok {
		return
	}
	if ctx.Session.Status != "active" {
		writeMobileError(w, http.StatusConflict, "call_not_active")
		return
	}
	req := r.Clone(r.Context())
	query := req.URL.Query()
	query.Set("tenant_id", ctx.Session.TenantID)
	query.Set("agent", ctx.AvatarID)
	if locale := r.URL.Query().Get("locale"); locale != "" {
		query.Set("locale", locale)
	}
	req.URL.RawQuery = query.Encode()
	s.voiceWithPackageQuota(w, req, ctx.Session.TenantID)
}

func (s *server) mobileContext(w http.ResponseWriter, r *http.Request, requireCallAuth bool) (string, *store.Customer, store.CustomerAuthSettings, bool) {
	requestContext := r.Context()
	if _, ok := auth.FromContext(requestContext); !ok && !s.cfg.AuthDisabled {
		if token := strings.TrimSpace(r.URL.Query().Get("access_token")); token != "" && s.auth != nil {
			if ac, parseErr := s.auth.ParseBearer("Bearer " + token); parseErr == nil {
				requestContext = auth.WithContext(requestContext, ac)
				r = r.WithContext(requestContext)
			} else {
				writeMobileError(w, http.StatusUnauthorized, "unauthorized")
				return "", nil, store.CustomerAuthSettings{}, false
			}
		}
	}
	tenantID := s.publicCustomerTenantID(r)
	settings, err := s.store.GetCustomerAuthSettings(r.Context(), tenantID)
	if err != nil {
		writeMobileError(w, http.StatusBadGateway, "auth_policy_unavailable")
		return "", nil, settings, false
	}
	var customer *store.Customer
	if ac, ok := auth.FromContext(r.Context()); ok && ac.Role == auth.RoleCustomer && ac.TenantID == tenantID {
		if c, getErr := s.store.GetCustomer(r.Context(), tenantID, ac.UserID); getErr == nil && c.Status == "active" {
			customer = c
		}
	} else if strings.TrimSpace(r.Header.Get("Authorization")) != "" && !s.cfg.AuthDisabled {
		writeMobileError(w, http.StatusUnauthorized, "unauthorized")
		return "", nil, settings, false
	}
	if requireCallAuth && settings.Enabled && settings.RequireAuthForWorkforce && customer == nil {
		writeMobileError(w, http.StatusUnauthorized, "customer_auth_required")
		return "", nil, settings, false
	}
	return tenantID, customer, settings, true
}

func (s *server) mobileOwnedCall(w http.ResponseWriter, r *http.Request) (store.CallSessionContext, bool) {
	if s.store == nil || s.calls == nil {
		writeMobileError(w, http.StatusServiceUnavailable, "call_service_unavailable")
		return store.CallSessionContext{}, false
	}
	callID := strings.TrimSpace(r.PathValue("call_id"))
	ctx, err := s.store.GetCallSessionContext(r.Context(), callID)
	if err != nil {
		writeMobileError(w, http.StatusNotFound, "call_not_found")
		return store.CallSessionContext{}, false
	}
	tenantID, customer, _, ok := s.mobileContext(w, r, false)
	if !ok || tenantID != ctx.Session.TenantID {
		if ok {
			writeMobileError(w, http.StatusNotFound, "call_not_found")
		}
		return store.CallSessionContext{}, false
	}
	if customer != nil {
		if ctx.CustomerID != customer.ID {
			writeMobileError(w, http.StatusNotFound, "call_not_found")
			return store.CallSessionContext{}, false
		}
	} else if ctx.CustomerID != "" && !s.cfg.AuthDisabled {
		writeMobileError(w, http.StatusNotFound, "call_not_found")
		return store.CallSessionContext{}, false
	}
	return ctx, true
}

func mobileCallPayload(ctx store.CallSessionContext) map[string]any {
	return map[string]any{"call_id": ctx.Session.ID, "status": ctx.Session.Status, "avatar_id": ctx.AvatarID, "started_at": ctx.Session.StartedAt, "ended_at": ctx.Session.EndedAt}
}

func mobileAvatar(agent workforce.Agent) map[string]any {
	return map[string]any{"id": agent.ID, "name": agent.Name, "role": agent.Role, "trait": agent.Trait, "voice": agent.Voice, "image": agent.Image, "greeting": agent.Greeting}
}

func containsAgent(agents []workforce.Agent, id string) bool {
	for _, agent := range agents {
		if agent.ID == id {
			return true
		}
	}
	return false
}

func mobileIdempotencyKey(tenant, subject, route, key string) string {
	h := sha256.Sum256([]byte(strings.Join([]string{tenant, subject, route, key}, "\x00")))
	return "mobile:" + hex.EncodeToString(h[:])
}

func mobileCached(key string) ([]byte, bool) {
	value, ok := mobileIdempotency.Load(key)
	if !ok {
		return nil, false
	}
	entry := value.(mobileCacheEntry)
	if time.Now().After(entry.ExpiresAt) {
		mobileIdempotency.Delete(key)
		return nil, false
	}
	return append([]byte(nil), entry.Body...), true
}

func mobileStoreCache(key string, body []byte) {
	mobileIdempotency.Store(key, mobileCacheEntry{ExpiresAt: time.Now().Add(10 * time.Minute), Body: append([]byte(nil), body...)})
}

func writeMobileError(w http.ResponseWriter, status int, code string) {
	writeJSON(w, status, map[string]any{"error": code, "code": code})
}
