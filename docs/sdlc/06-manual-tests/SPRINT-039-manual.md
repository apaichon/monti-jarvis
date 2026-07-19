---
id: MANUAL-SPRINT-039
sprint: SPRINT-039
release_target: v2.15.0
feature: FEAT-0035
updated: 2026-07-19
status: ready
---

# SPRINT-039 Manual UAT — Theme Branding & Colors (v2.15.0)

**Feature:** [FEAT-0035](../01-features/FEAT-0035-theme-color-customization.md)  
**Sprint:** [SPRINT-039](../03-sprints/SPRINT-039.md)  
**Tasks:** [TASK-0149](../04-tasks/TASK-0149.md) · [TASK-0150](../04-tasks/TASK-0150.md) · [TASK-0151](../04-tasks/TASK-0151.md) · [TASK-0152](../04-tasks/TASK-0152.md)  
**Spec:** [DES-0037](../02-design/37-theme-color-customization-spec.md)  
**Owner:** Tester  

Mark each box **[x]** pass / **[ ]** fail + defect note. Sign off when **must** scenarios pass.

**UI reference:** Caller/embed chrome — header **logo · brand name · subtitle**, **Start call**, chat, **Send**.

---

## 0. Preconditions

| Check | Done |
| --- | --- |
| Branch/worktree has S39 code (`feature/sprint-039-theme-colors` or merged main) | [ ] |
| Postgres + Redis (+ MinIO for logo upload) up | [ ] |
| `make build` / `make restart` (Go + customer + tenant portals) | [ ] |
| `AUTH_DISABLED=false` + `JWT_SECRET` for tenant login | [ ] |
| Active tenant seed (e.g. `admin@demo.local` / `demo-admin`) | [ ] |
| Tools: browser, `curl`, `jq`; optional second browser/profile for tenant isolation | [ ] |
| Optional: `GEMINI_API_KEY` for live chat/voice (not required for theme chrome) | [ ] |

**Recommended env** (`infra/.env.dev`):

```bash
AUTH_DISABLED=false
JWT_SECRET=your-long-random-secret-at-least-32-characters
EMBED_ALLOW_EMPTY_ORIGINS=true
APP_PUBLIC_URL=http://localhost:8091
```

**Worktree note** (if not on main):

```bash
cd .worktrees/SPRINT-039
# ensure this tree’s binary/portals are what you hit on :8091
```

---

## 1. Init infrastructure

```bash
make infra-init   # if needed
make build
make restart
make infra-check
```

| Step | Expected | Result |
| --- | --- | --- |
| 1.1 `curl -fsS http://localhost:8091/healthz \| jq .ok` | `true` | [ ] |
| 1.2 Theme unit tests | `go test ./internal/store/ -count=1 -run 'Theme\|Contrast\|Hex\|Branding\|Preset\|CSS'` green | [ ] |
| 1.3 Server package tests | `go test ./cmd/server/ -count=1` green | [ ] |

---

## 2. Auth tokens

```bash
export MONTI=http://localhost:8091

TENANT=$(curl -sS -X POST "$MONTI/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@demo.local","password":"demo-admin"}' | jq -r .access_token)

echo "tenant token prefix: ${TENANT:0:16}…"

# Optional second tenant (isolation) — use real credentials if available
# TENANT_B=$(curl -sS -X POST "$MONTI/api/auth/login" ... | jq -r .access_token)
```

| Step | Expected | Result |
| --- | --- | --- |
| 2.1 Tenant A login | Non-empty access token | [ ] |
| 2.2 `GET /api/tenant/theme` without Bearer | **401** | [ ] |
| 2.3 `GET /api/tenant/theme` with Bearer | **200** JSON with `draft_tokens`, `draft_branding`, `contrast_report` | [ ] |

```bash
curl -sS -H "Authorization: Bearer $TENANT" "$MONTI/api/tenant/theme" | jq '{preset, brand:.draft_branding.brand_name, contrast:.contrast_report.ok}'
```

---

## 3. Scenarios

### S1 — Tenant Theme UI loads (TASK-0151 · FEAT AC 1)

**Browser**

1. Open `http://localhost:8091/tenant/login` → sign in as active tenant admin.  
2. Open **Theme** in the left nav (or `http://localhost:8091/tenant/theme`).

| Step | Expected | Result |
| --- | --- | --- |
| S1.1 Page loads | Brand identity + color pickers + live preview visible | [ ] |
| S1.2 Nav | **Theme** link present between Embed and Knowledge (or adjacent) | [ ] |
| S1.3 Defaults | Dark preset / draft tokens populated; preview shows Start call + Send | [ ] |

---

### S2 — Brand name, logo, subtitle draft + publish (TASK-0149/0151 · FEAT AC 1–2)

**Browser**

1. Brand name: `Libra Tech Co.,Ltd`  
2. Subtitle: `AI · text & voice`  
3. Upload a small PNG/JPEG logo (or leave logo empty for default Monti mark).  
4. **Save draft** → toast success.  
5. **Publish** → toast success.

**API check**

