---
id: DES-0040
title: Customer Generative Workspace Specification
status: backlog
hold_reason: security_review_required
updated: 2026-07-25
sprint: SPRINT-044
owner: SA
feature: FEAT-0038
---

# Customer Generative Workspace — Design Spec (ON HOLD)

> **ON HOLD:** This is a retained design record only. Its runtime was removed;
> it must not be treated as an enabled customer or tenant execution path.

**Sprint:** SPRINT-044 · **Release target:** v2.18.0
**Feature:** [FEAT-0038](../01-features/FEAT-0038-customer-generative-ai.md)
**Depends on:** [23-customer-auth-spec.md](23-customer-auth-spec.md),
[24-authenticated-workforce-selection-spec.md](24-authenticated-workforce-selection-spec.md),
[34-platform-billing-quota-ai-cost-spec.md](34-platform-billing-quota-ai-cost-spec.md),
[39-tenant-ai-config-extensibility-spec.md](39-tenant-ai-config-extensibility-spec.md)

## 1. Goals

- Provide one usable server-side provider path: Claude Messages API.
- Keep credentials encrypted at rest, masked in reads, and absent from browser,
  logs, audit events, and error messages.
- Persist tenant/customer job and artifact metadata in Postgres; store bytes in
  MinIO under `km/generative/{tenant}/{customer}/{job}/{artifact}/`.
- Make unsupported provider states explicit and safe until their runtime
  contracts are implemented.

## 2. Non-goals

No arbitrary shell/CLI execution, subscription login, provider marketplace,
unlimited generation, billing overage, public anonymous jobs, or replacement of
Gemini inbound chat/voice.

## 3. Runtime contract

1. Parse the customer bearer token and derive the tenant from the token unless a
   supplied `X-Tenant-Id` or `tenant_id` query value matches it exactly.
2. Apply the generation Redis rate bucket (default 10 jobs/minute per tenant).
3. Create or reuse a Postgres job by `(tenant_id, customer_id, idempotency_key)`.
4. Run Claude with a server timeout, bounded prompt/output body, and a fixed
   safety system instruction. Never pass credentials to the browser.
5. Convert the typed response to one artifact, upload it to the tenant/customer
   prefix, then write metadata and mark the job complete. On failure, persist a
   bounded safe code/message and no provider response body.
6. Record idempotent AI usage and redacted audit metadata only after the logical
   job result is known.

## 4. Provider states

| Provider | Release state | Credential path | Outputs |
| --- | --- | --- | --- |
| Claude | usable | API key | html, image, canvas, link, report, doc |
| Codex | not_configured | not accepted yet | html, canvas, doc |
| Antigravity | unsupported | not accepted | html, canvas, doc |
| Grok CLI | not_configured | not accepted yet | html, report, doc |

## 5. Data model

All tables are in Postgres `callcenter`, have audit columns, and enforce tenant
and customer foreign keys. Migration: `scripts/migrations/029_customer_generative_ai.sql`.

| Table | Contract |
| --- | --- |
| `customer_generation_connections` | One provider row per tenant/customer; encrypted API-key ciphertext/nonce, version, last4, mode, status, expiry. |
| `customer_generation_jobs` | Bounded prompt, provider/output type, queued/running/completed/failed state, attempts, safe error, idempotency key. |
| `customer_generation_artifacts` | Type, MIME, filename, object key, size, SHA-256, tenant/customer/job ownership. |

## 6. Security and failure policy

- `TENANT_SECRET_ENCRYPTION_KEY` is required to save API keys.
- Expired credentials fail closed with `credential_expired`.
- Cross-tenant or cross-customer reads return `not_found`/`tenant_context_mismatch`.
- HTML and SVG artifact responses carry a sandbox CSP; the UI uses a sandboxed
  iframe for HTML previews.
- Raw provider errors, prompts containing customer content, credentials, and
  generated bodies are never included in audit events.
- Duplicate create requests reuse the existing job and do not start a second
  provider invocation or metering event.

## 7. Verification curl block

```sh
curl -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  https://example.test/api/customer/generative/providers

curl -X PUT -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"mode":"api_key","api_key":"<key>"}' \
  https://example.test/api/customer/generative/connections/claude

curl -X POST -H "Authorization: Bearer $CUSTOMER_TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"provider":"claude","output_type":"html","prompt":"Create a small landing page"}' \
  https://example.test/api/customer/generative/jobs
```

See [02-workflow.md](02-workflow.md) §97–99, [03-er-diagram.md](03-er-diagram.md),
[04-api-spec.md](04-api-spec.md), and [05-ux-ui.md](05-ux-ui.md).
