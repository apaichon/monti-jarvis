package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"strings"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/store"
	"github.com/libra/monti-jarvis/internal/workforce"
)

var tenantToolKeyRE = regexp.MustCompile(`^[a-z][a-z0-9_]{1,63}$`)
var tenantSkillSlugRE = regexp.MustCompile(`^[a-z][a-z0-9-]{1,63}$`)

type tenantPromptBody struct {
	SystemPrompt string `json:"system_prompt"`
	Enabled      *bool  `json:"enabled"`
}

type tenantGeminiKeyBody struct {
	APIKey string `json:"api_key"`
}

type tenantToolBody struct {
	ToolKey     string         `json:"tool_key"`
	DisplayName string         `json:"display_name"`
	Description string         `json:"description"`
	HandlerKey  string         `json:"handler_key"`
	InputSchema map[string]any `json:"input_schema"`
	Enabled     bool           `json:"enabled"`
}

type tenantSkillBody struct {
	Slug     string   `json:"slug"`
	Name     string   `json:"name"`
	Prompt   string   `json:"prompt"`
	ToolIDs  []string `json:"tool_ids"`
	AgentIDs []string `json:"agent_ids"`
	Enabled  bool     `json:"enabled"`
}

func (s *server) tenantAIID(r *http.Request) (string, bool) {
	ac, ok := auth.FromContext(r.Context())
	if !ok || strings.TrimSpace(ac.TenantID) == "" || ac.Role != auth.RoleTenantAdmin {
		return "", false
	}
	return ac.TenantID, true
}

func writeTenantAIError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrTenantSecretNotConfigured):
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "tenant secret encryption is not configured", "code": "secret_not_configured"})
	case errors.Is(err, store.ErrTenantSecretInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid provider key", "code": "secret_invalid"})
	case errors.Is(err, store.ErrTenantPromptInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "prompt is invalid or too long", "code": "prompt_invalid"})
	case errors.Is(err, store.ErrTenantToolInUse):
		writeJSON(w, http.StatusConflict, map[string]any{"error": "tool is assigned to a skill", "code": "tool_in_use"})
	case errors.Is(err, store.ErrTenantToolInvalid), errors.Is(err, store.ErrTenantSkillInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid tenant AI configuration", "code": "validation_error"})
	case errors.Is(err, store.ErrTenantAIConfigNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "tenant AI resource not found", "code": "not_found"})
	default:
		if err != nil && strings.Contains(err.Error(), "duplicate") {
			writeJSON(w, http.StatusConflict, map[string]any{"error": "resource already exists", "code": "duplicate_key"})
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
	}
}

func (s *server) getTenantGeminiKey(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantAIID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	row, err := s.store.GetTenantAIConfig(r.Context(), tenantID)
	if err != nil {
		writeTenantAIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (s *server) putTenantGeminiKey(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantAIID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body tenantGeminiKeyBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON", "code": "validation_error"})
		return
	}
	row, err := s.store.PutTenantGeminiKey(r.Context(), tenantID, body.APIKey)
	if err != nil {
		writeTenantAIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (s *server) deleteTenantGeminiKey(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantAIID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := s.store.DeleteTenantGeminiKey(r.Context(), tenantID); err != nil {
		writeTenantAIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"configured": false})
}

func (s *server) getTenantPrompt(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantAIID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	agentID := strings.TrimSpace(r.PathValue("agent_id"))
	if !s.tenantAgentKnown(r, tenantID, agentID) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "agent not found", "code": "agent_not_found"})
		return
	}
	row, err := s.store.GetTenantAgentConfig(r.Context(), tenantID, agentID)
	if err != nil {
		writeTenantAIError(w, err)
		return
	}
	row.SystemPrompt = strings.TrimSpace(row.SystemPrompt)
	writeJSON(w, http.StatusOK, map[string]any{"agent_id": row.AgentID, "enabled": row.Enabled, "system_prompt": row.SystemPrompt, "max_length": 8000, "updated_at": row.UpdatedAt})
}

