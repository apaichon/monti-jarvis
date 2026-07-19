# Monti AI Call Center — Roadmap (36 core + S37–S41 product/security + S42 bug fix + S43 tenant AI & config)

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
| 14 | Tenant | Embed to Web (vanilla loader + iframe) | D | 1, 6 ✅ v1.5.0 |
| 15 | Tenant | Set Scope and KM | D | 2, 6 ✅ v1.6.0 |
| 16 | Tenant | Settings, Locale, Limit user tier, group, Quota for customer call time per day, call minute per each call | D | 13, 15 ✅ v1.7.0 |
| 17 | Tenant | Test and Preview | D | 15, 16 ✅ v1.8.0 |
| 18 | Tenant | Customer Tier | D | 16 ✅ v1.9.0 |
| 19 | Tenant | Customer Account Import, Domain Rules, Integration | E | 3, 18 ✅ v2.0.0 |
| 20 | Customer | Auth (how to integrate if solution must integrate to existing system of tenant) | E | 19 ✅ v2.1.0 |
| 21 | Customer | Select AI Workforce to Conversation must login with OTP before and time limit with quota management setting | A | 1, 5 ✅ v2.2.0 |
| 22 | Platform / Tenant | Conversation Records to Minio with optional (encrypt or not), Knowledge Gap | F | 1, 3 ✅ v2.3.0 |
| 23 | Tenant, Customer | Tickets, AI conversation ask to open ticket to human in the loop | F | 22 ✅ v2.4.0 |
| 24 | Tenant | Customer Review AI Satisfaction after conversation, Tenant view statistics | F | 22, 23 ✅ v2.5.0 |
| 25 | Tenant | Dashboard : Call Center Statistics, Call Quota Usage | F | 22 (ClickHouse) ✅ v2.6.0 · [FEAT-0027](../01-features/FEAT-0027-tenant-call-center-statistics.md) |
| 26 | Tenant | Monitoring : System Performance | F | 25 ✅ v2.7.0 |
| **27** | **Customer / Integrator** | **Mobile Call API and SDK for inbound voice integration** | **G** | **1, 20** ✅ v2.8.0 |
| 28 | Platform | Audit Log | G | 3 ✅ v2.9.0 · [FEAT-0030](../01-features/FEAT-0030-cross-tenant-audit-log.md) |
| 29 | Platform | Monitoring : System Performance | G | 28 ✅ v2.10.0 · [FEAT-0031](../01-features/FEAT-0031-platform-system-performance-monitoring.md) |
| 30 | Platform | Dashboard: Overall Call Center Statistics and by Tenants | G | 29 (ClickHouse) ✅ v2.11.0 · [FEAT-0032](../01-features/FEAT-0032-platform-call-center-statistics.md) |
| **31** | **Platform** | **Monitoring: Billing, Quota Usages, AI Infra Cost Usage** | **G** | **30** ✅ v2.12.0 · [FEAT-0033](../01-features/FEAT-0033-platform-billing-quota-ai-cost-usage.md) |
| **32** | **Tuning** | **gRPC switch mode, Cache on Prod** | **H** | **25+** ✅ v2.13.0 · [SPRINT-032](../03-sprints/SPRINT-032.md) |
| **33** | **Tuning** | **Partition, Index, Hardening** | **H** | **32** · planned · TASK-0144 UAT carry-over |
| 34 | Infra | Design Large Scale Control multiple tenant servers, Auto Scale with k8s | I | 33 |
| 35 | Infra | Canary Deployment, A/B Testing launch feature to tenant selected | I | 34 |
| 36 | Infra | Backup Restore Archive, Full,select range,Incremental, by admin platform , by tenant | I | 34 |
| **37** | **Tenant / Integrator** | **Embed SDKs: Vue · React · Svelte · Web Component** | **D+** | **14** · [FEAT-0017](../01-features/FEAT-0017-embed-framework-sdks.md) ✅ v2.14.0 · [SPRINT-037](../03-sprints/SPRINT-037.md) |
| **38** | **Customer / Platform** | **Central call center brand portal** (all tenants’ brands) | **J** | **1, 5, 6, 7** · [FEAT-0018](../01-features/FEAT-0018-central-brand-call-portal.md) · backlog |
| **39** | **Tenant / Platform** | **Theme branding & color customization** | **D+** | **14, 16** · [FEAT-0035](../01-features/FEAT-0035-theme-color-customization.md) ✅ v2.15.0 · [SPRINT-039](../03-sprints/SPRINT-039.md) |
| **40** | **Tenant / Integrator** | **Outbound calling with Twilio** | **G** | **1, 20, 27** · backlog |
| **41** | **Security / Platform** | **AI call-center security hardening: encrypted localStorage, env secrets, read-only DB, tenant isolation** | **H** | **19, 20, 32, 33** · backlog |
| **42** | **Quality / Tenant** | **Bug fix: session, login menu, nav scroll/grouping, document scope** | **Q** | **3, 15, 20** · [FEAT-0036](../01-features/FEAT-0036-tenant-ux-bugfix.md) ✅ v2.16.0 · [SPRINT-042](../03-sprints/SPRINT-042.md) |
| **43** | **Tenant / Platform** | **Embed auth mode · env config groups · tenant Gemini key · system prompt · tools · skills** | **D+** | **14, 15, 16, 39** · backlog |

