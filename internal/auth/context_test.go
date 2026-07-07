package auth

import (
	"context"
	"testing"
)

func TestResolveTenantAuthDisabled(t *testing.T) {
	got := ResolveTenant(context.Background(), "acme", true, "demo")
	if got != "acme" {
		t.Fatalf("expected acme, got %q", got)
	}
	got = ResolveTenant(context.Background(), "", true, "demo")
	if got != "demo" {
		t.Fatalf("expected demo, got %q", got)
	}
}

func TestResolveTenantPlatformAdmin(t *testing.T) {
	ctx := WithContext(context.Background(), AuthContext{
		Role:     RolePlatformAdmin,
		TenantID: "",
	})
	got := ResolveTenant(ctx, "other", false, "demo")
	if got != "other" {
		t.Fatalf("expected header override other, got %q", got)
	}
	got = ResolveTenant(ctx, "", false, "demo")
	if got != "demo" {
		t.Fatalf("expected demo fallback, got %q", got)
	}
}

func TestResolveTenantTenantAdmin(t *testing.T) {
	ctx := WithContext(context.Background(), AuthContext{
		Role:     RoleTenantAdmin,
		TenantID: "demo",
	})
	got := ResolveTenant(ctx, "other", false, "demo")
	if got != "demo" {
		t.Fatalf("expected claim demo, got %q", got)
	}
}