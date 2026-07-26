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
	BonusGrant     = "grant"
	BonusConsume   = "consume"
	BonusExpire    = "expire"
	BonusReverse   = "reverse"
	BonusReconcile = "reconcile"
)

var (
	ErrBonusInvalid     = errors.New("invalid bonus entitlement")
	ErrBonusUnavailable = errors.New("bonus entitlement unavailable")
)

// Bonus dimensions intentionally match the package and usage-ledger dimensions.
const (
	BonusAIEmployees        = "ai_employees"
	BonusMonthlyCallMinutes = "monthly_call_minutes"
	BonusMobileCallMinutes  = "mobile_call_minutes"
	BonusKMDocuments        = "km_documents"
	BonusStorageBytes       = "storage_bytes"
	BonusConcurrentCalls    = "concurrent_calls"
)

type ReferralRewardRule struct {
	Dimension   string    `json:"dimension"`
	GrantAmount int64     `json:"grant_amount"`
	ExpiryDays  int       `json:"expiry_days"`
	Active      bool      `json:"active"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type BonusLedgerEntry struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenant_id"`
	ReferralID     string     `json:"referral_id,omitempty"`
	Dimension      string     `json:"dimension"`
	Operation      string     `json:"operation"`
	Amount         int64      `json:"amount"`
	IdempotencyKey string     `json:"idempotency_key"`
	ExpiresAt      *time.Time `json:"expires_at,omitempty"`
	SourceType     string     `json:"source_type"`
	SourceID       string     `json:"source_id"`
	Reason         string     `json:"reason"`
	CreatedAt      time.Time  `json:"created_at"`
}

