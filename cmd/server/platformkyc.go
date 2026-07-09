package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/libra/monti-jarvis/internal/store"
)

type kycRejectRequest struct {
	Reason string `json:"reason"`
}

func (s *server) getPlatformTenantKYC(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "tenant_id is required")
		return
	}

	tenant, err := s.store.GetTenant(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	reg, regErr := s.store.GetTenantRegistration(r.Context(), tenantID)
	if regErr != nil && !errors.Is(regErr, store.ErrTenantNotFound) {
		writeError(w, http.StatusBadGateway, regErr.Error())
		return
	}

	profile, err := s.store.GetTenantKYCProfile(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	var regJSON any
	if regErr == nil {
		regJSON = map[string]any{
			"id":           reg.ID,
			"company_name": reg.CompanyName,
			"admin_email":  reg.AdminEmail,
			"status":       reg.Status,
			"created_at":   reg.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tenant": map[string]any{
			"id":         tenant.ID,
			"slug":       tenant.Slug,
			"name":       tenant.Name,
			"status":     tenant.Status,
			"created_at": tenant.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		},
		"registration": regJSON,
		"kyc":          kycProfileJSON(profile),
	})
}

func (s *server) approvePlatformTenantKYC(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "tenant_id is required")
		return
	}

	result, err := s.store.ApproveTenantKYC(r.Context(), tenantID)
	if err != nil {
		writePlatformKYCError(w, err)
		return
	}
	s.sendKYCApprovedEmail(r.Context(), result)
	writeJSON(w, http.StatusOK, platformKYCDecisionJSON(result))
}

func (s *server) rejectPlatformTenantKYC(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "tenant_id is required")
		return
	}

	var req kycRejectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(req.Reason) == "" {
		writeError(w, http.StatusBadRequest, "reason is required")
		return
	}

	result, err := s.store.RejectTenantKYC(r.Context(), tenantID, req.Reason)
	if err != nil {
		writePlatformKYCError(w, err)
		return
	}
	s.sendKYCRejectedEmail(r.Context(), result)
	writeJSON(w, http.StatusOK, platformKYCDecisionJSON(result))
}

func writePlatformKYCError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrTenantNotFound):
		writeError(w, http.StatusNotFound, "tenant not found")
	case errors.Is(err, store.ErrKYCReviewConflict):
		writeError(w, http.StatusConflict, "kyc package is not ready for review")
	default:
		if strings.Contains(err.Error(), "rejection reason is required") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
	}
}

func platformKYCDecisionJSON(result store.PlatformKYCDecisionResult) map[string]any {
	out := map[string]any{
		"tenant_id":           result.TenantID,
		"tenant_status":       result.TenantStatus,
		"registration_status": result.RegistrationStatus,
		"kyc_status":          result.KYCStatus,
		"reviewed_at":         result.ReviewedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		"reviewed_by":         result.ReviewedBy,
	}
	if result.RejectionReason != "" {
		out["rejection_reason"] = result.RejectionReason
	}
	return out
}