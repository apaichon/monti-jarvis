package store

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
)

var (
	ErrCustomerAuthDisabled   = errors.New("customer auth disabled")
	ErrCustomerAuthForbidden  = errors.New("customer auth forbidden")
	ErrOTPInvalid             = errors.New("otp invalid")
	ErrOTPExpired             = errors.New("otp expired")
	ErrCustomerSessionInvalid = errors.New("customer session invalid")
)

type CustomerAuthSettings struct {
	TenantID                 string    `json:"tenant_id"`
	Enabled                  bool      `json:"enabled"`
	AuthMode                 string    `json:"auth_mode"`
	AllowedDomains           []string  `json:"allowed_domains"`
	OTPTTLSeconds            int       `json:"otp_ttl_seconds"`
	SessionTTLSeconds        int       `json:"session_ttl_seconds"`
	RequireAuthForWorkforce  bool      `json:"require_auth_for_workforce"`
	CustomerDailyCallSeconds int       `json:"customer_daily_call_seconds"`
	CustomerMaxCallSeconds   int       `json:"customer_max_call_seconds"`
	CreatedAt                time.Time `json:"created_at,omitempty"`
	UpdatedAt                time.Time `json:"updated_at,omitempty"`
}

type CustomerAuthSettingsInput struct {
	Enabled                  *bool
	AuthMode                 string
	AllowedDomains           []string
	OTPTTLSeconds            int
	SessionTTLSeconds        int
	RequireAuthForWorkforce  *bool
	CustomerDailyCallSeconds *int
	CustomerMaxCallSeconds   *int
}

type CustomerOTPChallenge struct {
	ID         string         `json:"id"`
	TenantID   string         `json:"tenant_id"`
	Email      string         `json:"email"`
	CustomerID string         `json:"customer_id,omitempty"`
	Status     string         `json:"status"`
	Attempts   int            `json:"attempts"`
	ExpiresAt  time.Time      `json:"expires_at"`
	Metadata   map[string]any `json:"metadata,omitempty"`
	CreatedAt  time.Time      `json:"created_at,omitempty"`
}

type CustomerSession struct {
	ID               string     `json:"id"`
	TenantID         string     `json:"tenant_id"`
	CustomerID       string     `json:"customer_id"`
	RefreshTokenHash string     `json:"-"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
}

func (s *Store) ensureCustomerAuthSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmts := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.tenant_customer_auth_settings (
  tenant_id text PRIMARY KEY REFERENCES %s.tenants(id) ON DELETE CASCADE,
  enabled boolean NOT NULL DEFAULT false,
  auth_mode text NOT NULL DEFAULT 'optional' CHECK (auth_mode IN ('optional','required')),
  allowed_domains jsonb NOT NULL DEFAULT '[]'::jsonb,
  otp_ttl_seconds integer NOT NULL DEFAULT 600,
	session_ttl_seconds integer NOT NULL DEFAULT 604800,%s
)`, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`ALTER TABLE %s.tenant_customer_auth_settings ADD COLUMN IF NOT EXISTS require_auth_for_workforce boolean NOT NULL DEFAULT false`, schema),
		fmt.Sprintf(`ALTER TABLE %s.tenant_customer_auth_settings ADD COLUMN IF NOT EXISTS customer_daily_call_seconds integer NOT NULL DEFAULT 0`, schema),
		fmt.Sprintf(`ALTER TABLE %s.tenant_customer_auth_settings ADD COLUMN IF NOT EXISTS customer_max_call_seconds integer NOT NULL DEFAULT 0`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.customer_auth_identities (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  customer_id text NOT NULL REFERENCES %s.customers(id) ON DELETE CASCADE,
  provider text NOT NULL DEFAULT 'email_otp' CHECK (provider IN ('email_otp')),
  email_normalized text NOT NULL,
  verified_at timestamptz NOT NULL DEFAULT now(),%s
)`, schema, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS customer_auth_identities_email_uidx
ON %s.customer_auth_identities (tenant_id, provider, email_normalized)`, schema),
		fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS customer_auth_identities_customer_uidx
