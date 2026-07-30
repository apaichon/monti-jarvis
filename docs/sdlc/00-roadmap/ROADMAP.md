# Monti AI Call Center — Roadmap (36 core + S37–S67 commercial/tenant/customer UX tracks + S44 generative AI hold + S45 residual + S47 Langfuse backlog)

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
| **51** | **Platform / Tenant / Finance** | **Shared Cloud and Dedicated VM commercial plans: calculator, usage/quota controls, billing scheduler, receipts, tax invoices, and admin quote monitoring** | **P** | **9, 10, 12, 13, 25, 31, 45, 48, 50** · ✅ v2.25.0 · [FEAT-0044](../01-features/FEAT-0044-commercial-plans-billing-operations.md) · [SPRINT-051](../03-sprints/SPRINT-051.md) · [DES-0048](../02-design/48-commercial-plans-billing-operations-spec.md) |
| **52** | **Tenant / Platform** | **Tenant self-service avatar create/library: unlimited drafts; only active avatars capped by package `max_ai_employees`** | **D+** | **5, 13, 15, 16, 45, 50** · ✅ v2.24.0 · [FEAT-0043](../01-features/FEAT-0043-tenant-avatar-create-active-cap.md) · [SPRINT-052](../03-sprints/SPRINT-052.md) · [DES-0047](../02-design/47-tenant-avatar-create-active-cap-spec.md) |
| **53** | **Tenant / Customer** | **Tenant settings: auto-register customer when email + OTP is entered in conversation; show app/tag version on UI** | **D+** | **16, 19, 20, 21, 52** · ✅ **v2.26.0** · [FEAT-0046](../01-features/FEAT-0046-conversation-auto-register-app-version.md) · [SPRINT-053](../03-sprints/SPRINT-053.md) · [DES-0050](../02-design/50-conversation-auto-register-app-version-spec.md) |
| **54** | **Customer / Platform** | **Customer portal: pick tenant from list to call (no `tenant_id` query string required)** | **J** | **1, 5, 6, 20, 21, 38, 53** · ✅ **v2.27.0** · [FEAT-0045](../01-features/FEAT-0045-customer-portal-tenant-list.md) · [SPRINT-054](../03-sprints/SPRINT-054.md) · [DES-0049](../02-design/49-customer-portal-tenant-list-spec.md) |
| **55** | **Tenant** | **Call Center Statistics grouped by topic** | **F+** | **22, 25, 30** · ✅ **v2.28.0** · [FEAT-0047](../01-features/FEAT-0047-tenant-call-center-topic-statistics.md) · [SPRINT-055](../03-sprints/SPRINT-055.md) · [DES-0051](../02-design/51-tenant-call-center-topic-statistics-spec.md) · extends [FEAT-0027](../01-features/FEAT-0027-tenant-call-center-statistics.md) |
| **56** | **Customer** | **Caller desk UX revamp: v2 call rail, selected tenant card, mic/speaker device settings, avatar call grid, tenant switcher, account/footer, and larger Monti + company branding** | **A+** | **1, 5, 39, 54** · ✅ **v2.29.0** · [FEAT-0048](../01-features/FEAT-0048-caller-desk-branding-audio-devices.md) · [SPRINT-056](../03-sprints/SPRINT-056.md) · [DES-0052](../02-design/52-caller-desk-branding-audio-devices-spec.md) |
| **57** | **Customer / Product Web / Branding** | **Monti root-domain logo and social preview metadata: use new Monti logo and Open Graph/Twitter tags for Facebook/link sharing** | **O+** | **39, 48, 54, 56** · ✅ **v2.30.0** · [FEAT-0049](../01-features/FEAT-0049-monti-logo-social-preview-metadata.md) · [SPRINT-057](../03-sprints/SPRINT-057.md) · [DES-0053](../02-design/53-monti-logo-social-preview-metadata-spec.md) |
| **58** | **Customer / Tenant / Platform Admin** | **Portal UI language selector: EN, TH, and Japanese localized labels for call page, tenant portal, and admin pages** | **I18N** | **16, 20, 39, 54, 56, 57** · ✅ **v2.31.0** · [FEAT-0050](../01-features/FEAT-0050-portal-ui-language-selector.md) · [SPRINT-058](../03-sprints/SPRINT-058.md) · [DES-0054](../02-design/54-portal-ui-language-selector-spec.md) |
| **59** | **Customer / Conversation UX** | **Call conversation UX revamp: friendly desktop/mobile call interface with clearer controls, chat transcript, customer context, and light/dark mode** | **A+** | **1, 21, 39, 54, 56, 58** · ✅ **v2.32.0** · [FEAT-0051](../01-features/FEAT-0051-call-conversation-ux-revamp.md) · [SPRINT-059](../03-sprints/SPRINT-059.md) · [DES-0055](../02-design/55-call-conversation-ux-revamp-spec.md) |
| **60** | **Tenant / AI Operations / Security** | **Tenant-owned Gemini key enforcement: no production `GEMINI_API_KEY` env fallback; AI Settings key entry with live validation test** | **D+** | **41, 43, 52** · [FEAT-0052](../01-features/FEAT-0052-tenant-owned-gemini-key-enforcement.md) · [SPRINT-060](../03-sprints/SPRINT-060.md) · [DES-0056](../02-design/56-tenant-owned-gemini-key-enforcement-spec.md) · **planned** |
| **61** | **Tenant / Platform Admin / AI Operations** | **Tenant UX simplification: remove tenant system performance page from tenant portal; move Gemini status to tenant top bar** | **Q** | **26, 29, 43, 60** · [FEAT-0053](../01-features/FEAT-0053-tenant-gemini-status-top-bar.md) · [SPRINT-061](../03-sprints/SPRINT-061.md) · [DES-0057](../02-design/57-tenant-gemini-status-top-bar-spec.md) · **planned** |
| **62** | **Tenant / Growth / Quota** | **Referral code redemption: tenant can apply a referral code to add bonus quota with validation, limits, and ledger tracking** | **M+** | **13, 45, 46, 51** · [FEAT-0054](../01-features/FEAT-0054-referral-code-redemption.md) · [SPRINT-062](../03-sprints/SPRINT-062.md) · [DES-0058](../02-design/58-referral-code-redemption-spec.md) · **planned** |
| **63** | **Customer / Tenant / Quota** | **Queued concurrent-call admission: callers wait when tenant package concurrent-call limit is full, then start when another customer finishes** | **A+** | **13, 16, 21, 45, 51, 56** · backlog |
| **64** | **Infra / Platform Admin / DevOps** | **Full and incremental backup/restore for Postgres, ClickHouse, and MinIO with verified recovery runbooks** | **I+** | **2, 22, 25, 28, 29, 36, 41, 49** · backlog |
| **65** | **Tenant / Customer / KM** | **Tenant customer product catalog: upload files and render relevant products, menus, guides, packages, or business records during conversation** | **D+** | **2, 14, 15, 20, 21, 22, 39, 43, 54, 56** · backlog |
| **66** | **Tenant / Security / Back Office** | **Multi-user tenant permissions: tenant admins invite same-domain users and assign menu-level back-office access** | **E+** | **3, 6, 16, 19, 20, 28, 41, 42, 53** · backlog |
| **67** | **Customer / Tenant / Tickets** | **AI summary before call close: generate call recap, confirm unresolved items, and submit summary to ticket** | **F+** | **1, 21, 22, 23, 24, 43, 53, 55, 56** · backlog |
| **68** | **Tenant / Customer / Notifications** | **Call schedule email notifications: topic-based handover or sales links redirect customers into an auto-start prepared voice conversation** | **E+** | **1, 16, 20, 21, 23, 43, 53, 56, 65, 67** · backlog |

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
**Depends:** 9, 10, 12, 13, 25, 31, 45, 48, 50 · **Status:**
implementation complete; manual UAT/release pending

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

## Shipped: SPRINT-053 — Conversation Auto Customer Register (Email OTP) + App Version on UI ✅ v2.26.0

**Platform:** Tenant / Customer · **Feature:** Tenant setting to auto-register
customers when they enter email and complete OTP during conversation; surface
the same app/tag version across portals · **Depends:** 16, 19, 20, 21, 52 ·
**Status:** shipped · **Release:** v2.26.0 · **Closed:** 2026-07-29 · **Feature:** [FEAT-0046](../01-features/FEAT-0046-conversation-auto-register-app-version.md) · **Design:** [DES-0050](../02-design/50-conversation-auto-register-app-version-spec.md)

