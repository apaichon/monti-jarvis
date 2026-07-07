package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/libra/monti-jarvis/internal/env"
)

func TestWorkforceEndpointWhenLegacyEnabled(t *testing.T) {
	cfg := env.Load()
	cfg.LegacyUIEnabled = true
	s := &server{cfg: cfg}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/workforce", s.workforce)

	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/workforce", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload struct {
		Agents []struct {
			ID string `json:"id"`
		} `json:"agents"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if got, want := len(payload.Agents), 4; got != want {
		t.Fatalf("len(agents) = %d, want %d", got, want)
	}
}

func TestCompactHistoryCapsMessages(t *testing.T) {
	out := compactHistory(nil, "latest")
	if len(out) != 1 || out[0].Content != "latest" {
		t.Fatalf("compactHistory(nil) = %#v, want single latest message", out)
	}
}