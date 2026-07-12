# SPRINT-016 Manual UAT — Settings, Locale, Limits

**Sprint:** SPRINT-016 · **Release target:** v1.7.0 · **Feature:** FEAT-0016

## 0. Preconditions

```bash
make restart   # or infra up + server
# Active tenant_admin JWT (KYC approved)
go test ./internal/store/ ./internal/quota/ ./cmd/server/ -count=1
```

- [ ] Server healthy on `:8091`
- [ ] Unit tests green
- [ ] Active tenant with optional package entitlement (usage meters more useful with package)

## S1 — Settings GET/PUT

```bash
TOKEN=…
curl -sS -H "Authorization: Bearer $TOKEN" http://localhost:8091/api/tenant/settings | jq .
curl -sS -X PUT -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"locale":"th","timezone":"Asia/Bangkok","display_name":"Test Co","ai_reply_locale":"th"}' \
  http://localhost:8091/api/tenant/settings | jq .
```

- [ ] GET lazy-creates defaults (`en`, `Asia/Bangkok`)
- [ ] PUT returns saved locale `th`, display_name set
- [ ] Invalid locale `fr` → 400 `invalid_locale`
- [ ] Invalid timezone `Foo/Bar` → 400 `invalid_timezone`

## S2 — Usage snapshot

```bash
curl -sS -H "Authorization: Bearer $TOKEN" http://localhost:8091/api/tenant/usage | jq .
```

- [ ] Returns `tenant_id`, `period`, `usage` (ai_employees, monthly_call_minutes, km_documents, concurrent_calls)
- [ ] When entitled: `package` + `limits` present
- [ ] `daily_usage.call_minutes` present; `call_limits` present
- [ ] JWT tenant only (no other tenant data)

## S3 — Call limits

```bash
curl -sS -X PUT -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"max_minutes_per_call":2,"max_call_minutes_per_day":10}' \
  http://localhost:8091/api/tenant/call-limits | jq .
curl -sS -H "Authorization: Bearer $TOKEN" http://localhost:8091/api/tenant/call-limits | jq .
```

- [ ] Values saved
- [ ] Negative values → 400
- [ ] Values above package monthly ceiling are clamped (when package known)
- [ ] `0` means unset (no operational cap)

## S4 — UI `/tenant/settings`

```bash
open http://localhost:8091/tenant/settings
```

- [ ] Nav **Settings** visible when logged in
- [ ] Workspace section saves locale/timezone/display name
- [ ] Usage meters show package limits when entitled
- [ ] Call limits form saves; package ceiling note visible
- [ ] Tier/group scaffold labels save with help text
- [ ] Locale switch updates settings page EN/TH labels

## S5 — Voice daily / per-call enforce

```bash
# Set max_call_minutes_per_day=1, exhaust with a short call (>30s) or Redis:
# redis-cli INCRBY monti_jarvis:call_daily:{tenant}:{YYYYMMDD} 1
# Then open voice WS for that tenant → expect 429 daily_call_limit
```

- [ ] Daily cap hit → voice open denied with `code: daily_call_limit`
- [ ] `max_minutes_per_call=1` → session ends after ~1 minute (context timeout)
- [ ] S13 monthly quota still applies first
- [ ] Minutes added to both monthly S13 key and S16 `call_daily` key

## S6 — AI locale hint

```bash
# Set ai_reply_locale=th (or locale=th), send chat with ambiguous short message
curl -sS -X POST -H "Content-Type: application/json" -H "X-Tenant-Id: $TENANT" \
  -d '{"agent_id":"ava","message":"hello"}' http://localhost:8091/api/chat | jq .
```

- [ ] System prompt includes Thai preference (observe reply language bias or log)
- [ ] Embed uses same tenant setting (no customer login required for hint)

## S7 — Isolation

- [ ] Tenant A settings/limits do not affect tenant B
- [ ] Unauthenticated GET `/api/tenant/settings` → 401

## Sign-off

| Role | Name | Date | Pass |
| --- | --- | --- | --- |
| Tester | | | [ ] |
| Dev | | | [ ] |
