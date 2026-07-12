package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/store"
	"github.com/libra/monti-jarvis/internal/workforce"
)

type putEmbedBody struct {
	Enabled        *bool    `json:"enabled"`
	AllowedOrigins []string `json:"allowed_origins"`
	DefaultAgentID *string  `json:"default_agent_id"`
}

func writeEmbedError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrEmbedNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "embed not found", "code": "embed_not_found"})
	case errors.Is(err, store.ErrEmbedDisabled):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "embed disabled", "code": "embed_disabled"})
	case errors.Is(err, store.ErrTenantNotFound):
		writeError(w, http.StatusNotFound, "tenant not found")
	default:
		if err != nil && strings.Contains(err.Error(), "invalid origin") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
	}
}

func (s *server) getPublicEmbed(w http.ResponseWriter, r *http.Request) {
	key := strings.TrimSpace(r.PathValue("embed_key"))
	if key == "" {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "embed not found", "code": "embed_not_found"})
		return
	}
	cfg, err := s.store.GetEmbedConfigByKey(r.Context(), key)
	if err != nil {
		if errors.Is(err, store.ErrEmbedNotFound) {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "embed not found", "code": "embed_not_found"})
			return
		}
		writeEmbedError(w, err)
		return
	}
	if !cfg.Enabled {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "embed disabled", "code": "embed_disabled"})
		return
	}

	// Iframe UI is same-origin to Monti, so browser Origin is Monti — not the host site.
	// Loader passes parent_origin (query or X-Embed-Parent-Origin); prefer that for allowlist.
	parentOrigin := r.URL.Query().Get("parent_origin")
	if parentOrigin == "" {
		parentOrigin = r.Header.Get("X-Embed-Parent-Origin")
	}
	reqOrigin := store.EmbedCheckOrigin(parentOrigin, r.Header.Get("Origin"), r.Header.Get("Referer"))
	allowEmpty := s.cfg.EmbedAllowEmptyOrigins
	if !store.OriginAllowed(cfg.AllowedOrigins, reqOrigin, allowEmpty) {
		writeJSON(w, http.StatusForbidden, map[string]any{
			"error": "origin not allowed",
			"code":  "origin_not_allowed",
		})
		return
	}

	// CORS for public embed: echo request origin if allowed (or * when empty list + allowEmpty).
	// When resolve is called from the host page (not iframe), browser Origin is the host.
	corsOrigin := store.RequestOrigin(r.Header.Get("Origin"), r.Header.Get("Referer"))
	if len(cfg.AllowedOrigins) == 0 && allowEmpty {
		w.Header().Set("Access-Control-Allow-Origin", "*")
	} else if corsOrigin != "" && store.OriginAllowed(cfg.AllowedOrigins, corsOrigin, false) {
		w.Header().Set("Access-Control-Allow-Origin", corsOrigin)
		w.Header().Set("Vary", "Origin")
	} else if reqOrigin != "" && store.OriginAllowed(cfg.AllowedOrigins, reqOrigin, false) {
		// Same-origin iframe fetch: allow host to call with parent_origin if needed later
		w.Header().Set("Access-Control-Allow-Origin", reqOrigin)
		w.Header().Set("Vary", "Origin")
	}
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Tenant-Id, X-Embed-Parent-Origin")

	tenant, err := s.store.GetTenant(r.Context(), cfg.TenantID)
	if err != nil {
		writeEmbedError(w, err)
		return
	}

	agents := s.embedAgentsJSON(r, cfg.TenantID)
	writeJSON(w, http.StatusOK, map[string]any{
		"tenant_id":        cfg.TenantID,
		"slug":             tenant.Slug,
		"name":             tenant.Name,
		"embed_key":        cfg.EmbedKey,
		"enabled":          cfg.Enabled,
		"default_agent_id": cfg.DefaultAgentID,
		"agents":           agents,
	})
}

func (s *server) getTenantEmbed(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok || strings.TrimSpace(ac.TenantID) == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	cfg, err := s.store.GetOrCreateEmbedConfig(r.Context(), ac.TenantID)
	if err != nil {
		writeEmbedError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, embedConfigJSON(cfg, s.cfg.PublicBaseURL))
}

func (s *server) putTenantEmbed(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok || strings.TrimSpace(ac.TenantID) == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body putEmbedBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	// Load current for partial update defaults
	cur, err := s.store.GetOrCreateEmbedConfig(r.Context(), ac.TenantID)
	if err != nil {
		writeEmbedError(w, err)
		return
	}
	enabled := cur.Enabled
	if body.Enabled != nil {
		enabled = *body.Enabled
	}
	origins := cur.AllowedOrigins
	if body.AllowedOrigins != nil {
		origins = body.AllowedOrigins
	}
	defaultAgent := cur.DefaultAgentID
	if body.DefaultAgentID != nil {
		defaultAgent = strings.TrimSpace(*body.DefaultAgentID)
	}
	cfg, err := s.store.UpdateEmbedConfig(r.Context(), ac.TenantID, enabled, origins, defaultAgent)
	if err != nil {
		writeEmbedError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, embedConfigJSON(cfg, s.cfg.PublicBaseURL))
}

func (s *server) rotateTenantEmbedKey(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok || strings.TrimSpace(ac.TenantID) == "" {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	cfg, err := s.store.RotateEmbedKey(r.Context(), ac.TenantID)
	if err != nil {
		writeEmbedError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, embedConfigJSON(cfg, s.cfg.PublicBaseURL))
}

func embedConfigJSON(cfg *store.TenantEmbedConfig, publicBase string) map[string]any {
	base := strings.TrimRight(publicBase, "/")
	snippet := `<script src="` + base + `/embed/monti-embed.js" data-embed-key="` + cfg.EmbedKey + `" data-position="bottom-right" async></script>`
	return map[string]any{
		"tenant_id":        cfg.TenantID,
		"embed_key":        cfg.EmbedKey,
		"enabled":          cfg.Enabled,
		"allowed_origins":  cfg.AllowedOrigins,
		"default_agent_id": cfg.DefaultAgentID,
		"snippet":          snippet,
		"created_at":       cfg.CreatedAt,
		"updated_at":       cfg.UpdatedAt,
	}
}

func (s *server) embedAgentsJSON(r *http.Request, tenantID string) []map[string]any {
	ctx := r.Context()
	if s.store != nil && s.store.HasTenantAvatarAssignments(ctx, tenantID) {
		list, err := s.store.ListWorkforceAgents(ctx, tenantID)
		if err == nil && len(list) > 0 {
			out := make([]map[string]any, 0, len(list))
			for _, a := range list {
				out = append(out, map[string]any{
					"id": a.ID, "name": a.Name, "role": a.Role, "image": a.Image,
				})
			}
			return out
		}
	}
	all := workforce.All()
	out := make([]map[string]any, 0, len(all))
	for _, a := range all {
		out = append(out, map[string]any{
			"id": a.ID, "name": a.Name, "role": a.Role, "image": a.Image,
		})
	}
	return out
}
