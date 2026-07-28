package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/libra/monti-jarvis/internal/quota"
	"github.com/libra/monti-jarvis/internal/store"
)

type tenantAvatarCreateBody struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	Trait    string `json:"trait"`
	Color    string `json:"color"`
	ImageURL string `json:"image_url"`
	Greeting string `json:"greeting"`
	// Voice is a Gemini AI Studio speaker setting name (e.g. Aoede, Puck, Kore).
	Voice string `json:"voice"`
}

type tenantAvatarUpdateBody struct {
	Name     string `json:"name"`
	Role     string `json:"role"`
	Trait    string `json:"trait"`
	Color    string `json:"color"`
	ImageURL string `json:"image_url"`
	Greeting string `json:"greeting"`
	Voice    string `json:"voice"`
}

func (s *server) listTenantLibraryAvatars(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	items, err := s.store.ListTenantLibraryAvatars(r.Context(), tenantID)
	if err != nil {
		writeAvatarError(w, err)
		return
	}
	activeCount, err := s.store.CountActiveTenantAssignments(r.Context(), tenantID)
	if err != nil {
		writeAvatarError(w, err)
		return
	}
	limit := s.effectiveAIEmployeeLimit(r, tenantID)
	remaining := limit - activeCount
	if remaining < 0 {
		remaining = 0
	}
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		// Ensure primary speaker is present for UI dropdown selection.
		if len(it.Avatar.Voices) == 0 {
			if name := s.store.PrimarySpeakerName(r.Context(), it.Avatar.ID); name != "" {
				it.Avatar.Voices = []store.AvatarVoice{{Voice: name, Status: "active", Priority: 1}}
			}
		}
		out = append(out, tenantLibraryAvatarJSON(it))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"avatars": out,
		"cap": map[string]any{
			"active":    activeCount,
			"limit":     limit,
			"remaining": remaining,
		},
	})
}

func (s *server) createTenantLibraryAvatar(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body tenantAvatarCreateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	name := strings.TrimSpace(body.Name)
	if name == "" {
		writeError(w, http.StatusBadRequest, "name is required")
		return
	}
	speaker := strings.TrimSpace(body.Voice)
	if speaker == "" {
		speaker = "Aoede"
	}
	if _, ok := store.CanonicalGeminiSpeaker(speaker); !ok {
		writeError(w, http.StatusBadRequest, "invalid voice; choose a Gemini AI Studio speaker name")
		return
	}
	// Create path: NO CheckAIEmployees — library can grow beyond package active cap.
	av := store.Avatar{
		Name:     name,
		Role:     strings.TrimSpace(body.Role),
		Trait:    strings.TrimSpace(body.Trait),
		Color:    strings.TrimSpace(body.Color),
		ImageURL: strings.TrimSpace(body.ImageURL),
		Greeting: strings.TrimSpace(body.Greeting),
		Status:   "active",
		Flags:    map[string]any{},
		Voices: []store.AvatarVoice{{
			Voice: speaker,
		}},
	}
	if av.ImageURL == "" {
		av.ImageURL = "/images/default-avatar.jpg"
	}
	created, err := s.store.CreateTenantLibraryAvatar(r.Context(), tenantID, av)
	if err != nil {
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, "slug already exists")
			return
		}
		writeAvatarError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, tenantLibraryAvatarJSON(*created))
}

func (s *server) getTenantLibraryAvatar(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	item, err := s.store.GetTenantLibraryAvatar(r.Context(), tenantID, id)
	if err != nil {
		writeAvatarError(w, err)
		return
	}
	if len(item.Avatar.Voices) == 0 {
		if name := s.store.PrimarySpeakerName(r.Context(), item.Avatar.ID); name != "" {
			item.Avatar.Voices = []store.AvatarVoice{{Voice: name, Status: "active", Priority: 1}}
		}
	}
	writeJSON(w, http.StatusOK, tenantLibraryAvatarJSON(*item))
}