### 1) Tenant settings — auto-register customer via email OTP in conversation

Today customer OTP auth (S20/S21) can gate workforce selection when required.
Sprint 53 adds a **tenant-controlled auto-register path inside the live
conversation**:

- Tenant Settings exposes a toggle (working name:
  **auto-register customer on conversation email OTP**).
- When **enabled**, a customer in chat/voice can enter an **email** in the
  conversation UI; the system **sends an OTP** to that email.
- On successful OTP verify:
  - if no `customers` row exists for that tenant+email, **create** (register)
    the customer automatically;
  - bind the conversation/session to that customer identity;
  - continue the conversation under the registered customer (quota/attribution
    as existing customer-aware paths).
- When **disabled**, conversation does not offer this auto-register flow
  (existing optional/required auth policies remain unchanged unless combined
  explicitly).
- Respect existing domain allow/deny rules, rate limits, and tenant isolation
  from S19–S20.

#### Goals

- Capture identity mid-conversation without forcing full signup before first
  message when the tenant wants friction-light registration.
- One tenant setting, clear customer UX for email → OTP → continue.
- Reuse customer OTP challenge, session, and `customers` directory authorities
  — do not invent a second identity store.

#### Acceptance sketch (auto-register)

1. Tenant admin can enable/disable auto-register-on-conversation-OTP in
   Settings; default is **off** (safe).
2. With setting **on**, customer conversation UI prompts for email, requests
   OTP, and on verify creates or reuses the tenant-scoped customer and attaches
   the session.
3. With setting **off**, conversation does not auto-create customers via this
   path.
4. OTP abuse limits and domain policy still apply; wrong tenant cannot claim
   another tenant’s customer.
5. Existing “require OTP before workforce” policy continues to work and can
   coexist without double-registering.

### 2) Show app version (same as git tag / `VERSION`) on UI

Operators and tenants need to know which release is running.

- Expose a single **app version** string equal to the release tag / `VERSION`
  file (e.g. `v2.24.0` or `2.24.0` — pick one canonical format and use it
  everywhere).
- Show it on primary shells:
  - tenant console (e.g. sidebar footer or settings),
  - platform admin,
  - optionally customer portal footer (small, non-intrusive).
- Prefer build-time injection from `VERSION` (or `/api` health/version field)
  so UI always matches the tagged binary/static assets.

#### Acceptance sketch (version)

1. `VERSION` / release tag `vX.Y.Z` and UI version label match for a given
   deploy.
2. Version is visible without opening browser devtools.
3. No secrets or environment names are leaked via the version display.

### Sprint 53 deliverables

| Deliverable | Scope |
| --- | --- |
| Tenant setting | Boolean (or enum) for conversation auto-register via email OTP |
| Conversation UX | Email input → send OTP → verify → continue under customer |
| Customer store | Auto-create `customers` row when missing; reuse when present |
| Session bind | Attach verified customer to active conversation/session |
| API | Settings get/put; conversation OTP request/verify hooks |
| App version | Single source from `VERSION`/tag; show on tenant (+ admin) UI |
| Verification | Setting off/on, domain/rate limit, isolation, version match |

### Sprint 53 acceptance sketch (combined)

1. Auto-register flow works end-to-end when enabled; no auto-create when disabled.
2. OTP + domain + rate-limit gates match S20 behavior.
3. Tenant A cannot register or bind customers for tenant B.
4. UI shows app version matching the deployed `VERSION` / git tag.
5. Workforce/quota attribution uses the registered customer after verify.

**Out (unless pulled in):** SMS OTP, social login, mandatory email on every
anonymous demo path, changing S51 commercial catalog, marketing email blasts.

**Design note:** Extend S16 settings + S20 customer auth OTP rather than a new
auth product. Version display should not require a separate sprint from the
auto-register work if both stay small; ship together under S53.

---

## Shipped: SPRINT-054 — Customer Portal Tenant List to Call (No `tenant_id` Query) ✅ v2.27.0

**Platform:** Customer / Platform · **Feature:** Customer portal entry by
selecting a tenant (brand) from a list, without requiring `?tenant_id=` on the
URL · **Depends:** 1, 5, 6, 20, 21, 38, 53 · **Status:** shipped · **Release:** v2.27.0 · **Closed:** 2026-07-29 · **Feature:** [FEAT-0045](../01-features/FEAT-0045-customer-portal-tenant-list.md) · **Design:** [DES-0049](../02-design/49-customer-portal-tenant-list-spec.md) · **Sprint:** [SPRINT-054](../03-sprints/SPRINT-054.md)

### Problem today

The customer portal scopes the session with a query string:

```text
/ ?tenant_id=acme
```

Callers must know or be handed a tenant id. That blocks a clean multi-brand
entry surface and couples deep links to internal ids.

### Goal

1. **Tenant list to call to** — Customer portal shows a list of call-to
   tenants (public brand listings / active tenants eligible for inbound), then
   the caller picks one and continues to avatar selection / conversation.
2. **Remove required `tenant_id` query string** — Primary path does not depend
   on `?tenant_id=`. Tenant context is chosen in-app (and may still be
   optionally deep-linked for share/bookmark, but is not required).

### Scope

### In

- Public (or lightly gated) **tenant list** API for customers: safe fields only
  (display name, logo/brand, locale, slug/id for selection — no secrets).
- Customer portal **picker screen** before workforce/conversation when no
  tenant is selected.
- After pick: set tenant context in session/state (memory, path segment, or
  allowlisted slug route such as `/t/{slug}`) and load avatars/chat/voice for
  that tenant only.
- Deprecate **required** `tenant_id` query param; optional deep link may still
  preselect a tenant then strip or replace with cleaner routing.
- Tenant isolation: list and subsequent APIs never mix data across tenants.
- Align with S38 central brand portal intent where useful; S54 can be the
  minimal “list → call” slice without full multi-brand marketing hub.

### Out

- Full S38 multi-brand marketing portal (unless already in scope)
- Cross-tenant conversation history
- Platform admin bulk brand CMS redesign
- Changing OTP/auto-register rules beyond consuming selected tenant_id

### Deliverables

| Deliverable | Scope |
| --- | --- |
| Public tenant list API | Active/listable tenants for inbound call entry |
| Customer portal tenant picker | UI list → select → enter desk |
| Context without query string | In-app selection; optional slug/path deep link |
| Migration from `?tenant_id=` | Optional preselect; primary UX no longer requires it |
| Verification | Pick A vs B isolation; no list leak; call/chat after pick |

### Acceptance sketch

1. Opening the customer portal **without** `tenant_id` shows a tenant list (or
   empty state), not a broken desk.
2. Selecting a tenant loads that tenant’s workforce and conversation only.
3. Chat/voice APIs use the selected tenant context (header/session/path), not a
   required query string.
4. Optional deep link may preselect a tenant; primary docs and CTAs do not
   require `?tenant_id=`.
5. Tenant A’s customers/agents/KM are never visible after selecting tenant B.

**Out (unless pulled in):** S38 full brand portal polish, paid listing ranking,
or removing platform demo single-tenant defaults entirely.

**Design note:** Prefer `slug` or path-based deep links over raw `tenant_id`
query params. Reuse brand listing / public theme surfaces where they already
exist.

---

## Shipped: SPRINT-055 — Tenant Call Center Statistics Grouped by Topic

**Platform:** Tenant · **Feature:** Extend tenant call-center statistics with
**group by topic** breakdowns (conversation/call topic tags) · **Depends:** 22,
25, 30 · **Status:** shipped v2.28.0 · **Extends:** [FEAT-0027](../01-features/FEAT-0027-tenant-call-center-statistics.md)
(S25 dashboard)

### Problem

Sprint 25 shipped tenant call-center statistics and quota usage (volume,
channels, satisfaction, package context). Tenants still cannot see **which
topics** drive traffic — e.g. billing vs technical vs general — so staffing and
KM investment stay guesswork.

Customer desk already exposes coarse topics (`general`, `billing`, `technical`
and similar). Those should flow into analytics facts and tenant dashboards.

