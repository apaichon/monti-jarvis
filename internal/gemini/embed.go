package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

const defaultEmbedModel = "gemini-embedding-001"

type embedTask string

const (
	embedTaskDocument embedTask = "document"
	embedTaskQuery    embedTask = "query"
)

func normalizeEmbedModel(model string) string {
	model = strings.TrimSpace(strings.TrimPrefix(model, "models/"))
	switch model {
	case "", "text-embedding-004", "embedding-001":
		return defaultEmbedModel
	default:
		return model
	}
}

func (c *Client) embedModelName() string {
	if c.embedModel != "" {
		return c.embedModel
	}
	return defaultEmbedModel
}

func (c *Client) EmbedDocument(ctx context.Context, text string) ([]float32, error) {
	return c.embed(ctx, text, embedTaskDocument)
}

func (c *Client) EmbedQuery(ctx context.Context, text string) ([]float32, error) {
	return c.embed(ctx, text, embedTaskQuery)
}

func (c *Client) embed(ctx context.Context, text string, task embedTask) ([]float32, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is not configured")
	}
	text = strings.TrimSpace(text)
	if text == "" {
		return nil, fmt.Errorf("embed text is required")
	}

	model := c.embedModelName()
	payload := map[string]any{
		"model":   "models/" + model,
		"content": map[string]any{"parts": []map[string]string{{"text": formatEmbedText(model, text, task)}}},
	}
	if usesEmbedTaskType(model) {
		payload["taskType"] = embedTaskType(task)
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:embedContent?key=%s",
		url.PathEscape(model),
		url.QueryEscape(c.apiKey),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var out struct {
		Embedding struct {
			Values []float32 `json:"values"`
		} `json:"embedding"`
		Error struct {
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if out.Error.Message != "" {
			return nil, fmt.Errorf("gemini embed: %s", out.Error.Message)
		}
		return nil, fmt.Errorf("gemini embed: http %d", resp.StatusCode)
	}
	if len(out.Embedding.Values) == 0 {
		return nil, fmt.Errorf("gemini embed: empty vector")
	}
	return out.Embedding.Values, nil
}

func usesEmbedTaskType(model string) bool {
	return strings.HasPrefix(model, "gemini-embedding-001")
}

func embedTaskType(task embedTask) string {
	if task == embedTaskQuery {
		return "RETRIEVAL_QUERY"
	}
	return "RETRIEVAL_DOCUMENT"
}

func formatEmbedText(model, text string, task embedTask) string {
	if !strings.HasPrefix(model, "gemini-embedding-2") {
		return text
	}
	if task == embedTaskQuery {
		return "task: search result | query: " + text
	}
	return "title: none | text: " + text
}