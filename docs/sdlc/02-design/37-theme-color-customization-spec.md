---
id: DES-0037
title: Theme Branding and Color Customization Specification
status: shipped
updated: 2026-07-19
sprint: SPRINT-039
owner: SA
release_target: v2.15.0
---

# Theme Branding & Color Customization — Design Spec

**Sprint:** SPRINT-039 · **Release target:** v2.15.0  
**Feature:** [FEAT-0035](../01-features/FEAT-0035-theme-color-customization.md)  
**Depends on:** [17-embed-to-web-spec.md](17-embed-to-web-spec.md), [19-tenant-settings-limits-spec.md](19-tenant-settings-limits-spec.md)  
**UI reference:** Customer/embed caller chrome (header brand mark + name + subtitle, agent orb, Start call, chat, Send) — screenshot 2026-07-19

## 1. Goals

1. Let tenants customize the **caller-facing brand chrome** shown in customer desk and embed:
   - **Brand name** (e.g. `Libra Tech Co.,Ltd`)
   - **Logo** (header mark)
   - **Subtitle** (e.g. `AI · text & voice`)
2. Let tenants customize the **full color theme** for that surface (background, panels, borders, primary buttons, accents, text, status).
3. **Draft / publish / reset** with live preview that mirrors the screenshot layout.
4. Public resolve returns **published** branding + colors only (never draft).
5. Platform admin **read-only** summary for support.

## 2. Non-goals

| Out | Notes |
| --- | --- |
| White-label custom domain CNAME | Separate backlog |
| Changing agent portrait/name/role | Workforce/avatar product (already separate) |
| Arbitrary CSS / HTML injection | Structured fields + tokens only |
| Full design-system / Storybook product | Out |
| Mobile native theme API | Future consumer of same published payload |
| Replacing `brands` portal listing fields for S38 | May *read* logo/name as defaults; theme publish is source of truth for caller chrome |

## 3. Surface map (from screenshot)

```text
┌─ embed / caller chrome ─────────────────────────────────────┐
│ [agent ▾]          [LOGO] Brand Name                     [×] │  ← brand_name, logo_url, subtitle
│                         Subtitle                             │
│                                                              │
│                    (agent orb / portrait)                    │  ← agent product, not theme
│                    Agent name / trait                        │
│                    waveform (uses accent/primary)            │
│                                                              │
│  [00:00:00]     [ Start call  primary button ]               │  ← primary / on-primary
│  status panel (surface + line + muted text)                  │
│  chat bubbles (surface-elevated + text)                      │
│  [ Type a message… ]                    [ Send ]             │  ← input line; Send = primary
└──────────────────────────────────────────────────────────────┘
```

**Theme-owned:** logo, brand name, subtitle, all colors.  
**Not theme-owned:** agent list, portraits, transcripts, call timer logic.

## 4. Branding model

| Field | Type | Limits | UI mapping |
| --- | --- | --- | --- |
| `brand_name` | text | 1–80 chars when set; empty → fall back tenant/workspace name | Header strong title |
| `subtitle` | text | 0–120 chars | Header `.sub` under name |
| `logo_url` | text URL | https or same-origin path; max 2048; empty → default Monti mark | Header brand-mark `img` |
| `logo_alt` | text | 0–80 | `alt` on logo |

Logo upload (preferred for tenant admin):

1. `POST /api/tenant/theme/logo` multipart → MinIO `monti-jarvis` prefix `theme/{tenant_id}/logo.*`  
2. Response `logo_url` public or signed GET path served by app.  
3. PUT draft may also accept an existing absolute `logo_url` (https only).

**Fallback chain (published empty fields):**

| Field | Fallback |
| --- | --- |
| `brand_name` | `tenants.name` / workspace name / `Monti` |
| `subtitle` | `AI · text & voice` |
| `logo_url` | `/images/monti-logo.png` |

## 5. Color token model

### 5.1 Required tokens → CSS variables

| Token key | CSS variable | Screenshot role |
| --- | --- | --- |
| `primary` | `--mj-primary` | Start call, Send, focus rings |
| `primary_text` | `--mj-primary-text` | Label on primary buttons (usually `#ffffff`) |
| `accent` | `--mj-accent` | Waveform, halo, secondary highlight |
| `background` | `--mj-background` | Page / shell background |
| `surface` | `--mj-surface` | Cards, status panel, header bar |
| `surface_elevated` | `--mj-surface-elevated` | Chat bubbles, inputs |
| `text` | `--mj-text` | Primary ink |
| `muted` | `--mj-muted` | Subtitle, secondary labels |
| `line` | `--mj-line` | Borders, agent select outline |
| `success` | `--mj-success` | Positive status |
| `warn` | `--mj-warn` | Warnings (mic blocked, etc.) |
| `danger` | `--mj-danger` | Errors, End call emphasis |
| `overlay` | `--mj-overlay` | Dim scrim if needed (optional; default `rgba` derived) |

Format: `#RRGGBB` preferred; `#RGB` expanded. Reject named colors and arbitrary `url()`.

