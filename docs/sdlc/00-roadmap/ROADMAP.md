# Monti AI Call Center — Roadmap (36 core + S37–S53 commercial/tenant tracks + S44 generative AI hold + S45 residual + S47 Langfuse backlog)

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
| **41** | **Security / Platform** | **AI call-center security hardening: encrypted localStorage, env secrets, read-only DB, tenant isolation** | **H** | **19, 20, 32, 33** · [FEAT-0041](../01-features/FEAT-0041-ai-call-center-security-hardening.md) · [SPRINT-041](../03-sprints/SPRINT-041.md) · in_progress |
| **42** | **Quality / Tenant** | **Bug fix: session, login menu, nav scroll/grouping, document scope** | **Q** | **3, 15, 20** · [FEAT-0036](../01-features/FEAT-0036-tenant-ux-bugfix.md) ✅ v2.16.0 · [SPRINT-042](../03-sprints/SPRINT-042.md) |
| **43** | **Tenant / Platform** | **Embed auth mode · env config groups · tenant Gemini key · system prompt · tools · skills** | **D+** | **14, 15, 16, 39** · [FEAT-0037](../01-features/FEAT-0037-tenant-ai-config-extensibility.md) · [SPRINT-043](../03-sprints/SPRINT-043.md) ✅ v2.17.0 |
| **44** | **Customer / Tenant** | **Customer generative AI (Claude · Codex · Antigravity · Grok CLI) → HTML/image/canvas/link/report/doc** | **K** | **1, 20, 21, 43** · [SPRINT-044](../03-sprints/SPRINT-044.md) · backlog — ON HOLD |
| **45** | **Platform / Tenant / Mobile** | **AiaaS for mass-market packages: ฿500 · ฿1,000 · ฿1,500 · ฿2,000 with differentiated quotas and corrected usage/statistics** | **L** | **13, 16, 25, 27, 30, 31, 43** · [FEAT-0039](../01-features/FEAT-0039-aiaas-packages-usage-reconciliation.md) · [SPRINT-045](../03-sprints/SPRINT-045.md) · completed 13/13; residual usage ledger/UAT |
| **46** | **Platform / Tenant / Growth** | **Tenant referral affiliate program with configurable bonus-quota rewards and referral usage tracking** | **M** | **9, 10, 13, 31, 45** · [SPRINT-046](../03-sprints/SPRINT-046.md) · ✅ v2.20.0 · manual UAT deferred |
| **47** | **Platform / AI Operations** | **Langfuse real-time LLM observability and evaluation across chat, voice, mobile, RAG, tools, and generative jobs** | **N** | **21, 25, 27, 31, 43, 44** · backlog |
| **48** | **Customer / Growth / Tenant** | **Product web for marketing, advertising, lead capture, demos, tenant registration, and package conversion** | **O** | **4, 6, 9, 17, 20, 31, 39, 46** · [FEAT-0040](../01-features/FEAT-0040-product-web-growth.md) · [SPRINT-048](../03-sprints/SPRINT-048.md) · ✅ v2.21.0 |
| **49** | **Platform / DevOps** | **Harvest-course shared-server deployment of customer, platform-admin, tenant, and product web surfaces** | **P** | **41, 48** · [SPRINT-049](../03-sprints/SPRINT-049.md) · backlog — ON HOLD (operator rollout pending) |
| **50** | **Platform Admin / Finance** | **Admin promotional package grant: set active plan + issue tax invoice for a tenant** | **P** | **4, 9, 10, 11, 12, 13** · ✅ v2.23.0 · [FEAT-0042](../01-features/FEAT-0042-admin-promotion-package-grant.md) · [SPRINT-050](../03-sprints/SPRINT-050.md) · [DES-0046](../02-design/46-admin-promotion-package-grant-spec.md) |
| **51** | **Platform / Tenant / Finance** | **Shared Cloud and Dedicated VM commercial plans: calculator, usage/quota controls, billing scheduler, receipts, and tax invoices** | **P** | **9, 10, 12, 13, 25, 31, 45, 48, 50** · planned · [FEAT-0044](../01-features/FEAT-0044-commercial-plans-billing-operations.md) · [SPRINT-051](../03-sprints/SPRINT-051.md) · [DES-0048](../02-design/48-commercial-plans-billing-operations-spec.md) |
| **52** | **Tenant / Platform** | **Tenant self-service avatar create/library: unlimited drafts; only active avatars capped by package `max_ai_employees`** | **D+** | **5, 13, 15, 16, 45, 50** · ✅ v2.24.0 · [FEAT-0043](../01-features/FEAT-0043-tenant-avatar-create-active-cap.md) · [SPRINT-052](../03-sprints/SPRINT-052.md) · [DES-0047](../02-design/47-tenant-avatar-create-active-cap-spec.md) |
| **53** | **Tenant / Customer** | **Tenant settings: auto-register customer when email + OTP is entered in conversation; show app/tag version on UI** | **D+** | **16, 19, 20, 21, 52** · planned · design **approved** · [FEAT-0045](../01-features/FEAT-0045-conversation-auto-register-app-version.md) · [SPRINT-053](../03-sprints/SPRINT-053.md) · [DES-0048](../02-design/48-conversation-auto-register-app-version-spec.md) |

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

**Platform:** Security / Platform · **Feature:** Defense-in-depth browser, environment, database, and tenant-isolation controls · **Depends:** 19, 20, 32, 33 · **Status:** in_progress · [FEAT-0041](../01-features/FEAT-0041-ai-call-center-security-hardening.md) · [SPRINT-041](../03-sprints/SPRINT-041.md)

| Deliverable | Notes |
| --- | --- |
| Encrypted browser storage | Protect web `localStorage` data with Web Crypto, minimize persisted credentials, and define key/session expiry behavior |
| Environment and secret hardening | Validate required configuration, keep secrets out of client bundles/logs, and document rotation and production injection controls |
| Read-only AI/reporting database role | Route AI call-center and reporting read paths through a dedicated least-privilege read-only user; keep writes on separate controlled roles |
| Injection-resistant data access | Require parameterized, allowlisted queries and bounded inputs; read-only credentials are an additional containment layer, not a substitute for query safety |
| Tenant database isolation | Enforce tenant-scoped authorization and database policies/RLS where applicable so a tenant can read only its own data; add cross-tenant denial tests |

Sprint 41 is planned for **14 points** based on the last three recorded
closed-slice velocities (12, 13, 17). Implementation is gated on the
`review_pending` design pack and security-owner approval. See
[DES-0044](../02-design/44-ai-call-center-security-hardening-spec.md), the
[manual UAT plan](../06-manual-tests/SPRINT-041-security-manual.md), and the
[SPRINT-041 commitment](../03-sprints/SPRINT-041.md).

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

## Planned: SPRINT-043 — Embed Auth, Config Groups & Tenant AI Extensibility

**Platform:** Tenant / Platform · **Feature:** Embed auth toggle, lean env config, per-tenant Gemini + prompts/tools/skills · **Depends:** 14, 15, 16, 39 · **Status:** planned · [SPRINT-043](../03-sprints/SPRINT-043.md)

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

## Backlog add: SPRINT-044 — Customer Generative AI (Multi-Provider CLI & Artifact Outputs)

**Platform:** Customer / Tenant · **Feature:** Customers generate artifacts via Claude, Codex, Antigravity, Grok CLI using API key or subscription login · **Depends:** 1, 20, 21, 43 · **Status:** backlog — ON HOLD pending security review · [FEAT-0038](../01-features/FEAT-0038-customer-generative-ai.md)

First-class **customer** generative workspace (not voice call-center only): authenticate a provider, run generation jobs, and receive structured outputs stored under the tenant/customer boundary.

### Providers / runtimes

| Provider | Role |
| --- | --- |
| **Claude** (API / CLI) | Text + long-form generation, code, HTML, reports |
| **Codex** (API / CLI) | Code-first generation and tool-using agent tasks |
| **Antigravity** | Agent/runtime adapter (bounded server-side invoke; no raw secrets to browser) |
| **Grok CLI** | xAI Grok CLI/API path for generation jobs |

### Credential modes

| Mode | Notes |
| --- | --- |
| **API key** | Customer or tenant stores encrypted key; server-side use only; never echo plaintext after save |
| **Login / subscription** | OAuth or provider session/subscription token where the product supports it; refresh + expiry UX; revoke from settings |

Prefer **tenant-managed** keys for B2B call-center embeds; allow **customer-owned** keys when customer auth is required (align with S20/S21). Platform default keys optional and quota-metered.

### Output artifact types

