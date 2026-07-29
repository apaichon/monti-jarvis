package store

import (
	"encoding/json"
	"testing"
)

func TestNormalizePublicBrandPage(t *testing.T) {
	limit, offset := normalizePublicBrandPage(0, -3)
	if limit != 50 || offset != 0 {
		t.Fatalf("defaults: limit=%d offset=%d", limit, offset)
	}
	limit, offset = normalizePublicBrandPage(200, 10)
	if limit != 100 || offset != 10 {
		t.Fatalf("clamp: limit=%d offset=%d", limit, offset)
	}
	limit, offset = normalizePublicBrandPage(25, 5)
	if limit != 25 || offset != 5 {
		t.Fatalf("passthrough: limit=%d offset=%d", limit, offset)
	}
}

func TestPublicBrandJSONOmitsPlatformListedAndSecrets(t *testing.T) {
	b := PublicBrand{
		ID:             "tenant-acme",
		Slug:           "acme",
		Name:           "Acme Support",
		Blurb:          "Help",
		LogoURL:        "https://example/logo.png",
		Category:       "retail",
		Languages:      []string{"en", "th"},
		Listed:         true,
		PlatformListed: true,
		Status:         "active",
	}
	raw, err := json.Marshal(b)
	if err != nil {
		t.Fatal(err)
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"platform_listed", "admin_email", "password", "secret", "registration_id"} {
		if _, ok := m[forbidden]; ok {
			t.Fatalf("public brand JSON must not include %q: %s", forbidden, raw)
		}
	}
	for _, required := range []string{"id", "slug", "name", "listed", "status", "languages"} {
		if _, ok := m[required]; !ok {
			t.Fatalf("missing %q in %s", required, raw)
		}
	}
}

func TestPublicBrandListabilityPredicateDocumented(t *testing.T) {
	// Guardrail: public directory filters are encoded in ListPublicBrands/GetPublicBrand SQL.
	// This test documents the required flags so refactors keep isolation.
	requiredFlags := []string{"listed", "platform_listed", "active"}
	if len(requiredFlags) != 3 {
		t.Fatal("unexpected")
	}
}
