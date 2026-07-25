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

var (
	ErrUsageEventNotFound            = errors.New("usage event not found")
	ErrUsageReconciliationNotFound   = errors.New("usage reconciliation not found")
	ErrUsageValidation               = errors.New("invalid usage input")
	ErrUsageReconciliationInProgress = errors.New("usage reconciliation already running")
)

const (
	UsageStateApplied            = "applied"
	UsageStateCorrection         = "correction"
	UsageStateReversed           = "reversed"
	UsageStateVoided             = "voided"
	UsageReconciliationQueued    = "queued"
	UsageReconciliationRunning   = "running"
	UsageReconciliationCompleted = "completed"
	UsageReconciliationFailed    = "failed"
)

type UsageEvent struct {
	ID                    string    `json:"id"`
	TenantID              string    `json:"tenant_id"`
	IdempotencyKey        string    `json:"idempotency_key"`
	Dimension             string    `json:"dimension"`
	Unit                  string    `json:"unit"`
	Amount                int64     `json:"amount"`
	PeriodStart           time.Time `json:"period_start"`
	PeriodEnd             time.Time `json:"period_end"`
	SourceType            string    `json:"source_type"`
	SourceID              string    `json:"source_id"`
	EntitlementSnapshotID string    `json:"entitlement_snapshot_id"`
	State                 string    `json:"state"`
	CorrectionOf          string    `json:"correction_of"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
}

type UsageEventInput struct {
	TenantID, IdempotencyKey, Dimension, Unit                        string
	Amount                                                           int64
	PeriodStart, PeriodEnd                                           time.Time
	SourceType, SourceID, EntitlementSnapshotID, State, CorrectionOf string
}

type UsageReconciliationRun struct {
	ID               string         `json:"run_id"`
	TenantID         string         `json:"tenant_id,omitempty"`
	IdempotencyKey   string         `json:"idempotency_key"`
	StartDate        time.Time      `json:"start_date"`
	EndDate          time.Time      `json:"end_date"`
	DryRun           bool           `json:"dry_run"`
	Status           string         `json:"status"`
	SourceWatermarks map[string]any `json:"source_watermarks"`
	MismatchCount    int            `json:"mismatch_count"`
	CorrectionCount  int            `json:"correction_count"`
	ErrorCode        string         `json:"error_code,omitempty"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

type UsageReconciliationInput struct {
	TenantID, IdempotencyKey string
	StartDate, EndDate       time.Time
	DryRun                   bool
}

type UsageProjection struct {
	Dimension             string `json:"dimension"`
	Unit                  string `json:"unit"`
	Period                string `json:"period"`
	Consumed              int64  `json:"consumed"`
	Source                string `json:"source"`
	Freshness             string `json:"freshness"`
	EntitlementSnapshotID string `json:"entitlement_snapshot_id,omitempty"`
}

func (s *Store) ensureUsageSchema(ctx context.Context) error {
	if s == nil || s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s.usage_events (
  id text PRIMARY KEY, tenant_id text NOT NULL REFERENCES %s.tenants(id) ON DELETE CASCADE,
  idempotency_key text NOT NULL,
  dimension text NOT NULL CHECK (dimension IN ('ai_employees','monthly_call_minutes','mobile_call_minutes','km_documents','storage_bytes','concurrent_calls')),
  unit text NOT NULL CHECK (unit IN ('assignments','minutes','documents','bytes','calls')),
  amount bigint NOT NULL CHECK (amount > 0), period_start date NOT NULL, period_end date NOT NULL,
  source_type text NOT NULL, source_id text NOT NULL DEFAULT '', entitlement_snapshot_id text NOT NULL DEFAULT '',
  state text NOT NULL DEFAULT 'applied' CHECK (state IN ('applied','correction','reversed','voided')), correction_of text NOT NULL DEFAULT '',
  %s, UNIQUE (tenant_id, idempotency_key), CHECK (period_end >= period_start)
);
CREATE INDEX IF NOT EXISTS usage_events_period_idx ON %s.usage_events (tenant_id, period_start, period_end);
CREATE INDEX IF NOT EXISTS usage_events_source_idx ON %s.usage_events (source_type, source_id);
CREATE TABLE IF NOT EXISTS %s.usage_reconciliation_runs (
	  id text PRIMARY KEY, tenant_id text NOT NULL DEFAULT '', idempotency_key text NOT NULL,
  start_date date NOT NULL, end_date date NOT NULL, dry_run boolean NOT NULL DEFAULT true,
  status text NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','running','completed','failed')),
  source_watermarks jsonb NOT NULL DEFAULT '{}'::jsonb, mismatch_count integer NOT NULL DEFAULT 0,
  correction_count integer NOT NULL DEFAULT 0, error_code text NOT NULL DEFAULT '',
	  %s, CHECK (end_date >= start_date)
);`, schema, schema, auditColumnsDDL, schema, schema, schema, auditColumnsDDL))
	if err != nil {
		return err
	}
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`CREATE UNIQUE INDEX IF NOT EXISTS usage_reconciliation_runs_tenant_key_idx ON %s.usage_reconciliation_runs (tenant_id, idempotency_key);`, schema))
	return err
}

func normalizeUsageEvent(in UsageEventInput) (UsageEventInput, error) {
	in.TenantID, in.IdempotencyKey = strings.TrimSpace(in.TenantID), strings.TrimSpace(in.IdempotencyKey)
	in.Dimension, in.Unit = strings.TrimSpace(in.Dimension), strings.TrimSpace(in.Unit)
	in.SourceType, in.SourceID = strings.TrimSpace(in.SourceType), strings.TrimSpace(in.SourceID)
	in.EntitlementSnapshotID, in.State, in.CorrectionOf = strings.TrimSpace(in.EntitlementSnapshotID), strings.TrimSpace(in.State), strings.TrimSpace(in.CorrectionOf)
	if in.State == "" {
		in.State = UsageStateApplied
	}
	if in.PeriodStart.IsZero() {
		in.PeriodStart = time.Now().UTC()
	}
	if in.PeriodEnd.IsZero() {
		in.PeriodEnd = in.PeriodStart
	}
	in.PeriodStart, in.PeriodEnd = dateOnly(in.PeriodStart), dateOnly(in.PeriodEnd)
	if in.TenantID == "" || in.IdempotencyKey == "" || in.Dimension == "" || in.Unit == "" || in.SourceType == "" || in.Amount <= 0 || in.PeriodEnd.Before(in.PeriodStart) {
		return in, ErrUsageValidation
	}
	if !validUsageDimensionUnit(in.Dimension, in.Unit) {
		return in, ErrUsageValidation
	}
	switch in.State {
	case UsageStateApplied, UsageStateCorrection, UsageStateReversed, UsageStateVoided:
	default:
		return in, ErrUsageValidation
	}
	if in.State == UsageStateCorrection && in.CorrectionOf == "" {
		return in, ErrUsageValidation
	}
	return in, nil
}

func validUsageDimensionUnit(dimension, unit string) bool {
	switch dimension {
	case "ai_employees":
		return unit == "assignments"
	case "monthly_call_minutes", "mobile_call_minutes":
		return unit == "minutes"
	case "km_documents":
		return unit == "documents"
	case "storage_bytes":
		return unit == "bytes"
	case "concurrent_calls":
		return unit == "calls"
	default:
		return false
	}
}

func dateOnly(value time.Time) time.Time {
	utc := value.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func (s *Store) RecordUsageEvent(ctx context.Context, input UsageEventInput) (UsageEvent, bool, error) {
	input, err := normalizeUsageEvent(input)
	if err != nil {
		return UsageEvent{}, false, err
	}
	if s == nil || s.pg == nil {
		return UsageEvent{}, false, fmt.Errorf("usage ledger unavailable")
	}
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return UsageEvent{}, false, err
	}
	defer tx.Rollback(ctx)
	schema, id := quoteIdent(s.cfg.PostgresSchema), "ue_"+newStoreID()
	tag, err := tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.usage_events (id,tenant_id,idempotency_key,dimension,unit,amount,period_start,period_end,source_type,source_id,entitlement_snapshot_id,state,correction_of,created_by,updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$14) ON CONFLICT (tenant_id,idempotency_key) DO NOTHING`, schema), id, input.TenantID, input.IdempotencyKey, input.Dimension, input.Unit, input.Amount, input.PeriodStart, input.PeriodEnd, input.SourceType, input.SourceID, input.EntitlementSnapshotID, input.State, input.CorrectionOf, auditctx.ActorID(ctx))
	if err != nil {
		return UsageEvent{}, false, err
	}
	event, err := scanUsageEvent(tx.QueryRow(ctx, fmt.Sprintf(`SELECT id,tenant_id,idempotency_key,dimension,unit,amount,period_start,period_end,source_type,source_id,entitlement_snapshot_id,state,correction_of,created_at,updated_at FROM %s.usage_events WHERE tenant_id=$1 AND idempotency_key=$2`, schema), input.TenantID, input.IdempotencyKey))
	if err != nil {
		return UsageEvent{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return UsageEvent{}, false, err
	}
	return event, tag.RowsAffected() == 0, nil
}