| Output | Notes |
| --- | --- |
| **HTML template** | Stored snippet or full page template; preview sandbox; download |
| **Image** | Generated image asset → MinIO under tenant/customer prefix; gallery link |
| **Canvas** | Structured canvas/JSON or image export for design surfaces |
| **Link** | Shareable public or signed URL to artifact / result page |
| **Report** | PDF/Markdown report package (transcript-adjacent or free-form brief) |
| **Doc** | Document export (Markdown/DOCX/PDF as available); versioned object |

### Deliverables

| Deliverable | Notes |
| --- | --- |
| Provider adapter layer | Server-side adapters for Claude / Codex / Antigravity / Grok CLI; unified job API (`create` → `status` → `result`) |
| Credential vault UX | Customer (and/or tenant admin) settings: set API key **or** connect subscription login; test connection; revoke |
| Generate job UI | Customer portal surface: pick provider, prompt, output type(s), run job, show progress/errors |
| Artifact store | Persist outputs to Postgres metadata + MinIO objects; tenant isolation; optional retention |
| Safety & quota | Rate limits, max payload size, content policy hooks, usage events for billing/AI cost (ties S31/S43) |
| Audit | Who ran what provider/job; no secret leakage in logs |

### Acceptance sketch

1. Customer (or tenant-configured path) can save a **Claude** or **Grok** API key encrypted and run a generation that returns at least one of: HTML template, image, report, doc.  
2. At least one provider supports **subscription/login** connect (or documented stub + API key fallback if OAuth not available).  
3. **Codex** and **Antigravity** adapters appear in provider list with bounded invoke and failure codes when not configured.  
4. Outputs for **HTML, image, canvas, link, report, doc** have explicit types in API/UI; files land in tenant-scoped storage.  
5. Tenant A cannot read Tenant B jobs/artifacts/keys; keys never returned in full after save.  
6. Usage/quota path records generation events without provider raw bodies in audit logs.

### Phase K — Customer generative workspace (44)

- Multi-provider generative jobs for authenticated customers  
- Artifact outputs beyond live voice/chat transcript  
- Complements S43 (tenant Gemini for call agents) without replacing inbound call AI  

**Out (unless pulled in):** Training custom models; unrestricted shell on customer devices; free unlimited generation without quota; replacing Gemini voice pipeline for calls.

## Completed: SPRINT-045 — AiaaS for Mass-Market Packages and Usage Reconciliation

**Platform:** Platform / Tenant / Mobile · **Feature:** Simple monthly AiaaS packages for mass-market tenants with differentiated capacity · **Depends:** 13, 16, 25, 27, 30, 31, 43 · **Status:** completed 13/13; completion release cut/tag pending · [FEAT-0039](../01-features/FEAT-0039-aiaas-packages-usage-reconciliation.md) · [DES-0042](../02-design/42-aiaas-packages-usage-reconciliation-spec.md)

Offer a small, understandable package ladder in Thai baht. The following is the
roadmap baseline for product and technical estimation; final commercial values
must be approved before package catalog release.

The four rows are initialization defaults only. After seeding, platform admins
can change package names, prices, status, and quota rules through the existing
package-management surface. Existing tenant entitlement snapshots are not
rewritten by catalog edits; a changed package takes effect only through an
explicit reassignment/upgrade.

| Monthly price | AI avatars | KM documents | Storage | Concurrent calls | Mobile call minutes |
| ---: | ---: | ---: | ---: | ---: | ---: |
| **฿500** | 1 | 100 | 5 GB | 1 | 100 min |
| **฿1,000** | 3 | 300 | 20 GB | 2 | 300 min |
| **฿1,500** | 5 | 750 | 50 GB | 5 | 750 min |
| **฿2,000** | 10 | 1,500 | 100 GB | 10 | 1,500 min |

### Requirement change

- Replace the current generic package presentation with named AiaaS mass
  packages whose price, quota rules, billing period, and included features are
  stored in the package/entitlement authorities rather than hard-coded in UI.
- Add storage and mobile usage as first-class quota dimensions. Preserve the
  existing monthly call-minute, KM-document, AI-avatar, and concurrent-call
  dimensions, and define whether mobile minutes are a separate allowance or a
  shared pool before implementation.
- Make every quota check and usage response identify its dimension, period,
  unit, limit, consumed value, remaining value, and source. A rejected request
  must not increment usage; a released concurrent-call slot must not create
  negative or stale usage.
- Correct usage tracking for web and mobile call paths, KM ingest/deletion,
  avatar assignment/revocation, MinIO storage upload/delete, and concurrent
  voice lifecycle (including disconnect, timeout, retry, and failed-start
  paths).
- Align tenant statistics, platform statistics, billing usage, and package
  usage cards with the same definitions. Clearly distinguish historical
  activity from current enforcement counters and show mobile activity
  separately while allowing reconciliation to the package allowance.
- Add an upgrade/downgrade-safe entitlement snapshot so historical usage keeps
  its package/rate context and a changed package does not rewrite past facts.

### Deliverables

| Deliverable | Notes |
| --- | --- |
| Package catalog and entitlement model | Seed the four price points in THB; expose quota dimensions and units through the API; preserve tenant-specific entitlement snapshots. |
| Quota enforcement | Add storage and mobile dimensions; normalize web/mobile call-minute and concurrency checks; use Redis DB 4 with the `monti_jarvis:` prefix. |
| Usage ledger/projection | Idempotent usage events for calls, mobile calls, KM, avatars, and MinIO bytes; define correction/replay behavior without double counting. |
| Tenant and platform statistics | Report limit, used, remaining, and utilization by dimension; reconcile ClickHouse historical facts with Postgres/Redis authorities and label unavailable/stale data. |
| Billing and package UI | Show the four plans, included quota, current consumption, remaining quota, and upgrade/downgrade impact in baht. |
| Mobile contract | Apply the selected package to mobile call API/SDK sessions and return stable quota/rate-limit errors plus usage metadata. |
| Migration and verification | Backfill or explicitly mark legacy usage, test two-tenant isolation, package changes, retries, deletes, disconnects, date boundaries, and concurrent load. |

**Related design hold:** [DES-0041 — Roadmap 45 Codex CLI skills and downloadable artifacts](../02-design/41-roadmap-45-codex-cli-skills-artifacts.md)
is separate from this package/usage sprint and remains **ON HOLD**. Sprint 45
does not enable customer or tenant CLI/generative execution.

### Acceptance sketch

1. A tenant can compare and purchase the ฿500/฿1,000/฿1,500/฿2,000 plans;
   each plan returns the exact avatar, KM, storage, concurrency, and mobile
   allowance shown in the catalog.
2. The same entitlement is enforced for tenant web, embed, and mobile paths;
   over-limit requests fail with a dimension-specific response and do not
   consume additional quota.
3. Storage usage changes only after successful MinIO writes/deletes, KM usage
   follows document lifecycle, avatar usage follows active assignments, and
   concurrent usage is released on every terminal call path.
4. Tenant and platform dashboards agree on controlled fixtures for used,
   remaining, utilization, mobile minutes, and historical-vs-current periods;
   unavailable sources are not rendered as zero.
5. Replayed or duplicated usage events produce one logical usage result, and
   changing packages preserves historical usage facts and their entitlement
   context.
6. Automated and manual tests cover two tenants, all four tiers, web and
   mobile calls, KM/storage/avatar mutations, failed starts, disconnects,
   retries, quota exhaustion, and concurrent multi-user load.

**Out (unless pulled in):** Overage billing, automatic package upgrade,
enterprise custom pricing, cross-tenant quota pooling, and changing the core
Gemini voice pipeline.

## Completed: SPRINT-046 — Tenant Referral Affiliate Program and Bonus Quota — v2.20.0

**Platform:** Platform / Tenant / Growth · **Feature:** Tenant referral/affiliate
program that rewards qualified referrals with additional quota allowance ·
**Depends:** 9, 10, 13, 31, 45 · **Status:** ✅ shipped v2.20.0; manual UAT deferred

Allow an active tenant to invite another business to Monti. A referral becomes
qualified only after the referred tenant completes onboarding/KYC, is activated,
and completes its first paid package order. The referring tenant receives a
configurable bonus entitlement, increasing the quota available to its own
tenant/customer traffic without changing the purchased package price.

### Referral and reward baseline

| Area | Roadmap requirement |
| --- | --- |
| Referral identity | Tenant-specific referral code and link; attribution is captured at signup and cannot be changed after qualification. |
| Qualified referral | Referred tenant is new, passes required onboarding/KYC, becomes active, and has a paid order that is not refunded or voided during the qualification window. |
| Referrer reward | Configurable bonus quota, initially intended as additional monthly call minutes, mobile call minutes, KM documents, and storage; any concurrent-call increase requires capacity approval. |
| Referred-tenant reward | Optional one-time onboarding quota bonus or package promotion, separately configured from the referrer reward. |
| Affiliate status | `clicked` → `attributed` → `pending` → `qualified` → `granted`, with `reversed` and `rejected` outcomes. |
| Settlement | Keep quota rewards and any future cash commission as separate ledger types; no automatic cash payout is in the initial scope. |

