BEGIN;

CREATE TABLE IF NOT EXISTS callcenter.schema_migrations (
  version text PRIMARY KEY,
  applied_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE callcenter.packages
  ADD COLUMN IF NOT EXISTS deployment_mode text NOT NULL DEFAULT 'shared_cloud';
ALTER TABLE callcenter.packages
  ADD COLUMN IF NOT EXISTS purchase_mode text NOT NULL DEFAULT 'self_serve';

UPDATE callcenter.packages
SET deployment_mode = 'dedicated_vm', purchase_mode = 'quote'
WHERE id LIKE 'pkg-dedicated-%' OR slug LIKE 'dedicated-%';

UPDATE callcenter.packages
SET deployment_mode = 'shared_cloud', purchase_mode = 'self_serve'
WHERE id LIKE 'pkg-shared-%' OR slug LIKE 'shared-%';

CREATE TABLE IF NOT EXISTS callcenter.package_versions (
  id text PRIMARY KEY,
  package_id text NOT NULL REFERENCES callcenter.packages(id),
  version integer NOT NULL CHECK (version > 0),
  monthly_price_cents integer NOT NULL CHECK (monthly_price_cents >= 0),
  annual_price_cents integer NOT NULL CHECK (annual_price_cents >= 0),
  annual_discount_bps integer NOT NULL DEFAULT 2000 CHECK (annual_discount_bps BETWEEN 0 AND 10000),
  currency text NOT NULL,
  tax_rate_bps integer NOT NULL DEFAULT 700 CHECK (tax_rate_bps BETWEEN 0 AND 10000),
  rules_snapshot jsonb NOT NULL DEFAULT '{}'::jsonb,
  effective_from timestamptz NOT NULL DEFAULT now(),
  effective_until timestamptz,
  status text NOT NULL DEFAULT 'active' CHECK (status IN ('draft','active','retired')),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system',
  UNIQUE (package_id, version),
  CHECK (effective_until IS NULL OR effective_until > effective_from)
);
CREATE UNIQUE INDEX IF NOT EXISTS package_versions_one_active_idx
  ON callcenter.package_versions (package_id) WHERE status = 'active';

CREATE TABLE IF NOT EXISTS callcenter.dedicated_quote_requests (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES callcenter.tenants(id) ON DELETE CASCADE,
  package_version_id text NOT NULL REFERENCES callcenter.package_versions(id),
  company_legal_name text NOT NULL,
  contact_name text NOT NULL,
  contact_email text NOT NULL,
  contact_phone text NOT NULL,
  tax_registration_id text NOT NULL DEFAULT '',
  company_size text NOT NULL DEFAULT '',
  expected_concurrency integer NOT NULL CHECK (expected_concurrency > 0),
  preferred_region text NOT NULL DEFAULT '',
  notes text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'submitted'
    CHECK (status IN ('submitted','under_review','capacity_confirmed','quoted','accepted','provisioning','active','rejected','expired','withdrawn')),
  quoted_amount_cents integer CHECK (quoted_amount_cents >= 0),
  currency text NOT NULL DEFAULT 'THB',
  quote_expires_at timestamptz,
  capacity_snapshot jsonb NOT NULL DEFAULT '{}'::jsonb,
  idempotency_key text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system'
);
CREATE INDEX IF NOT EXISTS dedicated_quotes_tenant_created_idx
  ON callcenter.dedicated_quote_requests (tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS dedicated_quotes_status_created_idx
  ON callcenter.dedicated_quote_requests (status, created_at);
CREATE UNIQUE INDEX IF NOT EXISTS dedicated_quotes_tenant_idem_idx
  ON callcenter.dedicated_quote_requests (tenant_id, idempotency_key)
  WHERE idempotency_key <> '';

CREATE TABLE IF NOT EXISTS callcenter.tenant_subscriptions (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES callcenter.tenants(id) ON DELETE CASCADE,
  package_version_id text NOT NULL REFERENCES callcenter.package_versions(id),
  entitlement_id text NOT NULL DEFAULT '',
  quote_request_id text REFERENCES callcenter.dedicated_quote_requests(id),
  initial_order_id text UNIQUE REFERENCES callcenter.payment_orders(id),
  billing_interval text NOT NULL CHECK (billing_interval IN ('monthly','annual')),
  status text NOT NULL
    CHECK (status IN ('pending_payment','active','past_due','grace','suspended','cancelled','ended')),
  billing_anchor timestamptz NOT NULL,
  current_period_start timestamptz NOT NULL,
  current_period_end timestamptz NOT NULL,
  next_bill_at timestamptz,
  grace_until timestamptz,
  calculation_snapshot jsonb NOT NULL,
  provider text NOT NULL DEFAULT '',
  payment_method text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system',
  CHECK (current_period_end > current_period_start)
);
CREATE INDEX IF NOT EXISTS tenant_subscriptions_tenant_created_idx
  ON callcenter.tenant_subscriptions (tenant_id, created_at DESC);
CREATE UNIQUE INDEX IF NOT EXISTS tenant_subscriptions_one_live_idx
  ON callcenter.tenant_subscriptions (tenant_id)
  WHERE status IN ('active','past_due','grace','suspended');

CREATE TABLE IF NOT EXISTS callcenter.billing_cycles (
  id text PRIMARY KEY,
  subscription_id text NOT NULL REFERENCES callcenter.tenant_subscriptions(id) ON DELETE CASCADE,
  period_key text NOT NULL,
  period_start timestamptz NOT NULL,
  period_end timestamptz NOT NULL,
  status text NOT NULL
    CHECK (status IN ('scheduled','previewed','payment_pending','paid','documents_issued','settled','retry_wait','failed')),
  calculation_snapshot jsonb NOT NULL,
  order_id text REFERENCES callcenter.payment_orders(id),
  receipt_id text,
  tax_invoice_id text,
  attempt_count integer NOT NULL DEFAULT 0 CHECK (attempt_count >= 0),
  next_attempt_at timestamptz,
  last_error_code text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system',
  UNIQUE (subscription_id, period_key),
  CHECK (period_end > period_start)
);
CREATE INDEX IF NOT EXISTS billing_cycles_retry_idx
  ON callcenter.billing_cycles (status, next_attempt_at);

INSERT INTO callcenter.package_versions (
  id, package_id, version, monthly_price_cents, annual_price_cents,
  annual_discount_bps, currency, tax_rate_bps, rules_snapshot,
  effective_from, status, created_by, updated_by
)
SELECT
  'pkgv_' || replace(p.id, '-', '_') || '_v1',
  p.id,
  1,
  p.price_cents,
  ((p.price_cents::bigint * 12 * 8000 + 5000) / 10000)::integer,
  2000,
  p.currency,
  700,
  pl.rules,
  now(),
  'active',
  'system',
  'system'
FROM callcenter.packages p
JOIN callcenter.package_limits pl ON pl.package_id = p.id
WHERE p.status = 'active'
ON CONFLICT (package_id, version) DO NOTHING;

INSERT INTO callcenter.schema_migrations (version)
VALUES ('035_commercial_operations')
ON CONFLICT (version) DO NOTHING;

COMMIT;
