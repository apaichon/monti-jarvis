---
id: SPRINT-000
status: completed
start: 2026-07-07
end: 2026-07-07
updated: 2026-07-07
goal: "Go spike for Sprint 21 — AI workforce selection + conversation (pre-Svelte/LiveKit)."
roadmap_sprint: 21
release: v0.1.0
---

# SPRINT-000 — Prototype v0.1.0

## Goal

Validate AI workforce selection and inbound Q&A before the official 35-sprint roadmap. Shipped as embedded Go/HTML UI with direct Gemini relay.

## Roadmap alignment

| Official sprint | Feature | Prototype coverage |
| ---: | --- | --- |
| 21 | Select AI Workforce to Conversation | Full (4 agents, text + voice) |
| 1 | Conversation | Partial (voice/text, no LiveKit/NATS/Svelte) |
| 2 | KM and Scope | None |
| 3 | Auth | None |

## Shipped

- `internal/workforce/` — Ava, Max, Luna, Neo
- Text chat + Gemini Live voice relay
- Dark neon call-center UI with Monti logo
- Isolated infra (`monti_jarvis` DB, Redis 4, MinIO)

## Retire in official product

- Embedded HTML UI → **Svelte + shadcn-svelte**
- Direct `/ws/voice` Gemini relay → **LiveKit room + orchestrator**
- No NATS events, no ClickHouse, no multi-tenant auth

## Verification

- `go test ./...` — pass
- Tag `v0.1.0` pushed