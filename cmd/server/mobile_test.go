package main

import (
	"testing"
	"time"

	"github.com/libra/monti-jarvis/internal/workforce"
)

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
