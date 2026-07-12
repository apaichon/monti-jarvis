package main

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/store"
)

type patchKMGapBody struct {
	Status            string `json:"status"`
	Notes             string `json:"notes"`
	ResolvedDocumentID string `json:"resolved_document_id"`
}

// GET /api/tenant/km/gaps?status=&agent_id=&limit=
func (s *server) listTenantKMGaps(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok || strings.TrimSpace(ac.TenantID) == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	status := r.URL.Query().Get("status")
	agentID := r.URL.Query().Get("agent_id")
	limit := 100
	gaps, err := s.store.ListKMGaps(r.Context(), ac.TenantID, status, agentID, limit)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if gaps == nil {
		gaps = []store.KMGap{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"gaps": gaps})
}

// PATCH /api/tenant/km/gaps/{id}
func (s *server) patchTenantKMGap(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok || strings.TrimSpace(ac.TenantID) == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "id required")
		return
	}
	var body patchKMGapBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if strings.TrimSpace(body.Status) == "" {
		writeError(w, http.StatusBadRequest, "status required")
		return
	}
	g, err := s.store.UpdateKMGapStatus(r.Context(), ac.TenantID, id, body.Status, body.Notes, body.ResolvedDocumentID)
	if err != nil {
		if strings.Contains(err.Error(), "invalid status") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		// pgx no rows
		if strings.Contains(err.Error(), "no rows") {
			writeError(w, http.StatusNotFound, "gap not found")
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, g)
}
