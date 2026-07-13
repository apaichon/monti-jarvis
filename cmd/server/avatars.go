package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/libra/monti-jarvis/internal/quota"
	"github.com/libra/monti-jarvis/internal/store"
)

func (s *server) listAvatars(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	avatars, err := s.store.ListAvatars(r.Context(), status)
	if err != nil {
		writeAvatarError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"avatars": avatarListJSON(avatars)})
}

func (s *server) getAvatar(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	av, err := s.store.GetAvatar(r.Context(), id)
	if err != nil {
		writeAvatarError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, avatarJSON(*av))
}

type avatarVoiceBody struct {
	ID              string `json:"id"`
	VoiceProviderID string `json:"voice_provider_id"`
	VoiceID         string `json:"voice_id"`
	Voice           string `json:"voice"`
	Priority        int    `json:"priority"`
	Status          string `json:"status"`
}

type avatarBody struct {
	Slug      string            `json:"slug"`
	Name      string            `json:"name"`
	Role      string            `json:"role"`
	Trait     string            `json:"trait"`
	Color     string            `json:"color"`
	ImageURL  string            `json:"image_url"`
	Greeting  string            `json:"greeting"`
	Status    string            `json:"status"`
	Flags     map[string]any    `json:"flags"`
	Voices    []avatarVoiceBody `json:"voices"`
	VoicesSet bool              `json:"-"`
}

type avatarUpdateBody struct {
	Slug     string             `json:"slug"`
	Name     string             `json:"name"`
	Role     string             `json:"role"`
	Trait    string             `json:"trait"`
	Color    string             `json:"color"`
	ImageURL string             `json:"image_url"`
	Greeting string             `json:"greeting"`
	Status   string             `json:"status"`
	Flags    map[string]any     `json:"flags"`
	Voices   *[]avatarVoiceBody `json:"voices"`
}

func (s *server) createAvatar(w http.ResponseWriter, r *http.Request) {
	var body avatarBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	body.VoicesSet = true
	av, voices, err := s.buildAvatarFromBody(body, "", true)
	if err != nil {
		writeAvatarValidationError(w, err)
		return
	}
	av.Voices = voices
	created, err := s.store.CreateAvatar(r.Context(), *av)
	if err != nil {
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, "slug already exists")
			return
		}
		writeAvatarError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, avatarJSON(*created))
}

func (s *server) updateAvatar(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	existing, err := s.store.GetAvatar(r.Context(), id)
	if err != nil {
		writeAvatarError(w, err)
		return
	}
	var body avatarUpdateBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	merged := mergeAvatarUpdateBody(*existing, body)
	av, voices, err := s.buildAvatarFromBody(merged, id, false)
	if err != nil {
		writeAvatarValidationError(w, err)
		return
	}
	if av.Status == "active" {
		checkVoices := voices
		if !merged.VoicesSet {
			checkVoices = existing.Voices
		}
		if !hasActiveVoice(checkVoices) {
			writeAvatarValidationError(w, errAvatarValidation("active avatars require at least one active voice"))
			return
		}
	}
	updated, err := s.store.UpdateAvatar(r.Context(), *av)
	if err != nil {
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, "slug already exists")
			return
		}
		writeAvatarError(w, err)
		return
	}
	if merged.VoicesSet {
		if err := s.store.ReplaceAvatarVoices(r.Context(), id, voices); err != nil {
			writeAvatarError(w, err)
			return
		}
		refreshed, err := s.store.GetAvatar(r.Context(), id)
		if err != nil {
			writeAvatarError(w, err)
			return
		}
		updated = refreshed
	}
	writeJSON(w, http.StatusOK, avatarJSON(*updated))
}

func (s *server) archiveAvatar(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if err := s.store.ArchiveAvatar(r.Context(), id); err != nil {
		writeAvatarError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "archived"})
}

func (s *server) listTenantAvatars(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	assignments, err := s.store.ListTenantAvatarAssignments(r.Context(), tenantID)
	if err != nil {
		writeAvatarError(w, err)
		return
	}
	activeCount, err := s.store.CountActiveTenantAssignments(r.Context(), tenantID)
	if err != nil {
		writeAvatarError(w, err)
		return
	}
	maxAI := 0
	if eff, err := s.entitlements.GetEffective(r.Context(), tenantID); err == nil {
		maxAI = rulesInt(eff.Rules, "max_ai_employees")
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"tenant_id":   tenantID,
		"assignments": tenantAssignmentListJSON(assignments),
		"cap": map[string]any{
			"max_ai_employees": maxAI,
			"active_count":     activeCount,
			"override_allowed": s.demoAvatarCapOverrideAllowed(tenantID),
		},
	})
}

type assignAvatarBody struct {
	AvatarID string `json:"avatar_id"`
}

func (s *server) assignTenantAvatar(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	var body assignAvatarBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	avatarID := strings.TrimSpace(body.AvatarID)
	if avatarID == "" {
		writeError(w, http.StatusBadRequest, "avatar_id is required")
		return
	}
	ctx := r.Context()
	if err := s.checkAvatarAssignCap(ctx, tenantID, avatarID); err != nil {
		writeAvatarError(w, err)
		return
	}
	assignment, err := s.store.AssignAvatarToTenant(ctx, tenantID, avatarID)
	if err != nil {
		writeAvatarError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, tenantAssignmentJSON(*assignment))
}

