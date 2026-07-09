<script lang="ts">
  import { feedback } from '$lib/feedback.svelte';

  let {
    message = '',
    variant = 'success' as 'success' | 'error' | 'info'
  }: {
    message?: string;
    variant?: 'success' | 'error' | 'info';
  } = $props();

  $effect(() => {
    if (!message) return;
    feedback[variant](message);
  });
</script>