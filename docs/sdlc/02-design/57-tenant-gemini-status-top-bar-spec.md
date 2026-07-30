---
id: DES-0057
title: Tenant Gemini Status Top Bar and Performance Nav Removal
status: approved
updated: 2026-07-30
sprint: SPRINT-061
owner: SA
feature: FEAT-0053
release_target: v2.34.0
---

# DES-0057 — Tenant Gemini Status Top Bar

**Sprint:** SPRINT-061 · **Release target:** v2.34.0  
**Feature:** [FEAT-0053](../01-features/FEAT-0053-tenant-gemini-status-top-bar.md)  
**Tasks:** TASK-0220–0223  
**Depends:** DES-0056 (key status), DES-0029 (tenant performance — UI retired)

## 1. Goals

1. Remove System Performance from tenant navigation and primary discovery.
2. Show compact Gemini readiness in tenant top bar.
3. Keep platform-admin system performance (S29) unchanged.

## 2. Non-goals

- New key management (S60)
- Tenant-facing Postgres/Redis/MinIO probes
- Removing S26 API without support path

## 3. Status model

| State | Meaning | Top-bar label (EN) | Action |
| --- | --- | --- | --- |
| `ready` | `gemini_key_status=valid` | Gemini ready | none / optional AI Settings |
| `key_missing` | no key / status none | Gemini key missing | → `/tenant/ai` |
| `validation_failed` | status invalid | Gemini key invalid | → `/tenant/ai` |
| `degraded` | valid but recent provider errors | Gemini degraded | → `/tenant/ai` (help) |

Prefer mapping from S60 metadata; optional soft health from last runtime error cache.

## 4. API

| Method | Path | Role | Response |
| --- | --- | --- | --- |
| GET | `/api/tenant/ai/gemini-status` | tenant_admin | `{ "state", "label", "action_href", "last_validated_at", "last4" }` |

No secrets. Optional: include in existing session bootstrap later (out of scope if separate GET is enough).

### Tenant system performance

| Path | S61 policy |
| --- | --- |
| UI `/tenant/monitoring` (or system-performance route) | Remove from nav; route may 302 → `/tenant/ai` or show retired notice |
| `GET /api/tenant/system-performance` | **Support-only** — remain available; not linked in tenant UI |

## 5. UX

Tenant top bar (right of “All systems operational” or replacing infra wording):

```
[ ● Gemini ready ]   or   [ ● Gemini key missing → ]
```

Colors: green ready, amber missing/degraded, red invalid. Click → AI Settings when not ready.

## 6. Platform preservation

`GET /api/platform/system-performance` and admin Monitoring page unchanged.

## 7. Verification

- Tenant nav lacks System Performance
- Top bar reflects ready after S60 valid test
- Platform monitoring still loads

## See also

Workflow §135–136 · API Sprint 61 · UX T61
