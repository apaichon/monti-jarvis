package env

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	GeminiAPIKey    string
	GeminiModel     string
	GeminiLiveModel string
	// AllowPlatformGeminiFallback enables env GEMINI_API_KEY for tenants without
	// a usable key. Ignored when AppEnv is production/prod (always fail-closed).
	AllowPlatformGeminiFallback bool
	Voice                       string
	PostgresURL                 string
	// PostgresKMReadURL is the least-privilege connection used by KM/RAG read
	// paths. It must not be used for KM ingest or ticket mutation.
	PostgresKMReadURL string
	// PostgresTicketWriteURL is the dedicated ticket capability connection.
	PostgresTicketWriteURL   string
	PostgresRLSEnforced      bool
	PostgresSchema           string
	RedisURL                 string
	RedisPrefix              string
	MinioEndpoint            string
	MinioAccessKey           string
	MinioSecretKey           string
	MinioBucket              string
	MinioPrefix              string
	MinioUseSSL              bool
	DemoTenantID             string
	LegacyUIEnabled          bool
	NATSURL                  string
	LiveKitURL               string
	LiveKitAPIKey            string
	LiveKitAPISecret         string
	CustomerWebDir           string
	PlatformAdminWebDir      string
	ClickHouseURL            string
	ClickHouseDB             string
	ClickHouseUser           string
	ClickHousePassword       string
	ClickHouseKMReadUser     string
	ClickHouseKMReadPassword string
	GeminiEmbedModel         string
	AIUsageRateVersion       string
	AIUsagePricingAsOf       string
	AIUsageCurrency          string
	AIUsageInputPriceMicros  int64
	AIUsageOutputPriceMicros int64
	AIUsageAudioPriceMicros  int64
	AuthDisabled             bool
	CookieSecure             bool
	CookieSameSite           string
	AllowedOrigins           []string
	JWTSecret                string
	JWTAccessTTL             time.Duration
	JWTRefreshTTL            time.Duration
	AuthCacheEnabled         bool
	AuthWriteBehindEnabled   bool
	AuthEventsEnabled        bool
	AuthUserCacheTTL         time.Duration
	EntitlementCacheEnabled  bool
	EntitlementCacheTTL      time.Duration
	TenantRegisterEnabled    bool
	TenantRegisterRateLimit  int
	TenantWebDir             string
	ProductWebDir            string
	ProductWebEnabled        bool
	LeadCaptureEnabled       bool
	LeadRateLimitPerIP       int
	FunnelRateLimitPerIP     int
	LeadDedupeWindowHours    int
	PublicBaseURL            string
	ResendAPIKey             string
	ResendFromEmail          string
	GoogleOAuthClientID      string
	GoogleOAuthClientSecret  string
	// Optional full redirect URI override (must match Google Console exactly).
	// Prefer http://localhost:PORT/... for local HTTP — Google rejects http://*.local.
	GoogleOAuthRedirectURL   string
	GitHubOAuthClientID      string
	GitHubOAuthClientSecret  string
	GitHubOAuthRedirectURL   string
	ChillPayMerchantCode     string
	ChillPayAPIKey           string
	ChillPayMD5Key           string
	ChillPayBaseURL          string
	ChillPayRouteNo          int
	ChillPayCurrency         string
	ChillPayCallbackURL      string
	ChillPayReturnURL        string
	PaymentCallbackDevBypass bool
	PaymentMockAutoFulfill   bool
	BillingSchedulerEnabled  bool
	BillingSchedulerPoll     time.Duration
	BillingGracePeriod       time.Duration
	BillingRetryDelays       []time.Duration
	// Quota / rate limit (SPRINT-013)
	QuotaEnabled         bool
	QuotaFailOpen        bool
	RateLimitEnabled     bool
	RateLimitChatPerMin  int
	RateLimitKMPerMin    int
	RateLimitVoicePerMin int
	QuotaConcurrentTTL   time.Duration
	// Embed (SPRINT-014)
	EmbedAllowEmptyOrigins    bool
	TenantSecretEncryptionKey string
	TenantSecretKeyVersion    string
	ConfigGroups              []string
	ConfigError               string
	// Preview sandbox (SPRINT-017)
	PreviewMaxConcurrent int
	// Customer import (SPRINT-019)
	CustomerImportMaxBytes int64
	CustomerImportMaxRows  int
	MonitoringProbeTimeout time.Duration
	// Mobile Call API and SDK (SPRINT-027)
	MobileCallAPIEnabled  bool
	MobileWSMaxFrameBytes int
	MobilePushEnabled     bool
	MobilePushProvider    string
	MobilePushTokenTTL    time.Duration
	// Cross-tenant audit log (SPRINT-028)
	AuditLogMode          string
	AuditLogDir           string
	AuditLogFlushInterval time.Duration
	AuditLogRetention     time.Duration
	AuditLogBatchSize     int
	AuditLogQueueSize     int
	AuditLogRetryBackoff  time.Duration
	AppEnv                string
}

