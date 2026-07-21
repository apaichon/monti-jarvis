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
	Role             string            `json:"role"`
	Content          string            `json:"content,omitempty"`
	FunctionCall     *FunctionCall     `json:"function_call,omitempty"`
	FunctionResponse *FunctionResponse `json:"function_response,omitempty"`
}

// ToolDeclaration is the provider-neutral subset of a Gemini function
// declaration that the server exposes to the model.
type ToolDeclaration struct {
	Name        string         `json:"name"`
	Description string         `json:"description,omitempty"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type FunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args,omitempty"`
	ID   string         `json:"id,omitempty"`
}

type FunctionResponse struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response,omitempty"`
	ID       string         `json:"id,omitempty"`
}

type UsageMetadata struct {
	PromptTokenCount     uint64 `json:"promptTokenCount"`
	CandidatesTokenCount uint64 `json:"candidatesTokenCount"`
	TotalTokenCount      uint64 `json:"totalTokenCount"`
}

type ReplyResult struct {
	Text          string
	Usage         UsageMetadata
	Model         string
	FunctionCalls []FunctionCall
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
	result, err := c.ReplyWithUsage(ctx, systemPrompt, history)
	if err != nil {
		return "", err
	}
	return result.Text, nil
}

func (c *Client) ReplyWithUsage(ctx context.Context, systemPrompt string, history []Message) (ReplyResult, error) {
	return c.replyWithTools(ctx, systemPrompt, history, nil)
}

func (c *Client) ReplyWithTools(ctx context.Context, systemPrompt string, history []Message, tools []ToolDeclaration) (ReplyResult, error) {
	return c.replyWithTools(ctx, systemPrompt, history, tools)
}

func (c *Client) replyWithTools(ctx context.Context, systemPrompt string, history []Message, tools []ToolDeclaration) (ReplyResult, error) {
	if c.apiKey == "" {
		return ReplyResult{}, errors.New("GEMINI_API_KEY is not configured")
	}
	if c.model == "" {
		c.model = "gemini-flash-latest"
	}
	systemPrompt = strings.TrimSpace(systemPrompt)
	if systemPrompt == "" {
		return ReplyResult{}, errors.New("system prompt is required")
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
	if len(tools) > 0 {
		reqBody.Tools = []toolBundle{{FunctionDeclarations: tools}}
	}
	payload, err := json.Marshal(reqBody)
	if err != nil {
		return ReplyResult{}, err
	}

	endpoint := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		url.PathEscape(c.model),
		url.QueryEscape(c.apiKey),
	)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return ReplyResult{}, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return ReplyResult{}, err
	}
	defer resp.Body.Close()

	var out generateResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return ReplyResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if out.Error.Message != "" {
			return ReplyResult{}, fmt.Errorf("gemini: %s", out.Error.Message)
		}
		return ReplyResult{}, fmt.Errorf("gemini: http %d", resp.StatusCode)
	}
	for _, candidate := range out.Candidates {
		var chunks []string
		var calls []FunctionCall
		for _, p := range candidate.Content.Parts {
			if strings.TrimSpace(p.Text) != "" {
				chunks = append(chunks, p.Text)
			}
			if p.FunctionCall != nil && strings.TrimSpace(p.FunctionCall.Name) != "" {
				calls = append(calls, *p.FunctionCall)
			}
		}
		if len(calls) > 0 {
			return ReplyResult{
				Text:          strings.TrimSpace(strings.Join(chunks, "\n")),
				FunctionCalls: calls,
				Usage:         out.UsageMetadata,
				Model:         c.model,
			}, nil
		}
		if len(chunks) > 0 {
			return ReplyResult{
				Text:  strings.TrimSpace(strings.Join(chunks, "\n")),
				Usage: out.UsageMetadata,
				Model: c.model,
			}, nil
		}
	}
	return ReplyResult{}, errors.New("gemini returned no text")
}

func toGeminiContents(history []Message) []content {
	contents := make([]content, 0, len(history))
	for _, msg := range history {
		text := strings.TrimSpace(msg.Content)
		if text == "" && msg.FunctionCall == nil && msg.FunctionResponse == nil {
			continue
		}
		role := "user"
		if msg.Role == "assistant" || msg.Role == "model" {
			role = "model"
		}
		item := content{Role: role}
		if text != "" {
			item.Parts = append(item.Parts, part{Text: text})
		}
		if msg.FunctionCall != nil {
			item.Parts = append(item.Parts, part{FunctionCall: msg.FunctionCall})
		}
		if msg.FunctionResponse != nil {
			item.Parts = append(item.Parts, part{FunctionResponse: msg.FunctionResponse})
		}
		contents = append(contents, item)
	}
	return contents
}

type generateRequest struct {
	SystemInstruction content          `json:"system_instruction"`
	Contents          []content        `json:"contents"`
	GenerationConfig  generationConfig `json:"generationConfig"`
	Tools             []toolBundle     `json:"tools,omitempty"`
}

type toolBundle struct {
	FunctionDeclarations []ToolDeclaration `json:"function_declarations"`
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
	Text             string            `json:"text,omitempty"`
	FunctionCall     *FunctionCall     `json:"functionCall,omitempty"`
	FunctionResponse *FunctionResponse `json:"functionResponse,omitempty"`
}

type generateResponse struct {
	Candidates []struct {
		Content content `json:"content"`
	} `json:"candidates"`
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
	UsageMetadata UsageMetadata `json:"usageMetadata"`
}