type BonusBalance struct {
	Dimension string     `json:"dimension"`
	Unit      string     `json:"unit"`
	Granted   int64      `json:"granted"`
	Used      int64      `json:"used"`
	Expired   int64      `json:"expired"`
	Reversed  int64      `json:"reversed"`
	Remaining int64      `json:"remaining"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

func validBonusDimension(dimension string) bool {
	switch strings.TrimSpace(dimension) {
	case BonusAIEmployees, BonusMonthlyCallMinutes, BonusMobileCallMinutes,
		BonusKMDocuments, BonusStorageBytes, BonusConcurrentCalls:
		return true
	default:
		return false
	}
}

func bonusUnit(dimension string) string {
	switch dimension {
	case BonusAIEmployees:
		return "assignments"
	case BonusMonthlyCallMinutes, BonusMobileCallMinutes:
		return "minutes"
	case BonusKMDocuments:
		return "documents"
	case BonusStorageBytes:
		return "bytes"
	case BonusConcurrentCalls:
		return "calls"
	default:
		return ""
	}
}

func defaultReferralRewardRules() []ReferralRewardRule {
	return []ReferralRewardRule{
		{Dimension: BonusAIEmployees, GrantAmount: 1, ExpiryDays: 90, Active: true},
		{Dimension: BonusMonthlyCallMinutes, GrantAmount: 60, ExpiryDays: 90, Active: true},
		{Dimension: BonusMobileCallMinutes, GrantAmount: 60, ExpiryDays: 90, Active: true},
		{Dimension: BonusKMDocuments, GrantAmount: 10, ExpiryDays: 90, Active: true},
		{Dimension: BonusStorageBytes, GrantAmount: 1073741824, ExpiryDays: 90, Active: true},
		{Dimension: BonusConcurrentCalls, GrantAmount: 1, ExpiryDays: 90, Active: true},
	}
}

func (s *Store) ensureBonusSchema(ctx context.Context) error {
	if s == nil || s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	if _, err := s.pg.Exec(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s.referral_reward_rules (
  dimension text PRIMARY KEY CHECK (dimension IN ('ai_employees','monthly_call_minutes','mobile_call_minutes','km_documents','storage_bytes','concurrent_calls')),
  grant_amount bigint NOT NULL CHECK (grant_amount >= 0),
  expiry_days integer NOT NULL DEFAULT 90 CHECK (expiry_days >= 0),
  active boolean NOT NULL DEFAULT true,%s
);
CREATE TABLE IF NOT EXISTS %s.tenant_bonus_ledger (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  referral_id text REFERENCES %s.tenant_referrals(id) ON DELETE CASCADE,
  dimension text NOT NULL CHECK (dimension IN ('ai_employees','monthly_call_minutes','mobile_call_minutes','km_documents','storage_bytes','concurrent_calls')),
  operation text NOT NULL CHECK (operation IN ('grant','consume','expire','reverse','reconcile')),
  amount bigint NOT NULL CHECK (amount > 0),
  idempotency_key text NOT NULL,
  expires_at timestamptz,
  source_type text NOT NULL DEFAULT '',
  source_id text NOT NULL DEFAULT '',
  reason text NOT NULL DEFAULT '',%s,
  UNIQUE (tenant_id, idempotency_key)
);
CREATE INDEX IF NOT EXISTS tenant_bonus_ledger_balance_idx
  ON %s.tenant_bonus_ledger (tenant_id, dimension, created_at);
CREATE INDEX IF NOT EXISTS tenant_bonus_ledger_referral_idx
  ON %s.tenant_bonus_ledger (referral_id, dimension);`, schema, auditColumnsDDL, schema, schema, schema, auditColumnsDDL, schema, schema)); err != nil {
		return err
	}
	for _, rule := range defaultReferralRewardRules() {
		if _, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.referral_reward_rules (dimension, grant_amount, expiry_days, active, created_by, updated_by)
VALUES ($1, $2, $3, $4, 'system', 'system') ON CONFLICT (dimension) DO NOTHING`, schema), rule.Dimension, rule.GrantAmount, rule.ExpiryDays, rule.Active); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) ListReferralRewardRules(ctx context.Context) ([]ReferralRewardRule, error) {
	if s == nil || s.pg == nil {
		return nil, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`SELECT dimension, grant_amount, expiry_days, active, updated_at FROM %s.referral_reward_rules ORDER BY dimension`, schema))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]ReferralRewardRule, 0)
	for rows.Next() {
		var rule ReferralRewardRule
		if err := rows.Scan(&rule.Dimension, &rule.GrantAmount, &rule.ExpiryDays, &rule.Active, &rule.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, rule)
	}
	return out, rows.Err()
}

func (s *Store) UpsertReferralRewardRule(ctx context.Context, rule ReferralRewardRule) (ReferralRewardRule, error) {
	rule.Dimension = strings.TrimSpace(rule.Dimension)
	if !validBonusDimension(rule.Dimension) || rule.GrantAmount < 0 || rule.ExpiryDays < 0 {
		return ReferralRewardRule{}, ErrBonusInvalid
	}
	if s == nil || s.pg == nil {
		return ReferralRewardRule{}, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	actor := auditctx.ActorID(ctx)
	return rule, s.pg.QueryRow(ctx, fmt.Sprintf(`
INSERT INTO %s.referral_reward_rules (dimension, grant_amount, expiry_days, active, created_by, updated_by)
VALUES ($1, $2, $3, $4, $5, $5)
ON CONFLICT (dimension) DO UPDATE SET grant_amount=EXCLUDED.grant_amount, expiry_days=EXCLUDED.expiry_days, active=EXCLUDED.active, updated_by=EXCLUDED.updated_by, updated_at=now()
RETURNING dimension, grant_amount, expiry_days, active, updated_at`, schema), rule.Dimension, rule.GrantAmount, rule.ExpiryDays, rule.Active, actor).Scan(&rule.Dimension, &rule.GrantAmount, &rule.ExpiryDays, &rule.Active, &rule.UpdatedAt)
}

func grantReferralBonusesTx(ctx context.Context, tx pgx.Tx, schema, referralID, tenantID string) error {
	rows, err := tx.Query(ctx, fmt.Sprintf(`SELECT dimension, grant_amount, expiry_days FROM %s.referral_reward_rules WHERE active=true AND grant_amount > 0`, schema))
	if err != nil {
		return err
	}
	type reward struct {
		dimension  string
		amount     int64
		expiryDays int
	}
	rewards := make([]reward, 0)
	for rows.Next() {
		var item reward
		if err := rows.Scan(&item.dimension, &item.amount, &item.expiryDays); err != nil {
			rows.Close()
			return err
		}
		rewards = append(rewards, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	actor := auditctx.ActorID(ctx)
	for _, item := range rewards {
		var expiresAt *time.Time
		if item.expiryDays > 0 {
			t := time.Now().UTC().Add(time.Duration(item.expiryDays) * 24 * time.Hour)
			expiresAt = &t
		}
		_, err := tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_bonus_ledger (id, tenant_id, referral_id, dimension, operation, amount, idempotency_key, expires_at, source_type, source_id, reason, created_by, updated_by)
VALUES ($1, $2, $3, $4, 'grant', $5, $6, $7, 'referral', $3, 'qualified referral reward', $8, $8)
		ON CONFLICT (tenant_id, idempotency_key) DO NOTHING`, schema), "bonus_"+newStoreID(), tenantID, referralID, item.dimension, item.amount, "referral:"+referralID+":grant:"+item.dimension, expiresAt, actor)
		if err != nil {
			return err
		}
	}
	return nil
}

