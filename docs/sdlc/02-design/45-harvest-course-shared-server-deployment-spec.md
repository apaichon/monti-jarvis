---
id: DES-0045
title: Harvest-course Shared-Server Deployment Specification
status: approved
updated: 2026-07-27
sprint: SPRINT-049
owner: SA
---

# DES-0045 — Harvest-course Shared-Server Deployment

**Sprint:** SPRINT-049 · **Release target:** deployment-only  
**Task:** [TASK-0181](../04-tasks/TASK-0181.md)  
**Depends on:** [43-product-web-growth-spec.md](43-product-web-growth-spec.md),
[44-ai-call-center-security-hardening-spec.md](44-ai-call-center-security-hardening-spec.md)

## 1. Goals

- Package the customer, platform-admin, tenant, and product web builds into one
  Monti image.
- Run the image beside the existing Harvest-course services on the same host.
- Preserve one same-origin contract for browser cookies, API calls, SSE, and
  voice/preview WebSockets.
- Make deployment deterministic through explicit source validation, environment
  variables, Nginx validation, health checks, and rollback instructions.

## 2. Non-goals (Sprint 49)

- Changing any customer, tenant, admin, or product feature behavior.
- Adding a new database table, migration, API endpoint, domain, or TLS
  certificate.
- Moving Harvest-course services into the Monti image.
- Executing a production rollout, rotating real secrets, or changing DNS.
- Splitting the four portals into separate containers or origins.

## 3. Deployment topology

```text
Browser
  │ HTTPS monti.devclub.dev
  ▼
Harvest-course host Nginx :80/:443
  │ one upstream, WebSocket upgrade preserved
  ▼
monti-jarvis container 127.0.0.1:${MONTI_HOST_PORT:-18092}:8091
  ├─ /              customer-web/build
  ├─ /admin/*       platform-admin-web/build
  ├─ /tenant/*      tenant-web/build
  ├─ /product/*     product-web/build
  └─ /api/* /ws/*   Go handlers and voice/preview transports
```

The existing Go router owns path dispatch. Nginx must proxy the full origin to
the same upstream; it must not mount individual portal directories or rewrite
`/admin`, `/tenant`, or `/product` to a different host.

| Surface | Public path | Runtime directory |
| --- | --- | --- |
| Customer portal | `/` | `/app/apps/customer-web/build` |
| Platform admin | `/admin/` | `/app/apps/platform-admin-web/build` |
| Tenant web | `/tenant/` | `/app/apps/tenant-web/build` |
| Product web | `/product/` | `/app/apps/product-web/build` |

## 4. Build contract

The deployment Dockerfile uses a Node build stage and a Go runtime stage.
Each portal must provide `package.json` and `package-lock.json`; each build
must produce a non-empty `build/index.html` before the image is exported.

```text
apps/customer-web/package*.json       → npm ci → build/
apps/platform-admin-web/package*.json → npm ci → build/
apps/tenant-web/package*.json         → npm ci → build/
apps/product-web/package*.json        → npm ci → build/
                                              ↓
                                     Go binary + four build dirs
```

The deployment script validates all four lockfiles before creating the build
context. Generated `node_modules`, `.svelte-kit`, existing `build` output,
`.git`, worktrees, and local runtime state are excluded from the context.

## 5. Environment contract

| Variable | Required | Default | Rule |
| --- | --- | --- | --- |
| `MONTI_SOURCE_DIR` | no | `repos/monti-jarvis` | Checkout used as build source; wrapper overrides it explicitly |
| `MONTI_PUBLIC_DOMAIN` | production | `monti.devclub.dev` | Must resolve to the Harvest-course host |
| `MONTI_HOST_PORT` | no | `18092` | Loopback-only host port; must not collide with Harvest services |
| `MONTI_PRODUCT_WEB_ENABLED` | no | `true` | Must remain true to serve `/product/` |
| `MONTI_AUTH_DISABLED` | production | `false` | Production must not use the anonymous demo bypass |
| `MONTI_LIVEKIT_DOMAIN` | no | `monti-live.devclub.dev` | Separate WebRTC signaling hostname |
| `MONTI_*` secrets | yes | none | Postgres, Redis, MinIO, ClickHouse, Gemini, LiveKit, JWT |

Inside the container the application paths are explicit:

```text
CUSTOMER_WEB_DIR=/app/apps/customer-web/build
PLATFORM_ADMIN_WEB_DIR=/app/apps/platform-admin-web/build
TENANT_WEB_DIR=/app/apps/tenant-web/build
PRODUCT_WEB_DIR=/app/apps/product-web/build
PRODUCT_WEB_ENABLED=true
```

Secrets are supplied by the host environment file and are never copied into
the build context or browser bundles.

## 6. Runtime and Nginx contract

- The Go process listens on container port `8091`.
- Docker publishes only `127.0.0.1:${MONTI_HOST_PORT}:8091`.
- Nginx terminates TLS, proxies `Host`, `X-Real-IP`, forwarded headers, and
  WebSocket `Upgrade`/`Connection` headers.
