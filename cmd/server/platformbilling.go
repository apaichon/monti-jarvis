package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/libra/monti-jarvis/internal/store"
)

// GET /api/platform/billing/orders?tenant_id=&status=&limit=&offset=
func (s *server) listPlatformBillingOrders(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	items, err := s.store.ListPaymentOrders(r.Context(), store.PaymentOrderListFilter{
		TenantID: r.URL.Query().Get("tenant_id"),
		Status:   r.URL.Query().Get("status"),
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	out := make([]map[string]any, 0, len(items))
	for _, it := range items {
		row := paymentOrderJSON(it.PaymentOrder, nil)
		row["package_name"] = it.PackageName
		row["tenant_name"] = it.TenantName
		out = append(out, row)
	}
	writeJSON(w, http.StatusOK, map[string]any{"orders": out})
}

// GET /api/platform/billing/orders/{id}
func (s *server) getPlatformBillingOrder(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	order, err := s.store.GetPaymentOrderByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrPaymentOrderNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	docs, _ := s.store.ListPaymentDocumentsByOrder(r.Context(), order.ID)
	writeJSON(w, http.StatusOK, paymentOrderJSON(*order, docs))
}

// GET /api/platform/billing/documents?tenant_id=&doc_type=&status=
func (s *server) listPlatformBillingDocuments(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	docs, err := s.store.ListPaymentDocuments(r.Context(), store.PaymentDocumentListFilter{
		TenantID: r.URL.Query().Get("tenant_id"),
		DocType:  r.URL.Query().Get("doc_type"),
		Status:   r.URL.Query().Get("status"),
		Limit:    limit,
		Offset:   offset,
	})
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

// GET /api/platform/billing/documents/{id}
func (s *server) getPlatformBillingDocument(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	doc, err := s.store.GetPaymentDocumentByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, store.ErrPaymentOrderNotFound) {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if strings.EqualFold(r.URL.Query().Get("format"), "html") {
		order, _ := s.store.GetPaymentOrderByID(r.Context(), doc.OrderID)
		if order == nil {
			order = &store.PaymentOrder{OrderNo: doc.OrderID, Status: doc.Status, TransactionID: ""}
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(renderPaymentDocumentHTML(*doc, *order)))
		return
	}
	writeJSON(w, http.StatusOK, paymentDocumentJSON(*doc))
}

type voidDocBody struct {
	Reason string `json:"reason"`
}

// POST /api/platform/billing/documents/{id}/void
func (s *server) voidPlatformBillingDocument(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	var body voidDocBody
	_ = json.NewDecoder(r.Body).Decode(&body)
	doc, err := s.store.VoidPaymentDocument(r.Context(), id, body.Reason)
	if err != nil {
		if errors.Is(err, store.ErrPaymentOrderNotFound) {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, paymentDocumentJSON(*doc))
}

// POST /api/platform/billing/documents/{id}/reissue
func (s *server) reissuePlatformBillingDocument(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	var body voidDocBody
	_ = json.NewDecoder(r.Body).Decode(&body)
	reason := strings.TrimSpace(body.Reason)
	if reason == "" {
		reason = "reissued by platform admin"
	}
	doc, err := s.store.ReissuePaymentDocument(r.Context(), id, reason)
	if err != nil {
		if errors.Is(err, store.ErrPaymentOrderNotFound) {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, paymentDocumentJSON(*doc))
}

// GET /api/platform/billing/seller-branding
func (s *server) getSellerBranding(w http.ResponseWriter, r *http.Request) {
	b, err := s.store.GetSellerBranding(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sellerBrandingJSON(b))
}

type sellerBrandingBody struct {
	Name    string `json:"name"`
	Address string `json:"address"`
	TaxID   string `json:"tax_id"`
	Branch  string `json:"branch"`
}

// PUT /api/platform/billing/seller-branding
func (s *server) putSellerBranding(w http.ResponseWriter, r *http.Request) {
	var body sellerBrandingBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	b, err := s.store.UpsertSellerBranding(r.Context(), body.Name, body.Address, body.TaxID, body.Branch)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, sellerBrandingJSON(b))
}

func sellerBrandingJSON(b store.PlatformSellerBranding) map[string]any {
	return map[string]any{
		"id":      b.ID,
		"name":    b.Name,
		"address": b.Address,
		"tax_id":  b.TaxID,
		"branch":  b.Branch,
	}
}
