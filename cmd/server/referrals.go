package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/libra/monti-jarvis/internal/store"
)

func (s *server) getTenantReferralCode(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	code, err := s.store.GetOrCreateTenantReferralCode(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, code)
}

func (s *server) recordReferralClick(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Code        string `json:"code"`
		Source      string `json:"source"`
		LandingPath string `json:"landing_path"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	click, err := s.store.RecordReferralClick(r.Context(), body.Code, body.Source, body.LandingPath)
	if err != nil {
		if errors.Is(err, store.ErrReferralInvalid) {
			writeError(w, http.StatusBadRequest, "referral code is invalid")
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, click)
}

func (s *server) listTenantReferrals(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := s.store.ListTenantReferrals(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	balances, err := s.store.ListTenantBonusBalances(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tenant_id": tenantID, "referrals": items, "bonus": balances})
}

func (s *server) qualifyPlatformReferral(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "referral id is required")
		return
	}
	item, err := s.store.QualifyReferral(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrReferralNotQualified) {
			writeJSON(w, http.StatusConflict, map[string]any{
				"error":    "referral_not_qualified",
				"referral": item,
			})
			return
		}
		if errors.Is(err, store.ErrReferralNotFound) {
			writeError(w, http.StatusNotFound, "referral not found")
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *server) listReferralRewardRules(w http.ResponseWriter, r *http.Request) {
	rules, err := s.store.ListReferralRewardRules(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rules": rules})
}

func (s *server) putReferralRewardRule(w http.ResponseWriter, r *http.Request) {
	dimension := strings.TrimSpace(r.PathValue("dimension"))
	var body struct {
		GrantAmount int64 `json:"grant_amount"`
		ExpiryDays  int   `json:"expiry_days"`
		Active      bool  `json:"active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	rule, err := s.store.UpsertReferralRewardRule(r.Context(), store.ReferralRewardRule{
		Dimension: dimension, GrantAmount: body.GrantAmount, ExpiryDays: body.ExpiryDays, Active: body.Active,
	})
	if err != nil {
		if errors.Is(err, store.ErrBonusInvalid) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (s *server) reversePlatformReferral(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Reason string `json:"reason"`
	}
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	item, err := s.store.ReverseReferral(r.Context(), strings.TrimSpace(r.PathValue("id")), body.Reason)
	if err != nil {
		if errors.Is(err, store.ErrReferralNotFound) {
			writeError(w, http.StatusNotFound, "referral not found")
			return
		}
		if errors.Is(err, store.ErrReferralNotQualified) {
			writeJSON(w, http.StatusConflict, map[string]any{"error": "referral_not_qualified", "referral": item})
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, item)
}