### 5.2 Presets

| Preset | Notes |
| --- | --- |
| `dark` | Default; pre-S39 Monti dark parity (screenshot baseline). |
| `light` | Light surfaces, dark text; primary/accent brand-safe. |
| `branded` | Custom tokens + branding fields are source of truth. |

### 5.3 Contrast rules

| Pair | Min ratio |
| --- | --- |
| `text` on `surface` | 4.5 |
| `text` on `background` | 4.5 |
| `primary_text` on `primary` | 4.5 |
| `muted` on `surface` | 3.0 (AA large / UI) |

**Publish:** if any pair fails → `409 contrast_confirmation_required` unless `confirm_low_contrast: true`.

## 6. Data model

### `tenant_themes`

| Column | Type | Notes |
| --- | --- | --- |
| `tenant_id` | text PK FK | → tenants |
| `preset` | text | dark \| light \| branded |
| `draft_branding` | jsonb | `{ brand_name, subtitle, logo_url, logo_alt }` |
| `published_branding` | jsonb | public snapshot |
| `draft_tokens` | jsonb | full color map |
| `published_tokens` | jsonb | public snapshot |
| `published_at` | timestamptz null | |
| `draft_updated_at` | timestamptz | |
| audit | | created_at, updated_at, created_by, updated_by |

Optional Redis: `monti_jarvis:theme:pub:{tenant_id}` TTL 60s — published payload only.

## 7. API summary

| Method | Path | Auth | Purpose |
| --- | --- | --- | --- |
| `GET` | `/api/tenant/theme` | tenant_admin | Draft + published branding/tokens + contrast |
| `PUT` | `/api/tenant/theme` | tenant_admin | Save draft branding + tokens + preset |
| `POST` | `/api/tenant/theme/publish` | tenant_admin | Publish draft |
| `POST` | `/api/tenant/theme/reset` | tenant_admin | Reset draft colors (± branding) to preset |
| `POST` | `/api/tenant/theme/logo` | tenant_admin | Upload logo → URL |
| `GET` | `/api/public/theme/{tenant_id}` | public | Published branding + tokens |
| `GET` | `/api/public/embed/{embed_key}` | public | Extend with `theme: { branding, tokens, preset }` |
| `GET` | `/api/admin/tenants/{id}/theme` | platform_admin | Read-only summary |

### PUT body (draft)

```json
{
  "preset": "branded",
  "branding": {
    "brand_name": "Libra Tech Co.,Ltd",
    "subtitle": "AI · text & voice",
    "logo_url": "https://…/theme/demo/logo.png",
    "logo_alt": "Libra Tech"
  },
  "tokens": {
    "primary": "#3b9eff",
    "primary_text": "#ffffff",
    "accent": "#8b5cf6",
    "background": "#050814",
    "surface": "#0c1425",
    "surface_elevated": "#121c30",
    "text": "#f4f7ff",
    "muted": "#8390aa",
    "line": "#3d5a80",
    "success": "#3dd68c",
    "warn": "#f0b83f",
    "danger": "#ff5c7a"
  }
}
```

### Public 200

```json
{
  "tenant_id": "demo",
  "preset": "branded",
  "source": "published",
  "branding": {
    "brand_name": "Libra Tech Co.,Ltd",
    "subtitle": "AI · text & voice",
    "logo_url": "https://…/logo.png",
    "logo_alt": "Libra Tech"
  },
  "tokens": { "primary": "#3b9eff" }
}
```

## 8. Client application

1. On customer desk + embed load, resolve published theme (embed resolve extension preferred).  
2. Set CSS variables on shell root (`:root` or `.embed-shell`).  
3. Bind header: `logo_url` → img, `brand_name` → title, `subtitle` → sub line.  
4. Primary buttons / Send use `var(--mj-primary)` and `var(--mj-primary-text)`.  
5. Waveform/halo prefer `var(--mj-accent)` with agent color as optional override only if product keeps agent-specific accent.

**Vanilla `monti-embed.js`:** outer launcher may stay default; **iframe interior** must apply branding + tokens.

## 9. Tenant admin UX

Theme editor (T20) has two columns:

1. **Brand identity** — name, subtitle, logo upload/preview  
2. **Colors** — preset + token pickers  
3. **Live preview** — miniature of screenshot chrome (header, orb placeholder, Start call, chat, Send)

## 10. RBAC

| Role | Capability |
| --- | --- |
| tenant_admin active | Full draft/publish/reset/logo |
| platform_admin | Read summary |
| public | Published branding + tokens only |
| customer | No write |

## 11. Verification

```bash
# Save branded chrome matching screenshot identity
curl -sS -X PUT -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"preset":"branded","branding":{"brand_name":"Libra Tech Co.,Ltd","subtitle":"AI · text & voice","logo_url":"https://example.com/logo.png"},"tokens":{...}}' \
  "$MONTI/api/tenant/theme"

curl -sS -X POST -H "Authorization: Bearer $TOKEN" -d '{}' \
  "$MONTI/api/tenant/theme/publish"

curl -sS "$MONTI/api/public/theme/$TENANT_ID" | jq .branding,.tokens
```

