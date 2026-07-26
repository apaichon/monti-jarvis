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

const (
	ReferralCodeActive   = "active"
	ReferralCodeDisabled = "disabled"

	ReferralClicked    = "clicked"
	ReferralAttributed = "attributed"
	ReferralPending    = "pending"
	ReferralQualified  = "qualified"
	ReferralRejected   = "rejected"
	ReferralReversed   = "reversed"
)

var (
	ErrReferralNotFound          = errors.New("referral not found")
	ErrReferralInvalid           = errors.New("referral code is invalid")
	ErrReferralSelf              = errors.New("self-referral is not allowed")
	ErrReferralAlreadyAttributed = errors.New("tenant already has a referral attribution")
	ErrReferralNotQualified      = errors.New("referral is not yet qualified")
)

type TenantReferralCode struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenant_id"`
	Code      string    `json:"code"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ReferralClick struct {
	ID          string    `json:"id"`
	Code        string    `json:"code"`
	Source      string    `json:"source"`
	LandingPath string    `json:"landing_path"`
	ClickedAt   time.Time `json:"clicked_at"`
}

type TenantReferral struct {
	ID                  string     `json:"id"`
	ReferrerTenantID    string     `json:"referrer_tenant_id"`
	ReferredTenantID    string     `json:"referred_tenant_id"`
	Code                string     `json:"code"`
	Status              string     `json:"status"`
	Source              string     `json:"source"`
	QualificationReason string     `json:"qualification_reason,omitempty"`
	AttributedAt        *time.Time `json:"attributed_at,omitempty"`
	QualifiedAt         *time.Time `json:"qualified_at,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

