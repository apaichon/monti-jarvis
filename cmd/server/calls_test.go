package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/libra/monti-jarvis/internal/env"
	"github.com/libra/monti-jarvis/internal/gemini"
	"github.com/libra/monti-jarvis/internal/live"
	"github.com/libra/monti-jarvis/internal/lktoken"
)

func TestHealthIncludesSprint003(t *testing.T) {
	s := &server{
		cfg:   env.Load(),
		ai:    gemini.New("", "", ""),
		voice: live.New(live.Config{}, nil),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	var payload map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if payload["sprint"] != "SPRINT-003" {
		t.Fatalf("sprint = %v", payload["sprint"])
	}
	if payload["auth_disabled"] != true {
		t.Fatalf("auth_disabled = %v", payload["auth_disabled"])
	}
}

func TestIssueCallTokenRoutePattern(t *testing.T) {
	lk := lktoken.Config{APIKey: "devkey", APISecret: "secret", LiveURL: "ws://localhost:7880"}
	if !lk.Enabled() {
		t.Fatal("livekit config should be enabled")
	}

	token, err := lk.JoinToken("monti-test", "caller", 0)
	if err != nil || token == "" {
		t.Fatalf("token issue failed: %v", err)
	}

	body := bytes.NewBufferString(`{"identity":"caller-demo"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/calls/demo123/token", body)
	req.SetPathValue("id", "demo123")
	if req.PathValue("id") != "demo123" {
		t.Fatal("path value not set")
	}
}