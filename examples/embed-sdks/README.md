# Embed SDK smoke demos (FEAT-0017)

Static HTML demos that load built packages via import maps / esm.sh-style paths after local build.

## Prerequisites

1. Monti server running (`make restart` → `http://localhost:8091`)
2. Tenant embed enabled with key `emb_…` and allowlist including `http://localhost:5500`
3. Build packages:

```bash
cd packages/embed-core && npm i && npm run build
cd ../embed-web-component && npm i && npm run build
```

## Serve

```bash
# from repo root
npx --yes serve examples/embed-sdks -p 5500
```

Open:

| Demo | URL |
| --- | --- |
| Web Component | http://localhost:5500/web-component.html |
| Core (vanilla TS API) | http://localhost:5500/core.html |
| Framework snippet reference | http://localhost:5500/framework-snippets.html |

Edit `EMBED_KEY` and `API_BASE` in each HTML file (or use query `?key=emb_…&base=http://localhost:8091`).

Vue/React/Svelte full Vite apps are not required for smoke: use `framework-snippets.html` plus package unit builds. For host app integration, copy snippets from tenant admin **Embed → Framework** or [EMBED_WEB_INTEGRATION.md](../../docs/EMBED_WEB_INTEGRATION.md).
