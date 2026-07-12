# Monti AI Call Center — 35-Sprint Roadmap

**Blueprint:** `docs/monti_multi_tenant_ai_call_center_blueprint.md` (v2.0)  
**Tech stack:** Svelte + shadcn-svelte · Go + Fiber · Postgres · NATS.io · LiveKit · Redis 8 · MinIO · ClickHouse (analytics + vector RAG)

## Prototype status

| Item | Status |
| --- | --- |
| `monti-jarvis` v0.1.0 Go spike | Shipped — maps to **Sprint 21** (workforce + conversation) |
| Official Sprint 1 | **Shipped v0.2.0** — Customer: Conversation (Svelte + LiveKit + NATS + Gemini voice) |
| Official Sprint 2 | **Shipped v0.3.0** — Customer: KM and Scope (ClickHouse RAG, per-avatar KB, citations) |
| Official Sprint 3 | **Shipped v0.4.0** — Backend: Auth (JWT, RBAC, Redis cache, NATS events) |
| Official Sprint 4 | **Shipped v0.5.0** — Platform Admin: Portal + Packages (login, profile, catalog UI) |
| Official Sprint 5 | **Shipped v0.6.0** — Platform Admin: Avatars (catalog + tenant assignment + portrait upload) |
| Official Sprint 6 | **Shipped v0.7.0** — Tenant: Register (public signup, OAuth, email verify, KYC backoffice; no HeyGen) |
| Official Sprint 7 | **Shipped v0.8.0** — Platform Admin: KYC Tenant (review queue, approve/reject, tenant activation) |
| Official Sprint 8 | **Code shipped** — Payment Gateway (VERIFY with S9) |
| Official Sprint 9–12 | **Commerce chain built** — Buy Package → Billing ledger → Receipt ops → Tax compliance |

---

## Sprint index

| Sprint | Platform | Feature | Phase | Depends on |
| ---: | --- | --- | --- | --- |
| 1 | Customer | Conversation | A | — ✅ v0.2.0 |
| 2 | Customer | Add KM and Scope | A | 1 ✅ v0.3.0 |
| 3 | Backend | Auth | B | — ✅ v0.4.0 |
| 4 | Platform Admin | Packages | B | 3 ✅ v0.5.0 |
| 5 | Platform Admin | Avatars | B | 3, 4 ✅ v0.6.0 |
| 6 | Tenant | Register | C | 3 ✅ v0.7.0 |
| 7 | Platform Admin | KYC Tenant | C | 6 ✅ v0.8.0 |
| 8 | Platform Admin | Payment Gateway | C | 3 ✅ *(code shipped; VERIFY with Sprint 9)* |
| 9 | Tenant | Buy Package (+ MVP receipt/tax) | C | 4, 6, 8 ✅ v1.3.0 |
| 10 | Platform Admin | Billing ledger | C | 9 ✅ v1.3.0 |
| 11 | Platform Admin | Receipt ops | C | 10 ✅ v1.3.0 |
| 12 | Tenant | Tax Invoice compliance | C | 10, 11 ✅ v1.3.0 |
| 13 | Platform Admin | Quota, Rate Limit | B | 3, 4 ✅ v1.4.0 |
| 14 | Tenant | Embed to Web | D | 1, 6 ✅ v1.5.0 |
| 15 | Tenant | Set Scope and KM | D | 2, 6 ✅ v1.6.0 |
| 16 | Tenant | Settings, Locale, Limit user tier, group, Quota for customer call time per day, call minute per each call | D | 13, 15 |
| 17 | Tenant | Test and Preview | D | 15, 16 |
| 18 | Tenant | Customer Tier | D | 16 |
| 19 | Customer | Register | E | 3 |
| 20 | Customer | Auth | E | 19 |
| 21 | Customer | Select AI Workforce to Conversation | A | 1, 5 |
| 22 | Platform / Tenant | Conversation Records, Knowledge Gap | F | 1, 3 |
| 23 | Tenant | Tickets | F | 22 |
| 24 | Tenant | Review | F | 22, 23 |
| 25 | Tenant | Dashboard | F | 22 (ClickHouse) |
| 26 | Tenant | Monitoring | F | 25 |
| 27 | Platform | Audit Log | G | 3 |
| 28 | Platform | Monitoring | G | 27 |
| 29 | Platform | Dashboard | G | 28 (ClickHouse) |
| 30 | Platform | Monitoring | G | 29 |
| 31 | Tuning | gRPC, Cache on Prod | H | 25+ |
| 32 | Tuning | Partition, Hardening | H | 31 |
| 33 | Infra | Scale, Auto Scale | I | 32 |
| 34 | Infra | Canary Deployment | I | 33 |
| 35 | Infra | Backup Restore Archive | I | 33 |

