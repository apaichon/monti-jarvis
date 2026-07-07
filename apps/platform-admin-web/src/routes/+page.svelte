<script lang="ts">
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { isPlatformAdmin, loginPath } from '$lib/auth/session';

  onMount(() => {
    goto(isPlatformAdmin() ? `${base}/packages` : loginPath());
  });
</script>

<p>Redirecting…</p>