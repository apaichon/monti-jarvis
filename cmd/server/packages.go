package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/packages"
	"github.com/libra/monti-jarvis/internal/store"
)

func (s *server) listRuleSchemas(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	schemas, err := s.store.ListRuleSchemas(r.Context(), status)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	type schemaOut struct {
		ID      string          `json:"id"`
		Version int             `json:"version"`
		Name    string          `json:"name"`
		Status  string          `json:"status"`
		Fields  json.RawMessage `json:"fields"`
	}
	out := make([]schemaOut, 0, len(schemas))
	for _, rs := range schemas {
		out = append(out, schemaOut{
			ID: rs.ID, Version: rs.Version, Name: rs.Name, Status: rs.Status, Fields: rs.Fields,
		})
	}
	writeJSON(w, http.StatusOK, map[string]any{"schemas": out})
}

func (s *server) listPackages(w http.ResponseWriter, r *http.Request) {
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	pkgs, err := s.store.ListPackages(r.Context(), status)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"packages": packageListJSON(pkgs)})
}

func (s *server) getPackage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	pkg, err := s.store.GetPackage(r.Context(), id)
	if err != nil {
		writePackageError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, packageJSON(*pkg))
}

type packageBody struct {
	Slug          string         `json:"slug"`
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	Status        string         `json:"status"`
	PriceCents    int            `json:"price_cents"`
	Currency      string         `json:"currency"`
	BillingPeriod string         `json:"billing_period"`
	RulesSchemaID string         `json:"rules_schema_id"`
	Rules         map[string]any `json:"rules"`
}

func (s *server) createPackage(w http.ResponseWriter, r *http.Request) {
	var body packageBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	pkg, err := s.buildPackageFromBody(r, body, "")
	if err != nil {
		writePackageValidationError(w, err)
		return
	}
	created, err := s.store.CreatePackage(r.Context(), *pkg)
	if err != nil {
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, "slug already exists")
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, packageJSON(*created))
}

func (s *server) updatePackage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	existing, err := s.store.GetPackage(r.Context(), id)
	if err != nil {
		writePackageError(w, err)
		return
	}
	var body packageBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	merged := mergePackageBody(*existing, body)
	pkg, err := s.buildPackageFromBody(r, merged, id)
	if err != nil {
		writePackageValidationError(w, err)
		return
	}
	pkg.ID = id
	updated, err := s.store.UpdatePackage(r.Context(), *pkg)
	if err != nil {
		if isUniqueViolation(err) {
			writeError(w, http.StatusConflict, "slug already exists")
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, packageJSON(*updated))
}

func (s *server) archivePackage(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if err := s.store.ArchivePackage(r.Context(), id); err != nil {
		writePackageError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "archived"})
}

func (s *server) getTenantEntitlement(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	eff, err := s.entitlements.GetEffective(r.Context(), tenantID)
	if err != nil {
		writeEntitlementError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, eff)
}

type assignEntitlementBody struct {
	PackageID string `json:"package_id"`
}

func (s *server) assignTenantEntitlement(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	var body assignEntitlementBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	packageID := strings.TrimSpace(body.PackageID)
	if packageID == "" {
		writeError(w, http.StatusBadRequest, "package_id is required")
		return
	}
	_, err := s.store.AssignEntitlement(r.Context(), tenantID, packageID)
	if err != nil {
		writeEntitlementError(w, err)
		return
	}
	s.entitlements.Invalidate(r.Context(), tenantID)
	eff, err := s.entitlements.GetEffective(r.Context(), tenantID)
	if err != nil {
		writeEntitlementError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, eff)
}

func (s *server) revokeTenantEntitlement(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	if err := s.store.RevokeEntitlement(r.Context(), tenantID); err != nil {
		writeEntitlementError(w, err)
		return
	}
	s.entitlements.Invalidate(r.Context(), tenantID)
	writeJSON(w, http.StatusOK, map[string]string{"status": "revoked"})
}

func (s *server) entitlementMe(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tenantID := strings.TrimSpace(ac.TenantID)
	if tenantID == "" {
		tenantID = s.cfg.DemoTenantID
	}
	eff, err := s.entitlements.GetEffective(r.Context(), tenantID)
	if err != nil {
		writeEntitlementError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, eff)
}

