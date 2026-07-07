<script lang="ts">
  let {
    color = '#00b7ff',
    count = 42,
    mini = false
  }: {
    color?: string;
    count?: number;
    mini?: boolean;
  } = $props();

  const bars = $derived(
    Array.from({ length: count }, (_, i) => ({
      i,
      h: mini
        ? 2 + Math.round(Math.abs(Math.sin(i)) * 10)
        : 5 + Math.round(Math.abs(Math.sin(i * 0.67)) * 32)
    }))
  );
</script>

<div class="wave {mini ? 'mini' : ''}" style="--assistant-color:{color}" aria-hidden="true">
  {#each bars as bar (bar.i)}
    <span class="bar" style="--i:{bar.i};--h:{bar.h}"></span>
  {/each}
</div>