### Goal

- Tenant Call Center Statistics dashboard can **group and filter by topic**.
- Metrics (calls, minutes, chat sessions, satisfaction where available) break
  down **per topic** for a date range, tenant-scoped only.
- Topic values come from conversation/call metadata already captured (or a
  small extension to persist topic on completed records / ClickHouse facts).

### Scope

### In

- Persist or project **topic** on analytics path (Postgres completed records
  and/or ClickHouse call-center facts used by S25 APIs).
- Tenant stats API: series and/or rows **grouped by topic** (and date where
  already supported).
- Tenant dashboard UI: topic breakdown table/chart; optional topic filter.
- Empty/unknown topic bucket when topic was not set.
- Strict tenant isolation (same as S25).

### Out

- Free-form multi-label taxonomy admin product (unless a minimal allowlist is
  needed)
- Platform-wide cross-tenant topic leaderboard (S30-style) unless a thin
  follow-on
- Changing customer topic picker UX beyond ensuring topic is stored
- Replacing S25 quota usage views

### Deliverables

| Deliverable | Scope |
| --- | --- |
| Topic on facts | Write topic into stats pipeline from conversation/call end |
| API group-by-topic | Tenant statistics endpoint(s) return topic aggregates |
| Dashboard | Tenant Call Center Stats: group by topic table/chart + date range |
| Isolation | Tenant A never sees tenant B topic totals |
| Verification | Seed two topics; counts match; unknown bucket; date filter |

### Acceptance sketch

1. Completed conversations/calls with a topic appear in tenant stats **grouped
   by that topic**.
2. Tenant can open Call Center Statistics and see topic breakdown for a chosen
   date range (counts and, where already shown, minutes/satisfaction).
3. Rows with missing topic land in a stable **Unknown / unset** bucket.
4. Changing date range refreshes topic aggregates without cross-tenant leak.
5. Existing non-topic S25 summary metrics remain available and consistent.

**Out (unless pulled in):** ML auto-topic classification, multi-select topics
per call, or platform topic benchmarking.

**Design note:** Prefer extending existing S25 ClickHouse/Postgres stats
contracts with a `topic` dimension rather than a second analytics product.

---

## Shipped: SPRINT-056 — Caller Desk Branding + Mic/Speaker Settings ✅ v2.29.0

**Platform:** Customer · **Feature:** Revamp customer call desk and tenant/brand
list for the v2 call-page design: Monti hero/status, selected tenant card,
collapsible audio settings, AI avatar call grid, tenant switcher, account card,
secure footer, and microphone/speaker device selection for voice · **Depends:**
1, 5, 39, 54 · **Status:** shipped · **Release:** v2.29.0 · **Closed:**
2026-07-29 · **Feature:** [FEAT-0048](../01-features/FEAT-0048-caller-desk-branding-audio-devices.md) · **Design:** [DES-0052](../02-design/52-caller-desk-branding-audio-devices-spec.md) · **Sprint:** [SPRINT-056](../03-sprints/SPRINT-056.md)
**Mockups:** [call page](../02-design/mockups/s56-caller-desk-branding/call-page.png) ·
[tenant list](../02-design/mockups/s56-caller-desk-branding/tenant-list.png) ·
[composite](../02-design/mockups/s56-caller-desk-branding/new-call-design-composite.png)

**Call-page v2 design anchors:** (1) Monti AI Call Center hero with online
status; (2) selected tenant card with large company logo and change-tenant
action; (3) audio settings panel with microphone, speaker, refresh, device
levels, and audio test; (4) AI avatar call cards with active avatar emphasis;
(5) tenant switcher for owned tenants plus add-tenant entry; (6) signed-in
account, sign-out, encrypted-data notice, and app version footer.

### Problem today

After S54, callers pick a brand and enter the desk, but:

1. **Brand identity is small** — company logo and Monti product mark compete with
   dense controls; mockups show a large Monti hero mark plus a clear selected-
   company card with logo.
2. **Tenant list cards under-emphasize logos** — directory should lead with
   large brand marks (and Monti hero) so multi-company discovery feels like a
   branded hub, not a text list.
3. **No mic/speaker settings** — voice calls use the browser default input/
   output; callers cannot pick microphone or speaker when multiple devices
   are available (common on desktops and headsets).

### Goal

1. **Call page revamp** — Larger Monti product logo/hero; prominent **selected
   company logo** and name on the control rail; avatar call cards; tenant
   switcher; signed-in account and secure footer (per v2 mockup).
2. **Tenant list revamp** — Larger Monti hero + **large per-company logos** on
   brand cards (per mockup).
3. **Mic and speaker settings** — Before/during voice, caller can select
   input (mic) and output (speaker/headphones); persist choice for the session
   (and optionally local preference).

### Scope

### In

- Customer web visual revamp aligned to mockups (layout, logo scale, brand card).
- Use published tenant theme branding (`logo_url`, brand name) and Monti product
  mark assets; fallbacks when logo missing.
- Device enumeration via browser media APIs (`enumerateDevices`, `getUserMedia`
  permission prompt as needed).
- UI: mic select + speaker select + refresh + audio test (labels EN/TH as
  needed); apply to Gemini voice / LiveKit capture and playback paths already
  used by the desk.
- Session or `localStorage` preference for last mic/speaker device ids.
- Accessible labels; graceful fallback when device list empty or permission denied.
- Keep S54 routing (`/`, `/t/{slug}`) and Live · OK system status (non-technical).

### Out

- Changing tenant theme publish pipeline beyond consuming existing public theme.
- Native mobile OS-only audio routing (beyond browser APIs).
- Full redesign of OTP/auth flows (layout polish only if needed for logo space).
- S55 topic statistics.
- Platform admin CMS for logos (reuse S39 theme / brand profile).

### Deliverables

| Deliverable | Scope |
| --- | --- |
| Call page logo revamp | Large Monti hero + selected company logo card on desk |
| Tenant list logo revamp | Large Monti hero + large brand logos on directory cards |
| Audio settings panel | Collapsible panel with mic, speaker, refresh, levels, and test action |
| Mic settings | Select/list input devices; use for voice capture |
| Speaker settings | Select/list output devices; use for voice playback |
| Avatar call grid | Per-avatar cards with active/selected state and call CTA |
| Tenant switcher | Owned tenant cards plus add-tenant entry |
| Account/footer | Signed-in identity, sign-out, encrypted-data notice, and version |
| Preference persist | Session and/or local last-used device ids |
| Verification | Logos load per brand A/B; device switch works; denied permission UX |

### Acceptance sketch

1. Tenant list shows a large Monti mark and each company card shows a large
   brand logo (or monogram fallback), matching mockup intent.
2. Call desk shows a large Monti product logo, online status, and a clear
   selected-company logo block (not a tiny header chip only).
3. Caller can open audio settings and choose **microphone** and **speaker**
   when multiple devices exist.
4. Audio settings provide refresh and test actions without breaking the call
   page when permissions are denied or labels are unavailable.
5. Avatar cards show callable AI agents with a visible active/selected state.
6. Tenant switcher shows owned tenants and an add-tenant path without requiring
   `tenant_id` query-string use.
7. Starting a voice call uses the selected mic/speaker (or safe default).
8. Switching brand updates company logo to the newly selected tenant only.
9. Permission denied / no devices shows clear non-technical messaging (not
   raw API errors).

**Out (unless pulled in):** Advanced calibrated tone generator, noise
suppression UI, or multi-language voice device naming beyond browser labels.

**Design note:** Prefer client-side device selection wired into existing
`GeminiVoice` / LiveKit paths; do not invent a server audio device registry.
Reuse `GET /api/public/theme/{tenant_id}` and public brand `logo_url` for marks.

**Tasks:** TASK-0202 (branding), TASK-0203 (mic/speaker), TASK-0204 (UAT).  
**Worktree:** `.worktrees/SPRINT-056` · branch `feature/sprint-056-caller-desk-branding-audio`

---

## Shipped: SPRINT-057 — Monti Logo and Social Preview Metadata ✅ v2.30.0

