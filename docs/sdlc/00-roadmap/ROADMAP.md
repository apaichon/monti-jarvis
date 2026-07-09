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
| Official Sprint 8 | **In progress** — Platform Admin: Payment Gateway (ChillPay/mock config, callbacks) |

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
| 9 | Tenant | Buy Package | C | 4, 6, 8 🔄 |
| 10 | Platform Admin | Billing | C | 9 |
| 11 | Platform Admin | Receipt | C | 10 |
| 12 | Tenant | Tax Invoice | C | 10, 11 |
| 13 | Platform Admin | Quota, Rate Limit | B | 3, 4 |
| 14 | Tenant | Embed to Web | D | 1, 6 |
| 15 | Tenant | Set Scope and KM | D | 2, 6 |
| 16 | Tenant | Settings, Locale, Limit, Quota | D | 13, 15 |
| 17 | Tenant | Test and Preview | D | 15, 16 |
| 18 | Tenant | Customer Tier | D | 16 |
| 19 | Customer | Register | E | 3 |
| 20 | Customer | Auth | E | 19 |
| 21 | Customer | Select AI Workforce to Conversation | A | 1, 5 |
| 22 | Platform / Tenant | Conversation Records | F | 1, 3 |
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

Onboarding and monetization.

- Tenant registration → KYC → payment gateway → package purchase → billing → receipt → tax invoice

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

## Current sprint: SPRINT-009 *(in progress)*

**Platform:** Tenant  
**Feature:** Buy Package  
**Goal:** Tenant ChillPay checkout, callback fulfillment, entitlement — **combined E2E verify with SPRINT-008 gateway**.

**Release target:** v1.0.0

See [SPRINT-009](../03-sprints/SPRINT-009.md) · [FEAT-0009](../01-features/FEAT-0009-buy-package.md).

## Prior: SPRINT-008 *(code shipped; UAT with Sprint 9)*

**Platform:** Platform Admin · **Feature:** Payment Gateway · **v0.9.0** *(tag at combined close)*

See [SPRINT-008](../03-sprints/SPRINT-008.md) · [FEAT-0008](../01-features/FEAT-0008-payment-gateway.md).

## Last shipped: SPRINT-007

**Platform:** Platform Admin · **Feature:** KYC Tenant · **v0.8.0**

See [SPRINT-007](../03-sprints/SPRINT-007.md) · [FEAT-0007](../01-features/FEAT-0007-kyc-tenant.md).

## Prior: SPRINT-006

**Platform:** Tenant · **Feature:** Register · **v0.7.0**

See [SPRINT-006](../03-sprints/SPRINT-006.md) · [FEAT-0006](../01-features/FEAT-0006-tenant-register.md).

## Prior: SPRINT-005

**Platform:** Platform Admin · **Feature:** Avatars · **v0.6.0**

See [SPRINT-005](../03-sprints/SPRINT-005.md) · [FEAT-0005](../01-features/FEAT-0005-avatar-catalog.md).

## Prior: SPRINT-004

**Platform:** Platform Admin · **Feature:** Portal + Packages · **v0.5.0**

See [SPRINT-004](../03-sprints/SPRINT-004.md) · [FEAT-0004](../01-features/FEAT-0004-packages-entitlements.md).

## Next sprint: SPRINT-010

**Platform:** Platform Admin  
**Feature:** Billing  
**Goal:** Billing records and usage metering after package purchase (depends on Sprint 9).