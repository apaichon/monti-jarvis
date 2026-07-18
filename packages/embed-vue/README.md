# @monti/embed-vue

Vue 3 component + plugin for the Monti Jarvis web embed.

## Install

```bash
npm install @monti/embed-vue @monti/embed-core
```

## Usage

```vue
<script setup lang="ts">
import { MontiEmbedVue } from "@monti/embed-vue";
</script>

<template>
  <MontiEmbedVue
    embed-key="emb_YOUR_KEY"
    api-base="http://localhost:8091"
    position="bottom-right"
    @open="() => {}"
    @close="() => {}"
    @error="(e) => console.error(e)"
  />
</template>
```

### Plugin

```ts
import { createApp } from "vue";
import { createMontiEmbedPlugin } from "@monti/embed-vue";
import App from "./App.vue";

createApp(App).use(createMontiEmbedPlugin({ apiBase: "http://localhost:8091" })).mount("#app");
```

### Props

Same as core: `embedKey`, `apiBase`, `parentOrigin`, `position`, `agentId`, `theme`, `locale`, `open`, `inline`, `skipResolve`.

Events: `open`, `close`, `ready`, `error`, `destroy`, `update:open`.
