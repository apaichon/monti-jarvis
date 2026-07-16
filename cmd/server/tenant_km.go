package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/km"
	"github.com/libra/monti-jarvis/internal/quota"
	"github.com/libra/monti-jarvis/internal/scope"
	"github.com/libra/monti-jarvis/internal/workforce"
)

func (s *server) tenantIDFromAuth(r *http.Request) (string, bool) {
	ac, ok := auth.FromContext(r.Context())
	if !ok || strings.TrimSpace(ac.TenantID) == "" {
		return "", false
	}
	return ac.TenantID, true
}

func (s *server) tenantKM(r *http.Request) *km.Service {
	tid, _ := s.tenantIDFromAuth(r)
	if s.km == nil {
		return nil
	}
	return s.km.WithTenant(tid)
}

// tenantKMAgentAvailable permits custom avatars only when they are actively
// assigned to this tenant. Tenants without assignments retain the built-in
// workforce fallback used by local/demo deployments.
func (s *server) tenantKMAgentAvailable(r *http.Request, tenantID, agentID string) bool {
	agentID = strings.ToLower(strings.TrimSpace(agentID))
	if agentID == "" || s.store == nil {
		return false
	}
	if !s.store.HasTenantAvatarAssignments(r.Context(), tenantID) {
		return tenantKMAgentIDAllowed(agentID, false, nil)
	}
	agents, err := s.customerWorkforceAgents(r, tenantID)
	if err != nil {
		return false
	}
	assignedIDs := make([]string, 0, len(agents))
	for _, agent := range agents {
		assignedIDs = append(assignedIDs, agent.ID)
	}
	return tenantKMAgentIDAllowed(agentID, true, assignedIDs)
}

func tenantKMAgentIDAllowed(agentID string, hasAssignments bool, assignedIDs []string) bool {
	agentID = strings.ToLower(strings.TrimSpace(agentID))
	if agentID == "" {
		return false
	}
	if !hasAssignments {
		return scope.ValidAgent(agentID)
	}
	for _, assignedID := range assignedIDs {
		if agentID == strings.ToLower(strings.TrimSpace(assignedID)) {
			return true
		}
	}
	return false
}

// GET /api/tenant/km/scopes
func (s *server) listTenantKMScopes(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.tenantIDFromAuth(r); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	type item struct {
		ID    string `json:"id"`
		Label string `json:"label"`
	}
	labels := map[string]string{
		"general": "General", "billing": "Billing", "technical": "Technical",
	}
	var scopes []item
	for _, id := range scope.AllScopes() {
		scopes = append(scopes, item{ID: id, Label: labels[id]})
	}
	writeJSON(w, http.StatusOK, map[string]any{"scopes": scopes})
}

// GET /api/tenant/km/agents
func (s *server) listTenantKMAgents(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	ctx := r.Context()
	type agentRow struct {
		ID            string         `json:"id"`
		Name          string         `json:"name"`
		Role          string         `json:"role"`
		DocCount      int            `json:"doc_count"`
		ChunkCount    int            `json:"chunk_count"`
		ByScope       map[string]int `json:"by_scope"`
		DefaultScopes []string       `json:"default_scopes"`
		Assigned      bool           `json:"assigned"`
	}

	var agents []workforce.Agent
	assigned := false
	if s.store != nil && s.store.HasTenantAvatarAssignments(ctx, tenantID) {
		list, err := s.store.ListWorkforceAgents(ctx, tenantID)
		if err == nil && len(list) > 0 {
			assigned = true
			for _, a := range list {
				agents = append(agents, workforce.FromWorkforceAgent(a))
			}
		}
	}
	if len(agents) == 0 {
		agents = workforce.All()
	}

	out := make([]agentRow, 0, len(agents))
	for _, a := range agents {
		row := agentRow{
			ID: a.ID, Name: a.Name, Role: a.Role,
			DefaultScopes: append([]string{}, scope.AgentScopes[a.ID]...),
			Assigned:      assigned,
			ByScope:       map[string]int{"general": 0, "billing": 0, "technical": 0},
		}
		if len(row.DefaultScopes) == 0 {
			row.DefaultScopes = []string{"general"}
		}
		if s.km != nil && s.store != nil {
			docs, chunks, err := s.store.CountAgentKnowledge(ctx, tenantID, a.ID)
			if err == nil {
				row.DocCount = docs
				row.ChunkCount = chunks
			}
			if by, err := s.store.CountAgentKnowledgeByScope(ctx, tenantID, a.ID); err == nil {
				row.ByScope = by
			}
		}
		out = append(out, row)
	}
	writeJSON(w, http.StatusOK, map[string]any{"agents": out})
}