The initial reward values are product-configurable rather than hard-coded. The
technical design must support both fixed bonuses and percentage bonuses with
per-dimension caps, expiry, and a maximum number of rewarded referrals per
tenant and billing period. Bonus quota must be represented separately from the
base package entitlement so the tenant can see what was purchased, earned, used,
and what will expire.

### Requirement change

- Add referral attribution and qualification records under the
  `callcenter` schema with tenant-scoped authorization and an immutable audit
  trail for status changes.
- Add a bonus-entitlement ledger that can grant, consume, expire, reverse, and
  reconcile quota bonuses for each S45 dimension: AI avatars, KM documents,
  storage, concurrent calls, monthly call minutes, and mobile call minutes.
- Increase a tenant's available quota through an explicit bonus layer; never
  mutate the purchased package limit or rewrite historical usage facts.
- Make quota APIs and tenant/platform statistics show `base_limit`,
  `bonus_granted`, `bonus_used`, `bonus_remaining`, `total_limit`, and expiry
  where applicable. Existing usage counters must continue to report actual
  consumption, not reward grants.
- Track referral conversions, qualified revenue, granted quota value, consumed
  bonus quota, expiry, reversal, and fraud/rejection outcomes in platform
  reporting without exposing another tenant's private customer or billing data.
- Ensure mobile, embed, KM, avatar, MinIO storage, and concurrent-call paths use
  the same entitlement resolution as the purchased AiaaS package and consume
  bonus quota deterministically according to the documented order.

### Deliverables

| Deliverable | Notes |
| --- | --- |
| Tenant referral UX | Referral code/link, invite status, qualified referrals, earned quota, expiry, and conversion summary. |
| Platform affiliate administration | Configure campaign dates, qualification window, reward dimensions/values, caps, expiry, approval/reversal, and fraud review. |
| Attribution and qualification service | Idempotent signup/order attribution; block self-referral, duplicate attribution, circular referrals, and refunded-order rewards. |
| Bonus quota ledger | Append-only grant/consume/expire/reverse records with deterministic idempotency keys and tenant isolation. |
| Quota enforcement integration | Resolve base plus valid bonus entitlements for web, embed, mobile, KM, avatars, storage, and concurrency using Redis DB 4 and the `monti_jarvis:` prefix. |
| Statistics and billing usage | Add referral funnel and bonus-quota metrics to platform views; add base-versus-bonus breakdowns to tenant usage and S45 package reporting. |
| Audit and verification | Record actor, reason, source order, campaign, and timestamps; test qualification, reversal, expiry, retries, abuse controls, and two-tenant isolation. |

### Acceptance sketch

1. A tenant can generate a referral link and see its lifecycle without seeing
   the referred tenant's private records.
2. A referral is rewarded once, only after the configured qualification event;
   retries and duplicate callbacks do not grant quota twice.
3. The referrer sees the bonus entitlement separately from its purchased S45
   package, and the combined available quota is enforced consistently on web,
   embed, KM, storage, avatar, concurrent-call, and mobile paths.
4. Bonus consumption, expiry, reversal, and remaining values appear in tenant
   and platform statistics; quota usage is never mistaken for referral revenue
   or quota grants.
5. Refunded/voided orders, self-referrals, duplicate referrals, expired
   campaigns, and fraud-rejected referrals grant no usable bonus quota.
6. Controlled tests prove tenant isolation, idempotent ledger behavior,
   package upgrade/downgrade compatibility, mobile usage reconciliation, and
   accurate base-versus-bonus reporting.

**Out:** Multi-level/downline commissions, public affiliate
marketplace, automatic cash payouts, cross-tenant quota pooling, and referral
rewards that bypass payment, KYC, rate limits, or tenant isolation.

## Backlog add: SPRINT-047 — Langfuse Real-Time LLM Observability and Evaluation

**Platform:** Platform / AI Operations · **Feature:** Real-time traces, metrics,
feedback, and evaluation for every supported LLM interaction · **Depends:** 21,
25, 27, 31, 43, 44 · **Status:** backlog

Integrate Langfuse as the operational observability and evaluation layer for
Monti's AI workloads. The integration must make model behavior measurable in
near real time without putting prompts, transcripts, customer identity, or
provider secrets into telemetry by default.

### Requirement change

- Instrument text chat, voice relay, mobile calls, RAG retrieval, tool calls,
  tenant skills, and S44 generative jobs with trace/session/generation/span
  relationships that preserve tenant and request correlation.
- Capture model/provider, model version, prompt version, latency, token or audio
  usage when available, estimated cost state, retries, errors, fallback path,
  tool name, retrieval count, and quota/usage event references as structured
  metadata.
- Support real-time operational views for request volume, latency, error rate,
  time-to-first-token, token/audio usage, cost coverage, fallback frequency,
  RAG retrieval behavior, and quota impact by tenant, channel, avatar, model,
  and provider.
- Add evaluation pipelines for answer relevance, groundedness/citation use,
  safety/policy adherence, tool-call correctness, language/locale quality,
  voice completion quality, and artifact output validity. Support online sample
  evaluation plus replay/offline evaluation against approved datasets.
- Allow tenant-safe feedback and evaluation scores without exposing one
  tenant's traces or datasets to another tenant. Platform operators receive
  aggregate cross-tenant views with bounded drill-down and redaction.
- Version prompts, model configuration, evaluation rubrics, datasets, and score
  definitions so results remain comparable after runtime changes.
- Keep Langfuse delivery asynchronous, sampled/configurable, and fail-open:
  Langfuse outage or timeout must not fail chat, voice, mobile calls, RAG,
  archive, quota enforcement, billing, or artifact generation.

### Telemetry data boundary

| Data | Policy |
| --- | --- |
| Tenant/request correlation | Use opaque tenant-scoped IDs; never send customer email, phone, auth token, or raw session secret. |
| Prompt/response/transcript | Redact or hash by default; raw content requires an explicit platform policy, bounded retention, and tenant consent/configuration. |
| Voice/audio | Send duration and provider usage metadata only by default; no audio payloads in Langfuse. |
| KM/RAG | Record document/scope identifiers and retrieval metrics without copying document content or embeddings. |
| Tools and skills | Record allowlisted tool/skill names, status, latency, and bounded error codes; never secrets or unrestricted arguments. |
| Provider credentials | Never send API keys, OAuth tokens, subscription tokens, or raw authorization headers. |

### Deliverables

| Deliverable | Notes |
| --- | --- |
| Langfuse adapter | Central Go integration for traces, generations, observations, scores, batching, retries, sampling, redaction, and shutdown flush. |
| Runtime instrumentation | Wire chat, Gemini text/voice, mobile voice, RAG, tools/skills, and S44 provider adapters to a common trace contract. |
| Real-time operations dashboard | Platform-admin views for latency, errors, usage/cost coverage, model/provider, channel, tenant aggregates, and degraded telemetry state. |
| Evaluation service | Configurable online evaluators plus offline replay jobs, dataset/version management, score persistence, and threshold alerts. |
| Tenant feedback path | Tenant-scoped rating/feedback and evaluation summaries connected to existing satisfaction/statistics without leaking raw customer content. |
| Privacy and retention controls | Redaction, sampling, retention, consent/configuration, RBAC, audit events, and documented Langfuse deployment/secrets. |
| Verification and runbook | Failure-mode tests, trace completeness checks, score reproducibility, cost/usage reconciliation, and operator troubleshooting guidance. |

### Acceptance sketch

1. A completed chat, voice, mobile, RAG, tool, or generative job produces a
   correlated Langfuse trace when telemetry is enabled, with model/config
   metadata and outcome status but no secrets or prohibited PII.
2. Langfuse being unavailable, slow, or rate-limited does not change the
   success/failure behavior of the customer interaction or quota/billing path.
3. Platform operators can see near-real-time latency, error, usage, cost-state,
   and fallback metrics, with tenant/channel/model/provider filters and safe
   partial-failure states.
4. An approved evaluation dataset can be replayed against a versioned prompt
   and model configuration; scores for relevance, groundedness, safety, and
   output validity are reproducible and attributable to that version.
5. Tenant feedback and evaluation results are isolated; aggregate dashboards
   reconcile with call-center statistics, AI usage facts, and quota records
   without becoming a second usage authority.
6. Tests cover redaction, tenant isolation, sampling, duplicate/retry delivery,
   provider metadata missing/unavailable, Langfuse outage, high concurrency,
   voice/mobile lifecycle completion, and prompt/model version changes.

