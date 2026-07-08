<script lang="ts">
  let {
    message = '',
    variant = 'success' as 'success' | 'error' | 'info'
  }: {
    message?: string;
    variant?: 'success' | 'error' | 'info';
  } = $props();
</script>

{#if message}
  <p class="status-message {variant}" role="status">{message}</p>
{/if}