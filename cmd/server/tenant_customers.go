package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/customerimport"
	"github.com/libra/monti-jarvis/internal/store"
)

type customerBody struct {
	Email       *string         `json:"email"`
	Phone       *string         `json:"phone"`
	DisplayName *string         `json:"display_name"`
	Locale      *string         `json:"locale"`
	TierID      *string         `json:"tier_id"`
	GroupIDs    *[]string       `json:"group_ids"`
	Source      *string         `json:"source"`
	ExternalID  *string         `json:"external_id"`
	Status      *string         `json:"status"`
	Metadata    *map[string]any `json:"metadata"`
}

type domainRuleBody struct {
	Domain         *string `json:"domain"`
	Policy         *string `json:"policy"`
	DefaultTierID  *string `json:"default_tier_id"`
	DefaultGroupID *string `json:"default_group_id"`
	Active         *bool   `json:"active"`
}

func writeCustomerError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrCustomerNotFound), errors.Is(err, store.ErrImportNotFound),
		errors.Is(err, store.ErrDomainRuleNotFound), errors.Is(err, store.ErrTierNotFound), errors.Is(err, store.ErrGroupNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": "resource not found", "code": "not_found"})
	case errors.Is(err, store.ErrCustomerConflict):
		writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error(), "code": "customer_conflict"})
	case errors.Is(err, store.ErrDomainRuleTaken):
		writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error(), "code": "domain_rule_exists"})
	case errors.Is(err, store.ErrInvalidLocale):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "locale must be en or th", "code": "validation_error"})
	default:
		if err != nil && (strings.Contains(err.Error(), "invalid") || strings.Contains(err.Error(), "required") || strings.Contains(err.Error(), "must be")) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "code": "validation_error"})
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
	}
}

func customerInputFromBody(body customerBody, current *store.Customer) store.CustomerInput {
	in := store.CustomerInput{Source: "manual", Status: "active", Metadata: map[string]any{}, GroupIDs: []string{}}
	if current != nil {
		in = store.CustomerInput{
			Email: current.Email, Phone: current.Phone, DisplayName: current.DisplayName, Locale: current.Locale,
			TierID: current.TierID, GroupIDs: append([]string(nil), current.GroupIDs...), Source: current.Source,
			ExternalID: current.ExternalID, Status: current.Status, Metadata: current.Metadata,
		}
	}
	if body.Email != nil {
		in.Email = *body.Email
	}
	if body.Phone != nil {
		in.Phone = *body.Phone
	}
	if body.DisplayName != nil {
		in.DisplayName = *body.DisplayName
	}
	if body.Locale != nil {
		in.Locale = *body.Locale
	}
	if body.TierID != nil {
		in.TierID = *body.TierID
	}
	if body.GroupIDs != nil {
		in.GroupIDs = *body.GroupIDs
	}
	if body.Source != nil {
		in.Source = *body.Source
	}
	if body.ExternalID != nil {
		in.ExternalID = *body.ExternalID
	}
	if body.Status != nil {
		in.Status = *body.Status
	}
	if body.Metadata != nil {
		in.Metadata = *body.Metadata
	}
	return in
}

func (s *server) publishCustomerUpsert(tenantID string, result *store.CustomerUpsertResult) {
	if s.bus == nil || result == nil || result.Customer == nil {
		return
	}
	event := map[string]any{
		"type": "customer.upserted", "tenant_id": tenantID, "customer_id": result.Customer.ID,
		"source": result.Customer.Source, "external_id": result.Customer.ExternalID,
		"outcome": result.Outcome, "occurred_at": time.Now().UTC(),
	}
	if err := s.bus.PublishJSON("monti.customer.upserted", event); err != nil {
		log.Printf("customer event warning: %v", err)
	}
}

func (s *server) listTenantCustomers(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	rows, err := s.store.ListCustomers(r.Context(), tenantID, store.CustomerListFilter{
		Query: r.URL.Query().Get("q"), Status: r.URL.Query().Get("status"), TierID: r.URL.Query().Get("tier_id"), Limit: limit,
	})
	if err != nil {
		writeCustomerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"customers": rows, "next_cursor": ""})
}

