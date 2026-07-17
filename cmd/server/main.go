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

	"github.com/libra/monti-jarvis/internal/audit"
	"github.com/libra/monti-jarvis/internal/auth"
	"github.com/libra/monti-jarvis/internal/calls"
	"github.com/libra/monti-jarvis/internal/clickhouse"
	"github.com/libra/monti-jarvis/internal/customerweb"
	"github.com/libra/monti-jarvis/internal/entitlements"
	"github.com/libra/monti-jarvis/internal/env"
	"github.com/libra/monti-jarvis/internal/gemini"
	"github.com/libra/monti-jarvis/internal/km"
	"github.com/libra/monti-jarvis/internal/live"
	"github.com/libra/monti-jarvis/internal/lktoken"
	"github.com/libra/monti-jarvis/internal/natsbus"
	"github.com/libra/monti-jarvis/internal/observability"
	"github.com/libra/monti-jarvis/internal/payment"
	"github.com/libra/monti-jarvis/internal/platformweb"
	"github.com/libra/monti-jarvis/internal/quota"
	"github.com/libra/monti-jarvis/internal/rag"
	"github.com/libra/monti-jarvis/internal/resend"
	"github.com/libra/monti-jarvis/internal/store"
	"github.com/libra/monti-jarvis/internal/tenantoauth"
	"github.com/libra/monti-jarvis/internal/tenantregister"
	"github.com/libra/monti-jarvis/internal/tenantweb"
	"github.com/libra/monti-jarvis/internal/web"
	"github.com/libra/monti-jarvis/internal/workforce"
)

type server struct {
	cfg             env.Config
	ai              *gemini.Client
	voice           *live.Relay
	calls           *calls.Service
	store           *store.Store
	km              *km.Service
	rag             *rag.Service
	ch              *clickhouse.Client
	bus             *natsbus.Bus
	auth            *auth.Service
	guard           *auth.HTTPGuard
	entitlements    *entitlements.Service
	quota           *quota.Service
	static          http.Handler
	platform        http.Handler
	tenant          http.Handler
	legacy          http.Handler
	registerLimiter *tenantregister.RateLimiter
	mailer          *resend.Client
	tenantOAuth     *tenantoauth.Service
	monitoring      *observability.Service
	audit           *audit.Writer
}

type chatRequest struct {
	SessionID string           `json:"session_id"`
	AgentID   string           `json:"agent_id"`
	Topic     string           `json:"topic"`
	Message   string           `json:"message"`
	History   []gemini.Message `json:"history"`
}

