# 02 — Design

System design artifacts for Monti Jarvis.

| Doc | Description | Status |
| --- | --- | --- |
| [auth-spec.md](auth-spec.md) | **Sprint 3** JWT, RBAC, route policy — **review first** | `review_pending` |
| [architecture.md](architecture.md) | Layers, packages, infra topology | `review_pending` |
| [workflow.md](workflow.md) | Chat, voice, call, KM, auth sequences | `review_pending` |
| [er-diagram.md](er-diagram.md) | Postgres + ClickHouse + MinIO entities | `review_pending` |
| [api-spec.md](api-spec.md) | REST, WebSocket, SSE contract | `review_pending` |
| [ux-ui.md](ux-ui.md) | ASCII wireframes with API mapping | `review_pending` |

**Sprint 3 gate:** Approve [auth-spec.md](auth-spec.md) (incl. open questions §13) before TASK-0010 implementation.

**Stack:** Go `net/http` · SvelteKit customer portal · Postgres · Redis 8 · MinIO · ClickHouse · NATS · LiveKit · Gemini

**Verify:** [test matrix](../05-test-scenarios/TEST-MATRIX.md) · [manual tests](../06-manual-tests/) · [deploy](../07-deployment/LOCAL-DEV.md)