**Out (unless pulled in):** Training or fine-tuning models, sending raw audio
or unrestricted customer transcripts to third parties, automatic model
switching based only on an evaluation score, and using Langfuse as the billing
or quota source of truth.

## Completed: SPRINT-045 — AiaaS for Mass-Market Packages and Usage Reconciliation

**Platform:** Platform / Tenant / Mobile · **Feature:** Simple monthly AiaaS packages for mass-market tenants with differentiated capacity · **Depends:** 13, 16, 25, 27, 30, 31, 43 · **Status:** completed 13/13; completion release cut/tag pending · [FEAT-0039](../01-features/FEAT-0039-aiaas-packages-usage-reconciliation.md) · [DES-0042](../02-design/42-aiaas-packages-usage-reconciliation-spec.md)

The completed Sprint 45 slice covers package initialization, dimensioned quota
enforcement, usage reconciliation, reporting, mobile/load verification, and
manual UAT. TASK-0166, TASK-0167, and TASK-0168 are completed; the remaining
release action is the final version/tag cut.

Offer a small, understandable package ladder in Thai baht. The following is the
roadmap baseline for product and technical estimation; final commercial values
must be approved before package catalog release.

The four rows are initialization defaults only. After seeding, platform admins
can change package names, prices, status, and quota rules through the existing
package-management surface. Existing tenant entitlement snapshots are not
rewritten by catalog edits; a changed package takes effect only through an
explicit reassignment/upgrade.

| Monthly price | AI avatars | KM documents | Storage | Concurrent calls | Mobile call minutes |
| ---: | ---: | ---: | ---: | ---: | ---: |
| **฿500** | 1 | 100 | 5 GB | 1 | 100 min |
| **฿1,000** | 3 | 300 | 20 GB | 2 | 300 min |
| **฿1,500** | 5 | 750 | 50 GB | 5 | 750 min |
| **฿2,000** | 10 | 1,500 | 100 GB | 10 | 1,500 min |

### Requirement change

- Replace the current generic package presentation with named AiaaS mass
  packages whose price, quota rules, billing period, and included features are
  stored in the package/entitlement authorities rather than hard-coded in UI.
- Add storage and mobile usage as first-class quota dimensions. Preserve the
  existing monthly call-minute, KM-document, AI-avatar, and concurrent-call
  dimensions, and define whether mobile minutes are a separate allowance or a
  shared pool before implementation.
- Make every quota check and usage response identify its dimension, period,
  unit, limit, consumed value, remaining value, and source. A rejected request
  must not increment usage; a released concurrent-call slot must not create
  negative or stale usage.
- Correct usage tracking for web and mobile call paths, KM ingest/deletion,
  avatar assignment/revocation, MinIO storage upload/delete, and concurrent
  voice lifecycle (including disconnect, timeout, retry, and failed-start
  paths).
- Align tenant statistics, platform statistics, billing usage, and package
  usage cards with the same definitions. Clearly distinguish historical
  activity from current enforcement counters and show mobile activity
  separately while allowing reconciliation to the package allowance.
- Add an upgrade/downgrade-safe entitlement snapshot so historical usage keeps
  its package/rate context and a changed package does not rewrite past facts.

### Deliverables

| Deliverable | Notes |
| --- | --- |
| Package catalog and entitlement model | Seed the four price points in THB; expose quota dimensions and units through the API; preserve tenant-specific entitlement snapshots. |
| Quota enforcement | Add storage and mobile dimensions; normalize web/mobile call-minute and concurrency checks; use Redis DB 4 with the `monti_jarvis:` prefix. |
| Usage ledger/projection | Idempotent usage events for calls, mobile calls, KM, avatars, and MinIO bytes; define correction/replay behavior without double counting. |
| Tenant and platform statistics | Report limit, used, remaining, and utilization by dimension; reconcile ClickHouse historical facts with Postgres/Redis authorities and label unavailable/stale data. |
| Billing and package UI | Show the four plans, included quota, current consumption, remaining quota, and upgrade/downgrade impact in baht. |
| Mobile contract | Apply the selected package to mobile call API/SDK sessions and return stable quota/rate-limit errors plus usage metadata. |
| Migration and verification | Backfill or explicitly mark legacy usage, test two-tenant isolation, package changes, retries, deletes, disconnects, date boundaries, and concurrent load. |

**Related design hold:** [DES-0041 — Roadmap 45 Codex CLI skills and downloadable artifacts](../02-design/41-roadmap-45-codex-cli-skills-artifacts.md)
is separate from this package/usage sprint and remains **ON HOLD**. Sprint 45
does not enable customer or tenant CLI/generative execution.

### Acceptance sketch

1. A tenant can compare and purchase the ฿500/฿1,000/฿1,500/฿2,000 plans;
   each plan returns the exact avatar, KM, storage, concurrency, and mobile
   allowance shown in the catalog.
2. The same entitlement is enforced for tenant web, embed, and mobile paths;
   over-limit requests fail with a dimension-specific response and do not
   consume additional quota.
3. Storage usage changes only after successful MinIO writes/deletes, KM usage
   follows document lifecycle, avatar usage follows active assignments, and
   concurrent usage is released on every terminal call path.
4. Tenant and platform dashboards agree on controlled fixtures for used,
   remaining, utilization, mobile minutes, and historical-vs-current periods;
   unavailable sources are not rendered as zero.
5. Replayed or duplicated usage events produce one logical usage result, and
   changing packages preserves historical usage facts and their entitlement
   context.
6. Automated and manual tests cover two tenants, all four tiers, web and
   mobile calls, KM/storage/avatar mutations, failed starts, disconnects,
   retries, quota exhaustion, and concurrent multi-user load.

**Out (unless pulled in):** Overage billing, automatic package upgrade,
enterprise custom pricing, cross-tenant quota pooling, and changing the core
Gemini voice pipeline.

## Completed partial: SPRINT-046 — Tenant Referral Affiliate Program and Bonus Quota — v2.19.0

**Platform:** Platform / Tenant / Growth · **Feature:** Tenant referral/affiliate
program that rewards qualified referrals with additional quota allowance ·
**Depends:** 9, 10, 13, 31, 45 · **Status:** completed partial; 3/10 points
shipped in v2.19.0; manual UAT deferred — **test later**

The shipped slice delivers tenant-scoped referral attribution and qualification
with immutable/idempotent registration capture, qualification gates, and an
append-only status event trail. Sprint 45 usage-ledger and mobile verification
dependencies are complete; bonus-quota grants, manual referral UAT, and the
full affiliate UX remain future slices.

Allow an active tenant to invite another business to Monti. A referral becomes
qualified only after the referred tenant completes onboarding/KYC, is activated,
and completes its first paid package order. The referring tenant receives a
configurable bonus entitlement, increasing the quota available to its own
tenant/customer traffic without changing the purchased package price.

### Referral and reward baseline

| Area | Roadmap requirement |
| --- | --- |
| Referral identity | Tenant-specific referral code and link; attribution is captured at signup and cannot be changed after qualification. |
| Qualified referral | Referred tenant is new, passes required onboarding/KYC, becomes active, and has a paid order that is not refunded or voided during the qualification window. |
| Referrer reward | Configurable bonus quota, initially intended as additional monthly call minutes, mobile call minutes, KM documents, and storage; any concurrent-call increase requires capacity approval. |
| Referred-tenant reward | Optional one-time onboarding quota bonus or package promotion, separately configured from the referrer reward. |
| Affiliate status | `clicked` → `attributed` → `pending` → `qualified` → `granted`, with `reversed` and `rejected` outcomes. |
| Settlement | Keep quota rewards and any future cash commission as separate ledger types; no automatic cash payout is in the initial scope. |

The initial reward values are product-configurable rather than hard-coded. The
technical design must support both fixed bonuses and percentage bonuses with
per-dimension caps, expiry, and a maximum number of rewarded referrals per
tenant and billing period. Bonus quota must be represented separately from the
base package entitlement so the tenant can see what was purchased, earned, used,
and what will expire.

### Requirement change

- Add referral attribution and qualification records under the
  `callcenter` schema with tenant-scoped authorization and an immutable audit
  trail for status changes.
- Add a bonus-entitlement ledger that can grant, consume, expire, reverse, and
  reconcile quota bonuses for each S45 dimension: AI avatars, KM documents,
  storage, concurrent calls, monthly call minutes, and mobile call minutes.
- Increase a tenant's available quota through an explicit bonus layer; never
  mutate the purchased package limit or rewrite historical usage facts.
- Make quota APIs and tenant/platform statistics show `base_limit`,
  `bonus_granted`, `bonus_used`, `bonus_remaining`, `total_limit`, and expiry
  where applicable. Existing usage counters must continue to report actual
  consumption, not reward grants.
