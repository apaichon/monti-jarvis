<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import RulesForm from '$lib/components/RulesForm.svelte';
  import { createPackage, listRuleSchemas, type RuleSchema } from '$lib/api/packages';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  let schemas = $state<RuleSchema[]>([]);
  let schemaId = $state('rules-v1');
  let fields = $state<RuleSchema['fields']>({});
  let rules = $state<Record<string, boolean | number>>({});
  let slug = $state('');
  let name = $state('');
  let description = $state('');
  let status = $state('draft');
  let priceCents = $state(0);
  let currency = $state('USD');
  let billingPeriod = $state('monthly');
  let saving = $state(false);

  onMount(async () => {
    const res = await listRuleSchemas();
    schemas = res.schemas.filter((s) => s.status === 'active');
    if (schemas[0]) {
      schemaId = schemas[0].id;
      fields = schemas[0].fields;
    }
  });

  $effect(() => {
    const schema = schemas.find((s) => s.id === schemaId);
    if (schema) fields = schema.fields;
  });

  async function submit(e: Event) {
    e.preventDefault();
    saving = true;
    try {
      const created = await createPackage({
        slug,
        name,
        description,
        status,
        price_cents: priceCents,
        currency,
        billing_period: billingPeriod,
        rules_schema_id: schemaId,
        rules
      });
      goto(`${base}/packages/${created.id}`);
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Create failed');
    } finally {
      saving = false;
    }
  }
</script>

<p><a class="link" href="{base}/packages">← Packages</a></p>
<h1 style="margin:8px 0 20px;font-size:24px">New package</h1>

<form onsubmit={submit}>
  <div class="card" style="margin-bottom:16px">
    <h2 style="margin:0 0 16px;font-size:16px">Metadata</h2>
    <div class="field">
      <label for="slug">Slug *</label>
      <input id="slug" bind:value={slug} required placeholder="my-package" />
    </div>
    <div class="field">
      <label for="name">Name *</label>
      <input id="name" bind:value={name} required />
    </div>
    <div class="field">
      <label for="description">Description</label>
      <textarea id="description" rows="3" bind:value={description}></textarea>
    </div>
    <div class="field">
      <label for="status">Status</label>
      <select id="status" bind:value={status}>
        <option value="draft">draft</option>
        <option value="active">active</option>
      </select>
    </div>
    <div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:12px">
      <div class="field">
        <label for="price">Price (¢)</label>
        <input id="price" type="number" min="0" bind:value={priceCents} />
      </div>
      <div class="field">
        <label for="currency">Currency</label>
        <select id="currency" bind:value={currency}>
          <option value="USD">USD</option>
          <option value="THB">THB</option>
          <option value="JPY">JPY (Yen)</option>
          <option value="KRW">KRW (Won)</option>
          <option value="CNY">CNY (Yuan)</option>
        </select>
      </div>
      <div class="field">
        <label for="billing">Billing</label>
        <select id="billing" bind:value={billingPeriod}>
          <option value="monthly">monthly</option>
          <option value="annual">annual</option>
          <option value="one_time">one_time</option>
        </select>
      </div>
    </div>
  </div>

  <div class="card" style="margin-bottom:16px">
    <h2 style="margin:0 0 16px;font-size:16px">Rules — schema {schemaId}</h2>
    {#if schemas.length > 1}
      <div class="field">
        <label for="schema">Schema</label>
        <select id="schema" bind:value={schemaId}>
          {#each schemas as schema (schema.id)}
            <option value={schema.id}>{schema.id} — {schema.name}</option>
          {/each}
        </select>
      </div>
    {/if}
    <RulesForm {fields} bind:rules />
  </div>

  <div style="display:flex;gap:10px;justify-content:flex-end">
    <a class="btn ghost" href="{base}/packages">Cancel</a>
    <button class="btn" type="submit" disabled={saving}>{saving ? 'Creating…' : 'Create package'}</button>
  </div>
</form>