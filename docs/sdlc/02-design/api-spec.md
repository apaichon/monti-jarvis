---
id: DES-0004
title: API Specification
status: review_pending
updated: 2026-07-07
sprint: SPRINT-003
---

# API Specification — Monti Jarvis

**Base URL:** `http://localhost:8091`  
**Auth (Sprint 3 — draft):** `AUTH_DISABLED=true` (default) — same as v0.3.0. When `AUTH_DISABLED=false`, use `Authorization: Bearer <access_token>` on protected routes. See [auth-spec.md](auth-spec.md).  
**CORS:** `*` — methods `GET, POST, OPTIONS`; headers `Content-Type, X-Tenant-Id`, `Authorization`

> **REVIEW PENDING** — Auth section below is design-only until [auth-spec.md](auth-spec.md) is approved.

## Health & infra

### `GET /healthz`

Liveness + feature flags.

```json
{
  "ok": true,
  "gemini": true,
  "voice": true,
  "livekit": true,
  "nats": true,
  "rag": true,
  "sprint": "SPRINT-003",
  "auth_disabled": true,
  "customer_web": "apps/customer-web/build"
}
```

### `GET /api/infra`

Dependency health.

```json
{
  "postgres": "ok",
  "redis": "ok",
  "minio": "ok",
  "clickhouse": "ok",
  "nats": "ok",
  "livekit": "configured"
}
```

## Workforce

### `GET /api/workforce`

```json
{
  "agents": [
    {
      "id": "ava",
      "name": "Ava",
      "role": "General Support",
      "trait": "Warm & Patient",
      "color": "#008cff",
      "image": "/images/ava.jpg",
      "voice": "Aoede",
      "popular": true,
      "greeting": "..."
    }
  ]
}
```

## Chat (text + RAG)

### `POST /api/chat`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `session_id` | string | no | Reuse id; generated if empty |
| `agent_id` | string | yes | `ava` \| `max` \| `luna` \| `neo` |
| `topic` | string | no | `general` \| `billing` \| `technical` |
| `message` | string | yes | User question |
| `history` | array | no | `{role, content}` prior turns |

**Response 200:**

```json
{
  "session_id": "abc123",
  "agent_id": "max",
  "reply": "Invoices are due within 15 days.",
  "sources": [
    {
      "chunk_id": "...",
      "document_id": "...",
      "scope": "billing",
      "excerpt": "Invoices are issued...",
      "score": 0.82
    }
  ],
  "missing_km": false
}
```

**Errors:** `400` validation · `502` Gemini failure

## Voice

### `GET /ws/voice`

WebSocket upgrade.

| Query | Description |
| --- | --- |
| `agent` | Agent id |
| `topic` | Topic tab for RAG preload |

**Client → server messages:**

```json
{"type": "audio", "data": "<base64 PCM16>"}
{"type": "text", "text": "..."}
{"type": "end"}
```

**Server → client messages:**

```json
{"type": "ready", "agent_id": "ava", "voice": "Aoede"}
{"type": "audio", "data": "<base64>"}
{"type": "transcript", "role": "user|assistant", "text": "..."}
{"type": "text", "text": "..."}
{"type": "interrupted"}
{"type": "turn_complete"}
{"type": "error", "message": "..."}
```

Static assets: `/recorder.js`, `/player.js` (AudioWorklet).

## Call sessions

### `POST /api/calls`

Create session. **Response:** `CallSession` `{id, tenant_id, room_name, status, started_at}`

### `GET /api/calls/{id}`

Get session.

### `POST /api/calls/{id}/token`

LiveKit join token. Body: `{identity?}`. **Response:** `{token, url, identity, room_name}`

### `POST /api/calls/{id}/end`

End session.

### `GET /api/calls/{id}/turns`

List turns. **Response:** `{turns: [{id, role, content, created_at}]}`

### `POST /api/calls/{id}/turns`

Add turn. Body: `{role, content}`

### `GET /api/calls/{id}/events`

SSE stream. Event `turn` with turn JSON payload.

## Auth (Sprint 3 — draft)

### `POST /api/auth/login`

| Field | Type | Required |
| --- | --- | --- |
| `email` | string | yes |
| `password` | string | yes |

**Response 200:** `{access_token, refresh_token, expires_in, token_type, user}`

**Errors:** `401` invalid credentials · `503` auth not configured

### `POST /api/auth/refresh`

Body: `{ "refresh_token": "..." }` → new token pair.

### `POST /api/auth/logout`

Bearer access token or body `{refresh_token}` → revokes refresh.

### `GET /api/auth/me`

Bearer required → `{id, email, display_name, role, tenant_id}`.

### Protected routes (when `AUTH_DISABLED=false`)

| Route | Roles |
| --- | --- |
| `POST /api/km/agents/{id}/documents` | `tenant_admin`, `platform_admin` |
| `POST /api/km/agents/{id}/reset` | `tenant_admin`, `platform_admin` |
| `POST /api/km/seed` | `platform_admin` |

Public unchanged: `/api/chat`, `/ws/voice`, `GET /api/km/*`, `/api/workforce`.

## Knowledge base (per avatar)

### `GET /api/km/agents/{agent_id}`

KB summary: `{agent_id, tenant_id, scope, document_count, chunk_count, documents[]}`

### `GET /api/km/agents/{agent_id}/documents`

List documents for agent.

### `POST /api/km/agents/{agent_id}/documents`

Multipart: `file` (required), `scope` (optional, defaults per agent).

**Response 201:** document record `{id, status, chunk_count, ...}`

### `POST /api/km/agents/{agent_id}/reset`

Clear all KB for agent (Postgres + MinIO + ClickHouse).

**Response:** `{agent_id, status: "reset"}`

### `POST /api/km/seed`

Ingest sample files from `docs/samples/km/{agent}.md` for all four agents.

## Static routes

| Route | Handler |
| --- | --- |
| `/` | Svelte customer portal |
| `/legacy/` | Legacy HTML UI |
| `/images/*` | Static assets |

## Error envelope

```json
{"error": "human-readable message"}
```

See [auth-spec.md](auth-spec.md), [ux-ui.md](ux-ui.md), and [docs/KM_SETUP.md](../../KM_SETUP.md).