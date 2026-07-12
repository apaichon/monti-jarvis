package tenantoauth

import "testing"

func TestResolveOAuthRedirectURL(t *testing.T) {
	// Explicit override wins.
	got := resolveOAuthRedirectURL("http://monti-jarvis-dev.local:8091", "http://localhost:8091/api/public/tenant/oauth/google/callback", "google")
	want := "http://localhost:8091/api/public/tenant/oauth/google/callback"
	if got != want {
		t.Fatalf("explicit got %q want %q", got, want)
	}

	// Non-loopback http → rewrite to localhost (Google policy).
	got = resolveOAuthRedirectURL("http://monti-jarvis-dev.local:8091", "", "google")
	if got != want {
		t.Fatalf("rewrite got %q want %q", got, want)
	}

	// localhost stays.
	got = resolveOAuthRedirectURL("http://localhost:8091", "", "google")
	if got != want {
		t.Fatalf("localhost got %q want %q", got, want)
	}

	// https custom host stays (production).
	got = resolveOAuthRedirectURL("https://monti.example.com", "", "google")
	wantHTTPS := "https://monti.example.com/api/public/tenant/oauth/google/callback"
	if got != wantHTTPS {
		t.Fatalf("https got %q want %q", got, wantHTTPS)
	}

	if OAuthCallbackPath("GitHub") != "/api/public/tenant/oauth/github/callback" {
		t.Fatal("callback path")
	}
}
