package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/libra/monti-jarvis/internal/store"
)

type putSettingsBody struct {
	Locale         *string `json:"locale"`
	Timezone       *string `json:"timezone"`
	DisplayName    *string `json:"display_name"`
	AIReplyLocale  *string `json:"ai_reply_locale"`
	UserTierLabel  *string `json:"user_tier_label"`
	UserGroupLabel *string `json:"user_group_label"`
}

type putCallLimitsBody struct {
	MaxMinutesPerCall    *int `json:"max_minutes_per_call"`
	MaxCallMinutesPerDay *int `json:"max_call_minutes_per_day"`
}

func writeSettingsError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrInvalidLocale):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "locale must be en or th", "code": "invalid_locale"})
	case errors.Is(err, store.ErrInvalidTimezone):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid timezone (use IANA name)", "code": "invalid_timezone"})
	case errors.Is(err, store.ErrSettingsNotFound), errors.Is(err, store.ErrCallLimitsNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		if err != nil && strings.Contains(err.Error(), "must be >= 0") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
	}
}

// GET /api/tenant/settings
func (s *server) getTenantSettings(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	row, err := s.store.GetOrCreateTenantSettings(r.Context(), tenantID)
	if err != nil {
		writeSettingsError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

// PUT /api/tenant/settings
func (s *server) putTenantSettings(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body putSettingsBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	row, err := s.store.UpdateTenantSettings(r.Context(), tenantID, store.UpdateTenantSettingsInput{
		Locale:         body.Locale,
		Timezone:       body.Timezone,
		DisplayName:    body.DisplayName,
		AIReplyLocale:  body.AIReplyLocale,
		UserTierLabel:  body.UserTierLabel,
		UserGroupLabel: body.UserGroupLabel,
	})
	if err != nil {
		writeSettingsError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

// GET /api/tenant/usage — package limits + usage for JWT tenant (same shape as platform S13).
func (s *server) getTenantUsage(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if s.quota == nil {
		writeJSON(w, http.StatusOK, map[string]any{
			"tenant_id": tenantID,
			"status":    "none",
			"period":    "",
			"package":   nil,
			"limits":    nil,
			"usage": map[string]int{
				"ai_employees": 0, "monthly_call_minutes": 0, "km_documents": 0, "concurrent_calls": 0,
			},
			"call_limits": nil,
			"daily_usage": map[string]int{"call_minutes": 0},
		})
		return
	}
	snap, err := s.quota.Snapshot(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	// Enrich with S16 operational caps + daily usage.
	out := map[string]any{
		"tenant_id": snap.TenantID,
		"package":   snap.Package,
		"status":    snap.Status,
		"period":    snap.Period,
		"limits":    snap.Limits,
		"usage":     snap.Usage,
	}
	tz := "Asia/Bangkok"
	if s.store != nil {
		tz = s.store.TenantTimezone(r.Context(), tenantID)
		if lim, err := s.store.GetOrCreateTenantCallLimits(r.Context(), tenantID); err == nil {
			out["call_limits"] = lim
		}
	}
	daily, _ := s.quota.GetDailyCallMinutes(r.Context(), tenantID, tz)
	out["daily_usage"] = map[string]any{
		"call_minutes": daily,
		"timezone":     tz,
	}
	writeJSON(w, http.StatusOK, out)
}

// GET /api/tenant/call-limits
func (s *server) getTenantCallLimits(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	row, err := s.store.GetOrCreateTenantCallLimits(r.Context(), tenantID)
	if err != nil {
		writeSettingsError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

// PUT /api/tenant/call-limits
func (s *server) putTenantCallLimits(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body putCallLimitsBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	cur, err := s.store.GetOrCreateTenantCallLimits(r.Context(), tenantID)
	if err != nil {
		writeSettingsError(w, err)
		return
	}
	maxCall := cur.MaxMinutesPerCall
	maxDay := cur.MaxCallMinutesPerDay
	if body.MaxMinutesPerCall != nil {
		maxCall = *body.MaxMinutesPerCall
	}
	if body.MaxCallMinutesPerDay != nil {
		maxDay = *body.MaxCallMinutesPerDay
	}
	if maxCall < 0 || maxDay < 0 {
		writeError(w, http.StatusBadRequest, "call limits must be >= 0")
		return
	}
	// Clamp under package monthly ceiling when known (0 = unset).
	if s.quota != nil {
		if snap, err := s.quota.Snapshot(r.Context(), tenantID); err == nil && snap != nil && snap.Limits != nil {
			pkg := snap.Limits.MaxMonthlyCallMinutes
			if pkg > 0 {
				if maxDay > pkg {
					maxDay = pkg
				}
				if maxCall > pkg {
					maxCall = pkg
				}
			}
		}
	}
	row, err := s.store.UpdateTenantCallLimits(r.Context(), tenantID, maxCall, maxDay)
	if err != nil {
		writeSettingsError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}
