---
id: DES-0044
title: AI Call-Center Security Hardening Specification
status: review_pending
updated: 2026-07-26
sprint: SPRINT-041
owner: SA
---

# DES-0044 — AI Call-Center Security Hardening

## Goals

1. Keep browser persistence from becoming a durable plaintext credential store.
2. Fail closed when production secrets or database privileges are unsafe.
3. Make read access materially less powerful than write access.
4. Enforce tenant scope at both the application and database layers for the
   selected high-risk read paths.
5. Produce repeatable evidence without exposing credentials, prompts,
   transcripts, or tenant data.

## Non-goals

- A new identity provider, SSO, or customer-auth protocol.
- A complete rewrite of all historical SQL or every table into RLS at once.
- Network segmentation, WAF procurement, HSM/KMS deployment, or device trust.
- Encrypting secrets in browser storage with a key stored next to the secret.

## Security invariants

| Invariant | Enforcement |
| --- | --- |
| Browser never receives server secrets | Build-time bundle scan, response redaction, metadata-only APIs |
| Refresh credentials are not durable plaintext | HttpOnly cookie preferred; versioned migration removes legacy local/session keys |
| Production cannot run with weak auth/config | Startup validator fails closed; development bypass is explicit and non-production |
| Read role cannot mutate | PostgreSQL `CONNECT` + schema/table `SELECT` grants only; no sequence write or DDL grants |
| Tenant context is request-local | `SET LOCAL app.tenant_id` inside a transaction; rollback on every path |
| Tenant scope is deny-by-default | Application authorization plus PostgreSQL policy for selected tables |
| Query shape is bounded | Parameters, allowlisted identifiers, max page size, fixed sort map |

## Environment and secret contract

| Variable | Required | Rule |
| --- | --- | --- |
| `APP_ENV` | yes | `production` rejects development bypasses and unsafe cookie/CORS settings |
| `JWT_SECRET` | production | At least 32 bytes; never returned by health or diagnostics |
| `POSTGRES_URL` | yes | Writer URL; must not be exposed to any web build or browser response |
| `POSTGRES_READONLY_URL` | production/read paths | Dedicated role URL; startup verifies it is distinct from writer authority |
| `POSTGRES_RLS_ENFORCED` | production | Must be true for tables declared in the policy inventory |
| `COOKIE_SECURE` | production | Must be true when refresh credentials use cookies |
| `ALLOWED_ORIGINS` | production | Explicit allowlist; wildcard is rejected with credentials enabled |

Startup logs may emit variable names, booleans, role names, and remediation
codes. They must never emit values, URLs with credentials, JWTs, cookies,
provider keys, SQL, prompts, or request bodies.

## Browser storage contract

### Current risk

Tenant and customer web clients currently persist access/refresh tokens in
browser storage, while OAuth/preview flows can briefly carry tokens in query
parameters. A browser-storage wrapper alone cannot defend against XSS because
the page can also access the decryption key.

### Target behavior

- Migrate refresh credentials to an HttpOnly, Secure, SameSite cookie set by
  the server; access tokens remain memory-first with a short-lived fallback
  only where an existing redirect requires it.
- Remove query-string tokens immediately after bootstrap and never log the URL.
- Use Web Crypto AES-GCM only for approved low-sensitivity preferences or a
  migration envelope, with a non-persistent session key and expiry metadata.
- On logout, expiry, migration failure, or tenant switch, clear current and
  known legacy keys and revoke the server session when possible.
- Reject unversioned/plaintext storage after the migration grace period and
  show a safe re-authentication state.

## Runtime and database boundary

The Go server owns two pools when configured: a writer pool for mutations and
a read-only pool for AI/reporting reads. Read handlers must select the pool
explicitly; a missing read pool is a startup error for production paths that
declare read-only operation mandatory.

The proposed migration `scripts/migrations/032_security_capability_grants.sql` adds or
verifies grants and policies; it does not create business entities. RLS is
enabled only for the approved inventory first, with application fallback and
an explicit policy-inventory check for tables not yet migrated.

## API-to-database-user routing

The API contract separates database identities by capability. The database
user is never selected from a client-supplied tenant ID; Go resolves the
authenticated tenant, chooses the capability pool, and sets transaction-local
tenant context before any query.

