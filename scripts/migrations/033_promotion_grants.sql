-- Sprint 50: admin promotional package grants (active plan + tax invoice audit).
\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.promotion_grants (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES :POSTGRES_SCHEMA.tenants(id) ON DELETE CASCADE,
  package_id text NOT NULL REFERENCES :POSTGRES_SCHEMA.packages(id),
  order_id text NOT NULL UNIQUE REFERENCES :POSTGRES_SCHEMA.payment_orders(id) ON DELETE CASCADE,
  entitlement_id text NOT NULL DEFAULT '',
  tax_invoice_id text NOT NULL DEFAULT '',
  reason text NOT NULL,
  idempotency_key text NOT NULL DEFAULT '',
  valid_until timestamptz,
  amount_cents int NOT NULL DEFAULT 0,
  status text NOT NULL DEFAULT 'issued' CHECK (status IN ('issued', 'failed')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system'
);

CREATE INDEX IF NOT EXISTS promotion_grants_tenant_created_idx
  ON :POSTGRES_SCHEMA.promotion_grants (tenant_id, created_at DESC);

CREATE UNIQUE INDEX IF NOT EXISTS promotion_grants_tenant_idempotency_uidx
  ON :POSTGRES_SCHEMA.promotion_grants (tenant_id, idempotency_key)
  WHERE idempotency_key <> '';
