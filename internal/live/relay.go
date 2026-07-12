package live

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/libra/monti-jarvis/internal/rag"
	"github.com/libra/monti-jarvis/internal/workforce"
)

const (
	liveURL          = "wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1beta.GenerativeService.BidiGenerateContent"
	// Allow enough time to load tenant KM chunks into the Live setup prompt.
	voiceRAGTimeout = 4 * time.Second
)

type Config struct {
	APIKey string
	Model  string
}

// LocaleHintFunc returns an optional system-prompt suffix (e.g. preferred reply language).
type LocaleHintFunc func(ctx context.Context, tenantID string) string

type Relay struct {
	cfg        Config
	rag        *rag.Service
	LocaleHint LocaleHintFunc
}

func New(cfg Config, ragSvc *rag.Service) *Relay {
	return &Relay{cfg: cfg, rag: ragSvc}
}

func (r *Relay) Enabled() bool {
	return r.cfg.APIKey != ""
}

func (r *Relay) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if r.cfg.APIKey == "" {
			http.Error(w, "GEMINI_API_KEY is not configured", http.StatusServiceUnavailable)
			return
		}

		agent := workforce.Resolve(req.URL.Query().Get("agent"))
		topic := strings.TrimSpace(req.URL.Query().Get("topic"))
		tenantID := strings.TrimSpace(req.URL.Query().Get("tenant_id"))
		lang := normalizeLang(req.URL.Query().Get("lang"))
		client, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Printf("voice upgrade: %v", err)
			return
		}
		defer client.Close()

		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()

		var clientWrite sync.Mutex
		send := func(msg serverMsg) error {
			clientWrite.Lock()
			defer clientWrite.Unlock()
			return client.WriteJSON(msg)
		}
		// Early status so UI can show loading while Gemini connects (often several seconds).
		_ = send(serverMsg{Type: "status", Message: "Connecting to AI voice…"})

		// Scope voice RAG to the same tenant as chat/embed (query tenant_id).
		relay := r
		if tenantID != "" && r.rag != nil {
			cp := *r
			cp.rag = r.rag.WithTenant(tenantID)
			relay = &cp
		}
		gem, err := relay.dial(ctx, agent, topic, tenantID, lang, func(msg string) {
			_ = send(serverMsg{Type: "status", Message: msg})
		})
		if err != nil {
			log.Printf("gemini live dial: %v", err)
			_ = send(serverMsg{Type: "error", Message: "Gemini Live connection failed — try again"})
			return
		}
		defer gem.Close()

		_ = send(serverMsg{
			Type:      "ready",
			Model:     normalizeModel(r.cfg.Model),
			Voice:     agent.Voice,
			AgentID:   agent.ID,
			AgentName: agent.Name,
			StartedAt: time.Now().UnixMilli(),
			Message:   "Connected — agent is greeting you…",
		})

		var gemWrite sync.Mutex
		writeGem := func(value any) error {
			gemWrite.Lock()
			defer gemWrite.Unlock()
			return gem.WriteJSON(value)
		}

		// Speak first: trigger opening greeting (do not wait for caller).
		if err := writeGem(openingGreetingContent(agent, lang)); err != nil {
			log.Printf("voice greeting trigger: %v", err)
		} else {
			_ = send(serverMsg{Type: "status", Message: "Agent greeting…"})
		}

		go func() {
			defer cancel()
			for {
				_, raw, err := gem.ReadMessage()
				if err != nil {
					if ctx.Err() == nil {
						log.Printf("gemini live read: %v", err)
					}
					return
				}
				var frame geminiFrame
				if err := json.Unmarshal(raw, &frame); err != nil {
					continue
				}
				if frame.ServerContent == nil {
					continue
				}
				sc := frame.ServerContent
				if sc.InputTranscription != nil && strings.TrimSpace(sc.InputTranscription.Text) != "" {
					_ = send(serverMsg{Type: "transcript", Role: "user", Text: sc.InputTranscription.Text})
				}
				if sc.OutputTranscription != nil && strings.TrimSpace(sc.OutputTranscription.Text) != "" {
					_ = send(serverMsg{Type: "transcript", Role: "assistant", Text: sc.OutputTranscription.Text})
				}
				if sc.Interrupted {
					_ = send(serverMsg{Type: "interrupted"})
				}
				if sc.ModelTurn != nil {
					for _, part := range sc.ModelTurn.Parts {
						if part.InlineData != nil && part.InlineData.Data != "" {
							_ = send(serverMsg{Type: "audio", Data: part.InlineData.Data})
						}
						if strings.TrimSpace(part.Text) != "" {
							_ = send(serverMsg{Type: "text", Text: part.Text})
						}
					}
				}
				if sc.TurnComplete || sc.GenerationComplete {
					_ = send(serverMsg{Type: "turn_complete"})
				}
			}
		}()

		for {
			if ctx.Err() != nil {
				return
			}
			_, raw, err := client.ReadMessage()
			if err != nil {
				return
			}
			var msg clientMsg
			if json.Unmarshal(raw, &msg) != nil {
				continue
			}
			switch msg.Type {
			case "audio":
				if msg.Data == "" {
					continue
				}
				err = writeGem(map[string]any{"realtimeInput": map[string]any{
					"audio": map[string]any{"mimeType": "audio/pcm;rate=16000", "data": msg.Data},
				}})
			case "text":
				text := strings.TrimSpace(msg.Text)
				if text == "" {
					continue
				}
				err = writeGem(map[string]any{"clientContent": map[string]any{
					"turns":        []map[string]any{{"role": "user", "parts": []map[string]any{{"text": text}}}},
					"turnComplete": true,
				}})
			case "end":
				return
			}
			if err != nil {
				log.Printf("gemini live write: %v", err)
				return
			}
		}
	}
}

