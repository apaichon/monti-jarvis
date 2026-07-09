package store

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
)

var ErrUserNotFound = errors.New("user not found")

type AuthUser struct {
	ID              string
	Email           string
	PasswordHash    string
	DisplayName     string
	Status          string
	Role            string
	TenantID        string
	AuthProvider    string
	EmailVerifiedAt *time.Time
}

type RefreshTokenRow struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
}

func (s *Store) ensureAuthSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmts := []string{
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.tenants (
  id text PRIMARY KEY,
  slug text NOT NULL UNIQUE,
  name text NOT NULL,
  status text NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'suspended')),%s
)`, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.users (
  id text PRIMARY KEY,
  email text NOT NULL UNIQUE,
  password_hash text NOT NULL,
  display_name text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'disabled')),%s
)`, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.user_roles (
  user_id text NOT NULL REFERENCES %s.users(id) ON DELETE CASCADE,
  role text NOT NULL CHECK (role IN ('platform_admin', 'tenant_admin', 'customer')),
  tenant_id text REFERENCES %s.tenants(id) ON DELETE CASCADE,%s,
  PRIMARY KEY (user_id, role)
)`, schema, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.refresh_tokens (
  id text PRIMARY KEY,
  user_id text NOT NULL REFERENCES %s.users(id) ON DELETE CASCADE,
  token_hash text NOT NULL UNIQUE,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,%s
)`, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS refresh_tokens_user_idx ON %s.refresh_tokens (user_id)`, schema),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS refresh_tokens_expires_idx ON %s.refresh_tokens (expires_at)`, schema),
	}
	for _, stmt := range stmts {
		if _, err := s.pg.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	return s.seedAuthUsers(ctx)
}

func (s *Store) seedAuthUsers(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	demoTenant := s.cfg.DemoTenantID
	if demoTenant == "" {
		demoTenant = "demo"
	}

	const platformHash = `$2a$12$lQ5/HO3QPAqZQxC76Am3NOuV7U/UWpnFVvuUynx.ABDL/ZiiRGBSW`
	const demoAdminHash = `$2a$12$5n9IyLvIFmjwBczbeJJ1J.pHXTsnwRX4uADFrIol6xw2TLYsF9qq2`

	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenants (id, slug, name, status)
VALUES ($1, $2, $3, 'active')
ON CONFLICT (id) DO NOTHING`, schema), demoTenant, demoTenant, "Demo Tenant")
	if err != nil {
		return err
	}

	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.users (id, email, password_hash, display_name, status)
VALUES
  ('usr_platform', 'platform@monti.local', $1, 'Monti Platform', 'active'),
  ('usr_demo_admin', 'admin@demo.local', $2, 'Demo Admin', 'active')
ON CONFLICT (email) DO NOTHING`, schema), platformHash, demoAdminHash)
	if err != nil {
		return err
	}

	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.user_roles (user_id, role, tenant_id)
VALUES ('usr_platform', 'platform_admin', NULL)
ON CONFLICT (user_id, role) DO NOTHING`, schema))
	if err != nil {
		return err
	}

	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.user_roles (user_id, role, tenant_id)
VALUES ('usr_demo_admin', 'tenant_admin', $1)
ON CONFLICT (user_id, role) DO NOTHING`, schema), demoTenant)
	return err
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (AuthUser, error) {
	if s.pg == nil {
		return AuthUser{}, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	row := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT u.id, u.email, COALESCE(u.password_hash, ''), u.display_name, u.status,
       COALESCE(r.role, ''), COALESCE(r.tenant_id, ''),
       COALESCE(u.auth_provider, 'email'), u.email_verified_at
FROM %s.users u
LEFT JOIN %s.user_roles r ON r.user_id = u.id
WHERE lower(u.email) = lower($1)
ORDER BY CASE r.role
  WHEN 'platform_admin' THEN 1
  WHEN 'tenant_admin' THEN 2
  ELSE 3
END
LIMIT 1`, schema, schema), email)

	var user AuthUser
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.Status, &user.Role, &user.TenantID, &user.AuthProvider, &user.EmailVerifiedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AuthUser{}, ErrUserNotFound
		}
		return AuthUser{}, err
	}
	return user, nil
}

func (s *Store) GetUserByID(ctx context.Context, userID string) (AuthUser, error) {
	if s.pg == nil {
		return AuthUser{}, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	row := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT u.id, u.email, COALESCE(u.password_hash, ''), u.display_name, u.status,
       COALESCE(r.role, ''), COALESCE(r.tenant_id, ''),
       COALESCE(u.auth_provider, 'email'), u.email_verified_at
FROM %s.users u
LEFT JOIN %s.user_roles r ON r.user_id = u.id
WHERE u.id = $1
ORDER BY CASE r.role
  WHEN 'platform_admin' THEN 1
  WHEN 'tenant_admin' THEN 2
  ELSE 3
END
LIMIT 1`, schema, schema), userID)

	var user AuthUser
	if err := row.Scan(&user.ID, &user.Email, &user.PasswordHash, &user.DisplayName, &user.Status, &user.Role, &user.TenantID, &user.AuthProvider, &user.EmailVerifiedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AuthUser{}, ErrUserNotFound
		}
		return AuthUser{}, err
	}
	return user, nil
}

func (s *Store) SaveRefreshToken(ctx context.Context, id, userID, tokenHash string, expiresAt time.Time) error {
	if s.pg == nil {
		return fmt.Errorf("postgres unavailable")
	}
	actor := auditctx.ActorID(ctx)
	if actor == auditctx.SystemActor && userID != "" {
		actor = userID
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.refresh_tokens (id, user_id, token_hash, expires_at, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $5)`, schema), id, userID, tokenHash, expiresAt, actor)
	return err
}

func (s *Store) GetRefreshToken(ctx context.Context, tokenHash string) (RefreshTokenRow, error) {
	if s.pg == nil {
		return RefreshTokenRow{}, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	row := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id, user_id, token_hash, expires_at, revoked_at
FROM %s.refresh_tokens
WHERE token_hash = $1`, schema), tokenHash)

	var rt RefreshTokenRow
	if err := row.Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.RevokedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RefreshTokenRow{}, fmt.Errorf("refresh token not found")
		}
		return RefreshTokenRow{}, err
	}
	return rt, nil
}

func (s *Store) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	if s.pg == nil {
		return fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.ActorID(ctx)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.refresh_tokens
SET revoked_at = now(), updated_by = $2
WHERE token_hash = $1 AND revoked_at IS NULL`, schema), tokenHash, actor)
	return err
}