---

## Phase definitions

### Phase A — Customer core (1, 2, 21)

Prove inbound AI call value before billing complexity.

- **Sprint 1:** Svelte customer portal, LiveKit voice room, transcript, NATS call events, Postgres sessions, Redis 8 active state.
- **Sprint 2:** KM ingest → MinIO → embed → ClickHouse `km_embeddings`; scope enforcement; RAG in orchestrator.
- **Sprint 21:** Platform avatar catalog + tenant assignment; workforce picker in conversation UI.

### Phase B — Platform foundation (3, 4, 5, 13)

Multi-tenant SaaS skeleton.

- Auth (JWT/session, RBAC: platform / tenant / customer)
- Commercial packages and platform-managed avatars
- Quota and rate limits (Redis 8 counters + Postgres entitlements)

### Phase C — Tenant commerce (6–12)

Onboarding and monetization (one chain — see [15-commerce-chain-plan.md](../02-design/15-commerce-chain-plan.md)).

- Tenant registration → KYC → payment gateway → **buy package** (method → ChillPay → status → entitlement → **MVP receipt/tax**) → **platform billing** → **receipt ops** → **tax invoice compliance**

### Phase D — Tenant go-live (14–18)

- Web embed widget, tenant KM/scope admin, locale/settings/limits, test sandbox, customer tiers

### Phase E — Customer identity (19–20)

- Optional customer accounts for history and tier benefits

### Phase F — Tenant operations (22–26)

- Conversation records, tickets, QA review, ClickHouse dashboards and monitoring

### Phase G — Platform operations (27–30)

- Cross-tenant audit, monitoring, dashboards

### Phase H — Production tuning (31–32)

- gRPC internal APIs, Redis 8 cache strategy, ClickHouse partitioning, security hardening

### Phase I — Infra scale (33–35)

- Autoscale, canary deployments, backup/restore/archive

---

## Sprint file convention

Each active sprint gets:

```text
docs/sdlc/README.md
docs/sdlc/00-roadmap/ROADMAP.md
docs/sdlc/01-features/FEAT-NNNN-<slug>.md
docs/sdlc/02-design/          01-architecture … 09-platform-admin-portal-spec (NN- prefix)
docs/sdlc/03-sprints/SPRINT-NNN.md
docs/sdlc/04-tasks/TASK-NNNN.md
```

Use `sprint-plan` skill when opening a new sprint.

## Shipped: Phase C commerce (SPRINT-008–012) — v1.3.0 / v1.3.1

**Closed 2026-07-11.** Gateway → buy package → billing ledger → receipt ops → tax compliance.

| Sprint | Feature | UI / API highlights | Release |
| ---: | --- | --- | --- |
| 8 | Payment Gateway | `/admin/settings/payment` | v1.3.0 |
| 9 | Buy Package | `/tenant/billing` method → pay → return + MVP docs | v1.3.0 |
| 10 | Billing ledger | `/admin/billing` · `GET /api/platform/billing/orders` | v1.3.0 |
| 11 | Receipt ops | `/admin/billing/receipts` · void/reissue · seller branding | v1.3.0 |
| 12 | Tax compliance | `/tenant/billing/tax` · `/tenant/billing/documents` | v1.3.0 |