**Platform:** Customer / Product Web / Branding · **Feature:** Replace the
root-domain/customer-facing Monti brand mark with the new logo asset and add
social sharing metadata so pasted root-domain links render the Monti image,
title, and description correctly on Facebook and other rich-preview surfaces ·
**Depends:** 39, 48, 54, 56 · **Status:** shipped · **Release:** v2.30.0 ·
**Closed:** 2026-07-30 ·
**Feature:** [FEAT-0049](../01-features/FEAT-0049-monti-logo-social-preview-metadata.md) ·
**Sprint:** [SPRINT-057](../03-sprints/SPRINT-057.md) ·
**Design:** [DES-0053](../02-design/53-monti-logo-social-preview-metadata-spec.md)

### Problem today

The customer-facing root domain and shared links do not consistently expose the
current Monti brand preview. When the URL is pasted into Facebook, chat apps, or
other social surfaces, the preview can be missing, stale, or generic. The root
experience needs to use the new Monti logo image and publish explicit Open
Graph/Twitter metadata for a polished product preview.

### Goal

1. **Use the new Monti logo** — Root-domain and customer-facing brand surfaces
   use the approved Monti AI Call Center logo image from
   `/Users/apaichon/Projects/libra/monti/images/logo.png`.
2. **Add rich preview metadata** — Root HTML includes Open Graph and Twitter
   card metadata for image, title, description, URL, site name, and content
   type.
3. **Validate link sharing** — Local and production builds expose absolute,
   crawlable social-preview image URLs that Facebook/debugger-style crawlers can
   fetch.

### Scope

### In

- Add the approved Monti logo asset to the appropriate public/static asset path
  for the root-domain app, preserving image quality and cache-safe naming.
- Update root-domain visible branding to use the new logo where product/customer
  surfaces show the Monti mark.
- Add Open Graph tags such as `og:title`, `og:description`, `og:image`,
  `og:image:alt`, `og:url`, `og:type`, and `og:site_name`.
- Add Twitter/X card tags such as `twitter:card`, `twitter:title`,
  `twitter:description`, `twitter:image`, and `twitter:image:alt`.
- Ensure metadata uses production absolute URLs for shared previews and a safe
  local fallback during development.
- Add or verify favicon/apple-touch/icon metadata where the root app already
  supports it.
- Manual validation notes for Facebook Sharing Debugger or equivalent rich-link
  crawler after deployment.

### Out

- Redesigning the full product-web landing page.
- Tenant-specific social previews or per-tenant Open Graph generation.
- Dynamic image generation for every route.
- Changing tenant company logos or uploaded brand assets.
- Paid social campaign tracking pixels or marketing attribution.

### Deliverables

| Deliverable | Scope |
| --- | --- |
| Logo asset | New Monti logo copied into the root app public/static path with cache-safe filename |
| Visible branding | Root/customer-facing Monti mark updated to use the approved logo |
| Open Graph tags | Title, description, image, image alt, URL, type, and site name metadata |
| Twitter card tags | Summary-large-image metadata aligned with the Open Graph preview |
| Preview validation | Local HTML/source check and post-deploy rich-preview debugger checklist |
| Verification | Root page renders logo; shared URL metadata resolves to accessible image and description |

### Acceptance sketch

1. Root-domain page uses the new Monti AI Call Center logo image from the
   approved asset, not the older/placeholder mark.
2. Page source includes `og:image` and related Open Graph metadata with an
   absolute production image URL suitable for Facebook link previews.
3. Page source includes description, title, site name, and Twitter card metadata
   consistent with the Monti AI Call Center brand.
4. The preview image URL returns the image with a crawler-readable content type
   and does not require authentication.
5. A local source check and a production rich-preview validation checklist are
   captured before marking the sprint complete.

---

## Shipped: SPRINT-058 — Portal UI Language Selector (EN · TH · JA) ✅ v2.31.0

**Platform:** Customer / Tenant / Platform Admin · **Feature:** UI language
selector with localized labels for English, Thai, and Japanese on the customer
call page, tenant portal, and platform admin · **Depends:** 16, 20, 39, 54, 56, 57 ·
**Status:** shipped · **Release:** v2.31.0 · **Closed:** 2026-07-30 ·
**Feature:** [FEAT-0050](../01-features/FEAT-0050-portal-ui-language-selector.md) ·
**Design:** [DES-0054](../02-design/54-portal-ui-language-selector-spec.md) ·
**Sprint:** [SPRINT-058](../03-sprints/SPRINT-058.md)

### Goal

1. Language selector on customer, tenant, and admin portals.
2. EN / TH / JA catalogs for primary chrome (nav, actions, statuses, empty/error).
3. Persist `monti_jarvis:ui_lang` in the browser; EN fallback for missing keys.
4. Keep UI language separate from S16 AI reply / workspace locale.

### Commitment (12/12)

| Task | Points | Focus |
| --- | ---: | --- |
| TASK-0208 | 3 | Shared i18n runtime + LanguageSelector |
| TASK-0209 | 3 | Customer directory + call desk labels |
| TASK-0210 | 3 | Tenant portal shell + primary pages |
| TASK-0211 | 3 | Platform admin + cross-surface UAT |

### Design pack

- Deep spec DES-0054 · workflow §130–131 · ER Sprint 58 · API Sprint 58 · UX C58/T58/A58
- Manual UAT: [SPRINT-058-portal-ui-language-selector-manual.md](../06-manual-tests/SPRINT-058-portal-ui-language-selector-manual.md)

**Worktree:** `.worktrees/SPRINT-058` · branch `feature/sprint-058-portal-ui-i18n`

---

## Shipped: SPRINT-059 — Call Conversation UX Revamp ✅ v2.32.0

**Platform:** Customer / Conversation UX · **Feature:** responsive call
conversation workspace with a prominent live avatar, clear controls, transcript,
context rails, and light/dark presentation · **Depends:** 1, 21, 39, 54, 56, 58 ·
**Status:** shipped · **Release:** v2.32.0 · **Closed:** 2026-07-30 ·
**Feature:** [FEAT-0051](../01-features/FEAT-0051-call-conversation-ux-revamp.md) ·
**Design:** [DES-0055](../02-design/55-call-conversation-ux-revamp-spec.md) ·
**Sprint:** [SPRINT-059](../03-sprints/SPRINT-059.md)

### Delivered

1. Desktop/mobile conversation shell with scroll-safe transcript and composer.
2. Large live avatar that remains visible during active calls.
3. Collapsed icon-led tenant, customer, and device context sections.
4. Icon-only light/dark switch with persisted browser preference.
5. Per-avatar default, dark, and light portrait uploads.
6. Real speaker/microphone toggles and numeric keypad composer input.

**Manual UAT:** [SPRINT-059-call-conversation-ux-manual.md](../06-manual-tests/SPRINT-059-call-conversation-ux-manual.md)

---

## Planned: SPRINT-060 — Tenant-Owned Gemini Key Enforcement

**Status:** planned · **Release target:** v2.33.0 · **Feature:** [FEAT-0052](../01-features/FEAT-0052-tenant-owned-gemini-key-enforcement.md) ·
**Design:** [DES-0056](../02-design/56-tenant-owned-gemini-key-enforcement-spec.md) · **Sprint:** [SPRINT-060](../03-sprints/SPRINT-060.md) ·
**Worktree:** `.worktrees/SPRINT-060-062` · branch `docs/sprint-060-062-plan`

**Commitment (12):** TASK-0216 validation API (4) · TASK-0217 fail-closed runtime (3) · TASK-0218 AI Settings UX (3) · TASK-0219 audit/UAT (2)

---

## Planned: SPRINT-061 — Tenant UX Simplification (Gemini Status Top Bar)

**Status:** planned · **Release target:** v2.34.0 · **Feature:** [FEAT-0053](../01-features/FEAT-0053-tenant-gemini-status-top-bar.md) ·
**Design:** [DES-0057](../02-design/57-tenant-gemini-status-top-bar-spec.md) · **Sprint:** [SPRINT-061](../03-sprints/SPRINT-061.md) ·
**Depends:** S60 · **Worktree:** `.worktrees/SPRINT-060-062`

**Commitment (12):** TASK-0220 status API (3) · TASK-0221 remove performance nav (3) · TASK-0222 top-bar chip (3) · TASK-0223 platform UAT (3)

---

## Planned: SPRINT-062 — Referral Code Redemption Adds Bonus Quota

