package store

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
	"github.com/libra/monti-jarvis/internal/tenantregister"
)

var (
	ErrTenantSlugTaken       = errors.New("slug already taken")
	ErrTenantEmailRegistered = errors.New("email already registered")
	ErrTenantNotActive   = errors.New("tenant not active")
	ErrKYCReviewConflict = errors.New("kyc review conflict")
)

type TenantRegistration struct {
	ID          string
	TenantID    string
	CompanyName string
	AdminEmail  string
	Status      string
	CreatedAt   time.Time
}

type TenantListItem struct {
	ID             string
	Slug           string
	Name           string
	Status         string
	RegistrationID string
	AdminEmail     string
	KYCStatus      string
	CreatedAt      time.Time
}

type TenantSummary struct {
	ID        string
	Slug      string
	Name      string
	Status    string
	CreatedAt time.Time
}

type RegisterTenantInput struct {
	CompanyName         string
	Slug                string
	AdminEmail          string
	AdminPassword       string
	AdminDisplayName    string
	PasswordHash        string
	AuthProvider        string
	EmailVerified       bool
	OAuthProvider       string
	OAuthProviderUserID string
}

type RegisterTenantResult struct {
	TenantID       string
	Slug           string
	RegistrationID string
	UserID         string
	EmailVerified  bool
	AuthProvider   string
}

func (s *Store) ensureTenantRegisterSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmts := []string{
		fmt.Sprintf(`ALTER TABLE %s.tenants DROP CONSTRAINT IF EXISTS tenants_status_check`, schema),
		fmt.Sprintf(`ALTER TABLE %s.tenants ADD CONSTRAINT tenants_status_check
  CHECK (status IN ('pending_kyc', 'active', 'suspended'))`, schema),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.tenant_registrations (
  id text PRIMARY KEY,
  tenant_id text NOT NULL UNIQUE REFERENCES %s.tenants(id) ON DELETE CASCADE,
  company_name text NOT NULL,
  admin_email text NOT NULL,
  status text NOT NULL DEFAULT 'submitted'
    CHECK (status IN ('submitted', 'approved', 'rejected')),%s
)`, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.brands (
  id text PRIMARY KEY,
  tenant_id text NOT NULL UNIQUE REFERENCES %s.tenants(id) ON DELETE CASCADE,
  name text NOT NULL,
  status text NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'disabled')),%s
)`, schema, schema, auditColumnsDDL),
		fmt.Sprintf(`CREATE INDEX IF NOT EXISTS tenant_registrations_tenant_idx
ON %s.tenant_registrations (tenant_id)`, schema),
		fmt.Sprintf(`ALTER TABLE %s.tenant_registrations
  ADD COLUMN IF NOT EXISTS rejection_reason text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.tenant_registrations
  ADD COLUMN IF NOT EXISTS reviewed_at timestamptz`, schema),
		fmt.Sprintf(`ALTER TABLE %s.tenant_registrations
  ADD COLUMN IF NOT EXISTS reviewed_by text NOT NULL DEFAULT ''`, schema),
	}
	for _, stmt := range stmts {
		if _, err := s.pg.Exec(ctx, stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetTenant(ctx context.Context, tenantID string) (TenantSummary, error) {
	if s.pg == nil {
		return TenantSummary{}, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var t TenantSummary
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id, slug, name, status, created_at
FROM %s.tenants WHERE id = $1`, schema), tenantID).Scan(&t.ID, &t.Slug, &t.Name, &t.Status, &t.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TenantSummary{}, ErrTenantNotFound
		}
		return TenantSummary{}, err
	}
	return t, nil
}

func (s *Store) GetTenantRegistration(ctx context.Context, tenantID string) (TenantRegistration, error) {
	if s.pg == nil {
		return TenantRegistration{}, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var reg TenantRegistration
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id, tenant_id, company_name, admin_email, status, created_at
FROM %s.tenant_registrations WHERE tenant_id = $1`, schema), tenantID).Scan(
		&reg.ID, &reg.TenantID, &reg.CompanyName, &reg.AdminEmail, &reg.Status, &reg.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TenantRegistration{}, ErrTenantNotFound
		}
		return TenantRegistration{}, err
	}
	return reg, nil
}

func (s *Store) TenantSlugExists(ctx context.Context, slug string) (bool, error) {
	if s.pg == nil {
		return false, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var exists bool
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT EXISTS(SELECT 1 FROM %s.tenants WHERE slug = $1 OR id = $1)`, schema), slug).Scan(&exists)
	return exists, err
}

func (s *Store) UserEmailExists(ctx context.Context, email string) (bool, error) {
	if s.pg == nil {
		return false, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var exists bool
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT EXISTS(SELECT 1 FROM %s.users WHERE lower(email) = lower($1))`, schema), email).Scan(&exists)
	return exists, err
}

