<script lang="ts">
  import { cn } from '$lib/utils';
  import type { Snippet } from 'svelte';

  let {
    open = false,
    onClose,
    class: className = '',
    children
  }: {
    open?: boolean;
    onClose?: () => void;
    class?: string;
    children: Snippet;
  } = $props();

  let dialogEl = $state<HTMLDialogElement | null>(null);

  $effect(() => {
    if (!dialogEl) return;
    if (open && !dialogEl.open) {
      dialogEl.showModal();
    } else if (!open && dialogEl.open) {
      dialogEl.close();
    }
  });

  function onDialogClose() {
    onClose?.();
  }

  function onBackdropClick(e: MouseEvent) {
    if (e.target === dialogEl) {
      onClose?.();
    }
  }
</script>

<dialog
  bind:this={dialogEl}
  class={cn('ui-dialog', className)}
  onclose={onDialogClose}
  onclick={onBackdropClick}
>
  {@render children()}
</dialog>