<script lang="ts">
  import type { Agent } from '$lib/api/workforce';

  let {
    agent,
    mini = false,
    speaking = false,
    tone = '',
    theme = 'dark'
  }: {
    agent: Agent;
    mini?: boolean;
    speaking?: boolean;
    tone?: string;
    theme?: 'dark' | 'light' | string;
  } = $props();

  const themedImage = $derived(
    theme === 'light'
      ? agent.image_light_url || agent.image_light || agent.image
      : agent.image_dark_url || agent.image_dark || agent.image
  );
  const still = $derived(themedImage || `/images/${agent.id}.jpg`);
  const toned = $derived((tone && agent.expressions?.[tone]) || '');
  const src = $derived(toned || (speaking && agent.speaking_image ? agent.speaking_image : still));
  const style = $derived(`--assistant-color:${agent.color}`);
</script>

<div class="portrait photo {mini ? 'mini' : ''}" style={style}>
  <img {src} alt={agent.name} loading="lazy" />
</div>