// GET /api/tenant/km/agents/{agent_id}/documents
func (s *server) listTenantKMDocuments(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	agentID := strings.ToLower(strings.TrimSpace(r.PathValue("agent_id")))
	if !s.tenantKMAgentAvailable(r, tenantID, agentID) {
		writeError(w, http.StatusBadRequest, "unknown agent_id")
		return
	}
	svc := s.tenantKM(r)
	if svc == nil {
		writeError(w, http.StatusBadGateway, "knowledge service unavailable")
		return
	}
	docs, err := svc.ListAgentDocuments(r.Context(), agentID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	pub := make([]map[string]any, 0, len(docs))
	for _, d := range docs {
		pub = append(pub, km.PublicDocument(d))
	}
	writeJSON(w, http.StatusOK, map[string]any{"agent_id": agentID, "documents": pub})
}

// POST /api/tenant/km/agents/{agent_id}/documents
func (s *server) uploadTenantKMDocument(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	agentID := strings.ToLower(strings.TrimSpace(r.PathValue("agent_id")))
	if !s.tenantKMAgentAvailable(r, tenantID, agentID) {
		writeError(w, http.StatusBadRequest, "unknown agent_id")
		return
	}
	if err := r.ParseMultipartForm(8 << 20); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not read file")
		return
	}
	kmScope := strings.TrimSpace(r.FormValue("scope"))
	if kmScope == "" {
		kmScope = scope.DefaultScope(agentID)
	}
	if !scope.ValidScope(kmScope) {
		writeError(w, http.StatusBadRequest, "invalid scope")
		return
	}

	ctx := r.Context()
	if s.quota != nil {
		if err := s.quota.AllowRate(ctx, tenantID, quota.BucketKM); err != nil {
			writeQuotaError(w, err)
			return
		}
		if err := s.quota.CheckKMDocument(ctx, tenantID); err != nil {
			writeQuotaError(w, err)
			return
		}
	}
	svc := s.tenantKM(r)
	if svc == nil {
		writeError(w, http.StatusBadGateway, "knowledge service unavailable")
		return
	}
	doc, err := svc.Ingest(ctx, agentID, header.Filename, data, kmScope)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, km.PublicDocument(doc))
}

// PATCH /api/tenant/km/documents/{id}
func (s *server) patchTenantKMDocument(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.tenantIDFromAuth(r); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "id required")
		return
	}
	var body struct {
		KMScope string `json:"km_scope"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	svc := s.tenantKM(r)
	if svc == nil {
		writeError(w, http.StatusBadGateway, "knowledge service unavailable")
		return
	}
	doc, err := svc.UpdateDocumentScope(r.Context(), id, body.KMScope)
	if err != nil {
		if errors.Is(err, km.ErrNotFound) {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		if strings.Contains(err.Error(), "invalid scope") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, km.PublicDocument(doc))
}

// DELETE /api/tenant/km/documents/{id}
func (s *server) deleteTenantKMDocument(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.tenantIDFromAuth(r); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		writeError(w, http.StatusBadRequest, "id required")
		return
	}
	svc := s.tenantKM(r)
	if svc == nil {
		writeError(w, http.StatusBadGateway, "knowledge service unavailable")
		return
	}
	if err := svc.DeleteDocument(r.Context(), id); err != nil {
		if errors.Is(err, km.ErrNotFound) {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id})
}

// POST /api/tenant/km/agents/{agent_id}/reset
func (s *server) resetTenantKMAgent(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	agentID := strings.ToLower(strings.TrimSpace(r.PathValue("agent_id")))
	if !s.tenantKMAgentAvailable(r, tenantID, agentID) {
		writeError(w, http.StatusBadRequest, "unknown agent_id")
		return
	}
	svc := s.tenantKM(r)
	if svc == nil {
		writeError(w, http.StatusBadGateway, "knowledge service unavailable")
		return
	}
	if err := svc.ResetAgent(r.Context(), agentID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"agent_id": agentID,
		"status":   "reset",
		"message":  "knowledge base cleared for agent",
		"at":       time.Now().UTC(),
	})
}
