package main

import (
	"context"
	"errors"
	"testing"

	"github.com/libra/monti-jarvis/internal/env"
	"github.com/libra/monti-jarvis/internal/store"
)

func TestResolveTenantGeminiAPIKeyFailsClosedInProduction(t *testing.T) {
	t.Parallel()

	s := &server{cfg: env.Config{
		AppEnv:                      "production",
		GeminiAPIKey:                "platform-key-must-not-be-used",
		AllowPlatformGeminiFallback: true,
	}}
	key, err := s.resolveTenantGeminiAPIKey(context.Background(), "tenant-1")
	if key != "" {
		t.Fatalf("production resolver returned platform key %q", key)
	}
	if !errors.Is(err, store.ErrTenantGeminiKeyRequired) {
		t.Fatalf("production resolver error = %v, want tenant key required", err)
	}
}

func TestResolveTenantGeminiAPIKeyAllowsExplicitDevelopmentFallback(t *testing.T) {
	t.Parallel()

	s := &server{cfg: env.Config{
		AppEnv:                      "dev",
		GeminiAPIKey:                "development-platform-key",
		AllowPlatformGeminiFallback: true,
	}}
	key, err := s.resolveTenantGeminiAPIKey(context.Background(), "tenant-1")
	if err != nil {
		t.Fatalf("development resolver returned error: %v", err)
	}
	if key != "development-platform-key" {
		t.Fatalf("development resolver key = %q", key)
	}
}

func TestGeminiFailureStatus(t *testing.T) {
	t.Parallel()

	if got := geminiFailureStatus("auth"); got != store.GeminiKeyStatusInvalid {
		t.Fatalf("auth failure status = %q", got)
	}
	for _, class := range []string{"network", "quota", "unknown"} {
		if got := geminiFailureStatus(class); got != store.GeminiKeyStatusDegraded {
			t.Fatalf("%s failure status = %q", class, got)
		}
	}
}