---

## Phase definitions

### Phase A — Customer core (1, 2, 21)

Prove inbound AI call value before billing complexity.

- **Sprint 1:** Svelte customer portal, LiveKit voice room, transcript, NATS call events, Postgres sessions, Redis 8 active state.
- **Sprint 2:** KM ingest → MinIO → embed → ClickHouse `km_embeddings`; scope enforcement; RAG in orchestrator.
- **Sprint 21:** Customer OTP-required workforce selection where configured; customer-aware call time and quota enforcement.

### Phase B — Platform foundation (3, 4, 5, 13)

Multi-tenant SaaS skeleton.

- Auth (JWT/session, RBAC: platform / tenant / customer)
- Commercial packages and platform-managed avatars
- Quota and rate limits (Redis 8 counters + Postgres entitlements)

### Phase C — Tenant commerce (6–12)

Onboarding and monetization (one chain — see [15-commerce-chain-plan.md](../02-design/15-commerce-chain-plan.md)).

- Tenant registration → KYC → payment gateway → **buy package** (method → ChillPay → status → entitlement → **MVP receipt/tax**) → **platform billing** → **receipt ops** → **tax invoice compliance**

### Phase D — Tenant go-live (14–18)

- Web embed widget (vanilla `monti-embed.js` + iframe), tenant KM/scope admin, locale/settings/limits, test sandbox, customer tiers

### Phase D+ — Integrator embed SDKs (37)

- First-class packages for host apps: **Vue 3**, **React**, **Svelte**, and a **Web Component** (`<monti-embed>`) on top of S14 public resolve + embed surface
- Shared `@monti/embed-core` + per-framework wrappers; keep zero-dep script tag path
- Feature: [FEAT-0017](../01-features/FEAT-0017-embed-framework-sdks.md) · Depends on Sprint 14 (shipped v1.5.0)
- **Sprint 37:** Embed Framework SDKs — **shipped v2.14.0** · [SPRINT-037](../03-sprints/SPRINT-037.md)

### Phase E — Customer identity (19–20)

- Optional customer accounts for history and tier benefits

### Shipped SPRINT-021 — Authenticated workforce selection and quota limits

- Status: **shipped** · Release: **v2.2.0**
- Feature: [FEAT-0023](../01-features/FEAT-0023-authenticated-workforce-selection.md)
- Sprint: [SPRINT-021](../03-sprints/SPRINT-021.md)
- Scope: require OTP before workforce selection where tenant policy demands it, preserve optional-auth tenants, and enforce customer-aware time/quota limits.

### Phase F — Tenant operations (22–26)

- Conversation records, tickets, QA review, ClickHouse dashboards and monitoring

### Shipped SPRINT-022 — Conversation records and knowledge gaps

