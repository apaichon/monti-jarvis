package main

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/live"
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
		case "quota_unavailable":
			status = http.StatusServiceUnavailable
		case "feature_disabled":
			status = http.StatusForbidden
		case "no_entitlement":
			status = http.StatusForbidden
		case "daily_call_limit", "per_call_limit", "preview_concurrent", "queue_full", "queue_timeout":
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

// voiceWS enforces package quotas (rate, concurrent, monthly + S16 daily/per-call), then relays.
func (s *server) voiceWS(w http.ResponseWriter, r *http.Request) {
	tenantID, _, err := s.requestTenantContext(r)
	if err != nil {
		writeEmbedContextError(w, err)
		return
	}
	s.voiceWithPackageQuota(w, r, tenantID)
}

// voiceWithPackageQuota applies the same package metering as production voice
// (used by /ws/voice and tenant preview voice — preview also logs source=preview).
func (s *server) voiceWithPackageQuota(w http.ResponseWriter, r *http.Request, tenantID string) {
	r = withVoiceTenant(r, tenantID)
	if s.quota == nil || s.voice == nil {
		if s.voice != nil {
			s.voice.Handler().ServeHTTP(w, r)
		} else {
			writeError(w, http.StatusServiceUnavailable, "voice not configured")
		}
		return
	}
	ctx := r.Context()

	if err := s.quota.AllowRate(ctx, tenantID, quota.BucketVoice); err != nil {
		writeQuotaError(w, err)
		return
	}
	if err := s.quota.CheckFeature(ctx, tenantID, quota.DimVoiceEnabled); err != nil {
		writeQuotaError(w, err)
		return
	}
	mobileCall := r.URL.Query().Get("mobile") == "1"
	if mobileCall {
		if err := s.quota.CheckMobileMinutes(ctx, tenantID, 0); err != nil {
			writeQuotaError(w, err)
			return
		}
	} else if err := s.quota.CheckMonthlyMinutes(ctx, tenantID, 0); err != nil {
		writeQuotaError(w, err)
		return
	}

	// S16 operational caps (under package monthly); S18 tier overrides when tier_id set.
	maxPerCall := 0
	maxPerDay := 0
	tz := "Asia/Bangkok"
	if s.store != nil && tenantID != "" {
		tz = s.store.TenantTimezone(ctx, tenantID)
		if lim, err := s.store.GetOrCreateTenantCallLimits(ctx, tenantID); err == nil && lim != nil {
			maxPerCall = lim.MaxMinutesPerCall
			maxPerDay = lim.MaxCallMinutesPerDay
		}
		if tierID := strings.TrimSpace(r.URL.Query().Get("tier_id")); tierID != "" {
			if t, err := s.store.GetCustomerTier(ctx, tenantID, tierID); err == nil && t != nil && t.Active {
				if t.MaxMinutesPerCall > 0 {
					maxPerCall = t.MaxMinutesPerCall
				}
				if t.MaxCallMinutesPerDay > 0 {
					maxPerDay = t.MaxCallMinutesPerDay
				}
			}
		}
		if err := s.quota.CheckDailyCallMinutes(ctx, tenantID, tz, maxPerDay); err != nil {
			writeQuotaError(w, err)
			return
		}
	}

	req := r
	relay := s.voice
	if relay != nil {
		cp := *relay
		cp.Admission = func(admissionCtx context.Context, admissionTenantID string, emit func(live.AdmissionUpdate) error) (live.AdmissionResult, error) {
			admission, err := s.quota.WaitForQueuedConcurrent(admissionCtx, admissionTenantID, r.URL.Query().Get("admission_id"), func(update quota.QueueUpdate) error {
				return emit(liveAdmissionUpdate(update))
			})
			if err != nil {
				return live.AdmissionResult{}, liveQuotaError(err)
			}
			if admission == nil {
				return live.AdmissionResult{}, nil
			}
			started := time.Now()
			result := live.AdmissionResult{
				Release: func() {
					if admission.Release != nil {
						admission.Release()
					}
					s.recordCallQuotaMinutes(admissionTenantID, r.URL.Query().Get("call_id"), mobileCall, tz, started)
				},
				Capacity: liveAdmissionCapacity(admission.AdmissionID, admission.Snapshot),
			}
			if maxPerCall > 0 {
				deadline := time.Duration(maxPerCall) * time.Minute
				cctx, cancel := context.WithTimeout(admissionCtx, deadline)
				result.Context = cctx
				result.Cancel = cancel
			}
			return result, nil
		}
		relay = &cp
	}
	relay.Handler().ServeHTTP(w, req)
}