func (s *server) updateTenantLibraryAvatar(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	existing, err := s.store.GetTenantLibraryAvatar(r.Context(), tenantID, id)
	if err != nil {
		writeAvatarError(w, err)
		return
	}
	if existing.Avatar.OwnerTenantID != tenantID {
		writeError(w, http.StatusForbidden, "only tenant-owned avatars can be edited here")
		return
	}
	var body tenantAvatarUpdateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	av := existing.Avatar
	if n := strings.TrimSpace(body.Name); n != "" {
		av.Name = n
	}
	if body.Role != "" {
		av.Role = strings.TrimSpace(body.Role)
	}
	if body.Trait != "" {
		av.Trait = strings.TrimSpace(body.Trait)
	}
	if body.Color != "" {
		av.Color = strings.TrimSpace(body.Color)
	}
	if body.ImageURL != "" {
		av.ImageURL = strings.TrimSpace(body.ImageURL)
	}
	if body.Greeting != "" {
		av.Greeting = strings.TrimSpace(body.Greeting)
	}
	updated, err := s.store.UpdateTenantOwnedAvatar(r.Context(), tenantID, av)
	if err != nil {
		writeAvatarError(w, err)
		return
	}
	if v := strings.TrimSpace(body.Voice); v != "" {
		if err := s.store.SetTenantAvatarSpeaker(r.Context(), tenantID, id, v); err != nil {
			if strings.Contains(err.Error(), "invalid gemini") {
				writeError(w, http.StatusBadRequest, "invalid voice; choose a Gemini AI Studio speaker name")
				return
			}
			writeAvatarError(w, err)
			return
		}
		if refreshed, gerr := s.store.GetAvatar(r.Context(), id); gerr == nil {
			updated = refreshed
		}
	}
	item := store.TenantLibraryAvatar{Avatar: *updated, AssignmentStatus: existing.AssignmentStatus}
	writeJSON(w, http.StatusOK, tenantLibraryAvatarJSON(item))
}

func (s *server) listGeminiSpeakerVoices(w http.ResponseWriter, r *http.Request) {
	if _, ok := s.tenantIDFromAuth(r); !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"source":  "https://aistudio.google.com/generate-speech",
		"docs":    "https://ai.google.dev/gemini-api/docs/speech-generation#voices",
		"provider": store.GeminiLiveProviderID,
		"voices":  store.GeminiSpeakerVoices(),
	})
}

func (s *server) activateTenantLibraryAvatar(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	item, err := s.store.GetTenantLibraryAvatar(r.Context(), tenantID, id)
	if err != nil {
		writeAvatarError(w, err)
		return
	}
	if item.AssignmentStatus == "active" {
		writeJSON(w, http.StatusOK, tenantLibraryAvatarJSON(*item))
		return
	}
	// Active cap: package max_ai_employees (+ bonus via quota service).
	if err := s.checkAvatarAssignCap(r.Context(), tenantID, id); err != nil {
		var qe *quota.Error
		if errors.Is(err, store.ErrMaxAIEmployeesExceeded) || errors.Is(err, quota.ErrLimitExceeded) || errors.As(err, &qe) {
			writeJSON(w, http.StatusConflict, map[string]any{
				"error": "active avatar limit reached for your package",
				"code":  "quota_exceeded",
			})
			return
		}
		writeAvatarError(w, err)
		return
	}
	activated, err := s.store.ActivateTenantAvatar(r.Context(), tenantID, id)
	if err != nil {
		writeAvatarError(w, err)
		return
	}
	if s.quota != nil {
		if count, countErr := s.store.CountActiveTenantAssignments(r.Context(), tenantID); countErr == nil {
			_ = s.quota.ConsumeBonusUsage(r.Context(), tenantID, quota.DimMaxAIEmployees, count, "avatar_assignment", id)
		}
	}
	writeJSON(w, http.StatusOK, tenantLibraryAvatarJSON(*activated))
}

func (s *server) deactivateTenantLibraryAvatar(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if err := s.store.DeactivateTenantAvatar(r.Context(), tenantID, id); err != nil {
		writeAvatarError(w, err)
		return
	}
	item, err := s.store.GetTenantLibraryAvatar(r.Context(), tenantID, id)
	if err != nil {
		writeJSON(w, http.StatusOK, map[string]string{"status": "disabled"})
		return
	}
	writeJSON(w, http.StatusOK, tenantLibraryAvatarJSON(*item))
}

func (s *server) deleteTenantLibraryAvatar(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if err := s.store.ArchiveTenantOwnedAvatar(r.Context(), tenantID, id); err != nil {
		writeAvatarError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "archived"})
}

func (s *server) effectiveAIEmployeeLimit(r *http.Request, tenantID string) int {
	if s.entitlements == nil {
		return 0
	}
	eff, err := s.entitlements.GetEffective(r.Context(), tenantID)
	if err != nil {
		return 0
	}
	return rulesInt(eff.Rules, "max_ai_employees")
}

func tenantLibraryAvatarJSON(item store.TenantLibraryAvatar) map[string]any {
	av := item.Avatar
	voice := ""
	if len(av.Voices) > 0 {
		voice = av.Voices[0].Voice
	}
	// Voices may not be loaded on list path — load primary when empty.
	out := map[string]any{
		"id":                av.ID,
		"slug":              av.Slug,
		"name":              av.Name,
		"role":              av.Role,
		"trait":             av.Trait,
		"color":             av.Color,
		"image_url":         av.ImageURL,
		"greeting":          av.Greeting,
		"status":            av.Status,
		"owner_tenant_id":   av.OwnerTenantID,
		"assignment_status": item.AssignmentStatus,
		"flags":             av.Flags,
		"voice":             voice,
	}
	return out
}