**Status:** planned · **Release target:** v2.35.0 · **Feature:** [FEAT-0054](../01-features/FEAT-0054-referral-code-redemption.md) ·
**Design:** [DES-0058](../02-design/58-referral-code-redemption-spec.md) · **Sprint:** [SPRINT-062](../03-sprints/SPRINT-062.md) ·
**Worktree:** `.worktrees/SPRINT-060-062`

**Commitment (12):** TASK-0224 validate/apply API (4) · TASK-0225 bonus ledger/usage (3) · TASK-0226 tenant UX (3) · TASK-0227 platform reverse/UAT (2)


## Detail (planned as SPRINT-060): Tenant-Owned Gemini Key Enforcement

**Platform:** Tenant / AI Operations / Security · **Feature:** Stop using a
shared `GEMINI_API_KEY` environment fallback for tenant AI runtime in production.
Tenant admins must enter their own Gemini API key in AI Settings, test that the
key can connect to Gemini, and only validated keys can power tenant AI calls ·
**Depends:** 41, 43, 52 · **Status:** backlog

### Problem today

Sprint 43 added encrypted tenant Gemini keys, but the runtime can still fall
back to a platform `GEMINI_API_KEY` from environment when a tenant has no key.
That makes production cost attribution and tenant isolation weaker than the
commercial model needs. Tenants need a clear AI Settings flow to provide,
validate, rotate, and remove their own Gemini key without exposing plaintext
secrets to the browser.

### Goal

1. **Tenant-owned key required** — Production chat/voice AI runtime uses the
   tenant's validated Gemini key and does not use a shared env `GEMINI_API_KEY`
   fallback for tenant calls.
2. **AI Settings key management** — Tenant admins can save, replace, delete, and
   see masked metadata for their Gemini key in AI Settings.
3. **Test Gemini connection** — Tenant admins can click a test action that
   verifies the saved or proposed key can connect to Gemini before it is marked
   valid for runtime use.

### Scope

### In

- Tenant AI Settings UI for Gemini key entry, replacement, deletion, masked
  last-four display, validation status, and last-tested timestamp.
- Backend validation endpoint/action that tests the key against Gemini with a
  bounded, non-bill-heavy request and returns non-secret success/failure details.
- Store encrypted key material only server-side; persist metadata such as
  `last4`, `status`, `last_validated_at`, and validation error class.
- Runtime resolution change: production tenant AI calls fail closed with a clear
  configuration error when no validated tenant Gemini key exists.
- Explicit dev/test behavior: any env-key fallback must be opt-in and reported
  as non-production only.
- Audit events for key create/replace/delete/test, without storing plaintext.
- Readiness or posture signal that reports whether shared env Gemini fallback is
  disabled for production tenant runtime.

### Out

- Exposing Gemini keys to browser JavaScript or platform-admin read paths.
- Platform-funded shared Gemini pool for tenant production traffic.
- Multi-provider key marketplace or non-Gemini provider support.
- Tenant billing changes, AI usage pricing changes, or quota plan redesign.
- Translating or rewriting tenant KM content.

### Deliverables

| Deliverable | Scope |
| --- | --- |
| AI Settings UI | Key entry, replace/delete, masked metadata, test action, validation state |
| Key validation | Server-side Gemini connectivity test with bounded request and safe errors |
| Runtime enforcement | Production AI calls require validated tenant Gemini key; no env fallback |
| Secret storage | Encrypted-at-rest key storage with no plaintext read API |
| Ops posture | Readiness/security signal for env fallback disabled in production |
| Verification | Valid key succeeds; invalid key fails; missing key blocks tenant AI call; no plaintext leaks |

### Acceptance sketch

1. Tenant admin can enter a Gemini key in AI Settings and run **Test connection**.
2. A valid key is marked usable with masked metadata and `last_validated_at`.
3. An invalid key shows a safe, non-secret failure reason and is not used for
   tenant AI runtime.
4. In production mode, tenant chat/voice AI calls never use `GEMINI_API_KEY`
   from environment when the tenant key is missing or invalid.
5. Key create/replace/delete/test actions are audited without plaintext secret
   values.

---

## Detail (planned as SPRINT-061): Tenant UX Simplification: Gemini Status in Top Bar

**Platform:** Tenant / Platform Admin / AI Operations · **Feature:** Tenant
admins should not need a dedicated System Performance page for infrastructure
details. Remove the tenant-facing system performance route from normal tenant
navigation and move the Gemini connectivity/status signal into the tenant top
bar, while keeping deeper health diagnostics in platform-admin operations ·
**Depends:** 26, 29, 43, 59 · **Status:** backlog

### Problem today

Sprint 26 exposed tenant-safe system performance monitoring, but that surface is
too operational for most tenant admins. The tenant portal should stay focused on
business workflows: avatars, KM, customers, billing, usage, and AI settings.
For tenant operators, the important live signal is whether Gemini is configured
and reachable; deeper dependency health belongs to platform administration.

### Goal

1. **Remove tenant performance page from tenant UX** — Hide or retire the tenant
   System Performance navigation/page from normal tenant-admin workflows.
2. **Gemini status in top bar** — Show a compact Gemini status indicator in the
   tenant portal top bar so tenants can immediately see configured, valid,
   degraded, or missing-key state.
3. **Keep platform diagnostics** — Preserve platform-admin system performance
   monitoring and operational probes for support teams.

### Scope

### In

- Remove tenant System Performance from tenant nav and route discovery; preserve
  backend compatibility only if needed for platform/support tooling.
- Add tenant top-bar Gemini status indicator with concise states such as
  `Gemini ready`, `Gemini key missing`, `Gemini validation failed`, and
  `Gemini degraded`.
- Link the top-bar Gemini status to Tenant AI Settings when action is needed.
- Reuse S43/S59 key metadata and validation status rather than adding a second
  status model.
- Keep platform-admin performance dashboards and dependency probes available for
  operators.
- Non-technical tenant copy; no raw provider errors or internal hostnames.

### Out

- Removing platform-admin system performance monitoring.
- Showing Redis, Postgres, MinIO, NATS, ClickHouse, or LiveKit internals to
  tenant admins.
- New Gemini key-management behavior beyond what S59 owns.
- Changing customer call-page status beyond what tenant configuration requires.

### Deliverables

| Deliverable | Scope |
| --- | --- |
| Tenant nav cleanup | Remove System Performance from tenant portal navigation |
| Top-bar status | Compact Gemini readiness/status indicator in tenant layout |
| AI Settings link | Status action routes tenant admin to AI Settings when remediation is needed |
| API reuse | Consume S43/S59 Gemini key/status metadata with safe tenant-scoped response |
| Platform preservation | Platform-admin diagnostics remain available for support/ops |
| Verification | Tenant no longer sees performance page in nav; Gemini status updates across ready/missing/failed states |

### Acceptance sketch

1. Tenant admin no longer sees a dedicated System Performance page in normal
   tenant navigation.
2. Tenant top bar shows Gemini status for the current tenant without exposing
   infrastructure internals.
3. Missing or invalid Gemini key status links directly to AI Settings.
4. Platform admin can still access full system performance diagnostics.
5. Existing S26/S29 operational APIs are either preserved for compatible support
   use or deliberately deprecated in the sprint design.

---

## Detail (planned as SPRINT-062): Referral Code Redemption Adds Bonus Quota

**Platform:** Tenant / Growth / Quota · **Feature:** Let a tenant enter and
apply a referral code from another tenant or campaign to receive eligible bonus
quota, with validation, abuse controls, expiry rules, and quota-ledger tracking ·
**Depends:** 13, 45, 46, 51 · **Status:** backlog

### Problem today

Sprint 46 established referral attribution and bonus-quota rewards, but the
tenant experience is still referral-link-first. Operators also need a simple
manual redemption flow: enter a referral code in the tenant portal, validate it,
and apply eligible bonus quota to the current tenant without support manually
editing entitlements.

### Goal

1. **Apply referral code** — Tenant admin can enter a referral code and see
   whether it is valid, expired, already used, self-owned, or ineligible.
2. **Grant bonus quota** — A valid code creates a bounded bonus-quota grant for
   the tenant without mutating the base paid package entitlement.
