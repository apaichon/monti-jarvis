# 02 — Design

System design artifacts for Monti Jarvis.

**Naming:** `NN-<slug>.md` — numeric prefix is **DES id** (oldest → newest). New docs take the next `DES-NNNN` and matching `NN-` filename.

| # | Doc | Sprint | Description | Status |
| ---: | --- | --- | --- | --- |
| 01 | [01-architecture.md](01-architecture.md) | 1+ | Layers, packages, infra topology | `review_pending` |
| 02 | [02-workflow.md](02-workflow.md) | 1+ | Chat, voice, call, KM, auth, packages, satisfaction, analytics, monitoring sequences | `approved` |
| 03 | [03-er-diagram.md](03-er-diagram.md) | 1+ | Postgres + ClickHouse + MinIO entities, satisfaction ratings, analytics facts, monitoring storage contract | `approved` |
| 04 | [04-api-spec.md](04-api-spec.md) | 1+ | REST, WebSocket, SSE, satisfaction, call-center statistics, and monitoring contracts | `approved` |
| 05 | [05-ux-ui.md](05-ux-ui.md) | 1+ | ASCII wireframes - customer, tenant, platform, analytics, and monitoring surfaces | `approved` |
| 06 | [06-auth-spec.md](06-auth-spec.md) | 3 | JWT, RBAC, route policy | `shipped` |
| 07 | [07-auth-cache-events-spec.md](07-auth-cache-events-spec.md) | 3 | Redis cache, write-behind, NATS auth events | `approved` |
| 08 | [08-packages-spec.md](08-packages-spec.md) | 4 | Package catalog + tenant entitlements (jsonb rules) | `approved` |
| 09 | [09-platform-admin-portal-spec.md](09-platform-admin-portal-spec.md) | 4 | Platform admin portal `/admin` | `shipped` |
| 10 | [10-avatars-spec.md](10-avatars-spec.md) | 5 | Avatar catalog + tenant assignment | `approved` |
| 11 | [11-tenant-register-spec.md](11-tenant-register-spec.md) | 6 | Tenant self-registration + pending_kyc | `approved` |
| 12 | [12-kyc-tenant-spec.md](12-kyc-tenant-spec.md) | 7 | Platform KYC review + approve/reject | `shipped` |
| 13 | [13-payment-gateway-spec.md](13-payment-gateway-spec.md) | 8 | ChillPay gateway config + callbacks | `approved` |
| 14 | [14-buy-package-spec.md](14-buy-package-spec.md) | 9 | Tenant checkout + callback fulfillment | `approved` |
| 15 | [15-commerce-chain-plan.md](15-commerce-chain-plan.md) | 9–12 | Phase C commerce chain re-scope (S8–S12) | `approved` |
| 16 | [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md) | 13 | Redis quotas + rate limits + platform usage | **`shipped`** v1.4.0 |
| 17 | [17-embed-to-web-spec.md](17-embed-to-web-spec.md) | 14 | Tenant web embed widget + public key | **`shipped`** v1.5.0 |
| 18 | [18-tenant-scope-km-spec.md](18-tenant-scope-km-spec.md) | 15 | Tenant self-service KM + scopes | **`shipped`** v1.6.0 |
| 19 | [19-tenant-settings-limits-spec.md](19-tenant-settings-limits-spec.md) | 16 | Settings, locale, call limits | **`shipped`** v1.7.0 |
| 20 | [20-tenant-test-preview-spec.md](20-tenant-test-preview-spec.md) | 17 | Tenant test/preview sandbox | **`shipped`** v1.8.0 |
| 21 | [21-customer-tier-spec.md](21-customer-tier-spec.md) | 18 | Customer tier catalog + groups | **`shipped`** v1.9.0 |
| 22 | [22-customer-account-import-spec.md](22-customer-account-import-spec.md) | 19 | Customer directory, CSV imports, domain defaults, integration identity | **`shipped`** v2.0.0 |
| 23 | [23-customer-auth-spec.md](23-customer-auth-spec.md) | 20 | Customer credentials, sessions, domain enforcement, quota isolation | `shipped` |
| 24 | [24-authenticated-workforce-selection-spec.md](24-authenticated-workforce-selection-spec.md) | 21 | Authenticated workforce selection and customer quota limits | `approved` |
| 25 | [25-conversation-records-knowledge-gaps-spec.md](25-conversation-records-knowledge-gaps-spec.md) | 22 | Conversation records, MinIO archive, and knowledge gaps | `approved` |
| 26 | [26-tickets-human-escalation-spec.md](26-tickets-human-escalation-spec.md) | 23 | Tenant tickets and customer-confirmed human escalation | `approved` |
| 27 | [27-customer-satisfaction-statistics-spec.md](27-customer-satisfaction-statistics-spec.md) | 24 | Customer 1-5 satisfaction review and tenant statistics | `approved` |
| 28 | [28-call-center-statistics-spec.md](28-call-center-statistics-spec.md) | 25 | Tenant ClickHouse call-center statistics and quota usage | `approved` |
| 29 | [29-tenant-system-performance-spec.md](29-tenant-system-performance-spec.md) | 26 | Tenant-safe dependency health, latency, and analytics freshness | `approved` |
| 30 | [30-mobile-call-api-sdk-spec.md](30-mobile-call-api-sdk-spec.md) | 27 | Versioned mobile call API, voice transport, and typed SDK contract | `approved` |

