---
id: DES-0020
title: Tenant Test and Preview Sandbox Specification
status: shipped
updated: 2026-07-12
sprint: SPRINT-017
owner: SA
---

# Tenant Test & Preview — Design Spec

**Sprint:** SPRINT-017 · **Release target:** v1.8.0  
**Feature:** [FEAT-0019](../01-features/FEAT-0019-tenant-test-preview.md)  
**Depends on:** conversation (S1), tenant KM (S15), settings/locale (S16), quotas (S13)

## 1. Goals

1. Tenant admin **previews** chat/voice as their own tenant before go-live.  
2. Preview uses real **RAG + locale** for that tenant.  
3. Preview **uses package quotas** (rate limits, concurrent, monthly minutes, S16 daily/per-call) so go-live testing is realistic.  
4. Sessions tagged `source=preview` for logging; UI matches customer **embed** (avatar + chat + voice).

## 2. Non-goals

| Out | Sprint |
| --- | --- |
| Separate staging DB | Future |
| Customer tiers | S18 |
| Customer accounts | S19–20 |
| Full call archive product | S22 |

## 3. Policy matrix

| Dimension | Production (embed/customer) | Preview (tenant admin) |
| --- | --- | --- |
| Rate limit chat/voice | Yes | Yes |
| Package monthly minutes | Yes | **Yes** (charged) |
| S16 daily / per-call caps | Yes | **Yes** |
| Concurrent package slots | Yes | **Yes** |
| Session `source` | `production` | `preview` (logged) |
| RAG + tenant scope | Yes | Yes |
| AI locale hint | Yes | Yes |
| Avatar embed UI | Yes | Yes |

## 4. Data model

### Option A (preferred) — `call_sessions.source`

Add nullable column (or use existing free-form field if present):

| Column | Type | Notes |
| --- | --- | --- |
| source | text | `production` \| `preview` (default `production`) |

Chat-only preview without LiveKit may skip session row or insert lightweight row with `source=preview`.

### Redis

| Key | Value | TTL |
| --- | --- | --- |
| `{prefix}preview:concurrent:{tenant}` | int | ~1h |

## 5. API summary

| Method | Path | Auth |
| --- | --- | --- |
| POST | `/api/tenant/preview/chat` | tenant_admin active |
| GET | `/ws/tenant/preview/voice` | tenant_admin active (query token or cookie not used — Bearer via first message **or** ticket query — prefer same WS auth pattern as existing if any; else JWT query `access_token` **short-lived** documented) |

**Simpler voice option:** `GET /ws/voice?preview=1&tenant_id=` with middleware that requires JWT for preview=1.

### POST chat body

```json
{
  "agent_id": "ava",
  "topic": "general",
  "message": "ราคาเท่าไหร่",
  "history": [],
  "session_id": "optional"
}
```

### Response

Same shape as public chat (`reply`, `sources`, `missing_km`, `session_id`) plus `"mode":"preview"`.

## 6. Enforcement sketch

```text
preview chat:
  AllowRate(chat)
  RAG WithTenant(jwt.tenant)
  AI locale from settings
  // no CheckMonthlyMinutes / no AddCallMinutes

preview voice open:
  AllowRate(voice)
  AcquirePreviewConcurrent (not package concurrent)
  // no CheckDaily / no CheckMonthly
  on end: release preview concurrent only
```

## 7. UX

Route **`/tenant/preview`** (T10). Banner: preview does not use package minutes.  
See [05-ux-ui.md](05-ux-ui.md).

## 8. Verification

```bash
# minutes key before
curl -sS -X POST -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"agent_id":"ava","message":"hello"}' http://localhost:8091/api/tenant/preview/chat | jq .
# minutes key after — unchanged
```

## 9. Related

| Artifact | Path |
| --- | --- |
| Feature | [FEAT-0019](../01-features/FEAT-0019-tenant-test-preview.md) |
| Sprint | [SPRINT-017](../03-sprints/SPRINT-017.md) |
| Quotas | [16-quota-rate-limit-spec.md](16-quota-rate-limit-spec.md) |
| Settings | [19-tenant-settings-limits-spec.md](19-tenant-settings-limits-spec.md) |
