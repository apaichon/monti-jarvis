<script lang="ts">
  import '../app.css';
  import FeedbackDialog from '$lib/components/FeedbackDialog.svelte';

  let { children } = $props();
</script>

{@render children()}
<FeedbackDialog />