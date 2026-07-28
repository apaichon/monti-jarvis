-- Sprint 52: tenant-owned avatar library (active cap separate from create).
\set ON_ERROR_STOP on

ALTER TABLE :POSTGRES_SCHEMA.ai_avatars
  ADD COLUMN IF NOT EXISTS owner_tenant_id text REFERENCES :POSTGRES_SCHEMA.tenants(id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS ai_avatars_owner_tenant_idx
  ON :POSTGRES_SCHEMA.ai_avatars (owner_tenant_id)
  WHERE owner_tenant_id IS NOT NULL;
