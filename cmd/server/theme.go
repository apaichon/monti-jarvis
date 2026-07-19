package main

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/store"
)

const maxThemeLogoBytes = 1 << 20 // 1MB

type putThemeBody struct {
	Preset   string              `json:"preset"`
	Branding store.ThemeBranding `json:"branding"`
	Tokens   store.ThemeTokens   `json:"tokens"`
}

type publishThemeBody struct {
	ConfirmLowContrast bool `json:"confirm_low_contrast"`
}

type resetThemeBody struct {
	Preset         string `json:"preset"`
	ResetBranding  bool   `json:"reset_branding"`
}

func writeThemeError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrInvalidThemePreset):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid preset", "code": "invalid_preset"})
	case errors.Is(err, store.ErrInvalidThemeTokens):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "code": "invalid_theme_tokens"})
	case errors.Is(err, store.ErrInvalidThemeBranding):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "code": "invalid_theme_branding"})
	case errors.Is(err, store.ErrContrastConfirmationNeeded):
		writeJSON(w, http.StatusConflict, map[string]any{
			"error": "contrast confirmation required",
			"code":  "contrast_confirmation_required",
		})
	case errors.Is(err, store.ErrThemeNotFound):
		writeError(w, http.StatusNotFound, "theme not found")
	default:
		if err != nil && strings.Contains(err.Error(), "unsupported image") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
	}
}

func themeJSON(row *store.TenantTheme) map[string]any {
	if row == nil {
		return map[string]any{}
	}
	out := map[string]any{
		"tenant_id":           row.TenantID,
		"preset":              row.Preset,
		"draft_branding":      row.DraftBranding,
		"published_branding":  row.PublishedBranding,
		"draft_tokens":        row.DraftTokens,
		"published_tokens":    row.PublishedTokens,
		"draft_updated_at":    row.DraftUpdatedAt,
		"created_at":          row.CreatedAt,
		"updated_at":          row.UpdatedAt,
		"contrast_report":     row.ContrastReport,
	}
	if row.PublishedAt != nil {
		out["published_at"] = row.PublishedAt
	} else {
		out["published_at"] = nil
	}
	return out
}

// GET /api/tenant/theme
func (s *server) getTenantTheme(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	row, err := s.store.GetOrCreateTenantTheme(r.Context(), tenantID)
	if err != nil {
		writeThemeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, themeJSON(row))
}

// PUT /api/tenant/theme
func (s *server) putTenantTheme(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body putThemeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	row, err := s.store.UpdateTenantThemeDraft(r.Context(), tenantID, body.Preset, body.Branding, body.Tokens)
	if err != nil {
		writeThemeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, themeJSON(row))
}

// POST /api/tenant/theme/publish
func (s *server) publishTenantTheme(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body publishThemeBody
	_ = json.NewDecoder(r.Body).Decode(&body)
	row, err := s.store.PublishTenantTheme(r.Context(), tenantID, body.ConfirmLowContrast)
	if err != nil {
		if errors.Is(err, store.ErrContrastConfirmationNeeded) {
			// include contrast report for client
			if cur, e2 := s.store.GetOrCreateTenantTheme(r.Context(), tenantID); e2 == nil {
				writeJSON(w, http.StatusConflict, map[string]any{
					"error":           "contrast confirmation required",
					"code":            "contrast_confirmation_required",
					"contrast_report": cur.ContrastReport,
				})
				return
			}
		}
		writeThemeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, themeJSON(row))
}

// POST /api/tenant/theme/reset
func (s *server) resetTenantTheme(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body resetThemeBody
	_ = json.NewDecoder(r.Body).Decode(&body)
	preset := body.Preset
	if preset == "" {
		preset = "dark"
	}
	row, err := s.store.ResetTenantThemeDraft(r.Context(), tenantID, preset, body.ResetBranding)
	if err != nil {
		writeThemeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, themeJSON(row))
}

