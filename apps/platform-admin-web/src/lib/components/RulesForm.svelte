<script lang="ts">
  import type { RuleFieldSpec } from '$lib/api/packages';

  let {
    fields,
    rules = $bindable({} as Record<string, boolean | number>)
  }: {
    fields: Record<string, RuleFieldSpec>;
    rules?: Record<string, boolean | number>;
  } = $props();

  $effect(() => {
    for (const [key, spec] of Object.entries(fields)) {
      if (rules[key] === undefined && spec.default !== undefined) {
        rules[key] = spec.default as boolean | number;
      }
    }
  });

  function labelFor(key: string, spec: RuleFieldSpec) {
    return spec.description || key.replaceAll('_', ' ');
  }
</script>

<div class="grid gap-3">
  {#each Object.entries(fields) as [key, spec] (key)}
    <div class="field">
      <label for={key}>{labelFor(key, spec)}</label>
      {#if spec.type === 'bool'}
        <label style="display:flex;align-items:center;gap:8px">
          <input id={key} type="checkbox" bind:checked={rules[key] as boolean} />
          <span>{rules[key] ? 'yes' : 'no'}</span>
        </label>
      {:else}
        <input
          id={key}
          type="number"
          min={spec.min ?? 0}
          max={spec.max}
          bind:value={rules[key] as number}
        />
      {/if}
    </div>
  {/each}
</div>