- Status: **shipped** · Release: **v2.3.0**
- Feature: [FEAT-0024](../01-features/FEAT-0024-conversation-records-knowledge-gaps.md)
- Sprint: [SPRINT-022](../03-sprints/SPRINT-022.md)
- Scope: archive conversation artifacts to MinIO, support configurable archive protection, and surface knowledge-gap candidates for tenant review.

### Phase G — Mobile integration and platform operations (27–31)

- **Sprint 27:** Mobile Call API and SDK for inbound voice integration
- **Sprint 28:** Cross-tenant audit log — **shipped v2.9.0** · [SPRINT-028](../03-sprints/SPRINT-028.md)
- **Sprint 29:** Platform system performance monitoring — **shipped v2.10.0** · [SPRINT-029](../03-sprints/SPRINT-029.md)
- **Sprint 30:** Platform overall call-center statistics by tenant — **shipped v2.11.0** · [SPRINT-030](../03-sprints/SPRINT-030.md)
- **Sprint 31:** Platform billing, quota, and AI infrastructure cost usage — **shipped v2.12.0** · [SPRINT-031](../03-sprints/SPRINT-031.md)
- **Sprint 32:** Platform billing usage readiness and reconciliation — **shipped v2.13.0** · [SPRINT-032](../03-sprints/SPRINT-032.md)

### Phase H — Production tuning (32–33)

- gRPC internal APIs, Redis 8 cache strategy, ClickHouse partitioning, security hardening

Sprint 32 shipped the controlled billing-usage reconciliation harness and automated source-error coverage. Sprint 33 (partition/index/hardening) remains **planned**. Manual browser/responsive UAT carry-over in TASK-0144. gRPC and production-cache implementation remain outside shipped scope. **v2.16.0** is SPRINT-042 tenant UX bugfix.

### Phase I — Infra scale (34–36)

- Autoscale, canary deployments, backup/restore/archive

### Phase J — Central multi-brand call portal (38)

- Platform-hosted **call center hub**: search/browse **all listed tenant brands**, brand profile, language + AI employee, start chat/voice
- Complements per-tenant **embed** (S14): hub = Monti multi-brand directory; embed = tenant’s own website
- Tenant opt-in listing + brand profile fields; platform moderate/unlist
- Blueprint §5.1 Customer Portal · Feature: [FEAT-0018](../01-features/FEAT-0018-central-brand-call-portal.md)
- *Pull forward after S14–18 if hub-first distribution is priority*

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

## Shipped: SPRINT-016 — Settings, Locale, Limits — v1.7.0

**Closed 2026-07-12.** Tenant settings (locale/timezone), usage snapshot, operational call caps (daily + per-call), AI locale hint, voice caption polish.

| Task | Points | Outcome |
| --- | ---: | --- |
| TASK-0072 | 3 | `tenant_settings` + `tenant_call_limits` schema, Redis daily keys |
| TASK-0073 | 5 | Settings/usage/limits APIs + voice enforce |
| TASK-0074 | 4 | `/tenant/settings` UI |
| TASK-0075 | 3 | Locale / AI reply language wiring |
| TASK-0076 | 1 | Manual UAT checklist |

Sprint: [SPRINT-016.md](../03-sprints/SPRINT-016.md) · Feature: [FEAT-0016](../01-features/FEAT-0016-tenant-settings-locale-limits.md) · Spec: [19-tenant-settings-limits-spec.md](../02-design/19-tenant-settings-limits-spec.md) · UAT: [SPRINT-016-manual.md](../06-manual-tests/SPRINT-016-manual.md)

### ⚠ Production launch gate (carry forward)

**Before production launch to end customers** — after integrating **tenant customer-user authentication** (S19–20) — **must ensure rate limit and quota management work** under real multi-user traffic:

1. Package quotas (S13) — monthly minutes, concurrent, KM, features  
2. API rate limits (S13) — chat / voice / KM per minute  
3. Operational call caps (S16) — daily + per-call minutes  
4. Tenant isolation (customer of A ≠ quota of B)  
5. Production env flags for quota/rate-limit fail mode  

Do **not** open customer production traffic until this gate is signed off (DevOps + Tester).

## Shipped: SPRINT-017 — Test and Preview — v1.8.0

