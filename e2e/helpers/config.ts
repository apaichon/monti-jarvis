// Central E2E configuration, all overridable via environment.
//
// The tests run against the Go server serving the *built* SvelteKit apps
// (tenant-web at base path `/tenant`), so the API and UI are same-origin —
// exactly the production topology.

export const E2E_PORT = process.env.E2E_PORT ?? '8099';
export const BASE_URL = process.env.E2E_BASE_URL ?? `http://localhost:${E2E_PORT}`;

// Postgres — the tenant-register flow requires a live DB. POSTGRES_URL /
// POSTGRES_SCHEMA are loaded from infra/.env.dev in global-setup when not
// already set, matching what the Go server reads.
export const PG_SCHEMA = process.env.POSTGRES_SCHEMA ?? 'callcenter';

// Seeded platform admin (internal/store/auth.go). Password documented in
// docs/sdlc/04-tasks/TASK-0028.md.
export const PLATFORM_EMAIL = process.env.E2E_PLATFORM_EMAIL ?? 'platform@monti.local';
export const PLATFORM_PASSWORD = process.env.E2E_PLATFORM_PASSWORD ?? 'monti-platform';