func Load() Config {
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "dev"
	}
	configError := loadConfigFiles(appEnv)

	return Config{
		Port:            envOr("PORT", "8091"),
		GeminiAPIKey:    os.Getenv("GEMINI_API_KEY"),
		GeminiModel:     envOr("GEMINI_MODEL", "gemini-flash-latest"),
		GeminiLiveModel: envOr("GEMINI_LIVE_MODEL", "gemini-2.5-flash-native-audio-latest"),
		// Default true in non-prod for local DX; production never allows fallback.
		AllowPlatformGeminiFallback: envBool("ALLOW_PLATFORM_GEMINI_FALLBACK", true),
		Voice:                       envOr("VOICE", "Aoede"),
		PostgresURL:                 os.Getenv("POSTGRES_URL"),
		PostgresKMReadURL:           envOr("POSTGRES_KM_READONLY_URL", envOr("POSTGRES_READONLY_URL", "")),
		PostgresTicketWriteURL:      os.Getenv("POSTGRES_TICKET_WRITE_URL"),
		PostgresRLSEnforced:         envBool("POSTGRES_RLS_ENFORCED", false),
		PostgresSchema:              envOr("POSTGRES_SCHEMA", "callcenter"),
		RedisURL:                    os.Getenv("REDIS_URL"),
		RedisPrefix:                 envOr("REDIS_PREFIX", "monti_jarvis:"),
		MinioEndpoint:               os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey:              os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey:              os.Getenv("MINIO_SECRET_KEY"),
		MinioBucket:                 envOr("MINIO_BUCKET", "monti-jarvis"),
		MinioPrefix:                 envOr("MINIO_PREFIX", "calls/"),
		MinioUseSSL:                 envBool("MINIO_USE_SSL", false),
		DemoTenantID:                envOr("DEMO_TENANT_ID", "demo"),
		LegacyUIEnabled:             envBool("LEGACY_UI_ENABLED", false),
		NATSURL:                     envOr("NATS_URL", "nats://localhost:4222"),
		LiveKitURL:                  envOr("LIVEKIT_URL", "ws://localhost:7880"),
		LiveKitAPIKey:               envOr("LIVEKIT_API_KEY", "devkey"),
		LiveKitAPISecret:            envOr("LIVEKIT_API_SECRET", "secret"),
		CustomerWebDir:              envOr("CUSTOMER_WEB_DIR", "apps/customer-web/build"),
		PlatformAdminWebDir:         envOr("PLATFORM_ADMIN_WEB_DIR", "apps/platform-admin-web/build"),
		ClickHouseURL:               envOr("CLICKHOUSE_URL", "http://localhost:8123"),
		ClickHouseDB:                envOr("CLICKHOUSE_DB", "monti_jarvis"),
		ClickHouseUser:              envOr("CLICKHOUSE_USER", "monti"),
		ClickHousePassword:          envOr("CLICKHOUSE_PASSWORD", "monti"),
		ClickHouseKMReadUser:        envOr("CLICKHOUSE_KM_READONLY_USER", envOr("CLICKHOUSE_USER", "monti")),
		ClickHouseKMReadPassword:    envOr("CLICKHOUSE_KM_READONLY_PASSWORD", envOr("CLICKHOUSE_PASSWORD", "monti")),
		GeminiEmbedModel:            envOr("GEMINI_EMBED_MODEL", "gemini-embedding-001"),
		AIUsageRateVersion:          envOr("AI_USAGE_RATE_VERSION", "unconfigured"),
		AIUsagePricingAsOf:          envOr("AI_USAGE_PRICING_AS_OF", ""),
		AIUsageCurrency:             envOr("AI_USAGE_CURRENCY", "USD"),
		AIUsageInputPriceMicros:     envInt64("AI_USAGE_INPUT_PRICE_MICROS", 0),
		AIUsageOutputPriceMicros:    envInt64("AI_USAGE_OUTPUT_PRICE_MICROS", 0),
		AIUsageAudioPriceMicros:     envInt64("AI_USAGE_AUDIO_PRICE_MICROS", 0),
		AuthDisabled:                envBool("AUTH_DISABLED", true),
		CookieSecure:                envBool("COOKIE_SECURE", false),
		CookieSameSite:              strings.ToLower(envOr("COOKIE_SAMESITE", "lax")),
		AllowedOrigins:              splitOrigins(os.Getenv("ALLOWED_ORIGINS")),
		JWTSecret:                   os.Getenv("JWT_SECRET"),
		JWTAccessTTL:                envDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:               envDuration("JWT_REFRESH_TTL", 168*time.Hour),
		AuthCacheEnabled:            envBool("AUTH_CACHE_ENABLED", os.Getenv("REDIS_URL") != ""),
		AuthWriteBehindEnabled:      envBool("AUTH_WRITE_BEHIND_ENABLED", os.Getenv("REDIS_URL") != ""),
		AuthEventsEnabled:           envBool("AUTH_EVENTS_ENABLED", envOr("NATS_URL", "nats://localhost:4222") != ""),
		AuthUserCacheTTL:            envDuration("AUTH_USER_CACHE_TTL", 15*time.Minute),
		EntitlementCacheEnabled:     envBool("ENTITLEMENT_CACHE_ENABLED", os.Getenv("REDIS_URL") != ""),
		EntitlementCacheTTL:         envDuration("ENTITLEMENT_CACHE_TTL", 15*time.Minute),
		TenantRegisterEnabled:       envBool("TENANT_REGISTER_ENABLED", true),
		TenantRegisterRateLimit:     envInt("TENANT_REGISTER_RATE_LIMIT", 5),
		TenantWebDir:                envOr("TENANT_WEB_DIR", "apps/tenant-web/build"),
		ProductWebDir:               envOr("PRODUCT_WEB_DIR", "apps/product-web/build"),
		ProductWebEnabled:           envBool("PRODUCT_WEB_ENABLED", true),
		LeadCaptureEnabled:          envBool("LEAD_CAPTURE_ENABLED", true),
		LeadRateLimitPerIP:          envInt("LEAD_RATE_LIMIT_PER_IP", 10),
		FunnelRateLimitPerIP:        envInt("FUNNEL_RATE_LIMIT_PER_IP", 120),
		LeadDedupeWindowHours:       envInt("LEAD_DEDUPE_WINDOW_HOURS", 24),
		PublicBaseURL:               envOr("APP_PUBLIC_URL", "http://localhost:8091"),
		ResendAPIKey:                resolveResendAPIKey(),
		ResendFromEmail:             resolveResendFrom(),
		GoogleOAuthClientID:         os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		GoogleOAuthClientSecret:     os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		GoogleOAuthRedirectURL: envOr(
			"GOOGLE_OAUTH_REDIRECT_URL",
			envOr("OAUTH_GOOGLE_REDIRECT_URL", ""),
		),
		GitHubOAuthClientID:     os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
		GitHubOAuthClientSecret: os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"),
		GitHubOAuthRedirectURL: envOr(
			"GITHUB_OAUTH_REDIRECT_URL",
			envOr("OAUTH_GITHUB_REDIRECT_URL", ""),
		),
		ChillPayMerchantCode:     os.Getenv("CHILLPAY_MERCHANT_CODE"),
		ChillPayAPIKey:           os.Getenv("CHILLPAY_API_KEY"),
		ChillPayMD5Key:           os.Getenv("CHILLPAY_MD5_KEY"),
		ChillPayBaseURL:          envOr("CHILLPAY_BASE_URL", "https://sandbox-appsrv2.chillpay.co/api/v2/Payment"),
		ChillPayRouteNo:          envInt("CHILLPAY_ROUTE_NO", 1),
		ChillPayCurrency:         envOr("CHILLPAY_CURRENCY", "764"),
		ChillPayCallbackURL:      os.Getenv("CHILLPAY_CALLBACK_URL"),
		ChillPayReturnURL:        os.Getenv("CHILLPAY_RETURN_URL"),
		PaymentCallbackDevBypass: envBool("PAYMENT_CALLBACK_DEV_BYPASS", false),
		PaymentMockAutoFulfill:   envBool("PAYMENT_MOCK_AUTO_FULFILL", false),
		BillingSchedulerEnabled:  envBool("BILLING_SCHEDULER_ENABLED", false),
		BillingSchedulerPoll:     envDuration("BILLING_SCHEDULER_POLL_INTERVAL", time.Minute),
		BillingGracePeriod:       envDuration("BILLING_GRACE_PERIOD", 72*time.Hour),
		BillingRetryDelays:       envDurationList("BILLING_RETRY_DELAYS", []time.Duration{time.Hour, 6 * time.Hour, 24 * time.Hour}),
		// Default on when Redis is configured (same pattern as entitlement cache).
		QuotaEnabled:              envBool("QUOTA_ENABLED", os.Getenv("REDIS_URL") != ""),
		QuotaFailOpen:             envBool("QUOTA_FAIL_OPEN", true),
		RateLimitEnabled:          envBool("RATE_LIMIT_ENABLED", os.Getenv("REDIS_URL") != ""),
		RateLimitChatPerMin:       envInt("RATE_LIMIT_CHAT_PER_MIN", 60),
		RateLimitKMPerMin:         envInt("RATE_LIMIT_KM_PER_MIN", 30),
		RateLimitVoicePerMin:      envInt("RATE_LIMIT_VOICE_PER_MIN", 20),
		QuotaConcurrentTTL:        envDuration("QUOTA_CONCURRENT_TTL", 2*time.Hour),
		EmbedAllowEmptyOrigins:    envBool("EMBED_ALLOW_EMPTY_ORIGINS", true),
		TenantSecretEncryptionKey: os.Getenv("TENANT_SECRET_ENCRYPTION_KEY"),
		TenantSecretKeyVersion:    envOr("TENANT_SECRET_KEY_VERSION", "v1"),
		ConfigGroups:              splitConfigGroups(os.Getenv("CONFIG_GROUPS")),
		ConfigError:               configError,
		PreviewMaxConcurrent:      envInt("PREVIEW_MAX_CONCURRENT", 2),
		CustomerImportMaxBytes:    int64(envInt("CUSTOMER_IMPORT_MAX_BYTES", 2*1024*1024)),
		CustomerImportMaxRows:     envInt("CUSTOMER_IMPORT_MAX_ROWS", 5000),
		MonitoringProbeTimeout:    envDuration("MONITORING_PROBE_TIMEOUT", 2*time.Second),
		MobileCallAPIEnabled:      envBool("MOBILE_CALL_API_ENABLED", false),
		MobileWSMaxFrameBytes:     positiveEnvInt("MOBILE_WS_MAX_FRAME_BYTES", 32768),
		MobilePushEnabled:         envBool("MOBILE_PUSH_ENABLED", false),
		MobilePushProvider:        envOr("MOBILE_PUSH_PROVIDER", "auto"),
		MobilePushTokenTTL:        envDuration("MOBILE_PUSH_TOKEN_TTL", 15*time.Minute),
		AuditLogMode:              envOr("AUDIT_LOG_MODE", "spool"),
		AuditLogDir:               envOr("AUDIT_LOG_DIR", "./var/audit"),
		AuditLogFlushInterval:     envDuration("AUDIT_LOG_FLUSH_INTERVAL", 5*time.Second),
		AuditLogRetention:         envDuration("AUDIT_LOG_RETENTION", time.Hour),
		AuditLogBatchSize:         positiveEnvInt("AUDIT_LOG_BATCH_SIZE", 500),
		AuditLogQueueSize:         positiveEnvInt("AUDIT_LOG_QUEUE_SIZE", 10000),
		AuditLogRetryBackoff:      envDuration("AUDIT_LOG_RETRY_BACKOFF", time.Second),
		AppEnv:                    appEnv,
	}
}

