package payment

import (
	"context"
	"fmt"
	"strings"

	"github.com/libra/monti-jarvis/internal/env"
	"github.com/libra/monti-jarvis/internal/payment/chillpay"
	stripepay "github.com/libra/monti-jarvis/internal/payment/stripe"
	"github.com/libra/monti-jarvis/internal/store"
)

const ProviderMock = "mock"
const ProviderChillPay = "chillpay"
const ProviderStripe = "stripe"

// Gateway resolves config and runs provider health checks.
type Gateway struct {
	cfg   env.Config
	store *store.Store
}

func NewGateway(cfg env.Config, st *store.Store) *Gateway {
	return &Gateway{cfg: cfg, store: st}
}

// ResolvedConfig merges DB row with env overrides.
type ResolvedConfig struct {
	Provider             string
	Mode                 string
	Status               string
	MerchantCode         string
	APIKey               string
	MD5Key               string
	BaseURL              string
	RouteNo              int
	Currency             string
	CallbackURL          string
	ReturnURL            string
	StripePublishableKey string
	StripeSecretKey      string
	StripeWebhookSecret  string
	StripeAPIBaseURL     string
	StripeSuccessURL     string
	StripeCancelURL      string
	Configured           bool
}

func (g *Gateway) Resolve(row store.PaymentGatewayConfig) ResolvedConfig {
	out := ResolvedConfig{
		Provider:             row.Provider,
		Mode:                 row.Mode,
		Status:               row.Status,
		MerchantCode:         row.MerchantCode,
		APIKey:               row.APIKey,
		MD5Key:               row.MD5Key,
		BaseURL:              row.BaseURL,
		RouteNo:              row.RouteNo,
		Currency:             row.Currency,
		CallbackURL:          row.CallbackURL,
		ReturnURL:            row.ReturnURL,
		StripePublishableKey: row.StripePublishableKey,
		StripeSecretKey:      row.StripeSecretKey,
		StripeWebhookSecret:  row.StripeWebhookSecret,
		StripeAPIBaseURL:     row.StripeAPIBaseURL,
		StripeSuccessURL:     row.StripeSuccessURL,
		StripeCancelURL:      row.StripeCancelURL,
	}
	if provider := strings.ToLower(strings.TrimSpace(g.cfg.PaymentGatewayProvider)); provider != "" {
		previousProvider := strings.ToLower(strings.TrimSpace(out.Provider))
		out.Provider = provider
		if previousProvider != "" && previousProvider != provider {
			out.CallbackURL = ""
			out.ReturnURL = ""
		}
		if strings.TrimSpace(out.Status) == "" || strings.EqualFold(out.Status, "inactive") {
			out.Status = "active"
		}
	}
	switch strings.ToLower(strings.TrimSpace(out.Provider)) {
	case ProviderChillPay:
		if v := strings.TrimSpace(g.cfg.ChillPayMerchantCode); v != "" {
			out.MerchantCode = v
		}
		if v := strings.TrimSpace(g.cfg.ChillPayAPIKey); v != "" {
			out.APIKey = v
		}
		if v := strings.TrimSpace(g.cfg.ChillPayMD5Key); v != "" {
			out.MD5Key = v
		}
		if v := strings.TrimSpace(g.cfg.ChillPayBaseURL); v != "" {
			out.BaseURL = v
		}
		if g.cfg.ChillPayRouteNo > 0 {
			out.RouteNo = g.cfg.ChillPayRouteNo
		}
		if v := strings.TrimSpace(g.cfg.ChillPayCurrency); v != "" {
			out.Currency = v
		}
		if v := strings.TrimSpace(g.cfg.ChillPayCallbackURL); v != "" {
			out.CallbackURL = v
		} else if strings.TrimSpace(out.CallbackURL) == "" {
			out.CallbackURL = DefaultCallbackURL(g.cfg.PublicBaseURL)
		}
		if v := strings.TrimSpace(g.cfg.ChillPayReturnURL); v != "" {
			out.ReturnURL = v
		} else if strings.TrimSpace(out.ReturnURL) == "" {
			// Default browser return after ChillPay -> tenant billing return page.
			out.ReturnURL = defaultTenantBillingReturnURL(g.cfg.PublicBaseURL)
		}
	case ProviderStripe:
		if v := strings.TrimSpace(g.cfg.StripePublishableKey); v != "" {
			out.StripePublishableKey = v
		}
		if v := strings.TrimSpace(g.cfg.StripeSecretKey); v != "" {
			out.StripeSecretKey = v
		}
		if v := strings.TrimSpace(g.cfg.StripeWebhookSecret); v != "" {
			out.StripeWebhookSecret = v
		}
		if v := strings.TrimSpace(g.cfg.StripeAPIBaseURL); v != "" {
			out.StripeAPIBaseURL = v
		}
		if v := strings.TrimSpace(g.cfg.StripeSuccessURL); v != "" {
			out.StripeSuccessURL = v
		}
		if v := strings.TrimSpace(g.cfg.StripeCancelURL); v != "" {
			out.StripeCancelURL = v
		}
		if strings.TrimSpace(out.CallbackURL) == "" {
			out.CallbackURL = DefaultStripeCallbackURL(g.cfg.PublicBaseURL)
		}
		if strings.TrimSpace(out.ReturnURL) == "" {
			out.ReturnURL = defaultTenantBillingReturnURL(g.cfg.PublicBaseURL)
		}
		if strings.TrimSpace(out.StripeSuccessURL) == "" {
			out.StripeSuccessURL = out.ReturnURL
		}
		if strings.TrimSpace(out.StripeCancelURL) == "" {
			out.StripeCancelURL = strings.TrimRight(strings.TrimSpace(g.cfg.PublicBaseURL), "/") + "/tenant/billing"
		}
	}
	if out.RouteNo <= 0 {
		out.RouteNo = 1
	}
	if strings.TrimSpace(out.Currency) == "" {
		out.Currency = "764"
	}
	out.Configured = strings.TrimSpace(out.Provider) != "" && out.Status == "active"
	return out
}

