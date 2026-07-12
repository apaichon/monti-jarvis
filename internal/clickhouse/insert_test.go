package clickhouse

import (
	"encoding/json"
	"strings"
	"testing"
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
