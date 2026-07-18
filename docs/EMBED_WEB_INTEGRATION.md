# Embed Monti on Your Website

How to integrate the **Monti Jarvis web embed** (floating chat widget) into a third-party site.

**Sprint:** SPRINT-014 · **Design:** [sdlc/02-design/17-embed-to-web-spec.md](sdlc/02-design/17-embed-to-web-spec.md) · **UAT:** [sdlc/06-manual-tests/SPRINT-014-manual.md](sdlc/06-manual-tests/SPRINT-014-manual.md)

---

## 1. What you get

| Piece | URL / asset |
| --- | --- |
| Loader script | `{MONTI_HOST}/embed/monti-embed.js` |
| Chat UI (iframe) | `{MONTI_HOST}/embed?key={EMBED_KEY}&parent_origin={HOST_ORIGIN}` |
| Public config API | `GET {MONTI_HOST}/api/public/embed/{EMBED_KEY}?parent_origin=…` |
| Tenant admin UI | `{MONTI_HOST}/tenant/embed` |

Visitors see a floating button (default bottom-right). Click opens an iframe with **agent portrait**, **voice call**, and **text chat** for **your tenant** (same experience as the Monti caller desk, compact).

```text
Your website (https://shop.example)
  └── monti-embed.js  (script tag)
        └── iframe → Monti /embed?key=emb_…&parent_origin=https://shop.example
              └── GET /api/public/embed/emb_…?parent_origin=…  (allowlist check)
              └── POST /api/chat  + X-Tenant-Id
              └── WS /ws/voice?tenant_id=…  (optional voice)
```

---

## 2. Prerequisites