func (r *Relay) dial(ctx context.Context, agent workforce.Agent, topic, tenantID, lang string, status func(string)) (*websocket.Conn, error) {
	start := time.Now()
	if status != nil {
		status("Preparing agent…")
	}
	prompt := workforce.SystemPrompt(agent)
	// Session language: explicit query lang wins; else tenant settings hint.
	if instr := languageInstruction(lang); instr != "" {
		prompt += "\n\n" + instr
	} else if r.LocaleHint != nil && tenantID != "" {
		if hint := r.LocaleHint(ctx, tenantID); hint != "" {
			prompt += "\n\n" + hint
		}
	}
	prompt += "\n\nWhen a call connects, speak first with a short greeting — do not wait for the caller to speak."

	type ragOut struct {
		result rag.Result
		err    error
	}
	ragCh := make(chan ragOut, 1)
	go func() {
		if r.rag == nil || !r.rag.Enabled() {
			ragCh <- ragOut{}
			return
		}
		ragCtx, cancel := context.WithTimeout(ctx, voiceRAGTimeout)
		defer cancel()
		result, err := r.rag.RetrieveForVoice(ragCtx, agent.ID, topic)
		ragCh <- ragOut{result: result, err: err}
	}()

	if status != nil {
		status("Opening Gemini Live…")
	}
	endpoint := liveURL + "?key=" + url.QueryEscape(r.cfg.APIKey)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}

	select {
	case out := <-ragCh:
		if out.err != nil {
			log.Printf("voice rag preload: %v (using base prompt)", out.err)
		} else if r.rag != nil {
			prompt = r.rag.BuildVoicePrompt(prompt, agent.ID, topic, out.result)
		}
	case <-time.After(voiceRAGTimeout):
		log.Printf("voice rag preload: timeout after %s (using base prompt)", voiceRAGTimeout)
	}

	if status != nil {
		status("Configuring session…")
	}
	if err := conn.WriteJSON(r.setupMessage(agent, prompt)); err != nil {
		_ = conn.Close()
		return nil, err
	}
	var ack map[string]json.RawMessage
	if err := conn.ReadJSON(&ack); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if _, ok := ack["setupComplete"]; !ok {
		_ = conn.Close()
		return nil, &setupError{Ack: ack}
	}
	log.Printf("voice dial agent=%s topic=%s lang=%s rag+setup=%s", agent.ID, topic, lang, time.Since(start).Round(time.Millisecond))
	return conn, nil
}

func normalizeLang(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "th", "en", "auto":
		return v
	default:
		return ""
	}
}

