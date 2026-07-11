---
id: MANUAL-SPRINT-013
sprint: SPRINT-013
release_target: v1.4.0
feature: FEAT-0013
updated: 2026-07-11
status: deferred_uat
---

# SPRINT-013 Manual UAT — Quota & Rate Limit (v1.4.0)

**Feature:** [FEAT-0013](../01-features/FEAT-0013-quota-rate-limit.md)  
**Tasks:** TASK-0057 … TASK-0061 · **Spec:** [16-quota-rate-limit-spec.md](../02-design/16-quota-rate-limit-spec.md)  
**Owner:** Tester

Map every checkbox to pass **[x]** / fail **[ ]** + defect note. Sign off only when all **must** scenarios pass.

---

## 0. Preconditions

| Check | Done |
| --- | --- |
| Branch has SPRINT-013 code (quota package + enforcement + usage UI) | [ ] |
| `GEMINI_API_KEY` set in `infra/.env.dev` (for chat/voice regression) | [ ] |
| Redis DB **4** available (`REDIS_URL=redis://localhost:6379/4`) | [ ] |
| Postgres `monti_jarvis` + schema `callcenter` | [ ] |
| Tools: `curl`, `jq`, browser | [ ] |

**Recommended env for this UAT** (`infra/.env.dev`):

```bash
AUTH_DISABLED=false
JWT_SECRET=your-long-random-secret-at-least-32-characters
QUOTA_ENABLED=true
QUOTA_FAIL_OPEN=true
RATE_LIMIT_ENABLED=true
RATE_LIMIT_CHAT_PER_MIN=60
RATE_LIMIT_KM_PER_MIN=30
RATE_LIMIT_VOICE_PER_MIN=20
```

For **S4 rate-limit burst** only, temporarily set `RATE_LIMIT_CHAT_PER_MIN=3`, restart, then restore.

---

## 1. Init infrastructure

```bash
make infra-init    # or ensure schema already applied
make build
make restart
make infra-check
```

| Step | Expected | Result |
| --- | --- | --- |
| 1.1 `curl -fsS http://localhost:8091/healthz \| jq .` | `ok: true` | [X] |
| 1.2 `curl -fsS http://localhost:8091/api/infra \| jq '{quota, rate_limit, redis, entitlement_cache}'` | `quota` and `rate_limit` are `ok` (or `disabled` only if Redis/env off) | [X] |
| 1.3 Unit tests | `go test ./internal/quota/... ./cmd/server/ -count=1 -run Quota` green | [X] |

**If `quota: disabled`:** confirm `REDIS_URL` and `QUOTA_ENABLED` — stop and fix before continuing.

---

## 2. Auth tokens (reuse)

```bash
PLATFORM=$(curl -sS -X POST http://localhost:8091/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"platform@monti.local","password":"monti-platform"}' | jq -r .access_token)

TENANT=$(curl -sS -X POST http://localhost:8091/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@demo.local","password":"demo-admin"}' | jq -r .access_token)

echo "platform=${PLATFORM:0:12}… tenant=${TENANT:0:12}…"
```

| Step | Expected | Result |
| --- | --- | --- |
| 2.1 Platform login | Non-empty `access_token` | [X] |
| 2.2 Tenant admin login | Non-empty `access_token` | [X] |
| 2.3 `GET /api/platform/tenants/demo/usage` **without** Bearer | `401` | [ ] |
| 2.4 Same with `Authorization: Bearer $TENANT` | `403` | [ ] |

---

## 3. Scenarios

### S1 — Platform usage snapshot (TASK-0060 · FEAT AC 5)

```bash
curl -sS -H "Authorization: Bearer $PLATFORM" \
  http://localhost:8091/api/platform/tenants/demo/usage | jq .
```

| Step | Expected | Result |
| --- | --- | --- |
| 3.1.1 API 200 | `tenant_id: demo`, `period` like `2026-07`, `status` active or none | [ ] |
| 3.1.2 When Starter assigned | `package.slug` present; `limits.max_km_documents` etc. integers; `usage` object | [ ] |
| 3.1.3 Browser | Open `http://localhost:8091/admin/login` → Tenants → **Usage** on `demo` | [ ] |
| 3.1.4 UI bars | Shows AI employees / minutes / KM / concurrent; Refresh reloads | [ ] |
| 3.1.5 Empty package | Revoke entitlement (or unknown tenant with no plan) → UI “No active package” without crash | [ ] |

**Re-assign Starter if revoked:**

```bash
curl -sS -X POST -H "Authorization: Bearer $PLATFORM" -H 'Content-Type: application/json' \
  -d '{"package_id":"pkg-starter"}' \
  http://localhost:8091/api/platform/tenants/demo/entitlement | jq .
```

---

### S2 — KM document quota exceed (TASK-0059 · FEAT AC 1)