type chatResponse struct {
	SessionID   string       `json:"session_id"`
	AgentID     string       `json:"agent_id"`
	Reply       string       `json:"reply"`
	Sources     []rag.Source `json:"sources,omitempty"`
	MissingKM   bool         `json:"missing_km,omitempty"`
	TicketOffer *ticketOffer `json:"ticket_offer,omitempty"`
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
	auditWriter, err := audit.New(audit.Config{
		Mode: cfg.AuditLogMode, Dir: cfg.AuditLogDir, FlushInterval: cfg.AuditLogFlushInterval,
		Retention: cfg.AuditLogRetention, BatchSize: cfg.AuditLogBatchSize, QueueSize: cfg.AuditLogQueueSize,
		RetryBackoff: cfg.AuditLogRetryBackoff,
	}, ch)
	if err != nil {
		log.Printf("infra warning: audit log: %v", err)
	} else {
		auditWriter.Start(rootCtx)
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
	if mailer.Enabled() {
		log.Printf("mailer: resend enabled from=%s", cfg.ResendFromEmail)
	} else {
		log.Printf("mailer: resend disabled (set RESEND_API_KEY + RESEND_FROM_EMAIL or RESEND_FROM_ADDR with a verified domain)")
	}
	tenantOAuth := tenantoauth.New(st.Redis(), tenantoauth.Config{
		PublicBaseURL:      cfg.PublicBaseURL,
		GoogleClientID:     cfg.GoogleOAuthClientID,
		GoogleClientSecret: cfg.GoogleOAuthClientSecret,
		GoogleRedirectURL:  cfg.GoogleOAuthRedirectURL,
		GitHubClientID:     cfg.GitHubOAuthClientID,
		GitHubClientSecret: cfg.GitHubOAuthClientSecret,
		GitHubRedirectURL:  cfg.GitHubOAuthRedirectURL,
		RedisPrefix:        cfg.RedisPrefix,
	})

	entSvc := entitlements.New(st, cfg)
	quotaSvc := quota.New(entSvc, st, cfg)

	voiceRelay := live.New(live.Config{APIKey: cfg.GeminiAPIKey, Model: cfg.GeminiLiveModel, MobileMaxFrameBytes: cfg.MobileWSMaxFrameBytes}, ragSvc)
	// S16: tenant AI reply locale preference in voice system prompt.
	voiceRelay.LocaleHint = func(ctx context.Context, tenantID string) string {
		if st == nil {
			return ""
		}
		return st.AIReplyLocaleHint(ctx, tenantID)
	}

	s := &server{
		cfg:             cfg,
		ai:              ai,
		voice:           voiceRelay,
		calls:           calls.New(st, bus, lk, cfg.DemoTenantID),
		store:           st,
		km:              kmSvc,
		rag:             ragSvc,
		ch:              ch,
		bus:             bus,
		auth:            authSvc,
		guard:           guard,
		entitlements:    entSvc,
		quota:           quotaSvc,
		static:          customerweb.Handler(cfg.CustomerWebDir),
		platform:        platformweb.Handler(cfg.PlatformAdminWebDir),
		tenant:          tenantweb.Handler(cfg.TenantWebDir),
		legacy:          web.Handler(),
		registerLimiter: registerLimiter,
		mailer:          mailer,
		tenantOAuth:     tenantOAuth,
		monitoring:      newMonitoringService(st, ch, bus, ai, cfg),
		audit:           auditWriter,
	}
	s.backfillCallCenterAnalytics(rootCtx)
	voiceRelay.AgentResolver = s.resolveAssignedWorkforceAgent

	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", s.health)
	mux.HandleFunc("GET /api/infra", s.infra)
	mux.HandleFunc("GET /api/public/brands", s.publicBrands)
	mux.HandleFunc("GET /api/public/brands/{slug}", s.publicBrand)
	mux.Handle("PUT /api/tenant/brand", guard.RequireTenantAdminActive(http.HandlerFunc(s.putTenantBrand)))
	mux.Handle("PUT /api/platform/tenants/{tenant_id}/brand-listing", guard.RequirePlatformAdmin(http.HandlerFunc(s.putPlatformBrandListing)))
	mux.Handle("POST /api/calls", guard.OptionalBearer(http.HandlerFunc(s.createCall)))
	mux.Handle("GET /api/calls/{id}", guard.OptionalBearer(http.HandlerFunc(s.getCall)))
	mux.Handle("POST /api/calls/{id}/token", guard.OptionalBearer(http.HandlerFunc(s.issueCallToken)))
	mux.Handle("POST /api/calls/{id}/end", guard.OptionalBearer(http.HandlerFunc(s.endCall)))
	mux.Handle("POST /api/calls/{id}/audio", guard.OptionalBearer(http.HandlerFunc(s.archiveCallAudio)))
	mux.Handle("POST /api/calls/{id}/rating", guard.OptionalBearer(http.HandlerFunc(s.submitCallRating)))
	mux.Handle("GET /api/calls/{id}/turns", guard.OptionalBearer(http.HandlerFunc(s.listCallTurns)))
	mux.Handle("POST /api/calls/{id}/turns", guard.OptionalBearer(http.HandlerFunc(s.addCallTurn)))
	mux.Handle("GET /api/calls/{id}/events", guard.OptionalBearer(http.HandlerFunc(s.callEvents)))

	mux.HandleFunc("GET /api/workforce", s.workforce)
	mux.Handle("GET /api/customer/portal-policy", guard.OptionalBearer(http.HandlerFunc(s.customerPortalPolicy)))
	mux.Handle("GET /api/customer/workforce", guard.OptionalBearer(http.HandlerFunc(s.customerWorkforce)))
	mux.Handle("GET /api/customer/quota", guard.OptionalBearer(http.HandlerFunc(s.customerQuota)))
	mux.Handle("GET /api/mobile/v1/bootstrap", s.mobileAPI(guard.OptionalBearer(http.HandlerFunc(s.mobileBootstrap))))
	mux.Handle("POST /api/mobile/v1/calls", s.mobileAPI(guard.OptionalBearer(http.HandlerFunc(s.mobileCreateCall))))
	mux.Handle("GET /api/mobile/v1/calls/{call_id}", s.mobileAPI(guard.OptionalBearer(http.HandlerFunc(s.mobileGetCall))))
	mux.Handle("GET /api/mobile/v1/calls/{call_id}/transcript", s.mobileAPI(guard.OptionalBearer(http.HandlerFunc(s.mobileTranscript))))
	mux.Handle("POST /api/mobile/v1/calls/{call_id}/end", s.mobileAPI(guard.OptionalBearer(http.HandlerFunc(s.mobileEndCall))))
	mux.Handle("POST /api/mobile/v1/calls/{call_id}/rating", s.mobileAPI(guard.OptionalBearer(http.HandlerFunc(s.mobileRateCall))))
	mux.Handle("POST /api/chat", guard.OptionalBearer(http.HandlerFunc(s.chat)))
	mux.Handle("GET /api/km/agents/{agent_id}", guard.OptionalBearer(http.HandlerFunc(s.getAgentKnowledge)))
	mux.Handle("GET /api/km/agents/{agent_id}/documents", guard.OptionalBearer(http.HandlerFunc(s.listAgentDocuments)))
	mux.Handle("POST /api/km/agents/{agent_id}/documents", guard.RequireKMWrite(http.HandlerFunc(s.uploadAgentDocument)))
	mux.Handle("POST /api/km/agents/{agent_id}/reset", guard.RequireKMWrite(http.HandlerFunc(s.resetAgentKnowledge)))
	mux.Handle("POST /api/km/seed", guard.RequirePlatformAdmin(http.HandlerFunc(s.seedKnowledge)))
	mux.HandleFunc("POST /api/public/tenant/register", s.registerTenant)
	mux.HandleFunc("GET /api/public/tenant/verify-email", s.verifyTenantEmail)
	mux.HandleFunc("POST /api/public/tenant/verify-email", s.verifyTenantEmail)
	// Shared tenant OAuth (login + register): one start/callback per provider.
	mux.HandleFunc("GET /api/public/tenant/oauth/providers", s.oauthProviders)
	mux.HandleFunc("GET /api/public/tenant/oauth/{provider}", s.startTenantOAuth)
	mux.HandleFunc("GET /api/public/tenant/oauth/{provider}/callback", s.tenantOAuthCallback)
	mux.HandleFunc("POST /api/public/tenant/oauth/complete", s.completeTenantOAuth)
	// Legacy paths (pre-rename register/oauth) — keep callbacks working until consoles updated.
	mux.HandleFunc("GET /api/public/tenant/register/oauth/providers", s.oauthProviders)
	mux.HandleFunc("GET /api/public/tenant/register/oauth/{provider}", s.startTenantOAuth)
	mux.HandleFunc("GET /api/public/tenant/register/oauth/{provider}/callback", s.tenantOAuthCallback)
	mux.HandleFunc("POST /api/public/tenant/register/oauth/complete", s.completeTenantOAuth)
	mux.Handle("GET /api/platform/tenants", guard.RequirePlatformAdmin(http.HandlerFunc(s.listPlatformTenants)))
	mux.Handle("GET /api/platform/audit-logs", guard.RequirePlatformAdmin(http.HandlerFunc(s.listPlatformAuditLogs)))
	mux.Handle("GET /api/platform/audit-logs/health", guard.RequirePlatformAdmin(http.HandlerFunc(s.platformAuditHealth)))
	mux.Handle("GET /api/platform/system-performance", guard.RequirePlatformAdmin(http.HandlerFunc(s.getPlatformSystemPerformance)))
	mux.Handle("GET /api/platform/call-center/statistics", guard.RequirePlatformAdmin(http.HandlerFunc(s.getPlatformCallCenterStatistics)))
	mux.Handle("GET /api/platform/tenants/{tenant_id}/kyc", guard.RequirePlatformAdmin(http.HandlerFunc(s.getPlatformTenantKYC)))
	mux.Handle("POST /api/platform/tenants/{tenant_id}/kyc/approve", guard.RequirePlatformAdmin(http.HandlerFunc(s.approvePlatformTenantKYC)))
	mux.Handle("POST /api/platform/tenants/{tenant_id}/kyc/reject", guard.RequirePlatformAdmin(http.HandlerFunc(s.rejectPlatformTenantKYC)))
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
	mux.Handle("GET /api/platform/payment-gateway", guard.RequirePlatformAdmin(http.HandlerFunc(s.getPaymentGateway)))
	mux.Handle("PUT /api/platform/payment-gateway", guard.RequirePlatformAdmin(http.HandlerFunc(s.putPaymentGateway)))
	mux.Handle("POST /api/platform/payment-gateway/test", guard.RequirePlatformAdmin(http.HandlerFunc(s.testPaymentGateway)))
	// Sprint 10–11: platform billing ledger + receipt ops
	mux.Handle("GET /api/platform/billing/orders", guard.RequirePlatformAdmin(http.HandlerFunc(s.listPlatformBillingOrders)))
	mux.Handle("GET /api/platform/billing/orders/{id}", guard.RequirePlatformAdmin(http.HandlerFunc(s.getPlatformBillingOrder)))
	mux.Handle("GET /api/platform/billing/documents", guard.RequirePlatformAdmin(http.HandlerFunc(s.listPlatformBillingDocuments)))
	mux.Handle("GET /api/platform/billing/documents/{id}", guard.RequirePlatformAdmin(http.HandlerFunc(s.getPlatformBillingDocument)))
	mux.Handle("POST /api/platform/billing/documents/{id}/void", guard.RequirePlatformAdmin(http.HandlerFunc(s.voidPlatformBillingDocument)))
	mux.Handle("POST /api/platform/billing/documents/{id}/reissue", guard.RequirePlatformAdmin(http.HandlerFunc(s.reissuePlatformBillingDocument)))
	mux.Handle("GET /api/platform/billing/seller-branding", guard.RequirePlatformAdmin(http.HandlerFunc(s.getSellerBranding)))
	mux.Handle("PUT /api/platform/billing/seller-branding", guard.RequirePlatformAdmin(http.HandlerFunc(s.putSellerBranding)))
	mux.HandleFunc("POST /api/callbacks/chillpay", s.chillpayCallback)
	// Browser return from ChillPay (GET or POST) → fulfill if paid → SPA status page.
	// Path may include /{orderRef} because ChillPay often strips query strings.
	mux.HandleFunc("GET /api/callbacks/chillpay/return", s.chillpayBrowserReturn)
	mux.HandleFunc("POST /api/callbacks/chillpay/return", s.chillpayBrowserReturn)
	mux.HandleFunc("GET /api/callbacks/chillpay/return/{ref}", s.chillpayBrowserReturn)
	mux.HandleFunc("POST /api/callbacks/chillpay/return/{ref}", s.chillpayBrowserReturn)
	// SPRINT-014 embed
	mux.HandleFunc("GET /api/public/embed/{embed_key}", s.getPublicEmbed)
	mux.HandleFunc("OPTIONS /api/public/embed/{embed_key}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, X-Tenant-Id, X-Embed-Parent-Origin")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusNoContent)
	})
	mux.Handle("GET /api/tenant/embed", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantEmbed)))
	mux.Handle("PUT /api/tenant/embed", guard.RequireTenantAdminActive(http.HandlerFunc(s.putTenantEmbed)))
	mux.Handle("POST /api/tenant/embed/rotate-key", guard.RequireTenantAdminActive(http.HandlerFunc(s.rotateTenantEmbedKey)))
	// SPRINT-015 — tenant KM + knowledge gaps
	mux.Handle("GET /api/tenant/km/scopes", guard.RequireTenantAdminActive(http.HandlerFunc(s.listTenantKMScopes)))
	mux.Handle("GET /api/tenant/km/agents", guard.RequireTenantAdminActive(http.HandlerFunc(s.listTenantKMAgents)))
	mux.Handle("GET /api/tenant/km/agents/{agent_id}/documents", guard.RequireTenantAdminActive(http.HandlerFunc(s.listTenantKMDocuments)))
	mux.Handle("POST /api/tenant/km/agents/{agent_id}/documents", guard.RequireTenantAdminActive(http.HandlerFunc(s.uploadTenantKMDocument)))
	mux.Handle("POST /api/tenant/km/agents/{agent_id}/reset", guard.RequireTenantAdminActive(http.HandlerFunc(s.resetTenantKMAgent)))
	mux.Handle("PATCH /api/tenant/km/documents/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.patchTenantKMDocument)))
	mux.Handle("DELETE /api/tenant/km/documents/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.deleteTenantKMDocument)))
	mux.Handle("GET /api/tenant/km/gaps", guard.RequireTenantAdminActive(http.HandlerFunc(s.listTenantKMGaps)))
	mux.Handle("PATCH /api/tenant/km/gaps/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.patchTenantKMGap)))

	// SPRINT-016 — tenant settings, usage, call limits
	mux.Handle("GET /api/tenant/settings", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantSettings)))
	mux.Handle("PUT /api/tenant/settings", guard.RequireTenantAdminActive(http.HandlerFunc(s.putTenantSettings)))
	mux.Handle("GET /api/tenant/usage", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantUsage)))
	mux.Handle("GET /api/tenant/call-limits", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantCallLimits)))
	mux.Handle("PUT /api/tenant/call-limits", guard.RequireTenantAdminActive(http.HandlerFunc(s.putTenantCallLimits)))

	// SPRINT-017 — tenant test / preview sandbox
	mux.Handle("GET /api/tenant/preview/scenarios", guard.RequireTenantAdminActive(http.HandlerFunc(s.listPreviewScenarios)))
	mux.Handle("POST /api/tenant/preview/chat", guard.RequireTenantAdminActive(http.HandlerFunc(s.tenantPreviewChat)))
	mux.HandleFunc("GET /ws/tenant/preview/voice", s.tenantPreviewVoiceWS)

	// SPRINT-018 — customer tiers + groups
	mux.Handle("GET /api/tenant/tiers", guard.RequireTenantAdminActive(http.HandlerFunc(s.listTenantTiers)))
	mux.Handle("POST /api/tenant/tiers", guard.RequireTenantAdminActive(http.HandlerFunc(s.createTenantTier)))
	mux.Handle("GET /api/tenant/tiers/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantTier)))
	mux.Handle("PUT /api/tenant/tiers/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.putTenantTier)))
	mux.Handle("DELETE /api/tenant/tiers/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.deleteTenantTier)))
	mux.Handle("GET /api/tenant/groups", guard.RequireTenantAdminActive(http.HandlerFunc(s.listTenantGroups)))
	mux.Handle("POST /api/tenant/groups", guard.RequireTenantAdminActive(http.HandlerFunc(s.createTenantGroup)))
	mux.Handle("GET /api/tenant/groups/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantGroup)))
	mux.Handle("PUT /api/tenant/groups/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.putTenantGroup)))
	mux.Handle("DELETE /api/tenant/groups/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.deleteTenantGroup)))

	// SPRINT-019 — tenant customer directory, CSV imports, domain rules
	mux.Handle("GET /api/tenant/customers", guard.RequireTenantAdminActive(http.HandlerFunc(s.listTenantCustomers)))
	mux.Handle("POST /api/tenant/customers", guard.RequireTenantAdminActive(http.HandlerFunc(s.createTenantCustomer)))
	mux.Handle("GET /api/tenant/customers/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantCustomer)))
	mux.Handle("PUT /api/tenant/customers/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.putTenantCustomer)))
	mux.Handle("DELETE /api/tenant/customers/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.deleteTenantCustomer)))
	mux.Handle("POST /api/tenant/customer-imports", guard.RequireTenantAdminActive(http.HandlerFunc(s.importTenantCustomers)))
	mux.Handle("GET /api/tenant/customer-imports/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantCustomerImport)))
	mux.Handle("GET /api/tenant/customer-domain-rules", guard.RequireTenantAdminActive(http.HandlerFunc(s.listTenantCustomerDomainRules)))
	mux.Handle("POST /api/tenant/customer-domain-rules", guard.RequireTenantAdminActive(http.HandlerFunc(s.createTenantCustomerDomainRule)))
	mux.Handle("PUT /api/tenant/customer-domain-rules/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.putTenantCustomerDomainRule)))
	mux.Handle("DELETE /api/tenant/customer-domain-rules/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.deleteTenantCustomerDomainRule)))

	// SPRINT-020 — customer email OTP auth.
	mux.Handle("GET /api/tenant/customer-auth/settings", guard.RequireTenantAdminActive(http.HandlerFunc(s.getCustomerAuthSettings)))
	mux.Handle("PUT /api/tenant/customer-auth/settings", guard.RequireTenantAdminActive(http.HandlerFunc(s.putCustomerAuthSettings)))
	mux.Handle("GET /api/tenant/conversation-records", guard.RequireTenantAdminActive(http.HandlerFunc(s.listTenantConversationRecords)))
	mux.Handle("GET /api/tenant/conversation-records/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantConversationRecord)))
	mux.Handle("GET /api/tenant/conversation-records/{id}/archive-objects/{object_id}/content", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantConversationArchiveObjectContent)))
	mux.Handle("POST /api/tenant/conversation-records/{id}/archive/retry", guard.RequireTenantAdminActive(http.HandlerFunc(s.retryTenantConversationArchive)))
	mux.Handle("GET /api/tenant/knowledge-gaps", guard.RequireTenantAdminActive(http.HandlerFunc(s.listTenantKnowledgeGaps)))
	mux.Handle("PATCH /api/tenant/knowledge-gaps/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.patchTenantKnowledgeGap)))
	// SPRINT-023 — customer-confirmed human escalation and tenant ticket queue.
	mux.Handle("POST /api/customer/tickets", guard.OptionalBearer(http.HandlerFunc(s.createCustomerTicket)))
	mux.Handle("GET /api/tenant/tickets", guard.RequireTenantAdminActive(http.HandlerFunc(s.listTenantTickets)))
	mux.Handle("GET /api/tenant/tickets/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantTicket)))
	mux.Handle("PATCH /api/tenant/tickets/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.patchTenantTicket)))
	mux.Handle("POST /api/tenant/tickets/{id}/events", guard.RequireTenantAdminActive(http.HandlerFunc(s.addTenantTicketEvent)))
	// SPRINT-024 — tenant-scoped customer satisfaction statistics.
	mux.Handle("GET /api/tenant/satisfaction/statistics", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantSatisfactionStatistics)))
	// SPRINT-025 — tenant-scoped call-center usage and quota statistics.
	mux.Handle("GET /api/tenant/call-center/statistics", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantCallCenterStatistics)))
	mux.Handle("GET /api/tenant/system-performance", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantSystemPerformance)))
	mux.HandleFunc("POST /api/customer/auth/request-otp", s.requestCustomerOTP)
	mux.HandleFunc("POST /api/customer/auth/verify-otp", s.verifyCustomerOTP)
	mux.HandleFunc("POST /api/customer/auth/refresh", s.refreshCustomerAuth)
	mux.HandleFunc("POST /api/customer/auth/logout", s.logoutCustomerAuth)
	mux.HandleFunc("GET /api/customer/me", s.customerMe)

	mux.Handle("GET /api/tenant/packages", guard.RequireTenantAdminActive(http.HandlerFunc(s.listTenantPackages)))
	mux.Handle("POST /api/tenant/checkout", guard.RequireTenantAdminActive(http.HandlerFunc(s.tenantCheckout)))
	mux.Handle("GET /api/tenant/orders/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantOrder)))
	mux.Handle("GET /api/tenant/orders/{id}/documents/{doc_type}", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantOrderDocument)))
	// Sprint 12: tax profile + document vault
	mux.Handle("GET /api/tenant/tax-profile", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantTaxProfile)))
	mux.Handle("PUT /api/tenant/tax-profile", guard.RequireTenantAdminActive(http.HandlerFunc(s.putTenantTaxProfile)))
	mux.Handle("GET /api/tenant/billing/documents", guard.RequireTenantAdminActive(http.HandlerFunc(s.listTenantBillingDocuments)))
	mux.Handle("GET /api/tenant/billing/documents/{id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.getTenantBillingDocument)))
	mux.Handle("POST /api/dev/mock-pay/{order_id}", guard.RequireTenantAdminActive(http.HandlerFunc(s.mockPayOrder)))
	mux.Handle("POST /api/platform/avatars/{id}/image", guard.RequirePlatformAdmin(http.HandlerFunc(s.uploadAvatarImage)))
	mux.HandleFunc("GET /api/assets/avatars/{id}/{file}", s.serveAvatarAsset)
	mux.Handle("GET /api/entitlements/me", guard.RequireTenantAdminOrPlatform(http.HandlerFunc(s.entitlementMe)))
	mux.Handle("GET /api/platform/tenants/{tenant_id}/usage", guard.RequirePlatformAdmin(http.HandlerFunc(s.getPlatformTenantUsage)))
	mux.HandleFunc("GET /ws/voice", s.voiceWS)
	mux.Handle("GET /ws/mobile/v1/calls/{call_id}", s.mobileAPI(http.HandlerFunc(s.mobileVoiceWS)))
	mux.Handle("/api/", http.HandlerFunc(s.apiJSONNotFound))
	mux.Handle("/ws/", http.HandlerFunc(s.apiJSONNotFound))
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

	var handler http.Handler = mux
	if auditWriter != nil {
		handler = auditWriter.Middleware(mux, s.auditActor, cfg.DemoTenantID)
	}
	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           withCORS(cfg.AuthDisabled, withCommonHeaders(handler)),
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
	if auditWriter != nil {
		_ = auditWriter.Close(shutdownCtx)
	}
}

func (s *server) apiJSONNotFound(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusNotFound, map[string]any{"error": "route not found", "code": "not_found", "path": r.URL.Path})
}

func (s *server) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"ok":                 true,
		"gemini":             s.ai != nil && s.ai.Enabled(),
		"voice":              s.voice != nil && s.voice.Enabled(),
		"livekit":            s.cfg.LiveKitAPIKey != "",
		"nats":               s.bus != nil && s.bus.Enabled(),
		"legacy_ui":          s.cfg.LegacyUIEnabled,
		"sprint":             "SPRINT-020",
		"auth_disabled":      s.cfg.AuthDisabled,
		"tenant_register":    s.cfg.TenantRegisterEnabled,
		"rag":                s.rag != nil && s.rag.Enabled(),
		"customer_web":       s.cfg.CustomerWebDir,
		"platform_admin_web": s.cfg.PlatformAdminWebDir,
		"tenant_web":         s.cfg.TenantWebDir,
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
		"postgres":   h.Postgres,
		"redis":      h.Redis,
		"minio":      h.Minio,
		"clickhouse": h.ClickHouse,
		"nats":       h.NATS,
		"livekit":    h.LiveKit,
	}
	if s.auth != nil && s.auth.Enabled() {
		out["auth_cache"] = s.auth.CacheStatus()
		out["auth_events"] = s.auth.EventsStatus()
		out["auth_write_behind_lag"] = s.auth.WriteBehindLag(ctx)
	}
	if s.entitlements != nil {
		out["entitlement_cache"] = s.entitlements.CacheStatus()
	}
	if s.quota != nil {
		out["quota"] = s.quota.Status(ctx)
		out["rate_limit"] = s.quota.RateLimitStatus(ctx)
	} else {
		out["quota"] = "disabled"
		out["rate_limit"] = "disabled"
	}
	if s.store != nil {
		if row, err := s.store.GetPaymentGatewayConfig(ctx); err == nil {
			gw := payment.NewGateway(s.cfg, s.store)
			resolved := gw.Resolve(row)
			out["payment_gateway"] = map[string]any{
				"configured": strings.TrimSpace(resolved.Provider) != "" && resolved.Status == "active",
				"provider":   resolved.Provider,
				"mode":       resolved.Mode,
				"status":     resolved.Status,
			}
		}
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

func (s *server) resolveAssignedWorkforceAgent(ctx context.Context, tenantID, agentID string) (workforce.Agent, bool) {
	if s.store == nil || strings.TrimSpace(tenantID) == "" {
		return workforce.Agent{}, false
	}
	agents, err := s.store.ListWorkforceAgents(ctx, tenantID)
	if err != nil {
		log.Printf("resolve workforce tenant=%s agent=%s: %v", tenantID, agentID, err)
		return workforce.Agent{}, false
	}
	return workforce.FindAssigned(agentID, agents)
}

func (s *server) resolveWorkforceAgent(ctx context.Context, tenantID, agentID string) workforce.Agent {
	if agent, ok := s.resolveAssignedWorkforceAgent(ctx, tenantID, agentID); ok {
		return agent
	}
	return workforce.Resolve(agentID)
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

	tenantID := s.quotaTenant(r)
	agent := s.resolveWorkforceAgent(r.Context(), tenantID, req.AgentID)
	customer, _, ok := s.enforceCustomerPortalAccess(w, r, tenantID, agent.ID, "chat")
	if !ok {
		return
	}
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

	if s.quota != nil {
		if err := s.quota.AllowRate(ctx, tenantID, quota.BucketChat); err != nil {
			writeQuotaError(w, err)
			return
		}
	}

	topic := strings.TrimSpace(req.Topic)
	var ragResult rag.Result
	useRAG := true
	if s.quota != nil {
		if err := s.quota.CheckFeature(ctx, tenantID, quota.DimRAGEnabled); err != nil {
			// Soft skip RAG when package disables it (chat still works).
			useRAG = false
		}
	}
	// Always scope RAG to the request tenant (embed/chat X-Tenant-Id) — not DemoTenantID.
	ragSvc := s.rag
	if ragSvc != nil && tenantID != "" {
		ragSvc = ragSvc.WithTenant(tenantID)
	}
	if useRAG && ragSvc != nil {
		var err error
		ragResult, err = ragSvc.Retrieve(ctx, agent.ID, topic, req.Message)
		if err != nil {
			log.Printf("chat rag tenant=%s agent=%s: %v", tenantID, agent.ID, err)
		}
	}
	prompt := workforce.SystemPrompt(agent)
	if s.store != nil && tenantID != "" {
		if hint := s.store.AIReplyLocaleHint(ctx, tenantID); hint != "" {
			prompt += "\n\n" + hint
		}
	}
	if useRAG && ragSvc != nil {
		prompt = ragSvc.AugmentPrompt(prompt, agent.ID, topic, req.Message, ragResult)
	}

	reply, err := s.ai.Reply(ctx, prompt, history)
	if err != nil {
		writeError(w, http.StatusBadGateway, err.Error())
		return
	}

	s.store.SaveExchange(context.Background(), sessionID, agent.ID, req.Message, reply)
	var conversationID string
	if s.store != nil {
		customerID := ""
		if customer != nil {
			customerID = customer.ID
			_ = s.store.RecordCustomerUsage(r.Context(), tenantID, customer.ID, sessionID, agent.ID, "chat", 1, "committed", "")
		}
		endedAt := time.Now().UTC()
		rec, err := s.store.UpsertConversationRecord(r.Context(), store.ConversationRecordInput{
			TenantID: tenantID, CallID: sessionID, CustomerID: customerID, AvatarID: agent.ID,
			Channel: "chat", Status: "archived", EndedAt: &endedAt, Summary: map[string]any{"topic": topic, "message_count": len(history)},
		})
		if err == nil {
			conversationID = rec.ID
			_, _ = s.store.ArchiveConversationTranscript(r.Context(), tenantID, sessionID, map[string]any{
				"session_id": sessionID, "tenant_id": tenantID, "agent_id": agent.ID, "topic": topic,
				"question": req.Message, "reply": reply, "sources": ragResult.Sources,
			}, "")
			s.projectCallCenterRecord(r.Context(), tenantID, rec.ID)
		}
	}
	if ragResult.MissingKM && s.store != nil {
		// Persist FAQ backlog for tenant KM improvement (also logged to ClickHouse qa_events).
		src := store.KMGapSourceChat
		// Embed UI uses same chat API; source still "chat" unless client sends later.
		_, _ = s.store.RecordKMGap(context.Background(), store.KMGap{
			TenantID:  tenantID,
			AgentID:   agent.ID,
			Topic:     topic,
			Question:  req.Message,
			SessionID: sessionID,
			Source:    src,
		})
		customerID := ""
		if customer != nil {
			customerID = customer.ID
		}
		_, _ = s.store.CreateKnowledgeGapCandidate(r.Context(), store.KnowledgeGapInput{
			TenantID: tenantID, ConversationRecordID: conversationID, AvatarID: agent.ID, CustomerID: customerID,
			Question: req.Message, AnswerExcerpt: reply, GapReason: "no_source", Confidence: 0,
		})
	}
	writeJSON(w, http.StatusOK, chatResponse{
		SessionID:   sessionID,
		AgentID:     agent.ID,
		Reply:       reply,
		Sources:     ragResult.Sources,
		MissingKM:   ragResult.MissingKM,
		TicketOffer: ticketOfferForMessage(req.Message, topic),
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
		// Embed UI is meant to run inside tenant iframes (SPRINT-014).
		if strings.HasPrefix(r.URL.Path, "/embed") {
			w.Header().Set("Content-Security-Policy", "frame-ancestors *")
			// Avoid legacy DENY that would block third-party embedding.
			w.Header().Del("X-Frame-Options")
		}
		next.ServeHTTP(w, r)
	})
}

func withCORS(_ bool, next http.Handler) http.Handler {
	// Include embed parent-origin header so host-page preflight from Vue/etc. succeeds.
	allowHeaders := "Content-Type, X-Tenant-Id, Authorization, X-Embed-Parent-Origin"
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
