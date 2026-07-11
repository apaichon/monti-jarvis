package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/payment"
	"github.com/libra/monti-jarvis/internal/payment/chillpay"
	"github.com/libra/monti-jarvis/internal/store"
)

type checkoutBody struct {
	PackageID     string `json:"package_id"`
	PaymentMethod string `json:"payment_method"`
}

func (s *server) listTenantPackages(w http.ResponseWriter, r *http.Request) {
	pkgs, err := s.store.ListPackages(r.Context(), "active")
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	ac, _ := auth.FromContext(r.Context())
	tenantID := strings.TrimSpace(ac.TenantID)

	var current map[string]any
	ent, err := s.store.GetActiveEntitlement(r.Context(), tenantID)
	if err == nil && ent.Package != nil {
		current = map[string]any{
			"package_id":   ent.PackageID,
			"package_name": ent.Package.Name,
			"status":       ent.Status,
		}
	} else if errors.Is(err, store.ErrEntitlementNotFound) {
		current = nil
	} else if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	out := make([]map[string]any, 0, len(pkgs))
	for _, p := range pkgs {
		out = append(out, tenantPackageJSON(p))
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"packages":            out,
		"current_entitlement": current,
		"payment_methods": []map[string]string{
			{"id": payment.MethodCreditCard, "label": "Credit Card", "channel_code": payment.ChannelCreditCard},
			{"id": payment.MethodQRPromptPay, "label": "QR PromptPay", "channel_code": payment.ChannelQRPromptPay},
		},
	})
}

