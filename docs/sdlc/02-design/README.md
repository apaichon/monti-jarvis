# 02 — Design

System design artifacts for Monti Jarvis.

**Naming:** `NN-<slug>.md` — numeric prefix is **DES id** (oldest → newest). New docs take the next `DES-NNNN` and matching `NN-` filename.

| # | Doc | Sprint | Description | Status |
| ---: | --- | --- | --- | --- |
| 01 | [01-architecture.md](01-architecture.md) | 1+ | Layers, packages, infra topology | `review_pending` |
| 02 | [02-workflow.md](02-workflow.md) | 1+ | Chat, voice, call, KM, auth, packages, satisfaction, analytics, monitoring sequences | `shipped` |
| 03 | [03-er-diagram.md](03-er-diagram.md) | 1+ | Postgres + ClickHouse + MinIO entities, satisfaction ratings, analytics facts, monitoring storage contract | `shipped` |
| 04 | [04-api-spec.md](04-api-spec.md) | 1+ | REST, WebSocket, SSE, satisfaction, call-center statistics, and monitoring contracts | `shipped` |
| 05 | [05-ux-ui.md](05-ux-ui.md) | 1+ | ASCII wireframes - customer, tenant, platform, analytics, and monitoring surfaces | `shipped` |
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
| 31 | [31-cross-tenant-audit-log-spec.md](31-cross-tenant-audit-log-spec.md) | 28 | Cross-tenant audit event contract, local spool, ClickHouse delivery, retention, and platform query | `approved` |
| 32 | [32-platform-system-performance-spec.md](32-platform-system-performance-spec.md) | 29 | Platform cross-tenant system performance, bounded probes, tenant analytics freshness, and audit delivery health | **`shipped`** v2.10.0 |
| 33 | [33-platform-call-center-statistics-spec.md](33-platform-call-center-statistics-spec.md) | 30 | Platform aggregate call-center statistics, tenant breakdown, freshness, satisfaction, and package labels | **`shipped`** v2.11.0 |
| 34 | [34-platform-billing-quota-ai-cost-spec.md](34-platform-billing-quota-ai-cost-spec.md) | 31 | Platform billing, quota enforcement snapshot, AI usage metering, cost coverage, and reconciliation | `approved` |
| 35 | [35-production-transport-cache-tuning-spec.md](35-production-transport-cache-tuning-spec.md) | 32 | Production transport/cache tuning design track | `review_pending` |
| 37 | [37-theme-color-customization-spec.md](37-theme-color-customization-spec.md) | 39 | Tenant brand chrome (name/logo/subtitle) + color tokens, draft/publish, customer+embed | **`shipped`** v2.15.0 |
| 38 | [38-tenant-ux-bugfix-spec.md](38-tenant-ux-bugfix-spec.md) | 42 | Session expiry, first-login menu, nav groups/scroll, KM document scope | **`shipped`** v2.16.0 |
| 39 | [39-tenant-ai-config-extensibility-spec.md](39-tenant-ai-config-extensibility-spec.md) | 43 | Embed auth, grouped config, encrypted tenant Gemini key, prompts, tools, and skills | `review_pending` |

**Sprint design pack:** Run **`sprint-tech-specs`** when opening each sprint — updates `02`–`05` (cumulative) and adds `NN-<domain>-spec.md` when needed. Templates: `.claude/skills/sprint-tech-specs/references/`.

**Sprint 22:** Design pack approved; implementation in progress — [25-conversation-records-knowledge-gaps-spec.md](25-conversation-records-knowledge-gaps-spec.md); workflow §66–67; API § Conversation Records & Knowledge Gaps; UX T15. [SPRINT-022](../03-sprints/SPRINT-022.md).

**Sprint 23:** Design pack approved; release shipped v2.4.0 — [26-tickets-human-escalation-spec.md](26-tickets-human-escalation-spec.md); workflow §68–70; API § Tickets & Human Escalation; UX T16/C15. [SPRINT-023](../03-sprints/SPRINT-023.md).

