package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/store"
)

// GET /api/tenant/tax-profile
func (s *server) getTenantTaxProfile(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tenantID := strings.TrimSpace(ac.TenantID)
	if tenantID == "" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	p, err := s.store.GetTenantTaxProfile(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, taxProfileJSON(p))
}

type taxProfileBody struct {
	CompanyName string `json:"company_name"`
	TaxID       string `json:"tax_id"`
	Branch      string `json:"branch"`
	Address     string `json:"address"`
	// When true, reissue active tax invoices with new buyer fields.
	RefreshInvoices bool `json:"refresh_invoices"`
}

// PUT /api/tenant/tax-profile
func (s *server) putTenantTaxProfile(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tenantID := strings.TrimSpace(ac.TenantID)
	if tenantID == "" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	var body taxProfileBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	p, err := s.store.UpsertTenantTaxProfile(r.Context(), tenantID, body.CompanyName, body.TaxID, body.Branch, body.Address)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	refreshed := 0
	if body.RefreshInvoices {
		refreshed, err = s.store.RefreshTenantTaxInvoices(r.Context(), tenantID)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
	}
	out := taxProfileJSON(p)
	out["invoices_refreshed"] = refreshed
	writeJSON(w, http.StatusOK, out)
}

// GET /api/tenant/billing/documents
func (s *server) listTenantBillingDocuments(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tenantID := strings.TrimSpace(ac.TenantID)
	if tenantID == "" {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	docs, err := s.store.ListTenantPaymentDocuments(r.Context(), tenantID)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	out := make([]map[string]any, 0, len(docs))
	for _, d := range docs {
		out = append(out, paymentDocumentJSON(d))
	}
	writeJSON(w, http.StatusOK, map[string]any{"documents": out})
}

// GET /api/tenant/billing/documents/{id}
func (s *server) getTenantBillingDocument(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	tenantID := strings.TrimSpace(ac.TenantID)
	docID := strings.TrimSpace(r.PathValue("id"))
	doc, err := s.store.GetPaymentDocumentByID(r.Context(), docID)
	if err != nil {
		if errors.Is(err, store.ErrPaymentOrderNotFound) {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if doc.TenantID != tenantID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if strings.EqualFold(r.URL.Query().Get("format"), "html") {
		order, _ := s.store.GetPaymentOrderByID(r.Context(), doc.OrderID)
		if order == nil {
			order = &store.PaymentOrder{OrderNo: doc.OrderID, Status: doc.Status}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(renderPaymentDocumentHTML(*doc, *order)))
		return
	}
	writeJSON(w, http.StatusOK, paymentDocumentJSON(*doc))
}

func taxProfileJSON(p store.TenantTaxProfile) map[string]any {
	return map[string]any{
		"tenant_id":    p.TenantID,
		"company_name": p.CompanyName,
		"tax_id":       p.TaxID,
		"branch":       p.Branch,
		"address":      p.Address,
	}
}