3. **Track redemption lifecycle** — Tenant and platform views show applied code,
   bonus quota, expiry, usage, and reversal state.

### Scope

### In

- Tenant portal referral-code input and validation flow.
- Server-side validation for code existence, campaign status, expiry, tenant
  eligibility, one-use/per-period limits, self-referral, duplicate redemption,
  and fraud/abuse rules.
- Bonus quota grant through the existing quota/bonus ledger layer, separate from
  base package limits.
- Supported quota dimensions aligned to S45/S46, such as call minutes, mobile
  minutes, KM documents, storage, or active-avatar bonus where configured.
- Tenant usage UI shows base quota plus applied referral bonus and expiry.
- Platform admin can inspect, revoke, or reverse redemption grants.
- Idempotent apply behavior so retries never grant quota twice.

### Out

- Cash payout, affiliate commission settlement, or accounting payouts.
- Cross-tenant quota pooling beyond the specific redeemed bonus grant.
- Manual SQL entitlement edits as the product path.
- Unlimited concurrent-call increases without capacity approval.
- Applying referral codes to invoices, tax documents, or historical usage.

### Deliverables

| Deliverable | Scope |
| --- | --- |
| Tenant redemption UI | Referral code input, validation state, apply action, bonus summary |
| Validation API | Tenant-scoped code validation/apply endpoint with clear safe errors |
| Bonus quota ledger | Idempotent grant records with dimension, amount, expiry, source code |
| Usage display | Tenant usage shows base quota, referral bonus, consumed bonus, expiry |
| Platform oversight | Admin view/action for redemption audit, revoke, and reversal |
| Verification | Valid apply, duplicate retry, expired code, self-referral, cross-tenant isolation |

### Acceptance sketch

1. Tenant admin can apply a valid referral code and immediately see the granted
   bonus quota in usage/quota views.
2. The base paid package entitlement remains unchanged; bonus quota is tracked
   as a separate grant source.
3. Duplicate, expired, self-owned, unknown, or ineligible codes do not grant
   quota and show non-technical messages.
4. Applying the same code twice, including retry after timeout, cannot create
   duplicate quota grants.
5. Platform admin can audit and reverse a referral-code quota grant.

---

## Backlog: SPRINT-063 — Queued Concurrent-Call Admission

**Platform:** Customer / Tenant / Quota · **Feature:** Enforce each tenant
package's total concurrent-call limit with a tenant-scoped waiting queue. When a
caller starts a voice call and the tenant is already at its concurrent-call
quota, the caller waits until another customer finishes or the queue timeout
expires · **Depends:** 13, 16, 21, 45, 51, 56 · **Status:** backlog

### Problem today

Sprint 13 enforces `max_concurrent_calls` by rejecting calls over the package
limit. That protects capacity, but it creates a poor caller experience during
brief traffic spikes: the caller must retry manually even if another customer
finishes seconds later. Tenants need package-based concurrency protection with a
controlled wait queue instead of immediate rejection.

### Goal

1. **Queue over-limit callers** — If a tenant has no available concurrent-call
   slot, the start-call flow places the caller in a tenant-scoped queue.
2. **Start when slot opens** — When another customer finishes, disconnects, or
   times out, the next queued caller is admitted and the voice call starts.
3. **Respect package limits** — The active-call count must never exceed the
   tenant package's concurrent-call limit, including bonus quota where approved.

### Scope

### In

- Tenant-scoped concurrent-call queue keyed by tenant and call channel.
- Admission controller that checks package concurrency, reserves a slot, or
  enqueues the caller with position and estimated wait metadata.
- Server-side worker/goroutine path to promote queued callers when active calls
  release slots; implementation must remain safe across server restarts and
  multiple app instances.
- Queue timeout, cancellation, browser disconnect cleanup, and idempotent retry
  behavior.
- Customer UI states for waiting, position, timeout, cancel, and automatic call
  start when admitted.
- Redis DB 4 keys under `monti_jarvis:` for active slots, queue entries, TTLs,
  and promotion locks.
- Tenant/admin visibility for active calls, queued callers, package limit, and
  recent queue timeouts.
- Metrics/audit events for queued, admitted, cancelled, timed out, and rejected
  states.

### Out

- Exceeding the tenant package concurrent-call limit.
- Infinite queues or unbounded goroutine creation per waiting caller.
- Prioritized paid queue tiers, callback scheduling, or human-agent routing.
- Cross-tenant queue sharing or borrowing unused capacity from another tenant.
- Replacing monthly-minute, daily-limit, or rate-limit enforcement.

### Deliverables

| Deliverable | Scope |
| --- | --- |
| Queue data model | Redis-backed tenant queue, active-slot lease, TTL, and promotion lock |
| Admission API/WS | Start-call response supports admitted vs queued with queue position |
| Promotion worker | Bounded goroutine/worker promotes queued callers when slots release |
| Customer UI | Waiting screen, position, cancel, timeout, and auto-start when admitted |
| Tenant visibility | Active/queued count and package concurrent limit in usage/status UI |
| Verification | Two tenants, over-limit queue, disconnect cleanup, timeout, restart, no over-admission |

### Acceptance sketch

1. When active calls are below the tenant package limit, a caller starts
   immediately and consumes one concurrent slot.
2. When active calls equal the tenant package limit, the next caller is queued
   instead of receiving immediate `quota_exceeded`.
3. When an active call ends, the first eligible queued caller is admitted and
   active calls never exceed the package limit.
4. If a queued caller cancels, closes the browser, or times out, the queue entry
   is cleaned up and the next caller position is updated.
5. Multiple tenants have independent queues and cannot consume each other's
   concurrent-call capacity.

---

## Backlog: SPRINT-064 — Full and Incremental Backup/Restore

**Platform:** Infra / Platform Admin / DevOps · **Feature:** Provide scheduled
full and incremental backups plus verified restore workflows for Monti
Postgres, ClickHouse, and MinIO data, including operator runbooks, manifests,
retention, and audit evidence · **Depends:** 2, 22, 25, 28, 29, 36, 41, 49 ·
**Status:** backlog

### Problem today

Monti stores tenant configuration, calls, transcripts, analytics, embeddings,
and KM assets across Postgres, ClickHouse, and MinIO. Production readiness
requires repeatable recovery, not only best-effort snapshots. Operators need
full and incremental backup jobs, off-host retention, and a restore path that is
tested against staging/local before any production recovery is attempted.

### Goal

1. **Back up all Monti data stores** — Capture Postgres, ClickHouse, and MinIO
   data with full backups and incremental changes.
2. **Restore with evidence** — Restore into staging/local and verify schema,
   row counts, objects, checksums, and application-level consistency.
3. **Operate safely** — Enforce encryption, retention, audit logs, explicit
   production confirmation, and clear recovery runbooks.

### Scope

### In

- Postgres backup for database `monti_jarvis`, schema `callcenter`, including
  full logical/physical backup and incremental/PITR or WAL strategy where
  selected.
- ClickHouse backup for database `monti_jarvis`, including full snapshots and
  incremental partition/object backups where supported.
- MinIO backup for bucket `monti-jarvis`, including full bucket backup and
  incremental object sync/versioned backup for prefixes `calls/` and `km/`.
- Backup manifests with timestamp, environment, app version, schema/database
  names, bucket/prefix list, object counts, sizes, checksums, and source backup
  cursor/checkpoint.
- Restore runner for staging/local dry runs and controlled production restore
  with explicit operator confirmation and maintenance-window evidence.
- Encrypted backups, off-host storage target, retention policy, pruning, and
  failed-backup alerting.
- Platform/admin operations visibility for last backup, status, size, retention
  window, RPO/RTO estimate, and last restore verification.
- Recovery runbooks and automated verification scripts that produce deployment
  or readiness evidence.

### Out

- Backing up unrelated HarvestMax databases, schemas, buckets, or app storage.
- Restoring production without explicit operator confirmation and rollback plan.
- Per-customer or tenant self-service restore.
- Data warehouse redesign, archive product UX, or cold-storage analytics.
- Schema migrations unrelated to backup/restore correctness.

### Deliverables

