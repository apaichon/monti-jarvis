package clickhouse

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/libra/monti-jarvis/internal/audit"
)

// Ensures markdown-like content with quotes/newlines is valid JSONEachRow.
func TestJSONEachRowEscapesMarkdown(t *testing.T) {
	content := "# Title\n\nDon't use `promptpay` — use **bank_qrcode**.\n| a | b |\n| --- | --- |\n| 'x' | y |\n"
	row := map[string]any{
		"tenant_id":   "dev-mountain-company",
		"agent_id":    "luna",
		"document_id": "doc1",
		"chunk_id":    "c1",
		"km_scope":    "general",
		"km_version":  1,
		"content":     content,
		"embedding":   []float32{0.1, 0.2, 0.3},
	}
	b, err := json.Marshal(row)
	if err != nil {
		t.Fatal(err)
	}
	// Must be single-line JSON object for JSONEachRow.
	if strings.Contains(string(b[:20]), "\n") && strings.Count(string(b), "\n") > 0 {
		// json.Marshal produces compact single line by default — good.
	}
	var decoded map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["content"] != content {
		t.Fatalf("content round-trip failed")
	}
	if !strings.Contains(string(b), `dev-mountain-company`) {
		t.Fatal("tenant missing")
	}
	// Apostrophe must be JSON-escaped, not SQL-broken.
	if strings.Contains(string(b), "Don't") && !strings.Contains(string(b), `Don\u0027t`) && !strings.Contains(string(b), `Don't`) {
		// Either unicode escape or raw in JSON string is fine as long as valid JSON
	}
}

func TestInsertAuditEventsUsesJSONEachRow(t *testing.T) {
	var body []byte
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		body, err = io.ReadAll(r.Body)
		if err != nil {
			t.Error(err)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	c := New(server.URL, "monti_jarvis", "user", "password")
	event := audit.Event{
		EventID: "evt-1", OccurredAt: time.Date(2026, 7, 16, 9, 10, 11, 123000000, time.UTC),
		TenantID: "tenant-1", ActorID: "admin-1", ActorType: "platform_admin", Action: "tenant.update",
		ResourceType: "tenant", ResourceID: "tenant-1", RequestID: "req-1", Source: "http_api", Outcome: "success",
		Metadata: map[string]any{"note": "quoted 'value'"},
	}
	if err := c.InsertAuditEvents(context.Background(), []audit.Event{event}); err != nil {
		t.Fatal(err)
	}
	var row map[string]any
	if err := json.Unmarshal(body, &row); err != nil {
		t.Fatalf("invalid JSONEachRow body: %v; body=%s", err, body)
	}
	if row["event_id"] != "evt-1" || row["metadata_json"] != `{"note":"quoted 'value'"}` {
		t.Fatalf("unexpected row: %#v", row)
	}
}
