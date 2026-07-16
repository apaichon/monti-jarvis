package env

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                    string
	GeminiAPIKey            string
	GeminiModel             string
	GeminiLiveModel         string
	Voice                   string
	PostgresURL             string
	PostgresSchema          string
	RedisURL                string
	RedisPrefix             string
	MinioEndpoint           string
	MinioAccessKey          string
	MinioSecretKey          string
	MinioBucket             string
	MinioPrefix             string
	MinioUseSSL             bool
	DemoTenantID            string
	LegacyUIEnabled         bool
	NATSURL                 string
	LiveKitURL              string
	LiveKitAPIKey           string
	LiveKitAPISecret        string
	CustomerWebDir          string
	PlatformAdminWebDir     string
	ClickHouseURL           string
	ClickHouseDB            string
	ClickHouseUser          string
	ClickHousePassword      string
	GeminiEmbedModel        string
	AuthDisabled            bool
	JWTSecret               string
	JWTAccessTTL            time.Duration
	JWTRefreshTTL           time.Duration
	AuthCacheEnabled        bool
	AuthWriteBehindEnabled  bool
	AuthEventsEnabled       bool
	AuthUserCacheTTL        time.Duration
	EntitlementCacheEnabled bool
	EntitlementCacheTTL     time.Duration
	TenantRegisterEnabled   bool
	TenantRegisterRateLimit int
	TenantWebDir            string
	PublicBaseURL           string
	ResendAPIKey            string
	ResendFromEmail         string
	GoogleOAuthClientID     string
	GoogleOAuthClientSecret string
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
	// Quota / rate limit (SPRINT-013)
	QuotaEnabled         bool
	QuotaFailOpen        bool
	RateLimitEnabled     bool
	RateLimitChatPerMin  int
	RateLimitKMPerMin    int
	RateLimitVoicePerMin int
	QuotaConcurrentTTL   time.Duration
	// Embed (SPRINT-014)
	EmbedAllowEmptyOrigins bool
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
	_ = godotenv.Load("infra/.env." + appEnv)
	_ = godotenv.Load("infra/.env")
	_ = godotenv.Load(".env." + appEnv)
	_ = godotenv.Load()

	return Config{
		Port:                    envOr("PORT", "8091"),
		GeminiAPIKey:            os.Getenv("GEMINI_API_KEY"),
		GeminiModel:             envOr("GEMINI_MODEL", "gemini-flash-latest"),
		GeminiLiveModel:         envOr("GEMINI_LIVE_MODEL", "gemini-2.5-flash-native-audio-latest"),
		Voice:                   envOr("VOICE", "Aoede"),
		PostgresURL:             os.Getenv("POSTGRES_URL"),
		PostgresSchema:          envOr("POSTGRES_SCHEMA", "callcenter"),
		RedisURL:                os.Getenv("REDIS_URL"),
		RedisPrefix:             envOr("REDIS_PREFIX", "monti_jarvis:"),
		MinioEndpoint:           os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey:          os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey:          os.Getenv("MINIO_SECRET_KEY"),
		MinioBucket:             envOr("MINIO_BUCKET", "monti-jarvis"),
		MinioPrefix:             envOr("MINIO_PREFIX", "calls/"),
		MinioUseSSL:             envBool("MINIO_USE_SSL", false),
		DemoTenantID:            envOr("DEMO_TENANT_ID", "demo"),
		LegacyUIEnabled:         envBool("LEGACY_UI_ENABLED", false),
		NATSURL:                 envOr("NATS_URL", "nats://localhost:4222"),
		LiveKitURL:              envOr("LIVEKIT_URL", "ws://localhost:7880"),
		LiveKitAPIKey:           envOr("LIVEKIT_API_KEY", "devkey"),
		LiveKitAPISecret:        envOr("LIVEKIT_API_SECRET", "secret"),
		CustomerWebDir:          envOr("CUSTOMER_WEB_DIR", "apps/customer-web/build"),
		PlatformAdminWebDir:     envOr("PLATFORM_ADMIN_WEB_DIR", "apps/platform-admin-web/build"),
		ClickHouseURL:           envOr("CLICKHOUSE_URL", "http://localhost:8123"),
		ClickHouseDB:            envOr("CLICKHOUSE_DB", "monti_jarvis"),
		ClickHouseUser:          envOr("CLICKHOUSE_USER", "monti"),
		ClickHousePassword:      envOr("CLICKHOUSE_PASSWORD", "monti"),
		GeminiEmbedModel:        envOr("GEMINI_EMBED_MODEL", "gemini-embedding-001"),
		AuthDisabled:            envBool("AUTH_DISABLED", true),
		JWTSecret:               os.Getenv("JWT_SECRET"),
		JWTAccessTTL:            envDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:           envDuration("JWT_REFRESH_TTL", 168*time.Hour),
		AuthCacheEnabled:        envBool("AUTH_CACHE_ENABLED", os.Getenv("REDIS_URL") != ""),
		AuthWriteBehindEnabled:  envBool("AUTH_WRITE_BEHIND_ENABLED", os.Getenv("REDIS_URL") != ""),
		AuthEventsEnabled:       envBool("AUTH_EVENTS_ENABLED", envOr("NATS_URL", "nats://localhost:4222") != ""),
		AuthUserCacheTTL:        envDuration("AUTH_USER_CACHE_TTL", 15*time.Minute),
		EntitlementCacheEnabled: envBool("ENTITLEMENT_CACHE_ENABLED", os.Getenv("REDIS_URL") != ""),
		EntitlementCacheTTL:     envDuration("ENTITLEMENT_CACHE_TTL", 15*time.Minute),
		TenantRegisterEnabled:   envBool("TENANT_REGISTER_ENABLED", true),
		TenantRegisterRateLimit: envInt("TENANT_REGISTER_RATE_LIMIT", 5),
		TenantWebDir:            envOr("TENANT_WEB_DIR", "apps/tenant-web/build"),
		PublicBaseURL:           envOr("APP_PUBLIC_URL", "http://localhost:8091"),
		ResendAPIKey:            resolveResendAPIKey(),
		ResendFromEmail:         resolveResendFrom(),
		GoogleOAuthClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		GoogleOAuthClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
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
		// Default on when Redis is configured (same pattern as entitlement cache).
		QuotaEnabled:           envBool("QUOTA_ENABLED", os.Getenv("REDIS_URL") != ""),
		QuotaFailOpen:          envBool("QUOTA_FAIL_OPEN", true),
		RateLimitEnabled:       envBool("RATE_LIMIT_ENABLED", os.Getenv("REDIS_URL") != ""),
		RateLimitChatPerMin:    envInt("RATE_LIMIT_CHAT_PER_MIN", 60),
		RateLimitKMPerMin:      envInt("RATE_LIMIT_KM_PER_MIN", 30),
		RateLimitVoicePerMin:   envInt("RATE_LIMIT_VOICE_PER_MIN", 20),
		QuotaConcurrentTTL:     envDuration("QUOTA_CONCURRENT_TTL", 2*time.Hour),
		EmbedAllowEmptyOrigins: envBool("EMBED_ALLOW_EMPTY_ORIGINS", true),
		PreviewMaxConcurrent:   envInt("PREVIEW_MAX_CONCURRENT", 2),
		CustomerImportMaxBytes: int64(envInt("CUSTOMER_IMPORT_MAX_BYTES", 2*1024*1024)),
		CustomerImportMaxRows:  envInt("CUSTOMER_IMPORT_MAX_ROWS", 5000),
		MonitoringProbeTimeout: envDuration("MONITORING_PROBE_TIMEOUT", 2*time.Second),
		MobileCallAPIEnabled:   envBool("MOBILE_CALL_API_ENABLED", false),
		MobileWSMaxFrameBytes:  positiveEnvInt("MOBILE_WS_MAX_FRAME_BYTES", 32768),
		MobilePushEnabled:      envBool("MOBILE_PUSH_ENABLED", false),
		MobilePushProvider:     envOr("MOBILE_PUSH_PROVIDER", "auto"),
		MobilePushTokenTTL:     envDuration("MOBILE_PUSH_TOKEN_TTL", 15*time.Minute),
		AuditLogMode:           envOr("AUDIT_LOG_MODE", "spool"),
		AuditLogDir:            envOr("AUDIT_LOG_DIR", "./var/audit"),
		AuditLogFlushInterval:  envDuration("AUDIT_LOG_FLUSH_INTERVAL", 5*time.Second),
		AuditLogRetention:      envDuration("AUDIT_LOG_RETENTION", time.Hour),
		AuditLogBatchSize:      positiveEnvInt("AUDIT_LOG_BATCH_SIZE", 500),
		AuditLogQueueSize:      positiveEnvInt("AUDIT_LOG_QUEUE_SIZE", 10000),
		AuditLogRetryBackoff:   envDuration("AUDIT_LOG_RETRY_BACKOFF", time.Second),
		AppEnv:                 appEnv,
	}
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