**Closed 2026-07-12.** Tenant preview desk (embed-like avatar UI), package-charged chat/voice, scenarios, greeting-first voice + language picker, connecting status UX.

| Task | Points | Outcome |
| --- | ---: | --- |
| TASK-0077 | 3 | Preview session `source` + schema |
| TASK-0078 | 5 | Preview chat/voice APIs (package quotas apply) |
| TASK-0079 | 4 | `/tenant/preview` embed-like UI |
| TASK-0080 | 3 | Scenarios, embed link, lang/voice UX |
| TASK-0081 | 1 | Manual UAT |

Sprint: [SPRINT-017.md](../03-sprints/SPRINT-017.md) · Feature: [FEAT-0019](../01-features/FEAT-0019-tenant-test-preview.md) · Spec: [20-tenant-test-preview-spec.md](../02-design/20-tenant-test-preview-spec.md) · UAT: [SPRINT-017-manual.md](../06-manual-tests/SPRINT-017-manual.md) · Screens: [screenshots/s17](../../screenshots/s17/README.md)

### Production launch gate (still open)

Before **customer production** after tenant **customer-user auth** (S19–20): verify **rate limit + package quota** under multi-user load.

## Shipped: SPRINT-018 — Customer Tier — v1.9.0

**Closed 2026-07-12.** Tenant tier catalog + groups, REST CRUD, `/tenant/tiers` UI, preview `tier_id` locale/cap overrides, settings link.

| Task | Points | Outcome |
| --- | ---: | --- |
| TASK-0082 | 3 | `customer_tiers` + `customer_groups` schema |
| TASK-0083 | 5 | Tiers/groups REST APIs |
| TASK-0084 | 4 | `/tenant/tiers` UI |
| TASK-0085 | 3 | Preview tier_id + settings link |
| TASK-0086 | 1 | Manual UAT |

Sprint: [SPRINT-018.md](../03-sprints/SPRINT-018.md) · Feature: [FEAT-0020](../01-features/FEAT-0020-customer-tier.md) · Spec: [21-customer-tier-spec.md](../02-design/21-customer-tier-spec.md) · UAT: [SPRINT-018-manual.md](../06-manual-tests/SPRINT-018-manual.md)

### Production launch gate (still open)

Before **customer production** after tenant **customer-user auth** (S19–20): verify rate limit + quota **with tier overrides**.

## Shipped: SPRINT-019 — Customer Account Import, Domain Rules, Integration — v2.0.0

**Closed 2026-07-13.** Tenant customer directory, CSV dry-run/commit import, idempotent integration identity, domain defaults, and `/tenant/customers` UI.

| Task | Points | Outcome |
| --- | ---: | --- |
| TASK-0087 | 3 | Customer, import-job, domain-rule, and group-membership schema |
| TASK-0088 | 5 | Tenant customer, CSV import, and domain-rule APIs |
| TASK-0089 | 4 | Tenant `/tenant/customers` management and import UI |
| TASK-0090 | 3 | Tier/group binding and idempotent integration contracts |
| TASK-0091 | 1 | Automated smoke coverage and signed two-tenant UAT |

Two-tenant UAT passed. Customer authentication remains SPRINT-020, and production customer traffic remains blocked until auth plus quota/rate-limit isolation are signed off under multi-user load.

Sprint: [SPRINT-019.md](../03-sprints/SPRINT-019.md) · Feature: [FEAT-0021](../01-features/FEAT-0021-customer-account-import.md) · Spec: [22-customer-account-import-spec.md](../02-design/22-customer-account-import-spec.md)

## Shipped: SPRINT-020 — Customer Authentication and Domain Enforcement — v2.1.0

**Closed 2026-07-13.** Customer email OTP authentication, tenant auth settings, customer sessions, tenant-context customer portal, and authenticated chat/call tenant routing.

| Task | Points | Outcome |
| --- | ---: | --- |
| TASK-0092 | 3 | Customer OTP identity/session schema and auth settings |
| TASK-0093 | 5 | Customer OTP request/verify, session, claim, and profile APIs |
| TASK-0094 | 3 | Tenant customer-auth configuration UI |
| TASK-0095 | 4 | Customer login/account UX and authenticated context wiring |
| TASK-0096 | 1 | Authenticated tenant smoke, manual checklist, and production gate evidence |

