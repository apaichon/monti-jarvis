<script lang="ts">
  let {
    color = '#00b7ff',
    count = 42,
    mini = false,
    active = false,
    level = 0
  }: {
    color?: string;
    count?: number;
    mini?: boolean;
    active?: boolean;
    level?: number;
  } = $props();

  const amp = $derived(active ? Math.min(1, level * 18) : 0);

  const bars = $derived(
    Array.from({ length: count }, (_, i) => {
      const base = mini
        ? 2 + Math.round(Math.abs(Math.sin(i)) * 10)
        : 5 + Math.round(Math.abs(Math.sin(i * 0.67)) * 32);
      const liveBoost = active ? Math.round(amp * (8 + Math.abs(Math.sin(i * 0.9)) * 24)) : 0;
      return { i, h: base + liveBoost };
    })
  );
</script>

<div
  class="wave {mini ? 'mini' : ''}"
  class:live={active && amp > 0.03}
  style="--assistant-color:{color};--live-amp:{amp}"
  aria-hidden="true"
>
  {#each bars as bar (bar.i)}
    <span class="bar" style="--i:{bar.i};--h:{bar.h}"></span>
  {/each}
</div>

<style>
  .wave.live .bar {
    animation-duration: 0.55s;
  }
</style>