ON %s.customer_auth_identities (tenant_id, provider, customer_id)`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.customer_otp_challenges (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  email_normalized text NOT NULL,
  customer_id text REFERENCES %s.customers(id) ON DELETE SET NULL,
  code_hash text NOT NULL,
  status text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','verified','expired','blocked')),
  attempts integer NOT NULL DEFAULT 0,
  expires_at timestamptz NOT NULL,
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,%s
)`, schema, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS customer_otp_challenges_lookup_idx
ON %s.customer_otp_challenges (tenant_id, email_normalized, created_at DESC)`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.customer_sessions (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  customer_id text NOT NULL REFERENCES %s.customers(id) ON DELETE CASCADE,
  refresh_token_hash text NOT NULL UNIQUE,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,%s
)`, schema, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS customer_sessions_customer_idx
ON %s.customer_sessions (tenant_id, customer_id, expires_at DESC)`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.customer_auth_events (
  id bigserial PRIMARY KEY,
  tenant_id text NOT NULL,
  customer_id text NOT NULL DEFAULT '',
  email_normalized text NOT NULL DEFAULT '',
  event text NOT NULL,
  ip text NOT NULL DEFAULT '',
  user_agent text NOT NULL DEFAULT '',
  metadata jsonb NOT NULL DEFAULT '{}'::jsonb,%s
)`, schema, auditColumnsDDL),
	}
	for _, stmt := range stmts {
		if _, err := s.pg.Exec(ctx, stmt); err != nil {
			return fmt.Errorf("customer auth schema: %w", err)
		}
	}
	return nil
}

func defaultCustomerAuthSettings(tenantID string) CustomerAuthSettings {
	return CustomerAuthSettings{
		TenantID: tenantID, Enabled: false, AuthMode: "optional",
		AllowedDomains: []string{}, OTPTTLSeconds: 600, SessionTTLSeconds: int((7 * 24 * time.Hour).Seconds()),
		RequireAuthForWorkforce: false, CustomerDailyCallSeconds: 0, CustomerMaxCallSeconds: 0,
	}
}

func (s *Store) GetCustomerAuthSettings(ctx context.Context, tenantID string) (CustomerAuthSettings, error) {
	if s.pg == nil {
		return CustomerAuthSettings{}, fmt.Errorf("postgres is not available")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	out := defaultCustomerAuthSettings(tenantID)
	var domains []byte
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT tenant_id,enabled,auth_mode,allowed_domains,otp_ttl_seconds,session_ttl_seconds,
require_auth_for_workforce,customer_daily_call_seconds,customer_max_call_seconds,created_at,updated_at
FROM %s.tenant_customer_auth_settings WHERE tenant_id=$1`, schema), tenantID).Scan(
		&out.TenantID, &out.Enabled, &out.AuthMode, &domains, &out.OTPTTLSeconds, &out.SessionTTLSeconds,
		&out.RequireAuthForWorkforce, &out.CustomerDailyCallSeconds, &out.CustomerMaxCallSeconds, &out.CreatedAt, &out.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return out, nil
	}
	if err != nil {
		return CustomerAuthSettings{}, err
	}
	_ = json.Unmarshal(domains, &out.AllowedDomains)
	if out.AllowedDomains == nil {
		out.AllowedDomains = []string{}
	}
	return out, nil
}

func (s *Store) PutCustomerAuthSettings(ctx context.Context, tenantID string, in CustomerAuthSettingsInput) (CustomerAuthSettings, error) {
	cur, err := s.GetCustomerAuthSettings(ctx, tenantID)
	if err != nil {
		return CustomerAuthSettings{}, err
	}
	if in.Enabled != nil {
		cur.Enabled = *in.Enabled
	}
	if strings.TrimSpace(in.AuthMode) != "" {
		cur.AuthMode = strings.ToLower(strings.TrimSpace(in.AuthMode))
	}
	if cur.AuthMode != "optional" && cur.AuthMode != "required" {
		return CustomerAuthSettings{}, fmt.Errorf("auth_mode must be optional or required")
	}
	if in.OTPTTLSeconds > 0 {
		cur.OTPTTLSeconds = in.OTPTTLSeconds
	}
	if cur.OTPTTLSeconds < 60 || cur.OTPTTLSeconds > 1800 {
		return CustomerAuthSettings{}, fmt.Errorf("otp_ttl_seconds must be between 60 and 1800")
	}
	if in.SessionTTLSeconds > 0 {
		cur.SessionTTLSeconds = in.SessionTTLSeconds
	}
	if cur.SessionTTLSeconds < 3600 || cur.SessionTTLSeconds > int((30*24*time.Hour).Seconds()) {
		return CustomerAuthSettings{}, fmt.Errorf("session_ttl_seconds must be between 3600 and 2592000")
	}
	if in.RequireAuthForWorkforce != nil {
		cur.RequireAuthForWorkforce = *in.RequireAuthForWorkforce
	}
	if in.CustomerDailyCallSeconds != nil {
		cur.CustomerDailyCallSeconds = *in.CustomerDailyCallSeconds
	}
	if in.CustomerMaxCallSeconds != nil {
		cur.CustomerMaxCallSeconds = *in.CustomerMaxCallSeconds
	}
	if cur.CustomerDailyCallSeconds < 0 || cur.CustomerDailyCallSeconds > 24*60*60 {
		return CustomerAuthSettings{}, fmt.Errorf("customer_daily_call_seconds must be between 0 and 86400")
	}
	if cur.CustomerMaxCallSeconds < 0 || cur.CustomerMaxCallSeconds > 24*60*60 {
		return CustomerAuthSettings{}, fmt.Errorf("customer_max_call_seconds must be between 0 and 86400")
	}
	domains := make([]string, 0, len(in.AllowedDomains))
	for _, item := range in.AllowedDomains {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		domain, err := NormalizeCustomerDomain(item)
		if err != nil {
			return CustomerAuthSettings{}, err
		}
		domains = append(domains, domain)
	}
	if in.AllowedDomains != nil {
		cur.AllowedDomains = domains
	}
	raw, _ := json.Marshal(cur.AllowedDomains)
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.tenant_customer_auth_settings
(tenant_id,enabled,auth_mode,allowed_domains,otp_ttl_seconds,session_ttl_seconds,
require_auth_for_workforce,customer_daily_call_seconds,customer_max_call_seconds,created_by,updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$10)
ON CONFLICT (tenant_id) DO UPDATE SET enabled=EXCLUDED.enabled,auth_mode=EXCLUDED.auth_mode,
allowed_domains=EXCLUDED.allowed_domains,otp_ttl_seconds=EXCLUDED.otp_ttl_seconds,
session_ttl_seconds=EXCLUDED.session_ttl_seconds,
require_auth_for_workforce=EXCLUDED.require_auth_for_workforce,
customer_daily_call_seconds=EXCLUDED.customer_daily_call_seconds,
customer_max_call_seconds=EXCLUDED.customer_max_call_seconds,
updated_by=EXCLUDED.updated_by,updated_at=now()`, schema),
		tenantID, cur.Enabled, cur.AuthMode, raw, cur.OTPTTLSeconds, cur.SessionTTLSeconds,
		cur.RequireAuthForWorkforce, cur.CustomerDailyCallSeconds, cur.CustomerMaxCallSeconds, actor)
	if err != nil {
		return CustomerAuthSettings{}, err
	}
	return s.GetCustomerAuthSettings(ctx, tenantID)
}

