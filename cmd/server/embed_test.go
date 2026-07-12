package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/libra/monti-jarvis/internal/store"
)

func TestWriteEmbedError_NotFound(t *testing.T) {
	rr := httptest.NewRecorder()
	writeEmbedError(rr, store.ErrEmbedNotFound)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("status %d", rr.Code)
	}
	var body map[string]any
	_ = json.NewDecoder(rr.Body).Decode(&body)
	if body["code"] != "embed_not_found" {
		t.Fatalf("%#v", body)
	}
}

func TestWriteEmbedError_InvalidOrigin(t *testing.T) {
	rr := httptest.NewRecorder()
	writeEmbedError(rr, store.ValidateOrigin("bad"))
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("status %d body %s", rr.Code, rr.Body.String())
	}
}

func TestEmbedConfigJSON_Snippet(t *testing.T) {
	cfg := &store.TenantEmbedConfig{
		TenantID: "demo",
		EmbedKey: "emb_test",
		Enabled:  true,
	}
	out := embedConfigJSON(cfg, "http://localhost:8091/")
	snip, _ := out["snippet"].(string)
	if snip == "" || !containsAll(snip, "monti-embed.js", "emb_test") {
		t.Fatalf("snippet %q", snip)
	}
}

func containsAll(s string, parts ...string) bool {
	for _, p := range parts {
		if !jsonContains(s, p) {
			return false
		}
	}
	return true
}

func jsonContains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 ||
		(func() bool {
			for i := 0; i+len(sub) <= len(s); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
			return false
		})())
}
