---
id: MANUAL-SPRINT-014
sprint: SPRINT-014
release_target: v1.5.0
feature: FEAT-0014
updated: 2026-07-11
status: draft
---

# SPRINT-014 Manual UAT — Embed to Web (v1.5.0)

**Feature:** [FEAT-0014](../01-features/FEAT-0014-embed-to-web.md)  
**Tasks:** TASK-0062 … TASK-0066 · **Spec:** [17-embed-to-web-spec.md](../02-design/17-embed-to-web-spec.md)  
**Owner:** Tester

Mark each box **[x]** pass / **[ ]** fail + defect note. Sign off when **must** scenarios pass.

---

## 0. Preconditions

| Check | Done |
| --- | --- |
| SPRINT-014 code present (embed schema, APIs, loader, `/embed`, tenant UI) | [ ] |
| Postgres + Redis up; `make build` / portals built | [ ] |
| `AUTH_DISABLED=false` + `JWT_SECRET` for tenant login | [ ] |
| Active tenant seed: `admin@demo.local` / `demo-admin` | [ ] |
| Tools: `curl`, `jq`, browser; optional `npx serve` for fixture | [ ] |
| Optional: `GEMINI_API_KEY` for real chat replies | [ ] |

**Recommended env** (`infra/.env.dev`):

```bash
AUTH_DISABLED=false
JWT_SECRET=your-long-random-secret-at-least-32-characters
EMBED_ALLOW_EMPTY_ORIGINS=true
APP_PUBLIC_URL=http://localhost:8091
```

---

## 1. Init infrastructure

```bash
make infra-init   # if needed
make build        # customer + tenant + Go
make restart
make infra-check
```

| Step | Expected | Result |
| --- | --- | --- |
| 1.1 `curl -fsS http://localhost:8091/healthz \| jq .ok` | `true` | [ ] |
| 1.2 `curl -fsS -o /dev/null -w '%{http_code}' http://localhost:8091/embed/monti-embed.js` | **200** | [ ] |
| 1.3 Unit smoke | `go test ./internal/store/ ./cmd/server/ -count=1 -run Embed` green | [ ] |

---

## 2. Auth tokens

```bash
TENANT=$(curl -sS -X POST http://localhost:8091/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@demo.local","password":"demo-admin"}' | jq -r .access_token)

echo "tenant token prefix: ${TENANT:0:16}…"
```

| Step | Expected | Result |
| --- | --- | --- |
| 2.1 Tenant login | Non-empty access token | [ ] |
| 2.2 `GET /api/tenant/embed` without Bearer | **401** | [ ] |

---

## 3. Scenarios

### S1 — Tenant enable embed + lazy config (TASK-0065 · FEAT AC 1, 6)

**Browser**

1. Open `http://localhost:8091/tenant/login` → `admin@demo.local` / `demo-admin`
2. Nav **Embed** → `http://localhost:8091/tenant/embed`

**API**

```bash
curl -sS -H "Authorization: Bearer $TENANT" \
  http://localhost:8091/api/tenant/embed | jq .
```

| Step | Expected | Result |
| --- | --- | --- |
| 3.1.1 First GET | 200; `embed_key` starts with `emb_`; `snippet` present | [ ] |
| 3.1.2 UI loads | Enable checkbox, key, origins textarea, Copy snippet | [ ] |
| 3.1.3 Enable + Save | Toggle on → Save → success toast; `enabled: true` | [ ] |
| 3.1.4 Empty origins help | Warning that empty allowlist allows any origin in dev | [ ] |
| 3.1.5 Copy snippet | Clipboard / feedback “Snippet copied” | [ ] |

```bash
# Enable via API if preferred
curl -sS -X PUT -H "Authorization: Bearer $TENANT" -H 'Content-Type: application/json' \
  -d '{"enabled":true,"allowed_origins":[],"default_agent_id":"ava"}' \
  http://localhost:8091/api/tenant/embed | jq '{enabled,embed_key,default_agent_id}'

KEY=$(curl -sS -H "Authorization: Bearer $TENANT" \
  http://localhost:8091/api/tenant/embed | jq -r .embed_key)
echo "KEY=$KEY"
```

---

