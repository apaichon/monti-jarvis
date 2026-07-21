package store

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
)

var (
	ErrEmbedNotFound = errors.New("embed not found")
	ErrEmbedDisabled = errors.New("embed disabled")
)

// TenantEmbedConfig is one row of callcenter.tenant_embed_configs.
type TenantEmbedConfig struct {
	TenantID       string    `json:"tenant_id"`
	EmbedKey       string    `json:"embed_key"`
	Enabled        bool      `json:"enabled"`
	AuthRequired   bool      `json:"auth_required"`
	AllowedOrigins []string  `json:"allowed_origins"`
	DefaultAgentID string    `json:"default_agent_id,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
}

func (s *Store) ensureEmbedSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmts := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.tenant_embed_configs (
  tenant_id text PRIMARY KEY REFERENCES %s.tenants(id) ON DELETE CASCADE,
  embed_key text NOT NULL,
  enabled boolean NOT NULL DEFAULT false,
  allowed_origins jsonb NOT NULL DEFAULT '[]'::jsonb,
  default_agent_id text,%s
)`, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`ALTER TABLE %s.tenant_embed_configs ADD COLUMN IF NOT EXISTS auth_required boolean NOT NULL DEFAULT false`, schema),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS tenant_embed_configs_embed_key_uidx
ON %s.tenant_embed_configs (embed_key)`, schema),
	}
	for _, stmt := range stmts {
		if _, err := s.pg.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("embed schema: %w", err)
		}
	}
	return nil
}

// NewEmbedKey returns emb_ + 32 hex chars.
func NewEmbedKey() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return "emb_" + hex.EncodeToString(b[:]), nil
}

// GetEmbedConfigByTenant returns config or ErrEmbedNotFound.
func (s *Store) GetEmbedConfigByTenant(ctx context.Context, tenantID string) (*TenantEmbedConfig, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	return s.scanEmbedConfig(ctx, fmt.Sprintf(`
SELECT tenant_id, embed_key, enabled, auth_required, allowed_origins, COALESCE(default_agent_id, ''), created_at, updated_at
FROM %s.tenant_embed_configs WHERE tenant_id = $1`, schema), tenantID)
}

// GetEmbedConfigByKey returns config by public key or ErrEmbedNotFound.
func (s *Store) GetEmbedConfigByKey(ctx context.Context, embedKey string) (*TenantEmbedConfig, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	return s.scanEmbedConfig(ctx, fmt.Sprintf(`
SELECT tenant_id, embed_key, enabled, auth_required, allowed_origins, COALESCE(default_agent_id, ''), created_at, updated_at
FROM %s.tenant_embed_configs WHERE embed_key = $1`, schema), embedKey)
}

// GetOrCreateEmbedConfig lazy-creates a disabled config with a new key.
func (s *Store) GetOrCreateEmbedConfig(ctx context.Context, tenantID string) (*TenantEmbedConfig, error) {
	cfg, err := s.GetEmbedConfigByTenant(ctx, tenantID)
	if err == nil {
		return cfg, nil
	}
	if !errors.Is(err, ErrEmbedNotFound) {
		return nil, err
	}
	key, err := NewEmbedKey()
	if err != nil {
		return nil, err
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_embed_configs (tenant_id, embed_key, enabled, allowed_origins, created_by, updated_by)
VALUES ($1, $2, false, '[]'::jsonb, $3, $3)
ON CONFLICT (tenant_id) DO NOTHING`, schema), tenantID, key, actor)
	if err != nil {
		return nil, err
	}
	return s.GetEmbedConfigByTenant(ctx, tenantID)
}

