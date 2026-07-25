-- Sprint 43: embed auth and tenant AI configuration.
\set ON_ERROR_STOP on

ALTER TABLE :POSTGRES_SCHEMA.tenant_embed_configs
  ADD COLUMN IF NOT EXISTS auth_required boolean NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.tenant_ai_configs (
  tenant_id text PRIMARY KEY REFERENCES :POSTGRES_SCHEMA.tenants(id) ON DELETE CASCADE,
  gemini_key_ciphertext bytea,
  gemini_key_nonce bytea,
  gemini_key_version text,
  gemini_key_last4 text,
  gemini_key_updated_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system'
);

-- CREATE TABLE IF NOT EXISTS does not upgrade a table created by an older
-- server. Keep existing installations compatible with encrypted key reads.
ALTER TABLE :POSTGRES_SCHEMA.tenant_ai_configs
  ADD COLUMN IF NOT EXISTS gemini_key_ciphertext bytea,
  ADD COLUMN IF NOT EXISTS gemini_key_nonce bytea,
  ADD COLUMN IF NOT EXISTS gemini_key_version text,
  ADD COLUMN IF NOT EXISTS gemini_key_last4 text,
  ADD COLUMN IF NOT EXISTS gemini_key_updated_at timestamptz;

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.tenant_agent_configs (
  tenant_id text NOT NULL REFERENCES :POSTGRES_SCHEMA.tenants(id) ON DELETE CASCADE,
  agent_id text NOT NULL,
  system_prompt text NOT NULL DEFAULT '',
  enabled boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system',
  PRIMARY KEY (tenant_id, agent_id)
);
ALTER TABLE :POSTGRES_SCHEMA.tenant_agent_configs
  ADD COLUMN IF NOT EXISTS system_prompt text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS enabled boolean NOT NULL DEFAULT true;

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.tenant_call_tools (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES :POSTGRES_SCHEMA.tenants(id) ON DELETE CASCADE,
  tool_key text NOT NULL,
  display_name text NOT NULL,
  description text NOT NULL,
  handler_key text NOT NULL,
  input_schema jsonb NOT NULL,
  enabled boolean NOT NULL DEFAULT false,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system',
  UNIQUE (tenant_id, tool_key)
);
ALTER TABLE :POSTGRES_SCHEMA.tenant_call_tools
  ADD COLUMN IF NOT EXISTS tool_key text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS display_name text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS description text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS handler_key text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS input_schema jsonb NOT NULL DEFAULT '{}'::jsonb,
  ADD COLUMN IF NOT EXISTS enabled boolean NOT NULL DEFAULT false;

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.tenant_skills (
  id text PRIMARY KEY,
  tenant_id text NOT NULL REFERENCES :POSTGRES_SCHEMA.tenants(id) ON DELETE CASCADE,
  slug text NOT NULL,
  name text NOT NULL,
  prompt text NOT NULL DEFAULT '',
  enabled boolean NOT NULL DEFAULT true,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system',
  UNIQUE (tenant_id, slug)
);
ALTER TABLE :POSTGRES_SCHEMA.tenant_skills
  ADD COLUMN IF NOT EXISTS slug text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS name text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS prompt text NOT NULL DEFAULT '',
  ADD COLUMN IF NOT EXISTS enabled boolean NOT NULL DEFAULT true;

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.tenant_skill_tools (
  tenant_id text NOT NULL REFERENCES :POSTGRES_SCHEMA.tenants(id) ON DELETE CASCADE,
  skill_id text NOT NULL,
  tool_id text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system',
  PRIMARY KEY (tenant_id, skill_id, tool_id)
);

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.tenant_agent_skills (
  tenant_id text NOT NULL REFERENCES :POSTGRES_SCHEMA.tenants(id) ON DELETE CASCADE,
  agent_id text NOT NULL,
  skill_id text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system',
  PRIMARY KEY (tenant_id, agent_id, skill_id)
);

CREATE INDEX IF NOT EXISTS tenant_agent_configs_agent_idx
  ON :POSTGRES_SCHEMA.tenant_agent_configs (tenant_id, agent_id);
CREATE INDEX IF NOT EXISTS tenant_call_tools_tenant_idx
  ON :POSTGRES_SCHEMA.tenant_call_tools (tenant_id, enabled);
CREATE INDEX IF NOT EXISTS tenant_skills_tenant_idx
  ON :POSTGRES_SCHEMA.tenant_skills (tenant_id, enabled);
