package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/libra/monti-jarvis/internal/env"
	"github.com/libra/monti-jarvis/internal/workforce"
)

func TestMobileAPIGate(t *testing.T) {
	s := &server{cfg: env.Config{MobileCallAPIEnabled: false}}
	recorder := httptest.NewRecorder()
	s.mobileAPI(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/mobile/v1/bootstrap", nil))
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("disabled mobile API status = %d, want 404", recorder.Code)
	}

	s.cfg.MobileCallAPIEnabled = true
	recorder = httptest.NewRecorder()
	s.mobileAPI(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/api/mobile/v1/bootstrap", nil))
	if recorder.Code != http.StatusNoContent {
		t.Fatalf("enabled mobile API status = %d, want 204", recorder.Code)
	}
}

func TestMobileIdempotencyKeyIsScoped(t *testing.T) {
	first := mobileIdempotencyKey("tenant-a", "customer-a", "create", "same")
	if first == mobileIdempotencyKey("tenant-b", "customer-a", "create", "same") {
		t.Fatal("tenant must scope idempotency keys")
	}
	if first == mobileIdempotencyKey("tenant-a", "customer-b", "create", "same") {
		t.Fatal("caller must scope idempotency keys")
	}
	if first == mobileIdempotencyKey("tenant-a", "customer-a", "end", "same") {
		t.Fatal("route must scope idempotency keys")
	}
}

func TestMobileIdempotencyCacheExpires(t *testing.T) {
	key := "mobile:test-expiry"
	mobileIdempotency.Store(key, mobileCacheEntry{ExpiresAt: time.Now().Add(-time.Second), Body: []byte(`{"ok":true}`)})
	if _, ok := mobileCached(key); ok {
		t.Fatal("expired idempotency response must not be replayed")
	}
}

func TestMobileAvatarSelectionRequiresAssignment(t *testing.T) {
	agents := []workforce.Agent{{ID: "ava"}}
	if !containsAgent(agents, "ava") {
		t.Fatal("assigned avatar should be selectable")
	}
	if containsAgent(agents, "unassigned") {
		t.Fatal("unassigned avatar must not be selectable")
	}
}