### S2 — Public resolve API (TASK-0063 · FEAT AC 2)

```bash
# Unknown key
curl -sS -o /tmp/emb404.json -w '%{http_code}' \
  http://localhost:8091/api/public/embed/emb_does_not_exist
echo; cat /tmp/emb404.json

# Good key
curl -sS "http://localhost:8091/api/public/embed/$KEY" | jq .
```

| Step | Expected | Result |
| --- | --- | --- |
| 3.2.1 Unknown key | **404**, `code: embed_not_found` | [ ] |
| 3.2.2 Enabled key | **200**: `tenant_id`, `slug`, `name`, `agents` array, `embed_key` | [ ] |
| 3.2.3 Disabled | PUT `enabled:false` → public GET → **404** `embed_disabled` | [ ] |
| 3.2.4 Re-enable | PUT `enabled:true` before continuing | [ ] |

**Disable / re-enable:**

```bash
curl -sS -X PUT -H "Authorization: Bearer $TENANT" -H 'Content-Type: application/json' \
  -d '{"enabled":false}' http://localhost:8091/api/tenant/embed | jq .enabled
curl -sS -w '\n%{http_code}\n' "http://localhost:8091/api/public/embed/$KEY"
curl -sS -X PUT -H "Authorization: Bearer $TENANT" -H 'Content-Type: application/json' \
  -d '{"enabled":true,"allowed_origins":[]}' http://localhost:8091/api/tenant/embed | jq .enabled
```

---

### S3 — Origin allowlist deny (TASK-0063 · FEAT AC 3)

```bash
curl -sS -X PUT -H "Authorization: Bearer $TENANT" -H 'Content-Type: application/json' \
  -d '{"enabled":true,"allowed_origins":["https://shop.example"]}' \
  http://localhost:8091/api/tenant/embed | jq .allowed_origins

# Allowed origin
curl -sS -H 'Origin: https://shop.example' \
  "http://localhost:8091/api/public/embed/$KEY" | jq '{tenant_id,code}'

# Denied origin
curl -sS -o /tmp/emb403.json -w '%{http_code}' \
  -H 'Origin: https://evil.example' \
  "http://localhost:8091/api/public/embed/$KEY"
echo; cat /tmp/emb403.json
```

| Step | Expected | Result |
| --- | --- | --- |
| 3.3.1 Allowed Origin | **200** | [ ] |
| 3.3.2 Evil Origin | **403**, `code: origin_not_allowed` | [ ] |
| 3.3.3 Missing Origin with allowlist | **403** (no Origin/Referer match) | [ ] |

**Restore empty allowlist for fixture tests:**

```bash
curl -sS -X PUT -H "Authorization: Bearer $TENANT" -H 'Content-Type: application/json' \
  -d '{"enabled":true,"allowed_origins":[]}' \
  http://localhost:8091/api/tenant/embed | jq .allowed_origins
```

---

### S4 — Loader + iframe chat (TASK-0064 · FEAT AC 1, 4, 7)

1. Put real key into `docs/fixtures/embed-demo.html` (`data-embed-key`).
2. Serve fixture (do **not** use `file://` if allowlist is non-empty):

```bash
# Terminal A
npx --yes serve docs/fixtures -p 5500
# Open http://localhost:5500/embed-demo.html
```

Optional allowlist for local fixture:

```bash
curl -sS -X PUT -H "Authorization: Bearer $TENANT" -H 'Content-Type: application/json' \
  -d '{"enabled":true,"allowed_origins":["http://localhost:5500"]}' \
  http://localhost:8091/api/tenant/embed | jq .
```

| Step | Expected | Result |
| --- | --- | --- |
| 3.4.1 Loader script | Network 200 for `/embed/monti-embed.js` | [ ] |
| 3.4.2 Launcher | Floating 💬 bottom-right on host page | [ ] |
| 3.4.3 Open panel | Iframe loads `/embed?key=…` | [ ] |
| 3.4.4 Agents | Agent name/select visible for tenant | [ ] |
| 3.4.5 Chat | Send message → assistant reply (or clear error if Gemini down) | [ ] |
| 3.4.6 No tenant JWT | Embed works without tenant login cookie | [ ] |

**Direct embed URL smoke:**

