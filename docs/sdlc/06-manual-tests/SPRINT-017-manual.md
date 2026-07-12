# SPRINT-017 Manual UAT — Test and Preview

**Sprint:** SPRINT-017 · **Release target:** v1.8.0 · **Feature:** FEAT-0019

## 0. Preconditions

```bash
make restart
go test ./internal/quota/ ./internal/store/ ./cmd/server/ -count=1
# Active tenant_admin JWT + KM data optional
```

- [ ] Server healthy on `:8091`
- [ ] Unit tests green

## S1 — Scenarios + preview chat

```bash
TOKEN=…
curl -sS -H "Authorization: Bearer $TOKEN" http://localhost:8091/api/tenant/preview/scenarios | jq .
curl -sS -X POST -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"agent_id":"ava","topic":"general","message":"Hello"}' \
  http://localhost:8091/api/tenant/preview/chat | jq .
```

- [ ] Scenarios list non-empty
- [ ] Chat returns `mode: preview`, `reply`, `tenant_id`
- [ ] Unauthorized without token → 401

## S2 — Minutes charged (package)

```bash
# Note Redis: monti_jarvis:quota:{tenant}:minutes:{YYYYMM}
# Run preview voice >30s then hang up
# Monthly + call_daily keys increase (same as production)
```

- [ ] Preview voice increments monthly package minutes
- [ ] Preview voice increments S16 `call_daily` when caps configured
- [ ] Session rows tagged `source=preview` when created

## S3 — Rate limit still applies

- [ ] Bursting chat beyond rate limit returns 429 `rate_limited` (if rate limits enabled)

## S4 — UI `/tenant/preview`

```bash
open http://localhost:8091/tenant/preview
```

- [ ] Nav **Preview** visible
- [ ] Banner states preview **uses package** rate limits & call minutes
- [ ] **Avatar portrait** visible (like embed) for selected agent
- [ ] Scenario chips fill input
- [ ] Send returns assistant reply
- [ ] Embed: if enabled → Open live embed; if not → link to `/tenant/embed`

## S5 — Isolation

- [ ] Tenant A preview uses A’s KM (not B)
- [ ] Non-admin token cannot call preview chat

## S6 — Concurrent preview voice (optional)

- [ ] Opening > `PREVIEW_MAX_CONCURRENT` (default 2) voice sessions returns `preview_concurrent`

## Sign-off

| Role | Name | Date | Pass |
| --- | --- | --- | --- |
| Tester | | | [ ] |
| Dev | | | [ ] |