Browser OTP/account smoke passed on the Libra Tech tenant, and automated Go/customer-web/tenant-web release gates passed. Before broad production customer traffic, re-run the documented multi-session quota/rate-limit checklist against the target deployment.

Sprint: [SPRINT-020.md](../03-sprints/SPRINT-020.md) · Feature: [FEAT-0022](../01-features/FEAT-0022-customer-auth.md) · Spec: [23-customer-auth-spec.md](../02-design/23-customer-auth-spec.md) · UAT: [SPRINT-020-manual.md](../06-manual-tests/SPRINT-020-manual.md)

## Shipped sprint: SPRINT-023

**Status:** completed · **Release:** v2.4.0 · **Commitment:** 16 points

Sprint: [SPRINT-023.md](../03-sprints/SPRINT-023.md) · Feature: [FEAT-0025](../01-features/FEAT-0025-tickets-human-escalation.md)

## Shipped sprint: SPRINT-024

**Status:** completed · **Release:** v2.5.0 · **Commitment:** 16 points

Sprint: [SPRINT-024.md](../03-sprints/SPRINT-024.md) · Feature: [FEAT-0026](../01-features/FEAT-0026-customer-satisfaction-statistics.md)

## Shipped sprint: SPRINT-025

**Status:** completed · **Release:** v2.6.0 · **Commitment:** 16 points

Sprint: [SPRINT-025.md](../03-sprints/SPRINT-025.md) · Feature: [FEAT-0027](../01-features/FEAT-0027-tenant-call-center-statistics.md)

## Shipped sprint: SPRINT-026

**Status:** completed · **Release:** v2.7.0 · **Commitment:** 16 points

Sprint: [SPRINT-026.md](../03-sprints/SPRINT-026.md) · Feature: [FEAT-0028](../01-features/FEAT-0028-tenant-system-performance-monitoring.md)

## Shipped sprint: SPRINT-027 — Mobile Call API and SDK

**Closed 2026-07-16.** Customer-safe mobile call API, bounded voice transport, typed SDK core, public brand discovery, and tenant policy enforcement shipped in v2.8.0.

Build a stable mobile integration contract for starting and ending inbound AI voice calls from a mobile application without coupling integrators to the web embed surface.

| Deliverable | Notes |
| --- | --- |
| Mobile call API contract | Authenticated session creation, tenant/avatar selection, call status, transcript events, end-call, and rating endpoints with versioned schemas |
| Voice transport adapter | Mobile-safe WebSocket/session handshake, reconnect behavior, audio permission/lifecycle guidance, and bounded failure states |
| SDK package | Typed client for the selected mobile integration target with token refresh, call lifecycle, transcript callbacks, and explicit end-call control |
| Tenant policy enforcement | Apply avatar assignment, customer auth, quota, rate-limit, and tenant isolation rules to mobile sessions |
| Sample integration | Small mobile reference app, API examples, compatibility matrix, and migration guidance from web embed |

Sprint: [SPRINT-027.md](../03-sprints/SPRINT-027.md) · Feature: [FEAT-0029](../01-features/FEAT-0029-mobile-call-api-sdk.md) · Spec: [30-mobile-call-api-sdk-spec.md](../02-design/30-mobile-call-api-sdk-spec.md)

The mobile API is feature-gated for local rollout. Push delivery remains optional and reports `not_configured` until an APNs/FCM provider adapter is deployed.

## Parallel build sprint: none

**Status:** no parallel stream

## Shipped: SPRINT-037 — Embed Framework SDKs

**Platform:** Tenant / Integrator · **Feature:** Vue · React · Svelte · Web Component packages · **Depends:** 14 · **Status:** shipped **v2.14.0** · [SPRINT-037](../03-sprints/SPRINT-037.md)

