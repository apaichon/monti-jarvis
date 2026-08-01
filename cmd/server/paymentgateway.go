package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/payment"
	"github.com/libra/monti-jarvis/internal/payment/chillpay"
	stripepay "github.com/libra/monti-jarvis/internal/payment/stripe"
	"github.com/libra/monti-jarvis/internal/store"
)

type stripePaymentGatewayPutRequest struct {
	PublishableKey string `json:"publishable_key"`
	SecretKey      string `json:"secret_key"`
	WebhookSecret  string `json:"webhook_secret"`
	APIBaseURL     string `json:"api_base_url"`
	SuccessURL     string `json:"success_url"`
	CancelURL      string `json:"cancel_url"`
}

type paymentGatewayPutRequest struct {
	Provider     string                         `json:"provider"`
	Mode         string                         `json:"mode"`
	MerchantCode string                         `json:"merchant_code"`
	APIKey       string                         `json:"api_key"`
	MD5Key       string                         `json:"md5_key"`
	BaseURL      string                         `json:"base_url"`
	RouteNo      int                            `json:"route_no"`
	Currency     string                         `json:"currency"`
	ReturnURL    string                         `json:"return_url"`
	Stripe       stripePaymentGatewayPutRequest `json:"stripe"`
}

type paymentGatewayReconcileRequest struct {
	Provider string `json:"provider"`
	Since    string `json:"since"`
	Limit    int    `json:"limit"`
	DryRun   bool   `json:"dry_run"`
}

func (s *server) getPaymentGateway(w http.ResponseWriter, r *http.Request) {
	row, err := s.store.GetPaymentGatewayConfig(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	gw := payment.NewGateway(s.cfg, s.store)
	resolved := gw.Resolve(row)
	lastAt, _ := s.store.LastPaymentCallbackAt(r.Context())
	writeJSON(w, http.StatusOK, paymentGatewayJSON(row, resolved, lastAt, ""))
}

func (s *server) putPaymentGateway(w http.ResponseWriter, r *http.Request) {
	var req paymentGatewayPutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if provider != "" && provider != payment.ProviderMock && provider != payment.ProviderChillPay && provider != payment.ProviderStripe {
		writeError(w, http.StatusBadRequest, "provider must be mock, chillpay, or stripe")
		return
	}
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode != "" && mode != "test" && mode != "live" {
		writeError(w, http.StatusBadRequest, "mode must be test or live")
		return
	}

	callbackURL := payment.DefaultCallbackURL(s.cfg.PublicBaseURL)
	if provider == payment.ProviderStripe {
		callbackURL = payment.DefaultStripeCallbackURL(s.cfg.PublicBaseURL)
	} else if v := strings.TrimSpace(s.cfg.ChillPayCallbackURL); v != "" {
		callbackURL = v
	}

	routeNo := req.RouteNo
	if routeNo <= 0 {
		routeNo = 1
	}
	currency := strings.TrimSpace(req.Currency)
	if currency == "" {
		currency = "764"
	}

	in := store.PaymentGatewayUpsert{
		Provider:               provider,
		Mode:                   mode,
		MerchantCode:           req.MerchantCode,
		BaseURL:                req.BaseURL,
		RouteNo:                routeNo,
		Currency:               currency,
		CallbackURL:            callbackURL,
		ReturnURL:              req.ReturnURL,
		SetAPIKey:              strings.TrimSpace(req.APIKey) != "",
		APIKey:                 req.APIKey,
		SetMD5Key:              strings.TrimSpace(req.MD5Key) != "",
		MD5Key:                 req.MD5Key,
		StripePublishableKey:   req.Stripe.PublishableKey,
		StripeSecretKey:        req.Stripe.SecretKey,
		StripeWebhookSecret:    req.Stripe.WebhookSecret,
		StripeAPIBaseURL:       req.Stripe.APIBaseURL,
		StripeSuccessURL:       req.Stripe.SuccessURL,
		StripeCancelURL:        req.Stripe.CancelURL,
		SetStripeSecret:        strings.TrimSpace(req.Stripe.SecretKey) != "",
		SetStripeWebhookSecret: strings.TrimSpace(req.Stripe.WebhookSecret) != "",
	}
	if provider != "" {
		in.Status = "active"
	}

	row, err := s.store.UpsertPaymentGatewayConfig(r.Context(), in)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	gw := payment.NewGateway(s.cfg, s.store)
	resolved := gw.Resolve(row)
	lastAt, _ := s.store.LastPaymentCallbackAt(r.Context())
	writeJSON(w, http.StatusOK, paymentGatewayJSON(row, resolved, lastAt, ""))
}

func (s *server) testPaymentGateway(w http.ResponseWriter, r *http.Request) {
	row, err := s.store.GetPaymentGatewayConfig(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	gw := payment.NewGateway(s.cfg, s.store)
	resolved := gw.Resolve(row)
	if strings.TrimSpace(resolved.Provider) == "" {
		writeError(w, http.StatusServiceUnavailable, "payment gateway not configured")
		return
	}
	if err := gw.Ping(r.Context(), resolved); err != nil {
		_ = s.store.UpdatePaymentGatewayTestStatus(r.Context(), "failed", err.Error())
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":       false,
			"provider": resolved.Provider,
			"message":  err.Error(),
		})
		return
	}
	_ = s.store.UpdatePaymentGatewayTestStatus(r.Context(), "ok", "")
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":       true,
		"provider": resolved.Provider,
		"message":  "credentials valid",
	})
}

