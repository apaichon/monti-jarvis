package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
)

var (
	ErrEmailNotVerified     = errors.New("email not verified")
	ErrVerificationInvalid  = errors.New("verification token invalid")
	ErrVerificationExpired  = errors.New("verification token expired")
	ErrOAuthIdentityInUse   = errors.New("oauth identity already linked")
)

func (s *Store) ensureTenantAuthSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmts := []string{
		fmt.Sprintf(`ALTER TABLE %s.users ADD COLUMN IF NOT EXISTS auth_provider text NOT NULL DEFAULT 'email'`, schema),
		fmt.Sprintf(`ALTER TABLE %s.users ADD COLUMN IF NOT EXISTS email_verified_at timestamptz`, schema),
		fmt.Sprintf(`ALTER TABLE %s.users ALTER COLUMN password_hash DROP NOT NULL`, schema),
		fmt.Sprintf(`UPDATE %s.users SET email_verified_at = COALESCE(email_verified_at, now()) WHERE auth_provider = 'email'`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.user_oauth_identities (
  provider text NOT NULL CHECK (provider IN ('google', 'github')),
  provider_user_id text NOT NULL,
  user_id text NOT NULL REFERENCES %s.users(id) ON DELETE CASCADE,
  email text NOT NULL,%s,
  PRIMARY KEY (provider, provider_user_id)
)`, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS user_oauth_identities_user_idx ON %s.user_oauth_identities (user_id)`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.email_verification_tokens (
  id text PRIMARY KEY,
  user_id text NOT NULL REFERENCES %s.users(id) ON DELETE CASCADE,
  token_hash text NOT NULL UNIQUE,
  expires_at timestamptz NOT NULL,
  used_at timestamptz,%s
)`, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS email_verification_tokens_user_idx ON %s.email_verification_tokens (user_id)`, schema),
	}
	for _, stmt := range stmts {
		if _, err := s.pg.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) OAuthIdentityUserID(ctx context.Context, provider, providerUserID string) (string, error) {
	if s.pg == nil {
		return "", fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var userID string
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT user_id FROM %s.user_oauth_identities WHERE provider = $1 AND provider_user_id = $2`, schema),
		provider, providerUserID).Scan(&userID)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	return userID, err
}

func (s *Store) CreateEmailVerificationToken(ctx context.Context, userID string, ttl time.Duration) (string, error) {
	if s.pg == nil {
		return "", fmt.Errorf("postgres unavailable")
	}
	if ttl <= 0 {
		ttl = 24 * time.Hour
	}
	raw, err := newRandomToken()
	if err != nil {
		return "", err
	}
	hash := hashToken(raw)
	id := "evt_" + newStoreID()
	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.SystemActor
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.email_verification_tokens (id, user_id, token_hash, expires_at, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $5)`, schema),
		id, userID, hash, time.Now().UTC().Add(ttl), actor)
	if err != nil {
		return "", err
	}
	return raw, nil
}

func (s *Store) VerifyEmailToken(ctx context.Context, rawToken string) (AuthUser, error) {
	if s.pg == nil {
		return AuthUser{}, fmt.Errorf("postgres unavailable")
	}
	hash := hashToken(rawToken)
	schema := quoteIdent(s.cfg.PostgresSchema)
	var userID string
	var expiresAt time.Time
	var usedAt *time.Time
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT user_id, expires_at, used_at FROM %s.email_verification_tokens WHERE token_hash = $1`, schema), hash).
		Scan(&userID, &expiresAt, &usedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return AuthUser{}, ErrVerificationInvalid
	}
	if err != nil {
		return AuthUser{}, err
	}
	if usedAt != nil {
		return AuthUser{}, ErrVerificationInvalid
	}
	if time.Now().UTC().After(expiresAt) {
		return AuthUser{}, ErrVerificationExpired
	}

	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return AuthUser{}, err
	}
	defer tx.Rollback(ctx)

	actor := auditctx.SystemActor
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.email_verification_tokens SET used_at = now(), updated_by = $2 WHERE token_hash = $1`, schema), hash, actor)
	if err != nil {
		return AuthUser{}, err
	}
	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.users SET email_verified_at = now(), updated_by = $2 WHERE id = $1`, schema), userID, actor)
	if err != nil {
		return AuthUser{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return AuthUser{}, err
	}
	return s.GetUserByID(ctx, userID)
}

func (s *Store) IsEmailVerified(ctx context.Context, userID string) (bool, error) {
	if s.pg == nil {
		return false, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var verifiedAt *time.Time
	var provider string
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT email_verified_at, auth_provider FROM %s.users WHERE id = $1`, schema), userID).
		Scan(&verifiedAt, &provider)
	if errors.Is(err, pgx.ErrNoRows) {
		return false, ErrUserNotFound
	}
	if err != nil {
		return false, err
	}
	if provider != "email" {
		return true, nil
	}
	return verifiedAt != nil, nil
}

func newRandomToken() (string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(b[:]), nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}