```bash
curl -sS -H "Authorization: Bearer $TENANT" "$MONTI/api/tenant/theme" | jq '{
  published_at,
  branding: .published_branding,
  source_tokens: (.published_tokens|keys)
}'
```

| Step | Expected | Result |
| --- | --- | --- |
| S2.1 Save draft | Draft branding reflects name/subtitle/logo | [ ] |
| S2.2 Publish | `published_at` set; `published_branding.brand_name` = Libra… | [ ] |
| S2.3 Live preview | Header preview shows logo + brand name + subtitle before leave page | [ ] |

---

### S3 — Full color theme on embed chrome (TASK-0150 · FEAT AC 3)

**Browser**

1. On Theme page, set distinctive colors (e.g. primary `#ff5500`, accent `#aa00ff`, surface `#1a1008`).  
2. Save draft → Publish (confirm low contrast if prompted).  
3. Ensure embed enabled: `/tenant/embed` → Enabled on, allowlist includes your host (or empty + `EMBED_ALLOW_EMPTY_ORIGINS=true`).  
4. Copy embed key. Open:

```text
http://localhost:8091/embed?key=emb_…&parent_origin=http://localhost:8091
```

| Step | Expected | Result |
| --- | --- | --- |
| S3.1 Header brand | Logo + **Libra Tech Co.,Ltd** + subtitle (not hard-coded Monti-only when published) | [ ] |
| S3.2 Start call | Button uses published primary/accent gradient | [ ] |
| S3.3 Send | Send button uses published primary colors | [ ] |
| S3.4 Panels | Background/surface/borders shift from default Monti blues toward published tokens | [ ] |

**API**

```bash
TENANT_ID=$(curl -sS -H "Authorization: Bearer $TENANT" "$MONTI/api/tenant/theme" | jq -r .tenant_id)
curl -sS "$MONTI/api/public/theme/$TENANT_ID" | jq '{source, branding, primary:.tokens.primary}'
```

| Step | Expected | Result |
| --- | --- | --- |
| S3.5 Public theme | `source` = `published` (after publish); tokens.primary matches | [ ] |
| S3.6 Embed resolve includes theme | `GET /api/public/embed/{key}?parent_origin=…` has `.theme.branding` and `.theme.tokens` | [ ] |

```bash
KEY=$(curl -sS -H "Authorization: Bearer $TENANT" "$MONTI/api/tenant/embed" | jq -r .embed_key)
curl -sS "$MONTI/api/public/embed/$KEY?parent_origin=http://localhost:8091" | jq '{name, theme_brand:.theme.branding.brand_name, primary:.theme.tokens.primary}'
```

---

### S4 — Customer desk branding (TASK-0150 · FEAT AC 2)

**Browser**

```text
http://localhost:8091/?tenant_id={tenant_id}
```

| Step | Expected | Result |
| --- | --- | --- |
| S4.1 Header | Brand logo/name/subtitle from published theme (not only “MONTI” hard-code) | [ ] |
| S4.2 Theme CSS | Page chrome picks up published colors when tenant_id is present | [ ] |

---

### S5 — Fallbacks when no custom publish (TASK-0150 · FEAT AC 4)

**API / second empty tenant or reset**

```bash
# Reset draft colors only (does not unpublish — use a fresh tenant OR note behavior)
curl -sS -X POST -H "Authorization: Bearer $TENANT" -H 'Content-Type: application/json' \
  -d '{"preset":"dark","reset_branding":false}' \
  "$MONTI/api/tenant/theme/reset" | jq .preset
```

| Step | Expected | Result |
| --- | --- | --- |
| S5.1 System default public | For tenant with empty `published_tokens`, public theme returns `source: system_default` and dark defaults | [ ] |
| S5.2 Empty branding fields | Fallback: workspace/Monti name, subtitle `AI · text & voice`, logo `/images/monti-logo.png` | [ ] |

---

### S6 — Contrast soft-gate (TASK-0149 · FEAT AC 5)

1. Set `text` and `surface` to near-identical dark colors (e.g. both `#111111`).  
2. Save draft — contrast report should show failures.  
3. Publish without confirm → expect **409** `contrast_confirmation_required`.  
4. Publish with confirm (UI dialog or API `confirm_low_contrast: true`) → **200**.

```bash
curl -sS -X PUT -H "Authorization: Bearer $TENANT" -H 'Content-Type: application/json' \
  -d "$(curl -sS -H "Authorization: Bearer $TENANT" "$MONTI/api/tenant/theme" | jq '{
    preset:.preset,
    branding:.draft_branding,
    tokens:(.draft_tokens + {text:"#111111", surface:"#121212", background:"#0a0a0a", muted:"#222222"})
  }')" "$MONTI/api/tenant/theme" | jq .contrast_report.ok

curl -sS -o /tmp/pub.json -w '%{http_code}' -X POST -H "Authorization: Bearer $TENANT" \
  -H 'Content-Type: application/json' -d '{}' "$MONTI/api/tenant/theme/publish"
echo
cat /tmp/pub.json | jq '{code, contrast_report}'

curl -sS -X POST -H "Authorization: Bearer $TENANT" -H 'Content-Type: application/json' \
  -d '{"confirm_low_contrast":true}' "$MONTI/api/tenant/theme/publish" | jq '{published_at, contrast:.contrast_report.ok}'
```

