# Feature: Platform Payment Gateway   (FEAT-0008)
**Sprint:** SPRINT-008   **Owner:** DEV   **Status:** active   **Release target:** v0.9.0

## Problem

Sprints 4–7 delivered packages, tenant onboarding, and KYC approval, but there is **no payment provider** configured. Platform operators cannot connect Stripe (or a dev mock) for tenant commerce. Sprint 9 (Buy Package) needs a **validated gateway config** and webhook receiver before checkout can ship.

## Scope

In:
- Postgres `payment_gateway_configs` — platform-level provider settings (provider id, mode, masked secrets, status)
- Postgres `payment_webhook_events` — idempotent webhook audit log (provider event id, type, payload hash)
- `internal/payment` — `Provider` interface; **Stripe** test-mode adapter; **mock** provider for local dev without keys
- Platform APIs (`platform_admin` only):
  - `GET /api/platform/payment-gateway` — current config + connection status
  - `PUT /api/platform/payment-gateway` — set provider, publishable key, secret key ref (env-backed), webhook secret, mode
  - `POST /api/platform/payment-gateway/test` — validate credentials (Stripe `balance` or mock ping)
- Webhook: `POST /api/webhooks/stripe` — signature verify, persist event, `200` ack (no checkout side-effects yet)
- Platform admin UI `/admin/settings/payment` — configure provider, test connection, view last webhook status
- `GET /api/infra` adds `payment_gateway` block (configured, provider, mode)
- Design pack via `sprint-tech-specs` (`13-payment-gateway-spec.md`)

Out:
- Tenant checkout / package purchase UI (Sprint 9)
- Invoices, receipts, tax invoices (Sprints 10–12)
- Subscription billing cycles and usage metering (Sprint 10+)
- Multiple concurrent providers per tenant (single platform default only)
- PCI card collection in Monti UI (Stripe Checkout / Elements in Sprint 9)
- Omise / 2C2P adapters (future; interface ready)
- Refunds, disputes, payout reporting

## Acceptance criteria

1. `ensureSchema` creates `payment_gateway_configs` (singleton row) and `payment_webhook_events` idempotently with audit columns.
2. `GET /api/platform/payment-gateway` returns provider, mode (`test`|`live`), masked secret tail, `webhook_configured`, `connection_status`; `403` for non-platform-admin.
3. `PUT /api/platform/payment-gateway` accepts `provider` (`stripe`|`mock`), keys, mode; stores secrets via env override pattern (`STRIPE_SECRET_KEY`) when set; persists publishable key in DB; returns updated config.
4. `POST /api/platform/payment-gateway/test` returns `200 {ok:true}` when Stripe test key valid or mock provider; `502` with message when keys invalid.
5. `POST /api/webhooks/stripe` verifies `Stripe-Signature`, inserts event once (idempotent on `provider_event_id`), returns `200`; rejects bad signature with `400`.
6. `/admin/settings/payment` — form for provider + keys + mode, **Test connection** button, feedback dialog on success/fail; nav link in platform shell.
7. `/api/infra` includes `payment_gateway: {configured, provider, mode}`.
8. `go test ./...`; manual UAT in `docs/sdlc/06-manual-tests/SPRINT-008-manual.md` (Tester, at VERIFY).

## Test notes

- API: RBAC, secret masking, idempotent webhook insert, signature verification unit tests
- Dev without Stripe: `mock` provider passes test endpoint
- Browser UAT: platform admin configures mock → test OK → simulate webhook curl
- E2E: `e2e/tests/platform-payment-gateway.spec.ts` (config page smoke)

## Links

- Sprint: [SPRINT-008](../03-sprints/SPRINT-008.md)
- Depends on: [FEAT-0003](FEAT-0003-auth-rbac.md) (platform_admin RBAC)
- Prior commerce: [FEAT-0004](FEAT-0004-packages-entitlements.md)
- Design: `13-payment-gateway-spec.md` (pending) · [04-api-spec.md](../02-design/04-api-spec.md) § Payment Gateway · [05-ux-ui.md](../02-design/05-ux-ui.md) § P13
- Roadmap: Sprint 8 · Phase C
- Next: Sprint 9 Buy Package