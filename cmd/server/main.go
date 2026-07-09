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

	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/calls"
	"github.com/libra/monti-jarvis/internal/clickhouse"
	"github.com/libra/monti-jarvis/internal/customerweb"
	"github.com/libra/monti-jarvis/internal/entitlements"
	"github.com/libra/monti-jarvis/internal/env"
	"github.com/libra/monti-jarvis/internal/gemini"
	"github.com/libra/monti-jarvis/internal/platformweb"
	"github.com/libra/monti-jarvis/internal/resend"
	"github.com/libra/monti-jarvis/internal/tenantregister"
	"github.com/libra/monti-jarvis/internal/tenantoauth"
	"github.com/libra/monti-jarvis/internal/tenantweb"
	"github.com/libra/monti-jarvis/internal/km"
	"github.com/libra/monti-jarvis/internal/live"
	"github.com/libra/monti-jarvis/internal/lktoken"
	"github.com/libra/monti-jarvis/internal/natsbus"
	"github.com/libra/monti-jarvis/internal/rag"
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
	km     *km.Service
	rag    *rag.Service
	ch     *clickhouse.Client
	bus    *natsbus.Bus
	auth         *auth.Service
	guard        *auth.HTTPGuard
	entitlements *entitlements.Service
	static          http.Handler
	platform        http.Handler
	tenant          http.Handler
	legacy          http.Handler
	registerLimiter *tenantregister.RateLimiter
	mailer          *resend.Client
	tenantOAuth     *tenantoauth.Service
}

type chatRequest struct {
	SessionID string           `json:"session_id"`
	AgentID   string           `json:"agent_id"`
	Topic     string           `json:"topic"`
	Message   string           `json:"message"`
	History   []gemini.Message `json:"history"`
}