| Deliverable | Notes |
| --- | --- |
| `@monti/embed-core` | Shared resolve, iframe lifecycle, open/close/destroy |
| `@monti/embed-vue` | Vue 3 component / plugin |
| `@monti/embed-react` | React component + hooks |
| `@monti/embed-svelte` | Svelte component |
| `@monti/embed-web-component` | `<monti-embed>` custom element |
| Docs + POCs | `EMBED_WEB_INTEGRATION.md` § Framework SDKs; `examples/embed-sdks` |
| Tenant UI snippets | Framework SDKs tab on `/tenant/embed` |

Feature: [FEAT-0017](../01-features/FEAT-0017-embed-framework-sdks.md) · Builds on [FEAT-0014](../01-features/FEAT-0014-embed-to-web.md) (vanilla loader remains supported)

## Backlog add: SPRINT-038 — Central Call Center Brand Portal

**Platform:** Customer / Platform · **Feature:** Multi-tenant brand directory + conversation · **Depends:** 1, 5, 6, 7 · **Status:** backlog

| Deliverable | Notes |
| --- | --- |
| Public brand directory | List/search **active + listed** tenant brands |
| Brand profile page | Logo, blurb, languages, AI workforce CTAs |
| Start chat / voice | Session under selected `tenant_id` + agent (KM/quota scoped) |
| Tenant opt-in | “List on central portal” + public brand fields |
| Platform moderate | Force-unlist; feature flags |
| Routes | e.g. `/brands`, `/brands/{slug}` → conversation pre-bound |

Feature: [FEAT-0018](../01-features/FEAT-0018-central-brand-call-portal.md) · Blueprint §5.1 · Complements [FEAT-0014](../01-features/FEAT-0014-embed-to-web.md) (per-site embed)

## Shipped: SPRINT-039 — Theme Branding & Color Customization

**Platform:** Tenant / Platform · **Feature:** Brand chrome + full color theme · **Depends:** 14, 16 · **Status:** shipped **v2.15.0** · [SPRINT-039](../03-sprints/SPRINT-039.md) · [FEAT-0035](../01-features/FEAT-0035-theme-color-customization.md) · DES-0037

| Deliverable | Notes |
| --- | --- |
| Brand identity | Editable **brand name**, **logo**, **subtitle** on caller/embed header |
| Theme presets | Light, dark, and branded palettes with safe defaults |
| Full color token editor | Primary (+ on-primary), accent, background, surfaces, text, muted, line, status colors |
| Preview and contrast | Live preview of embed chrome; contrast flags before publish |
| Scope and rollout | Apply branding+colors per tenant on customer + embed; draft/publish/reset |
| Commitment | 14 pts · TASK-0149–0152 · **shipped v2.15.0** |

## Backlog add: SPRINT-040 — Outbound Calling with Twilio

**Platform:** Tenant / Integrator · **Feature:** Provider-backed outbound AI voice calls · **Depends:** 1, 20, 27 · **Status:** backlog

| Deliverable | Notes |
| --- | --- |
| Outbound call initiation | Tenant-authorized recipient and AI workforce selection with explicit call status |
| Twilio voice adapter | Isolate Twilio setup, credentials, number configuration, and provider callbacks behind a bounded internal adapter |
| Call lifecycle | Track requested, ringing, connected, completed, failed, and retry-safe outcomes |
| Consent and enforcement | Apply recipient consent, tenant isolation, quota, rate limits, and operational call-window policies |
| Privacy and operations | Keep provider payloads bounded, support auditability, and define recording/transcript behavior before implementation |

## Backlog add: SPRINT-041 — AI Call-Center Security Hardening

**Platform:** Security / Platform · **Feature:** Defense-in-depth browser, environment, database, and tenant-isolation controls · **Depends:** 19, 20, 32, 33 · **Status:** backlog

| Deliverable | Notes |
| --- | --- |
| Encrypted browser storage | Protect web `localStorage` data with Web Crypto, minimize persisted credentials, and define key/session expiry behavior |
| Environment and secret hardening | Validate required configuration, keep secrets out of client bundles/logs, and document rotation and production injection controls |
| Read-only AI/reporting database role | Route AI call-center and reporting read paths through a dedicated least-privilege read-only user; keep writes on separate controlled roles |
| Injection-resistant data access | Require parameterized, allowlisted queries and bounded inputs; read-only credentials are an additional containment layer, not a substitute for query safety |
| Tenant database isolation | Enforce tenant-scoped authorization and database policies/RLS where applicable so a tenant can read only its own data; add cross-tenant denial tests |