func languageInstruction(lang string) string {
	switch lang {
	case "th":
		return "Speak and reply in Thai (ภาษาไทย) for this call unless the caller clearly switches language. You may use English for brand/product names."
	case "en":
		return "Speak and reply in English for this call unless the caller clearly switches language."
	case "auto":
		return "Detect the caller's language and reply in that language. You may switch languages if the caller switches. Prefer one language per turn."
	default:
		return ""
	}
}

// openingGreetingContent asks the model to speak the agent greeting immediately.
func openingGreetingContent(agent workforce.Agent, lang string) map[string]any {
	greet := strings.TrimSpace(agent.Greeting)
	if greet == "" {
		greet = "Hello, thank you for calling. How can I help you today?"
	}
	langHint := ""
	switch lang {
	case "th":
		langHint = " Speak the greeting in Thai."
		// Prefer a Thai-flavored nudge even if catalog greeting is English.
		greet = greet + " (Deliver this meaning warmly in Thai.)"
	case "en":
		langHint = " Speak the greeting in English."
	case "auto":
		langHint = " Prefer Thai if the tenant is Thai-facing; otherwise English. Keep one language."
	}
	text := "SYSTEM: The voice call just connected. Speak FIRST now — do not wait for the caller. " +
		"Deliver a short spoken opening greeting as " + agent.Name + " (" + agent.Role + "). " +
		"Base it on: «" + greet + "»." + langHint +
		" One short turn only, then listen."
	return map[string]any{"clientContent": map[string]any{
		"turns": []map[string]any{{
			"role":  "user",
			"parts": []map[string]any{{"text": text}},
		}},
		"turnComplete": true,
	}}
}

func (r *Relay) setupMessage(agent workforce.Agent, prompt string) map[string]any {
	voice := strings.TrimSpace(agent.Voice)
	if voice == "" {
		voice = "Aoede"
	}
	return map[string]any{"setup": map[string]any{
		"model": "models/" + normalizeModel(r.cfg.Model),
		"generationConfig": map[string]any{
			"responseModalities": []string{"AUDIO"},
			"speechConfig": map[string]any{
				"voiceConfig": map[string]any{
					"prebuiltVoiceConfig": map[string]any{"voiceName": voice},
				},
			},
		},
		"systemInstruction": map[string]any{
			"parts": []map[string]any{{"text": prompt}},
		},
		"inputAudioTranscription":  map[string]any{},
		"outputAudioTranscription": map[string]any{},
	}}
}

func normalizeModel(model string) string {
	model = strings.TrimSpace(model)
	model = strings.TrimPrefix(model, "models/")
	if model == "" {
		return "gemini-2.5-flash-native-audio-latest"
	}
	return model
}

type setupError struct {
	Ack map[string]json.RawMessage
}

func (e *setupError) Error() string {
	return "gemini live did not return setupComplete"
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type clientMsg struct {
	Type string `json:"type"`
	Data string `json:"data,omitempty"`
	Text string `json:"text,omitempty"`
}

type serverMsg struct {
	Type      string `json:"type"`
	Data      string `json:"data,omitempty"`
	Text      string `json:"text,omitempty"`
	Message   string `json:"message,omitempty"`
	Role      string `json:"role,omitempty"`
	Model     string `json:"model,omitempty"`
	Voice     string `json:"voice,omitempty"`
	AgentID   string `json:"agent_id,omitempty"`
	AgentName string `json:"agent_name,omitempty"`
	StartedAt int64  `json:"started_at_ms,omitempty"`
}

type geminiFrame struct {
	ServerContent *struct {
		ModelTurn *struct {
			Parts []struct {
				Text       string `json:"text,omitempty"`
				InlineData *struct {
					MimeType string `json:"mimeType"`
					Data     string `json:"data"`
				} `json:"inlineData,omitempty"`
			} `json:"parts,omitempty"`
		} `json:"modelTurn,omitempty"`
		InputTranscription *struct {
			Text string `json:"text"`
		} `json:"inputTranscription,omitempty"`
		OutputTranscription *struct {
			Text string `json:"text"`
		} `json:"outputTranscription,omitempty"`
		TurnComplete       bool `json:"turnComplete,omitempty"`
		GenerationComplete bool `json:"generationComplete,omitempty"`
		Interrupted        bool `json:"interrupted,omitempty"`
	} `json:"serverContent,omitempty"`
}