type chatResponse struct {
	SessionID string       `json:"session_id"`
	AgentID   string       `json:"agent_id"`
	Reply     string       `json:"reply"`
	Sources   []rag.Source `json:"sources,omitempty"`
	MissingKM bool         `json:"missing_km,omitempty"`
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

	ch := clickhouse.New(cfg.ClickHouseURL, cfg.ClickHouseDB, cfg.ClickHouseUser, cfg.ClickHousePassword)
	if ch != nil && ch.Enabled() {
		if err := ch.EnsureSchema(rootCtx); err != nil {
			log.Printf("infra warning: clickhouse schema: %v", err)
		} else if err := ch.Ping(rootCtx); err != nil {
			log.Printf("infra warning: clickhouse ping: %v", err)
		}
	}

	ai := gemini.New(cfg.GeminiAPIKey, cfg.GeminiModel, cfg.GeminiEmbedModel)
	ragSvc := rag.New(ch, ai, cfg.DemoTenantID)
	kmSvc := km.NewService(st, ch, ai, cfg.DemoTenantID)

	var authSvc *auth.Service
	if cfg.JWTSecret != "" {
		var err error
		authSvc, err = auth.NewService(auth.Dependencies{
			Store: st,
			Bus:   bus,
			CH:    ch,
			Cfg: auth.Config{
				JWTSecret:          cfg.JWTSecret,
				AccessTTL:          cfg.JWTAccessTTL,
				RefreshTTL:         cfg.JWTRefreshTTL,
				UserCacheTTL:       cfg.AuthUserCacheTTL,
				AuthDisabled:       cfg.AuthDisabled,
				CacheEnabled:       cfg.AuthCacheEnabled,
				WriteBehindEnabled: cfg.AuthWriteBehindEnabled,
				EventsEnabled:      cfg.AuthEventsEnabled,
				RedisPrefix:        cfg.RedisPrefix,
			},
		})
		if err != nil {
			log.Printf("infra warning: auth: %v", err)
		} else {
			authSvc.Start(rootCtx)
		}
	}
	if bus != nil && bus.Enabled() && cfg.AuthEventsEnabled {
		if err := bus.EnsureAuthStream(rootCtx); err != nil {
			log.Printf("infra warning: nats auth stream: %v", err)
		}
	}
	if ch != nil && ch.Enabled() {
		if err := ch.EnsureAuthEventsSchema(rootCtx); err != nil {
			log.Printf("infra warning: clickhouse auth_events: %v", err)
		}
	}
	guard := auth.NewHTTPGuard(authSvc, st, cfg.AuthDisabled)
	registerLimiter := tenantregister.NewRateLimiter(st.Redis(), cfg.RedisPrefix, cfg.TenantRegisterRateLimit)
	mailer := resend.New(cfg.ResendAPIKey, cfg.ResendFromEmail)
	tenantOAuth := tenantoauth.New(st.Redis(), tenantoauth.Config{
		PublicBaseURL:      cfg.PublicBaseURL,
		GoogleClientID:     cfg.GoogleOAuthClientID,
		GoogleClientSecret: cfg.GoogleOAuthClientSecret,
		GitHubClientID:     cfg.GitHubOAuthClientID,
		GitHubClientSecret: cfg.GitHubOAuthClientSecret,
		RedisPrefix:        cfg.RedisPrefix,
	})

	s := &server{
		cfg:          cfg,
		ai:           ai,
		voice:        live.New(live.Config{APIKey: cfg.GeminiAPIKey, Model: cfg.GeminiLiveModel}, ragSvc),
		calls:        calls.New(st, bus, lk, cfg.DemoTenantID),
		store:        st,
		km:           kmSvc,
		rag:          ragSvc,
		ch:           ch,
		bus:          bus,
		auth:         authSvc,
		guard:        guard,
		entitlements: entitlements.New(st, cfg),
		static:          customerweb.Handler(cfg.CustomerWebDir),
		platform:        platformweb.Handler(cfg.PlatformAdminWebDir),
		tenant:          tenantweb.Handler(cfg.TenantWebDir),
		legacy:          web.Handler(),
		registerLimiter: registerLimiter,
		mailer:          mailer,
		tenantOAuth:     tenantOAuth,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /api/infra", s.infra)
	mux.Handle("POST /api/calls", guard.OptionalBearer(http.HandlerFunc(s.createCall)))
	mux.Handle("GET /api/calls/{id}", guard.OptionalBearer(http.HandlerFunc(s.getCall)))
	mux.Handle("POST /api/calls/{id}/token", guard.OptionalBearer(http.HandlerFunc(s.issueCallToken)))
	mux.Handle("POST /api/calls/{id}/end", guard.OptionalBearer(http.HandlerFunc(s.endCall)))
	mux.Handle("GET /api/calls/{id}/turns", guard.OptionalBearer(http.HandlerFunc(s.listCallTurns)))
	mux.Handle("POST /api/calls/{id}/turns", guard.OptionalBearer(http.HandlerFunc(s.addCallTurn)))
	mux.Handle("GET /api/calls/{id}/events", guard.OptionalBearer(http.HandlerFunc(s.callEvents)))

	mux.HandleFunc("GET /api/workforce", s.workforce)
	mux.HandleFunc("POST /api/chat", s.chat)
	mux.Handle("GET /api/km/agents/{agent_id}", guard.OptionalBearer(http.HandlerFunc(s.getAgentKnowledge)))
	mux.Handle("GET /api/km/agents/{agent_id}/documents", guard.OptionalBearer(http.HandlerFunc(s.listAgentDocuments)))
	mux.Handle("POST /api/km/agents/{agent_id}/documents", guard.RequireKMWrite(http.HandlerFunc(s.uploadAgentDocument)))
	mux.Handle("POST /api/km/agents/{agent_id}/reset", guard.RequireKMWrite(http.HandlerFunc(s.resetAgentKnowledge)))
	mux.Handle("POST /api/km/seed", guard.RequirePlatformAdmin(http.HandlerFunc(s.seedKnowledge)))
	mux.HandleFunc("POST /api/public/tenant/register", s.registerTenant)
	mux.HandleFunc("GET /api/public/tenant/verify-email", s.verifyTenantEmail)
	mux.HandleFunc("POST /api/public/tenant/verify-email", s.verifyTenantEmail)
	mux.HandleFunc("GET /api/public/tenant/register/oauth/providers", s.oauthProviders)
	mux.HandleFunc("GET /api/public/tenant/register/oauth/{provider}", s.startTenantOAuth)
	mux.HandleFunc("GET /api/public/tenant/register/oauth/{provider}/callback", s.tenantOAuthCallback)
	mux.HandleFunc("POST /api/public/tenant/register/oauth/complete", s.completeTenantOAuth)
	mux.Handle("GET /api/platform/tenants", guard.RequirePlatformAdmin(http.HandlerFunc(s.listPlatformTenants)))
	mux.HandleFunc("GET /api/tenant/kyc", s.getTenantKYC)
	mux.HandleFunc("PUT /api/tenant/kyc", s.updateTenantKYC)
	mux.HandleFunc("POST /api/tenant/kyc/photo", s.uploadTenantKYCPhoto)
	mux.HandleFunc("POST /api/tenant/kyc/documents", s.uploadTenantKYCDocument)
	mux.HandleFunc("POST /api/tenant/kyc/submit", s.submitTenantKYC)
	mux.HandleFunc("GET /api/assets/kyc/{tenant_id}/{kind}/{file}", s.serveKYCAsset)
	mux.HandleFunc("POST /api/auth/login", s.login)
	mux.HandleFunc("POST /api/auth/refresh", s.refreshToken)
	mux.HandleFunc("POST /api/auth/logout", s.logout)
	mux.Handle("GET /api/auth/me", guard.RequireBearer(http.HandlerFunc(s.me)))
	mux.Handle("GET /api/platform/rule-schemas", guard.RequirePlatformAdmin(http.HandlerFunc(s.listRuleSchemas)))
	mux.Handle("GET /api/platform/packages", guard.RequirePlatformAdmin(http.HandlerFunc(s.listPackages)))
	mux.Handle("POST /api/platform/packages", guard.RequirePlatformAdmin(http.HandlerFunc(s.createPackage)))
	mux.Handle("GET /api/platform/packages/{id}", guard.RequirePlatformAdmin(http.HandlerFunc(s.getPackage)))
	mux.Handle("PUT /api/platform/packages/{id}", guard.RequirePlatformAdmin(http.HandlerFunc(s.updatePackage)))
	mux.Handle("DELETE /api/platform/packages/{id}", guard.RequirePlatformAdmin(http.HandlerFunc(s.archivePackage)))
	mux.Handle("GET /api/platform/tenants/{tenant_id}/entitlement", guard.RequirePlatformAdmin(http.HandlerFunc(s.getTenantEntitlement)))
	mux.Handle("POST /api/platform/tenants/{tenant_id}/entitlement", guard.RequirePlatformAdmin(http.HandlerFunc(s.assignTenantEntitlement)))
	mux.Handle("DELETE /api/platform/tenants/{tenant_id}/entitlement", guard.RequirePlatformAdmin(http.HandlerFunc(s.revokeTenantEntitlement)))
	mux.Handle("GET /api/platform/avatars", guard.RequirePlatformAdmin(http.HandlerFunc(s.listAvatars)))
	mux.Handle("POST /api/platform/avatars", guard.RequirePlatformAdmin(http.HandlerFunc(s.createAvatar)))
	mux.Handle("GET /api/platform/avatars/{id}", guard.RequirePlatformAdmin(http.HandlerFunc(s.getAvatar)))
	mux.Handle("PUT /api/platform/avatars/{id}", guard.RequirePlatformAdmin(http.HandlerFunc(s.updateAvatar)))
	mux.Handle("DELETE /api/platform/avatars/{id}", guard.RequirePlatformAdmin(http.HandlerFunc(s.archiveAvatar)))
	mux.Handle("GET /api/platform/tenants/{tenant_id}/avatars", guard.RequirePlatformAdmin(http.HandlerFunc(s.listTenantAvatars)))
	mux.Handle("POST /api/platform/tenants/{tenant_id}/avatars", guard.RequirePlatformAdmin(http.HandlerFunc(s.assignTenantAvatar)))
	mux.Handle("DELETE /api/platform/tenants/{tenant_id}/avatars/{avatar_id}", guard.RequirePlatformAdmin(http.HandlerFunc(s.revokeTenantAvatar)))
	mux.Handle("POST /api/platform/avatars/{id}/image", guard.RequirePlatformAdmin(http.HandlerFunc(s.uploadAvatarImage)))
	mux.HandleFunc("GET /api/assets/avatars/{id}/{file}", s.serveAvatarAsset)
	mux.Handle("GET /api/entitlements/me", guard.RequireTenantAdminOrPlatform(http.HandlerFunc(s.entitlementMe)))
	mux.HandleFunc("GET /ws/voice", s.voice.Handler())
	mux.HandleFunc("GET /legacy", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/legacy/", http.StatusFound)
	})
	mux.Handle("/legacy/", http.StripPrefix("/legacy", s.legacy))
	mux.HandleFunc("GET /admin", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/admin/", http.StatusFound)
	})
	mux.Handle("/admin/", s.platform)
	mux.HandleFunc("GET /tenant", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/tenant/", http.StatusFound)
	})
	mux.Handle("/tenant/", s.tenant)

	mux.Handle("/", s.static)

	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           withCORS(cfg.AuthDisabled, withCommonHeaders(mux)),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		health := st.Health(context.Background())
		log.Printf("monti-jarvis listening on :%s customer_web=%s platform_admin=%s tenant_web=%s auth_disabled=%t livekit=%t nats=%t postgres=%s redis=%s legacy_ui=%t",
			cfg.Port, cfg.CustomerWebDir, cfg.PlatformAdminWebDir, cfg.TenantWebDir, cfg.AuthDisabled, lk.Enabled(), bus != nil && bus.Enabled(),
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
		"sprint":              "SPRINT-006",
		"auth_disabled":       s.cfg.AuthDisabled,
		"tenant_register":     s.cfg.TenantRegisterEnabled,
		"rag":                 s.rag != nil && s.rag.Enabled(),
		"customer_web":        s.cfg.CustomerWebDir,
		"platform_admin_web":  s.cfg.PlatformAdminWebDir,
		"tenant_web":          s.cfg.TenantWebDir,
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
	h.ClickHouse = "disabled"
	if s.ch != nil && s.ch.Enabled() {
		if err := s.ch.Ping(ctx); err != nil {
			h.ClickHouse = err.Error()
		} else {
			h.ClickHouse = "ok"
		}
	}
	out := map[string]any{
		"postgres":    h.Postgres,
		"redis":       h.Redis,
		"minio":       h.Minio,
		"clickhouse":  h.ClickHouse,
		"nats":        h.NATS,
		"livekit":     h.LiveKit,
	}
	if s.auth != nil && s.auth.Enabled() {
		out["auth_cache"] = s.auth.CacheStatus()
		out["auth_events"] = s.auth.EventsStatus()
		out["auth_write_behind_lag"] = s.auth.WriteBehindLag(ctx)
	}
	if s.entitlements != nil {
		out["entitlement_cache"] = s.entitlements.CacheStatus()
	}
	writeJSON(w, http.StatusOK, out)
}

