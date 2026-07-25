-- Sprint 45: idempotent usage ledger and bounded reconciliation runs.
\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.usage_events (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES :POSTGRES_SCHEMA.tenants(id) ON DELETE CASCADE,
  idempotency_key text NOT NULL,
  dimension text NOT NULL CHECK (dimension IN ('ai_employees','monthly_call_minutes','mobile_call_minutes','km_documents','storage_bytes','concurrent_calls')),
  unit text NOT NULL CHECK (unit IN ('assignments','minutes','documents','bytes','calls')),
  amount bigint NOT NULL CHECK (amount > 0),
  period_start date NOT NULL,
  period_end date NOT NULL,
  source_type text NOT NULL,
  source_id text NOT NULL DEFAULT '',
  entitlement_snapshot_id text NOT NULL DEFAULT '',
  state text NOT NULL DEFAULT 'applied' CHECK (state IN ('applied','correction','reversed','voided')),
  correction_of text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system',
  UNIQUE (tenant_id, idempotency_key),
  CHECK (period_end >= period_start)
);
CREATE INDEX IF NOT EXISTS usage_events_period_idx ON :POSTGRES_SCHEMA.usage_events (tenant_id, period_start, period_end);
CREATE INDEX IF NOT EXISTS usage_events_source_idx ON :POSTGRES_SCHEMA.usage_events (source_type, source_id);

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.usage_reconciliation_runs (
  id text PRIMARY KEY,
  tenant_id text NOT NULL DEFAULT '',
  idempotency_key text NOT NULL,
  start_date date NOT NULL,
  end_date date NOT NULL,
  dry_run boolean NOT NULL DEFAULT true,
  status text NOT NULL DEFAULT 'queued' CHECK (status IN ('queued','running','completed','failed')),
  source_watermarks jsonb NOT NULL DEFAULT '{}'::jsonb,
  mismatch_count integer NOT NULL DEFAULT 0,
  correction_count integer NOT NULL DEFAULT 0,
  error_code text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system',
  CHECK (end_date >= start_date)
);

ALTER TABLE :POSTGRES_SCHEMA.usage_reconciliation_runs
  DROP CONSTRAINT IF EXISTS usage_reconciliation_runs_idempotency_key_key;
CREATE UNIQUE INDEX IF NOT EXISTS usage_reconciliation_runs_tenant_key_idx
  ON :POSTGRES_SCHEMA.usage_reconciliation_runs (tenant_id, idempotency_key);
