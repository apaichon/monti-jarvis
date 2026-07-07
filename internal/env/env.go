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
		Port:            envOr("PORT", "8091"),
		GeminiAPIKey:    os.Getenv("GEMINI_API_KEY"),
		GeminiModel:     envOr("GEMINI_MODEL", "gemini-flash-latest"),
		GeminiLiveModel: envOr("GEMINI_LIVE_MODEL", "gemini-2.5-flash-native-audio-latest"),
		Voice:           envOr("VOICE", "Aoede"),
		PostgresURL:     os.Getenv("POSTGRES_URL"),
		PostgresSchema:  envOr("POSTGRES_SCHEMA", "callcenter"),
		RedisURL:        os.Getenv("REDIS_URL"),
		RedisPrefix:     envOr("REDIS_PREFIX", "monti_jarvis:"),
		MinioEndpoint:   os.Getenv("MINIO_ENDPOINT"),
		MinioAccessKey:  os.Getenv("MINIO_ACCESS_KEY"),
		MinioSecretKey:  os.Getenv("MINIO_SECRET_KEY"),
		MinioBucket:     envOr("MINIO_BUCKET", "monti-jarvis"),
		MinioPrefix:     envOr("MINIO_PREFIX", "calls/"),
		MinioUseSSL:     envBool("MINIO_USE_SSL", false),
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