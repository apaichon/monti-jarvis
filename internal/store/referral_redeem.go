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
	ErrReferralRedeemSelf      = errors.New("self_referral")
	ErrReferralAlreadyRedeemed = errors.New("already_redeemed")
	ErrReferralIneligible      = errors.New("referral_ineligible")
)

// ReferralRedemption is a redeemer-side code apply record (S62).
type ReferralRedemption struct {
	ID               string         `json:"id"`
	RedeemerTenantID string         `json:"redeemer_tenant_id"`
	ReferrerTenantID string         `json:"referrer_tenant_id"`
	ReferralCode     string         `json:"referral_code"`
	Status           string         `json:"status"`
	IdempotencyKey   string         `json:"idempotency_key,omitempty"`
	AppliedAt        time.Time      `json:"applied_at"`
	ReversedAt       *time.Time     `json:"reversed_at,omitempty"`
	Bonus            []BonusBalance `json:"bonus,omitempty"`
}

func (s *Store) ensureReferralRedeemSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s.referral_redemptions (
  id text PRIMARY KEY,
  redeemer_tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  referrer_tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  referral_code text NOT NULL,
  status text NOT NULL DEFAULT 'applied' CHECK (status IN ('applied','reversed')),
  idempotency_key text NOT NULL DEFAULT '',
  applied_at timestamptz NOT NULL DEFAULT now(),
  reversed_at timestamptz,
  %s
);
CREATE UNIQUE INDEX IF NOT EXISTS referral_redemptions_redeemer_code_applied_uidx
  ON %s.referral_redemptions (redeemer_tenant_id, referral_code)
  WHERE status = 'applied';
CREATE INDEX IF NOT EXISTS referral_redemptions_referrer_idx
  ON %s.referral_redemptions (referrer_tenant_id, applied_at DESC);
`, schema, schema, schema, auditColumnsDDL, schema, schema))
	return err
}

// ValidateReferralCodeForRedeem dry-runs eligibility for the redeemer.
func (s *Store) ValidateReferralCodeForRedeem(ctx context.Context, redeemerTenantID, code string) (referrerTenantID string, preview []BonusBalance, err error) {
	code = normalizeReferralCode(code)
	if code == "" || strings.TrimSpace(redeemerTenantID) == "" {
		return "", nil, ErrReferralInvalid
	}
	if s.pg == nil {
		return "", nil, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var status, ownerTenant string
	err = s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT tenant_id, status FROM %s.tenant_referral_codes WHERE code=$1`, schema), code).Scan(&ownerTenant, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil, ErrReferralNotFound
	}
	if err != nil {
		return "", nil, err
	}
	if status != ReferralCodeActive {
		return "", nil, ErrReferralIneligible
	}
	if ownerTenant == redeemerTenantID {
		return "", nil, ErrReferralRedeemSelf
	}
	var applied bool
	if err := s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM %s.referral_redemptions WHERE redeemer_tenant_id=$1 AND referral_code=$2 AND status='applied')`, schema), redeemerTenantID, code).Scan(&applied); err != nil {
		return "", nil, err
	}
	if applied {
		return "", nil, ErrReferralAlreadyRedeemed
	}
	rules, err := s.ListReferralRewardRules(ctx)
	if err != nil {
		return "", nil, err
	}
	for _, r := range rules {
		if !r.Active || r.GrantAmount <= 0 {
			continue
		}
		preview = append(preview, BonusBalance{Dimension: r.Dimension, Unit: bonusUnit(r.Dimension), Remaining: r.GrantAmount, Granted: r.GrantAmount})
	}
	return ownerTenant, preview, nil
}

// RedeemReferralCode applies a code for the redeemer and grants bonus quota (idempotent).
func (s *Store) RedeemReferralCode(ctx context.Context, redeemerTenantID, code, idempotencyKey string) (ReferralRedemption, error) {
	code = normalizeReferralCode(code)
	idempotencyKey = strings.TrimSpace(idempotencyKey)
	if code == "" || strings.TrimSpace(redeemerTenantID) == "" {
		return ReferralRedemption{}, ErrReferralInvalid
	}
	if s.pg == nil {
		return ReferralRedemption{}, fmt.Errorf("postgres unavailable")
	}
	if err := s.ensureReferralRedeemSchema(ctx); err != nil {
		return ReferralRedemption{}, err
	}
	if err := s.ensureBonusSchema(ctx); err != nil {
		return ReferralRedemption{}, err
	}

	schema := quoteIdent(s.cfg.PostgresSchema)
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return ReferralRedemption{}, err
	}
	defer tx.Rollback(ctx)

	// Existing applied redemption → return it
	var existing ReferralRedemption
	err = tx.QueryRow(ctx, fmt.Sprintf(`
