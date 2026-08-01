package stripe

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const defaultAPIBaseURL = "https://api.stripe.com"

type Config struct {
	PublishableKey string
	SecretKey      string
	WebhookSecret  string
	APIBaseURL     string
	SuccessURL     string
	CancelURL      string
}

type Client struct {
	cfg        Config
	httpClient *http.Client
}

func NewClient(cfg Config) *Client {
	if strings.TrimSpace(cfg.APIBaseURL) == "" {
		cfg.APIBaseURL = defaultAPIBaseURL
	}
	return &Client{
		cfg: cfg,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type CheckoutSessionInput struct {
	OrderID     string
	OrderNo     string
	TenantID    string
	PackageID   string
	PackageName string
	AmountCents int
	Currency    string
	SuccessURL  string
	CancelURL   string
}

type CheckoutSession struct {
	ID              string            `json:"id"`
	URL             string            `json:"url"`
	ClientReference string            `json:"client_reference_id"`
	PaymentIntent   string            `json:"payment_intent"`
	PaymentStatus   string            `json:"payment_status"`
	Status          string            `json:"status"`
	AmountTotal     int64             `json:"amount_total"`
	Currency        string            `json:"currency"`
	ExpiresAt       int64             `json:"expires_at"`
	Metadata        map[string]string `json:"metadata"`
}

type ErrorResponse struct {
	Error struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error"`
}

type Event struct {
	ID   string `json:"id"`
	Type string `json:"type"`
	Data struct {
		Object json.RawMessage `json:"object"`
	} `json:"data"`
}

func (c *Client) Ping(ctx context.Context) error {
	if strings.TrimSpace(c.cfg.SecretKey) == "" {
		return fmt.Errorf("stripe secret_key is required")
	}
	req, err := c.newRequest(ctx, http.MethodGet, "/v1/account", nil)
	if err != nil {
		return err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("stripe account request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return c.responseError("stripe account", resp)
	}
	return nil
}

func (c *Client) CreateCheckoutSession(ctx context.Context, in CheckoutSessionInput) (CheckoutSession, error) {
	if strings.TrimSpace(c.cfg.SecretKey) == "" {
		return CheckoutSession{}, fmt.Errorf("stripe secret_key is required")
	}
	if in.AmountCents <= 0 {
		return CheckoutSession{}, fmt.Errorf("amount_cents must be positive")
	}
	successURL := firstNonEmpty(in.SuccessURL, c.cfg.SuccessURL)
	cancelURL := firstNonEmpty(in.CancelURL, c.cfg.CancelURL)
	if successURL == "" || cancelURL == "" {
		return CheckoutSession{}, fmt.Errorf("stripe success_url and cancel_url are required")
	}
	name := strings.TrimSpace(in.PackageName)
	if name == "" {
		name = "Monti package"
	}
	currency := NormalizeCurrency(in.Currency)
	form := url.Values{}
	form.Set("mode", "payment")
	form.Set("success_url", successURL)
	form.Set("cancel_url", cancelURL)
	form.Set("client_reference_id", strings.TrimSpace(in.OrderNo))
	form.Set("line_items[0][price_data][currency]", currency)
	form.Set("line_items[0][price_data][product_data][name]", name)
	form.Set("line_items[0][price_data][unit_amount]", strconv.Itoa(in.AmountCents))
	form.Set("line_items[0][quantity]", "1")
	metadata := map[string]string{
		"order_id":   strings.TrimSpace(in.OrderID),
		"order_no":   strings.TrimSpace(in.OrderNo),
		"tenant_id":  strings.TrimSpace(in.TenantID),
		"package_id": strings.TrimSpace(in.PackageID),
	}
	for key, value := range metadata {
		if value != "" {
			form.Set("metadata["+key+"]", value)
			form.Set("payment_intent_data[metadata]["+key+"]", value)
		}
	}
	req, err := c.newRequest(ctx, http.MethodPost, "/v1/checkout/sessions", strings.NewReader(form.Encode()))
	if err != nil {
		return CheckoutSession{}, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return CheckoutSession{}, fmt.Errorf("stripe checkout session request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return CheckoutSession{}, c.responseError("stripe checkout session", resp)
	}
	var out CheckoutSession
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return CheckoutSession{}, fmt.Errorf("stripe checkout session parse: %w", err)
	}
	if strings.TrimSpace(out.ID) == "" || strings.TrimSpace(out.URL) == "" {
		return CheckoutSession{}, fmt.Errorf("stripe checkout session response missing id or url")
	}
	return out, nil
}

func (c *Client) RetrieveCheckoutSession(ctx context.Context, sessionID string) (CheckoutSession, error) {
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return CheckoutSession{}, fmt.Errorf("stripe session_id is required")
	}
	req, err := c.newRequest(ctx, http.MethodGet, "/v1/checkout/sessions/"+url.PathEscape(sessionID), nil)
	if err != nil {
		return CheckoutSession{}, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return CheckoutSession{}, fmt.Errorf("stripe checkout session retrieve: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return CheckoutSession{}, c.responseError("stripe checkout session retrieve", resp)
	}
	var out CheckoutSession
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return CheckoutSession{}, fmt.Errorf("stripe checkout session parse: %w", err)
	}
	return out, nil
}

func (c *Client) ParseAndVerifyWebhook(raw []byte, signatureHeader string, now time.Time) (Event, error) {
	if err := VerifySignature(raw, signatureHeader, c.cfg.WebhookSecret, now); err != nil {
		return Event{}, err
	}
	var ev Event
	if err := json.Unmarshal(raw, &ev); err != nil {
		return Event{}, fmt.Errorf("stripe webhook parse: %w", err)
	}
	if strings.TrimSpace(ev.ID) == "" || strings.TrimSpace(ev.Type) == "" {
		return Event{}, fmt.Errorf("stripe webhook missing id or type")
	}
	return ev, nil
}

func ParseCheckoutSession(raw json.RawMessage) (CheckoutSession, error) {
	var session CheckoutSession
	if len(raw) == 0 {
		return session, fmt.Errorf("stripe checkout session payload is empty")
	}
	if err := json.Unmarshal(raw, &session); err != nil {
		return session, fmt.Errorf("stripe checkout session payload parse: %w", err)
	}
	return session, nil
}

func VerifySignature(raw []byte, signatureHeader, webhookSecret string, now time.Time) error {
	webhookSecret = strings.TrimSpace(webhookSecret)
	if webhookSecret == "" {
		return fmt.Errorf("stripe webhook_secret is required")
	}
	timestamp, signatures := parseSignatureHeader(signatureHeader)
	if timestamp == 0 || len(signatures) == 0 {
		return fmt.Errorf("stripe signature header is invalid")
	}
	if !now.IsZero() {
		eventTime := time.Unix(timestamp, 0)
		if now.Sub(eventTime) > 5*time.Minute || eventTime.Sub(now) > 5*time.Minute {
			return fmt.Errorf("stripe signature timestamp outside tolerance")
		}
	}
	expected := computeSignature(timestamp, raw, webhookSecret)
	for _, sig := range signatures {
		if hmac.Equal([]byte(expected), []byte(sig)) {
			return nil
		}
	}
	return fmt.Errorf("stripe signature mismatch")
}

func SignatureHeaderForTest(raw []byte, webhookSecret string, at time.Time) string {
	if at.IsZero() {
		at = time.Now().UTC()
	}
	ts := at.Unix()
	return fmt.Sprintf("t=%d,v1=%s", ts, computeSignature(ts, raw, webhookSecret))
}

func NormalizeCurrency(raw string) string {
	value := strings.ToLower(strings.TrimSpace(raw))
	switch value {
	case "", "764":
		return "thb"
	default:
		return value
	}
}

func (s CheckoutSession) Paid() bool {
	return strings.EqualFold(s.PaymentStatus, "paid") || strings.EqualFold(s.Status, "complete")
}

func (s CheckoutSession) FailedOrExpired() bool {
	return strings.EqualFold(s.Status, "expired") || strings.EqualFold(s.PaymentStatus, "failed")
}

func (s CheckoutSession) ExpiresAtTime() *time.Time {
	if s.ExpiresAt <= 0 {
		return nil
	}
	t := time.Unix(s.ExpiresAt, 0).UTC()
	return &t
}

func (c *Client) newRequest(ctx context.Context, method, path string, body io.Reader) (*http.Request, error) {
	endpoint := strings.TrimRight(strings.TrimSpace(c.cfg.APIBaseURL), "/") + path
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("stripe build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(c.cfg.SecretKey))
	return req, nil
}

func (c *Client) responseError(prefix string, resp *http.Response) error {
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	var er ErrorResponse
	if err := json.Unmarshal(body, &er); err == nil && strings.TrimSpace(er.Error.Message) != "" {
		return fmt.Errorf("%s HTTP %d: %s", prefix, resp.StatusCode, er.Error.Message)
	}
	if len(bytes.TrimSpace(body)) == 0 {
		return fmt.Errorf("%s HTTP %d", prefix, resp.StatusCode)
	}
	return fmt.Errorf("%s HTTP %d: %s", prefix, resp.StatusCode, string(body))
}

func parseSignatureHeader(header string) (int64, []string) {
	var timestamp int64
	var signatures []string
	for _, part := range strings.Split(header, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch strings.TrimSpace(key) {
		case "t":
			ts, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
			if err == nil {
				timestamp = ts
			}
		case "v1":
			if sig := strings.TrimSpace(value); sig != "" {
				signatures = append(signatures, sig)
			}
		}
	}
	return timestamp, signatures
}

func computeSignature(timestamp int64, raw []byte, webhookSecret string) string {
	payload := append([]byte(strconv.FormatInt(timestamp, 10)+"."), raw...)
	mac := hmac.New(sha256.New, []byte(webhookSecret))
	_, _ = mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return ""
}