**Sprint 24:** Design pack approved and shipped in v2.5.0 — [27-customer-satisfaction-statistics-spec.md](27-customer-satisfaction-statistics-spec.md); workflow §71-72; API § Customer Satisfaction; UX T17/C16. [SPRINT-024](../03-sprints/SPRINT-024.md).

**Sprint 25:** Design pack approved and shipped in v2.6.0 — [28-call-center-statistics-spec.md](28-call-center-statistics-spec.md); workflow §73-75; API § Tenant Call Center Statistics; UX T18. [SPRINT-025](../03-sprints/SPRINT-025.md).

**Sprint 26:** Design pack approved for implementation — [29-tenant-system-performance-spec.md](29-tenant-system-performance-spec.md); workflow §76; API § Tenant System Performance; UX T19. [SPRINT-026](../03-sprints/SPRINT-026.md).

**Sprint 27:** Design pack approved — [30-mobile-call-api-sdk-spec.md](30-mobile-call-api-sdk-spec.md); workflow §77–79; API § Mobile Call API and SDK; UX M1. [SPRINT-027](../03-sprints/SPRINT-027.md).

**Sprint 28:** Design pack approved and shipped in v2.9.0 — [31-cross-tenant-audit-log-spec.md](31-cross-tenant-audit-log-spec.md); workflow §80–82; API § Platform Audit Log; UX A20. [SPRINT-028](../03-sprints/SPRINT-028.md).

**Sprint 29:** Design pack approved and shipped in v2.10.0 — [32-platform-system-performance-spec.md](32-platform-system-performance-spec.md); workflow §83; API § Platform System Performance; UX A21. [SPRINT-029](../03-sprints/SPRINT-029.md).

**Sprint 30:** Design pack approved and shipped in v2.11.0 — [33-platform-call-center-statistics-spec.md](33-platform-call-center-statistics-spec.md); workflow §84; API § Platform Call Center Statistics; UX A22. [SPRINT-030](../03-sprints/SPRINT-030.md).

**Sprint 31:** ✅ Shipped v2.12.0 — [34-platform-billing-quota-ai-cost-spec.md](34-platform-billing-quota-ai-cost-spec.md); workflow §85–86; API § Platform Billing, Quota, and AI Usage; UX A23. [SPRINT-031](../03-sprints/SPRINT-031.md).

**Sprint 32:** ✅ Shipped v2.13.0 — [DES-0035](35-production-transport-cache-tuning-spec.md) remains a design-only roadmap track; [workflow](02-workflow.md) §87–88, [ER/fixture boundary](03-er-diagram.md), [API verification contract](04-api-spec.md), and [UX/UAT operator surface](05-ux-ui.md) support [SPRINT-032](../03-sprints/SPRINT-032.md), reusing [DES-0034](34-platform-billing-quota-ai-cost-spec.md). Manual UAT carries into Sprint 33.

**Sprint 42:** ✅ Shipped v2.16.0 — [38-tenant-ux-bugfix-spec.md](38-tenant-ux-bugfix-spec.md); workflow §91–92; API § Tenant UX Bug Fix; UX T21. [SPRINT-042](../03-sprints/SPRINT-042.md) · [FEAT-0036](../01-features/FEAT-0036-tenant-ux-bugfix.md).

**Sprint 43:** Design pack drafted and review-pending — [39-tenant-ai-config-extensibility-spec.md](39-tenant-ai-config-extensibility-spec.md); workflow §93–96; API § Tenant AI Configuration and Embed Auth; UX T22. [SPRINT-043](../03-sprints/SPRINT-043.md) · [FEAT-0037](../01-features/FEAT-0037-tenant-ai-config-extensibility.md).

**Sprint 39:** ✅ Shipped v2.15.0 — [37-theme-color-customization-spec.md](37-theme-color-customization-spec.md); workflow §89–90; API § Theme Color Customization; UX T20. [SPRINT-039](../03-sprints/SPRINT-039.md) · [FEAT-0035](../01-features/FEAT-0035-theme-color-customization.md).

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
