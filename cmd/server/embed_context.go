package main

import (
	"errors"
	"net/http"
	"strings"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/store"
)

var errEmbedOriginNotAllowed = errors.New("embed origin not allowed")

func embedKeyFromRequest(r *http.Request) string {
	if key := strings.TrimSpace(r.Header.Get("X-Monti-Embed-Key")); key != "" {
		return key
	}
	return strings.TrimSpace(r.URL.Query().Get("embed_key"))
}

func (s *server) resolveEmbedContext(r *http.Request) (*store.TenantEmbedConfig, error) {
	key := embedKeyFromRequest(r)
	if key == "" || s.store == nil {
		return nil, nil
	}
	cfg, err := s.store.GetEmbedConfigByKey(r.Context(), key)
	if err != nil {
		return nil, err
	}
	if !cfg.Enabled {
		return nil, store.ErrEmbedDisabled
	}
	parentOrigin := strings.TrimSpace(r.Header.Get("X-Embed-Parent-Origin"))
	if parentOrigin == "" {
		parentOrigin = strings.TrimSpace(r.URL.Query().Get("parent_origin"))
	}
	reqOrigin := store.EmbedCheckOrigin(parentOrigin, r.Header.Get("Origin"), r.Header.Get("Referer"))
	if !store.OriginAllowed(cfg.AllowedOrigins, reqOrigin, s.cfg.EmbedAllowEmptyOrigins) {
		return nil, errEmbedOriginNotAllowed
	}
	return cfg, nil
}

func (s *server) requestTenantContext(r *http.Request) (string, *store.TenantEmbedConfig, error) {
	cfg, err := s.resolveEmbedContext(r)
	if err != nil {
		return "", nil, err
	}
	if cfg != nil {
		if hint := strings.TrimSpace(r.Header.Get("X-Tenant-Id")); hint != "" && hint != cfg.TenantID {
			return "", nil, errors.New("tenant context mismatch")
		}
		return cfg.TenantID, cfg, nil
	}
	hint := r.Header.Get("X-Tenant-Id")
	if hint == "" {
		hint = r.URL.Query().Get("tenant_id")
	}
	return auth.ResolveTenant(r.Context(), hint, s.cfg.AuthDisabled, s.cfg.DemoTenantID), nil, nil
}

func writeEmbedContextError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrEmbedNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "embed not found", "code": "embed_not_found"})
	case errors.Is(err, store.ErrEmbedDisabled):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "embed disabled", "code": "embed_disabled"})
	case errors.Is(err, errEmbedOriginNotAllowed):
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "origin not allowed", "code": "origin_not_allowed"})
	case err != nil && strings.Contains(err.Error(), "tenant context mismatch"):
		writeJSON(w, http.StatusForbidden, map[string]any{"error": "tenant context mismatch", "code": "tenant_context_mismatch"})
	default:
		writeError(w, http.StatusBadGateway, err.Error())
	}
}