func DefaultCallbackURL(publicBase string) string {
	return strings.TrimRight(strings.TrimSpace(publicBase), "/") + "/api/callbacks/chillpay"
}

func DefaultStripeCallbackURL(publicBase string) string {
	return strings.TrimRight(strings.TrimSpace(publicBase), "/") + "/api/callbacks/stripe"
}

func defaultTenantBillingReturnURL(publicBase string) string {
	return strings.TrimRight(strings.TrimSpace(publicBase), "/") + "/tenant/billing/return"
}

func MaskSecret(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return ""
	}
	if len(value) <= 4 {
		return "****"
	}
	return "****" + value[len(value)-4:]
}

func (g *Gateway) Ping(ctx context.Context, resolved ResolvedConfig) error {
	switch strings.ToLower(strings.TrimSpace(resolved.Provider)) {
	case "", ProviderMock:
		return nil
	case ProviderChillPay:
		client := chillpay.NewClient(chillpay.Config{
			MerchantCode: resolved.MerchantCode,
			APIKey:       resolved.APIKey,
			MD5Key:       resolved.MD5Key,
			BaseURL:      resolved.BaseURL,
			RouteNo:      resolved.RouteNo,
			Currency:     resolved.Currency,
			CallbackURL:  resolved.CallbackURL,
			ReturnURL:    resolved.ReturnURL,
		})
		return client.Ping()
	case ProviderStripe:
		client := stripepay.NewClient(stripepay.Config{
			PublishableKey: resolved.StripePublishableKey,
			SecretKey:      resolved.StripeSecretKey,
			WebhookSecret:  resolved.StripeWebhookSecret,
			APIBaseURL:     resolved.StripeAPIBaseURL,
			SuccessURL:     resolved.StripeSuccessURL,
			CancelURL:      resolved.StripeCancelURL,
		})
		return client.Ping(ctx)
	default:
		return fmt.Errorf("unsupported provider %q", resolved.Provider)
	}
}

func (g *Gateway) VerifyCallback(resolved ResolvedConfig, form chillpay.CallbackForm) bool {
	if g.cfg.PaymentCallbackDevBypass {
		return true
	}
	if strings.EqualFold(resolved.Provider, ProviderMock) {
		return true
	}
	client := chillpay.NewClient(chillpay.Config{
		MerchantCode: resolved.MerchantCode,
		APIKey:       resolved.APIKey,
		MD5Key:       resolved.MD5Key,
		BaseURL:      resolved.BaseURL,
		RouteNo:      resolved.RouteNo,
		Currency:     resolved.Currency,
	})
	return client.VerifyCallback(form)
}