func (s *Store) FindCustomerByEmail(ctx context.Context, tenantID, email string) (*Customer, error) {
	email, err := NormalizeCustomerEmail(email)
	if err != nil {
		return nil, err
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	c, err := s.scanCustomer(ctx, tenantID, s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT id,tenant_id,email,email_normalized,phone,display_name,locale,tier_id,source,external_id,status,metadata,created_at,updated_at FROM %s.customers WHERE tenant_id=$1 AND email_normalized=$2`, schema), tenantID, email))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrCustomerNotFound
	}
	return c, err
}

func (s *Store) CreateCustomerOTPChallenge(ctx context.Context, tenantID, email, customerID, codeHash string, ttl time.Duration, metadata map[string]any) (CustomerOTPChallenge, error) {
	if ttl <= 0 {
		ttl = 10 * time.Minute
	}
	email, err := NormalizeCustomerEmail(email)
	if err != nil {
		return CustomerOTPChallenge{}, err
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	raw, _ := json.Marshal(metadata)
	id := "otp_" + newStoreID()
	actor := auditctx.ActorID(ctx)
	expiresAt := time.Now().UTC().Add(ttl)
	schema := quoteIdent(s.cfg.PostgresSchema)
	var customerArg any
	if strings.TrimSpace(customerID) != "" {
		customerArg = strings.TrimSpace(customerID)
	}
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.customer_otp_challenges
(id,tenant_id,email_normalized,customer_id,code_hash,expires_at,metadata,created_by,updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8)`, schema), id, tenantID, email, customerArg, codeHash, expiresAt, raw, actor)
	if err != nil {
		return CustomerOTPChallenge{}, err
	}
	return s.GetCustomerOTPChallenge(ctx, tenantID, id)
}

func (s *Store) GetCustomerOTPChallenge(ctx context.Context, tenantID, id string) (CustomerOTPChallenge, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	var out CustomerOTPChallenge
	var customerID *string
	var raw []byte
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT id,tenant_id,email_normalized,customer_id,status,attempts,expires_at,metadata,created_at
FROM %s.customer_otp_challenges WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id).Scan(
		&out.ID, &out.TenantID, &out.Email, &customerID, &out.Status, &out.Attempts, &out.ExpiresAt, &raw, &out.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return CustomerOTPChallenge{}, ErrOTPInvalid
	}
	if err != nil {
		return CustomerOTPChallenge{}, err
	}
	if customerID != nil {
		out.CustomerID = *customerID
	}
	out.Metadata = map[string]any{}
	_ = json.Unmarshal(raw, &out.Metadata)
	return out, nil
}

