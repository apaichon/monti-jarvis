package main

import (
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
	writeJSON(w, http.StatusOK, map[string]any{"tenant_id": tenantID, "referrals": items})
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
