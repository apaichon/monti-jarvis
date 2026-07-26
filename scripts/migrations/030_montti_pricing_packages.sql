-- Montti AI complete pricing realignment (docs/sales/Montti_AI_Complete_Pricing.md)
-- Shared Cloud: startups/SME self-serve (payment gateway)
-- Dedicated VM: quote-only after capacity check
\set ON_ERROR_STOP on

-- Archive legacy rows
UPDATE :POSTGRES_SCHEMA.packages
SET status = 'archived', updated_by = 'system', updated_at = now()
WHERE id IN (
  'pkg-starter','pkg-pro','pkg-enterprise',
  'pkg-aiaas-500','pkg-aiaas-1000','pkg-aiaas-1500','pkg-aiaas-2000'
) AND status <> 'archived';

-- rules-v2 schema (ensure present)
INSERT INTO :POSTGRES_SCHEMA.package_rule_schemas (id, version, name, fields, status)
VALUES (
  'rules-v2', 2, 'Montti shared/dedicated dimensions',
  '{
    "max_ai_employees": {"type":"int","min":0,"required":true},
    "max_monthly_call_minutes": {"type":"int","min":0,"required":true},
    "max_mobile_call_minutes": {"type":"int","min":0,"required":true},
    "max_km_documents": {"type":"int","min":0,"required":true},
    "max_storage_bytes": {"type":"int","min":0,"required":true},
    "max_concurrent_calls": {"type":"int","min":0,"required":true},
    "voice_enabled": {"type":"bool","required":true,"default":true},
    "rag_enabled": {"type":"bool","required":true,"default":true}
  }'::jsonb,
  'active'
)
ON CONFLICT (id) DO UPDATE SET fields = EXCLUDED.fields, name = EXCLUDED.name, status = 'active';

-- Shared Cloud (THB) — self-serve
INSERT INTO :POSTGRES_SCHEMA.packages (id, slug, name, description, status, price_cents, currency, billing_period, created_by, updated_by)
VALUES
  ('pkg-shared-launch', 'shared-launch', 'Launch',
   'Shared Cloud · self-serve · startups/SME · BYOK AI · concurrent voice 1',
   'active', 50000, 'THB', 'monthly', 'system', 'system'),
  ('pkg-shared-starter', 'shared-starter', 'Starter',
   'Shared Cloud · self-serve · startups/SME · BYOK AI · concurrent voice 2',
   'active', 90000, 'THB', 'monthly', 'system', 'system'),
  ('pkg-shared-growth', 'shared-growth', 'Growth',
   'Shared Cloud · self-serve · startups/SME · BYOK AI · concurrent voice 4',
   'active', 150000, 'THB', 'monthly', 'system', 'system'),
  ('pkg-shared-business', 'shared-business', 'Business',
   'Shared Cloud · self-serve · startups/SME · BYOK AI · concurrent voice 6',
   'active', 200000, 'THB', 'monthly', 'system', 'system')
ON CONFLICT (id) DO UPDATE SET
  slug = EXCLUDED.slug, name = EXCLUDED.name, description = EXCLUDED.description,
  status = 'active', price_cents = EXCLUDED.price_cents, currency = EXCLUDED.currency,
  billing_period = 'monthly', updated_by = 'system', updated_at = now();

-- KM document count is commercially unlimited; real KM cap is max_storage_bytes (1e6 sentinel).
INSERT INTO :POSTGRES_SCHEMA.package_limits (package_id, rules_schema_id, rules, created_by, updated_by) VALUES
  ('pkg-shared-launch', 'rules-v2',
   '{"max_ai_employees":2,"max_km_documents":1000000,"max_storage_bytes":1073741824,"max_concurrent_calls":1,"max_monthly_call_minutes":99999999,"max_mobile_call_minutes":99999999,"voice_enabled":true,"rag_enabled":true}'::jsonb,
   'system', 'system'),
  ('pkg-shared-starter', 'rules-v2',
   '{"max_ai_employees":5,"max_km_documents":1000000,"max_storage_bytes":5368709120,"max_concurrent_calls":2,"max_monthly_call_minutes":99999999,"max_mobile_call_minutes":99999999,"voice_enabled":true,"rag_enabled":true}'::jsonb,
   'system', 'system'),
  ('pkg-shared-growth', 'rules-v2',
   '{"max_ai_employees":10,"max_km_documents":1000000,"max_storage_bytes":10737418240,"max_concurrent_calls":4,"max_monthly_call_minutes":99999999,"max_mobile_call_minutes":99999999,"voice_enabled":true,"rag_enabled":true}'::jsonb,
   'system', 'system'),
  ('pkg-shared-business', 'rules-v2',
   '{"max_ai_employees":20,"max_km_documents":1000000,"max_storage_bytes":21474836480,"max_concurrent_calls":6,"max_monthly_call_minutes":99999999,"max_mobile_call_minutes":99999999,"voice_enabled":true,"rag_enabled":true}'::jsonb,
   'system', 'system')
