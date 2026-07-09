---
id: SPRINT-008
status: in_progress
start: 2026-07-09
end: 2026-07-23
updated: 2026-07-09
release_target: v0.9.0
goal: "Platform Admin: Payment Gateway — configure Stripe (or mock), test connection, and receive signed webhooks for Sprint 9 checkout."
roadmap_sprint: 8
platform: Platform Admin
depends_on: [SPRINT-003]
---

# SPRINT-008 — Platform Admin: Payment Gateway

## Goal

Let **platform admins** connect a **payment provider** (Stripe test mode or local **mock**), validate credentials, and accept **signed webhooks** — without tenant checkout yet. Unblocks Sprint 9 package purchase.

## Context from prior sprints

- **Sprint 4:** `packages`, `tenant_entitlements`, platform admin portal at `/admin`
- **Sprint 6–7:** Tenant register + KYC; approved tenants are `active`
- **No payment tables or provider code** exist today
- Pattern: optional external deps (Resend) — no-op / mock when unset; infra block on `/api/infra`

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| TASK-0035 | 3 | todo | devops | Payment gateway schema — `payment_gateway_configs` + `payment_webhook_events` |
| TASK-0036 | 5 | todo | dev | Provider abstraction — Stripe + mock adapters; platform config GET/PUT/test APIs |
| TASK-0037 | 3 | todo | dev | Stripe webhook receiver — signature verify, idempotent event log |
| TASK-0038 | 3 | todo | dev | Platform admin UI — `/admin/settings/payment` configure + test connection |
| TASK-0039 | 2 | todo | dev | Infra status, env docs, E2E smoke |

**Committed:** 16 points · **Target velocity:** 16 (avg from Sprints 1–7)

## Scope boundary

**In**
- Single platform-wide gateway config (not per-tenant)
- Providers: `stripe` (test/live mode) and `mock` (local dev)
- Secret handling: prefer `STRIPE_SECRET_KEY` / `STRIPE_WEBHOOK_SECRET` env; DB stores publishable key + metadata
- Platform APIs: get/update config, test connection
- Webhook endpoint logs events only (checkout fulfillment → Sprint 9)
- Platform admin settings screen with feedback dialog pattern from Sprint 6–7
- `sprint-tech-specs` design pack **before** TASK-0036 implementation

**Out** (→ backlog / later sprints)
- Tenant buy-package checkout (Sprint 9)
- PaymentIntent creation for real charges (Sprint 9)
- Billing, invoices, receipts, tax invoices (Sprints 10–12)
- Auto-assign Starter on KYC approve (Sprint 9)
- Omise / regional gateways (interface extensibility only)
- NATS payment events
- PCI — card data never touches Monti server

## Feature

- [FEAT-0008 — Platform Payment Gateway](../01-features/FEAT-0008-payment-gateway.md)

## Design pack (`sprint-tech-specs`)

| Artifact | Path | Status |
| --- | --- | --- |
| Payment gateway deep spec | [13-payment-gateway-spec.md](../02-design/13-payment-gateway-spec.md) | `pending` |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §25–27 | `pending` |
| ER diagram | [03-er-diagram.md](../02-design/03-er-diagram.md) | `pending` |
| API spec | [04-api-spec.md](../02-design/04-api-spec.md) § Payment Gateway | `pending` |
| UX/UI ASCII | [05-ux-ui.md](../02-design/05-ux-ui.md) § P13 | `pending` |

> Run **`/sprint-tech-specs`** before DEV starts **TASK-0035** (schema) / **TASK-0036** (APIs).

## Verification

```bash
make build && make test
make infra-init && make restart
# Platform admin: configure mock provider → test connection
open http://localhost:8091/admin/settings/payment
# Webhook (Stripe CLI or curl with mock signature in dev)
stripe listen --forward-to localhost:8091/api/webhooks/stripe
```

- Manual: `docs/sdlc/06-manual-tests/SPRINT-008-manual.md` (Tester, at VERIFY)
- E2E: `e2e/tests/platform-payment-gateway.spec.ts`

## Risks

| Risk | Mitigation |
| --- | --- |
| Secrets in DB | Store publishable key only; secret from env; mask in API responses |
| Webhook replay | Unique index on `(provider, provider_event_id)` |
| Stripe unavailable in CI | Default `mock` provider; Stripe tests behind build tag or skip if no key |
| Sprint 9 coupling | Webhook handler logs only; document event types Sprint 9 will consume |

## Definition of done

- Design pack approved · code reviewed · ACs verified · `make build` · tag **v0.9.0** at sprint close