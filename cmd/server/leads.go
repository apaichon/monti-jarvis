package main

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/libra/monti-jarvis/internal/auditctx"
	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/leads"
	"github.com/libra/monti-jarvis/internal/store"
)

type publicLeadRequest struct {
	Kind              string `json:"kind"`
	Email             string `json:"email"`
	FullName          string `json:"full_name"`
	CompanyName       string `json:"company_name"`
	Phone             string `json:"phone"`
	UseCase           string `json:"use_case"`
	PreferredChannel  string `json:"preferred_channel"`
	Language          string `json:"language"`
	ConsentMarketing  bool   `json:"consent_marketing"`
	ConsentContact    bool   `json:"consent_contact"`
	UTMSource         string `json:"utm_source"`
	UTMMedium         string `json:"utm_medium"`
	UTMCampaign       string `json:"utm_campaign"`
	UTMContent        string `json:"utm_content"`
	UTMTerm           string `json:"utm_term"`
	ReferralCode      string `json:"referral_code"`
	LandingPath       string `json:"landing_path"`
	PackageInterestID string `json:"package_interest_id"`
	Website           string `json:"website"`
}

func (s *server) createPublicLead(w http.ResponseWriter, r *http.Request) {
	if !s.cfg.LeadCaptureEnabled {
		writeLeadError(w, http.StatusServiceUnavailable, "lead capture is disabled", "LEAD_DISABLED")
		return
	}
	if s.store == nil || s.store.Health(r.Context()).Postgres != "ok" {
		writeLeadError(w, http.StatusServiceUnavailable, "lead capture unavailable", "LEAD_DISABLED")
		return
	}

	var req publicLeadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeLeadError(w, http.StatusBadRequest, "invalid JSON", "LEAD_VALIDATION")
		return
	}

	ctx := auth.WithRequestMeta(r.Context(), r)
	if s.leadLimiter != nil {
		meta := auth.RequestMetaFrom(ctx)
		allowed, err := s.leadLimiter.Allow(ctx, meta.IP)
		if err != nil {
			log.Printf("lead rate limit warning: %v", err)
		} else if !allowed {
			writeLeadError(w, http.StatusTooManyRequests, "too many lead submissions", "LEAD_RATE_LIMITED")
			return
		}
	}

	in := leads.LeadInput{
		Kind: req.Kind, Email: req.Email, FullName: req.FullName, CompanyName: req.CompanyName,
		Phone: req.Phone, UseCase: req.UseCase, PreferredChannel: req.PreferredChannel,
		Language: req.Language, ConsentMarketing: req.ConsentMarketing, ConsentContact: req.ConsentContact,
		UTMSource: req.UTMSource, UTMMedium: req.UTMMedium, UTMCampaign: req.UTMCampaign,
		UTMContent: req.UTMContent, UTMTerm: req.UTMTerm, ReferralCode: req.ReferralCode,
		LandingPath: req.LandingPath, PackageInterestID: req.PackageInterestID, Website: req.Website,
	}
	result, err := s.store.CreateLead(ctx, in, s.cfg.LeadDedupeWindowHours)
	if err != nil {
		writeLeadDomainError(w, err)
		return
	}
	status := http.StatusCreated
	if result.Deduped {
		status = http.StatusOK
	}
	writeJSON(w, status, map[string]any{
		"lead_id": result.Lead.ID,
		"status":  result.Lead.Status,
		"deduped": result.Deduped,
	})
}

func (s *server) listPlatformLeads(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeLeadError(w, http.StatusServiceUnavailable, "leads unavailable", "LEAD_DISABLED")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 50
	}
	if limit > 200 {
		limit = 200
	}
	if offset < 0 {
		offset = 0
	}
	items, total, err := s.store.ListLeads(r.Context(), store.LeadListFilters{
		Status: strings.TrimSpace(r.URL.Query().Get("status")),
		Kind:   strings.TrimSpace(r.URL.Query().Get("kind")),
		Source: strings.TrimSpace(r.URL.Query().Get("source")),
		Q:      strings.TrimSpace(r.URL.Query().Get("q")),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeLeadError(w, http.StatusBadGateway, err.Error(), "LEAD_UNAVAILABLE")
		return
	}
	if items == nil {
		items = []store.MarketingLead{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"items": items, "total": total, "limit": limit, "offset": offset,
	})
}

