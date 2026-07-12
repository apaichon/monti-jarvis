---
id: DES-0017
title: Tenant Embed to Web Specification
status: shipped
updated: 2026-07-12
sprint: SPRINT-014
owner: SA
release: v1.5.0
---

# Embed to Web — Design Spec

**Sprint:** SPRINT-014 · **Release:** v1.5.0  
**Feature:** [FEAT-0014](../01-features/FEAT-0014-embed-to-web.md)  
**Depends on:** customer conversation (S1), tenant active (S6–7), quotas (S13)

## 1. Goals

- Tenant pastes one script tag on their site → floating Monti widget.
- Traffic scoped to tenant via **embed_key** (not spoofable `tenant_id` alone).
- Origin allowlist reduces drive-by abuse of a leaked key.
- Hosted demo portal remains unchanged.

## 2. Non-goals

- Custom domain / CNAME white-label
- npm SDK
- Customer login inside widget
- Tenant KM editor (S15)

## 3. Environment

| Variable | Default | Notes |
| --- | --- | --- |
| `APP_PUBLIC_URL` | `http://localhost:8091` | Used in snippet host |
| `EMBED_ALLOW_EMPTY_ORIGINS` | `true` | When allowlist empty, allow any Origin (dev) |

## 4. Data model

### `tenant_embed_configs`

| Column | Type | Notes |
| --- | --- | --- |
| `tenant_id` | text PK FK | → `tenants.id` |
| `embed_key` | text UK | e.g. `emb_` + 32 hex |
| `enabled` | bool | default false |
| `allowed_origins` | jsonb | `["https://shop.example"]` |
| `default_agent_id` | text null | workforce agent id |
| audit cols | | `created_at`, `updated_at`, `created_by`, `updated_by` |

## 5. Public flow

```text
Host page loads monti-embed.js? data-embed-key=KEY
  → JS creates iframe src={PUBLIC}/embed?key=KEY
  → Embed page GET /api/public/embed/KEY?parent_origin=HOST_ORIGIN
  → 200 config → set tenant context → chat/voice APIs with X-Tenant-Id
```

Quota (S13) uses resolved `tenant_id`.

## 6. API summary

See [04-api-spec.md](04-api-spec.md) § Embed.

| Method | Path | Auth |
| --- | --- | --- |
| `GET` | `/api/public/embed/{embed_key}` | public |
| `GET` | `/api/tenant/embed` | tenant_admin active |
| `PUT` | `/api/tenant/embed` | tenant_admin active |
| `POST` | `/api/tenant/embed/rotate-key` | tenant_admin active |
| `GET` | `/embed/monti-embed.js` | public static |
| `GET` | `/embed` or `/embed/` | public SPA embed mode |

### Public resolve 200

```json
{
  "tenant_id": "demo",
  "slug": "demo",
  "name": "Demo Workspace",
  "embed_key": "emb_…",
  "enabled": true,
  "default_agent_id": "ava",
  "agents": [{ "id": "ava", "name": "Ava" }]
}
```

### Errors

| HTTP | code |
| ---: | --- |
| 404 | `embed_not_found`, `embed_disabled` |
| 403 | `origin_not_allowed` |
| 403 | tenant not active (admin APIs) |

## 7. Snippet

```html
<script
  src="http://localhost:8091/embed/monti-embed.js"
  data-embed-key="emb_REPLACE"
  data-position="bottom-right"
  async
></script>
```

## 8. Security

| Control | Behavior |
| --- | --- |
| embed_key | Public capability; rotate invalidates |
| allowed_origins | Match scheme+host[+port]; empty = permissive when env allows |
| frame-ancestors | Prefer CSP listing allowed origins or `*` if empty |
| No secrets in iframe URL | key is public by design |

## 9. RBAC

| Action | platform | tenant_admin | public |
| --- | --- | --- | --- |
| Manage embed config | no* | yes (own) | no |
| Resolve embed | — | — | yes if enabled |
| Use chat in embed | — | — | yes (tenant scoped) |

\* Platform may later add ops view — out of S14.

## 10. Verification

```bash
# After tenant enables embed
curl -s http://localhost:8091/api/public/embed/$KEY \
  -H 'Origin: https://allowed.example' | jq .
```

## 11. Related

| Artifact | Path |
| --- | --- |
| **Integrator guide** | [docs/EMBED_WEB_INTEGRATION.md](../../EMBED_WEB_INTEGRATION.md) |
| Workflow | [02-workflow.md](02-workflow.md) §37–39 |
| ER | [03-er-diagram.md](03-er-diagram.md) |
| API | [04-api-spec.md](04-api-spec.md) § Embed |
| UX | [05-ux-ui.md](05-ux-ui.md) § T7 · E1 |
| Sprint | [SPRINT-014](../03-sprints/SPRINT-014.md) |
