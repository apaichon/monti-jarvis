package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Client struct {
	apiKey     string
	model      string
	embedModel string
	http       *http.Client
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func New(apiKey, model, embedModel string) *Client {
	return &Client{
		apiKey:     apiKey,
		model:      model,
		embedModel: normalizeEmbedModel(embedModel),
		http:       &http.Client{Timeout: 45 * time.Second},
	}
}

func (c *Client) Enabled() bool {
	return c.apiKey != ""
}

func (c *Client) Reply(ctx context.Context, systemPrompt string, history []Message) (string, error) {
	if c.apiKey == "" {
		return "", errors.New("GEMINI_API_KEY is not configured")
	}
	if c.model == "" {
		c.model = "gemini-flash-latest"
	}
	systemPrompt = strings.TrimSpace(systemPrompt)
	if systemPrompt == "" {
		return "", errors.New("system prompt is required")
	}

	reqBody := generateRequest{
		SystemInstruction: content{
			Parts: []part{{Text: systemPrompt}},
		},
		Contents: toGeminiContents(history),
		GenerationConfig: generationConfig{
			Temperature:     0.7,
			TopP:            0.95,
			MaxOutputTokens: 900,
		},
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		url.PathEscape(c.model),
		url.QueryEscape(c.apiKey),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var out generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if out.Error.Message != "" {
			return "", fmt.Errorf("gemini: %s", out.Error.Message)
		}
		return "", fmt.Errorf("gemini: http %d", resp.StatusCode)
	}
	for _, candidate := range out.Candidates {
		var chunks []string
		for _, p := range candidate.Content.Parts {
			if strings.TrimSpace(p.Text) != "" {
				chunks = append(chunks, p.Text)
			}
		}
		if len(chunks) > 0 {
			return strings.TrimSpace(strings.Join(chunks, "\n")), nil
		}
	}
	return "", errors.New("gemini returned no text")
}

func toGeminiContents(history []Message) []content {
	contents := make([]content, 0, len(history))
	for _, msg := range history {
		text := strings.TrimSpace(msg.Content)
		if text == "" {
			continue
		}
		role := "user"
		if msg.Role == "assistant" || msg.Role == "model" {
			role = "model"
		}
		contents = append(contents, content{
			Role:  role,
			Parts: []part{{Text: text}},
		})
	}
	return contents
}

type generateRequest struct {
	SystemInstruction content          `json:"system_instruction"`
	Contents          []content        `json:"contents"`
	GenerationConfig  generationConfig `json:"generationConfig"`
}

type generationConfig struct {
	Temperature     float64 `json:"temperature"`
	TopP            float64 `json:"topP"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

type content struct {
	Role  string `json:"role,omitempty"`
	Parts []part `json:"parts"`
}

type part struct {
	Text string `json:"text"`
}

type generateResponse struct {
	Candidates []struct {
		Content content `json:"content"`
	} `json:"candidates"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}