func (s *Store) ensureReferralSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s.tenant_referral_codes (
  id text PRIMARY KEY,
  tenant_id text NOT NULL UNIQUE REFERENCES %s.tenants(id) ON DELETE CASCADE,
  code text NOT NULL UNIQUE,
  status text NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'disabled')),%s
);
CREATE TABLE IF NOT EXISTS %s.tenant_referrals (
  id text PRIMARY KEY,
  referrer_tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  referred_tenant_id text NOT NULL UNIQUE REFERENCES %s.tenants(id) ON DELETE CASCADE,
  referral_code_id text NOT NULL REFERENCES %s.tenant_referral_codes(id),
  code text NOT NULL,
  status text NOT NULL DEFAULT 'attributed'
    CHECK (status IN ('clicked', 'attributed', 'pending', 'qualified', 'rejected', 'reversed')),
  source text NOT NULL DEFAULT '',
  qualification_reason text NOT NULL DEFAULT '',
  attributed_at timestamptz,
  qualified_at timestamptz,%s
);
CREATE TABLE IF NOT EXISTS %s.tenant_referral_events (
  id text PRIMARY KEY,
  referral_id text NOT NULL REFERENCES %s.tenant_referrals(id) ON DELETE CASCADE,
  from_status text NOT NULL DEFAULT '',
  to_status text NOT NULL
    CHECK (to_status IN ('clicked', 'attributed', 'pending', 'qualified', 'rejected', 'reversed')),
  reason text NOT NULL DEFAULT '',
  source text NOT NULL DEFAULT '',
  event_at timestamptz NOT NULL DEFAULT now(),%s
);
CREATE INDEX IF NOT EXISTS tenant_referrals_referrer_idx
  ON %s.tenant_referrals (referrer_tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS tenant_referrals_status_idx
  ON %s.tenant_referrals (status, created_at DESC);
CREATE TABLE IF NOT EXISTS %s.tenant_referral_clicks (
  id text PRIMARY KEY,
  referral_code_id text NOT NULL REFERENCES %s.tenant_referral_codes(id) ON DELETE CASCADE,
  code text NOT NULL,
  source text NOT NULL DEFAULT '',
  landing_path text NOT NULL DEFAULT '',
  clicked_at timestamptz NOT NULL DEFAULT now(),%s
);`,
		schema, schema, auditColumnsDDL, schema, schema, schema, schema, auditColumnsDDL, schema, schema, auditColumnsDDL, schema, schema, schema, schema, auditColumnsDDL))
	return err
}

func (s *Store) RecordReferralClick(ctx context.Context, code, source, landingPath string) (ReferralClick, error) {
	if s == nil || s.pg == nil {
		return ReferralClick{}, fmt.Errorf("postgres unavailable")
	}
	code = normalizeReferralCode(code)
	if code == "" {
		return ReferralClick{}, ErrReferralInvalid
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var codeID, status string
	if err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT id, status FROM %s.tenant_referral_codes WHERE code=$1`, schema), code).Scan(&codeID, &status); errors.Is(err, pgx.ErrNoRows) {
		return ReferralClick{}, ErrReferralInvalid
	} else if err != nil {
		return ReferralClick{}, err
	} else if status != ReferralCodeActive {
		return ReferralClick{}, ErrReferralInvalid
	}
	var out ReferralClick
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_referral_clicks (id, referral_code_id, code, source, landing_path, created_by, updated_by)
VALUES ($1,$2,$3,$4,$5,$6,$6)
RETURNING id, code, source, landing_path, clicked_at`, schema), "refclick_"+newStoreID(), codeID, code, strings.TrimSpace(source), strings.TrimSpace(landingPath), auditctx.ActorID(ctx)).Scan(&out.ID, &out.Code, &out.Source, &out.LandingPath, &out.ClickedAt)
	return out, err
}

func normalizeReferralCode(code string) string {
	return strings.ToLower(strings.TrimSpace(code))
}

func (s *Store) GetOrCreateTenantReferralCode(ctx context.Context, tenantID string) (TenantReferralCode, error) {
	if s == nil || s.pg == nil {
		return TenantReferralCode{}, fmt.Errorf("postgres unavailable")
	}
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		return TenantReferralCode{}, ErrReferralInvalid
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.ActorID(ctx)
	code := "ref_" + newStoreID()
	var tenantStatus string
	if err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT status FROM %s.tenants WHERE id = $1`, schema), tenantID).Scan(&tenantStatus); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return TenantReferralCode{}, ErrReferralInvalid
		}
		return TenantReferralCode{}, err
	}
	if tenantStatus != "active" {
		return TenantReferralCode{}, ErrReferralInvalid
	}
	var out TenantReferralCode
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_referral_codes (id, tenant_id, code, created_by, updated_by)
VALUES ($1, $2, $3, $4, $4)
ON CONFLICT (tenant_id) DO UPDATE SET updated_by = EXCLUDED.updated_by
RETURNING id, tenant_id, code, status, created_at, updated_at`, schema),
		"refcode_"+newStoreID(), tenantID, code, actor).Scan(
		&out.ID, &out.TenantID, &out.Code, &out.Status, &out.CreatedAt, &out.UpdatedAt)
	if err != nil {
		return TenantReferralCode{}, err
	}
	return out, nil
}

func (s *Store) ListTenantReferrals(ctx context.Context, tenantID string) ([]TenantReferral, error) {
	if s == nil || s.pg == nil {
		return nil, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT id, referrer_tenant_id, referred_tenant_id, code, status, source,
       qualification_reason, attributed_at, qualified_at, created_at, updated_at
FROM %s.tenant_referrals
WHERE referrer_tenant_id = $1
ORDER BY created_at DESC`, schema), strings.TrimSpace(tenantID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := make([]TenantReferral, 0)
	for rows.Next() {
		var item TenantReferral
		if err := rows.Scan(&item.ID, &item.ReferrerTenantID, &item.ReferredTenantID, &item.Code,
			&item.Status, &item.Source, &item.QualificationReason, &item.AttributedAt,
			&item.QualifiedAt, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (s *Store) AttributeReferral(ctx context.Context, code, referredTenantID, source string) (TenantReferral, error) {
	if s == nil || s.pg == nil {
		return TenantReferral{}, fmt.Errorf("postgres unavailable")
	}
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return TenantReferral{}, err
	}
	defer tx.Rollback(ctx)
	item, err := s.attributeReferralTx(ctx, tx, code, referredTenantID, source)
	if err != nil {
		return TenantReferral{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return TenantReferral{}, err
	}
	return item, nil
}

func (s *Store) attributeReferralTx(ctx context.Context, tx pgx.Tx, code, referredTenantID, source string) (TenantReferral, error) {
	code = normalizeReferralCode(code)
	referredTenantID = strings.TrimSpace(referredTenantID)
	source = strings.TrimSpace(source)
	if code == "" || referredTenantID == "" {
		return TenantReferral{}, ErrReferralInvalid
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var codeID, referrerID, status string
	err := tx.QueryRow(ctx, fmt.Sprintf(`
SELECT id, tenant_id, status
FROM %s.tenant_referral_codes
WHERE code = $1
FOR UPDATE`, schema), code).Scan(&codeID, &referrerID, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantReferral{}, ErrReferralInvalid
	}
	if err != nil {
		return TenantReferral{}, err
	}
	if status != ReferralCodeActive {
		return TenantReferral{}, ErrReferralInvalid
	}
	if referrerID == referredTenantID {
		return TenantReferral{}, ErrReferralSelf
	}
	var tenantExists bool
	if err := tx.QueryRow(ctx, fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s.tenants WHERE id = $1)`, schema), referredTenantID).Scan(&tenantExists); err != nil {
		return TenantReferral{}, err
	}
	if !tenantExists {
		return TenantReferral{}, ErrReferralInvalid
	}

	var existing TenantReferral
	err = tx.QueryRow(ctx, fmt.Sprintf(`
SELECT id, referrer_tenant_id, referred_tenant_id, code, status, source,
       qualification_reason, attributed_at, qualified_at, created_at, updated_at
FROM %s.tenant_referrals
WHERE referred_tenant_id = $1
FOR UPDATE`, schema), referredTenantID).Scan(
		&existing.ID, &existing.ReferrerTenantID, &existing.ReferredTenantID, &existing.Code,
		&existing.Status, &existing.Source, &existing.QualificationReason, &existing.AttributedAt,
		&existing.QualifiedAt, &existing.CreatedAt, &existing.UpdatedAt)
	if err == nil {
		if existing.ReferrerTenantID == referrerID && existing.Code == code {
			return existing, nil
		}
		return TenantReferral{}, ErrReferralAlreadyAttributed
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return TenantReferral{}, err
	}
	var circular bool
	if err := tx.QueryRow(ctx, fmt.Sprintf(`
SELECT EXISTS(
  SELECT 1 FROM %s.tenant_referrals
  WHERE referrer_tenant_id = $1 AND referred_tenant_id = $2
)`, schema), referredTenantID, referrerID).Scan(&circular); err != nil {
		return TenantReferral{}, err
	}
	if circular {
		return TenantReferral{}, ErrReferralInvalid
	}

	actor := auditctx.ActorID(ctx)
	var item TenantReferral
	err = tx.QueryRow(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_referrals
  (id, referrer_tenant_id, referred_tenant_id, referral_code_id, code, status,
   source, attributed_at, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, now(), $8, $8)
RETURNING id, referrer_tenant_id, referred_tenant_id, code, status, source,
          qualification_reason, attributed_at, qualified_at, created_at, updated_at`, schema),
		"ref_"+newStoreID(), referrerID, referredTenantID, codeID, code, ReferralAttributed,
		source, actor).Scan(&item.ID, &item.ReferrerTenantID, &item.ReferredTenantID, &item.Code,
		&item.Status, &item.Source, &item.QualificationReason, &item.AttributedAt,
		&item.QualifiedAt, &item.CreatedAt, &item.UpdatedAt)
	if err != nil {
		return TenantReferral{}, err
	}
	if err := recordReferralEventTx(ctx, tx, schema, item.ID, "", ReferralAttributed, "", source); err != nil {
		return TenantReferral{}, err
	}
	return item, nil
}

func recordReferralEventTx(ctx context.Context, tx pgx.Tx, schema, referralID, fromStatus, toStatus, reason, source string) error {
	_, err := tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_referral_events
  (id, referral_id, from_status, to_status, reason, source, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $6, $7, $7)`, schema),
		"refevt_"+newStoreID(), referralID, fromStatus, toStatus, reason, source, auditctx.ActorID(ctx))
	return err
}

func (s *Store) GetReferral(ctx context.Context, referralID string) (TenantReferral, error) {
	if s == nil || s.pg == nil {
		return TenantReferral{}, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var item TenantReferral
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`
SELECT id, referrer_tenant_id, referred_tenant_id, code, status, source,
       qualification_reason, attributed_at, qualified_at, created_at, updated_at
FROM %s.tenant_referrals WHERE id = $1`, schema), strings.TrimSpace(referralID)).Scan(
		&item.ID, &item.ReferrerTenantID, &item.ReferredTenantID, &item.Code, &item.Status,
		&item.Source, &item.QualificationReason, &item.AttributedAt, &item.QualifiedAt,
		&item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantReferral{}, ErrReferralNotFound
	}
	if err != nil {
		return TenantReferral{}, err
	}
	return item, nil
}

func (s *Store) QualifyReferral(ctx context.Context, referralID string) (TenantReferral, error) {
	if s == nil || s.pg == nil {
		return TenantReferral{}, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return TenantReferral{}, err
	}
	defer tx.Rollback(ctx)

	var item TenantReferral
	err = tx.QueryRow(ctx, fmt.Sprintf(`
SELECT id, referrer_tenant_id, referred_tenant_id, code, status, source,
       qualification_reason, attributed_at, qualified_at, created_at, updated_at
FROM %s.tenant_referrals WHERE id = $1 FOR UPDATE`, schema), strings.TrimSpace(referralID)).Scan(
		&item.ID, &item.ReferrerTenantID, &item.ReferredTenantID, &item.Code, &item.Status,
		&item.Source, &item.QualificationReason, &item.AttributedAt, &item.QualifiedAt,
		&item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantReferral{}, ErrReferralNotFound
	}
	if err != nil {
		return TenantReferral{}, err
	}
	if item.Status == ReferralQualified {
		if err := tx.Commit(ctx); err != nil {
			return TenantReferral{}, err
		}
		return item, nil
	}
	if item.Status == ReferralRejected || item.Status == ReferralReversed {
		return item, ErrReferralNotQualified
	}

	var tenantStatus, kycStatus string
	var hasPaidOrder bool
	err = tx.QueryRow(ctx, fmt.Sprintf(`
SELECT t.status, COALESCE(k.status, ''), EXISTS(
  SELECT 1 FROM %s.payment_orders o
  WHERE o.tenant_id = t.id
    AND o.status = 'paid'
    AND NOT EXISTS (
      SELECT 1 FROM %s.payment_documents d
      WHERE d.order_id = o.id AND d.status = 'voided'
    )
)
FROM %s.tenants t
LEFT JOIN %s.tenant_kyc_profiles k ON k.tenant_id = t.id
WHERE t.id = $1`, schema, schema, schema, schema), item.ReferredTenantID).Scan(&tenantStatus, &kycStatus, &hasPaidOrder)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantReferral{}, ErrReferralNotFound
	}
	if err != nil {
		return TenantReferral{}, err
	}
	reason := ""
	switch {
	case tenantStatus != "active":
		reason = "tenant_not_active"
	case kycStatus != "approved":
		reason = "kyc_not_approved"
	case !hasPaidOrder:
		reason = "paid_order_required"
	}
	actor := auditctx.ActorID(ctx)
	previousStatus := item.Status
	previousReason := item.QualificationReason
	if reason != "" {
		_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_referrals
SET status = $2, qualification_reason = $3, updated_by = $4
WHERE id = $1`, schema), item.ID, ReferralPending, reason, actor)
		if err != nil {
			return TenantReferral{}, err
		}
		item.Status = ReferralPending
		item.QualificationReason = reason
		if previousStatus != ReferralPending || previousReason != reason {
			if err := recordReferralEventTx(ctx, tx, schema, item.ID, previousStatus, ReferralPending, reason, "qualification"); err != nil {
				return TenantReferral{}, err
			}
		}
		if err := tx.Commit(ctx); err != nil {
			return TenantReferral{}, err
		}
		return item, ErrReferralNotQualified
	}

	_, err = tx.Exec(ctx, fmt.Sprintf(`
UPDATE %s.tenant_referrals
SET status = $2, qualification_reason = '', qualified_at = COALESCE(qualified_at, now()), updated_by = $3
WHERE id = $1`, schema), item.ID, ReferralQualified, actor)
	if err != nil {
		return TenantReferral{}, err
	}
	item.Status = ReferralQualified
	item.QualificationReason = ""
	if err := recordReferralEventTx(ctx, tx, schema, item.ID, previousStatus, ReferralQualified, "", "qualification"); err != nil {
		return TenantReferral{}, err
	}
	if err := grantReferralBonusesTx(ctx, tx, schema, item.ID, item.ReferrerTenantID); err != nil {
		return TenantReferral{}, err
	}
	if item.QualifiedAt == nil {
		now := time.Now().UTC()
		item.QualifiedAt = &now
	}
	if err := tx.Commit(ctx); err != nil {
		return TenantReferral{}, err
	}
	return item, nil
}

