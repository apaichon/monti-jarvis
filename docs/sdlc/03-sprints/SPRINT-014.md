---
id: SPRINT-014
status: completed
start: 2026-07-11
end: 2026-07-12
closed: 2026-07-12
updated: 2026-07-12
release_target: v1.5.0
release: v1.5.0
goal: "Tenant: Embed to Web — public embed key, loader snippet/iframe, embed-mode customer UI, tenant admin config."
roadmap_sprint: 14
platform: Tenant
depends_on: [SPRINT-001, SPRINT-006]
---

# SPRINT-014 — Tenant: Embed to Web

## Goal

Let **active tenants** embed Monti conversation on their website via a **copy-paste snippet** (loader → iframe), scoped by a **public embed key**, with origin allowlist and a tenant admin config screen.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 closed (S11–S13) | 14, 14, 16 → **avg ~15** |
| Trailing average | **16** |
| **Commitment** | **16** |
| **Completed** | **16** |

## Context

| Sprint | Capability used |
| --- | --- |
| 1 | Customer chat + voice portal |
| 6–7 | Tenant identity, `active` after KYC |
| 13 | Quota applies to resolved tenant on hot paths |

**Gap:** No public multi-tenant entry except default `demo` / `AUTH_DISABLED`.

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0062](../04-tasks/TASK-0062.md) | 3 | completed | devops | `tenant_embed_configs` schema + lazy create |
| [TASK-0063](../04-tasks/TASK-0063.md) | 5 | completed | dev | Public resolve API, `parent_origin`, origin allowlist, tenant CRUD |
| [TASK-0064](../04-tasks/TASK-0064.md) | 4 | completed | dev | Loader JS + `/embed` portrait/voice/chat UI |
| [TASK-0065](../04-tasks/TASK-0065.md) | 3 | completed | dev | Tenant UI `/tenant/embed` — snippet, rotate, origins |
| [TASK-0066](../04-tasks/TASK-0066.md) | 1 | completed | tester | Manual UAT checklist + unit smoke |

**Committed:** 16 · **Completed:** 16

## Shipped summary (v1.5.0)

- Per-tenant embed config (`tenant_embed_configs`): key, enabled, allowed origins, default agent
- Public `GET /api/public/embed/{key}` with host-site `parent_origin` allowlist check
- `monti-embed.js` floating launcher + iframe; close control top-right (not over Send)
- Embed customer UI: avatar portrait, voice call, text chat; secure-context mic guidance
- Tenant admin `/tenant/embed` + integrator guide [EMBED_WEB_INTEGRATION.md](../../../EMBED_WEB_INTEGRATION.md) (security §7)

## Scope boundary

**In**
- One embed config per tenant (enable, key, allowed_origins, default_agent_id optional)
- Public `GET /api/public/embed/{embed_key}`
- Static loader e.g. `/embed/monti-embed.js`
- Embed customer UI (compact) for workforce + chat + voice when secure context
- Tenant admin APIs: get/put embed config, rotate key
- Origin allowlist enforcement when list non-empty

**Out**
- KM/scope tenant admin (**S15**)
- Locale/settings/limits (**S16**)
- Preview sandbox (**S17**)
- Customer accounts (**S19–20**)
- Custom domain / CNAME
- npm package publish

## Feature

- [FEAT-0014 — Embed to Web](../01-features/FEAT-0014-embed-to-web.md)

## Design pack (`sprint-tech-specs`)

| Artifact | Path | Status |
| --- | --- | --- |
| Embed deep spec | [17-embed-to-web-spec.md](../02-design/17-embed-to-web-spec.md) | shipped |
| Workflow | [02-workflow.md](../02-design/02-workflow.md) §37–39 | shipped |
| ER diagram | [03-er-diagram.md](../02-design/03-er-diagram.md) | shipped |
| API spec | [04-api-spec.md](../02-design/04-api-spec.md) § Embed | shipped |
| UX/UI ASCII | [05-ux-ui.md](../02-design/05-ux-ui.md) § T7 + E1 | shipped |

## Verification

```bash
make build && make test
go test ./internal/store/ ./cmd/server/ -count=1 -run Embed
# Tenant: /tenant/embed → copy snippet
# Host page with snippet → chat; voice needs HTTPS or localhost Monti host
```

- **Manual UAT:** [SPRINT-014-manual.md](../06-manual-tests/SPRINT-014-manual.md) (TASK-0066) — checklist shipped; full browser sign-off optional
- Integrator + security: [EMBED_WEB_INTEGRATION.md](../../../EMBED_WEB_INTEGRATION.md)

## Risks

| Risk | Mitigation |
| --- | --- |
| Clickjacking / open embed | Origin allowlist + host CSP `frame-src` |
| Key leakage | Rotate key; treat as public capability token |
| Mic on custom HTTP hosts | Document secure context; use localhost/HTTPS |
| parent_origin client-asserted | Soft gate; quotas + KM content review |

## Links

- Depends: [SPRINT-001](SPRINT-001.md), [SPRINT-006](SPRINT-006.md)
- Integrator guide: [docs/EMBED_WEB_INTEGRATION.md](../../../EMBED_WEB_INTEGRATION.md)
- Next: SPRINT-015 Set Scope and KM
