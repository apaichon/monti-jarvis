package env

import (
	"os"
	"testing"
)

func TestResolveResendFrom_EmailFull(t *testing.T) {
	t.Setenv("RESEND_FROM_EMAIL", "Monti <noreply@devclub.dev>")
	t.Setenv("RESEND_FROM_ADDR", "ignored@example.com")
	got := resolveResendFrom()
	if got != "Monti <noreply@devclub.dev>" {
		t.Fatalf("got %q", got)
	}
}

func TestResolveResendFrom_AddrAndName(t *testing.T) {
	_ = os.Unsetenv("RESEND_FROM_EMAIL")
	t.Setenv("RESEND_FROM_EMAIL", "")
	t.Setenv("RESEND_FROM_ADDR", "no-reply@devclub.dev")
	t.Setenv("RESEND_FROM_NAME", "Monti [No Reply]")
	got := resolveResendFrom()
	want := "Monti [No Reply] <no-reply@devclub.dev>"
	if got != want {
		t.Fatalf("got %q want %q", got, want)
	}
}

func TestResolveResendFrom_EmptyNoMontiLocal(t *testing.T) {
	t.Setenv("RESEND_FROM_EMAIL", "")
	t.Setenv("RESEND_FROM_ADDR", "")
	t.Setenv("RESEND_FROM_NAME", "")
	t.Setenv("RESEND_FROM", "")
	got := resolveResendFrom()
	if got != "" {
		t.Fatalf("expected empty from, got %q", got)
	}
	if containsMontiLocal(got) {
		t.Fatal("must not default to monti.local")
	}
}

func TestResolveResendAPIKey_Disabled(t *testing.T) {
	t.Setenv("RESEND_ENABLED", "false")
	t.Setenv("RESEND_API_KEY", "re_test")
	if got := resolveResendAPIKey(); got != "" {
		t.Fatalf("expected empty key when disabled, got %q", got)
	}
}

func containsMontiLocal(s string) bool {
	return len(s) > 0 && (s == "Monti <onboarding@monti.local>" || len(s) >= 11 && s[len(s)-11:] == "monti.local")
}