1. **Active tenant** (registered + KYC approved).
2. Monti server reachable from the browser at a public or local URL  
   (dev default: `http://localhost:8091` — preferred for voice; see [§7.3](#73-https--secure-context-voice--microphone)).
3. Tenant admin login: `/tenant/login`.
4. Optional: Gemini API key on the server for real AI replies and voice.

---

## 3. Tenant setup (one-time)

### 3.1 Open Embed settings

1. Sign in: `{MONTI_HOST}/tenant/login`
2. Open **Embed** in the nav, or go to `{MONTI_HOST}/tenant/embed`

### 3.2 Configure

| Field | Recommendation |
| --- | --- |
| **Enabled** | On for production traffic |
| **Embed key** | Public token (`emb_…`). Treat as a **capability id**, not a password. **Rotate** if leaked or staff leave. |
| **Allowed origins** | One origin per line — only sites that may host the widget |
| **Default agent** | Optional workforce id (e.g. `ava`) |

**Origins format:** `scheme://host[:port]` only — no path.

| Example | Valid? |
| --- | --- |
| `https://shop.example` | Yes |
| `http://localhost:5173` | Yes (local SPA) |
| `http://localhost:5500` | Yes (local fixture) |
| `https://shop.example/page` | No (path not allowed) |
| `shop.example` | No (missing scheme) |
| `http://monti-jarvis-dev.local:8091` | Valid syntax, but **do not** put Monti host here — allowlist is the **host site**, not Monti |

**Empty allowlist:** any origin is accepted when server env `EMBED_ALLOW_EMPTY_ORIGINS=true` (default in dev). **Do not leave empty in production** — set explicit origins.

### 3.3 Copy the snippet

Click **Copy snippet**. Example (local):

```html
<script
  src="http://localhost:8091/embed/monti-embed.js"
  data-embed-key="emb_REPLACE_WITH_YOUR_KEY"
  data-position="bottom-right"
  async
></script>
```

Production: replace the host with your Monti base URL (same value as `APP_PUBLIC_URL` on the server), preferably **HTTPS**.

### 3.4 API alternative (automation)

```bash
# Login as tenant_admin
TOKEN=$(curl -sS -X POST "$MONTI_HOST/api/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@your-tenant.local","password":"…"}' | jq -r .access_token)

# Get or create config + snippet
curl -sS -H "Authorization: Bearer $TOKEN" "$MONTI_HOST/api/tenant/embed" | jq .

# Enable + allowlist (production example)
curl -sS -X PUT -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{
    "enabled": true,
    "allowed_origins": ["https://www.yoursite.com", "https://yoursite.com"],
    "default_agent_id": "ava"
  }' \
  "$MONTI_HOST/api/tenant/embed" | jq .

# Rotate key (invalidates old snippets)
curl -sS -X POST -H "Authorization: Bearer $TOKEN" \
  "$MONTI_HOST/api/tenant/embed/rotate-key" | jq .embed_key
```

---

## 4. Website integration

### 4.1 Standard HTML

Paste **before** `</body>` on every page that should show the widget:

```html
<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <title>Your site</title>
  </head>
  <body>
    <!-- your content -->

    <script
      src="https://monti.example.com/embed/monti-embed.js"
      data-embed-key="emb_YOUR_KEY"
      data-position="bottom-right"
      async
    ></script>
  </body>
</html>
```

### 4.2 Script attributes

| Attribute | Required | Description |
| --- | --- | --- |
| `src` | Yes | Absolute URL to `monti-embed.js` on Monti host |
| `data-embed-key` | Yes | Key from tenant Embed settings |
| `data-position` | No | `bottom-right` (default), `bottom-left`, `top-right`, `top-left` |
| `data-base` | No | Override Monti origin if script is proxied from another host |
| `async` | Recommended | Non-blocking load |

### 4.3 Framework SDKs (Vue · React · Svelte · Web Component)

**Feature:** [FEAT-0017](sdlc/01-features/FEAT-0017-embed-framework-sdks.md) · packages under `packages/embed-*`

Prefer first-class packages over manual script injection. All four wrap **`@monti/embed-core`** and open the same `/embed` iframe surface (origin allowlist still enforced). The vanilla `monti-embed.js` path above remains the zero-dependency option.

| Package | npm | Host |
| --- | --- | --- |
| Core | `@monti/embed-core` | Shared resolve + lifecycle |
| Vue 3 | `@monti/embed-vue` | Composition component + plugin |
| React | `@monti/embed-react` | Component + `useMontiEmbed` |
| Svelte | `@monti/embed-svelte` | Svelte 4/5 component |
| Web Component | `@monti/embed-web-component` | `<monti-embed>` (Angular / plain HTML) |

**Common props:** `embedKey`, `apiBase` (Monti origin), optional `parentOrigin`, `position`, `agentId`, `theme`, `locale`, open/close, `destroy` on unmount.

#### Vue 3

```bash
npm install @monti/embed-vue @monti/embed-core
```

```vue
<script setup>
import { MontiEmbedVue } from '@monti/embed-vue'
</script>
<template>
  <MontiEmbedVue
    embed-key="emb_YOUR_KEY"
    api-base="https://monti.example.com"
    position="bottom-right"
    @error="(e) => console.error(e.code, e.message)"
  />
</template>
```

#### React

```bash
npm install @monti/embed-react @monti/embed-core
```

```tsx
import { MontiEmbedReact } from '@monti/embed-react'

export function MontiWidget() {
  return (
    <MontiEmbedReact
      embedKey="emb_YOUR_KEY"
      apiBase="https://monti.example.com"
      position="bottom-right"
      onError={(e) => console.error(e.code, e.message)}
    />
  )
}
```

#### Svelte

```bash
npm install @monti/embed-svelte @monti/embed-core
```

```svelte
<script>
  import MontiEmbed from '@monti/embed-svelte/MontiEmbed.svelte'
</script>
<MontiEmbed embedKey="emb_YOUR_KEY" apiBase="https://monti.example.com" />
```

#### Web Component

```bash
npm install @monti/embed-web-component @monti/embed-core
```

```html
<script type="module">
  import '@monti/embed-web-component'
</script>
<monti-embed
  embed-key="emb_YOUR_KEY"
  api-base="https://monti.example.com"
  position="bottom-right"
></monti-embed>
```

Events: `monti-open`, `monti-close`, `monti-ready`, `monti-error`, `monti-destroy`.

#### Errors (bad key / origin)

Pre-resolve surfaces clear codes: `embed_not_found`, `embed_disabled`, `origin_not_allowed` (same as public API). Framework packages emit `error` / `monti-error` with `{ code, message, status? }`.

#### Smoke demos

```bash
cd packages/embed-core && npm i && npm run build && npm test
cd ../embed-web-component && npm i && npm run build
npx --yes serve examples/embed-sdks -p 5500
# open http://localhost:5500/web-component.html?key=emb_…&base=http://localhost:8091
```

Tenant admin: **Embed → Framework SDKs** tab for copy-paste snippets per framework.

### 4.4 Single-page apps without packages (script inject)

If you cannot add npm packages, inject the loader once after mount (avoid double-inject on re-renders):

```js
// React example (useEffect once)
useEffect(() => {
  if (document.getElementById('monti-embed-script')) return;
  const s = document.createElement('script');
  s.id = 'monti-embed-script';
  s.src = 'https://monti.example.com/embed/monti-embed.js';
  s.async = true;
  s.dataset.embedKey = 'emb_YOUR_KEY';
  s.dataset.position = 'bottom-right';
  document.body.appendChild(s);
  return () => {
    // Optional cleanup: remove widget root if you remount often
    document.getElementById('monti-embed-root')?.remove();
    s.remove();
    window.__montiEmbedLoaded = false;
  };
}, []);
```

Next.js App Router: put the same logic in a client component (`'use client'`), or use `next/script` with `strategy="afterInteractive"` and `data-embed-key`. Prefer `@monti/embed-react` when possible.

### 4.5 WordPress / CMS

- **Theme footer:** Appearance → Theme File Editor → `footer.php` before `</body>`, or
- **Custom HTML block** in footer template, or
- Plugin that injects footer scripts (paste the snippet).

Ensure the published origin matches **Allowed origins** (e.g. `https://yoursite.com` vs `www`).

### 4.6 Direct iframe (advanced)

If you cannot run the loader script:

```html
<iframe
  title="Monti AI"
  src="https://monti.example.com/embed?key=emb_YOUR_KEY&parent_origin=https://www.yoursite.com"
  style="width:400px;height:680px;border:0;border-radius:16px"
  allow="microphone *; autoplay *"
  referrerpolicy="strict-origin-when-cross-origin"
></iframe>
```

You still need the key enabled and origin allowlist configured. Prefer the loader for the floating launcher UX and correct `parent_origin` wiring.

---

## 5. Local demo (developers)

```bash
# Terminal 1 — Monti
make restart

# Terminal 2 — serve fixture (so Origin is http://localhost:5500, not file://)
npx --yes serve docs/fixtures -p 5500
```

1. Login tenant → `/tenant/embed` → Enable → Save  
2. Set allowlist to `http://localhost:5500` (or leave empty with `EMBED_ALLOW_EMPTY_ORIGINS=true`)  
3. Copy key into `docs/fixtures/embed-demo.html` (`data-embed-key`)  
4. Open `http://localhost:5500/embed-demo.html` → click 💬 → chat / Start call  

For a Vue/Vite shop on port **5173**, allowlist `http://localhost:5173` and load the script from **`http://localhost:8091`** (not a custom `*.local` HTTP host if you need voice).

---

## 6. Public resolve API (for custom UIs)

```http
GET /api/public/embed/{embed_key}?parent_origin=https://www.yoursite.com
Origin: https://www.yoursite.com
```

Optional header (same meaning as query): `X-Embed-Parent-Origin: https://www.yoursite.com`

**How allowlist is evaluated**

| Source | When used |
| --- | --- |
| `parent_origin` query or `X-Embed-Parent-Origin` | Prefer when set (loader / iframe UI) |
| Browser `Origin` / `Referer` | Fallback (direct host-page `fetch` to Monti) |

The chat UI runs **inside an iframe on the Monti host**, so the browser’s native `Origin` for same-origin resolve is **Monti**, not your shop. The loader therefore passes the **host site** origin as `parent_origin`. Allowlist entries must match that host site (e.g. `https://shop.example` or `http://localhost:5173`), **not** the Monti host.

**200**

```json
{
  "tenant_id": "demo",
  "slug": "demo",
  "name": "Demo Workspace",
  "embed_key": "emb_…",
  "enabled": true,
  "default_agent_id": "ava",
  "agents": [{ "id": "ava", "name": "Ava", "role": "General Support" }]
}
```

| Status | `code` | Meaning |
| ---: | --- | --- |
| 404 | `embed_not_found` | Unknown key |
| 404 | `embed_disabled` | Tenant turned embed off |
| 403 | `origin_not_allowed` | Host origin not on allowlist |

Chat from a custom UI:

```http
POST /api/chat
Content-Type: application/json
X-Tenant-Id: {tenant_id from resolve}

{
  "agent_id": "ava",
  "topic": "general",
  "message": "Hello",
  "session_id": "",
  "history": []
}
```

Package **quotas** (SPRINT-013) apply to this `tenant_id` (rate limits, KM, concurrent voice, etc.).

---

## 7. Security guidelines

Use this section for production go-live and security review.

### 7.1 Threat model (what embed is / is not)

| Control | Strength | Notes |
| --- | --- | --- |
| Embed key (`emb_…`) | **Public capability** | Anyone with the key can attempt resolve; not a secret like an API password |
| Allowed origins | **Soft gate** | Stops casual hotlinking; `parent_origin` is client-asserted from the iframe path — do not treat as strong crypto auth |
| Tenant isolation | **Server-side** | Chat/voice use resolved `tenant_id`; KM and quotas are tenant-scoped |
| HTTPS | **Required in prod** | Protects users, cookies/session on your site, and enables voice |
| Auth on chat | **None (by design)** | Public embed is anonymous caller desk; do not put private admin data in agent greetings/KM without review |

**Do not** store payment secrets, admin JWTs, or private customer PII inside the embed key or iframe URL.

### 7.2 Production checklist

| # | Item | Action |
| --- | --- | --- |
| 1 | **Allowed origins** | Explicit HTTPS host origins only (`https://www…` and apex if both used). Include port only if non-default. |
| 2 | **Empty allowlist** | Set `EMBED_ALLOW_EMPTY_ORIGINS=false` in production; never ship empty allowlist in prod |
| 3 | **HTTPS both sides** | Your site **and** Monti (`APP_PUBLIC_URL`) on HTTPS |
| 4 | **Embed key hygiene** | Rotate after leak, employee offboarding, or unexpected usage spike; redeploy snippet |
| 5 | **Disable when unused** | Tenant Embed → **Enabled** off stops public resolve (`embed_disabled`) |
| 6 | **CSP on your site** | Allow Monti only for script / frame / connect / websocket (see below) |
| 7 | **Permissions-Policy** | Do not block microphone on the parent if voice is required |
| 8 | **Quotas** | Rely on S13 package limits so a public widget cannot exhaust unlimited capacity |
| 9 | **KM content** | Treat public embed as untrusted callers; avoid confidential SOPs in public agent knowledge |
| 10 | **Monitoring** | Watch 403 `origin_not_allowed`, 429 quota, and voice concurrent errors |

### 7.3 HTTPS & secure context (voice / microphone)

Browsers expose `navigator.mediaDevices` (mic) **only** in a [secure context](https://developer.mozilla.org/en-US/docs/Web/Security/Secure_Contexts):

| Monti iframe origin | Mic / voice |
| --- | --- |
| `https://monti.example.com` | Yes |
| `http://localhost:8091` | Yes |
| `http://127.0.0.1:8091` | Yes |
| `http://*.localhost:8091` | Yes (Chromium treats `*.localhost` as secure) |
| `http://monti-jarvis-dev.local:8091` | **No** — custom hostname over HTTP |
| `http://192.168.x.x:8091` | **No** |

**Symptoms when blocked:** `getUserMedia` undefined, “Cannot read properties of undefined (reading 'getUserMedia')”, or embed banner *Mic blocked*.

**Local fix:** point the snippet at localhost (or real HTTPS), not `http://*.local`:

```html
<script
  src="http://localhost:8091/embed/monti-embed.js"
  data-embed-key="emb_…"
  data-position="bottom-right"
  async
></script>
```

Keep **Allowed origins** as the shop origin (`http://localhost:5173`), not the Monti host.

**Loader behavior:** `monti-embed.js` sets on the iframe:

```text
allow="microphone *; autoplay *; …"
```

so the parent can **delegate** mic permission into the cross-origin Monti frame. That is necessary but **not sufficient** if the Monti URL itself is not a secure context.

**Parent Permissions-Policy:** if your site sends e.g. `Permissions-Policy: microphone=()`, the iframe cannot use the mic even with `allow`. Prefer:

```http
Permissions-Policy: microphone=(self "https://monti.example.com"), autoplay=(self "https://monti.example.com")
```

(or omit a restrictive microphone policy for pages that host the widget).

**User permission:** first **Start call** prompts for microphone on the **Monti origin**. Users must allow it; “denied” requires site settings → Microphone → Allow.

### 7.4 Origin allowlist rules

1. Match **exact** scheme + host + port (`https://www.a.com` ≠ `https://a.com` ≠ `http://www.a.com`).
2. List every deployment origin (prod, staging, preview) that will host the widget.
3. For local SPA ports, list each port you use (`5173`, `5174`, …).
4. `localhost` and `127.0.0.1` are **different** origins — add both if developers use both.
5. Do **not** allowlist the Monti host itself for “parent” — the parent is the third-party site.
6. Production: prefer HTTPS-only allowlist entries.

### 7.5 Content Security Policy (your website)

Minimum fragment so the loader and iframe work:

```http
Content-Security-Policy:
  script-src 'self' https://monti.example.com;
  frame-src https://monti.example.com;
  child-src https://monti.example.com;
  connect-src 'self' https://monti.example.com wss://monti.example.com;
  img-src 'self' https://monti.example.com data:;
  media-src 'self' https://monti.example.com blob:;
```

| Directive | Why |
| --- | --- |
| `script-src` | Loads `monti-embed.js` |
| `frame-src` / `child-src` | Chat UI iframe |
| `connect-src` | If host page ever calls Monti APIs directly; WS for custom UIs |
| `media-src` | Audio playback from agent voice |

Tighten `script-src` further if you use nonces/hashes for first-party scripts; keep Monti host allowlisted for the embed script.

### 7.6 Monti server / ops settings

| Env var | Production recommendation | Notes |
| --- | --- | --- |
| `APP_PUBLIC_URL` | `https://monti.example.com` | Baked into tenant **Copy snippet** |
| `EMBED_ALLOW_EMPTY_ORIGINS` | `false` | Empty allowlist must not mean “allow all” in prod |
| TLS termination | Required | LB or reverse proxy; avoid mixed content on HTTPS shops |
| `frame-ancestors` | Monti sets permissive framing for embed | Host CSP `frame-src` is the main lock on which sites can frame Monti from the **host** side |

Schema: `callcenter.tenant_embed_configs` (created on server start / infra-init).

### 7.7 Key rotation procedure

1. Tenant admin → **Embed** → **Rotate key** (or `POST /api/tenant/embed/rotate-key`).
2. Old `emb_…` immediately fails public resolve.
3. Update every host page / CMS / SPA env with the new key.
4. Verify one page: open widget → resolve 200 → chat or voice.
5. Optionally disable embed first if rotation must be coordinated across many sites.

### 7.8 What integrators must not do

- Do not put embed key in private backend-only env and then also ship it in HTML — it is **public by design**; still avoid logging it in third-party analytics if possible.
- Do not disable origin checks by leaving allowlist empty in production.
- Do not proxy Monti over HTTP on a custom hostname and expect voice to work.
- Do not rely on “obscure key” alone for multi-tenant isolation — isolation is `tenant_id` after resolve; protect KM content and quotas.
- Do not embed Monti on untrusted third-party domains you do not control (they could brand-spoof your support widget).

### 7.9 Data & privacy notes

- Visitors chat/voice as **anonymous public callers** unless you add your own auth layer later.
- Transcripts and call records follow Monti’s tenant data retention for the resolved tenant.
- Mic audio is processed for live voice (Gemini path); inform users via your site privacy policy if required by local law.
- Prefer minimizing PII in free-text chat prompts from the host page.

---

## 8. Troubleshooting

| Symptom | Likely cause | Fix |
| --- | --- | --- |
| No bubble | Script URL 404 / wrong host | Open `/embed/monti-embed.js` in browser; fix `src` |
| Panel opens, “Embed unavailable” | Key disabled / wrong key | Tenant Embed → Enabled + correct key |
| `origin_not_allowed` | Allowlist mismatch | Add exact host origin; include port; `localhost` vs `127.0.0.1` |
| Iframe 403 but curl with host `Origin` works | Parent origin not passed into iframe | Use current `monti-embed.js` (`parent_origin`); hard-refresh host to bust cache |
| `getUserMedia` undefined / Start call fails | Monti iframe not a **secure context** | Use `https://…` or `http://localhost:…` for Monti script `src` / `data-base` — not `http://*.local` |
| Banner “Mic blocked” | Same as above | See [§7.3](#73-https--secure-context-voice--microphone) |
| Mic permission denied | User or Permissions-Policy blocked mic | Browser site settings for Monti origin; relax parent `Permissions-Policy` |
| Close (✕) covers Send | Stale loader JS | Hard-refresh / cache-bust `monti-embed.js` (close is top-right of panel) |
| Chat fails / 502 | Gemini not configured | Set `GEMINI_API_KEY` on server |
| Chat 429 | Tenant quota / rate limit | Package limits (S13); wait or upgrade |
| Works on localhost, not prod | CORS / allowlist / mixed content | HTTPS both sides; add prod origin; no mixed HTTP Monti on HTTPS shop |
| Old snippet dead after rotate | Expected | Deploy new key everywhere |

---

## 9. Server configuration (ops)

| Env var | Default | Notes |
| --- | --- | --- |
| `APP_PUBLIC_URL` | `http://localhost:8091` | Host baked into snippet from tenant API |
| `EMBED_ALLOW_EMPTY_ORIGINS` | `true` | Empty allowlist = allow any origin (dev-friendly). Set **`false`** in production. |

Schema is created automatically: `callcenter.tenant_embed_configs` (on server start / infra-init).

---

## 10. Framework SDKs (roadmap)

Today’s path is the **vanilla script** above. Planned packages (**SPRINT-036** / [FEAT-0017](sdlc/01-features/FEAT-0017-embed-framework-sdks.md)):

| Package | Stack |
| --- | --- |
| `@monti/embed-vue` | Vue 3 |
| `@monti/embed-react` | React |
| `@monti/embed-svelte` | Svelte |
| `@monti/embed-web-component` | `<monti-embed>` custom element (any host) |
| `@monti/embed-core` | Shared resolve + iframe lifecycle |

Until those ship, use §3 script tag or mount the iframe yourself (see `poc/monti-embed` for a Vue host example). Roadmap: [ROADMAP.md](sdlc/00-roadmap/ROADMAP.md).

---

## 11. Related docs

| Doc | Purpose |
| --- | --- |
| [LOCAL-DEV.md](sdlc/07-deployment/LOCAL-DEV.md) | Run Monti locally |
| [17-embed-to-web-spec.md](sdlc/02-design/17-embed-to-web-spec.md) | Technical design |
| [SPRINT-014-manual.md](sdlc/06-manual-tests/SPRINT-014-manual.md) | Full UAT checklist |
| [embed-demo.html](fixtures/embed-demo.html) | Sample host page |
| [b-quik-tyre-embed-demo.html](fixtures/b-quik-tyre-embed-demo.html) | B-Quik tyre highlight demo host |
| [b-quik-tyre-highlight.md](samples/km/b-quik-tyre-highlight.md) | Sample KM (tyre FAQ from public highlight page) |
| [FEAT-0014](sdlc/01-features/FEAT-0014-embed-to-web.md) | Feature ACs (vanilla embed) |
| [FEAT-0017](sdlc/01-features/FEAT-0017-embed-framework-sdks.md) | Framework SDK packages (backlog) |
