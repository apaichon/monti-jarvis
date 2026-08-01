package payment

import (
	"testing"

	"github.com/libra/monti-jarvis/internal/env"
	"github.com/libra/monti-jarvis/internal/store"
)

func TestMaskSecret(t *testing.T) {
	if MaskSecret("") != "" {
		t.Fatal("empty")
	}
	if MaskSecret("ab") != "****" {
		t.Fatal("short")
	}
	if got := MaskSecret("abcdefghij"); got != "****ghij" {
		t.Fatalf("got %q", got)
	}
}

func TestDefaultCallbackURL(t *testing.T) {
	got := DefaultCallbackURL("http://localhost:8091/")
	want := "http://localhost:8091/api/callbacks/chillpay"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestGatewayResolvePaymentGatewayProviderEnvOverride(t *testing.T) {
	gw := NewGateway(env.Config{
		PublicBaseURL:          "http://localhost:8091",
		PaymentGatewayProvider: ProviderStripe,
		StripePublishableKey:   "pk_test",
		StripeSecretKey:        "sk_test",
		StripeWebhookSecret:    "whsec_test",
		StripeAPIBaseURL:       "https://api.stripe.com",
	}, nil)

	resolved := gw.Resolve(store.PaymentGatewayConfig{
		Provider:     ProviderChillPay,
		Status:       "active",
		CallbackURL:  "https://example.ngrok.dev/api/callbacks/chillpay",
		ReturnURL:    "https://example.ngrok.dev/api/callbacks/chillpay/return",
		MerchantCode: "legacy",
	})

	if resolved.Provider != ProviderStripe {
		t.Fatalf("provider = %q, want %q", resolved.Provider, ProviderStripe)
	}
	if !resolved.Configured {
		t.Fatalf("expected resolved config to be configured")
	}
	if resolved.CallbackURL != "http://localhost:8091/api/callbacks/stripe" {
		t.Fatalf("callback_url = %q", resolved.CallbackURL)
	}
	if resolved.StripeSecretKey != "sk_test" {
		t.Fatalf("stripe secret was not resolved from env")
	}
}
