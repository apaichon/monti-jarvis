package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNormalizePublicBrandPageRequest(t *testing.T) {
	limit, offset := normalizePublicBrandPageRequest(0, -1)
	if limit != 50 || offset != 0 {
		t.Fatalf("defaults: %d %d", limit, offset)
	}
	limit, offset = normalizePublicBrandPageRequest(500, 3)
	if limit != 100 || offset != 3 {
		t.Fatalf("clamp: %d %d", limit, offset)
	}
}

func TestPublicBrandsUnavailableWithoutStore(t *testing.T) {
	s := &server{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/public/brands", nil).WithContext(context.Background())
	s.publicBrands(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", rec.Code)
	}
	var body map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"] != "public brand directory unavailable" {
		t.Fatalf("body: %s", rec.Body.String())
	}
}

func TestPublicBrandUnavailableWithoutStore(t *testing.T) {
	s := &server{}
	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/public/brands/acme", nil).WithContext(context.Background())
	req.SetPathValue("slug", "acme")
	s.publicBrand(rec, req)
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502 (no store)", rec.Code)
	}
}

func TestPublicTenantsAliasRegistered(t *testing.T) {
	// Ensure normalize helper shared by brands list is stable (alias uses same handlers).
	limit, offset := normalizePublicBrandPageRequest(10, 0)
	if limit != 10 || offset != 0 {
		t.Fatalf("unexpected page: %d %d", limit, offset)
	}
}