func (s *server) chillpayCallback(w http.ResponseWriter, r *http.Request) {
	rawBody, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	r.Body = io.NopCloser(bytes.NewReader(rawBody))
	if err := r.ParseForm(); err != nil {
		writeError(w, http.StatusBadRequest, "invalid form")
		return
	}
	payloadHash := sha256Hex(string(rawBody))

	form := chillpay.CallbackForm{
		OrderNo:            r.FormValue("OrderNo"),
		Amount:             r.FormValue("Amount"),
		TransactionId:      r.FormValue("TransactionId"),
		CustomerId:         r.FormValue("CustomerId"),
		CustomerName:       r.FormValue("CustomerName"),
		BankCode:           r.FormValue("BankCode"),
		PaymentDate:        r.FormValue("PaymentDate"),
		PaymentStatus:      r.FormValue("PaymentStatus"),
		PaymentDescription: r.FormValue("PaymentDescription"),
		BankRefCode:        r.FormValue("BankRefCode"),
		Currency:           r.FormValue("Currency"),
		CreditCardToken:    r.FormValue("CreditCardToken"),
		CurrentDate:        r.FormValue("CurrentDate"),
		CurrentTime:        r.FormValue("CurrentTime"),
		CheckSum:           r.FormValue("CheckSum"),
	}
	if strings.TrimSpace(form.TransactionId) == "" {
		writeError(w, http.StatusBadRequest, "TransactionId is required")
		return
	}

	row, err := s.store.GetPaymentGatewayConfig(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	gw := payment.NewGateway(s.cfg, s.store)
	resolved := gw.Resolve(row)
	if strings.TrimSpace(resolved.Provider) == "" {
		writeError(w, http.StatusServiceUnavailable, "payment gateway not configured")
		return
	}
	if !gw.VerifyCallback(resolved, form) {
		writeError(w, http.StatusBadRequest, "invalid checksum")
		return
	}

	_, err = s.store.InsertPaymentCallbackEvent(r.Context(), store.PaymentCallbackEvent{
		Provider:      payment.ProviderChillPay,
		TransactionID: form.TransactionId,
		OrderNo:       form.OrderNo,
		PaymentStatus: form.PaymentStatus,
		Amount:        form.Amount,
		CustomerID:    form.CustomerId,
		PayloadHash:   payloadHash,
		ReceivedAt:    time.Now().UTC(),
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	if strings.TrimSpace(form.OrderNo) != "" {
		result, fulfillErr := s.store.FulfillPaymentOrder(r.Context(), form.OrderNo, form.TransactionId, form.PaymentStatus)
		if fulfillErr != nil && !errors.Is(fulfillErr, store.ErrPaymentOrderNotFound) {
			writeError(w, http.StatusBadGateway, fulfillErr.Error())
			return
		}
		if fulfillErr == nil && result.EntitlementChanged {
			s.entitlements.Invalidate(r.Context(), result.Order.TenantID)
		}
	}

	w.WriteHeader(http.StatusOK)
}

func (s *server) stripeCallback(w http.ResponseWriter, r *http.Request) {
	rawBody, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	payloadHash := sha256Hex(string(rawBody))

	row, err := s.store.GetPaymentGatewayConfig(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	gw := payment.NewGateway(s.cfg, s.store)
	resolved := gw.Resolve(row)
	if !strings.EqualFold(resolved.Provider, payment.ProviderStripe) {
		writeError(w, http.StatusServiceUnavailable, "stripe payment gateway not configured")
		return
	}
	client := stripepay.NewClient(stripepay.Config{
		PublishableKey: resolved.StripePublishableKey,
		SecretKey:      resolved.StripeSecretKey,
		WebhookSecret:  resolved.StripeWebhookSecret,
		APIBaseURL:     resolved.StripeAPIBaseURL,
		SuccessURL:     resolved.StripeSuccessURL,
		CancelURL:      resolved.StripeCancelURL,
	})

	var ev stripepay.Event
	signatureVerified := true
	if s.cfg.PaymentCallbackDevBypass {
		signatureVerified = false
		if err := json.Unmarshal(rawBody, &ev); err != nil {
			writeError(w, http.StatusBadRequest, "invalid stripe webhook JSON")
			return
		}
	} else {
		ev, err = client.ParseAndVerifyWebhook(rawBody, r.Header.Get("Stripe-Signature"), time.Now().UTC())
		if err != nil {
			_ = s.store.UpdatePaymentGatewayWebhookStatus(r.Context(), "failed", err.Error())
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
	}
	if strings.TrimSpace(ev.ID) == "" {
		writeError(w, http.StatusBadRequest, "stripe webhook missing event id")
		return
	}

	inserted, err := s.store.InsertPaymentCallbackEvent(r.Context(), store.PaymentCallbackEvent{
		Provider:          payment.ProviderStripe,
		TransactionID:     ev.ID,
		ProviderEventID:   ev.ID,
		EventType:         ev.Type,
		SignatureVerified: signatureVerified,
		ProcessingStatus:  "received",
		PaymentStatus:     ev.Type,
		PayloadHash:       payloadHash,
		ReceivedAt:        time.Now().UTC(),
	})
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	if !inserted {
		writeJSON(w, http.StatusOK, map[string]any{"received": true, "duplicate": true})
		return
	}

	session, sessionOK := stripeEventCheckoutSession(ev)
	if !sessionOK {
		_ = s.store.UpdatePaymentGatewayWebhookStatus(r.Context(), "ignored", "")
		writeJSON(w, http.StatusOK, map[string]any{"received": true, "ignored": true, "event_type": ev.Type})
		return
	}

	orderNo := firstNonEmpty(session.Metadata["order_no"], session.ClientReference)
	order, err := s.store.GetPaymentOrderByOrderNo(r.Context(), orderNo)
	if err != nil {
		_ = s.store.UpdatePaymentGatewayWebhookStatus(r.Context(), "unmatched", err.Error())
		writeJSON(w, http.StatusOK, map[string]any{"received": true, "matched": false, "event_type": ev.Type})
		return
	}
	txnID := firstNonEmpty(session.PaymentIntent, session.ID)
	_ = s.store.UpdatePaymentOrderProviderRefs(
		r.Context(),
		order.ID,
		txnID,
		"",
		session.ID,
		session.PaymentIntent,
		stripeProviderStatus(session),
		session.ExpiresAtTime(),
	)

	if !stripeSessionMatchesOrder(session, *order) {
		_ = s.store.UpdatePaymentGatewayWebhookStatus(r.Context(), "failed", "stripe session does not match local order")
		writeJSON(w, http.StatusOK, map[string]any{"received": true, "matched": false, "event_type": ev.Type})
		return
	}
	if session.Paid() {
		result, fulfillErr := s.store.FulfillPaymentOrder(r.Context(), order.OrderNo, txnID, "0")
		if fulfillErr != nil {
			_ = s.store.UpdatePaymentGatewayWebhookStatus(r.Context(), "failed", fulfillErr.Error())
			writeError(w, http.StatusBadGateway, fulfillErr.Error())
			return
		}
		if result.EntitlementChanged {
			s.entitlements.Invalidate(r.Context(), result.Order.TenantID)
		}
		_ = s.store.UpdatePaymentGatewayWebhookStatus(r.Context(), "processed", "")
		writeJSON(w, http.StatusOK, map[string]any{"received": true, "fulfilled": true, "order_no": order.OrderNo})
		return
	}
	if session.FailedOrExpired() {
		if _, fulfillErr := s.store.FulfillPaymentOrder(r.Context(), order.OrderNo, txnID, "2"); fulfillErr != nil {
			_ = s.store.UpdatePaymentGatewayWebhookStatus(r.Context(), "failed", fulfillErr.Error())
			writeError(w, http.StatusBadGateway, fulfillErr.Error())
			return
		}
		_ = s.store.UpdatePaymentGatewayWebhookStatus(r.Context(), "processed", "")
		writeJSON(w, http.StatusOK, map[string]any{"received": true, "failed": true, "order_no": order.OrderNo})
		return
	}

	_ = s.store.UpdatePaymentGatewayWebhookStatus(r.Context(), "received", "")
	writeJSON(w, http.StatusOK, map[string]any{"received": true, "pending": true, "order_no": order.OrderNo})
}

func (s *server) reconcilePaymentGateway(w http.ResponseWriter, r *http.Request) {
	var req paymentGatewayReconcileRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}
	}
	row, err := s.store.GetPaymentGatewayConfig(r.Context())
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	gw := payment.NewGateway(s.cfg, s.store)
	resolved := gw.Resolve(row)
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if provider == "" {
		provider = strings.ToLower(strings.TrimSpace(resolved.Provider))
	}
	if provider != payment.ProviderStripe {
		writeError(w, http.StatusBadRequest, "reconcile currently supports stripe")
		return
	}
	var since *time.Time
	if trimmed := strings.TrimSpace(req.Since); trimmed != "" {
		parsed, err := time.Parse(time.RFC3339, trimmed)
		if err != nil {
			writeError(w, http.StatusBadRequest, "since must be RFC3339")
			return
		}
		since = &parsed
	}
	orders, err := s.store.ListPaymentOrdersForProviderReconcile(r.Context(), provider, since, req.Limit)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}
	client := stripepay.NewClient(stripepay.Config{
		PublishableKey: resolved.StripePublishableKey,
		SecretKey:      resolved.StripeSecretKey,
		WebhookSecret:  resolved.StripeWebhookSecret,
		APIBaseURL:     resolved.StripeAPIBaseURL,
		SuccessURL:     resolved.StripeSuccessURL,
		CancelURL:      resolved.StripeCancelURL,
	})

	items := make([]map[string]any, 0, len(orders))
	for _, order := range orders {
		session, err := client.RetrieveCheckoutSession(r.Context(), order.ProviderSessionID)
		item := map[string]any{
			"order_id":            order.ID,
			"order_no":            order.OrderNo,
			"local_status":        order.Status,
			"provider_session_id": order.ProviderSessionID,
			"ok":                  err == nil,
		}
		if err != nil {
			item["error"] = err.Error()
			items = append(items, item)
			continue
		}
		item["provider_status"] = stripeProviderStatus(session)
		item["provider_payment_id"] = session.PaymentIntent
		item["matches"] = stripeSessionMatchesOrder(session, order)
		if !req.DryRun {
			_ = s.store.UpdatePaymentOrderProviderRefs(
				r.Context(), order.ID, firstNonEmpty(session.PaymentIntent, session.ID), "",
				session.ID, session.PaymentIntent, stripeProviderStatus(session), session.ExpiresAtTime(),
			)
			if order.Status == store.PaymentOrderStatusPending && session.Paid() && stripeSessionMatchesOrder(session, order) {
				result, fulfillErr := s.store.FulfillPaymentOrder(r.Context(), order.OrderNo, firstNonEmpty(session.PaymentIntent, session.ID), "0")
				if fulfillErr == nil && result.EntitlementChanged {
					s.entitlements.Invalidate(r.Context(), result.Order.TenantID)
				}
				item["fulfilled"] = fulfillErr == nil
				if fulfillErr != nil {
					item["fulfill_error"] = fulfillErr.Error()
				}
			}
		}
		items = append(items, item)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"provider": provider,
		"dry_run":  req.DryRun,
		"count":    len(items),
		"items":    items,
	})
}

func paymentGatewayJSON(row store.PaymentGatewayConfig, resolved payment.ResolvedConfig, lastAt *time.Time, connectionStatus string) map[string]any {
	configured := strings.TrimSpace(resolved.Provider) != "" && resolved.Status == "active"
	out := map[string]any{
		"configured":          configured,
		"provider":            resolved.Provider,
		"mode":                resolved.Mode,
		"status":              resolved.Status,
		"merchant_code":       resolved.MerchantCode,
		"api_key_masked":      payment.MaskSecret(resolved.APIKey),
		"md5_key_set":         strings.TrimSpace(resolved.MD5Key) != "",
		"base_url":            resolved.BaseURL,
		"route_no":            resolved.RouteNo,
		"currency":            resolved.Currency,
		"callback_url":        resolved.CallbackURL,
		"return_url":          resolved.ReturnURL,
		"connection_status":   connectionStatus,
		"last_callback_at":    nil,
		"last_test_status":    row.LastTestStatus,
		"last_tested_at":      timePtrRFC3339(row.LastTestedAt),
		"last_test_error":     row.LastTestError,
		"last_webhook_status": row.LastWebhookStatus,
		"last_webhook_at":     timePtrRFC3339(row.LastWebhookAt),
		"stripe": map[string]any{
			"publishable_key":    resolved.StripePublishableKey,
			"secret_key_set":     strings.TrimSpace(resolved.StripeSecretKey) != "",
			"webhook_secret_set": strings.TrimSpace(resolved.StripeWebhookSecret) != "",
			"api_base_url":       resolved.StripeAPIBaseURL,
			"success_url":        resolved.StripeSuccessURL,
			"cancel_url":         resolved.StripeCancelURL,
			"callback_url":       stripeCallbackURLForJSON(resolved),
		},
	}
	if lastAt != nil {
		out["last_callback_at"] = lastAt.UTC().Format(time.RFC3339)
	}
	if !configured && strings.TrimSpace(resolved.Provider) == "" {
		out["provider"] = ""
		out["status"] = "inactive"
	}
	return out
}

func stripeCallbackURLForJSON(resolved payment.ResolvedConfig) string {
	if strings.EqualFold(resolved.Provider, payment.ProviderStripe) {
		return resolved.CallbackURL
	}
	return ""
}

func stripeEventCheckoutSession(ev stripepay.Event) (stripepay.CheckoutSession, bool) {
	switch ev.Type {
	case "checkout.session.completed", "checkout.session.async_payment_succeeded", "checkout.session.async_payment_failed", "checkout.session.expired":
		session, err := stripepay.ParseCheckoutSession(ev.Data.Object)
		return session, err == nil && strings.TrimSpace(session.ID) != ""
	default:
		return stripepay.CheckoutSession{}, false
	}
}

func stripeProviderStatus(session stripepay.CheckoutSession) string {
	status := strings.TrimSpace(session.Status)
	if paymentStatus := strings.TrimSpace(session.PaymentStatus); paymentStatus != "" {
		if status != "" {
			return status + ":" + paymentStatus
		}
		return paymentStatus
	}
	return status
}

func stripeSessionMatchesOrder(session stripepay.CheckoutSession, order store.PaymentOrder) bool {
	if session.AmountTotal > 0 && int(session.AmountTotal) != order.AmountCents {
		return false
	}
	if strings.TrimSpace(session.Currency) == "" {
		return true
	}
	sessionCurrency := stripepay.NormalizeCurrency(session.Currency)
	orderCurrency := stripepay.NormalizeCurrency(order.Currency)
	return sessionCurrency == orderCurrency
}

func timePtrRFC3339(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339)
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}
