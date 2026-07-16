package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/libra/monti-jarvis/internal/store"
)

type brandProfileRequest struct {
	Name      string   `json:"name"`
	Blurb     string   `json:"blurb"`
	LogoURL   string   `json:"logo_url"`
	Category  string   `json:"category"`
	Languages []string `json:"languages"`
	Listed    *bool    `json:"listed"`
}

func (s *server) publicBrands(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	limit, offset = normalizePublicBrandPageRequest(limit, offset)
	items, total, err := s.store.ListPublicBrands(r.Context(), r.URL.Query().Get("q"), limit, offset)
	if err != nil {
		writeError(w, http.StatusBadGateway, "public brand directory unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"items": items, "total": total, "limit": limit, "offset": offset})
}

func (s *server) publicBrand(w http.ResponseWriter, r *http.Request) {
	brand, err := s.store.GetPublicBrand(r.Context(), strings.TrimSpace(r.PathValue("slug")))
	if err != nil {
		if errors.Is(err, store.ErrTenantNotFound) {
			writeError(w, http.StatusNotFound, "brand not found")
			return
		}
		writeError(w, http.StatusBadGateway, "public brand directory unavailable")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"item": brand})
}

func (s *server) putTenantBrand(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var req brandProfileRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	brand, err := s.store.PutBrandProfile(r.Context(), tenantID, store.BrandProfileInput{
		Name: req.Name, Blurb: req.Blurb, LogoURL: req.LogoURL, Category: req.Category, Languages: req.Languages, Listed: req.Listed,
	})
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"item": brand})
}

func decodeJSON(r *http.Request, dst any) error {
	return json.NewDecoder(r.Body).Decode(dst)
}

func (s *server) putPlatformBrandListing(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	var req struct {
		Listed bool `json:"listed"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	if err := s.store.SetBrandPlatformListed(r.Context(), tenantID, req.Listed); err != nil {
		writeError(w, http.StatusBadRequest, "brand listing update failed")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tenant_id": tenantID, "platform_listed": req.Listed})
}

func normalizePublicBrandPageRequest(limit, offset int) (int, int) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}
