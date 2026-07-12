package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/libra/monti-jarvis/internal/store"
)

type tierBody struct {
	Name                 *string `json:"name"`
	Slug                 *string `json:"slug"`
	Priority             *int    `json:"priority"`
	Description          *string `json:"description"`
	DefaultAgentID       *string `json:"default_agent_id"`
	AIReplyLocale        *string `json:"ai_reply_locale"`
	MaxMinutesPerCall    *int    `json:"max_minutes_per_call"`
	MaxCallMinutesPerDay *int    `json:"max_call_minutes_per_day"`
	Active               *bool   `json:"active"`
}

type groupBody struct {
	Name        *string `json:"name"`
	Slug        *string `json:"slug"`
	Description *string `json:"description"`
}

func writeTierError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrTierNotFound), errors.Is(err, store.ErrGroupNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error(), "code": "not_found"})
	case errors.Is(err, store.ErrInvalidSlug):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid slug", "code": "invalid_slug"})
	case errors.Is(err, store.ErrSlugTaken):
		writeJSON(w, http.StatusConflict, map[string]any{"error": "slug already exists", "code": "slug_taken"})
	case errors.Is(err, store.ErrInvalidLocale):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "locale must be en or th", "code": "invalid_locale"})
	default:
		if err != nil && (strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must be")) {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
	}
}

func (s *server) listTenantTiers(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	list, err := s.store.ListCustomerTiers(r.Context(), tenantID)
	if err != nil {
		writeTierError(w, err)
		return
	}
	if list == nil {
		list = []store.CustomerTier{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"tiers": list})
}

func (s *server) createTenantTier(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body tierBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	name, slug := "", ""
	if body.Name != nil {
		name = *body.Name
	}
	if body.Slug != nil {
		slug = *body.Slug
	}
	in := store.CreateCustomerTierInput{
		Name:     name,
		Slug:     slug,
		Priority: 0,
		Active:   body.Active,
	}
	if body.Priority != nil {
		in.Priority = *body.Priority
	}
	if body.Description != nil {
		in.Description = *body.Description
	}
	if body.DefaultAgentID != nil {
		in.DefaultAgentID = *body.DefaultAgentID
	}
	if body.AIReplyLocale != nil {
		in.AIReplyLocale = *body.AIReplyLocale
	}
	if body.MaxMinutesPerCall != nil {
		in.MaxMinutesPerCall = *body.MaxMinutesPerCall
	}
	if body.MaxCallMinutesPerDay != nil {
		in.MaxCallMinutesPerDay = *body.MaxCallMinutesPerDay
	}
	row, err := s.store.CreateCustomerTier(r.Context(), tenantID, in)
	if err != nil {
		writeTierError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, row)
}

func (s *server) getTenantTier(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	row, err := s.store.GetCustomerTier(r.Context(), tenantID, id)
	if err != nil {
		writeTierError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (s *server) putTenantTier(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	var body tierBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	row, err := s.store.UpdateCustomerTier(r.Context(), tenantID, id, store.UpdateCustomerTierInput{
		Name:                 body.Name,
		Slug:                 body.Slug,
		Priority:             body.Priority,
		Description:          body.Description,
		DefaultAgentID:       body.DefaultAgentID,
		AIReplyLocale:        body.AIReplyLocale,
		MaxMinutesPerCall:    body.MaxMinutesPerCall,
		MaxCallMinutesPerDay: body.MaxCallMinutesPerDay,
		Active:               body.Active,
	})
	if err != nil {
		writeTierError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (s *server) deleteTenantTier(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if err := s.store.DeleteCustomerTier(r.Context(), tenantID, id); err != nil {
		writeTierError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id})
}

func (s *server) listTenantGroups(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	list, err := s.store.ListCustomerGroups(r.Context(), tenantID)
	if err != nil {
		writeTierError(w, err)
		return
	}
	if list == nil {
		list = []store.CustomerGroup{}
	}
	writeJSON(w, http.StatusOK, map[string]any{"groups": list})
}

func (s *server) createTenantGroup(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body groupBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	name, slug, desc := "", "", ""
	if body.Name != nil {
		name = *body.Name
	}
	if body.Slug != nil {
		slug = *body.Slug
	}
	if body.Description != nil {
		desc = *body.Description
	}
	row, err := s.store.CreateCustomerGroup(r.Context(), tenantID, name, slug, desc)
	if err != nil {
		writeTierError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, row)
}

func (s *server) getTenantGroup(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	row, err := s.store.GetCustomerGroup(r.Context(), tenantID, id)
	if err != nil {
		writeTierError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (s *server) putTenantGroup(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	var body groupBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	cur, err := s.store.GetCustomerGroup(r.Context(), tenantID, id)
	if err != nil {
		writeTierError(w, err)
		return
	}
	name, slug, desc := cur.Name, cur.Slug, cur.Description
	if body.Name != nil {
		name = *body.Name
	}
	if body.Slug != nil {
		slug = *body.Slug
	}
	if body.Description != nil {
		desc = *body.Description
	}
	row, err := s.store.UpdateCustomerGroup(r.Context(), tenantID, id, name, slug, desc)
	if err != nil {
		writeTierError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (s *server) deleteTenantGroup(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if err := s.store.DeleteCustomerGroup(r.Context(), tenantID, id); err != nil {
		writeTierError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id})
}