ON CONFLICT (package_id) DO UPDATE SET
  rules_schema_id = EXCLUDED.rules_schema_id, rules = EXCLUDED.rules,
  updated_by = 'system', updated_at = now();

-- Dedicated VM (THB, same currency as Shared Cloud) — quote only
-- List prices converted from sheet EUR (€99/€299/€499/€849) at ≈฿38.5/EUR, rounded to ฿100.
INSERT INTO :POSTGRES_SCHEMA.packages (id, slug, name, description, status, price_cents, currency, billing_period, created_by, updated_by)
VALUES
  ('pkg-dedicated-launch', 'dedicated-launch', 'Dedicated Launch',
   'Dedicated VM · quote required · capacity check · 8 vCPU · 24 GB RAM · 300 GB SSD · up to 100 concurrent',
   'active', 380000, 'THB', 'monthly', 'system', 'system'),
  ('pkg-dedicated-growth', 'dedicated-growth', 'Dedicated Growth',
   'Dedicated VM · quote required · capacity check · 12 vCPU · 48 GB RAM · 400 GB SSD · up to 250 concurrent · white-label included',
   'active', 1150000, 'THB', 'monthly', 'system', 'system'),
  ('pkg-dedicated-business', 'dedicated-business', 'Dedicated Business',
   'Dedicated VM · quote required · capacity check · 16 vCPU · 64 GB RAM · 500 GB SSD · up to 500 concurrent',
   'active', 1920000, 'THB', 'monthly', 'system', 'system'),
  ('pkg-dedicated-enterprise', 'dedicated-enterprise', 'Dedicated Enterprise',
   'Dedicated VM · quote required · capacity check · 18 vCPU · 96 GB RAM · 600 GB SSD · up to 1000 concurrent',
   'active', 3270000, 'THB', 'monthly', 'system', 'system')
ON CONFLICT (id) DO UPDATE SET
  slug = EXCLUDED.slug, name = EXCLUDED.name, description = EXCLUDED.description,
  status = 'active', price_cents = EXCLUDED.price_cents, currency = EXCLUDED.currency,
  billing_period = 'monthly', updated_by = 'system', updated_at = now();

INSERT INTO :POSTGRES_SCHEMA.package_limits (package_id, rules_schema_id, rules, created_by, updated_by) VALUES
  ('pkg-dedicated-launch', 'rules-v2',
   '{"max_ai_employees":1000000,"max_km_documents":1000000,"max_storage_bytes":322122547200,"max_concurrent_calls":100,"max_monthly_call_minutes":99999999,"max_mobile_call_minutes":99999999,"voice_enabled":true,"rag_enabled":true}'::jsonb,
   'system', 'system'),
  ('pkg-dedicated-growth', 'rules-v2',
   '{"max_ai_employees":1000000,"max_km_documents":1000000,"max_storage_bytes":429496729600,"max_concurrent_calls":250,"max_monthly_call_minutes":99999999,"max_mobile_call_minutes":99999999,"voice_enabled":true,"rag_enabled":true}'::jsonb,
   'system', 'system'),
  ('pkg-dedicated-business', 'rules-v2',
   '{"max_ai_employees":1000000,"max_km_documents":1000000,"max_storage_bytes":536870912000,"max_concurrent_calls":500,"max_monthly_call_minutes":99999999,"max_mobile_call_minutes":99999999,"voice_enabled":true,"rag_enabled":true}'::jsonb,
   'system', 'system'),
  ('pkg-dedicated-enterprise', 'rules-v2',
   '{"max_ai_employees":1000000,"max_km_documents":1000000,"max_storage_bytes":644245094400,"max_concurrent_calls":1000,"max_monthly_call_minutes":99999999,"max_mobile_call_minutes":99999999,"voice_enabled":true,"rag_enabled":true}'::jsonb,
   'system', 'system')
ON CONFLICT (package_id) DO UPDATE SET
  rules_schema_id = EXCLUDED.rules_schema_id, rules = EXCLUDED.rules,
  updated_by = 'system', updated_at = now();