func (s *Store) VerifyCustomerOTPChallenge(ctx context.Context, tenantID, id, codeHash string, maxAttempts int) (CustomerOTPChallenge, error) {
	if maxAttempts <= 0 {
		maxAttempts = 5
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	chal, err := s.GetCustomerOTPChallenge(ctx, tenantID, id)
	if err != nil {
		return CustomerOTPChallenge{}, err
	}
	if chal.Status != "pending" {
		return CustomerOTPChallenge{}, ErrOTPInvalid
	}
	if time.Now().UTC().After(chal.ExpiresAt) {
		_, _ = s.pg.Exec(ctx, fmt.Sprintf(`UPDATE %s.customer_otp_challenges SET status='expired',updated_at=now() WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id)
		return CustomerOTPChallenge{}, ErrOTPExpired
	}
	var storedHash string
	err = s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT code_hash FROM %s.customer_otp_challenges WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id).Scan(&storedHash)
	if err != nil {
		return CustomerOTPChallenge{}, err
	}
	if !constantHashEqual(storedHash, codeHash) {
		status := "pending"
		if chal.Attempts+1 >= maxAttempts {
			status = "blocked"
		}
		_, _ = s.pg.Exec(ctx, fmt.Sprintf(`UPDATE %s.customer_otp_challenges SET attempts=attempts+1,status=$3,updated_at=now() WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id, status)
		return CustomerOTPChallenge{}, ErrOTPInvalid
	}
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`UPDATE %s.customer_otp_challenges SET status='verified',updated_at=now() WHERE tenant_id=$1 AND id=$2`, schema), tenantID, id)
	if err != nil {
		return CustomerOTPChallenge{}, err
	}
	return s.GetCustomerOTPChallenge(ctx, tenantID, id)
}

func constantHashEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func (s *Store) UpsertCustomerAuthIdentity(ctx context.Context, tenantID, customerID, email string) error {
	email, err := NormalizeCustomerEmail(email)
	if err != nil {
		return err
	}
	id := "cai_" + newStoreID()
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.customer_auth_identities
(id,tenant_id,customer_id,provider,email_normalized,verified_at,created_by,updated_by)
VALUES ($1,$2,$3,'email_otp',$4,now(),$5,$5)
ON CONFLICT (tenant_id,provider,email_normalized) DO UPDATE SET
customer_id=EXCLUDED.customer_id,verified_at=now(),updated_by=EXCLUDED.updated_by,updated_at=now()`, schema),
		id, tenantID, customerID, email, actor)
	return err
}

func (s *Store) CreateCustomerSession(ctx context.Context, tenantID, customerID, refreshTokenHash string, ttl time.Duration) (CustomerSession, error) {
	if ttl <= 0 {
		ttl = 7 * 24 * time.Hour
	}
	id := "csess_" + newStoreID()
	actor := auditctx.ActorID(ctx)
	expiresAt := time.Now().UTC().Add(ttl)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.customer_sessions
(id,tenant_id,customer_id,refresh_token_hash,expires_at,created_by,updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$6)`, schema), id, tenantID, customerID, refreshTokenHash, expiresAt, actor)
	if err != nil {
		return CustomerSession{}, err
	}
	return s.GetCustomerSessionByRefreshHash(ctx, refreshTokenHash)
}

func (s *Store) GetCustomerSessionByRefreshHash(ctx context.Context, refreshTokenHash string) (CustomerSession, error) {
	schema := quoteIdent(s.cfg.PostgresSchema)
	var out CustomerSession
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT id,tenant_id,customer_id,refresh_token_hash,expires_at,revoked_at FROM %s.customer_sessions WHERE refresh_token_hash=$1`, schema), refreshTokenHash).
		Scan(&out.ID, &out.TenantID, &out.CustomerID, &out.RefreshTokenHash, &out.ExpiresAt, &out.RevokedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return CustomerSession{}, ErrCustomerSessionInvalid
	}
	if err != nil {
		return CustomerSession{}, err
	}
	if out.RevokedAt != nil || time.Now().UTC().After(out.ExpiresAt) {
		return CustomerSession{}, ErrCustomerSessionInvalid
	}
	return out, nil
}

func (s *Store) RevokeCustomerSessionByRefreshHash(ctx context.Context, refreshTokenHash string) error {
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`UPDATE %s.customer_sessions SET revoked_at=COALESCE(revoked_at,now()),updated_at=now() WHERE refresh_token_hash=$1`, schema), refreshTokenHash)
	return err
}

func (s *Store) RecordCustomerAuthEvent(ctx context.Context, tenantID, customerID, email, event, ip, userAgent string, metadata map[string]any) {
	if s == nil || s.pg == nil {
		return
	}
	if metadata == nil {
		metadata = map[string]any{}
	}
	raw, _ := json.Marshal(metadata)
	actor := auditctx.ActorID(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, _ = s.pg.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.customer_auth_events
(tenant_id,customer_id,email_normalized,event,ip,user_agent,metadata,created_by,updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8)`, schema), tenantID, customerID, email, event, ip, userAgent, raw, actor)
}