- Track referral conversions, qualified revenue, granted quota value, consumed
  bonus quota, expiry, reversal, and fraud/rejection outcomes in platform
  reporting without exposing another tenant's private customer or billing data.
- Ensure mobile, embed, KM, avatar, MinIO storage, and concurrent-call paths use
  the same entitlement resolution as the purchased AiaaS package and consume
  bonus quota deterministically according to the documented order.

### Deliverables

| Deliverable | Notes |
| --- | --- |
| Tenant referral UX | Referral code/link, invite status, qualified referrals, earned quota, expiry, and conversion summary. |
| Platform affiliate administration | Configure campaign dates, qualification window, reward dimensions/values, caps, expiry, approval/reversal, and fraud review. |
| Attribution and qualification service | Idempotent signup/order attribution; block self-referral, duplicate attribution, circular referrals, and refunded-order rewards. |
| Bonus quota ledger | Append-only grant/consume/expire/reverse records with deterministic idempotency keys and tenant isolation. |
| Quota enforcement integration | Resolve base plus valid bonus entitlements for web, embed, mobile, KM, avatars, storage, and concurrency using Redis DB 4 and the `monti_jarvis:` prefix. |
| Statistics and billing usage | Add referral funnel and bonus-quota metrics to platform views; add base-versus-bonus breakdowns to tenant usage and S45 package reporting. |
| Audit and verification | Record actor, reason, source order, campaign, and timestamps; test qualification, reversal, expiry, retries, abuse controls, and two-tenant isolation. |

### Acceptance sketch

1. A tenant can generate a referral link and see its lifecycle without seeing
   the referred tenant's private records.
2. A referral is rewarded once, only after the configured qualification event;
   retries and duplicate callbacks do not grant quota twice.
3. The referrer sees the bonus entitlement separately from its purchased S45
   package, and the combined available quota is enforced consistently on web,
   embed, KM, storage, avatar, concurrent-call, and mobile paths.
4. Bonus consumption, expiry, reversal, and remaining values appear in tenant
   and platform statistics; quota usage is never mistaken for referral revenue
   or quota grants.
5. Refunded/voided orders, self-referrals, duplicate referrals, expired
   campaigns, and fraud-rejected referrals grant no usable bonus quota.
6. Controlled tests prove tenant isolation, idempotent ledger behavior,
   package upgrade/downgrade compatibility, mobile usage reconciliation, and
   accurate base-versus-bonus reporting.

**Out (unless pulled in):** Multi-level/downline commissions, public affiliate
marketplace, automatic cash payouts, cross-tenant quota pooling, and referral
rewards that bypass payment, KYC, rate limits, or tenant isolation.

## Backlog add: SPRINT-047 — Langfuse Real-Time LLM Observability and Evaluation

**Platform:** Platform / AI Operations · **Feature:** Real-time traces, metrics,
feedback, and evaluation for every supported LLM interaction · **Depends:** 21,
25, 27, 31, 43, 44 · **Status:** backlog

Integrate Langfuse as the operational observability and evaluation layer for
Monti's AI workloads. The integration must make model behavior measurable in
near real time without putting prompts, transcripts, customer identity, or
provider secrets into telemetry by default.

### Requirement change

- Instrument text chat, voice relay, mobile calls, RAG retrieval, tool calls,
  tenant skills, and S44 generative jobs with trace/session/generation/span
  relationships that preserve tenant and request correlation.
- Capture model/provider, model version, prompt version, latency, token or audio
  usage when available, estimated cost state, retries, errors, fallback path,
  tool name, retrieval count, and quota/usage event references as structured
  metadata.
- Support real-time operational views for request volume, latency, error rate,
  time-to-first-token, token/audio usage, cost coverage, fallback frequency,
  RAG retrieval behavior, and quota impact by tenant, channel, avatar, model,
  and provider.
- Add evaluation pipelines for answer relevance, groundedness/citation use,
  safety/policy adherence, tool-call correctness, language/locale quality,
  voice completion quality, and artifact output validity. Support online sample
  evaluation plus replay/offline evaluation against approved datasets.
- Allow tenant-safe feedback and evaluation scores without exposing one
  tenant's traces or datasets to another tenant. Platform operators receive
  aggregate cross-tenant views with bounded drill-down and redaction.
- Version prompts, model configuration, evaluation rubrics, datasets, and score
  definitions so results remain comparable after runtime changes.
- Keep Langfuse delivery asynchronous, sampled/configurable, and fail-open:
  Langfuse outage or timeout must not fail chat, voice, mobile calls, RAG,
  archive, quota enforcement, billing, or artifact generation.

### Telemetry data boundary

| Data | Policy |
| --- | --- |
| Tenant/request correlation | Use opaque tenant-scoped IDs; never send customer email, phone, auth token, or raw session secret. |
| Prompt/response/transcript | Redact or hash by default; raw content requires an explicit platform policy, bounded retention, and tenant consent/configuration. |
| Voice/audio | Send duration and provider usage metadata only by default; no audio payloads in Langfuse. |
| KM/RAG | Record document/scope identifiers and retrieval metrics without copying document content or embeddings. |
| Tools and skills | Record allowlisted tool/skill names, status, latency, and bounded error codes; never secrets or unrestricted arguments. |
| Provider credentials | Never send API keys, OAuth tokens, subscription tokens, or raw authorization headers. |

### Deliverables

| Deliverable | Notes |
| --- | --- |
| Langfuse adapter | Central Go integration for traces, generations, observations, scores, batching, retries, sampling, redaction, and shutdown flush. |
| Runtime instrumentation | Wire chat, Gemini text/voice, mobile voice, RAG, tools/skills, and S44 provider adapters to a common trace contract. |
| Real-time operations dashboard | Platform-admin views for latency, errors, usage/cost coverage, model/provider, channel, tenant aggregates, and degraded telemetry state. |
| Evaluation service | Configurable online evaluators plus offline replay jobs, dataset/version management, score persistence, and threshold alerts. |
| Tenant feedback path | Tenant-scoped rating/feedback and evaluation summaries connected to existing satisfaction/statistics without leaking raw customer content. |
| Privacy and retention controls | Redaction, sampling, retention, consent/configuration, RBAC, audit events, and documented Langfuse deployment/secrets. |
| Verification and runbook | Failure-mode tests, trace completeness checks, score reproducibility, cost/usage reconciliation, and operator troubleshooting guidance. |

### Acceptance sketch

1. A completed chat, voice, mobile, RAG, tool, or generative job produces a
   correlated Langfuse trace when telemetry is enabled, with model/config
   metadata and outcome status but no secrets or prohibited PII.
2. Langfuse being unavailable, slow, or rate-limited does not change the
   success/failure behavior of the customer interaction or quota/billing path.
3. Platform operators can see near-real-time latency, error, usage, cost-state,
   and fallback metrics, with tenant/channel/model/provider filters and safe
   partial-failure states.
4. An approved evaluation dataset can be replayed against a versioned prompt
   and model configuration; scores for relevance, groundedness, safety, and
   output validity are reproducible and attributable to that version.
5. Tenant feedback and evaluation results are isolated; aggregate dashboards
   reconcile with call-center statistics, AI usage facts, and quota records
   without becoming a second usage authority.
6. Tests cover redaction, tenant isolation, sampling, duplicate/retry delivery,
   provider metadata missing/unavailable, Langfuse outage, high concurrency,
   voice/mobile lifecycle completion, and prompt/model version changes.

**Out (unless pulled in):** Training or fine-tuning models, sending raw audio
or unrestricted customer transcripts to third parties, automatic model
switching based only on an evaluation score, and using Langfuse as the billing
or quota source of truth.

## Completed: SPRINT-048 — Product Web Sales, Marketing, Demo, and Tenant Conversion — v2.21.0

**Platform:** Customer / Growth / Tenant · **Feature:** [FEAT-0040](../01-features/FEAT-0040-product-web-growth.md) ·
**Sprint:** [SPRINT-048](../03-sprints/SPRINT-048.md) · **Depends:** 4, 6, 9, 17, 20, 31, 39, 46 ·
**Status:** ✅ **shipped v2.21.0** (merged to `main`, tag `v2.21.0`)

### Product intent

Build a public-facing Monti web presence based on the supplied product-web
reference: dark Monti brand, AI workforce positioning, product and solution
explanations, resources, transparent pricing, social proof, live demo entry,
and a clear sales/contact path. The site must support both self-serve tenant
conversion and assisted sales follow-up.

Reference input: `/Users/apaichon/Downloads/monti/product-web/` (homepage,
product/solutions/resources/pricing/about compositions, and brochure artwork).
The reference visuals are inspiration; package prices, claims, customer logos,
and metrics must come from approved Monti data before publication.

