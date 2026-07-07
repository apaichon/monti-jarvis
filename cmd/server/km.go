package main

import (
	"context"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/scope"
)

func (s *server) tenantID(r *http.Request) string {
	if v := strings.TrimSpace(r.Header.Get("X-Tenant-Id")); v != "" {
		return v
	}
	return s.cfg.DemoTenantID
}

func (s *server) getAgentKnowledge(w http.ResponseWriter, r *http.Request) {
	agentID := strings.ToLower(strings.TrimSpace(r.PathValue("agent_id")))
	if !scope.ValidAgent(agentID) {
		writeError(w, http.StatusBadRequest, "unknown agent_id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	info, err := s.km.AgentKnowledge(ctx, agentID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, info)
}

func (s *server) listAgentDocuments(w http.ResponseWriter, r *http.Request) {
	agentID := strings.ToLower(strings.TrimSpace(r.PathValue("agent_id")))
	if !scope.ValidAgent(agentID) {
		writeError(w, http.StatusBadRequest, "unknown agent_id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()
	docs, err := s.km.ListAgentDocuments(ctx, agentID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"agent_id": agentID, "documents": docs})
}

func (s *server) uploadAgentDocument(w http.ResponseWriter, r *http.Request) {
	agentID := strings.ToLower(strings.TrimSpace(r.PathValue("agent_id")))
	if !scope.ValidAgent(agentID) {
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

	ctx, cancel := context.WithTimeout(r.Context(), 120*time.Second)
	defer cancel()
	doc, err := s.km.Ingest(ctx, agentID, header.Filename, data, kmScope)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, doc)
}

func (s *server) resetAgentKnowledge(w http.ResponseWriter, r *http.Request) {
	agentID := strings.ToLower(strings.TrimSpace(r.PathValue("agent_id")))
	if !scope.ValidAgent(agentID) {
		writeError(w, http.StatusBadRequest, "unknown agent_id")
		return
	}
	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()
	if err := s.km.ResetAgent(ctx, agentID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"agent_id": agentID,
		"status":   "reset",
		"message":  "knowledge base cleared for agent",
	})
}

func (s *server) seedKnowledge(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 180*time.Second)
	defer cancel()

	agents := []string{"ava", "max", "luna", "neo"}
	var seeded []any
	for _, agentID := range agents {
		path := filepath.Join("docs", "samples", "km", agentID+".md")
		data, err := os.ReadFile(path)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "missing sample file: "+path)
			return
		}
		doc, err := s.km.Ingest(ctx, agentID, agentID+".md", data, scope.DefaultScope(agentID))
		if err != nil {
			writeError(w, http.StatusBadGateway, agentID+": "+err.Error())
			return
		}
		seeded = append(seeded, map[string]any{"agent_id": agentID, "document_id": doc.ID, "chunks": doc.ChunkCount})
	}
	writeJSON(w, http.StatusOK, map[string]any{"seeded": seeded})
}