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
	"github.com/libra/monti-jarvis/internal/workforce"
)

const liveURL = "wss://generativelanguage.googleapis.com/ws/google.ai.generativelanguage.v1beta.GenerativeService.BidiGenerateContent"

type Config struct {
	APIKey string
	Model  string
}

type Relay struct {
	cfg Config
}

func New(cfg Config) *Relay {
	return &Relay{cfg: cfg}
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
		client, err := upgrader.Upgrade(w, req, nil)
		if err != nil {
			log.Printf("voice upgrade: %v", err)
			return
		}
		defer client.Close()

		ctx, cancel := context.WithCancel(req.Context())
		defer cancel()

		gem, err := r.dial(ctx, agent)
		if err != nil {
			log.Printf("gemini live dial: %v", err)
			_ = client.WriteJSON(serverMsg{Type: "error", Message: "Gemini Live connection failed"})
			return
		}
		defer gem.Close()

		var clientWrite sync.Mutex
		send := func(msg serverMsg) error {
			clientWrite.Lock()
			defer clientWrite.Unlock()
			return client.WriteJSON(msg)
		}
		_ = send(serverMsg{
			Type:      "ready",
			Model:     normalizeModel(r.cfg.Model),
			Voice:     agent.Voice,
			AgentID:   agent.ID,
			AgentName: agent.Name,
			StartedAt: time.Now().UnixMilli(),
		})

		var gemWrite sync.Mutex
		writeGem := func(value any) error {
			gemWrite.Lock()
			defer gemWrite.Unlock()
			return gem.WriteJSON(value)
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

func (r *Relay) dial(ctx context.Context, agent workforce.Agent) (*websocket.Conn, error) {
	endpoint := liveURL + "?key=" + url.QueryEscape(r.cfg.APIKey)
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, endpoint, nil)
	if err != nil {
		return nil, err
	}
	if err := conn.WriteJSON(r.setup(agent)); err != nil {
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
	return conn, nil
}

func (r *Relay) setup(agent workforce.Agent) map[string]any {
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
			"parts": []map[string]any{{"text": workforce.SystemPrompt(agent)}},
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