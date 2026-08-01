package stripe

import (
	"testing"
	"time"
)

func TestVerifySignature(t *testing.T) {
	raw := []byte(`{"id":"evt_test","type":"checkout.session.completed"}`)
	secret := "whsec_test"
	now := time.Unix(1700000000, 0).UTC()
	header := SignatureHeaderForTest(raw, secret, now)

	if err := VerifySignature(raw, header, secret, now); err != nil {
		t.Fatalf("VerifySignature() error = %v", err)
	}
	if err := VerifySignature([]byte(`{"tampered":true}`), header, secret, now); err == nil {
		t.Fatalf("VerifySignature() succeeded for tampered payload")
	}
	if err := VerifySignature(raw, header, "wrong", now); err == nil {
		t.Fatalf("VerifySignature() succeeded for wrong secret")
	}
}

func TestNormalizeCurrency(t *testing.T) {
	tests := map[string]string{
		"":      "thb",
		"764":   "thb",
		"THB":   "thb",
		"usd":   "usd",
		" USD ": "usd",
	}
	for input, want := range tests {
		if got := NormalizeCurrency(input); got != want {
			t.Fatalf("NormalizeCurrency(%q) = %q, want %q", input, got, want)
		}
	}
}