func scanUsageEvent(row pgx.Row) (UsageEvent, error) {
	var event UsageEvent
	err := row.Scan(&event.ID, &event.TenantID, &event.IdempotencyKey, &event.Dimension, &event.Unit, &event.Amount, &event.PeriodStart, &event.PeriodEnd, &event.SourceType, &event.SourceID, &event.EntitlementSnapshotID, &event.State, &event.CorrectionOf, &event.CreatedAt, &event.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return UsageEvent{}, ErrUsageEventNotFound
	}
	return event, err
}

func (s *Store) CreateUsageReconciliationRun(ctx context.Context, input UsageReconciliationInput) (UsageReconciliationRun, bool, error) {
	input.TenantID, input.IdempotencyKey = strings.TrimSpace(input.TenantID), strings.TrimSpace(input.IdempotencyKey)
	if input.IdempotencyKey == "" || input.StartDate.IsZero() || input.EndDate.IsZero() || dateOnly(input.EndDate).Before(dateOnly(input.StartDate)) {
		return UsageReconciliationRun{}, false, ErrUsageValidation
	}
	input.StartDate, input.EndDate = dateOnly(input.StartDate), dateOnly(input.EndDate)
	if s == nil || s.pg == nil {
		return UsageReconciliationRun{}, false, fmt.Errorf("usage ledger unavailable")
	}
	tx, err := s.pg.Begin(ctx)
	if err != nil {
		return UsageReconciliationRun{}, false, err
	}
	defer tx.Rollback(ctx)
	schema, id := quoteIdent(s.cfg.PostgresSchema), "urec_"+newStoreID()
	tag, err := tx.Exec(ctx, fmt.Sprintf(`INSERT INTO %s.usage_reconciliation_runs (id,tenant_id,idempotency_key,start_date,end_date,dry_run,status,created_by,updated_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$8) ON CONFLICT (tenant_id,idempotency_key) DO NOTHING`, schema), id, input.TenantID, input.IdempotencyKey, input.StartDate, input.EndDate, input.DryRun, UsageReconciliationQueued, auditctx.ActorID(ctx))
	if err != nil {
		return UsageReconciliationRun{}, false, err
	}
	run, err := scanUsageReconciliationRun(tx.QueryRow(ctx, fmt.Sprintf(`SELECT id,tenant_id,idempotency_key,start_date,end_date,dry_run,status,source_watermarks,mismatch_count,correction_count,error_code,created_at,updated_at FROM %s.usage_reconciliation_runs WHERE tenant_id=$1 AND idempotency_key=$2`, schema), input.TenantID, input.IdempotencyKey))
	if err != nil {
		return UsageReconciliationRun{}, false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return UsageReconciliationRun{}, false, err
	}
	return run, tag.RowsAffected() == 0, nil
}

