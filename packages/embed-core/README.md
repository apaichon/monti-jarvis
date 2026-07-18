# @monti/embed-core

Shared Monti Jarvis **web embed** core: public resolve, iframe URL builder, floating/inline widget lifecycle.

Framework packages (`@monti/embed-vue`, `@monti/embed-react`, `@monti/embed-svelte`, `@monti/embed-web-component`) wrap this module. The zero-dependency script tag path (`monti-embed.js`) remains supported.

## Install

```bash
npm install @monti/embed-core
```

## Usage

```ts
import { MontiEmbed } from "@monti/embed-core";

const embed = new MontiEmbed({
  embedKey: "emb_YOUR_KEY",
  apiBase: "http://localhost:8091",
  // parentOrigin defaults to window.location.origin
  position: "bottom-right",
  agentId: "ava", // optional
  onOpen: () => console.log("opened"),
  onClose: () => console.log("closed"),
  onError: (err) => console.error(err.code, err.message),
});

await embed.mount();
// embed.open(); embed.close(); embed.destroy();
```

### Inline mount

```ts
const host = document.getElementById("chat")!;
const embed = new MontiEmbed({
  embedKey: "emb_YOUR_KEY",
  apiBase: "https://monti.example.com",
  container: host,
});
await embed.mount();
```

### Resolve only

```ts
import { resolveEmbed } from "@monti/embed-core";

const cfg = await resolveEmbed({
  apiBase: "http://localhost:8091",
  embedKey: "emb_…",
  parentOrigin: "https://shop.example",
});
```

## Props / options

| Option | Required | Description |
| --- | --- | --- |
| `embedKey` | yes | Public key from tenant Embed settings |
| `apiBase` | yes | Monti host origin |
| `parentOrigin` | no | Host site origin for allowlist (default: `window.location.origin`) |
| `position` | no | `bottom-right` (default), `bottom-left`, `top-right`, `top-left` |
| `agentId` | no | Default agent query param |
| `theme` / `locale` | no | Forwarded as query params when set |
| `container` | no | Inline mount element |
| `open` | no | Start open (floating mode) |
| `skipResolve` | no | Skip pre-resolve (iframe still validates) |

## Security

Origin allowlist and tenant isolation are enforced by the Monti server (SPRINT-014). See [EMBED_WEB_INTEGRATION.md](../../docs/EMBED_WEB_INTEGRATION.md).
