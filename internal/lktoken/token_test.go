package lktoken

import (
	"testing"
	"time"
)

func TestJoinToken(t *testing.T) {
	cfg := Config{APIKey: "devkey", APISecret: "secret", LiveURL: "ws://localhost:7880"}
	token, err := cfg.JoinToken("monti-room-1", "caller-1", time.Hour)
	if err != nil {
		t.Fatalf("JoinToken() error = %v", err)
	}
	if token == "" {
		t.Fatal("JoinToken() returned empty token")
	}
}

func TestJoinTokenRequiresConfig(t *testing.T) {
	cfg := Config{}
	if _, err := cfg.JoinToken("room", "caller", time.Hour); err == nil {
		t.Fatal("JoinToken() without config should fail")
	}
}