func envDurationList(key string, fallback []time.Duration) []time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return append([]time.Duration(nil), fallback...)
	}
	parts := strings.Split(raw, ",")
	out := make([]time.Duration, 0, len(parts))
	for _, part := range parts {
		value, err := time.ParseDuration(strings.TrimSpace(part))
		if err != nil || value <= 0 {
			return append([]time.Duration(nil), fallback...)
		}
		out = append(out, value)
	}
	if len(out) == 0 {
		return append([]time.Duration(nil), fallback...)
	}
	return out
}

// IsProduction reports whether the process is running as production.
func (c Config) IsProduction() bool {
	env := strings.ToLower(strings.TrimSpace(c.AppEnv))
	return env == "production" || env == "prod"
}

// PlatformGeminiFallbackAllowed is true only for non-production when the flag is set.
func (c Config) PlatformGeminiFallbackAllowed() bool {
	if c.IsProduction() {
		return false
	}
	return c.AllowPlatformGeminiFallback
}

// ValidateProductionSecurity enforces the Sprint 41 capability split before
// production serves traffic. Development keeps the existing local setup
// compatible; production must opt into explicit database principals.
func (c Config) ValidateProductionSecurity() error {
	if !strings.EqualFold(strings.TrimSpace(c.AppEnv), "production") {
		return nil
	}
	if c.AuthDisabled {
		return fmt.Errorf("AUTH_DISABLED must be false in production")
	}
	if len(strings.TrimSpace(c.JWTSecret)) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 bytes in production")
	}
	if !c.CookieSecure {
		return fmt.Errorf("COOKIE_SECURE must be true in production")
	}
	switch strings.ToLower(strings.TrimSpace(c.CookieSameSite)) {
	case "lax", "strict":
	default:
		return fmt.Errorf("COOKIE_SAMESITE must be lax or strict in production")
	}
	if len(c.AllowedOrigins) == 0 {
		return fmt.Errorf("ALLOWED_ORIGINS must contain an explicit origin in production")
	}
	for _, origin := range c.AllowedOrigins {
		parsed, err := url.Parse(origin)
		if err != nil || (parsed.Scheme != "https" && parsed.Scheme != "http") || parsed.Host == "" || strings.Contains(origin, "*") {
			return fmt.Errorf("ALLOWED_ORIGINS contains an invalid or wildcard origin")
		}
	}
	if !c.PostgresRLSEnforced {
		return fmt.Errorf("POSTGRES_RLS_ENFORCED must be true in production")
	}
	if strings.TrimSpace(c.PostgresURL) == "" {
		return fmt.Errorf("POSTGRES_URL is required in production")
	}
	if strings.TrimSpace(c.PostgresKMReadURL) == "" {
		return fmt.Errorf("POSTGRES_KM_READONLY_URL is required in production")
	}
	if strings.TrimSpace(c.PostgresTicketWriteURL) == "" {
		return fmt.Errorf("POSTGRES_TICKET_WRITE_URL is required in production")
	}
	writer := databasePrincipal(c.PostgresURL)
	kmRead := databasePrincipal(c.PostgresKMReadURL)
	ticketWrite := databasePrincipal(c.PostgresTicketWriteURL)
	if writer == kmRead || writer == ticketWrite || kmRead == ticketWrite {
		return fmt.Errorf("production database capability users must be distinct")
	}
	return nil
}

