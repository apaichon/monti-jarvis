package store

import (
	"strings"
	"testing"
)

func TestNewEmbedKey(t *testing.T) {
	k, err := NewEmbedKey()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(k, "emb_") || len(k) < 10 {
		t.Fatalf("key %q", k)
	}
	k2, _ := NewEmbedKey()
	if k == k2 {
		t.Fatal("keys should differ")
	}
}

func TestValidateOrigin(t *testing.T) {
	if err := ValidateOrigin("https://shop.example"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateOrigin("http://localhost:5500"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateOrigin("ftp://x"); err == nil {
		t.Fatal("expected error")
	}
	if err := ValidateOrigin("not-a-url"); err == nil {
		t.Fatal("expected error")
	}
}

func TestOriginAllowed(t *testing.T) {
	list := []string{"https://Shop.Example", "http://localhost:5500"}
	if !OriginAllowed(list, "https://shop.example", false) {
		t.Fatal("case insensitive host")
	}
	if !OriginAllowed(list, "https://shop.example/", false) {
		t.Fatal("trailing slash")
	}
	if OriginAllowed(list, "https://evil.example", false) {
		t.Fatal("should deny")
	}
	if OriginAllowed(nil, "https://any", false) {
		t.Fatal("empty + !allowEmpty")
	}
	if !OriginAllowed(nil, "https://any", true) {
		t.Fatal("empty + allowEmpty")
	}
	if OriginAllowed(list, "", false) {
		t.Fatal("empty request origin with allowlist")
	}
}

func TestRequestOrigin(t *testing.T) {
	if got := RequestOrigin("https://a.com", ""); got != "https://a.com" {
		t.Fatalf("got %q", got)
	}
	if got := RequestOrigin("", "https://b.com/path?q=1"); got != "https://b.com" {
		t.Fatalf("referer %q", got)
	}
	if got := RequestOrigin("", ""); got != "" {
		t.Fatalf("empty %q", got)
	}
}

func TestParseOrigin(t *testing.T) {
	if got := ParseOrigin("http://localhost:5173"); got != "http://localhost:5173" {
		t.Fatalf("got %q", got)
	}
	if got := ParseOrigin("https://Shop.Example/"); got != "https://shop.example" {
		t.Fatalf("got %q", got)
	}
	if ParseOrigin("http://localhost:5173/page") != "" {
		t.Fatal("path not allowed")
	}
	if ParseOrigin("ftp://x") != "" {
		t.Fatal("ftp not allowed")
	}
	if ParseOrigin("not-a-url") != "" {
		t.Fatal("invalid")
	}
}

func TestEmbedCheckOrigin(t *testing.T) {
	// Prefer parent_origin (host site) over iframe Origin (Monti host).
	got := EmbedCheckOrigin("http://localhost:5173", "http://monti.local:8091", "http://monti.local:8091/embed")
	if got != "http://localhost:5173" {
		t.Fatalf("prefer parent got %q", got)
	}
	// Fall back to browser Origin when parent not provided.
	got = EmbedCheckOrigin("", "https://shop.example", "")
	if got != "https://shop.example" {
		t.Fatalf("fallback Origin got %q", got)
	}
	// Invalid parent_origin ignored.
	got = EmbedCheckOrigin("not-valid", "https://shop.example", "")
	if got != "https://shop.example" {
		t.Fatalf("invalid parent fallback got %q", got)
	}
}
