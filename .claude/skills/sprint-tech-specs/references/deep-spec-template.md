# Deep domain spec template — docs/sdlc/02-design/NN-<slug>-spec.md

```markdown
---
id: DES-NNNN
title: <Domain> Specification
status: review_pending
updated: YYYY-MM-DD
sprint: SPRINT-NNN
owner: SA
---

# <Domain> — Design Spec

**Sprint:** SPRINT-NNN · **Release target:** vX.Y.Z  
**Feature:** [FEAT-NNNN](../01-features/FEAT-NNNN-<slug>.md)  
**Depends on:** [prior-spec.md](prior-spec.md)

## 1. Goals

- …

## 2. Non-goals (Sprint N)

- …

## 3. Environment

| Variable | Default | Description |
| --- | --- | --- |
| `FEATURE_ENABLED` | `true` | … |

## 4. Data model (Postgres `callcenter`)

### `table_name`

| Column | Type | Notes |
| --- | --- | --- |
| `id` | text PK | |
| audit cols | | `created_at`, `updated_at`, `created_by`, `updated_by` |

## 5. Redis / NATS / ClickHouse (if any)

| Key / subject | Purpose |
| --- | --- |
| `monti_jarvis:…` | … |

## 6. API summary

See [04-api-spec.md](04-api-spec.md) § <Domain>. Quick list:

| Method | Path | Role |
| --- | --- | --- |
| `GET` | `/api/...` | `platform_admin` |

## 7. RBAC

| Action | `platform_admin` | `tenant_admin` | `customer` |
| --- | --- | --- | --- |
| … | yes | no | no |

## 8. Verification

```bash
make test && make build
curl ...
```

## Approver sign-off

| Role | Name | Date | Approved |
| --- | --- | --- | --- |
| PM | | | ☐ |
| Dev | | | ☐ |
```