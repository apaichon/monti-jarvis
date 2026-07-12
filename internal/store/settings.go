package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
)

var (
	ErrSettingsNotFound   = errors.New("tenant settings not found")
	ErrCallLimitsNotFound = errors.New("tenant call limits not found")
	ErrInvalidLocale      = errors.New("invalid locale")
	ErrInvalidTimezone    = errors.New("invalid timezone")
)

// TenantSettings is one row of callcenter.tenant_settings.
type TenantSettings struct {
	TenantID       string    `json:"tenant_id"`
	Locale         string    `json:"locale"`
	Timezone       string    `json:"timezone"`
	DisplayName    string    `json:"display_name,omitempty"`
	AIReplyLocale  string    `json:"ai_reply_locale,omitempty"`
	UserTierLabel  string    `json:"user_tier_label,omitempty"`
	UserGroupLabel string    `json:"user_group_label,omitempty"`
	CreatedAt      time.Time `json:"created_at,omitempty"`
	UpdatedAt      time.Time `json:"updated_at,omitempty"`
}

// TenantCallLimits is one row of callcenter.tenant_call_limits.
type TenantCallLimits struct {
	TenantID              string    `json:"tenant_id"`
	MaxMinutesPerCall     int       `json:"max_minutes_per_call"`
	MaxCallMinutesPerDay  int       `json:"max_call_minutes_per_day"`
	CreatedAt             time.Time `json:"created_at,omitempty"`
	UpdatedAt             time.Time `json:"updated_at,omitempty"`
}

func (s *Store) ensureSettingsSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmts := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.tenant_settings (
  tenant_id text PRIMARY KEY REFERENCES %s.tenants(id) ON DELETE CASCADE,
  locale text NOT NULL DEFAULT 'en',
  timezone text NOT NULL DEFAULT 'Asia/Bangkok',
  display_name text NOT NULL DEFAULT '',
  ai_reply_locale text NOT NULL DEFAULT '',
  user_tier_label text NOT NULL DEFAULT '',
  user_group_label text NOT NULL DEFAULT '',%s
)`, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.tenant_call_limits (
  tenant_id text PRIMARY KEY REFERENCES %s.tenants(id) ON DELETE CASCADE,
  max_minutes_per_call integer NOT NULL DEFAULT 0,
  max_call_minutes_per_day integer NOT NULL DEFAULT 0,%s
)`, schema, schema, auditColumnsDDL),
	}
	for _, stmt := range stmts {
		if _, err := s.pg.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("settings schema: %w", err)
		}
	}
	return nil
}

// NormalizeLocale returns en|th or ErrInvalidLocale.
func NormalizeLocale(v string) (string, error) {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return "en", nil
	}
	switch v {
	case "en", "th":
		return v, nil
	default:
		return "", ErrInvalidLocale
	}
}

// NormalizeOptionalLocale allows empty (auto) or en|th.
func NormalizeOptionalLocale(v string) (string, error) {
	v = strings.ToLower(strings.TrimSpace(v))
	if v == "" {
		return "", nil
	}
	return NormalizeLocale(v)
}

// ValidateTimezone checks IANA timezone name.
func ValidateTimezone(tz string) error {
	tz = strings.TrimSpace(tz)
	if tz == "" {
		return ErrInvalidTimezone
	}
	if _, err := time.LoadLocation(tz); err != nil {
		return ErrInvalidTimezone
	}
	return nil
}

