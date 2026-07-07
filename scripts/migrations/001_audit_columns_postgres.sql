-- Monti Jarvis Postgres audit columns (idempotent)
-- Adds created_at, updated_at, created_by, updated_by to all callcenter tables.

\set ON_ERROR_STOP on

CREATE TABLE IF NOT EXISTS :POSTGRES_SCHEMA.schema_migrations (
  version text PRIMARY KEY,
  applied_at timestamptz NOT NULL DEFAULT now()
);

CREATE OR REPLACE FUNCTION :POSTGRES_SCHEMA.touch_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

ALTER TABLE :POSTGRES_SCHEMA.calls ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.calls ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.calls ADD COLUMN IF NOT EXISTS created_by text NOT NULL DEFAULT 'system';
ALTER TABLE :POSTGRES_SCHEMA.calls ADD COLUMN IF NOT EXISTS updated_by text NOT NULL DEFAULT 'system';
DROP TRIGGER IF EXISTS trg_calls_touch_updated_at ON :POSTGRES_SCHEMA.calls;
CREATE TRIGGER trg_calls_touch_updated_at
  BEFORE UPDATE ON :POSTGRES_SCHEMA.calls
  FOR EACH ROW EXECUTE FUNCTION :POSTGRES_SCHEMA.touch_updated_at();

ALTER TABLE :POSTGRES_SCHEMA.messages ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.messages ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.messages ADD COLUMN IF NOT EXISTS created_by text NOT NULL DEFAULT 'system';
ALTER TABLE :POSTGRES_SCHEMA.messages ADD COLUMN IF NOT EXISTS updated_by text NOT NULL DEFAULT 'system';
DROP TRIGGER IF EXISTS trg_messages_touch_updated_at ON :POSTGRES_SCHEMA.messages;
CREATE TRIGGER trg_messages_touch_updated_at
  BEFORE UPDATE ON :POSTGRES_SCHEMA.messages
  FOR EACH ROW EXECUTE FUNCTION :POSTGRES_SCHEMA.touch_updated_at();

ALTER TABLE :POSTGRES_SCHEMA.call_sessions ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.call_sessions ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.call_sessions ADD COLUMN IF NOT EXISTS created_by text NOT NULL DEFAULT 'system';
ALTER TABLE :POSTGRES_SCHEMA.call_sessions ADD COLUMN IF NOT EXISTS updated_by text NOT NULL DEFAULT 'system';
UPDATE :POSTGRES_SCHEMA.call_sessions SET created_at = COALESCE(started_at, created_at) WHERE started_at IS NOT NULL;
DROP TRIGGER IF EXISTS trg_call_sessions_touch_updated_at ON :POSTGRES_SCHEMA.call_sessions;
CREATE TRIGGER trg_call_sessions_touch_updated_at
  BEFORE UPDATE ON :POSTGRES_SCHEMA.call_sessions
  FOR EACH ROW EXECUTE FUNCTION :POSTGRES_SCHEMA.touch_updated_at();

ALTER TABLE :POSTGRES_SCHEMA.call_turns ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.call_turns ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.call_turns ADD COLUMN IF NOT EXISTS created_by text NOT NULL DEFAULT 'system';
ALTER TABLE :POSTGRES_SCHEMA.call_turns ADD COLUMN IF NOT EXISTS updated_by text NOT NULL DEFAULT 'system';
DROP TRIGGER IF EXISTS trg_call_turns_touch_updated_at ON :POSTGRES_SCHEMA.call_turns;
CREATE TRIGGER trg_call_turns_touch_updated_at
  BEFORE UPDATE ON :POSTGRES_SCHEMA.call_turns
  FOR EACH ROW EXECUTE FUNCTION :POSTGRES_SCHEMA.touch_updated_at();

ALTER TABLE :POSTGRES_SCHEMA.knowledge_documents ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.knowledge_documents ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.knowledge_documents ADD COLUMN IF NOT EXISTS created_by text NOT NULL DEFAULT 'system';
ALTER TABLE :POSTGRES_SCHEMA.knowledge_documents ADD COLUMN IF NOT EXISTS updated_by text NOT NULL DEFAULT 'system';
DROP TRIGGER IF EXISTS trg_knowledge_documents_touch_updated_at ON :POSTGRES_SCHEMA.knowledge_documents;
CREATE TRIGGER trg_knowledge_documents_touch_updated_at
  BEFORE UPDATE ON :POSTGRES_SCHEMA.knowledge_documents
  FOR EACH ROW EXECUTE FUNCTION :POSTGRES_SCHEMA.touch_updated_at();