SELECT id, redeemer_tenant_id, referrer_tenant_id, referral_code, status, idempotency_key, applied_at, reversed_at
FROM %s.referral_redemptions WHERE redeemer_tenant_id=$1 AND referral_code=$2 AND status='applied'`, schema), redeemerTenantID, code).Scan(
		&existing.ID, &existing.RedeemerTenantID, &existing.ReferrerTenantID, &existing.ReferralCode,
		&existing.Status, &existing.IdempotencyKey, &existing.AppliedAt, &existing.ReversedAt)
	if err == nil {
		if err := tx.Commit(ctx); err != nil {
			return ReferralRedemption{}, err
		}
		balances, _ := s.ListTenantBonusBalances(ctx, redeemerTenantID)
		existing.Bonus = balances
		return existing, nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return ReferralRedemption{}, err
	}

	var ownerTenant, status string
	err = tx.QueryRow(ctx, fmt.Sprintf(`SELECT tenant_id, status FROM %s.tenant_referral_codes WHERE code=$1 FOR UPDATE`, schema), code).Scan(&ownerTenant, &status)
	if errors.Is(err, pgx.ErrNoRows) {
		return ReferralRedemption{}, ErrReferralNotFound
	}
	if err != nil {
		return ReferralRedemption{}, err
	}
	if status != ReferralCodeActive {
		return ReferralRedemption{}, ErrReferralIneligible
	}
	if ownerTenant == redeemerTenantID {
		return ReferralRedemption{}, ErrReferralRedeemSelf
	}

	redemptionID := "red_" + newStoreID()
	actor := auditctx.ActorID(ctx)
	_, err = tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.referral_redemptions (id, redeemer_tenant_id, referrer_tenant_id, referral_code, status, idempotency_key, applied_at, created_by, updated_by)
VALUES ($1,$2,$3,$4,'applied',$5,now(),$6,$6)`, schema), redemptionID, redeemerTenantID, ownerTenant, code, idempotencyKey, actor)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate") || strings.Contains(err.Error(), "unique") {
			return ReferralRedemption{}, ErrReferralAlreadyRedeemed
		}
		return ReferralRedemption{}, err
	}

	// Grant bonuses to redeemer with source referral_redeem
	rows, err := tx.Query(ctx, fmt.Sprintf(`SELECT dimension, grant_amount, expiry_days FROM %s.referral_reward_rules WHERE active=true AND grant_amount > 0`, schema))
	if err != nil {
		return ReferralRedemption{}, err
	}
	type reward struct {
		dimension  string
		amount     int64
		expiryDays int
	}
	var rewards []reward
	for rows.Next() {
		var item reward
		if err := rows.Scan(&item.dimension, &item.amount, &item.expiryDays); err != nil {
			rows.Close()
			return ReferralRedemption{}, err
		}
		rewards = append(rewards, item)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return ReferralRedemption{}, err
	}
	for _, item := range rewards {
		var expiresAt *time.Time
		if item.expiryDays > 0 {
			t := time.Now().UTC().Add(time.Duration(item.expiryDays) * 24 * time.Hour)
			expiresAt = &t
		}
		_, err := tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_bonus_ledger (id, tenant_id, referral_id, dimension, operation, amount, idempotency_key, expires_at, source_type, source_id, reason, created_by, updated_by)
VALUES ($1,$2,NULL,$3,'grant',$4,$5,$6,'referral_redeem',$7,'redeemed referral code bonus',$8,$8)
ON CONFLICT (tenant_id, idempotency_key) DO NOTHING`, schema),
			"bonus_"+newStoreID(), redeemerTenantID, item.dimension, item.amount,
			"redeem:"+redemptionID+":grant:"+item.dimension, expiresAt, redemptionID, actor)
		if err != nil {
			return ReferralRedemption{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return ReferralRedemption{}, err
	}
	out := ReferralRedemption{
		ID:               redemptionID,
		RedeemerTenantID: redeemerTenantID,
		ReferrerTenantID: ownerTenant,
		ReferralCode:     code,
		Status:           "applied",
		IdempotencyKey:   idempotencyKey,
		AppliedAt:        time.Now().UTC(),
	}
	balances, _ := s.ListTenantBonusBalances(ctx, redeemerTenantID)
	out.Bonus = balances
	return out, nil
}

// ListTenantRedemptions lists codes the tenant redeemed.
func (s *Store) ListTenantRedemptions(ctx context.Context, tenantID string) ([]ReferralRedemption, error) {
	if s.pg == nil {
		return nil, fmt.Errorf("postgres unavailable")
	}
	_ = s.ensureReferralRedeemSchema(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT id, redeemer_tenant_id, referrer_tenant_id, referral_code, status, idempotency_key, applied_at, reversed_at
FROM %s.referral_redemptions WHERE redeemer_tenant_id=$1 ORDER BY applied_at DESC`, schema), tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []ReferralRedemption
	for rows.Next() {
		var item ReferralRedemption
		if err := rows.Scan(&item.ID, &item.RedeemerTenantID, &item.ReferrerTenantID, &item.ReferralCode, &item.Status, &item.IdempotencyKey, &item.AppliedAt, &item.ReversedAt); err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

// ReverseReferralRedemption reverses bonus from a redemption (platform).
func (s *Store) ReverseReferralRedemption(ctx context.Context, redemptionID, reason string) (ReferralRedemption, error) {
	if s.pg == nil {
		return ReferralRedemption{}, fmt.Errorf("postgres unavailable")
	}
	_ = s.ensureReferralRedeemSchema(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return ReferralRedemption{}, err
	}
	defer tx.Rollback(ctx)

	var item ReferralRedemption
	err = tx.QueryRow(ctx, fmt.Sprintf(`
SELECT id, redeemer_tenant_id, referrer_tenant_id, referral_code, status, idempotency_key, applied_at, reversed_at
FROM %s.referral_redemptions WHERE id=$1 FOR UPDATE`, schema), strings.TrimSpace(redemptionID)).Scan(
		&item.ID, &item.RedeemerTenantID, &item.ReferrerTenantID, &item.ReferralCode, &item.Status, &item.IdempotencyKey, &item.AppliedAt, &item.ReversedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return ReferralRedemption{}, ErrReferralNotFound
	}
	if err != nil {
		return ReferralRedemption{}, err
	}
	if item.Status == "reversed" {
		if err := tx.Commit(ctx); err != nil {
			return ReferralRedemption{}, err
		}
		return item, nil
	}
	actor := auditctx.ActorID(ctx)
	// Reverse grants by source_id = redemptionID
	rows, err := tx.Query(ctx, fmt.Sprintf(`
SELECT dimension, COALESCE(SUM(amount),0) FROM %s.tenant_bonus_ledger
WHERE tenant_id=$1 AND source_type='referral_redeem' AND source_id=$2 AND operation='grant'
GROUP BY dimension`, schema), item.RedeemerTenantID, item.ID)
	if err != nil {
		return ReferralRedemption{}, err
	}
	type g struct {
		dim string
		amt int64
	}
	var grants []g
	for rows.Next() {
		var itemG g
		if err := rows.Scan(&itemG.dim, &itemG.amt); err != nil {
			rows.Close()
			return ReferralRedemption{}, err
		}
		grants = append(grants, itemG)
	}
	rows.Close()
	for _, gr := range grants {
		if gr.amt <= 0 {
			continue
		}
		_, err := tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_bonus_ledger (id, tenant_id, dimension, operation, amount, idempotency_key, source_type, source_id, reason, created_by, updated_by)
VALUES ($1,$2,$3,'reverse',$4,$5,'referral_redeem',$6,$7,$8,$8)
ON CONFLICT (tenant_id, idempotency_key) DO NOTHING`, schema),
			"bonus_"+newStoreID(), item.RedeemerTenantID, gr.dim, gr.amt,
			"redeem:"+item.ID+":reverse:"+gr.dim, item.ID, reason, actor)
		if err != nil {
			return ReferralRedemption{}, err
		}
	}
	_, err = tx.Exec(ctx, fmt.Sprintf(`UPDATE %s.referral_redemptions SET status='reversed', reversed_at=now(), updated_by=$2 WHERE id=$1`, schema), item.ID, actor)
	if err != nil {
		return ReferralRedemption{}, err
	}
	item.Status = "reversed"
	now := time.Now().UTC()
	item.ReversedAt = &now
	if err := tx.Commit(ctx); err != nil {
		return ReferralRedemption{}, err
	}
	return item, nil
}