func scanUsageReconciliationRun(row pgx.Row) (UsageReconciliationRun, error) {
	var run UsageReconciliationRun
	var raw []byte
	err := row.Scan(&run.ID, &run.TenantID, &run.IdempotencyKey, &run.StartDate, &run.EndDate, &run.DryRun, &run.Status, &raw, &run.MismatchCount, &run.CorrectionCount, &run.ErrorCode, &run.CreatedAt, &run.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return UsageReconciliationRun{}, ErrUsageReconciliationNotFound
	}
	if err != nil {
		return UsageReconciliationRun{}, err
	}
	run.SourceWatermarks = map[string]any{}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &run.SourceWatermarks); err != nil {
			return UsageReconciliationRun{}, err
		}
	}
	return run, nil
}

func (s *Store) GetUsageReconciliationRun(ctx context.Context, id string) (UsageReconciliationRun, error) {
	if s == nil || s.pg == nil {
		return UsageReconciliationRun{}, fmt.Errorf("usage ledger unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	return scanUsageReconciliationRun(s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT id,tenant_id,idempotency_key,start_date,end_date,dry_run,status,source_watermarks,mismatch_count,correction_count,error_code,created_at,updated_at FROM %s.usage_reconciliation_runs WHERE id=$1`, schema), strings.TrimSpace(id)))
}

// ListUsageProjections reads historical, tenant-scoped ledger activity. It is
// deliberately separate from Redis-backed enforcement counters.
func (s *Store) ListUsageProjections(ctx context.Context, tenantID string, startDate, endDate time.Time) ([]UsageProjection, error) {
	if s == nil || s.pg == nil {
		return nil, fmt.Errorf("usage ledger unavailable")
	}
	if strings.TrimSpace(tenantID) == "" || startDate.IsZero() || endDate.IsZero() {
		return nil, ErrUsageValidation
	}
	startDate, endDate = dateOnly(startDate), dateOnly(endDate)
	if endDate.Before(startDate) {
		return nil, ErrUsageValidation
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	rows, err := s.pg.Query(ctx, fmt.Sprintf(`
SELECT dimension, unit, period_start,
       COALESCE(SUM(CASE WHEN state IN ('reversed','voided') THEN -amount ELSE amount END), 0)::bigint,
       COALESCE(MAX(source_type), ''), COALESCE(MAX(entitlement_snapshot_id), ''), MAX(updated_at)
FROM %s.usage_events
WHERE tenant_id=$1 AND period_start <= $3 AND period_end >= $2
GROUP BY dimension, unit, period_start
ORDER BY period_start, dimension`, schema), strings.TrimSpace(tenantID), startDate, endDate)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	projections := make([]UsageProjection, 0)
	for rows.Next() {
		var dimension, unit, source, snapshot string
		var period time.Time
		var consumed int64
		var freshness time.Time
		if err := rows.Scan(&dimension, &unit, &period, &consumed, &source, &snapshot, &freshness); err != nil {
			return nil, err
		}
		projections = append(projections, UsageProjection{
			Dimension: dimension, Unit: unit, Period: period.UTC().Format("2006-01-02"),
			Consumed: consumed, Source: source, Freshness: freshness.UTC().Format(time.RFC3339Nano),
			EntitlementSnapshotID: snapshot,
		})
	}
	return projections, rows.Err()
}

func (s *Store) RunUsageReconciliation(ctx context.Context, id string) error {
	if s == nil || s.pg == nil {
		return fmt.Errorf("usage ledger unavailable")
	}
	schema := quoteIdent(s.cfg.PostgresSchema)
	var tenantID string
	var startDate, endDate time.Time
	err := s.pg.QueryRow(ctx, fmt.Sprintf(`UPDATE %s.usage_reconciliation_runs SET status=$2,updated_at=now() WHERE id=$1 AND status=$3 RETURNING tenant_id,start_date,end_date`, schema), id, UsageReconciliationRunning, UsageReconciliationQueued).Scan(&tenantID, &startDate, &endDate)
	if errors.Is(err, pgx.ErrNoRows) {
		run, getErr := s.GetUsageReconciliationRun(ctx, id)
		if getErr != nil {
			return getErr
		}
		if run.Status == UsageReconciliationRunning {
			return ErrUsageReconciliationInProgress
		}
		return nil
	}
	if err != nil {
		return err
	}
	var count, amount int64
	var latest *time.Time
	err = s.pg.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*)::bigint,COALESCE(SUM(amount),0)::bigint,MAX(created_at) FROM %s.usage_events WHERE period_start <= $2 AND period_end >= $1 AND ($3='' OR tenant_id=$3)`, schema), startDate, endDate, tenantID).Scan(&count, &amount, &latest)
	if err != nil {
		_, _ = s.pg.Exec(ctx, fmt.Sprintf(`UPDATE %s.usage_reconciliation_runs SET status=$2,error_code=$3,updated_at=now() WHERE id=$1`, schema), id, UsageReconciliationFailed, "reconciliation_source_unavailable")
		return err
	}
	watermarks := map[string]any{"usage_events_count": count, "usage_events_amount": amount}
	if latest != nil {
		watermarks["usage_events_latest_created_at"] = latest.UTC().Format(time.RFC3339Nano)
	}
	payload, err := json.Marshal(watermarks)
	if err != nil {
		return err
	}
	_, err = s.pg.Exec(ctx, fmt.Sprintf(`UPDATE %s.usage_reconciliation_runs SET status=$2,source_watermarks=$3::jsonb,mismatch_count=0,correction_count=0,error_code='',updated_at=now() WHERE id=$1`, schema), id, UsageReconciliationCompleted, string(payload))
	return err
}