ALTER TABLE :POSTGRES_SCHEMA.knowledge_chunks ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.knowledge_chunks ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.knowledge_chunks ADD COLUMN IF NOT EXISTS created_by text NOT NULL DEFAULT 'system';
ALTER TABLE :POSTGRES_SCHEMA.knowledge_chunks ADD COLUMN IF NOT EXISTS updated_by text NOT NULL DEFAULT 'system';
DROP TRIGGER IF EXISTS trg_knowledge_chunks_touch_updated_at ON :POSTGRES_SCHEMA.knowledge_chunks;
CREATE TRIGGER trg_knowledge_chunks_touch_updated_at
  BEFORE UPDATE ON :POSTGRES_SCHEMA.knowledge_chunks
  FOR EACH ROW EXECUTE FUNCTION :POSTGRES_SCHEMA.touch_updated_at();

ALTER TABLE :POSTGRES_SCHEMA.tenants ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.tenants ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.tenants ADD COLUMN IF NOT EXISTS created_by text NOT NULL DEFAULT 'system';
ALTER TABLE :POSTGRES_SCHEMA.tenants ADD COLUMN IF NOT EXISTS updated_by text NOT NULL DEFAULT 'system';
DROP TRIGGER IF EXISTS trg_tenants_touch_updated_at ON :POSTGRES_SCHEMA.tenants;
CREATE TRIGGER trg_tenants_touch_updated_at
  BEFORE UPDATE ON :POSTGRES_SCHEMA.tenants
  FOR EACH ROW EXECUTE FUNCTION :POSTGRES_SCHEMA.touch_updated_at();

ALTER TABLE :POSTGRES_SCHEMA.users ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.users ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.users ADD COLUMN IF NOT EXISTS created_by text NOT NULL DEFAULT 'system';
ALTER TABLE :POSTGRES_SCHEMA.users ADD COLUMN IF NOT EXISTS updated_by text NOT NULL DEFAULT 'system';
DROP TRIGGER IF EXISTS trg_users_touch_updated_at ON :POSTGRES_SCHEMA.users;
CREATE TRIGGER trg_users_touch_updated_at
  BEFORE UPDATE ON :POSTGRES_SCHEMA.users
  FOR EACH ROW EXECUTE FUNCTION :POSTGRES_SCHEMA.touch_updated_at();

ALTER TABLE :POSTGRES_SCHEMA.user_roles ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.user_roles ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.user_roles ADD COLUMN IF NOT EXISTS created_by text NOT NULL DEFAULT 'system';
ALTER TABLE :POSTGRES_SCHEMA.user_roles ADD COLUMN IF NOT EXISTS updated_by text NOT NULL DEFAULT 'system';
DROP TRIGGER IF EXISTS trg_user_roles_touch_updated_at ON :POSTGRES_SCHEMA.user_roles;
CREATE TRIGGER trg_user_roles_touch_updated_at
  BEFORE UPDATE ON :POSTGRES_SCHEMA.user_roles
  FOR EACH ROW EXECUTE FUNCTION :POSTGRES_SCHEMA.touch_updated_at();

ALTER TABLE :POSTGRES_SCHEMA.refresh_tokens ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.refresh_tokens ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now();
ALTER TABLE :POSTGRES_SCHEMA.refresh_tokens ADD COLUMN IF NOT EXISTS created_by text NOT NULL DEFAULT 'system';
ALTER TABLE :POSTGRES_SCHEMA.refresh_tokens ADD COLUMN IF NOT EXISTS updated_by text NOT NULL DEFAULT 'system';
DROP TRIGGER IF EXISTS trg_refresh_tokens_touch_updated_at ON :POSTGRES_SCHEMA.refresh_tokens;
CREATE TRIGGER trg_refresh_tokens_touch_updated_at
  BEFORE UPDATE ON :POSTGRES_SCHEMA.refresh_tokens
  FOR EACH ROW EXECUTE FUNCTION :POSTGRES_SCHEMA.touch_updated_at();

INSERT INTO :POSTGRES_SCHEMA.schema_migrations (version)
VALUES ('001_audit_columns_postgres')
ON CONFLICT (version) DO NOTHING;