### Conversion funnel

```text
Advertising / SEO / referral link
        ↓
Product web landing page with source + campaign attribution
        ├─ Try live demo → no-auth demo → request contact / register CTA
        ├─ Book a demo → lead form → sales follow-up / calendar
        ├─ Contact sales → qualified lead pipeline → assisted tenant onboarding
        └─ Pricing / Start now → tenant registration → verify + KYC
                                      → package selection → buy package
                                      → tenant workspace / receipt confirmation
```

### Public product-web surfaces

| Surface | Purpose | Primary conversion |
| --- | --- | --- |
| Home | Explain Monti's AI call-center value with hero, proof points, use cases, and strong CTAs | Live demo, book a demo |
| Product | Explain AI agents, omnichannel support, knowledge, handover, analytics, security, and integrations | Try demo, view pricing |
| Solutions | Industry/team landing pages for support, sales qualification, booking, billing, healthcare, e-commerce, and internal helpdesk | Contact sales, book demo |
| Resources | Guides, blog, webinars, videos, case studies, and downloadable brochure | Lead capture, newsletter signup |
| Pricing | Data-driven package comparison with quotas, included features, billing terms, and upgrade path | Register, choose package |
| About | Trust, company story, security/compliance posture, and contact details | Contact sales |
| Live demo | Guided no-auth AI voice/text experience with QR/share entry and a follow-up CTA | Try demo, become a lead |
| Contact / book demo | Consent-aware lead form with requested use case, company, contact channel, and source attribution | Create qualified lead |

### Lead and customer lifecycle

- Capture campaign/source/referral attribution on first visit and preserve it
  through demo, contact, registration, and package purchase where permitted.
- Store only the minimum contact data required for follow-up; record consent,
  preferred contact method, language, company, use case, and lead status.
- Support lifecycle states such as `new`, `contacted`, `demo_scheduled`,
  `demo_completed`, `qualified`, `registered`, `kyc_pending`, `package_selected`,
  `paid`, `converted`, `lost`, and `unsubscribed` with actor and timestamps.
- Give sales a bounded lead view and follow-up notes without exposing tenant
  customer conversations, credentials, payment secrets, or another tenant's
  private data.
- Connect an existing referral code/link to campaign attribution without
  bypassing tenant isolation or the Sprint 46 qualification rules.

### Redirect and purchase rules

1. Every public CTA preserves safe `utm_*`, referral, and landing-page context
   without accepting arbitrary redirect URLs.
2. `Try live demo` opens the approved demo surface and offers `Book a demo`,
   `Contact sales`, and `Register` after the experience.
3. `Book a demo` and `Contact sales` create a lead before redirecting to a
   confirmation page; sales follow-up must not depend on a browser session.
4. `Register` redirects to the existing tenant registration flow, then to email
   verification/KYC and the package catalog when onboarding is complete.
5. A tenant can select a package from the public pricing context, but payment
   and entitlement creation remain authenticated and use the existing billing
   authority; public pricing must never create an entitlement by itself.
6. After successful payment, redirect to the tenant workspace and show the
   package, receipt/tax state, next setup step, and contact/support path.

### Deliverables

| Deliverable | Notes |
| --- | --- |
| Product-web shell | Responsive Svelte site using Monti dark-blue/blue visual language, logo, accessible navigation, footer, SEO metadata, and Thai/English-ready content structure. |
| Marketing pages | Home, Product, Solutions, Resources, Pricing, About, Contact, and live-demo entry with reusable sections and approved assets. |
| Lead capture | Contact/book-demo/newsletter forms, consent and unsubscribe handling, source/referral attribution, validation, spam/rate protection, and confirmation states. |
| Demo conversion | QR/shareable live-demo entry, demo completion CTA, optional scheduling link, and lead creation without requiring signup before the demo. |
| Tenant conversion | Safe redirect from product web to tenant registration, email verification/KYC, package catalog, authenticated checkout, receipt, and tenant workspace. |
| Sales operations | Tenant-safe lead list, lifecycle/status updates, follow-up notes, assignment, export/audit controls, and notification integration using existing platform authorization. |
| Measurement | Funnel analytics for acquisition, CTA, demo, lead, registration, KYC, purchase, referral, and campaign conversion; no raw customer content or secrets in analytics. |
| Verification | Responsive/accessibility/browser checks, form abuse and consent tests, attribution persistence, redirect allowlist tests, demo-to-lead flow, two-tenant isolation, and successful purchase smoke. |

### Acceptance sketch

1. A visitor can understand Monti's value, inspect approved product/solution
   content, see data-driven package options, and reach live demo, contact, or
   registration within one clear action from every primary page.
2. A visitor can try the live demo without tenant signup; a later contact or
   registration can be linked to the originating campaign/referral when consent
   and attribution policy allow it.
3. Book-demo and contact forms validate, rate-limit, record consent/source, show
   a confirmation state, and create exactly one deduplicated lead for retries.
4. A tenant CTA redirects only to approved Monti routes, preserves safe context,
   completes registration/KYC, and reaches the authenticated package purchase
   flow without leaking data across tenants.
5. A paid package returns the correct existing entitlement, receipt/tax state,
   and tenant workspace redirect; pricing content alone never grants quota.
6. Sales can see and progress leads from first contact through demo, registration,
   and paid conversion, while tenants can only see their own private records.
7. Funnel dashboards report visits, CTA clicks, demo starts/completions, leads,
   registrations, KYC completion, package purchases, referral conversions, and
   drop-off by campaign without exposing PII beyond approved roles.

**Out (unless pulled in):** Unapproved advertising claims, a full external CRM
replacement, automatic marketing emails without consent, arbitrary third-party
redirects, public tenant data, public payment handling, and hard-coded package
prices that bypass the platform catalog.

## Shipped: SPRINT-050 — Admin Promotional Package Grant (Active Plan + Tax Invoice) ✅ v2.23.0

**Platform:** Platform Admin / Finance · **Feature:** [FEAT-0042](../01-features/FEAT-0042-admin-promotion-package-grant.md)
· **Sprint plan:** [SPRINT-050](../03-sprints/SPRINT-050.md) · **Design:** [DES-0046](../02-design/46-admin-promotion-package-grant-spec.md)
· **Depends:** 4, 9, 10, 11, 12, 13 · **Status:** shipped · **Release:** v2.23.0 · **Closed:** 2026-07-28

Sprint 50 lets platform admins **give a promotional quota package to a tenant**
from the admin web. A promotion grant is not a silent entitlement tweak: the
operator **must** select an active catalog plan and the system **must** set that
plan as the tenant's **active entitlement** and **issue a tax invoice** for the
tenant in the same grant workflow.

### Why now

Operators already have package catalog management and a basic
`POST /api/platform/tenants/{id}/entitlement` assign path, but that path does
not create commercial documents. Promotional / complimentary / sales-approved
grants still need a finance-grade trail: active plan snapshot, tax invoice, and
audit of who granted what and why.

### Sprint 50 deliverables

| Deliverable | Scope |
| --- | --- |
| Promotion grant API | `platform_admin` endpoint that assigns an active package as the tenant's active plan and issues a tax invoice atomically (or rolls back) |
| Active plan enforcement | One active entitlement per tenant; promotion always results in `status=active` for the granted plan; previous active entitlement is superseded with an auditable reason |
| Tax invoice on grant | Create/issue a tax invoice for the tenant with package line items, promotion source, seller/buyer tax fields, immutable document number |
| Admin web UX | Tenant surface (e.g. entitlement / promotion grant form) to pick package, set validity/reason, confirm active plan + tax invoice outcome |
| Audit & isolation | Platform-only write path; tenant can read own active plan and own tax invoice; no cross-tenant leakage |
| Verification | Unit/API tests for atomic grant, UI smoke, UAT checklist |

### Sprint 50 acceptance sketch

1. A platform admin can open a tenant in admin web and grant a **promotion
   package** from the active catalog without using tenant checkout or payment
   providers.
2. Completing a grant **sets the selected package as the tenant's sole active
   plan/entitlement** with a rules snapshot and optional `valid_until`.
3. Completing a grant **issues a tax invoice** for that tenant tied to the
   promotion order/source; document is visible in platform receipt/tax search
   and to the tenant billing history when applicable.
4. Partial failure is not allowed: if tax invoice issuance fails, the active
   plan must not remain half-applied (transactional grant or compensating
   rollback).
5. Non-`platform_admin` roles cannot grant promotions; tenant A cannot see
   tenant B's grant or invoice.
6. Existing paid checkout and referral bonus-quota flows remain unchanged.