func (s *server) revokeTenantAvatar(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	avatarID := strings.TrimSpace(r.PathValue("avatar_id"))
	if err := s.store.RevokeTenantAssignment(r.Context(), tenantID, avatarID); err != nil {
		writeAvatarError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "disabled"})
}

func (s *server) checkAvatarAssignCap(ctx context.Context, tenantID, avatarID string) error {
	assignments, err := s.store.ListTenantAvatarAssignments(ctx, tenantID)
	if err != nil {
		return err
	}
	for _, a := range assignments {
		if a.AvatarID == avatarID && a.Status == "active" {
			return nil
		}
	}
	// The configured demo tenant is seeded with the complete avatar catalog so
	// platform admins can promote and demote demo agents independently of the
	// tenant's commercial package cap. Runtime quota checks remain unchanged.
	if s.demoAvatarCapOverrideAllowed(tenantID) {
		return nil
	}
	count, err := s.store.CountActiveTenantAssignments(ctx, tenantID)
	if err != nil {
		return err
	}
	// SPRINT-013: prefer quota service (structured 429).
	if s.quota != nil {
		if err := s.quota.CheckAIEmployees(ctx, tenantID, count+1); err != nil {
			return err
		}
		return nil
	}
	maxAI := 0
	eff, err := s.entitlements.GetEffective(ctx, tenantID)
	if err == nil {
		maxAI = rulesInt(eff.Rules, "max_ai_employees")
	} else if !errors.Is(err, store.ErrEntitlementNotFound) {
		return err
	}
	if count >= maxAI {
		return store.ErrMaxAIEmployeesExceeded
	}
	return nil
}

func (s *server) demoAvatarCapOverrideAllowed(tenantID string) bool {
	demoTenantID := strings.TrimSpace(s.cfg.DemoTenantID)
	if demoTenantID == "" {
		demoTenantID = "demo"
	}
	return strings.TrimSpace(tenantID) == demoTenantID
}

func (s *server) buildAvatarFromBody(body avatarBody, id string, requireVoices bool) (*store.Avatar, []store.AvatarVoice, error) {
	rawSlug := strings.TrimSpace(body.Slug)
	slug := strings.ToLower(rawSlug)
	name := strings.TrimSpace(body.Name)
	if slug == "" || name == "" {
		return nil, nil, errAvatarValidation("slug and name are required")
	}
	if rawSlug != slug {
		return nil, nil, errAvatarValidation("slug must be lowercase")
	}
	status := strings.TrimSpace(body.Status)
	if status == "" {
		status = "draft"
	}
	if status != "draft" && status != "active" && status != "archived" {
		return nil, nil, errAvatarValidation("status must be draft, active, or archived")
	}
	voices, err := parseAvatarVoices(body.Voices, requireVoices)
	if err != nil {
		return nil, nil, err
	}
	if requireVoices && len(voices) == 0 {
		return nil, nil, errAvatarValidation("at least one voice is required")
	}
	if body.VoicesSet && len(voices) == 0 {
		return nil, nil, errAvatarValidation("voices cannot be empty")
	}
	if status == "active" && !hasActiveVoice(voices) {
		return nil, nil, errAvatarValidation("active avatars require at least one active voice")
	}
	imageURL := strings.TrimSpace(body.ImageURL)
	if imageURL == "" {
		imageURL = "/images/" + slug + ".jpg"
	}
	if id == "" {
		id = slug
	}
	flags := body.Flags
	if flags == nil {
		flags = map[string]any{}
	}
	return &store.Avatar{
		ID:       id,
		Slug:     slug,
		Name:     name,
		Role:     strings.TrimSpace(body.Role),
		Trait:    strings.TrimSpace(body.Trait),
		Color:    strings.TrimSpace(body.Color),
		ImageURL: imageURL,
		Greeting: strings.TrimSpace(body.Greeting),
		Status:   status,
		Flags:    flags,
	}, voices, nil
}

func parseAvatarVoices(items []avatarVoiceBody, required bool) ([]store.AvatarVoice, error) {
	if !required && len(items) == 0 {
		return nil, nil
	}
	out := make([]store.AvatarVoice, 0, len(items))
	for i, item := range items {
		providerID := strings.TrimSpace(item.VoiceProviderID)
		voiceID := strings.TrimSpace(item.VoiceID)
		voice := strings.TrimSpace(item.Voice)
		if providerID == "" || voiceID == "" || voice == "" {
			return nil, errAvatarValidation("voice_provider_id, voice_id, and voice are required")
		}
		priority := item.Priority
		if priority <= 0 {
			priority = i + 1
		}
		status := strings.TrimSpace(item.Status)
		if status == "" {
			status = "active"
		}
		if status != "active" && status != "disabled" {
			return nil, errAvatarValidation("voice status must be active or disabled")
		}
		out = append(out, store.AvatarVoice{
			ID:              strings.TrimSpace(item.ID),
			VoiceProviderID: providerID,
			VoiceID:         voiceID,
			Voice:           voice,
			Priority:        priority,
			Status:          status,
		})
	}
	return out, nil
}

