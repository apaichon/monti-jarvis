<script lang="ts">
  import type { Agent } from '$lib/api/workforce';

  let {
    agent,
    mini = false,
    lipSync = false,
    lipLevel = 0
  }: {
    agent: Agent;
    mini?: boolean;
    lipSync?: boolean;
    lipLevel?: number;
  } = $props();

  const src = $derived(agent.image || `/images/${agent.id}.jpg`);
  const style = $derived(`--assistant-color:${agent.color}`);
  const mouthOpen = $derived(
    lipSync ? Math.min(1, Math.max(0, lipLevel * 14)) : 0
  );
  const talking = $derived(lipSync && mouthOpen > 0.04);
</script>

<div
  class="portrait photo {mini ? 'mini' : ''}"
  class:talking
  class:robot={agent.robot}
  style="{style};--mouth-open:{mouthOpen}"
>
  <img {src} alt={agent.name} loading="lazy" />
  {#if lipSync}
    <div class="lip-sync" aria-hidden="true">
      <span class="lip-mouth"></span>
    </div>
  {/if}
</div>

<style>
  .portrait.talking {
    animation: none;
  }

  .portrait.talking img {
    transform-origin: 50% 72%;
    transition: transform 45ms ease-out;
    transform: scaleY(calc(1 + var(--mouth-open) * 0.018));
  }

  .lip-sync {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }

  .lip-mouth {
    position: absolute;
    left: 36%;
    right: 36%;
    bottom: 21%;
    height: 7%;
    border-radius: 999px;
    background: rgb(28 14 18 / 72%);
    box-shadow: inset 0 1px 0 rgb(255 255 255 / 12%);
    transform: scaleY(calc(0.25 + var(--mouth-open) * 1.55));
    transform-origin: center;
    transition: transform 45ms ease-out;
    opacity: calc(0.35 + var(--mouth-open) * 0.65);
  }

  .portrait.robot .lip-mouth {
    left: 40%;
    right: 40%;
    bottom: 24%;
    height: 5%;
    background: rgb(0 168 255 / 55%);
    box-shadow: 0 0 8px rgb(0 168 255 / 45%);
  }
</style>