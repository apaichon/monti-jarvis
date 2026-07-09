// Postgres helper for the E2E harness.
//
// The tenant-register flow issues an email-verification token whose raw value
// is delivered only by email (Resend) and stored **hashed** (sha256 hex) in
// `email_verification_tokens`. In local/dev there is no mailer, so to drive
// the real /verify-email endpoint end-to-end we stub the *delivery* step: we
// insert a token row ourselves using the same hashing the server uses
// (internal/store/tenantauth.go `hashToken`). No product code is touched.
//
// This is the only place coupled to DB internals (schema, table, hash). If
// those change, update this file.

import { Pool } from 'pg';
import { createHash } from 'node:crypto';
import { PG_SCHEMA } from './config';

const SCHEMA_OK = /^[a-zA-Z_][a-zA-Z0-9_]*$/.test(PG_SCHEMA);

let pool: Pool | null = null;
let probed = false;
let ok = false;

function getPool(): Pool | null {
  if (!process.env.POSTGRES_URL || !SCHEMA_OK) return null;
  if (!pool) {
    pool = new Pool({
      connectionString: process.env.POSTGRES_URL,
      max: 4,
      allowExitOnIdle: true
    });
  }
  return pool;
}

export async function dbAvailable(): Promise<boolean> {
  if (probed) return ok;
  probed = true;
  const p = getPool();
  if (!p) return (ok = false);
  try {
    await p.query('SELECT 1');
    ok = true;
  } catch {
    ok = false;
  }
  return ok;
}

export function sha256hex(raw: string): string {
  return createHash('sha256').update(raw).digest('hex');
}

export async function getUserIdByEmail(email: string): Promise<string | null> {
  const p = getPool();
  if (!p) return null;
  const r = await p.query(`SELECT id FROM ${PG_SCHEMA}.users WHERE email = $1`, [email]);
  return r.rows[0]?.id ?? null;
}

// Mirrors CreateEmailVerificationToken: store sha256(raw) with a 1h expiry.
export async function insertVerificationToken(userId: string, rawToken: string): Promise<void> {
  const p = getPool();
  if (!p) throw new Error('Postgres not available for token injection');
  const id = 'evt_e2e_' + Math.random().toString(36).slice(2, 12);
  const expires = new Date(Date.now() + 60 * 60 * 1000);
  await p.query(
    `INSERT INTO ${PG_SCHEMA}.email_verification_tokens
       (id, user_id, token_hash, expires_at, created_by, updated_by)
     VALUES ($1, $2, $3, $4, 'e2e', 'e2e')`,
    [id, userId, sha256hex(rawToken), expires]
  );
}

// Best-effort removal of everything a single registration created, in FK order.
export async function cleanup(email: string, slug: string): Promise<void> {
  const p = getPool();
  if (!p) return;
  const stmts: Array<[string, unknown[]]> = [
    [`DELETE FROM ${PG_SCHEMA}.email_verification_tokens WHERE user_id IN (SELECT id FROM ${PG_SCHEMA}.users WHERE email=$1)`, [email]],
    [`DELETE FROM ${PG_SCHEMA}.user_roles WHERE user_id IN (SELECT id FROM ${PG_SCHEMA}.users WHERE email=$1)`, [email]],
    [`DELETE FROM ${PG_SCHEMA}.tenant_registrations WHERE admin_email=$1`, [email]],
    [`DELETE FROM ${PG_SCHEMA}.brands WHERE tenant_id IN (SELECT id FROM ${PG_SCHEMA}.tenants WHERE slug=$1)`, [slug]],
    [`DELETE FROM ${PG_SCHEMA}.users WHERE email=$1`, [email]],
    [`DELETE FROM ${PG_SCHEMA}.tenants WHERE slug=$1`, [slug]]
  ];
  for (const [sql, args] of stmts) {
    try {
      await p.query(sql, args);
    } catch {
      // best effort — leftover test rows are harmless (unique per run)
    }
  }
}

export async function closeDb(): Promise<void> {
  if (pool) {
    await pool.end();
    pool = null;
  }
}