func reverseReferralBonusesTx(ctx context.Context, tx pgx.Tx, schema, referralID, tenantID, reason string) error {
	rows, err := tx.Query(ctx, fmt.Sprintf(`
SELECT dimension, COALESCE(SUM(amount),0)
FROM %s.tenant_bonus_ledger
WHERE tenant_id=$1 AND referral_id=$2 AND operation='grant'
GROUP BY dimension`, schema), tenantID, referralID)
	if err != nil {
		return err
	}
	type grant struct {
		dimension string
		amount    int64
	}
	grants := make([]grant, 0)
	for rows.Next() {
		var item grant
		if err := rows.Scan(&item.dimension, &item.amount); err != nil {
			rows.Close()
			return err
		}
		grants = append(grants, item)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return err
	}
	rows.Close()
	actor := auditctx.ActorID(ctx)
	for _, item := range grants {
		if item.amount <= 0 {
			continue
		}
		_, err := tx.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_bonus_ledger (id, tenant_id, referral_id, dimension, operation, amount, idempotency_key, source_type, source_id, reason, created_by, updated_by)
VALUES ($1,$2,$3,$4,'reverse',$5,$6,'referral',$3,$7,$8,$8)
		ON CONFLICT (tenant_id,idempotency_key) DO NOTHING`, schema), "bonus_"+newStoreID(), tenantID, referralID, item.dimension, item.amount, "referral:"+referralID+":reverse:"+item.dimension, reason, actor)
		if err != nil {
			return err
		}
	}
	return nil
}

// ExpireBonusEntitlements materializes expiry events so the ledger remains
// auditable while balance reads remain deterministic and idempotent.
func (s *Store) ExpireBonusEntitlements(ctx context.Context) error {
	if s == nil || s.pg == nil {
		return fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.tenant_bonus_ledger (id, tenant_id, referral_id, dimension, operation, amount, idempotency_key, source_type, source_id, reason, created_by, updated_by)
SELECT 'bonus_' || g.id || '_expire', g.tenant_id, g.referral_id, g.dimension, 'expire', g.amount,
       g.id || ':expire', 'system', g.id, 'bonus grant expired', 'system', 'system'
FROM %s.tenant_bonus_ledger g
WHERE g.operation='grant' AND g.expires_at IS NOT NULL AND g.expires_at <= now()
  AND NOT EXISTS (SELECT 1 FROM %s.tenant_bonus_ledger e WHERE e.tenant_id=g.tenant_id AND e.idempotency_key=g.id || ':expire')
  AND NOT EXISTS (SELECT 1 FROM %s.tenant_bonus_ledger r WHERE r.tenant_id=g.tenant_id AND r.referral_id=g.referral_id AND r.dimension=g.dimension AND r.operation IN ('reverse','expire') AND r.source_id=g.id)`, schema, schema, schema, schema))
	return err
}

