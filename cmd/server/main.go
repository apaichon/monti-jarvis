package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/libra/monti-jarvis/internal/env"
	"github.com/libra/monti-jarvis/internal/gemini"
	"github.com/libra/monti-jarvis/internal/live"
	"github.com/libra/monti-jarvis/internal/store"
	"github.com/libra/monti-jarvis/internal/web"
	"github.com/libra/monti-jarvis/internal/workforce"
)

type server struct {
	cfg    env.Config
	ai     *gemini.Client
	voice  *live.Relay
	store  *store.Store
	static http.Handler
}

type chatRequest struct {
	SessionID string           `json:"session_id"`
	AgentID   string           `json:"agent_id"`
	Topic     string           `json:"topic"`
	Message   string           `json:"message"`
	History   []gemini.Message `json:"history"`
}

type chatResponse struct {
	SessionID string `json:"session_id"`
	AgentID   string `json:"agent_id"`
	Reply     string `json:"reply"`
}

func main() {
	cfg := env.Load()
	rootCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	st, warnings := store.Open(rootCtx, cfg)
	for _, warning := range warnings {
		log.Printf("infra warning: %s", warning)
	}
	defer st.Close()

	s := &server{
		cfg:    cfg,
		ai:     gemini.New(cfg.GeminiAPIKey, cfg.GeminiModel),
		voice:  live.New(live.Config{APIKey: cfg.GeminiAPIKey, Model: cfg.GeminiLiveModel}),
		store:  st,
		static: web.Handler(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /api/infra", s.infra)
	mux.HandleFunc("GET /api/workforce", s.workforce)
	mux.HandleFunc("POST /api/chat", s.chat)
	mux.HandleFunc("GET /ws/voice", s.voice.Handler())
	mux.Handle("/", s.static)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           withCommonHeaders(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		health := st.Health(context.Background())
		log.Printf("monti-jarvis listening on :%s text_model=%s live_model=%s gemini=%t postgres=%s redis=%s minio=%s workforce=%d",
			cfg.Port, cfg.GeminiModel, cfg.GeminiLiveModel, s.ai.Enabled(), health.Postgres, health.Redis, health.Minio, len(workforce.All()))
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %v", err)
		}
	}()

	<-rootCtx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}

func (s *server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":        true,
		"gemini":    s.ai.Enabled(),
		"voice":     s.voice.Enabled(),
		"workforce": len(workforce.All()),
	})
}

func (s *server) infra(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	writeJSON(w, http.StatusOK, s.store.Health(ctx))
}

func (s *server) workforce(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"agents": workforce.All()})
}

func (s *server) chat(w http.ResponseWriter, r *http.Request) {
	var req chatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}
	req.Message = strings.TrimSpace(req.Message)
	if req.Message == "" {
		writeError(w, http.StatusBadRequest, "message is required")
		return
	}

	agent := workforce.Resolve(req.AgentID)
	sessionID := strings.TrimSpace(req.SessionID)
	if sessionID == "" {
		sessionID = newID()
	}

	message := req.Message
	if topic := strings.TrimSpace(req.Topic); topic != "" && !strings.EqualFold(topic, "general") {
		message = "[" + topic + "] " + message
	}

	history := compactHistory(req.History, message)
	ctx, cancel := context.WithTimeout(r.Context(), 50*time.Second)
	defer cancel()

	reply, err := s.ai.Reply(ctx, workforce.SystemPrompt(agent), history)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	s.store.SaveExchange(context.Background(), sessionID, agent.ID, req.Message, reply)
	writeJSON(w, http.StatusOK, chatResponse{
		SessionID: sessionID,
		AgentID:   agent.ID,
		Reply:     reply,
	})
}

func compactHistory(history []gemini.Message, message string) []gemini.Message {
	const maxMessages = 16
	out := make([]gemini.Message, 0, maxMessages+1)
	start := 0
	if len(history) > maxMessages {
		start = len(history) - maxMessages
	}
	for _, item := range history[start:] {
		role := strings.TrimSpace(item.Role)
		content := strings.TrimSpace(item.Content)
		if content == "" {
			continue
		}
		if role != "assistant" && role != "model" {
			role = "user"
		}
		out = append(out, gemini.Message{Role: role, Content: content})
	}
	out = append(out, gemini.Message{Role: "user", Content: message})
	return out
}

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return hex.EncodeToString([]byte(time.Now().UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(b[:])
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func withCommonHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}