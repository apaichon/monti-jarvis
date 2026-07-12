---
id: SPRINT-015
status: completed
start: 2026-07-12
end: 2026-07-12
closed: 2026-07-12
updated: 2026-07-12
design_pack: shipped
release_target: v1.6.0
release: v1.6.0
goal: "Tenant: Set Scope and KM — self-service knowledge upload, scope tagging, document lifecycle, and tenant admin UI."
roadmap_sprint: 15
platform: Tenant
depends_on: [SPRINT-002, SPRINT-006]
---

# SPRINT-015 — Tenant: Set Scope and KM

## Goal

Give **active tenant admins** a portal to manage **knowledge documents** and **scopes** for their AI agents — upload, list, retag, delete, reset — so go-live tenants are not dependent on platform seed/`curl`.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S12–S14) | 14, 16, 16 → **avg ~15** |
| Trailing average | **16** |
| **Commitment** | **16** |
| **Completed** | **16** |

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0067](../04-tasks/TASK-0067.md) | 3 | completed | devops | Delete cascade; `km_gaps` schema; route wiring |
| [TASK-0068](../04-tasks/TASK-0068.md) | 5 | completed | dev | `/api/tenant/km/*` + quota |
| [TASK-0069](../04-tasks/TASK-0069.md) | 4 | completed | dev | Tenant UI `/tenant/km` |
| [TASK-0070](../04-tasks/TASK-0070.md) | 3 | completed | dev | Scope picker + matrix + gaps panel |
| [TASK-0071](../04-tasks/TASK-0071.md) | 1 | completed | tester | Manual UAT checklist |

**Committed:** 16 · **Completed:** 16

## Shipped summary (v1.6.0)

- Tenant KM admin APIs (scopes, agents, upload, list, scope patch, delete, reset)
- `callcenter.km_gaps` FAQ backlog on `missing_km` + list/patch APIs
- Tenant UI `/tenant/km` (Knowledge nav)
- ClickHouse inserts via **JSONEachRow** (safe for large markdown KM)
- Chat/voice RAG scoped to request **tenant_id** (fix embed multi-tenant bug)
- Voice RAG richer preload for Gemini Live product FAQs
- Shared OAuth paths: `/api/public/tenant/oauth/{provider}/callback` (login + register)
- ChillPay QR channel `bank_qrcode`
- Sample KM + embed fixture for B-Quik tyre highlight

## Design pack

| Artifact | Status |
| --- | --- |
| [18-tenant-scope-km-spec.md](../02-design/18-tenant-scope-km-spec.md) | shipped |
| Workflow §40–45 | shipped |
| ER + `km_gaps` | shipped |
| API Tenant KM | shipped |
| UX T8 | shipped |

## Links

- Feature: [FEAT-0015](../01-features/FEAT-0015-tenant-scope-km.md)
- UAT: [SPRINT-015-manual.md](../06-manual-tests/SPRINT-015-manual.md)
- Next: SPRINT-016 Settings / Locale / Limits