func databasePrincipal(raw string) string {
	raw = strings.TrimSpace(raw)
	if parsed, err := url.Parse(raw); err == nil && parsed.User != nil {
		if user := strings.TrimSpace(parsed.User.Username()); user != "" {
			return user
		}
	}
	// Keyword/value DSNs and malformed URLs cannot expose a principal safely;
	// retain the full value so exact duplicate URLs are still rejected.
	return raw
}

func splitOrigins(raw string) []string {
	seen := map[string]bool{}
	var origins []string
	for _, value := range strings.Split(raw, ",") {
		value = strings.TrimRight(strings.TrimSpace(value), "/")
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		origins = append(origins, value)
	}
	return origins
}

var configGroupNames = map[string]bool{
	"ai": true, "ops": true, "email": true, "features": true,
}

// loadConfigFiles preserves the legacy dotenv lookup while adding explicit
// per-environment groups. Existing process values always win.
func loadConfigFiles(appEnv string) string {
	files := []string{"infra/.env." + appEnv, "infra/.env", ".env." + appEnv, ".env"}
	for _, file := range files {
		if err := loadOptionalEnvFile(file); err != nil {
			return err.Error()
		}
	}
	groups := splitConfigGroups(os.Getenv("CONFIG_GROUPS"))
	seen := map[string]string{}
	for _, group := range groups {
		if !configGroupNames[group] {
			return fmt.Sprintf("unknown CONFIG_GROUPS value %q", group)
		}
		file := "infra/.env." + appEnv + "." + group
		values, err := godotenv.Read(file)
		if err != nil {
			return fmt.Sprintf("config group %s: %v", group, err)
		}
		for key, value := range values {
			if previous, ok := seen[key]; ok && previous != group {
				return fmt.Sprintf("config key %s defined in groups %s and %s", key, previous, group)
			}
			seen[key] = group
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, value)
			}
		}
	}
	return ""
}