**Goal:** When `usage >= max_km_documents`, next ingest returns **429** `quota_exceeded`.

Starter default: `max_km_documents: 50` — for a short UAT, temporarily lower limit:

**Option A — edit package via platform UI** `/admin/packages` → Starter → set `max_km_documents` to `2` → Save.  
**Option B — API** (if PUT packages accepts rules):

```bash
# Inspect current package
curl -sS -H "Authorization: Bearer $PLATFORM" \
  http://localhost:8091/api/platform/packages/pkg-starter | jq '.limits // .rules // .'
```

Then ensure demo has ≤2 docs or reset agent KM:

```bash
# Optional reset (platform) — pick a seeded agent id e.g. ava
curl -sS -X POST -H "Authorization: Bearer $PLATFORM" \
  http://localhost:8091/api/km/agents/ava/reset | jq .
```

Upload tiny markdown files until limit:

```bash
AGENT=ava
for i in 1 2 3; do
  echo "doc $i $(date)" > /tmp/qdoc$i.md
  code=$(curl -sS -o /tmp/qresp$i.json -w '%{http_code}' \
    -H "Authorization: Bearer $PLATFORM" \
    -F "file=@/tmp/qdoc$i.md" \
    "http://localhost:8091/api/km/agents/${AGENT}/documents")
  echo "upload $i → HTTP $code $(cat /tmp/qresp$i.json)"
done
```

| Step | Expected | Result |
| --- | --- | --- |
| 3.2.1 Under limit | HTTP **201** for docs while count &lt; limit | [X] |
| 3.2.2 At/over limit | HTTP **429**, body `code: quota_exceeded`, `dimension: max_km_documents` | [ ] |
| 3.2.3 No partial ingest | Failed upload not listed in `GET /api/km/agents/{agent}/documents` | [ ] |
| 3.2.4 Usage panel | `usage.km_documents` matches successful uploads | [ ] |

Restore package limits to seed defaults after test.

---

### S3 — Concurrent voice slots (TASK-0059 · FEAT AC 2)

Starter default: `max_concurrent_calls: 2`.

Requires `GEMINI_API_KEY` and voice enabled on package.

```bash
# Lower concurrent to 1 for easier UAT (optional) via package edit, then:
# Open two browser tabs to customer portal voice with same demo tenant
# Or use websocat / browser DevTools WS
```

**Browser path:**

1. Ensure `AUTH_DISABLED=true` **or** customer path resolves `demo` entitlement (default demo tenant).
2. Open `http://localhost:8091/` → start **voice** on an agent.
3. Open a **second** browser window → start voice again (if limit=1, second should fail).
4. With limit=2: open third session → expect reject.

| Step | Expected | Result |
| --- | --- | --- |
| 3.3.1 First session | Voice connects / ready | [ ] |
| 3.3.2 Over concurrent | New session **429** (or WS error) with `quota_exceeded` / `max_concurrent_calls` | [ ] |
| 3.3.3 After close | Closing a session frees a slot; new session succeeds | [ ] |
| 3.3.4 Usage | `concurrent_calls` rises while open; drops after close (refresh usage UI) | [ ] |

**Note:** If voice fails for Gemini config, mark **blocked** and still run monthly-minutes unit coverage as already auto-tested.

---

### S4 — Rate limit burst (TASK-0059 · FEAT AC 6) — optional but recommended

```bash
# Temporary: RATE_LIMIT_CHAT_PER_MIN=3 in infra/.env.dev → make restart
for i in $(seq 1 5); do
  code=$(curl -sS -o /tmp/chat$i.json -w '%{http_code}' \
    -X POST http://localhost:8091/api/chat \
    -H 'Content-Type: application/json' \
    -d "{\"agent_id\":\"ava\",\"message\":\"ping $i\"}")
  echo "chat $i → $code $(jq -c '{code,error,dimension}' /tmp/chat$i.json 2>/dev/null)"
done
```

| Step | Expected | Result |
| --- | --- | --- |
| 3.4.1 First N requests | HTTP 200 (or 502 if Gemini down — still counts against rate) | [ ] |
| 3.4.2 Over limit | HTTP **429**, `code: rate_limited`, prefer `Retry-After` header | [ ] |
| 3.4.3 Restore env | Set `RATE_LIMIT_CHAT_PER_MIN=60`, restart | [ ] |

If Gemini returns 502 before rate limit, still verify rate path via unit tests:

```bash
go test ./internal/quota/ -run TestAllowRate -count=1
```

- [ ] Auto TestAllowRate green (counts as S4 partial if live Gemini blocked)

---

### S5 — Feature flags (TASK-0059 · FEAT AC 3)

**5a — `rag_enabled=false` (soft skip)**

1. Edit package rules: `rag_enabled: false` for Starter (or Pro on demo).
2. `POST /api/chat` with a question that would use KM.