```bash
open "http://localhost:8091/embed?key=$KEY"
```

- [ ] Compact chat UI only (not full marketing portal chrome)

---

### S5 — Rotate key (TASK-0065 · FEAT AC 6)

```bash
OLD=$KEY
curl -sS -X POST -H "Authorization: Bearer $TENANT" \
  http://localhost:8091/api/tenant/embed/rotate-key | jq '{embed_key,enabled}'
NEW=$(curl -sS -H "Authorization: Bearer $TENANT" \
  http://localhost:8091/api/tenant/embed | jq -r .embed_key)

curl -sS -w '\nHTTP %{http_code}\n' "http://localhost:8091/api/public/embed/$OLD"
curl -sS "http://localhost:8091/api/public/embed/$NEW" | jq '{tenant_id,embed_key}'
```

| Step | Expected | Result |
| --- | --- | --- |
| 3.5.1 Rotate API/UI | New `emb_…` key returned | [ ] |
| 3.5.2 Old key | **404** `embed_not_found` | [ ] |
| 3.5.3 New key | **200** | [ ] |
| 3.5.4 UI confirm | Rotate confirm dialog; snippet updates | [ ] |

Export for later steps: `KEY=$NEW`

---

### S6 — Invalid origin format + defaults (TASK-0065)

| Step | Expected | Result |
| --- | --- | --- |
| 3.6.1 Bad origin save | Enter `not-a-url` → Save → **400** / error feedback | [ ] |
| 3.6.2 Default agent | Set `ava` → Save → public resolve `default_agent_id` matches | [ ] |

```bash
curl -sS -X PUT -H "Authorization: Bearer $TENANT" -H 'Content-Type: application/json' \
  -d '{"enabled":true,"allowed_origins":["not-a-url"]}' \
  http://localhost:8091/api/tenant/embed
# expect 400
```

---

### S7 — Quota still applies on embed path (FEAT AC 5) — optional

If S13 quotas active: exhaust KM or rate limit for tenant, then embed chat should return **429** with quota codes.

| Step | Expected | Result |
| --- | --- | --- |
| 3.7.1 | Over-limit chat from embed uses same tenant quotas | [ ] / skip |

---

### S8 — Regression (must)

| Step | Expected | Result |
| --- | --- | --- |
| 3.8.1 Customer portal `/` | Full demo UI still loads; workforce + chat | [ ] |
| 3.8.2 Tenant billing | `/tenant/billing` still loads when logged in | [ ] |
| 3.8.3 Platform admin | `/admin` login still works | [ ] |
| 3.8.4 Health | `/healthz` ok | [ ] |

```bash
go test ./internal/store/ ./cmd/server/ -count=1 -run 'Embed|Quota'
```

- [ ] Green

---

## 4. Defect log

| # | Scenario | Severity | Notes | Fixed? |
| --- | --- | --- | --- | --- |
| 1 | | P0/P1/P2 | | [ ] |
| 2 | | | | [ ] |

---

## 5. Teardown

```bash
# Optional: disable embed for demo
curl -sS -X PUT -H "Authorization: Bearer $TENANT" -H 'Content-Type: application/json' \
  -d '{"enabled":false,"allowed_origins":[]}' \
  http://localhost:8091/api/tenant/embed | jq .enabled
```

- [ ] Env left in team-default state
- [ ] Fixture key not committed with a production secret (keys are public capability tokens)

---

## 6. Sign-off

| Role | Name | Date | Result |
| --- | --- | --- | --- |
| Tester | | | ☐ Pass · ☐ Pass with notes · ☐ Fail |
| Dev (support) | | | ☐ |

**Must-pass before close:** S1, S2, S3, S4 (chat), S5, S8, unit tests.  
**Should-pass:** S6, S7.

**Links**

- Sprint: [SPRINT-014](../03-sprints/SPRINT-014.md)
- Task: [TASK-0066](../04-tasks/TASK-0066.md)
- Design: [17-embed-to-web-spec.md](../02-design/17-embed-to-web-spec.md)
- Fixture: [embed-demo.html](../../fixtures/embed-demo.html)
- Deploy: [LOCAL-DEV.md](../07-deployment/LOCAL-DEV.md)