func (s *server) workforce(w http.ResponseWriter, r *http.Request) {
	tenantID := auth.ResolveTenant(r.Context(), r.Header.Get("X-Tenant-Id"), s.cfg.AuthDisabled, s.cfg.DemoTenantID)
	if s.store != nil && s.store.HasTenantAvatarAssignments(r.Context(), tenantID) {
		agents, err := s.store.ListWorkforceAgents(r.Context(), tenantID)
		if err != nil {
			writeError(w, http.StatusBadGateway, err.Error())
			return
		}
		out := make([]workforce.Agent, 0, len(agents))
		for _, agent := range agents {
			out = append(out, workforce.FromWorkforceAgent(agent))
		}
		writeJSON(w, http.StatusOK, map[string]any{"agents": out})
		return
	}
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

	topic := strings.TrimSpace(req.Topic)
	ragResult, _ := s.rag.Retrieve(ctx, agent.ID, topic, req.Message)
	prompt := s.rag.AugmentPrompt(workforce.SystemPrompt(agent), agent.ID, topic, req.Message, ragResult)

	reply, err := s.ai.Reply(ctx, prompt, history)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	s.store.SaveExchange(context.Background(), sessionID, agent.ID, req.Message, reply)
	writeJSON(w, http.StatusOK, chatResponse{
		SessionID: sessionID,
		AgentID:   agent.ID,
		Reply:     reply,
		Sources:   ragResult.Sources,
		MissingKM: ragResult.MissingKM,
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

func withCORS(_ bool, next http.Handler) http.Handler {
	allowHeaders := "Content-Type, X-Tenant-Id, Authorization"
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", allowHeaders)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}