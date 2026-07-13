package store

import (
	"context"
	"fmt"
)

// auditColumnsDDL is appended to every callcenter table CREATE statement.
const auditColumnsDDL = `
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  created_by text NOT NULL DEFAULT 'system',
  updated_by text NOT NULL DEFAULT 'system'`

var auditableTables = []string{
	"calls",
	"messages",
	"call_sessions",
	"call_turns",
	"knowledge_documents",
	"knowledge_chunks",
	"tenants",
	"users",
	"user_roles",
	"refresh_tokens",
	"package_rule_schemas",
	"packages",
	"package_limits",
	"tenant_entitlements",
	"embedding_models",
	"voice_providers",
	"ai_avatars",
	"ai_avatar_voices",
	"tenant_avatar_assignments",
	"payment_gateway_configs",
	"payment_callback_events",
	"payment_orders",
	"tenant_embed_configs",
	"km_gaps",
	"tenant_settings",
	"tenant_call_limits",
	"customer_tiers",
	"customer_groups",
	"customers",
	"customer_group_members",
	"customer_import_jobs",
	"customer_domain_rules",
	"tenant_customer_auth_settings",
	"customer_auth_identities",
	"customer_otp_challenges",
	"customer_sessions",
	"customer_auth_events",
}

func (s *Store) ensureAuditSchema(ctx context.Context) error {
	if s.pg == nil {
		return nil
	}
	schema := quoteIdent(s.cfg.PostgresSchema)

	if _, err := s.pg.Exec(ctx, fmt.Sprintf(`
CREATE TABLE IF NOT EXISTS %s.schema_migrations (
  version text PRIMARY KEY,
  applied_at timestamptz NOT NULL DEFAULT now()
)`, schema)); err != nil {
		return err
	}

	if _, err := s.pg.Exec(ctx, fmt.Sprintf(`
CREATE OR REPLACE FUNCTION %s.touch_updated_at()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql`, schema)); err != nil {
		return err
	}

	for _, table := range auditableTables {
		qualified := fmt.Sprintf("%s.%s", schema, quoteIdent(table))
		stmts := []string{
			fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS created_at timestamptz NOT NULL DEFAULT now()`, qualified),
			fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS updated_at timestamptz NOT NULL DEFAULT now()`, qualified),
			fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS created_by text NOT NULL DEFAULT 'system'`, qualified),
			fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS updated_by text NOT NULL DEFAULT 'system'`, qualified),
		}
		for _, stmt := range stmts {
			if _, err := s.pg.Exec(ctx, stmt); err != nil {
				return fmt.Errorf("%s: %w", table, err)
			}
		}

		trigger := fmt.Sprintf("trg_%s_touch_updated_at", table)
		if _, err := s.pg.Exec(ctx, fmt.Sprintf(`DROP TRIGGER IF EXISTS %s ON %s`, quoteIdent(trigger), qualified)); err != nil {
			return err
		}
		if _, err := s.pg.Exec(ctx, fmt.Sprintf(`
CREATE TRIGGER %s
BEFORE UPDATE ON %s
FOR EACH ROW EXECUTE FUNCTION %s.touch_updated_at()`,
			quoteIdent(trigger), qualified, schema)); err != nil {
			return err
		}
	}

	if _, err := s.pg.Exec(ctx, fmt.Sprintf(`
UPDATE %s.call_sessions
SET created_at = COALESCE(started_at, created_at)
WHERE started_at IS NOT NULL`, schema)); err != nil {
		return err
	}

	_, err := s.pg.Exec(ctx, fmt.Sprintf(`
INSERT INTO %s.schema_migrations (version)
VALUES ('001_audit_columns_postgres')
ON CONFLICT (version) DO NOTHING`, schema))
	return err
}
