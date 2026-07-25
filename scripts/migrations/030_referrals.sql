-- Sprint 46: tenant referral attribution and qualification foundation.
-- Bonus quota grants are intentionally not represented by this migration.

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.tenant_referral_codes (
  id text PRIMARY KEY,
  tenant_id text NOT NULL UNIQUE REFERENCES :POSTGRES_SCHEMA.tenants(id) ON DELETE CASCADE,
  code text NOT NULL UNIQUE,
  status text NOT NULL DEFAULT 'active'
    CHECK (status IN ('active', 'disabled')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system'
);

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.tenant_referrals (
  id text PRIMARY KEY,
  referrer_tenant_id text NOT NULL REFERENCES :POSTGRES_SCHEMA.tenants(id) ON DELETE CASCADE,
  referred_tenant_id text NOT NULL UNIQUE REFERENCES :POSTGRES_SCHEMA.tenants(id) ON DELETE CASCADE,
  referral_code_id text NOT NULL REFERENCES :POSTGRES_SCHEMA.tenant_referral_codes(id),
  code text NOT NULL,
  status text NOT NULL DEFAULT 'attributed'
    CHECK (status IN ('clicked', 'attributed', 'pending', 'qualified', 'rejected', 'reversed')),
  source text NOT NULL DEFAULT '',
  qualification_reason text NOT NULL DEFAULT '',
  attributed_at timestamptz,
  qualified_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system'
);

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.tenant_referral_events (
  id text PRIMARY KEY,
  referral_id text NOT NULL REFERENCES :POSTGRES_SCHEMA.tenant_referrals(id) ON DELETE CASCADE,
  from_status text NOT NULL DEFAULT '',
  to_status text NOT NULL
    CHECK (to_status IN ('clicked', 'attributed', 'pending', 'qualified', 'rejected', 'reversed')),
  reason text NOT NULL DEFAULT '',
  source text NOT NULL DEFAULT '',
  event_at timestamptz NOT NULL DEFAULT now(),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS tenant_referrals_referrer_idx
  ON :POSTGRES_SCHEMA.tenant_referrals (referrer_tenant_id, created_at DESC);

CREATE INDEX IF NOT EXISTS tenant_referrals_status_idx
  ON :POSTGRES_SCHEMA.tenant_referrals (status, created_at DESC);