**Sprint design pack:** Run **`sprint-tech-specs`** when opening each sprint — updates `02`–`05` (cumulative) and adds `NN-<domain>-spec.md` when needed. Templates: `.claude/skills/sprint-tech-specs/references/`.

**Sprint 22:** Design pack approved; implementation in progress — [25-conversation-records-knowledge-gaps-spec.md](25-conversation-records-knowledge-gaps-spec.md); workflow §66–67; API § Conversation Records & Knowledge Gaps; UX T15. [SPRINT-022](../03-sprints/SPRINT-022.md).

**Sprint 23:** Design pack approved; release shipped v2.4.0 — [26-tickets-human-escalation-spec.md](26-tickets-human-escalation-spec.md); workflow §68–70; API § Tickets & Human Escalation; UX T16/C15. [SPRINT-023](../03-sprints/SPRINT-023.md).

**Sprint 24:** Design pack approved and shipped in v2.5.0 — [27-customer-satisfaction-statistics-spec.md](27-customer-satisfaction-statistics-spec.md); workflow §71-72; API § Customer Satisfaction; UX T17/C16. [SPRINT-024](../03-sprints/SPRINT-024.md).

**Sprint 25:** Design pack approved and shipped in v2.6.0 — [28-call-center-statistics-spec.md](28-call-center-statistics-spec.md); workflow §73-75; API § Tenant Call Center Statistics; UX T18. [SPRINT-025](../03-sprints/SPRINT-025.md).

**Sprint 26:** Design pack approved for implementation — [29-tenant-system-performance-spec.md](29-tenant-system-performance-spec.md); workflow §76; API § Tenant System Performance; UX T19. [SPRINT-026](../03-sprints/SPRINT-026.md).

**Sprint 27:** Design pack approved — [30-mobile-call-api-sdk-spec.md](30-mobile-call-api-sdk-spec.md); workflow §77–79; API § Mobile Call API and SDK; UX M1. [SPRINT-027](../03-sprints/SPRINT-027.md).

**Sprint 21:** Design pack approved; implementation in progress — [24-authenticated-workforce-selection-spec.md](24-authenticated-workforce-selection-spec.md); workflow §64–65; API § Authenticated Workforce Selection & Customer Quota; UX C14/T14. [SPRINT-021](../03-sprints/SPRINT-021.md).

**Sprint 20:** ✅ Shipped v2.1.0 — [23-customer-auth-spec.md](23-customer-auth-spec.md); workflow §59–63; API § Customer Authentication & Domain Enforcement; UX T13. [SPRINT-020](../03-sprints/SPRINT-020.md).

**Sprint 19:** ✅ Shipped v2.0.0 — [22-customer-account-import-spec.md](22-customer-account-import-spec.md); workflow §55–58; API § Customer Accounts & Imports; UX T12. [SPRINT-019](../03-sprints/SPRINT-019.md).

**Sprint 18:** ✅ Shipped v1.9.0 — [21-customer-tier-spec.md](21-customer-tier-spec.md); workflow §52–54; API § Tiers; UX T11. [SPRINT-018](../03-sprints/SPRINT-018.md).

**Sprint 17:** ✅ Shipped v1.8.0 — [20-tenant-test-preview-spec.md](20-tenant-test-preview-spec.md); workflow §49–51; API § Preview; UX T10. [SPRINT-017](../03-sprints/SPRINT-017.md).

**Sprint 16:** ✅ Shipped v1.7.0 — [19-tenant-settings-limits-spec.md](19-tenant-settings-limits-spec.md); workflow §46–48; API § Settings; UX T9.

**Sprint 15:** ✅ Shipped v1.6.0 — [18-tenant-scope-km-spec.md](18-tenant-scope-km-spec.md); workflow §40–45; API § Tenant KM; UX T8; `km_gaps`. [SPRINT-015](../03-sprints/SPRINT-015.md).

**Sprint 14:** ✅ Shipped v1.5.0 — [17-embed-to-web-spec.md](17-embed-to-web-spec.md); workflow §37–39; API § Embed; UX T7/E1.

**Sprint 13:** ✅ Shipped v1.4.0 — [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md); workflow §32–36; API § Quota; UX P14. [SPRINT-013](../03-sprints/SPRINT-013.md).

**Sprint 9 gate:** ✅ Design pack approved — TASK-0040+ per [14-buy-package-spec.md](14-buy-package-spec.md) + chain plan [15](15-commerce-chain-plan.md). Combined E2E with Sprint 8; Phase C closed v1.3.0.

**Sprint 8:** ✅ Code shipped — [13-payment-gateway-spec.md](13-payment-gateway-spec.md), P13 per [05-ux-ui.md](05-ux-ui.md).

**Sprint 7:** ✅ Shipped v0.8.0 — [12-kyc-tenant-spec.md](12-kyc-tenant-spec.md), P12 per [05-ux-ui.md](05-ux-ui.md).

**Sprint 6:** ✅ Shipped v0.7.0 — [11-tenant-register-spec.md](11-tenant-register-spec.md), tenant UI T1–T3 per [05-ux-ui.md](05-ux-ui.md).

**Stack:** Go `net/http` · SvelteKit customer portal · Postgres · Redis 8 · MinIO · ClickHouse · NATS · LiveKit · Gemini

**Verify:** [test matrix](../05-test-scenarios/TEST-MATRIX.md) · [manual tests](../06-manual-tests/) · [deploy](../07-deployment/LOCAL-DEV.md)

**Excel exports:** [excel-output/](excel-output/) mirrors the same `NN-` prefix order.
