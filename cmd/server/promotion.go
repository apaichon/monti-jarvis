package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/store"
)

type promotionGrantBody struct {
	PackageID      string  `json:"package_id"`
	Reason         string  `json:"reason"`
	ValidUntil     *string `json:"valid_until"`
	AmountCents    *int    `json:"amount_cents"`
	IdempotencyKey string  `json:"idempotency_key"`
}

func (s *server) createPromotionGrant(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "tenant_id is required")
		return
	}
	var body promotionGrantBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	packageID := strings.TrimSpace(body.PackageID)
	reason := strings.TrimSpace(body.Reason)
	if packageID == "" || reason == "" {
		writeError(w, http.StatusBadRequest, "package_id and reason are required")
		return
	}

	var validUntil *time.Time
	if body.ValidUntil != nil && strings.TrimSpace(*body.ValidUntil) != "" {
		raw := strings.TrimSpace(*body.ValidUntil)
		t, err := time.Parse(time.RFC3339, raw)
		if err != nil {
			t, err = time.Parse("2006-01-02", raw)
			if err != nil {
				writeError(w, http.StatusBadRequest, "valid_until must be RFC3339 or YYYY-MM-DD")
				return
			}
			t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 0, time.UTC)
		}
		validUntil = &t
	}

	result, err := s.store.GrantPromotion(r.Context(), store.PromotionGrantInput{
		TenantID:       tenantID,
		PackageID:      packageID,
		Reason:         reason,
		ValidUntil:     validUntil,
		AmountCents:    body.AmountCents,
		IdempotencyKey: strings.TrimSpace(body.IdempotencyKey),
	})
	if err != nil {
		writePromotionError(w, err)
		return
	}
	if s.entitlements != nil {
		s.entitlements.Invalidate(r.Context(), tenantID)
	}
	status := http.StatusCreated
	if result.Replayed {
		status = http.StatusOK
	}
	writeJSON(w, status, promotionGrantJSON(result))
}

func (s *server) listPromotionGrants(w http.ResponseWriter, r *http.Request) {
	tenantID := strings.TrimSpace(r.PathValue("tenant_id"))
	if tenantID == "" {
		writeError(w, http.StatusBadRequest, "tenant_id is required")
		return
	}
	limit := 50
	if q := strings.TrimSpace(r.URL.Query().Get("limit")); q != "" {
		if n, err := strconv.Atoi(q); err == nil {
			limit = n
		}
	}
	grants, err := s.store.ListPromotionGrants(r.Context(), tenantID, limit)
	if err != nil {
		writePromotionError(w, err)
		return
	}
	items := make([]map[string]any, 0, len(grants))
	for _, g := range grants {
		item := map[string]any{
			"id":             g.ID,
			"tenant_id":      g.TenantID,
			"package_id":     g.PackageID,
			"order_id":       g.OrderID,
			"entitlement_id": g.EntitlementID,
			"tax_invoice_id": g.TaxInvoiceID,
			"reason":         g.Reason,
			"amount_cents":   g.AmountCents,
			"status":         g.Status,
			"created_at":     g.CreatedAt,
			"created_by":     g.CreatedBy,
		}
		if g.ValidUntil != nil {
			item["valid_until"] = g.ValidUntil
		}
		if doc, err := s.store.GetPaymentDocumentByID(r.Context(), g.TaxInvoiceID); err == nil {
			item["tax_invoice_number"] = doc.DocNumber
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{"grants": items})
}

func (s *server) getPromotionGrant(w http.ResponseWriter, r *http.Request) {
	grantID := strings.TrimSpace(r.PathValue("grant_id"))
	if grantID == "" {
		writeError(w, http.StatusBadRequest, "grant_id is required")
		return
	}
	grant, err := s.store.GetPromotionGrant(r.Context(), grantID)
	if err != nil {
		writePromotionError(w, err)
		return
	}
	ent, err := s.store.GetActiveEntitlement(r.Context(), grant.TenantID)
	if err != nil {
		writePromotionError(w, err)
		return
	}
	tax, err := s.store.GetPaymentDocumentByID(r.Context(), grant.TaxInvoiceID)
	if err != nil {
		writePromotionError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, promotionGrantJSON(&store.PromotionGrantResult{
		Grant:       *grant,
		Entitlement: ent,
		TaxInvoice:  tax,
	}))
}

func promotionGrantJSON(result *store.PromotionGrantResult) map[string]any {
	g := result.Grant
	out := map[string]any{
		"id":           g.ID,
		"tenant_id":    g.TenantID,
		"package_id":   g.PackageID,
		"order_id":     g.OrderID,
		"reason":       g.Reason,
		"amount_cents": g.AmountCents,
		"status":       g.Status,
		"created_at":   g.CreatedAt,
		"created_by":   g.CreatedBy,
		"replayed":     result.Replayed,
	}
	if g.IdempotencyKey != "" {
		out["idempotency_key"] = g.IdempotencyKey
	}
	if g.ValidUntil != nil {
		out["valid_until"] = g.ValidUntil
	}
	if result.Entitlement != nil {
		out["entitlement"] = entitlementFromStore(result.Entitlement)
	}
	if result.TaxInvoice != nil {
		d := result.TaxInvoice
		out["tax_invoice"] = map[string]any{
			"id":           d.ID,
			"doc_type":     d.DocType,
			"doc_number":   d.DocNumber,
			"status":       d.Status,
			"amount_cents": d.AmountCents,
			"currency":     d.Currency,
			"issued_at":    d.IssuedAt,
		}
		out["currency"] = d.Currency
	}
	return out
}

func entitlementFromStore(ent *store.TenantEntitlement) map[string]any {
	if ent == nil {
		return nil
	}
	pkg := map[string]any{}
	if ent.Package != nil {
		pkg = map[string]any{
			"id":   ent.Package.ID,
			"slug": ent.Package.Slug,
			"name": ent.Package.Name,
		}
	}
	rules := ent.RulesSnapshot
	if rules == nil {
		rules = map[string]any{}
	}
	out := map[string]any{
		"tenant_id":       ent.TenantID,
		"package":         pkg,
		"status":          ent.Status,
		"rules_schema_id": ent.RulesSchemaID,
		"rules":           rules,
		"valid_from":      ent.ValidFrom,
	}
	if ent.ValidUntil != nil {
		out["valid_until"] = ent.ValidUntil
	}
	return out
}

func writePromotionError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, store.ErrPromotionReasonRequired):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, store.ErrPackageNotActive):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, store.ErrPackageNotFound),
		errors.Is(err, store.ErrTenantNotFound),
		errors.Is(err, store.ErrPromotionGrantNotFound),
		errors.Is(err, store.ErrEntitlementNotFound),
		errors.Is(err, store.ErrPaymentOrderNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, store.ErrIdempotencyConflict):
		writeError(w, http.StatusConflict, err.Error())
	default:
		if strings.Contains(err.Error(), "amount_cents") {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
	}
}
