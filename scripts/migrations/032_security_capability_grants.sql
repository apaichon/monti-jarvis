-- Sprint 41: capability-specific database grants.
--
-- Roles are provisioned by deployment/IAM, not created with application
-- passwords in this migration. The migration is applied only when
-- POSTGRES_KM_READONLY_ROLE and POSTGRES_TICKET_WRITE_ROLE are configured.
\set ON_ERROR_STOP on

GRANT USAGE ON SCHEMA :POSTGRES_SCHEMA TO :KM_READ_ROLE;
GRANT SELECT ON TABLE
  :POSTGRES_SCHEMA.knowledge_documents,
  :POSTGRES_SCHEMA.knowledge_chunks,
  :POSTGRES_SCHEMA.tenant_settings,
  :POSTGRES_SCHEMA.tenant_agent_configs,
  :POSTGRES_SCHEMA.tenant_embed_configs
TO :KM_READ_ROLE;

ALTER TABLE :POSTGRES_SCHEMA.knowledge_documents ENABLE ROW LEVEL SECURITY;
ALTER TABLE :POSTGRES_SCHEMA.knowledge_documents FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS knowledge_documents_tenant_boundary ON :POSTGRES_SCHEMA.knowledge_documents;
CREATE POLICY knowledge_documents_tenant_boundary ON :POSTGRES_SCHEMA.knowledge_documents
  USING (tenant_id = current_setting('app.tenant_id', true))
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true));

ALTER TABLE :POSTGRES_SCHEMA.knowledge_chunks ENABLE ROW LEVEL SECURITY;
ALTER TABLE :POSTGRES_SCHEMA.knowledge_chunks FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS knowledge_chunks_tenant_boundary ON :POSTGRES_SCHEMA.knowledge_chunks;
CREATE POLICY knowledge_chunks_tenant_boundary ON :POSTGRES_SCHEMA.knowledge_chunks
  USING (tenant_id = current_setting('app.tenant_id', true))
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true));

GRANT USAGE ON SCHEMA :POSTGRES_SCHEMA TO :TICKET_WRITE_ROLE;
GRANT SELECT, INSERT, UPDATE ON TABLE
  :POSTGRES_SCHEMA.tickets,
  :POSTGRES_SCHEMA.ticket_events
TO :TICKET_WRITE_ROLE;
GRANT SELECT ON TABLE
  :POSTGRES_SCHEMA.ai_avatars,
  :POSTGRES_SCHEMA.call_sessions,
  :POSTGRES_SCHEMA.conversation_records,
  :POSTGRES_SCHEMA.customers,
  :POSTGRES_SCHEMA.tenants,
  :POSTGRES_SCHEMA.users,
  :POSTGRES_SCHEMA.user_roles
TO :TICKET_WRITE_ROLE;

-- Ticket handlers set app.tenant_id with SET LOCAL inside every transaction.
-- FORCE makes the policy apply even when the migration owner is also the table
-- owner; an absent context therefore returns no rows and rejects writes.
ALTER TABLE :POSTGRES_SCHEMA.tickets ENABLE ROW LEVEL SECURITY;
ALTER TABLE :POSTGRES_SCHEMA.tickets FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS tickets_tenant_boundary ON :POSTGRES_SCHEMA.tickets;
CREATE POLICY tickets_tenant_boundary ON :POSTGRES_SCHEMA.tickets
  USING (tenant_id = current_setting('app.tenant_id', true))
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true));

ALTER TABLE :POSTGRES_SCHEMA.ticket_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE :POSTGRES_SCHEMA.ticket_events FORCE ROW LEVEL SECURITY;
DROP POLICY IF EXISTS ticket_events_tenant_boundary ON :POSTGRES_SCHEMA.ticket_events;
CREATE POLICY ticket_events_tenant_boundary ON :POSTGRES_SCHEMA.ticket_events
  USING (tenant_id = current_setting('app.tenant_id', true))
  WITH CHECK (tenant_id = current_setting('app.tenant_id', true));

-- Neither capability role receives sequence ownership, DELETE, DDL, or
-- unrestricted schema privileges. All KM document/chunk access requires the
-- transaction-local tenant context set by the application.
