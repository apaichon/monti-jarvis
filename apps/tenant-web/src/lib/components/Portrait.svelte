<script lang="ts">
  export type PortraitAgent = {
    id: string;
    name: string;
    color?: string;
    image?: string;
    speaking_image?: string;
    expressions?: Record<string, string>;
  };

  let {
    agent,
    mini = false,
    speaking = false,
    tone = ''
  }: {
    agent: PortraitAgent;
    mini?: boolean;
    speaking?: boolean;
    tone?: string;
  } = $props();

  const still = $derived(agent.image || `/images/${agent.id}.jpg`);
  const toned = $derived((tone && agent.expressions?.[tone]) || '');
  const src = $derived(toned || (speaking && agent.speaking_image ? agent.speaking_image : still));
  const style = $derived(`--assistant-color:${agent.color || '#00b7ff'}`);
</script>

<div class="portrait photo {mini ? 'mini' : ''}" style={style}>
  <img {src} alt={agent.name} loading="lazy" />
</div>

<style>
  .portrait {
    --assistant-color: #00b7ff;
    width: 168px;
    height: 168px;
    border-radius: 50%;
    overflow: hidden;
    border: 3px solid color-mix(in srgb, var(--assistant-color) 55%, transparent);
    box-shadow: 0 0 28px color-mix(in srgb, var(--assistant-color) 35%, transparent);
    background: #0a1528;
  }
  .portrait.mini {
    width: 48px;
    height: 48px;
    border-width: 2px;
  }
  .portrait img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
</style>