func (s *server) createTenantCustomer(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body customerBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	result, err := s.store.UpsertCustomer(r.Context(), tenantID, customerInputFromBody(body, nil))
	if err != nil {
		writeCustomerError(w, err)
		return
	}
	s.publishCustomerUpsert(tenantID, result)
	status := http.StatusCreated
	if result.Outcome == "updated" {
		status = http.StatusOK
	}
	writeJSON(w, status, map[string]any{"customer": result.Customer, "outcome": result.Outcome})
}

func (s *server) getTenantCustomer(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	row, err := s.store.GetCustomer(r.Context(), tenantID, strings.TrimSpace(r.PathValue("id")))
	if err != nil {
		writeCustomerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (s *server) putTenantCustomer(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	current, err := s.store.GetCustomer(r.Context(), tenantID, id)
	if err != nil {
		writeCustomerError(w, err)
		return
	}
	var body customerBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	row, err := s.store.UpdateCustomer(r.Context(), tenantID, id, customerInputFromBody(body, current))
	if err != nil {
		writeCustomerError(w, err)
		return
	}
	result := &store.CustomerUpsertResult{Customer: row, Outcome: "updated"}
	s.publishCustomerUpsert(tenantID, result)
	writeJSON(w, http.StatusOK, row)
}

func (s *server) deleteTenantCustomer(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	row, err := s.store.DeactivateCustomer(r.Context(), tenantID, strings.TrimSpace(r.PathValue("id")))
	if err != nil {
		writeCustomerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"id": row.ID, "status": row.Status})
}

func importErrorMaps(errorsIn []customerimport.RowError) []map[string]any {
	out := make([]map[string]any, 0, len(errorsIn))
	for _, item := range errorsIn {
		if len(out) == 100 {
			break
		}
		out = append(out, map[string]any{"row": item.Row, "field": item.Field, "code": item.Code, "message": item.Message})
	}
	return out
}

func rejectedRows(errorsIn []customerimport.RowError) int {
	seen := map[int]bool{}
	for _, item := range errorsIn {
		seen[item.Row] = true
	}
	return len(seen)
}

func (s *server) importTenantCustomers(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	maxBytes := s.cfg.CustomerImportMaxBytes
	if maxBytes <= 0 {
		maxBytes = 2 * 1024 * 1024
	}
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes+64*1024)
	if err := r.ParseMultipartForm(maxBytes); err != nil {
		writeJSON(w, http.StatusRequestEntityTooLarge, map[string]any{"error": "import too large", "code": "import_too_large"})
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "CSV file is required", "code": "import_invalid"})
		return
	}
	defer file.Close()
	parsed, err := customerimport.Parse(file, s.cfg.CustomerImportMaxRows)
	if err != nil {
		if strings.Contains(err.Error(), "exceeds maximum") {
			writeJSON(w, http.StatusRequestEntityTooLarge, map[string]any{"error": err.Error(), "code": "import_too_large"})
			return
		}
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error(), "code": "import_invalid"})
		return
	}
	dryRun := strings.EqualFold(r.FormValue("dry_run"), "true") || r.FormValue("dry_run") == "1"
	source := r.FormValue("source")
	if source == "" {
		source = "csv"
	}
	tiers, err := s.store.ListCustomerTiers(r.Context(), tenantID)
	if err != nil {
		writeCustomerError(w, err)
		return
	}
	groups, err := s.store.ListCustomerGroups(r.Context(), tenantID)
	if err != nil {
		writeCustomerError(w, err)
		return
	}
	tierBySlug := map[string]string{}
	for _, tier := range tiers {
		if tier.Active {
			tierBySlug[tier.Slug] = tier.ID
		}
	}
	groupBySlug := map[string]string{}
	for _, group := range groups {
		groupBySlug[group.Slug] = group.ID
	}
	valid := make([]store.CustomerInput, 0, len(parsed.Rows))
	for _, row := range parsed.Rows {
		in := store.CustomerInput{DisplayName: row.DisplayName, Email: row.Email, Phone: row.Phone, Locale: row.Locale, Source: source, ExternalID: row.ExternalID, Status: "active", Metadata: map[string]any{}}
		if row.Source != "" {
			in.Source = row.Source
		}
		invalid := false
		if row.TierSlug != "" {
			id, found := tierBySlug[row.TierSlug]
			if !found {
				parsed.Errors = append(parsed.Errors, customerimport.RowError{Row: row.Number, Field: "tier_slug", Code: "tier_not_found", Message: "Tier not found"})
				invalid = true
			} else {
				in.TierID = id
			}
		}
		for _, slug := range row.GroupSlugs {
			id, found := groupBySlug[slug]
			if !found {
				parsed.Errors = append(parsed.Errors, customerimport.RowError{Row: row.Number, Field: "group_slugs", Code: "group_not_found", Message: fmt.Sprintf("Group %s not found", slug)})
				invalid = true
				continue
			}
			in.GroupIDs = append(in.GroupIDs, id)
		}
		if !invalid {
			prepared, validateErr := s.store.ValidateCustomerInput(r.Context(), tenantID, in)
			if validateErr != nil {
				parsed.Errors = append(parsed.Errors, customerimport.RowError{Row: row.Number, Field: "row", Code: "validation_error", Message: validateErr.Error()})
				continue
			}
			valid = append(valid, prepared)
		}
	}
	created, updated := 0, 0
	rejected := rejectedRows(parsed.Errors)
	jobInput := store.CustomerImportJob{
		Filename: header.Filename, Mode: "commit", Status: "completed", TotalRows: parsed.Total,
		RejectedRows: rejected, Errors: importErrorMaps(parsed.Errors),
	}
	var job *store.CustomerImportJob
	if !dryRun {
		results, storedJob, importErr := s.store.CommitCustomerImport(r.Context(), tenantID, valid, jobInput)
		if importErr != nil {
			writeCustomerError(w, importErr)
			return
		}
		for i := range results {
			result := &results[i]
			if result.Outcome == "created" {
				created++
			} else {
				updated++
			}
			s.publishCustomerUpsert(tenantID, result)
		}
		job = storedJob
	} else {
		jobInput.Mode = "dry_run"
		jobInput.Status = "validated"
		var createErr error
		job, createErr = s.store.CreateCustomerImportJob(r.Context(), tenantID, jobInput)
		if createErr != nil {
			writeCustomerError(w, createErr)
			return
		}
	}
	job.CreatedRows = created
	job.UpdatedRows = updated
	job.AcceptedRows = parsed.Total - rejected
	status := http.StatusCreated
	if dryRun {
		status = http.StatusOK
	}
	writeJSON(w, status, job)
}