| API / runtime path | Principal | Allowed | Explicitly denied |
| --- | --- | --- | --- |
| `POST /api/chat` and `GET /ws/voice` KM/RAG retrieval | `monti_ai_km_ro` | Read approved `knowledge_documents`, `knowledge_chunks`, scopes, and agent-KM context for the authenticated tenant | INSERT/UPDATE/DELETE/DDL, ticket creation, cross-tenant reads |
| Tenant preview chat/voice KM retrieval | `monti_ai_km_ro` | Same read-only tenant-scoped KM contract; preview limits still apply | KM mutation, ticket writes, cross-tenant reads |
| `POST /api/tenant/km/*` ingest/reset/update | `monti_km_rw` | KM management writes for the authenticated tenant and permitted tenant role | AI runtime use of write credential, other tenant data, ticket mutation |
| `POST /api/customer/tickets` and ticket events/status | `monti_ticket_rw` | Append ticket/ticket-event rows and read the minimum scoped source data needed for the ticket | KM writes, entitlement/package writes, arbitrary conversation mutation, cross-tenant access |
| General controlled mutations | `monti_app_rw` | Existing bounded application writes not covered by the specialized users | Browser/client direct database access |

The preferred shared-Postgres model is one capability role per service path,
combined with RLS and `SET LOCAL app.tenant_id`. Where tenants have dedicated
database instances, each tenant receives distinct capability credentials and
those credentials are never reused for another tenant. A different database
user alone is not considered tenant isolation; the authenticated context,
policy, and query allowlist remain mandatory.

For a tenant-scoped read:

1. Authenticate and authorize the actor in Go.
2. Begin a transaction on the read pool.
3. Execute `SELECT set_config('app.tenant_id', $1, true)`.
4. Run parameterized SQL with bounded filters and allowlisted sort fields.
5. Commit/rollback before returning the connection to the pool.

No handler may set tenant context with a session-level `SET`, and no request
may trust a client-supplied tenant ID over the authenticated subject.

If an AI conversation needs to persist usage, session, or transcript metadata,
that write is dispatched to a separate controlled writer path after the
read-only KM retrieval. The AI/KM read principal itself remains incapable of
creating tickets or mutating KM.

## API summary

| Method | Path | Role | Purpose |
| --- | --- | --- | --- |
| `GET` | `/api/platform/security/posture` | platform admin | Metadata-only posture checks and remediation codes |
| `GET` | `/healthz` | public/operator | Existing liveness; never includes secret or tenant data |
| `GET` | `/readyz` | operator | Readiness includes configured pool/policy booleans, not credentials |

### Capability routing requirements

| API | Database user used by the handler path | Tenant rule |
| --- | --- | --- |
| `POST /api/chat` / `GET /ws/voice` KM lookup | `monti_ai_km_ro` | Tenant is derived from auth/embed context; no client tenant override |
| `POST /api/customer/tickets` | `monti_ticket_rw` | Ticket and source conversation are checked against the authenticated tenant |
| `GET/PATCH /api/tenant/tickets/{id}` and `POST /events` | `monti_ticket_rw` | Wrong-tenant IDs return safe `404 not_found`; no cross-tenant detail |
| `POST /api/tenant/km/*` | `monti_km_rw` | Tenant admin and active-tenant checks precede all writes |

The AI conversation path must not borrow `monti_ticket_rw` or the general
writer pool merely because a ticket may later be offered. Ticket creation is a
separate, customer-confirmed API transaction.

The posture endpoint is diagnostic only: it cannot mutate secrets, grants,
roles, policies, or tenant data. Isolation probes remain test-only fixtures and
are not exposed as a public production API.

## RBAC

- Platform admins may view aggregate posture metadata.
- Tenant users cannot view database role names, secret configuration, or other
  tenants' posture.
- Customers have no new security endpoint.
- Dev bypasses are allowed only under explicit non-production configuration and
  are reported as `degraded`, never `healthy`.

## Verification curl block

```bash
# Metadata only; response must contain booleans/codes and no secret values.
curl -fsS -H "Authorization: Bearer $PLATFORM_TOKEN" \
  "$BASE_URL/api/platform/security/posture" | jq

# Readiness must fail or report degraded when the read-only role is absent.
curl -fsS "$BASE_URL/readyz" | jq

# The two-tenant isolation and read-role denial checks run in the isolated
# test fixture; no production tenant ID or credential is used here.
go test ./internal/store ./cmd/server -run 'Security|Isolation|ReadOnly' -count=1
```

## Approval gate

Implementation requires security-owner approval of the cookie migration,
read-pool routing inventory, RLS policy inventory, and production fail-closed
rules. Status remains `review_pending` until that review is recorded.