func (s *server) tenantCheckout(w http.ResponseWriter, r *http.Request) {
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

	var body checkoutBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	packageID := strings.TrimSpace(body.PackageID)
	if packageID == "" {
		writeError(w, http.StatusBadRequest, "package_id is required")
		return
	}
	method, err := payment.NormalizePaymentMethod(body.PaymentMethod)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	pkg, err := s.store.GetPackage(r.Context(), packageID)
	if err != nil {
		if errors.Is(err, store.ErrPackageNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if pkg.Status != "active" {
		writeError(w, http.StatusConflict, "package is not active")
		return
	}

	gwRow, err := s.store.GetPaymentGatewayConfig(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	gw := payment.NewGateway(s.cfg, s.store)
	resolved := gw.Resolve(gwRow)
	if strings.TrimSpace(resolved.Provider) == "" || resolved.Status != "active" {
		writeError(w, http.StatusServiceUnavailable, "payment gateway not configured")
		return
	}

	provider := strings.ToLower(strings.TrimSpace(resolved.Provider))
	currency := strings.TrimSpace(resolved.Currency)
	if currency == "" {
		currency = "764"
	}

	order, err := s.store.CreatePaymentOrder(r.Context(), store.CreatePaymentOrderInput{
		TenantID:      tenantID,
		PackageID:     packageID,
		AmountCents:   pkg.PriceCents,
		Currency:      currency,
		Provider:      provider,
		PaymentMethod: method,
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	// SPA page shown to the tenant after payment (order_id for local mock / client).
	spaReturnURL := tenantSPAReturnURL(resolved.ReturnURL, s.cfg.PublicBaseURL, order.ID, order.OrderNo, "", "")
	// What ChillPay gets: server bridge with order ref in the *path* (query strings are often dropped).
	// Handler fulfills paid orders + redirects to /tenant/billing/return?…
	chillPayReturnURL := chillpayBrowserReturnURL(resolved.ReturnURL, resolved.CallbackURL, s.cfg.PublicBaseURL, order.OrderNo)

	var paymentURL string
	var transactionID string

	switch provider {
	case payment.ProviderMock:
		base := strings.TrimRight(strings.TrimSpace(s.cfg.PublicBaseURL), "/")
		paymentURL = base + "/tenant/billing/mock-pay?order_id=" + url.QueryEscape(order.ID)
		if s.cfg.PaymentMockAutoFulfill {
			result, err := s.store.FulfillPaymentOrder(r.Context(), order.OrderNo, "mock_"+order.ID, "0")
			if err != nil {
				writeError(w, http.StatusBadGateway, err.Error())
				return
			}
			if result.EntitlementChanged {
				s.entitlements.Invalidate(r.Context(), tenantID)
			}
			order = &result.Order
			paymentURL = spaReturnURL
		}
	default:
		client := chillpay.NewClient(chillpay.Config{
			MerchantCode: resolved.MerchantCode,
			APIKey:       resolved.APIKey,
			MD5Key:       resolved.MD5Key,
			BaseURL:      resolved.BaseURL,
			RouteNo:      resolved.RouteNo,
			Currency:     resolved.Currency,
			CallbackURL:  resolved.CallbackURL,
			ReturnURL:    chillPayReturnURL,
		})
		amountBaht := float64(pkg.PriceCents) / 100.0
		clientIP := clientIPFromRequest(r)
		// ChillPay CustName must be a person name — not an email (error 2032).
		custName := ""
		if user, uerr := s.store.GetUserByID(r.Context(), ac.UserID); uerr == nil {
			custName = user.DisplayName
		}
		log.Printf("chillpay checkout tenant=%s order_no=%s return_url=%s callback_url=%s",
			tenantID, order.OrderNo, chillPayReturnURL, resolved.CallbackURL)
		paymentURL, transactionID, err = client.InitPayment(chillpay.RequestInfo{
			OrderNo:     order.OrderNo,
			CustomerID:  tenantID,
			Amount:      amountBaht,
			Description: pkg.Name,
			ChannelCode: payment.ChannelCodeForMethod(method),
			IPAddress:   clientIP,
			CustEmail:   ac.Email,
			CustName:    custName,
		})
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		if err := s.store.UpdatePaymentOrderInit(r.Context(), order.ID, transactionID, paymentURL); err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
	}

	if provider == payment.ProviderMock && !s.cfg.PaymentMockAutoFulfill {
		if err := s.store.UpdatePaymentOrderInit(r.Context(), order.ID, "", paymentURL); err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"order_id":           order.ID,
		"order_no":           order.OrderNo,
		"package_id":         packageID,
		"amount_cents":       order.AmountCents,
		"currency":           order.Currency,
		"status":             order.Status,
		"payment_url":        paymentURL,
		"provider":           provider,
		"payment_method":     method,
		"return_url":         spaReturnURL,
		"chillpay_return_url": chillPayReturnURL,
	})
}

// Paths used after ChillPay browser redirect (harvest-style bridge → SPA).
const (
	tenantBillingReturnPath   = "/tenant/billing/return"
	chillpayBrowserReturnPath = "/api/callbacks/chillpay/return"
)

// chillpayBrowserReturnURL is the ReturnUrl sent to ChillPay.
// Embed order_no in the path — ChillPay often strips query params and may omit OrderNo in POST.
func chillpayBrowserReturnURL(configuredReturn, configuredCallback, publicBase, orderNo string) string {
	base := ""
	if host := absoluteHostBase(configuredReturn); host != "" {
		base = host + chillpayBrowserReturnPath
	} else if host := absoluteHostBase(configuredCallback); host != "" {
		base = host + chillpayBrowserReturnPath
	} else {
		base = strings.TrimRight(fallbackPublicBase(publicBase), "/") + chillpayBrowserReturnPath
	}
	orderNo = strings.TrimSpace(orderNo)
	if orderNo == "" {
		return base
	}
	return base + "/" + url.PathEscape(orderNo)
}

// tenantSPAReturnURL is the tenant portal page after the server bridge redirect.
func tenantSPAReturnURL(configuredReturn, publicBase, orderID, orderNo, status, txnID string) string {
	base := ""
	if host := absoluteHostBase(configuredReturn); host != "" {
		base = host + tenantBillingReturnPath
	} else {
		base = strings.TrimRight(fallbackPublicBase(publicBase), "/") + tenantBillingReturnPath
	}
	q := url.Values{}
	if strings.TrimSpace(orderID) != "" {
		q.Set("order_id", strings.TrimSpace(orderID))
	}
	if strings.TrimSpace(orderNo) != "" {
		q.Set("order_no", strings.TrimSpace(orderNo))
	}
	if strings.TrimSpace(status) != "" {
		q.Set("status", strings.TrimSpace(status))
	}
	if strings.TrimSpace(txnID) != "" {
		q.Set("txn_id", strings.TrimSpace(txnID))
	}
	if encoded := q.Encode(); encoded != "" {
		return base + "?" + encoded
	}
	return base
}

// checkoutReturnURL keeps SPA return helper for mock / tests.
func checkoutReturnURL(configured, publicBase, orderID string) string {
	return tenantSPAReturnURL(configured, publicBase, orderID, "", "", "")
}

func absoluteHostBase(raw string) string {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

func fallbackPublicBase(publicBase string) string {
	publicBase = strings.TrimRight(strings.TrimSpace(publicBase), "/")
	if publicBase == "" {
		return "http://localhost:8091"
	}
	return publicBase
}

// chillpayBrowserReturn receives the browser after ChillPay payment (GET or POST).
// Fulfills the order when payment succeeded (DB paid + package entitlement/quota),
// then redirects to /tenant/billing/return with order params.
func (s *server) chillpayBrowserReturn(w http.ResponseWriter, r *http.Request) {
	_ = r.ParseForm()
	_ = r.ParseMultipartForm(1 << 20)

	// Log every field ChillPay sent (sandbox field names vary).
	if len(r.Form) > 0 {
		log.Printf("chillpay browser return form=%v", r.Form)
	}

	pathRef := strings.TrimSpace(r.PathValue("ref"))
	orderNo := firstNonEmpty(
		pathRef,
		r.FormValue("OrderNo"),
		r.FormValue("orderNo"),
		r.FormValue("orderno"),
		r.URL.Query().Get("OrderNo"),
		r.URL.Query().Get("orderNo"),
		r.FormValue("order_no"),
		r.URL.Query().Get("order_no"),
	)
	statusRaw := firstNonEmpty(
		r.FormValue("PaymentStatus"),
		r.FormValue("paymentStatus"),
		r.FormValue("Status"),
		r.FormValue("status"),
		r.URL.Query().Get("PaymentStatus"),
		r.URL.Query().Get("status"),
	)
	txnID := firstNonEmpty(
		r.FormValue("TransactionId"),
		r.FormValue("transactionId"),
		r.FormValue("TransCode"),
		r.FormValue("transCode"),
		r.URL.Query().Get("TransactionId"),
		r.URL.Query().Get("txn_id"),
	)

	payStatus := normalizeChillPayPaymentStatus(statusRaw)
	log.Printf("chillpay browser return method=%s path_ref=%s order_no=%s status_raw=%s status=%s txn_id=%s",
		r.Method, pathRef, orderNo, statusRaw, payStatus, txnID)

	orderID := ""
	if s.store != nil && orderNo != "" {
		// Path ref may be order_id (ord_…) or order_no (MJ…).
		if order, err := s.store.GetPaymentOrderByOrderNo(r.Context(), orderNo); err == nil {
			orderID = order.ID
			orderNo = order.OrderNo
		} else if order, err := s.store.GetPaymentOrderByID(r.Context(), orderNo); err == nil {
			orderID = order.ID
			orderNo = order.OrderNo
		}

		// Fulfill when ChillPay reports success (callback may not have reached us).
		if orderNo != "" && payStatus != "" {
			result, fulfillErr := s.store.FulfillPaymentOrder(r.Context(), orderNo, txnID, payStatus)
			if fulfillErr != nil {
				log.Printf("chillpay browser return fulfill error order_no=%s: %v", orderNo, fulfillErr)
			} else {
				orderID = result.Order.ID
				if result.EntitlementChanged && s.entitlements != nil {
					s.entitlements.Invalidate(r.Context(), result.Order.TenantID)
					log.Printf("chillpay browser return fulfilled order_no=%s tenant=%s package=%s paid",
						orderNo, result.Order.TenantID, result.Order.PackageID)
				}
			}
		}
	}

	// Prefer public host that ChillPay can reach (ngrok).
	var resolved payment.ResolvedConfig
	if s.store != nil {
		gwRow, _ := s.store.GetPaymentGatewayConfig(r.Context())
		gw := payment.NewGateway(s.cfg, s.store)
		resolved = gw.Resolve(gwRow)
	}
	dest := tenantSPAReturnURL(resolved.ReturnURL, s.cfg.PublicBaseURL, orderID, orderNo, payStatus, txnID)
	http.Redirect(w, r, dest, http.StatusFound)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

// normalizeChillPayPaymentStatus maps browser-return status strings to callback codes.
// Callback uses "0"=success, "1"=pending, "2"=failed. Browser return may send
// "complete", "success", "Paid", etc. (observed: status=complete with empty OrderNo).
func normalizeChillPayPaymentStatus(raw string) string {
	s := strings.TrimSpace(strings.ToLower(raw))
	switch s {
	case "0", "success", "successful", "complete", "completed", "paid", "ok", "approve", "approved":
		return "0"
	case "2", "fail", "failed", "error", "cancel", "cancelled", "canceled", "reject", "rejected":
		return "2"
	case "1", "pending", "wait", "waitauthorize", "processing":
		return "1"
	default:
		return strings.TrimSpace(raw)
	}
}

func clientIPFromRequest(r *http.Request) string {
	if xff := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); xff != "" {
		parts := strings.Split(xff, ",")
		if p := strings.TrimSpace(parts[0]); p != "" {
			return p
		}
	}
	if xri := strings.TrimSpace(r.Header.Get("X-Real-IP")); xri != "" {
		return xri
	}
	host := r.RemoteAddr
	if i := strings.LastIndex(host, ":"); i >= 0 {
		host = host[:i]
	}
	host = strings.Trim(host, "[]")
	if host == "" {
		return "127.0.0.1"
	}
	return host
}

func (s *server) getTenantOrder(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	orderID := strings.TrimSpace(r.PathValue("id"))
	order, err := s.store.GetPaymentOrderByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, store.ErrPaymentOrderNotFound) {
			// Allow lookup by order_no (ChillPay may return OrderNo on browser redirect).
			order, err = s.store.GetPaymentOrderByOrderNo(r.Context(), orderID)
		}
		if err != nil {
			if errors.Is(err, store.ErrPaymentOrderNotFound) {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
	}
	if strings.TrimSpace(ac.TenantID) != order.TenantID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}

	// Ensure documents exist for paid orders (re-issue if callback path skipped).
	docs := []store.PaymentDocument{}
	if order.Status == store.PaymentOrderStatusPaid {
		_ = s.store.IssuePaymentDocuments(r.Context(), order.ID)
		docs, _ = s.store.ListPaymentDocumentsByOrder(r.Context(), order.ID)
	}

	writeJSON(w, http.StatusOK, paymentOrderJSON(*order, docs))
}

func (s *server) getTenantOrderDocument(w http.ResponseWriter, r *http.Request) {
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	orderID := strings.TrimSpace(r.PathValue("id"))
	docType := strings.TrimSpace(r.PathValue("doc_type"))
	switch docType {
	case store.PaymentDocTypeReceipt, store.PaymentDocTypeTaxInvoice:
	default:
		writeError(w, http.StatusBadRequest, "doc_type must be receipt or tax_invoice")
		return
	}

	order, err := s.store.GetPaymentOrderByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, store.ErrPaymentOrderNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if strings.TrimSpace(ac.TenantID) != order.TenantID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if order.Status != store.PaymentOrderStatusPaid {
		writeError(w, http.StatusConflict, "documents available only for paid orders")
		return
	}
	if err := s.store.IssuePaymentDocuments(r.Context(), order.ID); err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	doc, err := s.store.GetPaymentDocument(r.Context(), order.ID, docType)
	if err != nil {
		if errors.Is(err, store.ErrPaymentOrderNotFound) {
			writeError(w, http.StatusNotFound, "document not found")
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	format := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("format")))
	if format == "html" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(renderPaymentDocumentHTML(*doc, *order)))
		return
	}
	writeJSON(w, http.StatusOK, paymentDocumentJSON(*doc))
}

func (s *server) mockPayOrder(w http.ResponseWriter, r *http.Request) {
	if s.cfg.AppEnv == "prod" {
		writeError(w, http.StatusNotFound, "not found")
		return
	}
	ac, ok := auth.FromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}
	orderID := strings.TrimSpace(r.PathValue("order_id"))
	order, err := s.store.GetPaymentOrderByID(r.Context(), orderID)
	if err != nil {
		if errors.Is(err, store.ErrPaymentOrderNotFound) {
			writeError(w, http.StatusNotFound, err.Error())
			return
		}
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if strings.TrimSpace(ac.TenantID) != order.TenantID {
		writeError(w, http.StatusForbidden, "forbidden")
		return
	}
	if order.Provider != payment.ProviderMock {
		writeError(w, http.StatusConflict, "order is not a mock payment")
		return
	}

	// Optional fail path for UI testing: POST body {"result":"failed"}
	resultStatus := "0"
	var mockBody struct {
		Result string `json:"result"`
	}
	_ = json.NewDecoder(r.Body).Decode(&mockBody)
	if strings.EqualFold(strings.TrimSpace(mockBody.Result), "failed") {
		resultStatus = "2"
	}

	result, err := s.store.FulfillPaymentOrder(r.Context(), order.OrderNo, "mock_"+order.ID, resultStatus)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if result.EntitlementChanged {
		s.entitlements.Invalidate(r.Context(), order.TenantID)
	}
	docs, _ := s.store.ListPaymentDocumentsByOrder(r.Context(), order.ID)
	writeJSON(w, http.StatusOK, paymentOrderJSON(result.Order, docs))
}

func tenantPackageJSON(p store.Package) map[string]any {
	summary := map[string]any{}
	for _, key := range []string{"max_ai_employees", "max_monthly_call_minutes", "max_km_documents", "max_concurrent_calls"} {
		if v, ok := p.Rules[key]; ok {
			summary[key] = v
		}
	}
	return map[string]any{
		"id":             p.ID,
		"slug":           p.Slug,
		"name":           p.Name,
		"description":    p.Description,
		"price_cents":    p.PriceCents,
		"currency":       p.Currency,
		"billing_period": p.BillingPeriod,
		"rules_summary":  summary,
	}
}

func paymentOrderJSON(o store.PaymentOrder, docs []store.PaymentDocument) map[string]any {
	out := map[string]any{
		"id":              o.ID,
		"order_no":        o.OrderNo,
		"package_id":      o.PackageID,
		"status":          o.Status,
		"amount_cents":    o.AmountCents,
		"currency":        o.Currency,
		"payment_method":  o.PaymentMethod,
		"provider":        o.Provider,
		"transaction_id":  o.TransactionID,
		"created_at":      o.CreatedAt,
		"documents":       paymentDocumentsJSON(docs),
	}
	if o.PaidAt != nil {
		out["paid_at"] = o.PaidAt.UTC()
	} else {
		out["paid_at"] = nil
	}
	return out
}

func paymentDocumentsJSON(docs []store.PaymentDocument) []map[string]any {
	out := make([]map[string]any, 0, len(docs))
	for _, d := range docs {
		out = append(out, paymentDocumentJSON(d))
	}
	return out
}

func paymentDocumentJSON(d store.PaymentDocument) map[string]any {
	out := map[string]any{
		"id":               d.ID,
		"order_id":         d.OrderID,
		"tenant_id":        d.TenantID,
		"doc_type":         d.DocType,
		"doc_number":       d.DocNumber,
		"status":           d.Status,
		"buyer_name":       d.BuyerName,
		"buyer_address":    d.BuyerAddress,
		"buyer_tax_id":     d.BuyerTaxID,
		"seller_name":      d.SellerName,
		"seller_address":   d.SellerAddress,
		"seller_tax_id":    d.SellerTaxID,
		"package_name":     d.PackageName,
		"amount_cents":     d.AmountCents,
		"currency":         d.Currency,
		"vat_rate_bps":     d.VATRateBps,
		"net_cents":        d.NetCents,
		"vat_cents":        d.VATCents,
		"payment_method":   d.PaymentMethod,
		"reissued_from_id": d.ReissuedFromID,
		"void_reason":      d.VoidReason,
		"issued_at":        d.IssuedAt.UTC(),
	}
	if d.VoidedAt != nil {
		out["voided_at"] = d.VoidedAt.UTC()
	} else {
		out["voided_at"] = nil
	}
	return out
}

func renderPaymentDocumentHTML(d store.PaymentDocument, o store.PaymentOrder) string {
	title := "Receipt"
	if d.DocType == store.PaymentDocTypeTaxInvoice {
		title = "Tax Invoice"
	}
	if d.Status == store.PaymentDocStatusVoided {
		title = title + " (VOIDED)"
	}
	amount := fmt.Sprintf("%.2f", float64(d.AmountCents)/100)
	net := fmt.Sprintf("%.2f", float64(d.NetCents)/100)
	vat := fmt.Sprintf("%.2f", float64(d.VATCents)/100)
	rate := fmt.Sprintf("%.0f", float64(d.VATRateBps)/100)
	method := payment.MethodLabel(d.PaymentMethod)
	voidNote := ""
	if d.Status == store.PaymentDocStatusVoided {
		voidNote = fmt.Sprintf(`<p style="color:#b91c1c;font-weight:600">VOIDED%s</p>`,
			func() string {
				if d.VoidReason != "" {
					return ": " + escHTML(d.VoidReason)
				}
				return ""
			}())
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en"><head><meta charset="utf-8"/><title>%s %s</title>
<style>
  body{font-family:system-ui,sans-serif;max-width:720px;margin:32px auto;padding:0 16px;color:#111}
  h1{font-size:22px;margin:0 0 4px}
  .muted{color:#666;font-size:13px}
  table{width:100%%;border-collapse:collapse;margin:20px 0}
  th,td{text-align:left;padding:8px;border-bottom:1px solid #ddd;font-size:14px}
  .right{text-align:right}
  .box{border:1px solid #ddd;border-radius:8px;padding:12px 16px;margin:12px 0}
  @media print{.no-print{display:none}}
</style></head><body>
<button class="no-print" onclick="window.print()" style="margin-bottom:16px">Print / Save PDF</button>
<h1>%s</h1>
%s
<p class="muted">Document no. <strong>%s</strong> · Issued %s · Doc status: %s</p>
<div class="box">
  <strong>Seller</strong><br/>%s<br/>%s<br/>Tax ID: %s
</div>
<div class="box">
  <strong>Buyer</strong><br/>%s<br/>%s<br/>Tax ID: %s
</div>
<table>
  <tr><th>Description</th><th class="right">Amount</th></tr>
  <tr><td>%s<br/><span class="muted">Order %s · %s</span></td><td class="right">%s %s</td></tr>
  <tr><td>Net</td><td class="right">%s</td></tr>
  <tr><td>VAT (%s%%)</td><td class="right">%s</td></tr>
  <tr><td><strong>Total</strong></td><td class="right"><strong>%s %s</strong></td></tr>
</table>
<p class="muted">Payment method: %s · Order status: %s · Transaction: %s</p>
</body></html>`,
		title, d.DocNumber,
		title, voidNote, d.DocNumber, d.IssuedAt.UTC().Format("2006-01-02 15:04 UTC"), escHTML(d.Status),
		escHTML(d.SellerName), escHTML(d.SellerAddress), escHTML(d.SellerTaxID),
		escHTML(d.BuyerName), escHTML(d.BuyerAddress), escHTML(d.BuyerTaxID),
		escHTML(d.PackageName), escHTML(o.OrderNo), escHTML(method), amount, escHTML(d.Currency),
		net, rate, vat, amount, escHTML(d.Currency),
		escHTML(method), escHTML(o.Status), escHTML(o.TransactionID),
	)
}

func escHTML(s string) string {
	replacer := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
	)
	return replacer.Replace(s)
}
