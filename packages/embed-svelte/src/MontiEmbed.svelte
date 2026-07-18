<script lang="ts">
  import { onMount, onDestroy, createEventDispatcher } from "svelte";
  import {
    MontiEmbed,
    type EmbedError,
    type EmbedPosition,
    type EmbedResolveResult,
  } from "@monti/embed-core";

  interface Props {
    embedKey: string;
    apiBase: string;
    parentOrigin?: string;
    position?: EmbedPosition;
    agentId?: string;
    theme?: string;
    locale?: string;
    open?: boolean;
    inline?: boolean;
    skipResolve?: boolean;
  }

  let {
    embedKey,
    apiBase,
    parentOrigin = undefined,
    position = "bottom-right",
    agentId = undefined,
    theme = undefined,
    locale = undefined,
    open = false,
    inline = false,
    skipResolve = false,
  }: Props = $props();

  const dispatch = createEventDispatcher<{
    open: void;
    close: void;
    ready: EmbedResolveResult | undefined;
    error: EmbedError;
    destroy: void;
  }>();

  let host: HTMLDivElement | undefined = $state();
  let embed: MontiEmbed | null = null;
  let mounted = false;

  async function remount() {
    embed?.destroy();
    embed = null;
    if (!embedKey || !apiBase) return;
    embed = new MontiEmbed({
      embedKey,
      apiBase,
      parentOrigin,
      position,
      agentId,
      theme,
      locale,
      open,
      skipResolve,
      container: inline ? host ?? null : null,
      onOpen: () => dispatch("open"),
      onClose: () => dispatch("close"),
      onReady: (r) => dispatch("ready", r),
      onError: (e) => dispatch("error", e),
      onDestroy: () => dispatch("destroy"),
    });
    await embed.mount();
  }

  onMount(() => {
    mounted = true;
    void remount();
  });

  onDestroy(() => {
    mounted = false;
    embed?.destroy();
    embed = null;
  });

  // Remount when identity options change (after initial mount).
  $effect(() => {
    const _deps = [
      embedKey,
      apiBase,
      parentOrigin,
      position,
      agentId,
      theme,
      locale,
      inline,
      skipResolve,
    ];
    void _deps;
    if (!mounted) return;
    void remount();
  });

  // Controlled open without remounting.
  $effect(() => {
    if (!embed || !mounted) return;
    if (open) embed.open();
    else embed.close();
  });

  export function openEmbed() {
    embed?.open();
  }
  export function closeEmbed() {
    embed?.close();
  }
  export function toggleEmbed() {
    embed?.toggle();
  }
  export function getInstance() {
    return embed;
  }
</script>

<div
  bind:this={host}
  data-monti-embed-svelte="1"
  style={inline ? "width:100%;min-height:480px;height:100%" : "display:contents"}
></div>