func (s *Store) ListTenantBonusBalances(ctx context.Context, tenantID string) ([]BonusBalance, error) {
	if s == nil || s.pg == nil {
		return nil, fmt.Errorf("postgres unavailable")
	}
	if strings.TrimSpace(tenantID) == "" {
		return nil, ErrBonusInvalid
	}
	_ = s.ExpireBonusEntitlements(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT dimension,
       CASE WHEN dimension IN ('monthly_call_minutes','mobile_call_minutes') THEN 'minutes'
            WHEN dimension='ai_employees' THEN 'assignments'
            WHEN dimension='km_documents' THEN 'documents'
            WHEN dimension='storage_bytes' THEN 'bytes' ELSE 'calls' END,
       COALESCE(SUM(amount) FILTER (WHERE operation='grant'),0),
       COALESCE(SUM(amount) FILTER (WHERE operation='consume'),0),
       COALESCE(SUM(amount) FILTER (WHERE operation='expire'),0),
       COALESCE(SUM(amount) FILTER (WHERE operation='reverse'),0),
       MIN(expires_at) FILTER (WHERE operation='grant' AND expires_at > now())
FROM %s.tenant_bonus_ledger WHERE tenant_id=$1 GROUP BY dimension ORDER BY dimension`, schema), tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]BonusBalance, 0)
	for rows.Next() {
		var b BonusBalance
		if err := rows.Scan(&b.Dimension, &b.Unit, &b.Granted, &b.Used, &b.Expired, &b.Reversed, &b.ExpiresAt); err != nil {
			return nil, err
		}
		activeGranted := b.Granted - b.Expired
		b.Remaining = activeGranted - b.Used - b.Reversed
		if b.Remaining < 0 {
			b.Remaining = 0
		}
		out = append(out, b)
	}
	return out, rows.Err()
}

func (s *Store) ListTenantBonusLedger(ctx context.Context, tenantID string) ([]BonusLedgerEntry, error) {
	if s == nil || s.pg == nil {
		return nil, fmt.Errorf("postgres unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`SELECT id, tenant_id, COALESCE(referral_id,''), dimension, operation, amount, idempotency_key, expires_at, source_type, source_id, reason, created_at FROM %s.tenant_bonus_ledger WHERE tenant_id=$1 ORDER BY created_at DESC`, schema), strings.TrimSpace(tenantID))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]BonusLedgerEntry, 0)
	for rows.Next() {
		var e BonusLedgerEntry
		if err := rows.Scan(&e.ID, &e.TenantID, &e.ReferralID, &e.Dimension, &e.Operation, &e.Amount, &e.IdempotencyKey, &e.ExpiresAt, &e.SourceType, &e.SourceID, &e.Reason, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

// ConsumeBonus uses a transaction-level tenant/dimension lock so concurrent
// call and upload completions cannot spend the same reward twice.
func (s *Store) ConsumeBonus(ctx context.Context, tenantID, dimension string, amount int64, idempotencyKey, sourceType, sourceID string) error {
	if s == nil || s.pg == nil {
		return fmt.Errorf("postgres unavailable")
	}
	tenantID, dimension, idempotencyKey = strings.TrimSpace(tenantID), strings.TrimSpace(dimension), strings.TrimSpace(idempotencyKey)
	if tenantID == "" || !validBonusDimension(dimension) || amount <= 0 || idempotencyKey == "" {
		return ErrBonusInvalid
	}
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	schema := quoteIdent(s.cfg.PostgresSchema)
	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, tenantID+":"+dimension); err != nil {
		return err
	}
	var existing int64
	err = tx.QueryRow(ctx, fmt.Sprintf(`SELECT amount FROM %s.tenant_bonus_ledger WHERE tenant_id=$1 AND idempotency_key=$2`, schema), tenantID, idempotencyKey).Scan(&existing)
	if err == nil {
		if existing != amount {
			return ErrBonusInvalid
		}
		return tx.Commit(ctx)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	var granted, used, expired, reversed int64
	if err := tx.QueryRow(ctx, fmt.Sprintf(`
SELECT COALESCE(SUM(amount) FILTER (WHERE operation='grant' AND (expires_at IS NULL OR expires_at > now())),0),
       COALESCE(SUM(amount) FILTER (WHERE operation='consume'),0),
       COALESCE(SUM(amount) FILTER (WHERE operation IN ('expire','reverse')),0)
FROM %s.tenant_bonus_ledger WHERE tenant_id=$1 AND dimension=$2`, schema), tenantID, dimension).Scan(&granted, &used, &expired); err != nil {
		return err
	}
	_ = reversed
	if granted-used-expired < amount {
		return ErrBonusUnavailable
	}
	_, err = tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.tenant_bonus_ledger (id, tenant_id, dimension, operation, amount, idempotency_key, source_type, source_id, reason, created_by, updated_by) VALUES ($1,$2,$3,'consume',$4,$5,$6,$7,'bonus quota consumed',$8,$8)`, schema), "bonus_"+newStoreID(), tenantID, dimension, amount, idempotencyKey, strings.TrimSpace(sourceType), strings.TrimSpace(sourceID), auditctx.ActorID(ctx))
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}