func (s *Store) IsTenantActive(ctx context.Context, tenantID string) (bool, error) {
	if s.pg == nil {
		return false, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var status string
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT status FROM %s.tenants WHERE id = $1`, schema), tenantID).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return status == "active", nil
}

func (s *Store) RegisterTenant(ctx context.Context, in RegisterTenantInput) (*RegisterTenantResult, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres unavailable")
	}
	normalized := tenantregister.Normalize(tenantregister.Input{
		CompanyName:      in.CompanyName,
		Slug:             in.Slug,
		AdminEmail:       in.AdminEmail,
		AdminPassword:    in.AdminPassword,
		AdminDisplayName: in.AdminDisplayName,
	})
	if in.AuthProvider == "email" || in.AuthProvider == "" {
		if err := tenantregister.Validate(normalized); err != nil {
			return nil, err
		}
	} else if err := tenantregister.ValidateProfile(normalized); err != nil {
		return nil, err
	}

	slug := normalized.Slug
	taken, err := s.TenantSlugExists(ctx, slug)
	if err != nil {
		return nil, err
	}
	if taken {
		return nil, ErrTenantSlugTaken
	}
	emailTaken, err := s.UserEmailExists(ctx, normalized.AdminEmail)
	if err != nil {
		return nil, err
	}
	if emailTaken {
		return nil, ErrTenantEmailRegistered
	}
	if in.OAuthProvider != "" && in.OAuthProviderUserID != "" {
		linked, err := s.OAuthIdentityUserID(ctx, in.OAuthProvider, in.OAuthProviderUserID)
		if err != nil {
			return nil, err
		}
		if linked != "" {
			return nil, ErrOAuthIdentityInUse
		}
	}

	regID := newRegistrationID()
	userID := fmt.Sprintf("usr_%s_admin", slug)
	brandID := fmt.Sprintf("brand_%s", slug)
	actor := auditctx.SystemActor
	schema := quoteIdent(s.cfg.PostgresSchema)
	authProvider := in.AuthProvider
	if authProvider == "" {
		authProvider = "email"
	}
	var emailVerifiedAt *time.Time
	if in.EmailVerified {
		now := time.Now().UTC()
		emailVerifiedAt = &now
	}

	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenants (id, slug, name, status, created_by, updated_by)
VALUES ($1, $2, $3, 'pending_kyc', $4, $4)`, schema),
		slug, slug, normalized.CompanyName, actor)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.brands (id, tenant_id, name, status, created_by, updated_by)
VALUES ($1, $2, $3, 'active', $4, $4)`, schema),
		brandID, slug, normalized.CompanyName, actor)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.users (id, email, password_hash, display_name, status, auth_provider, email_verified_at, created_by, updated_by)
VALUES ($1, $2, NULLIF($3, ''), $4, 'active', $5, $6, $7, $7)`, schema),
		userID, normalized.AdminEmail, in.PasswordHash, normalized.AdminDisplayName, authProvider, emailVerifiedAt, actor)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.user_roles (user_id, role, tenant_id, created_by, updated_by)
VALUES ($1, 'tenant_admin', $2, $3, $3)`, schema),
		userID, slug, actor)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_registrations (id, tenant_id, company_name, admin_email, status, created_by, updated_by)
VALUES ($1, $2, $3, $4, 'submitted', $5, $5)`, schema),
		regID, slug, normalized.CompanyName, normalized.AdminEmail, actor)
	if err != nil {
		return nil, err
	}

	if in.OAuthProvider != "" && in.OAuthProviderUserID != "" {
		_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.user_oauth_identities (provider, provider_user_id, user_id, email, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $5)`, schema),
			in.OAuthProvider, in.OAuthProviderUserID, userID, normalized.AdminEmail, actor)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &RegisterTenantResult{
		TenantID:       slug,
		Slug:           slug,
		RegistrationID: regID,
		UserID:         userID,
		EmailVerified:  in.EmailVerified,
		AuthProvider:   authProvider,
	}, nil
}

func (s *Store) ListTenants(ctx context.Context, status, kycStatus string, limit, offset int) ([]TenantListItem, int, error) {
	if s.pg == nil {
		return nil, 0, fmt.Errorf("postgres unavailable")
	}
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	schema := quoteIdent(s.cfg.PostgresSchema)
	status = strings.TrimSpace(status)
	kycStatus = strings.TrimSpace(kycStatus)

	where := make([]string, 0, 2)
	args := []any{}
	if status != "" {
		args = append(args, status)
		where = append(where, fmt.Sprintf("t.status = $%d", len(args)))
	}
	if kycStatus != "" {
		args = append(args, kycStatus)
		where = append(where, fmt.Sprintf("COALESCE(k.status, 'draft') = $%d", len(args)))
	}
	whereSQL := ""
	if len(where) > 0 {
		whereSQL = " WHERE " + strings.Join(where, " AND ")
	}

	fromJoin := fmt.Sprintf(`FROM %s.tenants t
LEFT JOIN %s.tenant_registrations tr ON tr.tenant_id = t.id
LEFT JOIN %s.tenant_kyc_profiles k ON k.tenant_id = t.id`, schema, schema, schema)

	var total int
	countQuery := "SELECT COUNT(*) " + fromJoin + whereSQL
	if err := s.pg.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
SELECT t.id, t.slug, t.name, t.status, COALESCE(tr.id, ''), COALESCE(tr.admin_email, ''),
       COALESCE(k.status, 'draft'), t.created_at
%s%s ORDER BY t.created_at DESC LIMIT $%d OFFSET $%d`,
		fromJoin, whereSQL, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := s.pg.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	out := make([]TenantListItem, 0)
	for rows.Next() {
		var item TenantListItem
		if err := rows.Scan(&item.ID, &item.Slug, &item.Name, &item.Status, &item.RegistrationID, &item.AdminEmail, &item.KYCStatus, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, item)
	}
	return out, total, rows.Err()
}

func newRegistrationID() string {
	return "reg_" + newStoreID()
}

func newStoreID() string {
	return fmt.Sprintf("%x", time.Now().UTC().UnixNano())[:16]
}