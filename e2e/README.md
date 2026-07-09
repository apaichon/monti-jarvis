# SPRINT-006 E2E — Tenant Register (Playwright)

End-to-end tests for the tenant self-registration sprint. They drive a real
Chromium browser against the Go server serving the **built** SvelteKit apps
(`/tenant`, `/admin`), so UI + API are same-origin — the production topology.

## Coverage → sprint tasks

| Spec | Covers |
| --- | --- |
| `tests/tenant-register.spec.ts` | **TASK-0029** form (fields, slug auto-suggest, client validation) · **TASK-0026** register API (happy path → verify-email prompt, duplicate slug `409`, duplicate email `409`) |
| `tests/tenant-verify-login.spec.ts` | **TASK-0027** unverified login blocked (`403 email not verified`), email-verification issues a session, verified admin logs in → backoffice |
| `tests/platform-tenants.spec.ts` | **TASK-0028** `GET /api/platform/tenants` guards (`401`/`403`), pending tenant listed for platform admin (API + admin UI) |

TASK-0025 (schema) is exercised implicitly — every DB-backed test relies on the
migrated tables.

> Note: the implemented flow requires **email verification** before login (the
> register endpoint returns `verification_required` rather than tokens). That
> grew past the sprint doc's original "Out of scope" note, so these tests cover
> the **actual** behavior. Where a task's documented acceptance criteria differ
> from what shipped, the test asserts the shipped behavior.

## Prerequisites

- Node 18+ and the repo's Go toolchain.
- Local infra (**Postgres** required) running and migrated:
  ```bash
  make infra-check     # from repo root
  make infra-init
  ```
  Postgres-dependent tests **skip** (not fail) when `POSTGRES_URL` is
  unreachable, so the pure-UI/validation tests still run anywhere.

## Run

```bash
cd e2e
npm install
npm run install:browsers      # one-time: playwright install chromium
npm test
```

On first run `scripts/ensure-server.sh` runs `make build` and starts a
dedicated server on **:8099**. Later runs reuse it (`reuseExistingServer`).

Useful flags:

```bash
npm run test:headed           # watch the browser
npm run test:ui               # Playwright UI mode
npm run report                # open the last HTML report
E2E_SKIP_BUILD=1 npm test     # reuse an existing ./monti-jarvis binary
E2E_BASE_URL=http://localhost:8091 npm test   # target a server you manage
```

## How email verification is handled

The verification token is delivered only by email (Resend) and stored **hashed**
in Postgres. With no dev mailer, the tests stub the *delivery* step only:
`helpers/db.ts` inserts a token row using the same `sha256` hashing the server
uses, then the tests drive the real `/verify-email` endpoint and verify page.
**No product code is modified.** This is the one place coupled to DB internals
(schema/table/hash); update it if those change.

## Config knobs (env)

| Var | Default | Purpose |
| --- | --- | --- |
| `E2E_PORT` | `8099` | dedicated test server port |
| `E2E_BASE_URL` | `http://localhost:8099` | target origin |
| `POSTGRES_URL` / `POSTGRES_SCHEMA` | from `infra/.env.dev` | DB for token injection + cleanup |
| `E2E_PLATFORM_EMAIL` / `E2E_PLATFORM_PASSWORD` | `platform@monti.local` / `monti-platform` | seeded platform admin |
| `E2E_SKIP_BUILD` | `0` | skip `make build` in the server bootstrap |

Test tenants use unique slugs/emails and are cleaned up from Postgres after each
test (best-effort).
