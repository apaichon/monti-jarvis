<script lang="ts">
  import type { Agent } from '$lib/api/workforce';

  let {
    agent,
    mini = false
  }: {
    agent: Agent;
    mini?: boolean;
  } = $props();

  const src = $derived(agent.image || `/images/${agent.id}.jpg`);
  const style = $derived(`--assistant-color:${agent.color}`);
</script>

<div class="portrait photo {mini ? 'mini' : ''}" style={style}>
  <img {src} alt={agent.name} loading="lazy" />
</div>