// POST /api/tenant/theme/logo
func (s *server) uploadTenantThemeLogo(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	if err := r.ParseMultipartForm(maxThemeLogoBytes); err != nil {
		writeError(w, http.StatusBadRequest, "invalid multipart form")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeError(w, http.StatusBadRequest, "file is required")
		return
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, maxThemeLogoBytes+1))
	if err != nil {
		writeError(w, http.StatusBadRequest, "could not read file")
		return
	}
	if len(data) > maxThemeLogoBytes {
		writeError(w, http.StatusBadRequest, "image exceeds 1MB limit")
		return
	}
	contentType := strings.TrimSpace(header.Header.Get("Content-Type"))
	if contentType == "" || contentType == "application/octet-stream" {
		contentType = mimeFromFilename(header.Filename)
	}
	ctx, cancel := contextWithTimeout(r, 30*time.Second)
	defer cancel()
	_, logoURL, err := s.store.PutThemeLogo(ctx, tenantID, contentType, data)
	if err != nil {
		writeThemeError(w, err)
		return
	}
	// Merge into draft branding
	cur, err := s.store.GetOrCreateTenantTheme(ctx, tenantID)
	if err != nil {
		writeThemeError(w, err)
		return
	}
	branding := cur.DraftBranding
	branding.LogoURL = logoURL
	if branding.LogoAlt == "" {
		branding.LogoAlt = branding.BrandName
	}
	row, err := s.store.UpdateTenantThemeDraft(ctx, tenantID, cur.Preset, branding, cur.DraftTokens)
	if err != nil {
		writeThemeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"logo_url": logoURL,
		"theme":    themeJSON(row),
	})
}

// GET /api/public/theme/{tenant_id}
func (s *server) getPublicTheme(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "tenant_id required")
		return
	}
	pub, err := s.store.GetPublishedTheme(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			// still return system default for unknown? prefer 404
			writeError(w, http.StatusNotFound, "tenant not found")
			return
		}
		writeThemeError(w, err)
		return
	}
	w.Header().Set("Access-Control-Allow-Origin", "*")
	writeJSON(w, http.StatusOK, map[string]any{
		"tenant_id": pub.TenantID,
		"preset":    pub.Preset,
		"source":    pub.Source,
		"branding":  pub.Branding,
		"tokens":    pub.Tokens,
		"css_vars":  store.CSSVarMap(pub.Tokens),
	})
}

// GET /api/admin/tenants/{tenant_id}/theme
func (s *server) getAdminTenantTheme(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "tenant_id required")
		return
	}
	row, err := s.store.GetTenantTheme(r.Context(), tenantID)
	if err != nil {
		if errors.Is(err, store.ErrThemeNotFound) {
			writeJSON(w, http.StatusOK, map[string]any{
				"tenant_id":    tenantID,
				"preset":       "dark",
				"has_logo":     false,
				"brand_name":   "",
				"published_at": nil,
				"contrast_ok":  true,
				"source":       "system_default",
			})
			return
		}
		writeThemeError(w, err)
		return
	}
	hasLogo := strings.TrimSpace(row.PublishedBranding.LogoURL) != "" || strings.TrimSpace(row.DraftBranding.LogoURL) != ""
	brandName := row.PublishedBranding.BrandName
	if brandName == "" {
		brandName = row.DraftBranding.BrandName
	}
	report := store.EvaluateContrast(row.DraftTokens)
	out := map[string]any{
		"tenant_id":    tenantID,
		"preset":       row.Preset,
		"has_logo":     hasLogo,
		"brand_name":   brandName,
		"contrast_ok":  report.OK,
		"source":       "draft",
	}
	if len(row.PublishedTokens) > 0 {
		out["source"] = "published"
	}
	if row.PublishedAt != nil {
		out["published_at"] = row.PublishedAt
	} else {
		out["published_at"] = nil
	}
	writeJSON(w, http.StatusOK, out)
}

// GET /api/assets/theme/{tenant_id}/{file}
func (s *server) serveThemeAsset(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	filename := strings.TrimSpace(r.PathValue("file"))
	if tenantID == "" || filename == "" {
		writeError(w, http.StatusBadRequest, "invalid asset path")
		return
	}
	ctx, cancel := contextWithTimeout(r, 15*time.Second)
	defer cancel()
	rc, ct, err := s.store.GetThemeAsset(ctx, tenantID, filename)
	if err != nil {
		writeError(w, http.StatusNotFound, "asset not found")
		return
	}
	defer rc.Close()
	w.Header().Set("Content-Type", ct)
	w.Header().Set("Cache-Control", "public, max-age=3600")
	_, _ = io.Copy(w, rc)
}

func attachPublishedTheme(ctx context.Context, st *store.Store, tenantID string) map[string]any {
	if st == nil {
		return nil
	}
	pub, err := st.GetPublishedTheme(ctx, tenantID)
	if err != nil {
		return nil
	}
	return map[string]any{
		"preset":   pub.Preset,
		"source":   pub.Source,
		"branding": pub.Branding,
		"tokens":   pub.Tokens,
		"css_vars": store.CSSVarMap(pub.Tokens),
	}
}

// used by embed resolve
func (s *server) publicThemeForTenant(ctx context.Context, tenantID string) map[string]any {
	return attachPublishedTheme(ctx, s.store, tenantID)
}
