---
id: DEPLOY-SPRINT-044-CUSTOMER-GENERATIVE
status: backlog
hold_reason: security_review_required
updated: 2026-07-25
environment: local-dev-and-production
sprint: SPRINT-044
---

# Sprint 44 Customer Generative AI Setup (ON HOLD)

> **ON HOLD:** Do not configure or expose customer/tenant generative or CLI
> execution. This document is retained only for a future reviewed rollout.

## Required server configuration

```dotenv
# Stable deployment key; base64-encoded 32-byte value.
TENANT_SECRET_ENCRYPTION_KEY=<base64-encoded-32-byte-value>
TENANT_SECRET_KEY_VERSION=v1

# Claude-compatible Messages API. Keep provider credentials server-side.
CLAUDE_BASE_URL=https://api.anthropic.com
CLAUDE_MODEL=claude-sonnet-4-20250514
GENERATIVE_TIMEOUT=90s
RATE_LIMIT_GENERATION_PER_MIN=10
```

The encryption key must remain stable across restarts. It is used to encrypt
customer-owned Claude API keys; it is not the Claude API key itself. The
customer browser receives only masked connection metadata.

## Local verification

```bash
make infra-up
make infra-check
make db-migrate
make build
make test
cd apps/customer-web && npm run check && npm run build
```

For deterministic UAT, set `CLAUDE_BASE_URL` to a local Claude-compatible test
double that returns a bounded Messages API response. Do not use unrestricted
shell or CLI execution as a provider adapter.

## Storage and isolation

- Postgres metadata: `callcenter.customer_generation_*`.
- MinIO bytes: `km/generative/{tenant}/{customer}/{job}/{artifact}/` in bucket
  `monti-jarvis`.
- Redis rate bucket: DB 4 with the `monti_jarvis:` prefix.

Run [SPRINT-044-manual.md](../06-manual-tests/SPRINT-044-manual.md) after the
dependencies are available. The UAT gate must verify credential absence,
duplicate-job idempotency, rate limiting, and cross-tenant/customer reads.