func (s *Store) ReverseReferral(ctx context.Context, referralID, reason string) (TenantReferral, error) {
	if s == nil || s.pg == nil {
		return TenantReferral{}, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return TenantReferral{}, err
	}
	defer tx.Rollback(ctx)
	var item TenantReferral
	err = tx.QueryRow(ctx, fmt.Sprintf(`
SELECT id, referrer_tenant_id, referred_tenant_id, code, status, source,
       qualification_reason, attributed_at, qualified_at, created_at, updated_at
FROM %s.tenant_referrals WHERE id=$1 FOR UPDATE`, schema), strings.TrimSpace(referralID)).Scan(
		&item.ID, &item.ReferrerTenantID, &item.ReferredTenantID, &item.Code, &item.Status,
		&item.Source, &item.QualificationReason, &item.AttributedAt, &item.QualifiedAt,
		&item.CreatedAt, &item.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return TenantReferral{}, ErrReferralNotFound
	}
	if err != nil {
		return TenantReferral{}, err
	}
	if item.Status == ReferralReversed {
		return item, tx.Commit(ctx)
	}
	if item.Status != ReferralQualified {
		return item, ErrReferralNotQualified
	}
	actor := auditctx.ActorID(ctx)
	reason = strings.TrimSpace(reason)
	if reason == "" {
		reason = "qualification reversed"
	}
	if _, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.tenant_referrals SET status=$2, qualification_reason=$3, updated_by=$4 WHERE id=$1`, schema), item.ID, ReferralReversed, reason, actor); err != nil {
		return TenantReferral{}, err
	}
	if err := reverseReferralBonusesTx(ctx, tx, schema, item.ID, item.ReferrerTenantID, reason); err != nil {
		return TenantReferral{}, err
	}
	if err := recordReferralEventTx(ctx, tx, schema, item.ID, item.Status, ReferralReversed, reason, "platform"); err != nil {
		return TenantReferral{}, err
	}
	item.Status = ReferralReversed
	item.QualificationReason = reason
	if err := tx.Commit(ctx); err != nil {
		return TenantReferral{}, err
	}
	return item, nil
}
