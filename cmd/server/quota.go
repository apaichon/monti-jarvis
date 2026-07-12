package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/quota"
	"github.com/libra/monti-jarvis/internal/store"
)

// writeQuotaError maps quota.Error to HTTP status + structured JSON.
func writeQuotaError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	var qe *quota.Error
	if errors.As(err, &qe) {
		status := http.StatusTooManyRequests
		switch qe.Code {
		case "feature_disabled":
			status = http.StatusForbidden
		case "no_entitlement":
			status = http.StatusForbidden
		case "daily_call_limit", "per_call_limit":
			status = http.StatusTooManyRequests
		case "rate_limited":
			status = http.StatusTooManyRequests
			w.Header().Set("Retry-After", "60")
		default:
			status = http.StatusTooManyRequests
		}
		writeJSON(w, status, map[string]any{
			"error":     qe.Error(),
			"code":      qe.Code,
			"dimension": qe.Dimension,
			"limit":     qe.Limit,
			"usage":     qe.Usage,
		})
		return
	}
	if errors.Is(err, quota.ErrLimitExceeded) || errors.Is(err, quota.ErrRateLimited) {
		writeJSON(w, http.StatusTooManyRequests, map[string]any{
			"error": err.Error(),
			"code":  "quota_exceeded",
		})
		return
	}
	if errors.Is(err, quota.ErrFeatureDisabled) || errors.Is(err, quota.ErrNoEntitlement) {
		writeJSON(w, http.StatusForbidden, map[string]any{
			"error": err.Error(),
			"code":  "feature_disabled",
		})
		return
	}
	writeError(w, http.StatusBadGateway, err.Error())
}

func (s *server) quotaTenant(r *http.Request) string {
	// Prefer header; WebSocket / EventSource clients may pass tenant_id query.
	hint := r.Header.Get("X-Tenant-Id")
	if hint == "" {
		hint = r.URL.Query().Get("tenant_id")
	}
	return auth.ResolveTenant(r.Context(), hint, s.cfg.AuthDisabled, s.cfg.DemoTenantID)
}

// voiceWS enforces rate limit, feature flags, concurrent slots, S16 daily/per-call caps, then relays.
func (s *server) voiceWS(w http.ResponseWriter, r *http.Request) {
	if s.quota == nil {
		s.voice.Handler().ServeHTTP(w, r)
		return
	}
	tenantID := s.quotaTenant(r)
	ctx := r.Context()

	if err := s.quota.AllowRate(ctx, tenantID, quota.BucketVoice); err != nil {
		writeQuotaError(w, err)
		return
	}
	if err := s.quota.CheckFeature(ctx, tenantID, quota.DimVoiceEnabled); err != nil {
		writeQuotaError(w, err)
		return
	}
	if err := s.quota.CheckMonthlyMinutes(ctx, tenantID, 0); err != nil {
		writeQuotaError(w, err)
		return
	}

	// S16 operational caps (under package monthly).
	maxPerCall := 0
	maxPerDay := 0
	tz := "Asia/Bangkok"
	if s.store != nil && tenantID != "" {
		tz = s.store.TenantTimezone(ctx, tenantID)
		if lim, err := s.store.GetOrCreateTenantCallLimits(ctx, tenantID); err == nil && lim != nil {
			maxPerCall = lim.MaxMinutesPerCall
			maxPerDay = lim.MaxCallMinutesPerDay
		}
		if err := s.quota.CheckDailyCallMinutes(ctx, tenantID, tz, maxPerDay); err != nil {
			writeQuotaError(w, err)
			return
		}
	}

	release, err := s.quota.AcquireConcurrent(ctx, tenantID)
	if err != nil {
		writeQuotaError(w, err)
		return
	}
	started := time.Now()
	defer func() {
		if release != nil {
			release()
		}
		// Best-effort: at least 1 minute if session lasted > 30s, else ceil minutes.
		elapsed := time.Since(started)
		mins := int(elapsed.Minutes())
		if mins < 1 && elapsed >= 30*time.Second {
			mins = 1
		}
		if mins > 0 {
			bg := context.Background()
			_ = s.quota.AddCallMinutes(bg, tenantID, mins)
			_ = s.quota.AddDailyCallMinutes(bg, tenantID, tz, mins)
		}
	}()

	// Per-call max: cancel context after N minutes so the relay ends.
	req := r
	if maxPerCall > 0 {
		deadline := time.Duration(maxPerCall) * time.Minute
		cctx, cancel := context.WithTimeout(ctx, deadline)
		defer cancel()
		req = r.WithContext(cctx)
	}
	s.voice.Handler().ServeHTTP(w, req)
}

// getPlatformTenantUsage serves GET /api/platform/tenants/{tenant_id}/usage.
func (s *server) getPlatformTenantUsage(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "tenant_id is required")
		return
	}
	if s.store != nil {
		if _, err := s.store.GetTenant(r.Context(), tenantID); err != nil {
			if errors.Is(err, store.ErrTenantNotFound) {
				writeError(w, http.StatusNotFound, "tenant not found")
				return
			}
			// Fall through — Snapshot may still work for known demo tenants.
		}
	}
	if s.quota == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"tenant_id": tenantID,
			"status":    "none",
			"period":    time.Now().UTC().Format("2006-01"),
			"package":   nil,
			"limits":    nil,
			"usage": map[string]int{
				"ai_employees": 0, "monthly_call_minutes": 0, "km_documents": 0, "concurrent_calls": 0,
			},
		})
		return
	}
	snap, err := s.quota.Snapshot(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, snap)
}
