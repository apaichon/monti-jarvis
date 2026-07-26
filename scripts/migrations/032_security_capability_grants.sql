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
  :POSTGRES_SCHEMA.users,
  :POSTGRES_SCHEMA.user_roles
TO :TICKET_WRITE_ROLE;

-- Neither capability role receives sequence ownership, DELETE, DDL, or
-- unrestricted schema privileges. Tenant RLS policies are added only after
-- transaction-local app.tenant_id wiring is enabled for the target handlers.
