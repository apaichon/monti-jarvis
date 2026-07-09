package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/libra/monti-jarvis/internal/payment"
	"github.com/libra/monti-jarvis/internal/payment/chillpay"
	"github.com/libra/monti-jarvis/internal/store"
)

type paymentGatewayPutRequest struct {
	Provider     string `json:"provider"`
	Mode         string `json:"mode"`
	MerchantCode string `json:"merchant_code"`
	APIKey       string `json:"api_key"`
	MD5Key       string `json:"md5_key"`
	BaseURL      string `json:"base_url"`
	RouteNo      int    `json:"route_no"`
	Currency     string `json:"currency"`
	ReturnURL    string `json:"return_url"`
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
	if provider != "" && provider != payment.ProviderMock && provider != payment.ProviderChillPay {
		writeError(w, http.StatusBadRequest, "provider must be mock or chillpay")
		return
	}
	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if mode != "" && mode != "test" && mode != "live" {
		writeError(w, http.StatusBadRequest, "mode must be test or live")
		return
	}

	callbackURL := payment.DefaultCallbackURL(s.cfg.PublicBaseURL)
	if v := strings.TrimSpace(s.cfg.ChillPayCallbackURL); v != "" {
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
		Provider:     provider,
		Mode:         mode,
		MerchantCode: req.MerchantCode,
		BaseURL:      req.BaseURL,
		RouteNo:      routeNo,
		Currency:     currency,
		CallbackURL:  callbackURL,
		ReturnURL:    req.ReturnURL,
		SetAPIKey:    strings.TrimSpace(req.APIKey) != "",
		APIKey:       req.APIKey,
		SetMD5Key:    strings.TrimSpace(req.MD5Key) != "",
		MD5Key:       req.MD5Key,
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
		writeJSON(w, http.StatusBadGateway, map[string]any{
			"ok":       false,
			"provider": resolved.Provider,
			"message":  err.Error(),
		})
		return
	}
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
	w.WriteHeader(http.StatusOK)
}

func paymentGatewayJSON(row store.PaymentGatewayConfig, resolved payment.ResolvedConfig, lastAt *time.Time, connectionStatus string) map[string]any {
	configured := strings.TrimSpace(resolved.Provider) != "" && resolved.Status == "active"
	out := map[string]any{
		"configured":        configured,
		"provider":          resolved.Provider,
		"mode":              resolved.Mode,
		"status":            resolved.Status,
		"merchant_code":     resolved.MerchantCode,
		"api_key_masked":    payment.MaskSecret(resolved.APIKey),
		"md5_key_set":       strings.TrimSpace(resolved.MD5Key) != "",
		"base_url":          resolved.BaseURL,
		"route_no":          resolved.RouteNo,
		"currency":          resolved.Currency,
		"callback_url":      resolved.CallbackURL,
		"return_url":        resolved.ReturnURL,
		"connection_status": connectionStatus,
		"last_callback_at":  nil,
	}
	if lastAt != nil {
		out["last_callback_at"] = lastAt.UTC().Format(time.RFC3339)
	}
	if !configured && strings.TrimSpace(resolved.Provider) == "" {
		out["provider"] = ""
		out["status"] = "inactive"
	}
	_ = row
	return out
}

func sha256Hex(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