| Deliverable | Scope |
| --- | --- |
| Backup runner | Scheduled full backups for Postgres, ClickHouse, and MinIO with encryption and off-host target |
| Incremental strategy | WAL/PITR or selected Postgres incremental path, ClickHouse incremental snapshots, MinIO object delta sync/versioning |
| Restore runner | Staging/local restore first, optional gated production restore, cleanup and rollback hooks |
| Manifest/checksum | Per-backup metadata, checksums, cursor/checkpoint, retention, and verification output |
| Ops dashboard/runbook | Platform/admin backup status, RPO/RTO, last restore test, operator recovery procedure |
| Verification | Automated dry-run restore, consistency checks, corrupt backup handling, and DEP/readiness evidence |

### Acceptance sketch

1. A full backup captures Postgres `monti_jarvis.callcenter`, ClickHouse
   database `monti_jarvis`, and MinIO bucket `monti-jarvis` prefixes `calls/`
   and `km/`.
2. Incremental backup captures only changes since the last valid checkpoint or
   snapshot and records that checkpoint in the manifest.
3. Restore to staging/local reconstructs databases and bucket objects, then
   passes schema, count, checksum, and application consistency checks.
4. Production restore requires explicit operator confirmation, records audit
   evidence, and links the resulting DEP/readiness note.
5. Corrupt, missing, partial, or checksum-mismatched backups fail safely without
   silently leaving a partial restore marked successful.

---

## Backlog: SPRINT-065 — Tenant Customer Product Catalog

**Platform:** Tenant / Customer / KM · **Feature:** Let tenants manage a
customer-facing product/catalog library with downloadable files and structured
metadata, then render the most relevant catalog items inside chat or voice
conversation when the customer's request matches them · **Depends:** 2, 14, 15,
20, 21, 22, 39, 43, 54, 56 · **Status:** backlog

### Problem today

Monti can answer from tenant KM, but tenants also need a product-style catalog
surface that is designed for customer conversations. A restaurant may need to
show a food menu, a travel agency may show package guides, an insurer may show
coverage plans, a lender may show loan packages, and an HR/business tenant may
show attendance or employee-service documents. The AI should not only answer in
text; it should surface the relevant item, preview key details, and offer a
download when the customer asks about it.

### Goal

1. **Manage catalog assets** — Tenant admins can upload and organize
   customer-facing catalog files with metadata, categories, language, tags,
   eligibility, and publish state.
2. **Match customer intent** — Conversation retrieval finds relevant catalog
   items based on the customer's request, tenant scope, avatar role, language,
   and active publish rules.
3. **Render and download** — Customer conversation can show product cards,
   menus, packages, guides, or business records with safe file download links.

### Scope

### In

- Tenant portal catalog management for upload, edit, publish/unpublish,
  archive, versioning, and delete.
- File support for common customer-facing assets such as PDF, image, CSV/XLSX,
  DOCX, and structured JSON/CSV catalog rows where approved.
- Catalog item types for food menus, travel guides/packages, insurance
  packages, loan packages, employee attendance or HR service records, and
  configurable business-specific catalog categories.
- MinIO storage under tenant-scoped catalog prefixes with metadata in Postgres
  and optional searchable embeddings in ClickHouse.
- RAG/relevance pipeline that ranks catalog items and returns cited item IDs,
  snippets, thumbnails/previews where available, and download eligibility.
- Customer conversation rendering for related catalog cards, menu/package
  summaries, file previews, and download actions.
- Tenant controls for which avatars can use each catalog collection and whether
  files are public, authenticated-customer-only, or restricted by customer tier.
- Multilingual metadata and matching, aligned with EN/TH/Japanese localization
  and tenant avatar language settings.
- Audit events for upload, publish, download, and AI-rendered catalog
  recommendations.

### Out

- Full ecommerce checkout, cart, payment, invoice, or inventory reservation.
- Editing customer business systems of record from the conversation.
- Public search engine indexing of private tenant catalog files.
- Cross-tenant catalog sharing unless a later marketplace feature approves it.
- Unlimited file retention or storage beyond tenant package limits.

### Deliverables

| Deliverable | Scope |
| --- | --- |
| Catalog data model | Tenant-scoped collections, items, file versions, metadata, publish state, access policy |
| Upload/download API | Secure file upload, signed download links, thumbnail/preview metadata, storage quota checks |
| Tenant catalog UI | Manage food menus, travel guides, insurance/loan packages, HR records, and custom categories |
| Retrieval integration | Catalog indexing, relevance ranking, citations, language matching, avatar collection scope |
| Customer rendering | Conversation cards, file previews, related-product panels, and download action states |
| Verification | Tenant isolation, file ACLs, relevance quality, stale version handling, package quota enforcement |

### Acceptance sketch

1. Tenant admin can upload, tag, publish, and version customer-facing catalog
   assets without exposing them to other tenants.
2. When a customer asks a related question, the conversation surfaces relevant
   catalog items such as a menu, travel package, insurance plan, loan package,
   or business record with a concise summary and citation.
3. Download links are tenant-scoped, permission-checked, time-limited where
   needed, and record an audit event.
4. Unpublished, archived, expired, or restricted catalog items are not rendered
   to ineligible customers or avatars.
5. Catalog storage, indexing, and retrieval respect tenant package limits,
   language settings, and existing KM isolation rules.

---

## Backlog: SPRINT-066 — Multi-User Tenant Permissions

**Platform:** Tenant / Security / Back Office · **Feature:** Let tenant admins
invite same-domain staff users into the tenant portal and assign menu-level
back-office permissions so multiple users can manage operations without sharing
one owner account · **Depends:** 3, 6, 16, 19, 20, 28, 41, 42, 53 ·
**Status:** backlog

### Problem today

Tenant administration is owner-centric. As tenants grow, support, billing,
operations, content, and technical staff need controlled access to specific
tenant back-office menus. Sharing a single tenant admin account weakens audit
trails and makes it hard to limit who can change AI settings, quota, billing,
catalog content, customer records, or reports.

### Goal

1. **Invite tenant users** — Tenant owner/admin can add staff by email when the
   email domain matches the tenant's approved domain policy.
2. **Assign permissions** — Tenant owner/admin can grant roles or direct
   menu-level permissions for back-office pages and actions.
3. **Audit access** — Every invite, acceptance, permission change, suspension,
   and privileged action is tenant-scoped and auditable.

### Scope

### In

- Tenant user invite flow with same-domain validation, invite expiry, resend,
  revoke, accept, and first-login handoff.
- Tenant staff identity records linked to one tenant, with status such as
  invited, active, suspended, removed, and owner.
- Menu-level permission model for tenant portal sections such as dashboard,
  conversations, KM, catalog, avatars, AI settings, billing/quota, customers,
  tickets, reports, integrations, and user management.
- Built-in roles such as Owner, Admin, Operator, Billing, Content Manager,
  Analyst, and Read Only, with optional custom permission overrides.
- Server-side authorization checks for tenant APIs and UI navigation filtering
  so hidden menus cannot be accessed directly by URL/API call.
- Owner protection rules: at least one active owner remains, owners cannot be
  removed by lower-privilege users, and privilege escalation is blocked.
- Audit log events for invites, accepts, login, permission changes, suspension,
  removal, and denied access attempts.
- Platform admin visibility for tenant user count, owner list, suspicious
  invites, and support-only troubleshooting without cross-tenant data mutation.

### Out

- Cross-tenant workforce membership from one login unless a later federation
  feature approves it.
- Public customer account permission management.
- Enterprise SSO/SAML/SCIM provisioning.
- Fine-grained field-level permissions inside every entity.
- Sharing tenant admin credentials as an accepted workflow.

### Deliverables

| Deliverable | Scope |
| --- | --- |
| Tenant user model | Staff membership, role, status, invite token, domain policy, owner protection |
| Invite API/UI | Add same-domain email, resend, revoke, accept, suspend, remove, and list users |
| Permission matrix | Built-in roles, menu/action permissions, custom overrides, navigation filtering |
| Authorization layer | Server-side tenant API permission checks with direct URL/API denial tests |
| Audit trail | Cross-tenant audit events for membership, permission, login, and denied access |
| Verification | Same-domain validation, owner safety, role coverage, API bypass prevention, tenant isolation |

### Acceptance sketch

1. Tenant owner/admin can invite a staff email that matches the tenant's allowed
   domain, and invalid or external-domain emails are rejected with clear errors.