Manual: publish → open embed → header shows custom logo/name/subtitle; Start call / Send use primary; second tenant unchanged.

## 12. Infrastructure and commercial pricing

Theme branding adds negligible infrastructure load. Voice WebSocket
concurrency, database, Redis, ClickHouse, object storage, and traffic are the
capacity drivers. The following is a planning model for the current PCM16
JSON-WebSocket relay, not an SLA or a binding Contabo quote.

### 12.1 Exchange rate and assumptions

- Planning rate on 2026-07-19: **1 EUR ≈ ฿38.51** mid-market; verify the rate
  at invoicing. Source: [XE EUR/THB converter](https://www.xe.com/en-gb/currencyconverter/convert/?Amount=13&From=EUR&To=THB).
- Contabo prices may vary by region, VAT, billing term, promotion, and checkout
  availability. Source: [Contabo pricing](https://contabo.com/en/pricing/).
- “Double size” means provisioning 2× the expected peak concurrent voice
  sessions and selling only the planned capacity.
- Registered-user assumptions: 1,000 → 10 concurrent; 10,000 → 100;
  100,000 → 1,000; 1,000,000 → 5,000; 10,000,000 → 10,000.
- One continuous PCM16 session is approximately 0.684 Mbps and can consume
  about 220 GB/month of traffic. Validate actual limits with the k6 fake-
  Gemini WebSocket load test before committing capacity.

### 12.2 Production machine estimate

The shared production base is eight machines: two VDS L for PostgreSQL,
two VDS M for Redis, two VDS L for ClickHouse, one storage node, and one
operations/monitoring node. App relay nodes use VDS XL and are budgeted at
approximately 100 continuous sessions per node because traffic, not only CPU,
is limiting.

| Registered users | Peak concurrent | Reserved 2× | App nodes | Total machines | Contabo cost/month |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 1,000 | 10 | 20 | 1 VDS XL | 9 | **฿17,291** |
| 10,000 | 100 | 200 | 2 VDS XL | 10 | **฿20,464** |
| 100,000 | 1,000 | 2,000 | 20 VDS XL | 28 | **฿77,582** |
| 1,000,000 | 5,000 | 10,000 | 100 VDS XL | 108 | **฿331,440** |
| 10,000,000 | 10,000 | 20,000 | 200 VDS XL | 208 | **฿648,763** |

At one million registered users and above, upgrade database and analytics
roles to dedicated servers and use multiple regions. Ten million registered
users should not be offered on a single Contabo region using uncompressed
PCM relay; use Opus/WebRTC or a media gateway first.

### 12.3 Reserved cluster prices

Prices include a 25% operating reserve and target approximately 70% gross
margin. Six-month contracts receive 10% off; one-year prepaid contracts
receive 20% off.

| Reserved capacity | Monthly | 6-month total | 1-year total |
| --- | ---: | ---: | ---: |
| 1,000-user cluster | ฿65,900 | **฿355,860** | **฿632,640** |
| 10,000-user cluster | ฿73,900 | **฿399,060** | **฿709,440** |
| 100,000-user cluster | ฿329,000 | **฿1,776,600** | **฿3,158,400** |
| 1-million-user cluster | ฿1,390,000 | **฿7,506,000** | **฿13,344,000** |
| 10-million-user cluster | ฿2,790,000+ | **฿15,066,000+** | **฿26,784,000+** |

### 12.4 BYOK tenant packages

Gemini API usage is **not included**. Each tenant supplies its own Gemini API
credential and pays Google directly. Monti charges for platform capacity,
theme/branding, storage, support, and optional SLA only.

| Package | Concurrent calls | Monthly | 6-month total | 1-year total |
| --- | ---: | ---: | ---: | ---: |
| Starter | 10 | ฿3,490 | **฿18,846** | **฿33,504** |
| Growth | 50 | ฿9,990 | **฿53,946** | **฿95,904** |
| Business | 200 | ฿29,900 | **฿161,460** | **฿287,040** |
| Scale | 1,000 | ฿99,000 | **฿534,600** | **฿950,400** |
| Enterprise | 5,000 | ฿319,000+ | **฿1,722,600+** | **฿3,062,400+** |

Contract pricing is prepaid and non-refundable. Capacity upgrades are
prorated for the remaining term; downgrades apply at renewal. VAT, payment
fees, custom support, dedicated-region capacity, and Gemini provider charges
are excluded. Do not expose tenant Gemini keys to the browser or store them as
plaintext.

## 13. See also

- [02-workflow.md](02-workflow.md) §89–90  
- [03-er-diagram.md](03-er-diagram.md) Sprint 39  
- [04-api-spec.md](04-api-spec.md) Theme  
- [05-ux-ui.md](05-ux-ui.md) Sprint 39 T20  
- [SPRINT-039](../03-sprints/SPRINT-039.md)
