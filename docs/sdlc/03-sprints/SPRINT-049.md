---
id: SPRINT-049
status: backlog
hold_reason: operator_rollout_pending
start: 2026-07-27
end: 2026-07-31
updated: 2026-07-28
design_pack: approved
release_target: deployment-only
platform: DevOps / Shared Server
depends_on: [SPRINT-041, SPRINT-048]
goal: "Deploy customer, admin, tenant, and product web surfaces from one Monti image on the Harvest-course server."
parallel_track: true
---

# SPRINT-049 — Harvest-course Shared-Server Deployment

## Goal

Make the four existing Monti web surfaces deployable together on the
Harvest-course server, with one same-origin API/backend and explicit build,
environment, Nginx, health-check, and rollback instructions.

## Velocity

| Window | Points |
| --- | ---: |
| Last 3 recorded closed slices (S42, S45, S46) | 12, 13, 17 → **avg 14.0** |
| **Committed** | **8** |

The scope is limited to deployment plumbing and verification. It does not
change product behavior or imply a production rollout.

## Commitment

| Task | Points | Status | Owner | Outcome |
| --- | ---: | --- | --- | --- |
| [TASK-0181](../04-tasks/TASK-0181.md) Shared-server four-portal deployment | 8 | backlog (held) | devops | Deployment preparation verified; awaiting operator rollout and production secrets |
| **Total** | **8** | | | |

> note: Deployment preparation completed on 2026-07-26. Image build,
> compose validation, dependency-backed container smoke, and four-path route
> checks passed; live rollout remains operator-controlled. Sprint 49 is held
> on 2026-07-28 while the owner works on another task.

## Design pack

| Artifact | Status | Scope |
| --- | --- | --- |
| Deep spec | `approved` | [DES-0045](../02-design/45-harvest-course-shared-server-deployment-spec.md) |
| Workflow | `approved` | [02-workflow.md](../02-design/02-workflow.md) §115 |
| ER / deployment inventory | `approved` | [03-er-diagram.md](../02-design/03-er-diagram.md) Sprint 49 |
| API | `approved` | [04-api-spec.md](../02-design/04-api-spec.md) Sprint 49 |
| UX / operator surface | `approved` | [05-ux-ui.md](../02-design/05-ux-ui.md) Sprint 49 |

Implementation remains scoped to TASK-0181 and is paused while Sprint 49 is
held; no product behavior or production rollout is implied.

## Scope boundary

### In

- Include the product-web build in the Monti deployment Dockerfile and compose
  environment.
- Add a Monti-side deployment wrapper that passes this checkout as the build
  source to `harvest-deployment`.
- Verify all four portal paths and `/healthz` through the shared-server route.
- Document required DNS, TLS, environment, Nginx, deploy, and rollback steps.

### Out

- New domains, DNS records, TLS certificate issuance, or live production
  rollout.
- Splitting the portals into independent services.
- Changes to customer, tenant, admin, or product application behavior.

## Verification target

```bash
make product-web
make build
./scripts/deploy-shared-server.sh --help
docker compose -f prod/docker-compose.yml -f prod/docker-compose-monti.yml --env-file prod/.env.prod config
curl -fsS https://monti.devclub.dev/healthz
curl -fsS https://monti.devclub.dev/
curl -fsS https://monti.devclub.dev/admin/
curl -fsS https://monti.devclub.dev/tenant/
curl -fsS https://monti.devclub.dev/product/
nginx -t
git diff --check
```

## Release gate

Deployment preparation is complete when the image contains all four build
outputs, the shared Nginx configuration passes validation, and the four portal
routes plus API health respond on the target host. A rollout still requires
operator approval and real production secrets.