| Step | Expected | Result |
| --- | --- | --- |
| S6.1 Low contrast draft | `contrast_report.ok` = false | [ ] |
| S6.2 Publish without confirm | **409** + code `contrast_confirmation_required` | [ ] |
| S6.3 Publish with confirm | **200**, theme published | [ ] |

---

### S7 — Tenant isolation (TASK-0150 · FEAT AC 6)

| Step | Expected | Result |
| --- | --- | --- |
| S7.1 Tenant B cannot use Tenant A token for B’s id | Cross-tenant write denied (JWT tenant_id scoped) | [ ] |
| S7.2 Public theme for tenant B | Does not show Tenant A’s brand name/logo/colors | [ ] |
| S7.3 Embed key of A | Resolve never returns B’s theme | [ ] |

---

### S8 — Reset draft (TASK-0149 · FEAT AC 1)

**Browser:** Theme → **Reset colors** → confirm.

| Step | Expected | Result |
| --- | --- | --- |
| S8.1 Draft tokens | Restored to selected preset defaults | [ ] |
| S8.2 Published unchanged until publish | Public/embed still shows last published until re-publish | [ ] |

---

### S9 — Logo upload asset (TASK-0149)

```bash
# After browser upload, logo_url should be like /api/assets/theme/{tenant_id}/logo.png
LOGO=$(curl -sS -H "Authorization: Bearer $TENANT" "$MONTI/api/tenant/theme" | jq -r .draft_branding.logo_url)
echo "$LOGO"
curl -sS -o /dev/null -w '%{http_code}' "$MONTI$LOGO"
```

| Step | Expected | Result |
| --- | --- | --- |
| S9.1 Upload via UI | logo_url set on draft | [ ] |
| S9.2 Asset GET | **200** image when MinIO configured | [ ] |
| S9.3 MinIO down (optional) | Clear error on upload; no crash | [ ] |

---

### S10 — Platform admin read-only summary (TASK-0152 · FEAT AC support)

```bash
# platform admin login — use seed platform credentials for your env
PLATFORM=$(curl -sS -X POST "$MONTI/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"platform@monti.local","password":"…"}' | jq -r .access_token)

curl -sS -H "Authorization: Bearer $PLATFORM" \
  "$MONTI/api/admin/tenants/$TENANT_ID/theme" | jq .
```

| Step | Expected | Result |
| --- | --- | --- |
| S10.1 Read summary | brand_name, has_logo, preset, published_at, contrast_ok | [ ] |
| S10.2 No write API for platform on theme | Platform cannot PUT tenant theme (no route / 403) | [ ] |

---

### S11 — Auth / inactive (FEAT AC 6–7)

| Step | Expected | Result |
| --- | --- | --- |
| S11.1 No token | Theme APIs **401** | [ ] |
| S11.2 Non-admin role (if available) | **403** | [ ] |
| S11.3 Public never returns draft-only fields | Public payload has no `draft_*` keys | [ ] |

```bash
curl -sS "$MONTI/api/public/theme/$TENANT_ID" | jq 'keys'
# expect: tenant_id, preset, source, branding, tokens, css_vars — not draft_tokens
```

---

### S12 — Regression: vanilla embed loader (FEAT AC 7)

| Step | Expected | Result |
| --- | --- | --- |
| S12.1 `monti-embed.js` loads | **200** | [ ] |
| S12.2 Fixture / host page with script tag | Widget opens; iframe shows branded theme when published | [ ] |
| S12.3 Framework SDK props (if used) | `theme` prop still non-breaking | [ ] |

```bash
curl -fsS -o /dev/null -w '%{http_code}\n' "$MONTI/embed/monti-embed.js"
```

---

## 4. Defect log

| ID | Scenario | Severity | Notes |
| --- | --- | --- | --- |
| | | | |

---

## 5. Teardown

```bash
# optional: reset theme to dark for demo tenant
curl -sS -X POST -H "Authorization: Bearer $TENANT" -H 'Content-Type: application/json' \
  -d '{"preset":"dark","reset_branding":true}' "$MONTI/api/tenant/theme/reset"
# re-publish if you want demo clean:
# curl -sS -X POST -H "Authorization: Bearer $TENANT" -d '{"confirm_low_contrast":true}' ...
# make down   # only if tearing down shared infra
```

| Step | Result |
| --- | --- |
| Demo tenant left in known state (documented) | [ ] |

---

## 6. Sign-off

| Field | Value |
| --- | --- |
| Tester | |
| Date | |
| Build / commit | |
| Result | PASS / FAIL |
| Blocking defects | |

**Must-pass scenarios:** S1, S2, S3, S6, S7, S11.  
**Should-pass:** S4, S5, S8, S9, S10, S12.

---

## See also

- Integrator: [EMBED_WEB_INTEGRATION.md](../../EMBED_WEB_INTEGRATION.md)  
- Design: [37-theme-color-customization-spec.md](../02-design/37-theme-color-customization-spec.md)  
- Sprint: [SPRINT-039.md](../03-sprints/SPRINT-039.md)
