import { fileURLToPath } from 'node:url';
import { dirname, resolve } from 'node:path';
import { loadDotEnv } from './helpers/env';

// Load the same infra/.env.dev the Go server reads, so the Node-side Postgres
// helper (token injection / cleanup) can reach the same database. Existing
// environment variables are preserved.
export default function globalSetup(): void {
  const here = dirname(fileURLToPath(import.meta.url));
  loadDotEnv(resolve(here, '..', 'infra', '.env.dev'));
}
