package env

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	GeminiAPIKey    string
	GeminiModel     string
	GeminiLiveModel string
	Voice           string
	PostgresURL     string
	PostgresSchema  string
	RedisURL        string
	RedisPrefix     string
	MinioEndpoint   string
	MinioAccessKey  string
	MinioSecretKey  string
	MinioBucket     string
	MinioPrefix     string
	MinioUseSSL     bool
	DemoTenantID    string
	LegacyUIEnabled bool
	NATSURL         string
	LiveKitURL      string
	LiveKitAPIKey   string
	LiveKitAPISecret string
	CustomerWebDir      string
	PlatformAdminWebDir string
	ClickHouseURL      string
	ClickHouseDB       string
	ClickHouseUser     string
	ClickHousePassword string
	GeminiEmbedModel   string
	AuthDisabled           bool
	JWTSecret              string
	JWTAccessTTL           time.Duration
	JWTRefreshTTL          time.Duration
	AuthCacheEnabled       bool
	AuthWriteBehindEnabled bool
	AuthEventsEnabled      bool
	AuthUserCacheTTL         time.Duration
	EntitlementCacheEnabled  bool
	EntitlementCacheTTL      time.Duration
	TenantRegisterEnabled    bool
	TenantRegisterRateLimit  int
	TenantWebDir             string
	PublicBaseURL            string
	ResendAPIKey             string
	ResendFromEmail          string
	GoogleOAuthClientID      string
	GoogleOAuthClientSecret  string
	GitHubOAuthClientID      string
	GitHubOAuthClientSecret  string
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
		Port:             envOr("PORT", "8091"),
		GeminiAPIKey:     os.Getenv("GEMINI_API_KEY"),
		GeminiModel:      envOr("GEMINI_MODEL", "gemini-flash-latest"),
		GeminiLiveModel:  envOr("GEMINI_LIVE_MODEL", "gemini-2.5-flash-native-audio-latest"),
		Voice:            envOr("VOICE", "Aoede"),
		PostgresURL:      os.Getenv("POSTGRES_URL"),
		PostgresSchema:   envOr("POSTGRES_SCHEMA", "callcenter"),
		RedisURL:         os.Getenv("REDIS_URL"),
		RedisPrefix:      envOr("REDIS_PREFIX", "monti_jarvis:"),
		MinioEndpoint:    os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey:   os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey:   os.Getenv("MINIO_SECRET_KEY"),
		MinioBucket:      envOr("MINIO_BUCKET", "monti-jarvis"),
		MinioPrefix:      envOr("MINIO_PREFIX", "calls/"),
		MinioUseSSL:      envBool("MINIO_USE_SSL", false),
		DemoTenantID:     envOr("DEMO_TENANT_ID", "demo"),
		LegacyUIEnabled:  envBool("LEGACY_UI_ENABLED", false),
		NATSURL:          envOr("NATS_URL", "nats://localhost:4222"),
		LiveKitURL:       envOr("LIVEKIT_URL", "ws://localhost:7880"),
		LiveKitAPIKey:    envOr("LIVEKIT_API_KEY", "devkey"),
		LiveKitAPISecret: envOr("LIVEKIT_API_SECRET", "secret"),
		CustomerWebDir:      envOr("CUSTOMER_WEB_DIR", "apps/customer-web/build"),
		PlatformAdminWebDir: envOr("PLATFORM_ADMIN_WEB_DIR", "apps/platform-admin-web/build"),
		ClickHouseURL:      envOr("CLICKHOUSE_URL", "http://localhost:8123"),
		ClickHouseDB:       envOr("CLICKHOUSE_DB", "monti_jarvis"),
		ClickHouseUser:     envOr("CLICKHOUSE_USER", "monti"),
		ClickHousePassword: envOr("CLICKHOUSE_PASSWORD", "monti"),
		GeminiEmbedModel: envOr("GEMINI_EMBED_MODEL", "gemini-embedding-001"),
		AuthDisabled:           envBool("AUTH_DISABLED", true),
		JWTSecret:              os.Getenv("JWT_SECRET"),
		JWTAccessTTL:           envDuration("JWT_ACCESS_TTL", 15*time.Minute),
		JWTRefreshTTL:          envDuration("JWT_REFRESH_TTL", 168*time.Hour),
		AuthCacheEnabled:       envBool("AUTH_CACHE_ENABLED", os.Getenv("REDIS_URL") != ""),
		AuthWriteBehindEnabled: envBool("AUTH_WRITE_BEHIND_ENABLED", os.Getenv("REDIS_URL") != ""),
		AuthEventsEnabled:      envBool("AUTH_EVENTS_ENABLED", envOr("NATS_URL", "nats://localhost:4222") != ""),
		AuthUserCacheTTL:        envDuration("AUTH_USER_CACHE_TTL", 15*time.Minute),
		EntitlementCacheEnabled: envBool("ENTITLEMENT_CACHE_ENABLED", os.Getenv("REDIS_URL") != ""),
		EntitlementCacheTTL:     envDuration("ENTITLEMENT_CACHE_TTL", 15*time.Minute),
		TenantRegisterEnabled:   envBool("TENANT_REGISTER_ENABLED", true),
		TenantRegisterRateLimit: envInt("TENANT_REGISTER_RATE_LIMIT", 5),
		TenantWebDir:            envOr("TENANT_WEB_DIR", "apps/tenant-web/build"),
		PublicBaseURL:           envOr("APP_PUBLIC_URL", "http://localhost:8091"),
		ResendAPIKey:            os.Getenv("RESEND_API_KEY"),
		ResendFromEmail:         envOr("RESEND_FROM_EMAIL", "Monti <onboarding@monti.local>"),
		GoogleOAuthClientID:     os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		GoogleOAuthClientSecret: os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
		GitHubOAuthClientID:     os.Getenv("GITHUB_OAUTH_CLIENT_ID"),
		GitHubOAuthClientSecret: os.Getenv("GITHUB_OAUTH_CLIENT_SECRET"),
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