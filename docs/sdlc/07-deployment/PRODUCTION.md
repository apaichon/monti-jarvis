---
id: DEPLOY-PROD
status: draft
updated: 2026-07-07
environment: production
---

# Production Deployment (Stub)

Monti Jarvis is **not production-hardened** in Sprint 1–2. This document outlines the target shape from the [blueprint](../../monti_multi_tenant_ai_call_center_blueprint.md) for when Sprint 3+ auth and tenant isolation land.

## Target topology (future)

```text
                    ┌─────────────────┐
  Callers ─────────►│ CDN / WAF       │
                    └────────┬────────┘
                             │
                    ┌────────▼────────┐
                    │ Ingress (TLS)   │
                    └────────┬────────┘
                             │
              ┌──────────────┼──────────────┐
              │              │              │
     ┌────────▼────────┐ ┌───▼───┐ ┌────────▼────────┐
     │ monti-jarvis    │ │LiveKit│ │ Tenant admin    │
     │ API + portal    │ │ SFU   │ │ (later sprint)  │
     └────────┬────────┘ └───────┘ └─────────────────┘
              │
    ┌─────────┼─────────┬─────────────┐
    │         │         │             │
 Postgres  Redis     MinIO      ClickHouse
 (per-tenant   (sessions)  (recordings,  (KM vectors,
  RLS)                    km/)         analytics)
```

## Not yet implemented

- TLS termination and public DNS
- Auth / KYC (Sprint 3)
- Per-tenant RLS and secrets management
- Horizontal pod autoscaling
- Managed Gemini / LiveKit credentials rotation
- CI/CD pipeline and container image publish
- Backup/restore runbooks for KM and call recordings

## Pre-production checklist (when ready)

Use [`../08-readiness/RELEASE-READINESS.md`](../08-readiness/RELEASE-READINESS.md) as the gate. Additional prod items:

- [ ] Secrets in vault (not `.env` files on disk)
- [ ] Postgres backups + PITR
- [ ] MinIO versioning / lifecycle for `calls/` and `km/`
- [ ] ClickHouse replication or managed service SLA
- [ ] Rate limits on `/api/chat` and KM ingest
- [ ] Observability: structured logs, metrics, tracing
- [ ] DR drill documented

## Suggested release artifact (future)

```bash
# Illustrative — not wired in Makefile yet
docker build -t monti-jarvis:$(git describe --tags) .
# Deploy via Helm/K8s or VM systemd unit with:
#   - PORT=443 behind reverse proxy
#   - DATABASE_URL, REDIS_URL, MINIO_*, CLICKHOUSE_*, GEMINI_API_KEY from secrets
```

Until auth ships, **do not expose** the no-auth Sprint 1–2 build to the public internet.