func (s *server) putTenantPrompt(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantAIID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	agentID := strings.TrimSpace(r.PathValue("agent_id"))
	if !s.tenantAgentKnown(r, tenantID, agentID) {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "agent not found", "code": "agent_not_found"})
		return
	}
	var body tenantPromptBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON", "code": "validation_error"})
		return
	}
	enabled := true
	if body.Enabled != nil {
		enabled = *body.Enabled
	}
	row, err := s.store.PutTenantAgentConfig(r.Context(), tenantID, agentID, body.SystemPrompt, enabled)
	if err != nil {
		writeTenantAIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (s *server) tenantAgentKnown(r *http.Request, tenantID, agentID string) bool {
	if agentID == "" {
		return false
	}
	if s.store != nil && s.store.HasTenantAvatarAssignments(r.Context(), tenantID) {
		agents, err := s.store.ListWorkforceAgents(r.Context(), tenantID)
		if err != nil {
			return false
		}
		for _, agent := range agents {
			if strings.EqualFold(agent.ID, agentID) {
				return true
			}
		}
		return false
	}
	_, ok := workforce.Get(agentID)
	return ok
}

func (s *server) listTenantTools(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantAIID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := s.store.ListTenantTools(r.Context(), tenantID)
	if err != nil {
		writeTenantAIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tools": items})
}

func (s *server) createTenantTool(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantAIID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body tenantToolBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON", "code": "validation_error"})
		return
	}
	if !tenantToolKeyRE.MatchString(body.ToolKey) || body.HandlerKey != "create_ticket" {
		writeTenantAIError(w, store.ErrTenantToolInvalid)
		return
	}
	item, err := s.store.CreateTenantTool(r.Context(), store.TenantCallTool{TenantID: tenantID, ToolKey: body.ToolKey, DisplayName: body.DisplayName, Description: body.Description, HandlerKey: body.HandlerKey, InputSchema: body.InputSchema, Enabled: body.Enabled})
	if err != nil {
		writeTenantAIError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (s *server) updateTenantTool(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantAIID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body tenantToolBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON", "code": "validation_error"})
		return
	}
	if !tenantToolKeyRE.MatchString(body.ToolKey) || body.HandlerKey != "create_ticket" {
		writeTenantAIError(w, store.ErrTenantToolInvalid)
		return
	}
	item, err := s.store.UpdateTenantTool(r.Context(), store.TenantCallTool{ID: r.PathValue("id"), TenantID: tenantID, ToolKey: body.ToolKey, DisplayName: body.DisplayName, Description: body.Description, HandlerKey: body.HandlerKey, InputSchema: body.InputSchema, Enabled: body.Enabled})
	if err != nil {
		writeTenantAIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *server) deleteTenantTool(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantAIID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := s.store.DeleteTenantTool(r.Context(), tenantID, r.PathValue("id")); err != nil {
		writeTenantAIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "deleted"})
}

func (s *server) listTenantSkills(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantAIID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := s.store.ListTenantSkills(r.Context(), tenantID)
	if err != nil {
		writeTenantAIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"skills": items})
}

func (s *server) createTenantSkill(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantAIID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body tenantSkillBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON", "code": "validation_error"})
		return
	}
	if !tenantSkillSlugRE.MatchString(body.Slug) {
		writeTenantAIError(w, store.ErrTenantSkillInvalid)
		return
	}
	item, err := s.store.CreateTenantSkill(r.Context(), store.TenantSkill{TenantID: tenantID, Slug: body.Slug, Name: body.Name, Prompt: body.Prompt, ToolIDs: body.ToolIDs, AgentIDs: body.AgentIDs, Enabled: body.Enabled})
	if err != nil {
		writeTenantAIError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, item)
}

func (s *server) updateTenantSkill(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantAIID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body tenantSkillBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON", "code": "validation_error"})
		return
	}
	if !tenantSkillSlugRE.MatchString(body.Slug) {
		writeTenantAIError(w, store.ErrTenantSkillInvalid)
		return
	}
	item, err := s.store.UpdateTenantSkill(r.Context(), store.TenantSkill{ID: r.PathValue("id"), TenantID: tenantID, Slug: body.Slug, Name: body.Name, Prompt: body.Prompt, ToolIDs: body.ToolIDs, AgentIDs: body.AgentIDs, Enabled: body.Enabled})
	if err != nil {
		writeTenantAIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}

func (s *server) deleteTenantSkill(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantAIID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := s.store.DeleteTenantSkill(r.Context(), tenantID, r.PathValue("id")); err != nil {
		writeTenantAIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"status": "deleted"})
}

func (s *server) assignTenantSkill(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantAIID(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body tenantSkillBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid JSON", "code": "validation_error"})
		return
	}
	item, err := s.store.UpdateTenantSkillAssignments(r.Context(), tenantID, r.PathValue("id"), body.ToolIDs, body.AgentIDs)
	if err != nil {
		writeTenantAIError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, item)
}