**v1.3.1** — post-ship hardening: ChillPay OrderNo/CustName, browser return fulfill, OAuth login after KYC, billing package card UI, localStorage session.

Plan: [15-commerce-chain-plan.md](../02-design/15-commerce-chain-plan.md)

## Shipped: SPRINT-013 — Quota, Rate Limit — v1.4.0

**Closed 2026-07-11.** Redis quotas + rate limits on chat/voice/KM/avatars; platform usage panel.

| Task | Points | Outcome |
| --- | ---: | --- |
| TASK-0057 | 3 | Redis keys, env, `/api/infra` |
| TASK-0058 | 5 | `internal/quota` service |
| TASK-0059 | 4 | Enforce chat/voice/KM/avatars |
| TASK-0060 | 3 | Platform usage API + UI |
| TASK-0061 | 1 | Manual checklist (full browser UAT deferred) |

Sprint: [SPRINT-013.md](../03-sprints/SPRINT-013.md) · Feature: [FEAT-0013](../01-features/FEAT-0013-quota-rate-limit.md) · Spec: [16-quota-rate-limit-spec.md](../02-design/16-quota-rate-limit-spec.md) · UAT: [SPRINT-013-manual.md](../06-manual-tests/SPRINT-013-manual.md)

## Shipped: SPRINT-014 — Embed to Web — v1.5.0

**Closed 2026-07-12.** Tenant embed key, origin allowlist, loader iframe, portrait/voice/chat embed UI, tenant admin, integrator security guide.

| Task | Points | Outcome |
| --- | ---: | --- |
| TASK-0062 | 3 | `tenant_embed_configs` schema |
| TASK-0063 | 5 | Public resolve + `parent_origin` + tenant APIs |
| TASK-0064 | 4 | Loader JS + `/embed` portrait/voice/chat |
| TASK-0065 | 3 | Tenant `/tenant/embed` admin |
| TASK-0066 | 1 | Manual UAT checklist + unit smoke |

Sprint: [SPRINT-014.md](../03-sprints/SPRINT-014.md) · Feature: [FEAT-0014](../01-features/FEAT-0014-embed-to-web.md) · Spec: [17-embed-to-web-spec.md](../02-design/17-embed-to-web-spec.md) · Guide: [EMBED_WEB_INTEGRATION.md](../../EMBED_WEB_INTEGRATION.md) · UAT: [SPRINT-014-manual.md](../06-manual-tests/SPRINT-014-manual.md)

## Shipped: SPRINT-015 — Set Scope and KM — v1.6.0

**Closed 2026-07-12.** Tenant KM admin UI/APIs, `km_gaps`, multi-tenant RAG for embed, OAuth path rename, QR `bank_qrcode`.

| Task | Points | Outcome |
| --- | ---: | --- |
| TASK-0067 | 3 | Delete cascade + `km_gaps` |
| TASK-0068 | 5 | Tenant KM REST API |
| TASK-0069 | 4 | `/tenant/km` UI |
| TASK-0070 | 3 | Scope matrix + gaps panel |
| TASK-0071 | 1 | Manual UAT checklist |

Sprint: [SPRINT-015.md](../03-sprints/SPRINT-015.md) · Feature: [FEAT-0015](../01-features/FEAT-0015-tenant-scope-km.md) · Spec: [18-tenant-scope-km-spec.md](../02-design/18-tenant-scope-km-spec.md) · UAT: [SPRINT-015-manual.md](../06-manual-tests/SPRINT-015-manual.md)

## Current sprint: SPRINT-016 — Settings, Locale, Limits *(next open)*

**Platform:** Tenant · **Feature:** Settings, Locale, Limit user tier, group, quota for customer call time · **Depends:** 13, 15