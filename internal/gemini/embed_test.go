package gemini

import "testing"

func TestNormalizeEmbedModel(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"", defaultEmbedModel},
		{"text-embedding-004", defaultEmbedModel},
		{"models/text-embedding-004", defaultEmbedModel},
		{"gemini-embedding-001", "gemini-embedding-001"},
		{"gemini-embedding-2", "gemini-embedding-2"},
	}
	for _, tc := range tests {
		if got := normalizeEmbedModel(tc.in); got != tc.want {
			t.Fatalf("normalizeEmbedModel(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFormatEmbedTextGeminiEmbedding2(t *testing.T) {
	doc := formatEmbedText("gemini-embedding-2", "hello", embedTaskDocument)
	if doc != "title: none | text: hello" {
		t.Fatalf("document format: %q", doc)
	}
	query := formatEmbedText("gemini-embedding-2", "hello", embedTaskQuery)
	if query != "task: search result | query: hello" {
		t.Fatalf("query format: %q", query)
	}
}

func TestFormatEmbedTextGeminiEmbedding001(t *testing.T) {
	text := formatEmbedText("gemini-embedding-001", "hello", embedTaskQuery)
	if text != "hello" {
		t.Fatalf("expected raw text, got %q", text)
	}
}

func TestNewSetsEmbedModel(t *testing.T) {
	c := New("key", "chat", "text-embedding-004")
	if c.embedModel != defaultEmbedModel {
		t.Fatalf("embedModel = %q, want %q", c.embedModel, defaultEmbedModel)
	}
}