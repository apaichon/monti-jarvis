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
| 12 | [12-kyc-tenant-spec.md](12-kyc-tenant-spec.md) | 7 | Platform KYC review + approve/reject | `approved` |

**Sprint design pack:** Run **`sprint-tech-specs`** when opening each sprint — updates `02`–`05` (cumulative) and adds `NN-<domain>-spec.md` when needed. Templates: `.claude/skills/sprint-tech-specs/references/`.

**Sprint 7 gate:** ✅ Design pack approved — implement TASK-0030+ per [12-kyc-tenant-spec.md](12-kyc-tenant-spec.md), P12 per [05-ux-ui.md](05-ux-ui.md).

**Sprint 6:** ✅ Shipped v0.7.0 — [11-tenant-register-spec.md](11-tenant-register-spec.md), tenant UI T1–T3 per [05-ux-ui.md](05-ux-ui.md).

**Stack:** Go `net/http` · SvelteKit customer portal · Postgres · Redis 8 · MinIO · ClickHouse · NATS · LiveKit · Gemini

**Verify:** [test matrix](../05-test-scenarios/TEST-MATRIX.md) · [manual tests](../06-manual-tests/) · [deploy](../07-deployment/LOCAL-DEV.md)

**Excel exports:** [excel-output/](excel-output/) mirrors the same `NN-` prefix order.