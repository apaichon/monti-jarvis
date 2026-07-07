package auth

import "time"

type Config struct {
	JWTSecret          string
	AccessTTL          time.Duration
	RefreshTTL         time.Duration
	UserCacheTTL       time.Duration
	AuthDisabled       bool
	CacheEnabled       bool
	WriteBehindEnabled bool
	EventsEnabled      bool
	RedisPrefix        string
}