// UpdateEmbedConfig updates enabled, origins, default agent.
func (s *Store) UpdateEmbedConfig(ctx context.Context, tenantID string, enabled bool, origins []string, defaultAgentID string) (*TenantEmbedConfig, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	if origins == nil {
		origins = []string{}
	}
	// Normalize origins
	clean := make([]string, 0, len(origins))
	for _, o := range origins {
		o = strings.TrimSpace(o)
		if o == "" {
			continue
		}
		if err := ValidateOrigin(o); err != nil {
			return nil, err
		}
		clean = append(clean, strings.TrimRight(o, "/"))
	}
	raw, err := json.Marshal(clean)
	if err != nil {
		return nil, err
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	defaultAgentID = strings.TrimSpace(defaultAgentID)
	var agent any
	if defaultAgentID == "" {
		agent = nil
	} else {
		agent = defaultAgentID
	}
	// Ensure row exists
	if _, err := s.GetOrCreateEmbedConfig(ctx, tenantID); err != nil {
		return nil, err
	}
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_embed_configs
SET enabled = $2, allowed_origins = $3::jsonb, default_agent_id = $4, updated_by = $5, updated_at = now()
WHERE tenant_id = $1`, schema), tenantID, enabled, raw, agent, actor)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrEmbedNotFound
	}
	return s.GetEmbedConfigByTenant(ctx, tenantID)
}

func (s *Store) UpdateEmbedAuthRequired(ctx context.Context, tenantID string, required bool) (*TenantEmbedConfig, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	if _, err := s.GetOrCreateEmbedConfig(ctx, tenantID); err != nil {
		return nil, err
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	if _, err := s.pg.Exec(ctx, fmt.Sprintf(`UPDATE %s.tenant_embed_configs SET auth_required=$2, updated_by=$3, updated_at=now() WHERE tenant_id=$1`, schema), tenantID, required, actor); err != nil {
		return nil, err
	}
	return s.GetEmbedConfigByTenant(ctx, tenantID)
}

// RotateEmbedKey issues a new public key; old key stops resolving.
func (s *Store) RotateEmbedKey(ctx context.Context, tenantID string) (*TenantEmbedConfig, error) {
	if _, err := s.GetOrCreateEmbedConfig(ctx, tenantID); err != nil {
		return nil, err
	}
	key, err := NewEmbedKey()
	if err != nil {
		return nil, err
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tag, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_embed_configs
SET embed_key = $2, updated_by = $3, updated_at = now()
WHERE tenant_id = $1`, schema), tenantID, key, actor)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() == 0 {
		return nil, ErrEmbedNotFound
	}
	return s.GetEmbedConfigByTenant(ctx, tenantID)
}

func (s *Store) scanEmbedConfig(ctx context.Context, q string, arg any) (*TenantEmbedConfig, error) {
	var c TenantEmbedConfig
	var originsRaw []byte
	err := s.pg.QueryRow(ctx, q, arg).Scan(
		&c.TenantID, &c.EmbedKey, &c.Enabled, &c.AuthRequired, &originsRaw, &c.DefaultAgentID, &c.CreatedAt, &c.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrEmbedNotFound
	}
	if err != nil {
		return nil, err
	}
	c.AllowedOrigins = []string{}
	if len(originsRaw) > 0 {
		_ = json.Unmarshal(originsRaw, &c.AllowedOrigins)
	}
	if c.AllowedOrigins == nil {
		c.AllowedOrigins = []string{}
	}
	return &c, nil
}

// ValidateOrigin requires absolute http(s) URL with host (no path required).
func ValidateOrigin(origin string) error {
	u, err := url.Parse(origin)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Errorf("invalid origin %q (use scheme://host[:port])", origin)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid origin scheme %q", u.Scheme)
	}
	return nil
}

// OriginAllowed returns true if requestOrigin matches allowlist.
// Empty allowlist: allowed when allowEmpty is true.
func OriginAllowed(allowlist []string, requestOrigin string, allowEmpty bool) bool {
	if len(allowlist) == 0 {
		return allowEmpty
	}
	req := normalizeOrigin(requestOrigin)
	if req == "" {
		return false
	}
	for _, o := range allowlist {
		if normalizeOrigin(o) == req {
			return true
		}
	}
	return false
}

// RequestOrigin extracts Origin header or scheme+host from Referer.
func RequestOrigin(originHeader, refererHeader string) string {
	if o := strings.TrimSpace(originHeader); o != "" {
		return strings.TrimRight(o, "/")
	}
	ref := strings.TrimSpace(refererHeader)
	if ref == "" {
		return ""
	}
	u, err := url.Parse(ref)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	return u.Scheme + "://" + u.Host
}

// ParseOrigin returns scheme://host[:port] for a valid http(s) origin, else "".
func ParseOrigin(o string) string {
	o = strings.TrimSpace(o)
	o = strings.TrimRight(o, "/")
	if o == "" {
		return ""
	}
	u, err := url.Parse(o)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return ""
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ""
	}
	// Reject path/query/fragment in explicit parent_origin claims.
	if u.Path != "" && u.Path != "/" {
		return ""
	}
	if u.RawQuery != "" || u.Fragment != "" {
		return ""
	}
	return strings.ToLower(u.Scheme) + "://" + strings.ToLower(u.Host)
}

// EmbedCheckOrigin is the origin used for allowlist checks on public resolve.
// Prefer parent_origin (host site, passed through the iframe) over the browser
// Origin/Referer of the iframe document (which is the Monti host).
func EmbedCheckOrigin(parentOrigin, originHeader, refererHeader string) string {
	if p := ParseOrigin(parentOrigin); p != "" {
		return p
	}
	return RequestOrigin(originHeader, refererHeader)
}

func normalizeOrigin(o string) string {
	if p := ParseOrigin(o); p != "" {
		return p
	}
	o = strings.TrimSpace(o)
	o = strings.TrimRight(o, "/")
	return strings.ToLower(o)
}
