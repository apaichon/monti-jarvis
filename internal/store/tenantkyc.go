package store

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/libra/monti-jarvis/internal/auditctx"
)

type TenantKYCProfile struct {
	TenantID         string
	ContactName      string
	ContactPhone     string
	ContactAddress   string
	PhotoObjectKey   string
	BusinessDocKeys  []string
	Status           string
	SubmittedAt      *time.Time
	ReviewedAt       *time.Time
	ReviewedBy       string
	RejectionReason  string
	UpdatedAt        time.Time
}

type PlatformKYCDecisionResult struct {
	TenantID           string
	TenantStatus       string
	RegistrationStatus string
	KYCStatus          string
	RejectionReason    string
	ReviewedAt         time.Time
	ReviewedBy         string
	AdminEmail         string
	CompanyName        string
}

func (s *Store) ensureTenantKYCSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	stmt := fmt.Sprintf(`CREATE TABLE IF NOT EXISTS %s.tenant_kyc_profiles (
  tenant_id text PRIMARY KEY REFERENCES %s.tenants(id) ON DELETE CASCADE,
  contact_name text NOT NULL DEFAULT '',
  contact_phone text NOT NULL DEFAULT '',
  contact_address text NOT NULL DEFAULT '',
  photo_object_key text NOT NULL DEFAULT '',
  business_doc_keys jsonb NOT NULL DEFAULT '[]'::jsonb,
  status text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'submitted')),
  submitted_at timestamptz,%s
)`, schema, schema, auditColumnsDDL)
	_, err := s.pg.Exec(ctx, stmt)
	if err != nil {
		return err
	}
	migrations := []string{
		fmt.Sprintf(`ALTER TABLE %s.tenant_kyc_profiles
  ADD COLUMN IF NOT EXISTS reviewed_at timestamptz`, schema),
		fmt.Sprintf(`ALTER TABLE %s.tenant_kyc_profiles
  ADD COLUMN IF NOT EXISTS reviewed_by text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.tenant_kyc_profiles
  ADD COLUMN IF NOT EXISTS rejection_reason text NOT NULL DEFAULT ''`, schema),
		fmt.Sprintf(`ALTER TABLE %s.tenant_kyc_profiles DROP CONSTRAINT IF EXISTS tenant_kyc_profiles_status_check`, schema),
		fmt.Sprintf(`ALTER TABLE %s.tenant_kyc_profiles ADD CONSTRAINT tenant_kyc_profiles_status_check
  CHECK (status IN ('draft', 'submitted', 'approved', 'rejected'))`, schema),
	}
	for _, m := range migrations {
		if _, err := s.pg.Exec(ctx, m); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) GetTenantKYCProfile(ctx context.Context, tenantID string) (TenantKYCProfile, error) {
	if s.pg == nil {
		return TenantKYCProfile{}, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	row := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT tenant_id, contact_name, contact_phone, contact_address, photo_object_key,
       business_doc_keys, status, submitted_at, reviewed_at, reviewed_by, rejection_reason, updated_at
FROM %s.tenant_kyc_profiles WHERE tenant_id = $1`, schema), tenantID)
	var profile TenantKYCProfile
	var docsRaw []byte
	if err := row.Scan(&profile.TenantID, &profile.ContactName, &profile.ContactPhone, &profile.ContactAddress,
		&profile.PhotoObjectKey, &docsRaw, &profile.Status, &profile.SubmittedAt,
		&profile.ReviewedAt, &profile.ReviewedBy, &profile.RejectionReason, &profile.UpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TenantKYCProfile{TenantID: tenantID, Status: "draft", BusinessDocKeys: []string{}}, nil
		}
		return TenantKYCProfile{}, err
	}
	_ = json.Unmarshal(docsRaw, &profile.BusinessDocKeys)
	if profile.BusinessDocKeys == nil {
		profile.BusinessDocKeys = []string{}
	}
	return profile, nil
}

func (s *Store) UpsertTenantKYCProfile(ctx context.Context, tenantID, contactName, contactPhone, contactAddress string) (TenantKYCProfile, error) {
	if s.pg == nil {
		return TenantKYCProfile{}, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.ActorID(ctx)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_kyc_profiles (tenant_id, contact_name, contact_phone, contact_address, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $5)
ON CONFLICT (tenant_id) DO UPDATE SET
  contact_name = EXCLUDED.contact_name,
  contact_phone = EXCLUDED.contact_phone,
  contact_address = EXCLUDED.contact_address,
  updated_by = EXCLUDED.updated_by,
  updated_at = now()`, schema),
		tenantID, strings.TrimSpace(contactName), strings.TrimSpace(contactPhone), strings.TrimSpace(contactAddress), actor)
	if err != nil {
		return TenantKYCProfile{}, err
	}
	return s.GetTenantKYCProfile(ctx, tenantID)
}

func (s *Store) SetTenantKYCPhoto(ctx context.Context, tenantID, objectKey string) error {
	if s.pg == nil {
		return fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.ActorID(ctx)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_kyc_profiles (tenant_id, photo_object_key, created_by, updated_by)
VALUES ($1, $2, $3, $3)
ON CONFLICT (tenant_id) DO UPDATE SET photo_object_key = EXCLUDED.photo_object_key, updated_by = EXCLUDED.updated_by, updated_at = now()`, schema),
		tenantID, objectKey, actor)
	return err
}

func (s *Store) AppendTenantKYCDocument(ctx context.Context, tenantID, objectKey string) error {
	if s.pg == nil {
		return fmt.Errorf("postgres unavailable")
	}
	profile, err := s.GetTenantKYCProfile(ctx, tenantID)
	if err != nil {
		return err
	}
	keys := append(profile.BusinessDocKeys, objectKey)
	raw, _ := json.Marshal(keys)
	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.ActorID(ctx)
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_kyc_profiles (tenant_id, business_doc_keys, created_by, updated_by)
VALUES ($1, $2::jsonb, $3, $3)
ON CONFLICT (tenant_id) DO UPDATE SET business_doc_keys = EXCLUDED.business_doc_keys, updated_by = EXCLUDED.updated_by, updated_at = now()`, schema),
		tenantID, string(raw), actor)
	return err
}

func (s *Store) SubmitTenantKYC(ctx context.Context, tenantID string) (TenantKYCProfile, error) {
	if s.pg == nil {
		return TenantKYCProfile{}, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.ActorID(ctx)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_kyc_profiles (tenant_id, status, submitted_at, created_by, updated_by)
VALUES ($1, 'submitted', now(), $2, $2)
ON CONFLICT (tenant_id) DO UPDATE SET
  status = 'submitted',
  submitted_at = now(),
  reviewed_at = NULL,
  reviewed_by = '',
  rejection_reason = '',
  updated_by = EXCLUDED.updated_by,
  updated_at = now()`, schema),
		tenantID, actor)
	if err != nil {
		return TenantKYCProfile{}, err
	}
	_, _ = s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_registrations
SET status = 'submitted', rejection_reason = '', reviewed_at = NULL, reviewed_by = '', updated_by = $2, updated_at = now()
WHERE tenant_id = $1`, schema), tenantID, actor)
	return s.GetTenantKYCProfile(ctx, tenantID)
}