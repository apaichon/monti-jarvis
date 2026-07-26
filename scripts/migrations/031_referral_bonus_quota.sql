-- Sprint 46: configurable referral rewards and append-only bonus entitlement ledger.
\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.referral_reward_rules (
  dimension text PRIMARY KEY CHECK (dimension IN ('ai_employees','monthly_call_minutes','mobile_call_minutes','km_documents','storage_bytes','concurrent_calls')),
  grant_amount bigint NOT NULL CHECK (grant_amount >= 0),
  expiry_days integer NOT NULL DEFAULT 90 CHECK (expiry_days >= 0),
  active boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system'
);

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.tenant_bonus_ledger (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES :POSTGRES_SCHEMA.tenants(id) ON DELETE CASCADE,
  referral_id text REFERENCES :POSTGRES_SCHEMA.tenant_referrals(id) ON DELETE CASCADE,
  dimension text NOT NULL CHECK (dimension IN ('ai_employees','monthly_call_minutes','mobile_call_minutes','km_documents','storage_bytes','concurrent_calls')),
  operation text NOT NULL CHECK (operation IN ('grant','consume','expire','reverse','reconcile')),
  amount bigint NOT NULL CHECK (amount > 0),
  idempotency_key text NOT NULL,
  expires_at timestamptz,
  source_type text NOT NULL DEFAULT '',
  source_id text NOT NULL DEFAULT '',
  reason text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system',
  UNIQUE (tenant_id, idempotency_key)
);

CREATE INDEX IF NOT EXISTS tenant_bonus_ledger_balance_idx
  ON :POSTGRES_SCHEMA.tenant_bonus_ledger (tenant_id, dimension, created_at);
CREATE INDEX IF NOT EXISTS tenant_bonus_ledger_referral_idx
  ON :POSTGRES_SCHEMA.tenant_bonus_ledger (referral_id, dimension);

INSERT INTO :POSTGRES_SCHEMA.referral_reward_rules (dimension, grant_amount, expiry_days, active)
VALUES
  ('ai_employees', 1, 90, true),
  ('monthly_call_minutes', 60, 90, true),
  ('mobile_call_minutes', 60, 90, true),
  ('km_documents', 10, 90, true),
  ('storage_bytes', 1073741824, 90, true),
  ('concurrent_calls', 1, 90, true)
ON CONFLICT (dimension) DO NOTHING;

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.tenant_referral_clicks (
  id text PRIMARY KEY,
  referral_code_id text NOT NULL REFERENCES :POSTGRES_SCHEMA.tenant_referral_codes(id) ON DELETE CASCADE,
  code text NOT NULL,
  source text NOT NULL DEFAULT '',
  landing_path text NOT NULL DEFAULT '',
  clicked_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system'
);
CREATE INDEX IF NOT EXISTS tenant_referral_clicks_code_idx
  ON :POSTGRES_SCHEMA.tenant_referral_clicks (referral_code_id, clicked_at DESC);