func (s *server) buildPackageFromBody(r *http.Request, body packageBody, id string) (*store.Package, error) {
	slug := strings.TrimSpace(strings.ToLower(body.Slug))
	name := strings.TrimSpace(body.Name)
	if slug == "" || name == "" {
		return nil, errPackageValidation("slug and name are required")
	}
	schemaID := strings.TrimSpace(body.RulesSchemaID)
	if schemaID == "" {
		return nil, errPackageValidation("rules_schema_id is required")
	}
	if body.Rules == nil {
		return nil, errPackageValidation("rules is required")
	}
	schema, err := s.store.GetRuleSchema(r.Context(), schemaID)
	if err != nil {
		return nil, err
	}
	if schema.Status != "active" {
		return nil, errPackageValidation("rules schema is not active")
	}
	if err := packages.ValidateRules(schema.Fields, body.Rules); err != nil {
		return nil, err
	}
	status := strings.TrimSpace(body.Status)
	if status == "" {
		status = "draft"
	}
	currency := strings.TrimSpace(body.Currency)
	if currency == "" {
		currency = "USD"
	}
	billing := strings.TrimSpace(body.BillingPeriod)
	if billing == "" {
		billing = "monthly"
	}
	if id == "" {
		id = "pkg-" + slug
	}
	return &store.Package{
		ID:            id,
		Slug:          slug,
		Name:          name,
		Description:   strings.TrimSpace(body.Description),
		Status:        status,
		PriceCents:    body.PriceCents,
		Currency:      currency,
		BillingPeriod: billing,
		RulesSchemaID: schemaID,
		Rules:         body.Rules,
	}, nil
}

func mergePackageBody(existing store.Package, body packageBody) packageBody {
	out := packageBody{
		Slug:          existing.Slug,
		Name:          existing.Name,
		Description:   existing.Description,
		Status:        existing.Status,
		PriceCents:    existing.PriceCents,
		Currency:      existing.Currency,
		BillingPeriod: existing.BillingPeriod,
		RulesSchemaID: existing.RulesSchemaID,
		Rules:         existing.Rules,
	}
	if body.Slug != "" {
		out.Slug = body.Slug
	}
	if body.Name != "" {
		out.Name = body.Name
	}
	if body.Description != "" {
		out.Description = body.Description
	}
	if body.Status != "" {
		out.Status = body.Status
	}
	if body.Currency != "" {
		out.Currency = body.Currency
	}
	if body.BillingPeriod != "" {
		out.BillingPeriod = body.BillingPeriod
	}
	out.PriceCents = body.PriceCents
	if body.RulesSchemaID != "" {
		out.RulesSchemaID = body.RulesSchemaID
	}
	if body.Rules != nil {
		out.Rules = body.Rules
	}
	return out
}

func packageJSON(p store.Package) map[string]any {
	return map[string]any{
		"id":              p.ID,
		"slug":            p.Slug,
		"name":            p.Name,
		"description":     p.Description,
		"status":          p.Status,
		"price_cents":     p.PriceCents,
		"currency":        p.Currency,
		"billing_period":  p.BillingPeriod,
		"rules_schema_id": p.RulesSchemaID,
		"rules":           p.Rules,
		"created_at":      p.CreatedAt,
		"updated_at":      p.UpdatedAt,
	}
}

func packageListJSON(pkgs []store.Package) []map[string]any {
	out := make([]map[string]any, 0, len(pkgs))
	for _, p := range pkgs {
		out = append(out, packageJSON(p))
	}
	return out
}

type validationError struct{ msg string }

func errPackageValidation(msg string) error { return validationError{msg: msg} }

func (e validationError) Error() string { return e.msg }

func writePackageValidationError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, packages.ErrUnknownField),
		errors.Is(err, packages.ErrRequiredField),
		errors.Is(err, packages.ErrInvalidType),
		errors.Is(err, packages.ErrInvalidValue),
		errors.Is(err, packages.ErrInvalidSchema):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.As(err, &validationError{}):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		writePackageError(w, err)
	}
}

func writePackageError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrPackageNotFound),
		errors.Is(err, store.ErrRuleSchemaNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, store.ErrPackageHasEntitlements):
		writeError(w, http.StatusConflict, err.Error())
	default:
		var ve validationError
		if errors.As(err, &ve) {
			writeError(w, http.StatusBadRequest, ve.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
	}
}

func writeEntitlementError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrEntitlementNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, store.ErrTenantNotFound),
		errors.Is(err, store.ErrPackageNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	default:
		writeError(w, http.StatusBadGateway, err.Error())
	}
}

func isUniqueViolation(err error) bool {
	return err != nil && strings.Contains(err.Error(), "unique")
}

