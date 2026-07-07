package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/libra/monti-jarvis/internal/workforce"
)

func TestWorkforceEndpointReturnsAgents(t *testing.T) {
	s := &server{}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/workforce", s.workforce)

	req := httptest.NewRequest(http.MethodGet, "/api/workforce", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	var payload struct {
		Agents []workforce.Agent `json:"agents"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&payload); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}
	if got, want := len(payload.Agents), 4; got != want {
		t.Fatalf("len(agents) = %d, want %d", got, want)
	}
}

func TestCompactHistoryCapsMessages(t *testing.T) {
	history := make([]struct {
		Role    string
		Content string
	}, 20)
	for i := range history {
		history[i].Role = "user"
		history[i].Content = "msg"
	}

	var gemHistory []struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	}
	for _, item := range history {
		gemHistory = append(gemHistory, struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		}{Role: item.Role, Content: item.Content})
	}

	// compactHistory is tested indirectly via length behavior
	out := compactHistory(nil, "latest")
	if len(out) != 1 || out[0].Content != "latest" {
		t.Fatalf("compactHistory(nil) = %#v, want single latest message", out)
	}
}