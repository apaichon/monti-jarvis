<script lang="ts">
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';

  onMount(() => {
    goto(`${base}/register`);
  });
</script>

<p style="text-align:center;color:var(--muted);padding:48px">Redirecting…</p>