package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type CachedUser struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name"`
	Status        string `json:"status"`
	Role          string `json:"role"`
	TenantID      string `json:"tenant_id"`
	AuthProvider  string `json:"auth_provider,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	PasswordHash  string `json:"password_hash,omitempty"`
}

type CachedRefresh struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	Revoked   bool      `json:"revoked"`
}

type Cache struct {
	rdb        *redis.Client
	prefix     string
	userTTL    time.Duration
	refreshTTL time.Duration
	enabled    bool
}

func NewCache(rdb *redis.Client, prefix string, userTTL, refreshTTL time.Duration, enabled bool) *Cache {
	if prefix == "" {
		prefix = "monti_jarvis:"
	}
	return &Cache{
		rdb:        rdb,
		prefix:     prefix,
		userTTL:    userTTL,
		refreshTTL: refreshTTL,
		enabled:    enabled && rdb != nil,
	}
}

func (c *Cache) Enabled() bool {
	return c != nil && c.enabled
}

func (c *Cache) key(parts ...string) string {
	return c.prefix + strings.Join(parts, "")
}

func (c *Cache) GetUserByID(ctx context.Context, userID string) (CachedUser, bool, error) {
	if !c.Enabled() {
		return CachedUser{}, false, nil
	}
	raw, err := c.rdb.Get(ctx, c.key("auth:user:", userID)).Result()
	if err == redis.Nil {
		return CachedUser{}, false, nil
	}
	if err != nil {
		return CachedUser{}, false, err
	}
	var user CachedUser
	if err := json.Unmarshal([]byte(raw), &user); err != nil {
		return CachedUser{}, false, err
	}
	return user, true, nil
}

func (c *Cache) GetUserIDByEmail(ctx context.Context, email string) (string, bool, error) {
	if !c.Enabled() {
		return "", false, nil
	}
	email = strings.ToLower(strings.TrimSpace(email))
	id, err := c.rdb.Get(ctx, c.key("auth:user:email:", email)).Result()
	if err == redis.Nil {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}

func (c *Cache) PutUser(ctx context.Context, user CachedUser, includeHash bool) error {
	if !c.Enabled() {
		return nil
	}
	cp := user
	if !includeHash {
		cp.PasswordHash = ""
	}
	payload, err := json.Marshal(cp)
	if err != nil {
		return err
	}
	email := strings.ToLower(strings.TrimSpace(user.Email))
	pipe := c.rdb.Pipeline()
	pipe.Set(ctx, c.key("auth:user:", user.ID), payload, c.userTTL)
	if email != "" {
		pipe.Set(ctx, c.key("auth:user:email:", email), user.ID, c.userTTL)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func (c *Cache) StripPasswordHash(ctx context.Context, user CachedUser) error {
	user.PasswordHash = ""
	return c.PutUser(ctx, user, false)
}

func (c *Cache) GetRefresh(ctx context.Context, tokenHash string) (CachedRefresh, bool, error) {
	if !c.Enabled() {
		return CachedRefresh{}, false, nil
	}
	raw, err := c.rdb.Get(ctx, c.key("auth:refresh:", tokenHash)).Result()
	if err == redis.Nil {
		return CachedRefresh{}, false, nil
	}
	if err != nil {
		return CachedRefresh{}, false, err
	}
	var row CachedRefresh
	if err := json.Unmarshal([]byte(raw), &row); err != nil {
		return CachedRefresh{}, false, err
	}
	return row, true, nil
}

func (c *Cache) PutRefresh(ctx context.Context, tokenHash string, row CachedRefresh) error {
	if !c.Enabled() {
		return nil
	}
	payload, err := json.Marshal(row)
	if err != nil {
		return err
	}
	ttl := time.Until(row.ExpiresAt)
	if ttl <= 0 {
		ttl = c.refreshTTL
	}
	return c.rdb.Set(ctx, c.key("auth:refresh:", tokenHash), payload, ttl).Err()
}

func (c *Cache) RevokeRefresh(ctx context.Context, tokenHash string) error {
	if !c.Enabled() {
		return nil
	}
	row, ok, err := c.GetRefresh(ctx, tokenHash)
	if err != nil {
		return err
	}
	if !ok {
		row = CachedRefresh{Revoked: true}
	}
	row.Revoked = true
	return c.PutRefresh(ctx, tokenHash, row)
}

func (c *Cache) IsJTIDenied(ctx context.Context, jti string) (bool, error) {
	if !c.Enabled() || jti == "" {
		return false, nil
	}
	_, err := c.rdb.Get(ctx, c.key("auth:deny:jti:", jti)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (c *Cache) DenyJTI(ctx context.Context, jti string, ttl time.Duration) error {
	if !c.Enabled() || jti == "" {
		return nil
	}
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}
	return c.rdb.Set(ctx, c.key("auth:deny:jti:", jti), "1", ttl).Err()
}

const persistStream = "auth:wb:queue"
const persistDLQ = "auth:wb:dlq"

type PersistJob struct {
	Op        string    `json:"op"`
	ID        string    `json:"id,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
	TokenHash string    `json:"token_hash"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	Actor     string    `json:"actor,omitempty"`
}

func (c *Cache) EnqueuePersist(ctx context.Context, job PersistJob) error {
	if !c.Enabled() {
		return fmt.Errorf("redis cache not enabled")
	}
	payload, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return c.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: c.key(persistStream),
		Values: map[string]any{"job": string(payload)},
	}).Err()
}

func (c *Cache) PersistLag(ctx context.Context) (int64, error) {
	if !c.Enabled() {
		return 0, nil
	}
	return c.rdb.XLen(ctx, c.key(persistStream)).Result()
}