func (s *server) getTenantCustomerImport(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	job, err := s.store.GetCustomerImportJob(r.Context(), tenantID, strings.TrimSpace(r.PathValue("id")))
	if err != nil {
		writeCustomerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, job)
}

func ruleInput(body domainRuleBody, current *store.CustomerDomainRule) store.CustomerDomainRuleInput {
	in := store.CustomerDomainRuleInput{}
	if current != nil {
		in.Domain, in.Policy, in.DefaultTierID, in.DefaultGroupID = current.Domain, current.Policy, current.DefaultTierID, current.DefaultGroupID
		active := current.Active
		in.Active = &active
	}
	if body.Domain != nil {
		in.Domain = *body.Domain
	}
	if body.Policy != nil {
		in.Policy = *body.Policy
	}
	if body.DefaultTierID != nil {
		in.DefaultTierID = *body.DefaultTierID
	}
	if body.DefaultGroupID != nil {
		in.DefaultGroupID = *body.DefaultGroupID
	}
	if body.Active != nil {
		in.Active = body.Active
	}
	return in
}

func (s *server) listTenantCustomerDomainRules(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	rows, err := s.store.ListCustomerDomainRules(r.Context(), tenantID)
	if err != nil {
		writeCustomerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"rules": rows})
}

func (s *server) createTenantCustomerDomainRule(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	var body domainRuleBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	row, err := s.store.CreateCustomerDomainRule(r.Context(), tenantID, ruleInput(body, nil))
	if err != nil {
		writeCustomerError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, row)
}

func (s *server) putTenantCustomerDomainRule(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	current, err := s.store.GetCustomerDomainRule(r.Context(), tenantID, id)
	if err != nil {
		writeCustomerError(w, err)
		return
	}
	var body domainRuleBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	row, err := s.store.UpdateCustomerDomainRule(r.Context(), tenantID, id, ruleInput(body, current))
	if err != nil {
		writeCustomerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, row)
}

func (s *server) deleteTenantCustomerDomainRule(w http.ResponseWriter, r *http.Request) {
	tenantID, ok := s.tenantIDFromAuth(r)
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if err := s.store.DeleteCustomerDomainRule(r.Context(), tenantID, id); err != nil {
		writeCustomerError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "id": id})
}