2. Accepted staff users can sign in to the tenant portal and only see menus and
   actions allowed by their role/permission matrix.
3. Direct API or URL access to a restricted tenant menu/action is denied even if
   the UI link is hidden.
4. Permission changes take effect on the next request/session refresh and are
   recorded in the audit log with actor, target user, tenant, and before/after
   values.
5. The system prevents removing or downgrading the last active tenant owner and
   prevents lower-privilege users from granting themselves higher permissions.

---

## Backlog: SPRINT-067 — AI Summary Before Call Close

**Platform:** Customer / Tenant / Tickets · **Feature:** Before a call is
closed, generate an AI summary of the conversation, confirm unresolved items or
next steps, and submit the approved summary into the tenant ticket workflow ·
**Depends:** 1, 21, 22, 23, 24, 43, 53, 55, 56 · **Status:** backlog

### Problem today

Monti can hold a customer conversation and create tickets, but the close-call
handoff is still thin. Tenant staff need a concise, structured summary that
captures the customer's intent, issue, actions already taken, unresolved
questions, sentiment, related products/KM citations, and follow-up ownership.
Without that, support teams must reread transcripts before acting on a ticket.

### Goal

1. **Summarize before closing** — When the caller or AI ends a call, Monti
   generates a structured summary from the transcript and call metadata.
2. **Confirm next steps** — The close flow asks the customer or agent to confirm
   unresolved items, contact details, priority, and whether a ticket should be
   created.
3. **Submit to ticket** — Approved summaries create or update a tenant ticket
   with transcript references, attachments, and audit evidence.

### Scope

### In

- AI-generated close-call summary for voice and text conversations with fields
  such as customer request, issue category, key facts, actions taken,
  unresolved questions, sentiment, priority, suggested next action, and owner.
- Close-call customer UI that displays the summary, asks for confirmation or
  correction where appropriate, and supports "submit ticket" or "close without
  ticket" according to tenant settings.
- Tenant settings to require ticket creation for unresolved calls, allow optional
  ticket creation, or disable ticket submission for selected avatars/topics.
- Ticket payload mapping into the existing ticket workflow with call ID,
  transcript/recording references in MinIO, customer identity, tenant, avatar,
  topic, priority, summary, and attachments.
- Tenant back-office view that shows AI summary, original transcript link,
  confidence/warning flags, and customer confirmation state.
- Safety checks for hallucinated facts, missing transcript segments, sensitive
  data, and low-confidence summaries before ticket submission.
- Audit events for summary generation, customer confirmation, ticket submit,
  summary edit, and close-without-ticket.
- Analytics hooks for summary/ticket rate, unresolved-call rate, and summary
  correction frequency.

### Out

- Replacing the full conversation transcript or recording archive.
- Automatic legal/medical/financial advice classification beyond tenant-defined
  escalation rules.
- Human support SLA engine, assignment queues, or external CRM integrations
  unless already covered by ticket scope.
- Customer-visible ticket portal for tracking status.
- Silent ticket creation without tenant-configured consent/confirmation rules.

### Deliverables

| Deliverable | Scope |
| --- | --- |
| Summary generator | Structured AI summary from transcript, metadata, KM/product citations, and call outcome |
| Close-call UX | Summary preview, confirmation, correction, submit ticket, close without ticket |
| Ticket mapping | Create/update ticket with call references, priority, customer info, attachments, and audit |
| Tenant settings | Per-avatar/topic rules for required, optional, or disabled ticket submission |
| Safety checks | Low-confidence flags, missing transcript warnings, sensitive data handling, edit history |
| Verification | Voice/text close flow, ticket payload, rejected summary, tenant isolation, audit evidence |

### Acceptance sketch

1. Before a voice or text call closes, Monti generates a structured summary that
   references the correct tenant, call ID, customer, avatar, topic, and key
   transcript facts.
2. The caller or configured tenant flow can confirm, correct, submit to ticket,
   or close without ticket according to tenant settings.
3. Submitted tickets include the AI summary, unresolved items, priority,
   customer/contact details, transcript or recording references, and audit
   events.
4. Low-confidence, incomplete, or sensitive summaries are flagged and cannot be
   silently submitted as verified facts.
5. Tenant staff can open the ticket and trace the summary back to the original
   conversation record without crossing tenant boundaries.

---

## Backlog: SPRINT-068 — Call Schedule Email Notifications

**Platform:** Tenant / Customer / Notifications · **Feature:** Let tenants
schedule customer call notifications by topic, send email links for handover or
sales/product conversations, and redirect the customer into an automatically
started voice call with prepared conversation context · **Depends:** 1, 16, 20,
21, 23, 43, 53, 56, 64, 66 · **Status:** backlog

### Problem today

Monti supports customer-initiated calls, but tenants also need a proactive
schedule flow for handover, follow-up, renewal, sales, onboarding, or product
consultation. The customer should receive an email with a clear topic and a
secure link. When clicked, Monti should open the correct tenant/avatar call page
and prepare the AI to speak about the scheduled topic instead of starting from a
generic greeting.

### Goal

1. **Schedule topic calls** — Tenant staff can create scheduled call invites
   with customer email, topic, intent, avatar, product/catalog context, and time
   window.
2. **Send actionable email** — The customer receives a branded email with the
   topic, appointment details, and a secure redirect link.
3. **Auto-start prepared call** — Clicking the link opens the call page and
   starts or primes a speech conversation with the scheduled topic, handover
   notes, and relevant product context.

### Scope

### In

- Tenant schedule UI for one-off call invites with customer email, subject,
  topic, purpose such as handover, sales product, support follow-up, onboarding,
  renewal, or ticket continuation.
- Optional links to ticket, product/catalog item, KM document, previous call
  summary, assigned avatar, language, and preferred call time window.
- Email notification templates with tenant branding, topic, appointment window,
  call purpose, cancel/reschedule policy, and secure call link.
- Signed redirect tokens with expiry, single-use or limited-use policy, tenant
  and customer binding, topic payload, and replay protection.
- Customer redirect flow that lands on the correct call page, checks auth/OTP
  requirements, selects the scheduled tenant/avatar, and prepares the call
  context.
- Auto-start voice call when allowed by browser and tenant policy, with fallback
  "Start call" action when microphone permission or browser autoplay rules block
  automatic start.
- AI conversation primer that includes scheduled topic, handover notes, product
  or ticket context, language, and suggested opening message.
- Status tracking for scheduled, email sent, opened, call started, completed,
  missed, cancelled, expired, and failed.
- Audit and analytics events for schedule create, email send/open, redirect,
  auto-start, fallback start, and call outcome.

### Out

- Bulk marketing campaign automation or newsletter tooling.
- Calendar provider two-way sync unless a later integration feature approves it.
- SMS/LINE/WhatsApp notification channels.
- Cold outbound dialing without customer click/consent.
- Guaranteed browser microphone auto-start when browser permission rules block
  it.

### Deliverables

| Deliverable | Scope |
| --- | --- |
| Schedule model | Tenant-scoped scheduled call invite, customer, topic, avatar, time window, token state |
| Tenant schedule UI | Create, send, resend, cancel, expire, and inspect scheduled call notifications |
| Email notification | Branded template, topic details, secure redirect link, delivery/error status |
| Redirect/start flow | Signed token validation, tenant/avatar selection, auth/OTP gate, auto-start or fallback |
| AI topic primer | Prepared opening prompt with handover notes, product/catalog/ticket context, language |
| Verification | Token expiry, replay protection, wrong tenant denial, browser fallback, audit and status events |

### Acceptance sketch

1. Tenant staff can schedule a topic-based call notification for a customer email
   and send a branded email with the correct tenant, avatar, topic, and call
   window.
2. Clicking the email link validates a signed token, redirects to the correct
   customer call page, and applies the prepared conversation topic without
   exposing another tenant's data.
3. When browser and tenant policy allow it, the call starts automatically in
   speech mode; otherwise the page shows a clear start action with the same
   prepared topic context.
4. The AI opening and follow-up context reflect the scheduled purpose, such as
   handover, sales product discussion, support follow-up, onboarding, renewal,
   or ticket continuation.
5. Expired, revoked, already-used, or tampered links cannot start a call and are
   recorded with safe audit/status events.