func loadOptionalEnvFile(file string) error {
	values, err := godotenv.Read(file)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for key, value := range values {
		if _, exists := os.LookupEnv(key); !exists {
			_ = os.Setenv(key, value)
		}
	}
	return nil
}

func splitConfigGroups(raw string) []string {
	seen := map[string]bool{}
	var groups []string
	for _, value := range strings.Split(raw, ",") {
		value = strings.ToLower(strings.TrimSpace(value))
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		groups = append(groups, value)
	}
	return groups
}

func envInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envInt64(key string, fallback int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func positiveEnvInt(key string, fallback int) int {
	value := envInt(key, fallback)
	if value <= 0 {
		return fallback
	}
	return value
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func envOr(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}
	return parsed
}

// resolveResendAPIKey respects RESEND_ENABLED=false to force-disable mailer.
func resolveResendAPIKey() string {
	if !envBool("RESEND_ENABLED", true) {
		return ""
	}
	return strings.TrimSpace(os.Getenv("RESEND_API_KEY"))
}

// resolveResendFrom builds the Resend "from" header.
// Priority:
//  1. RESEND_FROM_EMAIL — full "Name <addr@domain>" or bare address
//  2. RESEND_FROM_ADDR (+ optional RESEND_FROM_NAME) — matches common Resend env naming
//  3. empty — mailer disabled until a verified domain sender is configured
//
// Never default to @monti.local: Resend rejects unverified domains with HTTP 403.
func resolveResendFrom() string {
	if from := strings.TrimSpace(os.Getenv("RESEND_FROM_EMAIL")); from != "" {
		return from
	}
	addr := strings.TrimSpace(os.Getenv("RESEND_FROM_ADDR"))
	if addr == "" {
		// Legacy alias used in some env files
		addr = strings.TrimSpace(os.Getenv("RESEND_FROM"))
	}
	if addr == "" {
		return ""
	}
	name := strings.TrimSpace(os.Getenv("RESEND_FROM_NAME"))
	if name == "" {
		return addr
	}
	return fmt.Sprintf("%s <%s>", name, addr)
}
