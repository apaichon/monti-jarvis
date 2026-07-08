<script lang="ts">
  import type { Agent } from '$lib/api/workforce';
  import { lipPresetFor, mouthOpenFromLevel } from '$lib/lipsync/presets';

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
  const preset = $derived(lipPresetFor(agent.id, agent.robot));
  const mouthOpen = $derived(
    lipSync && !mini ? mouthOpenFromLevel(lipLevel, preset) : 0
  );
  const talking = $derived(mouthOpen > 0.035);
  const useJawSync = $derived(lipSync && !mini);

  const jawStyle = $derived(
    [
      `--mouth-line:${(preset.mouthLine * 100).toFixed(1)}%`,
      `--mouth-open:${mouthOpen.toFixed(3)}`,
      `--jaw-drop:${(preset.jawDrop * 100).toFixed(1)}%`,
      `--jaw-scale:${preset.jawScale}`,
      `--mouth-width:${1 + mouthOpen * preset.mouthWidth}`,
      `--assistant-color:${agent.color}`
    ].join(';')
  );
</script>

<div
  class="portrait photo {mini ? 'mini' : ''}"
  class:talking
  class:robot={agent.robot}
  style={jawStyle}
  role="img"
  aria-label={agent.name}
>
  {#if useJawSync}
    <div class="face-stack" aria-hidden="true">
      <div class="face-layer face-upper" style:background-image="url({src})"></div>
      <div class="face-layer face-lower" style:background-image="url({src})"></div>
    </div>
  {:else}
    <img {src} alt={agent.name} loading="lazy" />
  {/if}
</div>

<style>
  .portrait {
    overflow: hidden;
  }

  .portrait.talking {
    animation: none;
  }

  .face-stack {
    position: absolute;
    inset: 0;
    border-radius: inherit;
    overflow: hidden;
  }

  .face-layer {
    position: absolute;
    inset: 0;
    background-size: cover;
    background-position: center top;
    background-repeat: no-repeat;
    will-change: transform;
  }

  /* Same photo — upper face fixed at the mouth seam */
  .face-upper {
    z-index: 2;
    clip-path: inset(0 0 calc(100% - var(--mouth-line)) 0);
  }

  /* Lower jaw: opens using the real pixels below the mouth line */
  .face-lower {
    z-index: 1;
    clip-path: inset(var(--mouth-line) 0 0 0);
    transform-origin: 50% 0%;
    transform: translateY(calc(var(--mouth-open) * var(--jaw-drop)))
      scale(var(--mouth-width), calc(1 + var(--mouth-open) * var(--jaw-scale)));
    transition: transform 40ms ease-out;
  }

  .portrait.robot .face-lower {
    filter: brightness(calc(1 + var(--mouth-open) * 0.08));
  }

  .portrait.photo img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }

</style>