**Out (unless pulled into S51):** Shared Cloud / Dedicated VM catalog redesign,
billing scheduler, proration calculator, dedicated quote provisioning, and
hard-coded tax rates.

---

## Planned: SPRINT-051 — Shared Cloud and Dedicated VM Commercial Operations

**Platform:** Platform / Tenant / Finance · **Feature:**
[FEAT-0044](../01-features/FEAT-0044-commercial-plans-billing-operations.md)
two-mode commercial catalog with price calculation, scheduled billing,
tax-invoice compliance, usage tracking, and quota management · **Sprint:**
[SPRINT-051](../03-sprints/SPRINT-051.md) · **Design:**
[DES-0048](../02-design/48-commercial-plans-billing-operations-spec.md) ·
**Depends:** 9, 10, 12, 13, 25, 31, 45, 48, 50 · **Status:** planned /
design `review_pending`

Sprint 51 turns the pricing reference into a single catalog and billing
authority for two service modes:

1. **Shared Cloud** — self-serve monthly or annual packages for startups and
   SMEs.
2. **Dedicated VM** — capacity-backed packages for larger organizations,
   subject to quote confirmation and infrastructure availability.

The values below are the proposed commercial baseline from the pricing
reference. They supersede the earlier S45 presentation values only after
catalog approval and migration. Existing tenant entitlement snapshots must not
be rewritten by a catalog edit.

### Commercial modes and package baseline

#### Shared Cloud — self-serve

| Plan | Monthly | Annual at 20% saving | AI avatars | KM/storage allowance | Concurrent voice | Voice / RAG |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Launch | ฿500 | ฿4,800/year (฿400 effective/mo) | 2 | Up to 1 GB | 1 | BYOK voice; RAG enabled |
| Starter | ฿900 | ฿8,640/year (฿720 effective/mo) | 5 | Up to 5 GB | 2 | BYOK voice; RAG enabled |
| Growth | ฿1,500 | ฿14,400/year (฿1,200 effective/mo) | 10 | Up to 10 GB | 4 | BYOK voice; RAG enabled |
| Business | ฿2,000 | ฿19,200/year (฿1,600 effective/mo) | 20 | Up to 20 GB | 6 | BYOK voice; RAG enabled |

Annual calculation: `monthly_price × 12 × (1 - annual_discount)`.
The 20% annual discount is a catalog setting, not a UI-only calculation.

#### Dedicated VM — quote and capacity confirmation

| Plan | Monthly reference | Annual estimate at 20% saving* | AI avatars | KM/storage allowance | Concurrent voice | Commercial flow |
| --- | ---: | ---: | ---: | ---: | ---: | --- |
| Dedicated Launch | ฿3,800 | ฿36,480/year | Unlimited | Up to 300 GB | 100 | Request quote |
| Dedicated Growth | ฿11,500 | ฿110,400/year | Unlimited | Up to 400 GB | 250 | Request quote |
| Dedicated Business | ฿19,200 | ฿184,320/year | Unlimited | Up to 500 GB | 500 | Request quote |
| Dedicated Enterprise | ฿32,700 | ฿313,920/year | Unlimited | Up to 600 GB | 1,000 | Request quote |

\* Dedicated annual values are planning estimates only. The final price,
discount, setup fee, capacity, SLA, and provisioning date come from the
approved quote. Dedicated voice remains BYOK and KM/RAG remains enabled.

#### Dedicated VM — no payment-gateway purchase

Dedicated cards must show **Request quotation**, never **Buy**. Selecting a
Dedicated package opens a company-information form containing legal company
name, contact name, work email, phone, optional tax/registration ID, company
size, expected concurrency, preferred region, and notes. Submitting the form:

- creates a tenant-scoped quote request;
- creates **no** payment order, receipt, tax invoice, entitlement, or
  subscription;
- enters platform capacity/sales review; and
- requires capacity confirmation, final quote, expiry, acceptance, and an
  explicit provisioning handoff before activation.

The existing checkout endpoint remains a hard guard and returns
`409 PACKAGE_REQUIRES_QUOTE` if a Dedicated package ID is submitted.

### Pricing and billing calculation

The calculator must use catalog data and return a transparent breakdown:

```text
subtotal       = base_plan_price + approved_addons + setup_fees
discount       = catalog_discount + approved_credits
taxable_amount = max(subtotal - discount, 0)
tax            = taxable_amount × configured_tax_rate
amount_due     = taxable_amount + tax
```

- Monthly plans bill at the tenant's billing anchor; annual plans bill once
  for twelve months after the annual discount is applied.
- Upgrades are prorated for the unused portion of the current period.
- Downgrades take effect at the next renewal unless finance explicitly approves
  a credit or immediate change.
- Quote-only Dedicated VM plans cannot create an entitlement until capacity,
  quote, payment terms, and provisioning approval are complete.
- Overage, auto-upgrade, setup fees, credits, and tax rates are configuration
  values. They must not be hard-coded in the pricing page or calculated from
  browser-supplied amounts.

### Billing scheduler

Implement one idempotent scheduler around the existing billing authority:

| Scheduled stage | Responsibility | Required outcome |
| --- | --- | --- |
| Renewal preview | Calculate next period, quota reset, discount, tax, and amount due | Tenant sees the upcoming charge and package/quota context |
| Invoice issue | Create one invoice for the billing period | Immutable invoice number and calculation snapshot |
| Payment attempt | Charge the approved payment method or create a finance task | Idempotent provider reference and payment status |
| Retry / dunning | Retry transient failures and notify the tenant | Bounded retry schedule, no duplicate invoice or entitlement |
| Grace / suspension | Apply configured grace period and restrict service safely | No silent quota reset or data deletion |
| Renewal settlement | Mark paid, extend entitlement, reset period counters | Subscription, invoice, payment, and quota period agree |

Scheduler requirements:

- Use the tenant billing timezone and a deterministic billing-period key.
- Re-running a job must produce the same result; provider callbacks and cron
  retries must be safe to replay.
- Keep payment state, invoice state, entitlement state, and quota state
  separate, with an auditable transition linking them.
- Dedicated provisioning must be a separate post-payment/capacity workflow and
  must not be hidden inside the monthly renewal job.

### Tax invoice and receipt roadmap

Every paid commercial mode must support receipt and tax-invoice documents with:

- seller identity, buyer legal name, billing address, tax ID, branch number,
  currency, service mode, plan, billing period, line items, discount, tax rate,
  tax amount, total, payment reference, and issue timestamp;
- immutable document numbers and states: `draft`, `issued`, `paid`, `void`,
  `reissued`, and `refunded`;
- a correction/reissue path that preserves the original document and reason;
- tenant download/history in billing and platform finance search with strict
  tenant isolation;
- a configurable tax policy, with the actual applicable rate and tax-invoice
  eligibility determined by the finance configuration rather than frontend
  constants.

### Usage tracking and quota management

Both modes use the same dimensioned usage contract. Each event identifies
`tenant_id`, `subscription_id`, `service_mode`, `dimension`, `period_start`,
`period_end`, `quantity`, `unit`, `source`, and an idempotency key.

| Dimension | Shared Cloud policy | Dedicated VM policy |
| --- | --- | --- |
| AI avatars | Hard limit by plan: 2 / 5 / 10 / 20 | Unlimited by plan; operational capacity is quote-controlled |
| KM/RAG storage | Hard storage limit: 1 / 5 / 10 / 20 GB | Hard storage limit: 300 / 400 / 500 / 600 GB |
| Concurrent voice | Hard limit: 1 / 2 / 4 / 6 | Capacity limit: 100 / 250 / 500 / 1,000 |
| Platform voice minutes | BYOK and commercially unlimited; track minutes, failures, and provider metadata | Same; track against the dedicated tenant and infrastructure |
| KM documents / retrieval | Track document lifecycle, bytes, retrievals, and failures | Same dimensions, isolated to the dedicated tenant |
| Mobile / embed calls | Track call minutes and concurrency if enabled | Same dimensions, with dedicated capacity attribution |

Quota behavior is explicit and consistent across web, embed, mobile, KM, and
voice paths:

- reject a request before incrementing a hard quota;
- release concurrent usage on disconnect, timeout, failed start, and retry;
- distinguish current enforcement counters from historical usage facts;
- show limit, used, remaining, utilization, period, and source in tenant and
  platform views;
- use Postgres `callcenter` as the billing/entitlement authority, Redis DB 4
  with the `monti_jarvis:` prefix for fast enforcement counters, and ClickHouse
  for reconciled historical analytics;
- never silently treat unavailable usage as zero; show stale/degraded state and
  keep the correction/replay path auditable.

### Tenant current plan, quota, and next bill

The billing page and tenant sidebar use one tenant-scoped current-plan response:

- package name, mode, status, and active entitlement snapshot;
- billing interval, current period, next bill date, projected amount, and
  billing state;
- quota dimensions with unit, limit/unlimited, used, remaining, utilization,
  source, and freshness; and
- latest receipt/tax-invoice links where available.

The sidebar must remove hard-coded `Enterprise` and `68% monthly allowance`.
Its compact percentage is the highest reliable utilization among finite quota
dimensions—not an invented average. When usage is unavailable it shows
`Usage unavailable`; promotion/manual plans show `No scheduled bill`, and
pre-activation Dedicated requests show `Quotation in review`.

### Sprint 51 deliverables

| Deliverable | Scope |
| --- | --- |
| Commercial catalog | Shared Cloud and Dedicated VM modes, plans, prices, discounts, tax settings, add-ons, and versioned effective dates |
| Pricing calculator | Monthly/annual totals, discount, tax, proration, credits, quote state, and line-item explanation |
| Subscription and entitlement model | Billing anchor, mode, plan version, period, renewal state, quota snapshot, and dedicated provisioning state |
| Usage ledger and projections | Idempotent calls, voice minutes, concurrency, avatars, KM/RAG, storage, mobile, and embed usage with reconciliation |
| Quota management | Shared hard ceilings, dedicated capacity ceilings, reset/grace behavior, and tenant/platform usage views |
| Billing scheduler | Preview, issue, payment, retry, dunning, suspension, renewal, and safe replay |
| Receipt and tax invoice | Immutable numbering, tax fields, download/history, void/reissue/refund workflow, and audit trail |
| Dedicated quote flow | Capacity check, quote approval, payment terms, provisioning handoff, and no-entitlement-before-approval guard |
| Current plan UX | One plan/quota/next-bill response shared by billing page and sidebar; no hard-coded package or usage |
| Verification | Two modes, monthly/annual cycles, proration, tax calculations, retry/idempotency, quota exhaustion, usage reconciliation, tenant isolation, and invoice corrections |

### Sprint 51 acceptance sketch

1. A tenant can switch between Shared Cloud and Dedicated VM and see an
   itemized monthly or annual calculation. Shared Cloud shows **Buy** and may
   open payment; Dedicated VM shows **Request quotation** and never opens
   payment.
2. Shared Cloud annual totals apply the configured 20% saving; Dedicated VM
   displays an indicative estimate and requires an approved quote before
   provisioning.
3. A paid subscription creates one entitlement snapshot with the selected mode,
   plan version, billing period, quota limits, and next billing date.
4. Renewal, retry, callback, and scheduler replay cannot create duplicate
   charges, invoices, quota resets, or entitlements.
5. Usage dashboards and quota enforcement agree for both modes across web,
   embed, mobile, voice, avatars, KM/RAG, and storage; two tenants cannot see
   each other's usage or invoices.
6. Paid orders produce receipt/tax-invoice documents with correct line items,
   tax calculation, immutable numbering, and auditable correction/reissue
   behavior.
7. Quota exhaustion blocks only the affected dimension, returns a stable error,
   and does not lose or double-count the usage event.
8. Current plan and sidebar show the same active package, billing period, next
   bill state, and quota freshness; no hard-coded plan or utilization remains.

**Out (unless separately approved):** unbounded Shared Cloud overage, automatic
plan upgrades, cross-tenant quota pooling, hidden Dedicated VM provisioning,
Dedicated payment-gateway checkout, hard-coded tax rules, and changing
historical invoices or entitlement snapshots.

---

## Shipped: SPRINT-052 — Tenant Avatar Create & Active Cap (Package Limit) ✅ v2.24.0

**Platform:** Tenant / Platform · **Feature:** [FEAT-0043](../01-features/FEAT-0043-tenant-avatar-create-active-cap.md)
· **Sprint plan:** [SPRINT-052](../03-sprints/SPRINT-052.md) · **Design:** [DES-0047](../02-design/47-tenant-avatar-create-active-cap-spec.md)
· **Depends:** 5, 13, 15, 16, 45, 50 · **Status:** shipped · **Release:** v2.24.0 · **Closed:** 2026-07-28

Today, platform admins own the avatar catalog and assign avatars to tenants,
with the package cap enforced on **active assignments**
(`max_ai_employees` / AI avatars in the plan). Tenants cannot build their own
avatar library. Sprint 52 moves avatar **creation** into the tenant console
while keeping commercial control on **how many can be active at once**.

### Policy (core rule)

| Action | Limit rule |
| --- | --- |
| **Create** avatar (draft / inactive / library) | **Not** limited by package avatar count; **not** blocked solely because storage quota is used for other KM assets. Portrait/object bytes may still count toward **storage** usage when uploaded, but storage exhaustion must not silently redefine the avatar **count** policy. |
| **Activate** avatar (set `active` for workforce / customer selection) | **Hard cap** = package entitlement `max_ai_employees` (plus any valid S46 bonus on that dimension). Activating beyond the cap returns a stable `409` / `quota_exceeded` (or equivalent) without creating a half-active state. |
| **Deactivate** / archive | Frees an active slot; draft/inactive rows remain in the tenant library. |

In short: tenants may **create many** avatars; they may **activate only up to**
the package’s total avatar limitation.

### Goals

- Tenant admin can create, edit, and list tenant-owned avatars (name, persona,
  portrait, voice binding as already supported by the catalog model).
- Separate **library** (all tenant avatars) from **active workforce** (subset
  eligible for customer selection / embed / mobile).
- Enforce active count from the effective entitlement only on activate /
  promote paths — not on create.
- Keep platform-admin catalog + assignment paths working for shared platform
  avatars; tenant-created avatars are scoped to the owning tenant.
- Show in tenant UI: active count / package limit, remaining active slots, and
  clear messaging when activation is blocked.

### Sprint 52 deliverables

| Deliverable | Scope |
| --- | --- |
| Tenant avatar CRUD | Create/update/list tenant-owned avatars; optional portrait upload reusing existing MinIO patterns |
| Active vs library states | Explicit active/inactive (or assignment) status; create defaults to inactive/draft |
| Package active cap | Activate path checks `max_ai_employees` (+ bonus if present); create path does not |
| Storage clarity | Avatar portrait bytes may accrue storage usage; package storage quota is independent of avatar **count**; do not refuse create solely for “avatar count” |
| Tenant UX | Avatar library screen: create, edit, activate/deactivate, show cap meter |
| Platform visibility | Platform admin can still list/inspect tenant avatars for support (read-only or existing ops paths) |
| Verification | Over-create allowed; over-activate blocked; deactivate frees slot; tenant isolation; workforce API only returns active |

### Sprint 52 acceptance sketch

1. A tenant admin can create more avatars than `max_ai_employees` without error,
   as long as other hard validations pass (auth, required fields).
2. Activating avatars succeeds only while
   `active_count < effective max_ai_employees` (package + valid bonus).
3. When at the active cap, activate returns a stable, user-visible error; no
   avatar becomes half-active.
4. Deactivating an active avatar frees a slot so another library avatar can be
   activated.
5. Customer/workforce/embed/mobile selection surfaces only **active** tenant
   avatars (plus any still-assigned platform catalog avatars per existing
   rules).
6. Tenant A cannot see or activate tenant B’s avatars.
7. Package storage quota (KM/bytes) does not substitute for the avatar **active**
   cap, and the avatar active cap does not rewrite storage policy.

**Out (unless pulled in):** HeyGen or third-party live avatar generation,
cross-tenant marketplace of avatars, removing platform catalog, changing S51
commercial plan matrix, or unlimited concurrent voice.

**Design note:** Prefer extending `ai_avatars` / `tenant_avatar_assignments`
(or a tenant-owned avatar table with the same active-cap semantics) rather than
a second workforce authority. Package rule remains `max_ai_employees` on the
entitlement snapshot.

---

## Planned: SPRINT-053 — Conversation Auto Customer Register (Email OTP) + App Version on UI

**Platform:** Tenant / Customer · **Feature:** [FEAT-0045](../01-features/FEAT-0045-conversation-auto-register-app-version.md)
· **Sprint plan:** [SPRINT-053](../03-sprints/SPRINT-053.md) · **Design:** [DES-0048](../02-design/48-conversation-auto-register-app-version-spec.md)
· **Depends:** 16, 19, 20, 21, 52 · **Status:** planned · **Design pack:** approved · **Worktree:** `.worktrees/SPRINT-053`

See sprint doc for deliverables, ACs, and design pack links. Summary:

1. Tenant setting `auto_register_on_conversation_otp` (default off).
2. Conversation email → OTP → auto-create/reuse customer → session bind when on.
3. App version on UI matches `VERSION` / git tag via `/api/version` and shell footers.