func (s *Store) scanSettings(ctx context.Context, q string, arg any) (*TenantSettings, error) {
	var row TenantSettings
	err := s.pg.QueryRow(ctx, q, arg).Scan(
		&row.TenantID, &row.Locale, &row.Timezone, &row.DisplayName,
		&row.AIReplyLocale, &row.UserTierLabel, &row.UserGroupLabel,
		&row.CreatedAt, &row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrSettingsNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// GetTenantSettings returns settings or ErrSettingsNotFound.
func (s *Store) GetTenantSettings(ctx context.Context, tenantID string) (*TenantSettings, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	return s.scanSettings(ctx, fmt.Sprintf(`
SELECT tenant_id, locale, timezone, COALESCE(display_name, ''), COALESCE(ai_reply_locale, ''),
       COALESCE(user_tier_label, ''), COALESCE(user_group_label, ''), created_at, updated_at
FROM %s.tenant_settings WHERE tenant_id = $1`, schema), tenantID)
}

// GetOrCreateTenantSettings lazy-creates defaults.
func (s *Store) GetOrCreateTenantSettings(ctx context.Context, tenantID string) (*TenantSettings, error) {
	cfg, err := s.GetTenantSettings(ctx, tenantID)
	if err == nil {
		return cfg, nil
	}
	if !errors.Is(err, ErrSettingsNotFound) {
		return nil, err
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_settings (tenant_id, locale, timezone, created_by, updated_by)
VALUES ($1, 'en', 'Asia/Bangkok', $2, $2)
ON CONFLICT (tenant_id) DO NOTHING`, schema), tenantID, actor)
	if err != nil {
		return nil, err
	}
	return s.GetTenantSettings(ctx, tenantID)
}

// UpdateTenantSettingsInput is a partial update payload.
type UpdateTenantSettingsInput struct {
	Locale         *string
	Timezone       *string
	DisplayName    *string
	AIReplyLocale  *string
	UserTierLabel  *string
	UserGroupLabel *string
}

// UpdateTenantSettings applies validated fields.
func (s *Store) UpdateTenantSettings(ctx context.Context, tenantID string, in UpdateTenantSettingsInput) (*TenantSettings, error) {
	cur, err := s.GetOrCreateTenantSettings(ctx, tenantID)
	if err != nil {
		return nil, err
	}
	locale := cur.Locale
	if in.Locale != nil {
		locale, err = NormalizeLocale(*in.Locale)
		if err != nil {
			return nil, err
		}
	}
	tz := cur.Timezone
	if in.Timezone != nil {
		tz = strings.TrimSpace(*in.Timezone)
		if err := ValidateTimezone(tz); err != nil {
			return nil, err
		}
	}
	display := cur.DisplayName
	if in.DisplayName != nil {
		display = strings.TrimSpace(*in.DisplayName)
	}
	aiLocale := cur.AIReplyLocale
	if in.AIReplyLocale != nil {
		aiLocale, err = NormalizeOptionalLocale(*in.AIReplyLocale)
		if err != nil {
			return nil, err
		}
	}
	tier := cur.UserTierLabel
	if in.UserTierLabel != nil {
		tier = strings.TrimSpace(*in.UserTierLabel)
	}
	group := cur.UserGroupLabel
	if in.UserGroupLabel != nil {
		group = strings.TrimSpace(*in.UserGroupLabel)
	}

	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_settings
SET locale = $2, timezone = $3, display_name = $4, ai_reply_locale = $5,
    user_tier_label = $6, user_group_label = $7, updated_by = $8, updated_at = now()
WHERE tenant_id = $1`, schema),
		tenantID, locale, tz, display, aiLocale, tier, group, actor)
	if err != nil {
		return nil, err
	}
	return s.GetTenantSettings(ctx, tenantID)
}

// GetTenantCallLimits returns limits or ErrCallLimitsNotFound.
func (s *Store) GetTenantCallLimits(ctx context.Context, tenantID string) (*TenantCallLimits, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var row TenantCallLimits
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT tenant_id, max_minutes_per_call, max_call_minutes_per_day, created_at, updated_at
FROM %s.tenant_call_limits WHERE tenant_id = $1`, schema), tenantID).Scan(
		&row.TenantID, &row.MaxMinutesPerCall, &row.MaxCallMinutesPerDay, &row.CreatedAt, &row.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCallLimitsNotFound
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

// GetOrCreateTenantCallLimits lazy-creates zeros (unset).
func (s *Store) GetOrCreateTenantCallLimits(ctx context.Context, tenantID string) (*TenantCallLimits, error) {
	cfg, err := s.GetTenantCallLimits(ctx, tenantID)
	if err == nil {
		return cfg, nil
	}
	if !errors.Is(err, ErrCallLimitsNotFound) {
		return nil, err
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_call_limits (tenant_id, max_minutes_per_call, max_call_minutes_per_day, created_by, updated_by)
VALUES ($1, 0, 0, $2, $2)
ON CONFLICT (tenant_id) DO NOTHING`, schema), tenantID, actor)
	if err != nil {
		return nil, err
	}
	return s.GetTenantCallLimits(ctx, tenantID)
}

// UpdateTenantCallLimits sets caps (must be >= 0).
func (s *Store) UpdateTenantCallLimits(ctx context.Context, tenantID string, maxPerCall, maxPerDay int) (*TenantCallLimits, error) {
	if maxPerCall < 0 || maxPerDay < 0 {
		return nil, fmt.Errorf("call limits must be >= 0")
	}
	if _, err := s.GetOrCreateTenantCallLimits(ctx, tenantID); err != nil {
		return nil, err
	}
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_call_limits
SET max_minutes_per_call = $2, max_call_minutes_per_day = $3, updated_by = $4, updated_at = now()
WHERE tenant_id = $1`, schema), tenantID, maxPerCall, maxPerDay, actor)
	if err != nil {
		return nil, err
	}
	return s.GetTenantCallLimits(ctx, tenantID)
}

// AIReplyLocaleHint returns a system-prompt line for preferred reply language, or empty.
func (s *Store) AIReplyLocaleHint(ctx context.Context, tenantID string) string {
	if s == nil || s.pg == nil || strings.TrimSpace(tenantID) == "" {
		return ""
	}
	row, err := s.GetTenantSettings(ctx, tenantID)
	if err != nil {
		return ""
	}
	lang := strings.TrimSpace(row.AIReplyLocale)
	if lang == "" {
		lang = strings.TrimSpace(row.Locale)
	}
	switch lang {
	case "th":
		return "Reply and speak exclusively in Thai (ภาษาไทย) for the entire conversation. Do not mix English sentences or provide bilingual translations unless the caller asks. Product/brand names may stay in English."
	case "en":
		return "Reply and speak exclusively in English for the entire conversation. Do not mix Thai or other languages, and do not provide dual-language translations unless the caller asks."
	default:
		return ""
	}
}

// TenantTimezone returns IANA tz or Asia/Bangkok default.
func (s *Store) TenantTimezone(ctx context.Context, tenantID string) string {
	if s == nil || s.pg == nil || tenantID == "" {
		return "Asia/Bangkok"
	}
	row, err := s.GetTenantSettings(ctx, tenantID)
	if err != nil || strings.TrimSpace(row.Timezone) == "" {
		return "Asia/Bangkok"
	}
	if err := ValidateTimezone(row.Timezone); err != nil {
		return "Asia/Bangkok"
	}
	return row.Timezone
}
