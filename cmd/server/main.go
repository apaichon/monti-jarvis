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

	"github.com/libra/monti-jarvis/internal/calls"
	"github.com/libra/monti-jarvis/internal/customerweb"
	"github.com/libra/monti-jarvis/internal/env"
	"github.com/libra/monti-jarvis/internal/gemini"
	"github.com/libra/monti-jarvis/internal/live"
	"github.com/libra/monti-jarvis/internal/lktoken"
	"github.com/libra/monti-jarvis/internal/natsbus"
	"github.com/libra/monti-jarvis/internal/store"
	"github.com/libra/monti-jarvis/internal/web"
	"github.com/libra/monti-jarvis/internal/workforce"
)

type server struct {
	cfg    env.Config
	ai     *gemini.Client
	voice  *live.Relay
	calls  *calls.Service
	store  *store.Store
	bus    *natsbus.Bus
	static http.Handler
	legacy http.Handler
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

	var bus *natsbus.Bus
	if cfg.NATSURL != "" {
		nc, err := natsbus.Connect(cfg.NATSURL)
		if err != nil {
			log.Printf("infra warning: nats: %v", err)
		} else {
			bus = nc
			defer bus.Close()
		}
	}

	lk := lktoken.Config{
		APIKey:    cfg.LiveKitAPIKey,
		APISecret: cfg.LiveKitAPISecret,
		LiveURL:   cfg.LiveKitURL,
	}

	s := &server{
		cfg:    cfg,
		ai:     gemini.New(cfg.GeminiAPIKey, cfg.GeminiModel),
		voice:  live.New(live.Config{APIKey: cfg.GeminiAPIKey, Model: cfg.GeminiLiveModel}),
		calls:  calls.New(st, bus, lk, cfg.DemoTenantID),
		store:  st,
		bus:    bus,
		static: customerweb.Handler(cfg.CustomerWebDir),
		legacy: web.Handler(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /api/infra", s.infra)
	mux.HandleFunc("POST /api/calls", s.createCall)
	mux.HandleFunc("GET /api/calls/{id}", s.getCall)
	mux.HandleFunc("POST /api/calls/{id}/token", s.issueCallToken)
	mux.HandleFunc("POST /api/calls/{id}/end", s.endCall)
	mux.HandleFunc("GET /api/calls/{id}/turns", s.listCallTurns)
	mux.HandleFunc("POST /api/calls/{id}/turns", s.addCallTurn)
	mux.HandleFunc("GET /api/calls/{id}/events", s.callEvents)

	mux.HandleFunc("GET /api/workforce", s.workforce)
	mux.HandleFunc("POST /api/chat", s.chat)
	mux.HandleFunc("GET /ws/voice", s.voice.Handler())
	mux.HandleFunc("GET /legacy", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/legacy/", http.StatusFound)
	})
	mux.Handle("/legacy/", http.StripPrefix("/legacy", s.legacy))

	mux.Handle("/", s.static)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           withCORS(withCommonHeaders(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		health := st.Health(context.Background())
		log.Printf("monti-jarvis listening on :%s customer_web=%s livekit=%t nats=%t postgres=%s redis=%s legacy_ui=%t",
			cfg.Port, cfg.CustomerWebDir, lk.Enabled(), bus != nil && bus.Enabled(),
			health.Postgres, health.Redis, cfg.LegacyUIEnabled)
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
		"ok":           true,
		"gemini":       s.ai != nil && s.ai.Enabled(),
		"voice":        s.voice != nil && s.voice.Enabled(),
		"livekit":      s.cfg.LiveKitAPIKey != "",
		"nats":         s.bus != nil && s.bus.Enabled(),
		"legacy_ui":    s.cfg.LegacyUIEnabled,
		"sprint":       "SPRINT-001",
		"customer_web": s.cfg.CustomerWebDir,
	})
}

func (s *server) infra(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	h := s.store.Health(ctx)
	h.NATS = "disabled"
	if s.bus != nil && s.bus.Enabled() {
		h.NATS = "ok"
	}
	h.LiveKit = "disabled"
	if s.cfg.LiveKitAPIKey != "" {
		h.LiveKit = "configured"
	}
	writeJSON(w, http.StatusOK, h)
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

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}