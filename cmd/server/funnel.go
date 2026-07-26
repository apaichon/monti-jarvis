package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/leads"
	"github.com/libra/monti-jarvis/internal/store"
)

type funnelEventRequest struct {
	EventName    string `json:"event_name"`
	PagePath     string `json:"page_path"`
	CTAID        string `json:"cta_id"`
	SessionKey   string `json:"session_key"`
	UTMSource    string `json:"utm_source"`
	UTMMedium    string `json:"utm_medium"`
	UTMCampaign  string `json:"utm_campaign"`
	UTMContent   string `json:"utm_content"`
	UTMTerm      string `json:"utm_term"`
	ReferralCode string `json:"referral_code"`
}

func (s *server) createFunnelEvent(w http.ResponseWriter, r *http.Request) {
	if s.store == nil || s.store.Health(r.Context()).Postgres != "ok" {
		writeLeadError(w, http.StatusServiceUnavailable, "funnel events unavailable", "FUNNEL_UNAVAILABLE")
		return
	}

	var req funnelEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeLeadError(w, http.StatusBadRequest, "invalid JSON", "LEAD_VALIDATION")
		return
	}

	ctx := auth.WithRequestMeta(r.Context(), r)
	meta := auth.RequestMetaFrom(ctx)
	if s.funnelLimiter != nil {
		allowed, err := s.funnelLimiter.Allow(ctx, meta.IP)
		if err != nil {
			log.Printf("funnel rate limit warning: %v", err)
		} else if !allowed {
			writeLeadError(w, http.StatusTooManyRequests, "too many funnel events", "FUNNEL_RATE_LIMITED")
			return
		}
	}

	in := leads.NormalizeFunnel(leads.FunnelInput{
		EventName: req.EventName, PagePath: req.PagePath, CTAID: req.CTAID,
		SessionKey: req.SessionKey, UTMSource: req.UTMSource, UTMMedium: req.UTMMedium,
		UTMCampaign: req.UTMCampaign, UTMContent: req.UTMContent, UTMTerm: req.UTMTerm,
		ReferralCode: req.ReferralCode,
	})
	if err := leads.ValidateFunnel(in); err != nil {
		if errors.Is(err, leads.ErrUnknownEvent) {
			writeLeadError(w, http.StatusBadRequest, "unknown funnel event", "FUNNEL_UNKNOWN_EVENT")
			return
		}
		writeLeadError(w, http.StatusBadRequest, "invalid funnel event", "LEAD_VALIDATION")
		return
	}

	id, err := s.store.InsertFunnelEvent(ctx, store.FunnelEventInput{
		EventName: in.EventName, PagePath: in.PagePath, CTAID: in.CTAID,
		SessionKey: in.SessionKey, UTMSource: in.UTMSource, UTMMedium: in.UTMMedium,
		UTMCampaign: in.UTMCampaign, UTMContent: in.UTMContent, UTMTerm: in.UTMTerm,
		ReferralCode: in.ReferralCode, ClientIPHash: hashClientIP(meta.IP),
	})
	if err != nil {
		writeLeadError(w, http.StatusBadGateway, err.Error(), "FUNNEL_UNAVAILABLE")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"id": id, "ok": true})
}

func hashClientIP(ip string) string {
	ip = strings.TrimSpace(ip)
	if ip == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(ip))
	return hex.EncodeToString(sum[:])
}
