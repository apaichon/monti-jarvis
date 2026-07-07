<script lang="ts">
  import { cn } from '$lib/utils';
  import type { Snippet } from 'svelte';

  let {
    class: className = '',
    children
  }: {
    class?: string;
    children: Snippet;
  } = $props();
</script>

<div
  class={cn(
    'rounded-2xl border border-monti-line/30 bg-monti-panel/70 p-6 shadow-2xl backdrop-blur',
    className
  )}
>
  {@render children()}
</div>