| Step | Expected | Result |
| --- | --- | --- |
| 3.5.1 Chat still works | HTTP 200 reply (no RAG sources required) | [ ] |
| 3.5.2 Not hard-blocked | No 403 on normal chat when RAG off | [ ] |

**5b — `voice_enabled=false` (hard block)**

1. Set `voice_enabled: false` on package.
2. Open voice / `GET /ws/voice?agent=ava`.

| Step | Expected | Result |
| --- | --- | --- |
| 3.5.3 Voice denied | **403** `feature_disabled`, `dimension: voice_enabled` | [ ] |
| 3.5.4 Restore flags | Re-enable voice + rag on package for regression | [ ] |

---

### S6 — Avatar assign cap (TASK-0059 · FEAT AC)

Starter: `max_ai_employees: 2`.

```bash
# List platform avatars
curl -sS -H "Authorization: Bearer $PLATFORM" \
  http://localhost:8091/api/platform/avatars | jq '.avatars[:5] | .[] | {id,slug}'

# Count current assignments
curl -sS -H "Authorization: Bearer $PLATFORM" \
  http://localhost:8091/api/platform/tenants/demo/avatars | jq .
```

Assign until over limit (use distinct active avatar ids):

```bash
# Example — replace AVATAR_ID
curl -sS -o /tmp/asg.json -w '%{http_code}' \
  -X POST -H "Authorization: Bearer $PLATFORM" -H 'Content-Type: application/json' \
  -d '{"avatar_id":"AVATAR_ID"}' \
  http://localhost:8091/api/platform/tenants/demo/avatars
echo; cat /tmp/asg.json
```

| Step | Expected | Result |
| --- | --- | --- |
| 3.6.1 Under limit | Assign succeeds 200 | [ ] |
| 3.6.2 Over limit | **429** `quota_exceeded` `max_ai_employees` | [ ] |
| 3.6.3 Usage panel | `ai_employees` matches active assignments | [ ] |

---

### S7 — Infra + fail-open smoke (TASK-0057)

| Step | Expected | Result |
| --- | --- | --- |
| 3.7.1 `GET /api/infra` | `quota`, `rate_limit` present | [ ] |
| 3.7.2 Env docs | `grep -E 'QUOTA_\|RATE_LIMIT_' infra/.env.dev.example` shows vars | [ ] |
| 3.7.3 Optional fail-open | With Redis stopped + `QUOTA_FAIL_OPEN=true`, chat still returns (not 500 from quota) — **dev only** | [ ] |

---

### S8 — Regression (must)

| Step | Expected | Result |
| --- | --- | --- |
| 3.8.1 Platform login UI | `/admin/login` → packages list loads | [ ] |
| 3.8.2 Packages CRUD smoke | List packages 200 | [ ] |
| 3.8.3 Customer portal | `AUTH_DISABLED=true` or public `/` — chat still works under normal limits | [ ] |
| 3.8.4 Entitlement | `GET /api/entitlements/me` with tenant token still 200 | [ ] |
| 3.8.5 Billing/login (if used) | No break of tenant login / billing routes | [ ] |

```bash
# Quick auto regression
go test ./internal/quota/... ./cmd/server/ -count=1
```

- [ ] All green

---

## 4. Defect log

| # | Scenario | Severity | Notes | Fixed? |
| --- | --- | --- | --- | --- |
| 1 | | P0/P1/P2 | | [ ] |
| 2 | | | | [ ] |

---

## 5. Teardown

```bash
# Restore package limits / flags if changed
# Restore RATE_LIMIT_CHAT_PER_MIN=60
make restart
# optional: make down
```

- [ ] Env restored to team defaults
- [ ] Demo still has usable Starter (or intended package)

---

## Deferred at release (v1.4.0)

Full browser UAT (S3 concurrent voice multi-session, S4 live Gemini rate burst, S5 package flag toggles) **deferred** by PM decision 2026-07-11.  
Auto coverage: `go test ./internal/quota/...` and `./cmd/server/ -run Quota` green.

Re-run remaining boxes without blocking the tag.

## 6. Sign-off

| Role | Name | Date | Result |
| --- | --- | --- | --- |
| Tester | | | ☐ Pass · ☐ Pass with notes · ☐ Fail |
| Dev (support) | | | ☐ |

**Must-pass before sprint close:** S1, S2, S6 (or S3 if avatars constrained), S8, unit tests.  
**Should-pass:** S3 voice concurrent, S4 rate limit, S5 flags.

**Links**

- Sprint: [SPRINT-013](../03-sprints/SPRINT-013.md)
- Task: [TASK-0061](../04-tasks/TASK-0061.md)
- Design: [16-quota-rate-limit-spec.md](../02-design/16-quota-rate-limit-spec.md)
- Deploy: [LOCAL-DEV.md](../07-deployment/LOCAL-DEV.md)
