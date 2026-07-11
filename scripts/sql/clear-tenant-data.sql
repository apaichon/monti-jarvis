-- =============================================================================
-- Clear ALL tenant data for re-testing: register → KYC → buy package → pay
-- Schema: callcenter (Monti Jarvis)
--
-- KEEPS (platform seed / catalog):
--   • platform_admin user  (platform@monti.local)
--   • packages + package_limits + package_rule_schemas
--   • payment_gateway_configs
--   • ai_avatars / voice_providers / embedding_models
--   • schema_migrations
--
-- REMOVES:
--   • every tenant (including demo)
--   • tenant admins / customers, OAuth, refresh + verify tokens
--   • registrations, KYC, brands, entitlements, avatar assignments
--   • payment_orders + payment_documents + callback events
--   • call/session/KM rows tied to tenants
--
-- Usage:
--   psql "$POSTGRES_URL" -v ON_ERROR_STOP=1 -f scripts/sql/clear-tenant-data.sql
--
-- Then restart so seed re-creates demo tenant + admin@demo.local:
--   make restart
--
-- Optional Redis flush for entitlement cache (after SQL):
--   redis-cli -n 4 KEYS 'monti_jarvis:*' | xargs -r redis-cli -n 4 DEL
--   # or only entitlements:
--   redis-cli -n 4 --scan --pattern 'monti_jarvis:entitlement:*' | xargs -r redis-cli -n 4 DEL
-- =============================================================================

BEGIN;

SET search_path TO callcenter, public;

-- Snapshot before wipe
SELECT 'tenants_before' AS metric, count(*)::text AS value FROM tenants
UNION ALL
SELECT 'users_before', count(*)::text FROM users
UNION ALL
SELECT 'payment_orders_before', count(*)::text FROM payment_orders;

-- ---------------------------------------------------------------------------
-- 1) Payment commerce (explicit; CASCADE from tenants also covers these)
-- ---------------------------------------------------------------------------
DELETE FROM payment_documents;
DELETE FROM payment_orders;
DELETE FROM payment_callback_events;

-- ---------------------------------------------------------------------------
-- 2) Tenant-scoped product data
-- ---------------------------------------------------------------------------
DELETE FROM tenant_entitlements;
DELETE FROM tenant_avatar_assignments;
DELETE FROM tenant_kyc_profiles;
DELETE FROM tenant_registrations;
DELETE FROM brands;

-- Knowledge / calls (tenant_id text, may not have FK)
DELETE FROM knowledge_chunks
WHERE document_id IN (SELECT id FROM knowledge_documents);
DELETE FROM knowledge_documents;
DELETE FROM call_turns;
DELETE FROM call_sessions;
DELETE FROM messages;
DELETE FROM calls;

-- ---------------------------------------------------------------------------
-- 3) Auth for non-platform users
--    Users are not FK'd to tenants; tenant_id lives on user_roles.
-- ---------------------------------------------------------------------------
DELETE FROM email_verification_tokens
WHERE user_id IN (
  SELECT u.id FROM users u
  WHERE u.id <> 'usr_platform'
    AND u.email <> 'platform@monti.local'
);

DELETE FROM refresh_tokens
WHERE user_id IN (
  SELECT u.id FROM users u
  WHERE u.id <> 'usr_platform'
    AND u.email <> 'platform@monti.local'
);

DELETE FROM user_oauth_identities
WHERE user_id IN (
  SELECT u.id FROM users u
  WHERE u.id <> 'usr_platform'
    AND u.email <> 'platform@monti.local'
);

DELETE FROM user_roles
WHERE role <> 'platform_admin'
   OR user_id <> 'usr_platform';

-- Keep only platform_admin account
DELETE FROM users
WHERE id <> 'usr_platform'
  AND email <> 'platform@monti.local';

-- ---------------------------------------------------------------------------
-- 4) Tenants last (CASCADE cleans any remaining FK children)
-- ---------------------------------------------------------------------------
DELETE FROM tenants;

-- Ensure platform role still present
INSERT INTO user_roles (user_id, role, tenant_id)
SELECT 'usr_platform', 'platform_admin', NULL
WHERE EXISTS (SELECT 1 FROM users WHERE id = 'usr_platform')
ON CONFLICT (user_id, role) DO NOTHING;

-- Snapshot after wipe
SELECT 'tenants_after' AS metric, count(*)::text AS value FROM tenants
UNION ALL
SELECT 'users_after', count(*)::text FROM users
UNION ALL
SELECT 'users_emails', string_agg(email, ', ' ORDER BY email) FROM users
UNION ALL
SELECT 'packages_kept', count(*)::text FROM packages
UNION ALL
SELECT 'gateway_kept', count(*)::text FROM payment_gateway_configs;

COMMIT;

-- =============================================================================
-- After this script:
--   make restart
-- Seed restores:
--   tenant  demo / admin@demo.local  password demo-admin
--   platform@monti.local             password monti-platform
-- Then re-test: register new tenant → verify email → KYC → buy package → pay
-- =============================================================================
