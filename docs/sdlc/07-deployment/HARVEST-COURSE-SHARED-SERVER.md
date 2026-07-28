---
id: DEPLOY-0049
status: planned
environment: harvest-course-shared-server
sprint: SPRINT-049
updated: 2026-07-26
---

# Harvest-course Shared-Server Deployment

Monti is deployed as one backend/image behind the existing HarvestMax host
Nginx. The four web surfaces remain same-origin and are selected by path:

| Surface | URL |
| --- | --- |
| Customer portal | `https://monti.devclub.dev/` |
| Platform admin | `https://monti.devclub.dev/admin/` |
| Tenant web | `https://monti.devclub.dev/tenant/` |
| Product web | `https://monti.devclub.dev/product/` |
| API and readiness | `https://monti.devclub.dev/healthz`, `/readyz` |

## Host layout

- `/opt/harvest-deployment` owns Docker Compose, host Nginx, TLS files, and
  deployment logs.
- `monti-jarvis` runs on the loopback host port configured by
  `MONTI_HOST_PORT` (default `18092`).
- Nginx proxies the full origin to Monti, preserving WebSocket upgrade headers
  for voice and preview traffic.
- The build source is the Monti checkout; the deployment checkout is not
  copied into the image.

## Configuration

Copy `prod/.env.monti.example` in the HarvestMax deployment checkout to
`prod/.env.monti` and set real values. At minimum, set the database, Redis,
MinIO, ClickHouse, Gemini, LiveKit, and JWT secrets. Keep
`MONTI_AUTH_DISABLED=false` in production.

The compose service sets these image paths:

```text
CUSTOMER_WEB_DIR=/app/apps/customer-web/build
PLATFORM_ADMIN_WEB_DIR=/app/apps/platform-admin-web/build
TENANT_WEB_DIR=/app/apps/tenant-web/build
PRODUCT_WEB_DIR=/app/apps/product-web/build
PRODUCT_WEB_ENABLED=true
```

## Deploy

From the Monti checkout:

```bash
./scripts/deploy-shared-server.sh --environment production
```

For a remote/server checkout, point the wrapper at the deployment repo:

```bash
HARVEST_DEPLOYMENT_ROOT=/opt/harvest-deployment \
  ./scripts/deploy-shared-server.sh --environment production --ref main
```

The deployment script builds the image, starts the Monti service, installs the
Monti Nginx vhosts, validates Nginx, reloads it, and checks all four portal
paths plus `/healthz`.

## Rollback

Do not delete persistent volumes during an application rollback. Restore the
previous image/tag or source ref, rerun the deployment wrapper with the same
environment file, validate Nginx, then re-run the four-path smoke check.

## First rollout checklist

- DNS for `monti.devclub.dev` points to the Harvest-course server and Cloudflare
  is using Full (strict) for the origin certificate.
- The host Nginx config includes the Monti vhost and does not conflict with
  Harvest-course’s existing `nerd.*` or API vhosts.
- `MONTI_HOST_PORT` is free and bound to `127.0.0.1` only.
- `docker compose ... config` succeeds without unresolved required secrets.
- `/healthz`, `/`, `/admin/`, `/tenant/`, and `/product/` return successful
  responses after reload.
