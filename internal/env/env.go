package env

import (
	"os"
	"strconv"

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
	CustomerWebDir   string
	ClickHouseURL      string
	ClickHouseDB       string
	ClickHouseUser     string
	ClickHousePassword string
	GeminiEmbedModel string
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
		CustomerWebDir:   envOr("CUSTOMER_WEB_DIR", "apps/customer-web/build"),
		ClickHouseURL:      envOr("CLICKHOUSE_URL", "http://localhost:8123"),
		ClickHouseDB:       envOr("CLICKHOUSE_DB", "monti_jarvis"),
		ClickHouseUser:     envOr("CLICKHOUSE_USER", "monti"),
		ClickHousePassword: envOr("CLICKHOUSE_PASSWORD", "monti"),
		GeminiEmbedModel: envOr("GEMINI_EMBED_MODEL", "gemini-embedding-001"),
	}
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