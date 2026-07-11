<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import RulesForm from '$lib/components/RulesForm.svelte';
  import {
    archivePackage,
    getPackage,
    listRuleSchemas,
    updatePackage,
    type Package,
    type RuleSchema
  } from '$lib/api/packages';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  const id = $derived($page.params.id);

  let pkg = $state<Package | null>(null);
  let schemas = $state<RuleSchema[]>([]);
  let fields = $state<RuleSchema['fields']>({});
  let rules = $state<Record<string, boolean | number>>({});
  let saving = $state(false);
  let showArchive = $state(false);
  let archiving = $state(false);

  onMount(async () => {
    try {
      const [p, schemaRes] = await Promise.all([getPackage(id), listRuleSchemas()]);
      pkg = p;
      schemas = schemaRes.schemas.filter((s) => s.status === 'active');
      rules = { ...p.rules };
      const schema = schemas.find((s) => s.id === p.rules_schema_id) ?? schemas[0];
      if (schema) fields = schema.fields;
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load package');
    }
  });

  async function save(e: Event) {
    e.preventDefault();
    if (!pkg) return;
    saving = true;
    try {
      pkg = await updatePackage(pkg.id, {
        slug: pkg.slug,
        name: pkg.name,
        description: pkg.description,
        status: pkg.status,
        price_cents: pkg.price_cents,
        currency: pkg.currency,
        billing_period: pkg.billing_period,
        rules_schema_id: pkg.rules_schema_id,
        rules
      });
      rules = { ...pkg.rules };
      feedback.success('Package saved');
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Save failed');
    } finally {
      saving = false;
    }
  }

  async function confirmArchive() {
    if (!pkg) return;
    archiving = true;
    try {
      await archivePackage(pkg.id);
      goto(`${base}/packages`);
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Archive failed');
      showArchive = false;
    } finally {
      archiving = false;
    }
  }
</script>

<p><a class="link" href="{base}/packages">← Packages</a></p>

{#if pkg}
  <h1 style="margin:8px 0 20px;font-size:24px">Edit: {pkg.name}</h1>
{:else}
  <h1 style="margin:8px 0 20px;font-size:24px">Edit package</h1>
{/if}

{#if pkg}
  <form onsubmit={save}>
    <div class="card" style="margin-bottom:16px">
      <h2 style="margin:0 0 16px;font-size:16px">Metadata</h2>
      <div class="field">
        <label for="slug">Slug</label>
        <input id="slug" bind:value={pkg.slug} required />
      </div>
      <div class="field">
        <label for="name">Name</label>
        <input id="name" bind:value={pkg.name} required />
      </div>
      <div class="field">
        <label for="description">Description</label>
        <textarea id="description" rows="3" bind:value={pkg.description}></textarea>
      </div>
      <div class="field">
        <label for="status">Status</label>
        <select id="status" bind:value={pkg.status}>
          <option value="draft">draft</option>
          <option value="active">active</option>
          <option value="archived">archived</option>
        </select>
      </div>
      <div style="display:grid;grid-template-columns:1fr 1fr 1fr;gap:12px">
        <div class="field">
          <label for="price">Price (¢)</label>
          <input id="price" type="number" min="0" bind:value={pkg.price_cents} />
        </div>
        <div class="field">
          <label for="currency">Currency</label>
          <select id="currency" bind:value={pkg.currency}>
            <option value="USD">USD</option>
            <option value="THB">THB</option>
            <option value="JPY">JPY (Yen)</option>
            <option value="KRW">KRW (Won)</option>
            <option value="CNY">CNY (Yuan)</option>
          </select>
        </div>
        <div class="field">
          <label for="billing">Billing</label>
          <select id="billing" bind:value={pkg.billing_period}>
            <option value="monthly">monthly</option>
            <option value="annual">annual</option>
            <option value="one_time">one_time</option>
          </select>
        </div>
      </div>
    </div>

    <div class="card" style="margin-bottom:16px">
      <h2 style="margin:0 0 16px;font-size:16px">Rules — schema {pkg.rules_schema_id}</h2>
      <RulesForm {fields} bind:rules />
    </div>

    <div style="display:flex;gap:10px;justify-content:space-between;flex-wrap:wrap">
      <button class="btn danger" type="button" disabled={pkg.status === 'archived'} onclick={() => (showArchive = true)}>
        Archive package
      </button>
      <button class="btn" type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save changes'}</button>
    </div>
  </form>
{/if}

{#if showArchive}
  <div class="modal-backdrop" role="presentation" onclick={() => (showArchive = false)}>
    <div class="card modal" role="dialog" onclick={(e) => e.stopPropagation()}>
      <h3 style="margin:0 0 12px">Archive "{pkg?.name}"?</h3>
      <p style="color:var(--muted);font-size:14px">Active entitlements block archive (409).</p>
      <div style="display:flex;gap:10px;justify-content:flex-end;margin-top:16px">
        <button class="btn ghost" type="button" onclick={() => (showArchive = false)}>Cancel</button>
        <button class="btn danger" type="button" disabled={archiving} onclick={confirmArchive}>
          {archiving ? 'Archiving…' : 'Archive'}
        </button>
      </div>
    </div>
  </div>
{/if}