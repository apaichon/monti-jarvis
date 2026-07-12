# SPRINT-018 Manual UAT — Customer Tier

**Sprint:** SPRINT-018 · **Release target:** v1.9.0 · **Feature:** FEAT-0020

## 0. Preconditions

```bash
make restart
go test ./internal/store/ ./cmd/server/ -count=1
# Active tenant_admin JWT
```

- [ ] Server healthy on `:8091`
- [ ] Unit tests green

## S1 — Create tier

```bash
TOKEN=…
curl -sS -X POST -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"VIP","slug":"vip","priority":100,"ai_reply_locale":"th","max_minutes_per_call":30}' \
  http://localhost:8091/api/tenant/tiers | jq .
curl -sS -H "Authorization: Bearer $TOKEN" http://localhost:8091/api/tenant/tiers | jq .
```

- [ ] 201 with id `tier_…`
- [ ] List includes VIP
- [ ] Duplicate slug → 409

## S2 — Groups

```bash
curl -sS -X POST -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"name":"Retail"}' http://localhost:8091/api/tenant/groups | jq .
```

- [ ] Group created
- [ ] List/delete works

## S3 — UI

```bash
open http://localhost:8091/tenant/tiers
```

- [ ] Nav **Tiers**
- [ ] Create/edit/delete tier in UI
- [ ] Groups section works
- [ ] Settings page links to Tiers

## S4 — Preview tier_id

```bash
TIER=… # from create
curl -sS -X POST -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d "{\"agent_id\":\"ava\",\"message\":\"hello\",\"tier_id\":\"$TIER\"}" \
  http://localhost:8091/api/tenant/preview/chat | jq '{mode,tier_id,tier_slug,reply:.reply[0:80]}'
```

- [ ] Response includes `tier_id` / `tier_slug`
- [ ] Preview UI tier select available when tiers exist

## S5 — Isolation

- [ ] Tenant A cannot GET/DELETE tenant B tier ids
- [ ] Unauthorized without JWT → 401

## Sign-off

| Role | Name | Date | Pass |
| --- | --- | --- | --- |
| Tester | | | [ ] |
| Dev | | | [ ] |
