package store

import "testing"

func TestNormalizeCustomerEmail(t *testing.T) {
	got, err := NormalizeCustomerEmail(" Jane.Doe@Example.COM ")
	if err != nil || got != "jane.doe@example.com" {
		t.Fatalf("got %q err=%v", got, err)
	}
	if _, err := NormalizeCustomerEmail("not-email"); err == nil {
		t.Fatal("expected invalid email")
	}
}

func TestNormalizeCustomerDomain(t *testing.T) {
	got, err := NormalizeCustomerDomain(" Example.COM. ")
	if err != nil || got != "example.com" {
		t.Fatalf("got %q err=%v", got, err)
	}
	for _, bad := range []string{"https://example.com", "localhost", "bad domain.com", "-bad.com"} {
		if _, err := NormalizeCustomerDomain(bad); err == nil {
			t.Fatalf("expected invalid domain %q", bad)
		}
	}
	if _, err := NormalizeCustomerDomain("ตัวอย่าง.com"); err == nil {
		t.Fatal("expected non-ASCII domain to require punycode")
	}
}

func TestContainsCredentialMetadata(t *testing.T) {
	if !containsCredentialMetadata(map[string]any{"profile": map[string]any{"access_token": "nope"}}) {
		t.Fatal("expected nested credential key to be rejected")
	}
	if containsCredentialMetadata(map[string]any{"segment": "vip", "score": 4}) {
		t.Fatal("ordinary metadata should be accepted")
	}
}

func TestNormalizeCustomerSource(t *testing.T) {
	got, err := NormalizeCustomerSource("")
	if err != nil || got != "manual" {
		t.Fatalf("got %q err=%v", got, err)
	}
	if _, err := NormalizeCustomerSource("bad/source"); err == nil {
		t.Fatal("expected invalid source")
	}
}