- Voice and preview connections use a long read/send timeout; normal static
  page requests still use the same upstream.
- Nginx configuration is installed only after `nginx -t` succeeds.
- A failed Nginx validation must leave the currently loaded configuration
  untouched and must stop the deployment before reload.

## 7. Deployment sequence

1. Wrapper resolves the Monti checkout and HarvestMax deployment checkout.
2. Deployment script loads environment files and validates required source
   files, Docker, Compose, and the network.
3. Source is copied into a temporary sanitized build context.
4. Docker builds and tags the four-portal image.
5. Compose starts/recreates `monti-jarvis` after dependencies are healthy.
6. Host Nginx vhosts are installed and validated, then reloaded.
7. Direct `/healthz` and public `/healthz`, `/`, `/admin/`, `/tenant/`, and
   `/product/` smoke checks run.
8. The deployment exits non-zero if any required check fails.

## 8. Rollback and failure handling

- Application rollback restores the previous image or source ref and reruns
  the same deployment command.
- Persistent Postgres, Redis, MinIO, NATS, ClickHouse, and audit volumes are
  never removed by an application rollback.
- If Nginx validation fails, do not reload Nginx; inspect `nginx -t` output and
  keep the existing active vhost.
- If a portal check fails after restart, inspect the Monti container logs and
  restore the previous image before changing DNS.

## 9. Data and security impact

This sprint adds no Postgres, Redis, NATS, ClickHouse, or MinIO entity. It
reuses existing application authorities and path routing. Same-origin serving
reduces cross-origin configuration, while production authentication and the
S41 security gate remain mandatory before customer traffic is opened.

## 10. API summary

No new API endpoint is introduced. The deployment verifies existing public
contracts:

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| `GET` | `/healthz` | public/operator | Liveness and safe dependency summary |
| `GET` | `/readyz` | operator | Readiness and production safety gate |
| `GET` | `/` | customer/public | Customer portal shell |
| `GET` | `/admin/*` | platform session in app | Platform admin shell |
| `GET` | `/tenant/*` | tenant session in app | Tenant web shell |
| `GET` | `/product/*` | public | Product marketing shell |
| `GET` | `/ws/*` | app/session dependent | Existing voice and preview upgrades |

## 11. RBAC

| Action | Browser | Host operator | Platform admin | Tenant admin |
| --- | --- | --- | --- | --- |
| View public/customer/product paths | yes | yes | yes | yes |
| View admin/tenant paths | app auth required | yes | app role required | app role required |
| Build image | no | yes | no | no |
| Install/reload Nginx | no | yes | no | no |
| Read deployment secrets | no | controlled host access | no | no |

## 12. Pre-Implementation Specs

- [x] Path-based same-origin topology is selected and matches existing Go
  router dispatch.
- [x] Four lockfiles and four build outputs are defined as image requirements.
- [x] Compose variables and runtime directories are defined for local and
  production overlays.
- [x] Nginx ownership, WebSocket headers, loopback port, and validation gate
  are defined.
- [x] Health and four-portal smoke checks are defined.
- [x] Rollback explicitly preserves persistent data volumes.

## 13. Code Style Guide / Folder Style

- Keep deployment orchestration in `harvest-deployment`; keep the Monti-side
  wrapper in `scripts/deploy-shared-server.sh`.
- Use Bash `set -euo pipefail`, explicit absolute/validated paths, and safe
  non-secret diagnostics.
- Keep Dockerfile stages ordered as dependency manifests → installs → source
  copy → builds → runtime copies.
- Keep Nginx as one full-origin proxy; do not duplicate application route logic
  in Nginx.
- Keep environment names prefixed `MONTI_` in host config and map them to the
  application’s existing non-prefixed runtime variables in Compose.

## 14. Verification

```bash
make product-web
make build
go test ./... -count=1
./scripts/deploy-shared-server.sh --help
docker compose -f prod/docker-compose.yml -f prod/docker-compose-monti.yml config --quiet
docker build -f docker/monti-jarvis/Dockerfile .
curl -fsS https://monti.devclub.dev/healthz
curl -fsS https://monti.devclub.dev/
curl -fsS https://monti.devclub.dev/admin/
curl -fsS https://monti.devclub.dev/tenant/
curl -fsS https://monti.devclub.dev/product/
nginx -t
git diff --check
```

## Approver sign-off

| Role | Name | Date | Approved |
| --- | --- | --- | --- |
| Requester / PM | User request | 2026-07-26 | ☑ |
| DevOps | Codex | 2026-07-26 | ☑ |

See [02-workflow.md](02-workflow.md) Sprint 49, [03-er-diagram.md](03-er-diagram.md)
Sprint 49, [04-api-spec.md](04-api-spec.md) Sprint 49, and
[05-ux-ui.md](05-ux-ui.md) Sprint 49.