func hasActiveVoice(voices []store.AvatarVoice) bool {
	for _, v := range voices {
		if v.Status == "active" {
			return true
		}
	}
	return false
}

func mergeAvatarUpdateBody(existing store.Avatar, body avatarUpdateBody) avatarBody {
	out := avatarBody{
		Slug:     existing.Slug,
		Name:     existing.Name,
		Role:     existing.Role,
		Trait:    existing.Trait,
		Color:    existing.Color,
		ImageURL: existing.ImageURL,
		Greeting: existing.Greeting,
		Status:   existing.Status,
		Flags:    existing.Flags,
	}
	if body.Slug != "" {
		out.Slug = body.Slug
	}
	if body.Name != "" {
		out.Name = body.Name
	}
	if body.Role != "" {
		out.Role = body.Role
	}
	if body.Trait != "" {
		out.Trait = body.Trait
	}
	if body.Color != "" {
		out.Color = body.Color
	}
	if body.ImageURL != "" {
		out.ImageURL = body.ImageURL
	}
	if body.Greeting != "" {
		out.Greeting = body.Greeting
	}
	if body.Status != "" {
		out.Status = body.Status
	}
	if body.Flags != nil {
		out.Flags = body.Flags
	}
	if body.Voices != nil {
		out.VoicesSet = true
		out.Voices = *body.Voices
	}
	return out
}

func avatarJSON(av store.Avatar) map[string]any {
	return map[string]any{
		"id":         av.ID,
		"slug":       av.Slug,
		"name":       av.Name,
		"role":       av.Role,
		"trait":      av.Trait,
		"color":      av.Color,
		"image_url":  av.ImageURL,
		"greeting":   av.Greeting,
		"status":     av.Status,
		"flags":      av.Flags,
		"voices":     avatarVoiceListJSON(av.Voices),
		"created_at": av.CreatedAt,
		"updated_at": av.UpdatedAt,
	}
}

func avatarListJSON(avatars []store.Avatar) []map[string]any {
	out := make([]map[string]any, 0, len(avatars))
	for _, av := range avatars {
		out = append(out, avatarJSON(av))
	}
	return out
}

func avatarVoiceJSON(v store.AvatarVoice) map[string]any {
	return map[string]any{
		"id":                v.ID,
		"voice_provider_id": v.VoiceProviderID,
		"voice_id":          v.VoiceID,
		"voice":             v.Voice,
		"priority":          v.Priority,
		"status":            v.Status,
	}
}

func avatarVoiceListJSON(voices []store.AvatarVoice) []map[string]any {
	out := make([]map[string]any, 0, len(voices))
	for _, v := range voices {
		out = append(out, avatarVoiceJSON(v))
	}
	return out
}

func tenantAssignmentJSON(a store.TenantAvatarAssignment) map[string]any {
	out := map[string]any{
		"avatar_id": a.AvatarID,
		"status":    a.Status,
	}
	if a.Avatar != nil {
		out["avatar"] = avatarSummaryJSON(*a.Avatar)
	}
	return out
}

func tenantAssignmentListJSON(assignments []store.TenantAvatarAssignment) []map[string]any {
	out := make([]map[string]any, 0, len(assignments))
	for _, a := range assignments {
		out = append(out, tenantAssignmentJSON(a))
	}
	return out
}

func avatarSummaryJSON(av store.Avatar) map[string]any {
	return map[string]any{
		"id":     av.ID,
		"name":   av.Name,
		"role":   av.Role,
		"status": av.Status,
	}
}

func rulesInt(rules map[string]any, key string) int {
	if rules == nil {
		return 0
	}
	v, ok := rules[key]
	if !ok {
		return 0
	}
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	default:
		return 0
	}
}

type avatarValidationError struct{ msg string }

func errAvatarValidation(msg string) error { return avatarValidationError{msg: msg} }

func (e avatarValidationError) Error() string { return e.msg }

func writeAvatarValidationError(w http.ResponseWriter, err error) {
	switch {
	case errors.As(err, &avatarValidationError{}):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writeAvatarError(w, err)
	}
}

func writeAvatarError(w http.ResponseWriter, err error) {
	var qe *quota.Error
	if errors.As(err, &qe) || errors.Is(err, quota.ErrLimitExceeded) {
		writeQuotaError(w, err)
		return
	}
	switch {
	case errors.Is(err, store.ErrAvatarNotFound),
		errors.Is(err, store.ErrTenantNotFound),
		errors.Is(err, store.ErrAssignmentNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, store.ErrAvatarHasAssignments),
		errors.Is(err, store.ErrMaxAIEmployeesExceeded):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, store.ErrVoiceProviderNotFound):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		var ve avatarValidationError
		if errors.As(err, &ve) {
			writeError(w, http.StatusBadRequest, ve.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
	}
}