func (s *server) recordCallQuotaMinutes(tenantID, callID string, mobile bool, timezone string, started time.Time) {
	if s == nil || s.quota == nil || started.IsZero() {
		return
	}
	elapsed := time.Since(started)
	mins := int(elapsed.Minutes())
	if mins < 1 && elapsed >= 30*time.Second {
		mins = 1
	}
	if mins <= 0 {
		return
	}
	bg := context.Background()
	if mobile {
		_ = s.quota.AddMobileCallMinutes(bg, tenantID, mins)
	} else {
		_ = s.quota.AddCallMinutes(bg, tenantID, mins)
	}
	_ = s.quota.AddDailyCallMinutes(bg, tenantID, timezone, mins)
	s.recordCallUsageEvent(tenantID, callID, mobile, mins)
}

func liveAdmissionUpdate(update quota.QueueUpdate) live.AdmissionUpdate {
	out := live.AdmissionUpdate{
		Type:    update.Type,
		Message: update.Message,
		AdmissionCapacity: live.AdmissionCapacity{
			AdmissionID:          update.AdmissionID,
			Position:             update.Position,
			EstimatedWaitSeconds: update.EstimatedWaitSeconds,
			QueueEnabled:         update.Snapshot.QueueEnabled,
			ActiveCalls:          update.Snapshot.ActiveCalls,
			QueuedCallers:        update.Snapshot.QueuedCallers,
			TotalCalls:           update.Snapshot.TotalCalls,
			MaxConcurrentCalls:   update.Snapshot.MaxConcurrentCalls,
			BusyStatus:           update.Snapshot.BusyStatus,
		},
	}
	return out
}

func liveAdmissionCapacity(admissionID string, snapshot quota.ConcurrentQueueSnapshot) *live.AdmissionCapacity {
	return &live.AdmissionCapacity{
		AdmissionID:        admissionID,
		QueueEnabled:       snapshot.QueueEnabled,
		ActiveCalls:        snapshot.ActiveCalls,
		QueuedCallers:      snapshot.QueuedCallers,
		TotalCalls:         snapshot.TotalCalls,
		MaxConcurrentCalls: snapshot.MaxConcurrentCalls,
		BusyStatus:         snapshot.BusyStatus,
	}
}

func liveQuotaError(err error) error {
	if err == nil {
		return nil
	}
	var qe *quota.Error
	if errors.As(err, &qe) {
		return &live.AdmissionError{Code: qe.Code, Message: qe.Error(), Err: err}
	}
	return err
}

func (s *server) recordCallUsageEvent(tenantID, callID string, mobile bool, minutes int) {
	if s == nil || s.store == nil || strings.TrimSpace(callID) == "" || minutes <= 0 {
		return
	}
	dimension, sourceType := "monthly_call_minutes", "call"
	if mobile {
		dimension, sourceType = "mobile_call_minutes", "mobile_call"
	}
	now := time.Now().UTC()
	snapshotID := ""
	if entitlement, err := s.store.GetActiveEntitlement(context.Background(), tenantID); err == nil && entitlement != nil {
		snapshotID = entitlement.ID
	}
	_, _, _ = s.store.RecordUsageEvent(context.Background(), store.UsageEventInput{
		TenantID: tenantID, IdempotencyKey: "call:" + callID + ":" + dimension,
		Dimension: dimension, Unit: "minutes", Amount: int64(minutes),
		PeriodStart: now, PeriodEnd: now, SourceType: sourceType, SourceID: callID,
		EntitlementSnapshotID: snapshotID,
	})
}

func withVoiceTenant(r *http.Request, tenantID string) *http.Request {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return r
	}
	req := r.Clone(r.Context())
	u := *r.URL
	query := u.Query()
	query.Set("tenant_id", tenantID)
	u.RawQuery = query.Encode()
	req.URL = &u
	return req
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
				"ai_employees": 0, "monthly_call_minutes": 0, "mobile_call_minutes": 0, "km_documents": 0, "storage_bytes": 0, "concurrent_calls": 0,
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

// getTenantConcurrentCallQueueStatus serves GET /api/tenant/concurrent-call-queue/status.
func (s *server) getTenantConcurrentCallQueueStatus(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.commercialTenantID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if s.quota == nil {
		writeJSON(w, http.StatusOK, quota.ConcurrentQueueSnapshot{
			QueueEnabled: false,
			BusyStatus:   "available",
		})
		return
	}
	snapshot, err := s.quota.ConcurrentQueueSnapshot(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, snapshot)
}
