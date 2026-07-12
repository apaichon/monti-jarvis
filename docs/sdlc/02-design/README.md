# 02 — Design

System design artifacts for Monti Jarvis.

**Naming:** `NN-<slug>.md` — numeric prefix is **DES id** (oldest → newest). New docs take the next `DES-NNNN` and matching `NN-` filename.

| # | Doc | Sprint | Description | Status |
| ---: | --- | --- | --- | --- |
| 01 | [01-architecture.md](01-architecture.md) | 1+ | Layers, packages, infra topology | `review_pending` |
| 02 | [02-workflow.md](02-workflow.md) | 1+ | Chat, voice, call, KM, auth, packages sequences | `approved` |
| 03 | [03-er-diagram.md](03-er-diagram.md) | 1+ | Postgres + ClickHouse + MinIO entities | `approved` |
| 04 | [04-api-spec.md](04-api-spec.md) | 1+ | REST, WebSocket, SSE contract | `approved` |
| 05 | [05-ux-ui.md](05-ux-ui.md) | 1+ | ASCII wireframes — customer + platform admin P0–P10 | `approved` |
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

**Sprint design pack:** Run **`sprint-tech-specs`** when opening each sprint — updates `02`–`05` (cumulative) and adds `NN-<domain>-spec.md` when needed. Templates: `.claude/skills/sprint-tech-specs/references/`.

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