## Shipped: SPRINT-042 — Bug Fix (Quality / Tenant UX)

**Platform:** Quality / Tenant · **Feature:** Fix session, first-login menu, navigation, and document scope defects · **Depends:** 3, 15, 20 · **Status:** shipped **v2.16.0** · [SPRINT-042](../03-sprints/SPRINT-042.md) · [FEAT-0036](../01-features/FEAT-0036-tenant-ux-bugfix.md) · DES-0038 · **12 pts** TASK-0154–0157  

Dedicated **bug-fix sprint** (not mixed with new product features). Prioritize production UX blockers first.

| Deliverable | Notes |
| --- | --- |
| Session expired | Clear handling when JWT/session expires: redirect to login, no silent failure, preserve `next` path, consistent toast/copy across tenant (and customer if same bug) |
| Login first time — menu missing | After first successful login, nav/menu must render without requiring a full page refresh |
| Tenant menu grouping + scroll | Group tenant nav items (e.g. Ops / Knowledge / Commerce / Settings); sidebar must scroll when items overflow without requiring click on last item to reveal rest |
| Add document scope | Fix/complete document ↔ scope assignment on tenant KM (upload or edit document can set scope; list/filter respects scope; no orphan docs outside allowlist) |

**Acceptance sketch**

1. Expire access token → next protected navigation returns to login with reason; re-login returns to intended page.  
2. Fresh browser / cleared storage → login → full tenant nav visible on first paint.  
3. Narrow viewport / long nav → scroll the sidebar; grouped sections remain usable.  
4. Tenant can attach a KM document to a scope; agent chat only retrieves in-scope docs.

## Backlog add: SPRINT-043 — Embed Auth, Config Groups & Tenant AI Extensibility

**Platform:** Tenant / Platform · **Feature:** Embed auth toggle, lean env config, per-tenant Gemini + prompts/tools/skills · **Depends:** 14, 15, 16, 39 · **Status:** backlog  

| Deliverable | Notes |
| --- | --- |
| Embed mode auth | Per-tenant (or embed config) flag `auth: true \| false` — when true, embed/caller requires customer auth (OTP/session) before workforce/chat; when false, keep public embed path |
| Manage configuration groups | Split env/config into groups: **core infra only** in primary env (Postgres, ClickHouse, Redis, LiveKit + minimal app bind); other parameters (AI pricing, audit spool, email, feature flags, etc.) in named groups or secondary config so operators do not mix secrets with app knobs |
| Tenant Gemini API key | Active tenant can store **their own** Gemini key (encrypted at rest); runtime uses tenant key when set, else platform default; never expose raw key after save |
| Tenant custom system prompt | Tenant-editable system prompt (or per-agent override) applied to chat/voice within safety bounds (length, no secret exfil instructions) |
| Tenant call tools | Tenant can enable/configure call-time **tools** (function calling) for the AI workforce — allowlist of tool defs, enable/disable, scoped to tenant |
| Tenant custom skills | Tenant-defined **skills** packages (prompt + tool bundles + optional KM hints) assignable to agents; CRUD in tenant admin |

**Acceptance sketch**

1. Embed with `auth=true` blocks chat/voice until customer authenticated; `auth=false` matches current public embed.  
2. Documented config groups: infra keys live in core env; non-infra keys load from grouped sources without breaking `make restart`.  
3. Tenant saves Gemini key → subsequent AI calls for that tenant use it; platform admin cannot read plaintext.  
4. Custom system prompt appears in orchestrator for that tenant’s agents.  
5. At least one tool + one skill can be registered and invoked under tenant isolation tests.

**Out (unless pulled in):** Full marketplace of third-party skills; multi-provider LLM switcher beyond Gemini; replacing platform-wide Gemini entirely for all tenants.