func (s *server) getPlatformLead(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeLeadError(w, http.StatusServiceUnavailable, "leads unavailable", "LEAD_DISABLED")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	lead, notes, events, err := s.store.GetLeadDetail(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrLeadNotFound) {
			writeLeadError(w, http.StatusNotFound, "lead not found", "LEAD_NOT_FOUND")
			return
		}
		writeLeadError(w, http.StatusBadGateway, err.Error(), "LEAD_UNAVAILABLE")
		return
	}
	if notes == nil {
		notes = []store.MarketingLeadNote{}
	}
	if events == nil {
		events = []store.MarketingLeadEvent{}
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"lead": lead, "notes": notes, "events": events,
	})
}

func (s *server) patchPlatformLead(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeLeadError(w, http.StatusServiceUnavailable, "leads unavailable", "LEAD_DISABLED")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	var body struct {
		Status     *string `json:"status"`
		AssignedTo *string `json:"assigned_to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeLeadError(w, http.StatusBadRequest, "invalid JSON", "LEAD_VALIDATION")
		return
	}
	if body.Status == nil && body.AssignedTo == nil {
		writeLeadError(w, http.StatusBadRequest, "status or assigned_to is required", "LEAD_VALIDATION")
		return
	}

	ctx := r.Context()
	if ac, ok := auth.FromContext(ctx); ok && ac.UserID != "" {
		ctx = auditctx.WithActor(ctx, ac.UserID)
	}
	lead, err := s.store.PatchLead(ctx, id, body.Status, body.AssignedTo)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrLeadNotFound):
			writeLeadError(w, http.StatusNotFound, "lead not found", "LEAD_NOT_FOUND")
		case errors.Is(err, store.ErrLeadInvalid):
			writeLeadError(w, http.StatusBadRequest, "invalid lead status", "LEAD_VALIDATION")
		default:
			writeLeadError(w, http.StatusBadGateway, err.Error(), "LEAD_UNAVAILABLE")
		}
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"lead": lead})
}

func (s *server) addPlatformLeadNote(w http.ResponseWriter, r *http.Request) {
	if s.store == nil {
		writeLeadError(w, http.StatusServiceUnavailable, "leads unavailable", "LEAD_DISABLED")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	var body struct {
		Body string `json:"body"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeLeadError(w, http.StatusBadRequest, "invalid JSON", "LEAD_VALIDATION")
		return
	}
	ctx := r.Context()
	if ac, ok := auth.FromContext(ctx); ok && ac.UserID != "" {
		ctx = auditctx.WithActor(ctx, ac.UserID)
	}
	note, err := s.store.AddLeadNote(ctx, id, body.Body)
	if err != nil {
		switch {
		case errors.Is(err, store.ErrLeadNotFound):
			writeLeadError(w, http.StatusNotFound, "lead not found", "LEAD_NOT_FOUND")
		case errors.Is(err, leads.ErrNoteEmpty), errors.Is(err, leads.ErrNoteTooLong):
			writeLeadError(w, http.StatusBadRequest, err.Error(), "LEAD_VALIDATION")
		default:
			writeLeadError(w, http.StatusBadGateway, err.Error(), "LEAD_UNAVAILABLE")
		}
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"note": note})
}

func writeLeadDomainError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, leads.ErrSpam):
		writeLeadError(w, http.StatusBadRequest, "spam detected", "LEAD_SPAM")
	case errors.Is(err, leads.ErrConsentRequired):
		writeLeadError(w, http.StatusBadRequest, "consent is required", "LEAD_CONSENT_REQUIRED")
	case errors.Is(err, leads.ErrValidation):
		writeLeadError(w, http.StatusBadRequest, "invalid lead fields", "LEAD_VALIDATION")
	default:
		writeLeadError(w, http.StatusBadGateway, err.Error(), "LEAD_UNAVAILABLE")
	}
}

func writeLeadError(w http.ResponseWriter, status int, message, code string) {
	writeJSON(w, status, map[string]any{"error": message, "code": code})
}
