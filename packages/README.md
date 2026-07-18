# Monti packages

| Package | npm name | Notes |
| --- | --- | --- |
| [monti-mobile-sdk](./monti-mobile-sdk) | `@monti/mobile-sdk` | Mobile call API client (SPRINT-027) |
| [embed-core](./embed-core) | `@monti/embed-core` | Shared web embed core (FEAT-0017) |
| [embed-vue](./embed-vue) | `@monti/embed-vue` | Vue 3 component / plugin |
| [embed-react](./embed-react) | `@monti/embed-react` | React component + hooks |
| [embed-svelte](./embed-svelte) | `@monti/embed-svelte` | Svelte component |
| [embed-web-component](./embed-web-component) | `@monti/embed-web-component` | `<monti-embed>` custom element |

## Build embed SDKs

```bash
cd packages/embed-core && npm install && npm run build && npm test
cd ../embed-vue && npm install && npm run build
cd ../embed-react && npm install && npm run build
cd ../embed-svelte && npm install && npm run build
cd ../embed-web-component && npm install && npm run build
```

Smoke demos: [examples/embed-sdks](../examples/embed-sdks).
Integrator guide: [docs/EMBED_WEB_INTEGRATION.md](../docs/EMBED_WEB_INTEGRATION.md).
