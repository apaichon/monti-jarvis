# @monti/embed-svelte

Svelte 4/5 component for the Monti Jarvis web embed (aligned with the Monti portal stack).

## Install

```bash
npm install @monti/embed-svelte @monti/embed-core
```

## Component (Svelte 5 runes)

```svelte
<script>
  import MontiEmbed from "@monti/embed-svelte/MontiEmbed.svelte";
</script>

<MontiEmbed embedKey="emb_YOUR_KEY" apiBase="http://localhost:8091" />
```

## Imperative mount

```ts
import { mountMontiEmbedSvelte } from "@monti/embed-svelte";

const handle = await mountMontiEmbedSvelte({
  embedKey: "emb_YOUR_KEY",
  apiBase: "http://localhost:8091",
});
// handle.destroy() on teardown
```
