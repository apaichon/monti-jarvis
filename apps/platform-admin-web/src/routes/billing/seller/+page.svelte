<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { getSellerBranding, putSellerBranding } from '$lib/api/billing';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  let name = $state('Monti Jarvis Platform');
  let address = $state('Bangkok, Thailand');
  let taxId = $state('');
  let branch = $state('00000');
  let loading = $state(true);
  let saving = $state(false);

  onMount(async () => {
    try {
      const b = await getSellerBranding();
      name = b.name;
      address = b.address;
      taxId = b.tax_id;
      branch = b.branch || '00000';
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load branding');
    } finally {
      loading = false;
    }
  });

  async function save(e: Event) {
    e.preventDefault();
    saving = true;
    try {
      await putSellerBranding({ name, address, tax_id: taxId, branch });
      feedback.success('Seller branding saved — new/reissued docs will use it');
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Save failed');
    } finally {
      saving = false;
    }
  }
</script>

<div style="margin-bottom:16px">
  <a href="{base}/billing" style="font-size:13px;color:var(--muted)">← Billing</a>
</div>

<div class="card" style="max-width:520px">
  <h1 style="margin:0 0 8px;font-size:22px">Seller branding</h1>
  <p style="margin:0 0 20px;color:var(--muted);font-size:13px">
    Company block printed on receipts and tax invoices (Sprint 11).
  </p>

  {#if loading}
    <p style="color:var(--muted)">Loading…</p>
  {:else}
    <form onsubmit={save}>
      <div class="field">
        <label for="name">Legal name</label>
        <input id="name" bind:value={name} required />
      </div>
      <div class="field">
        <label for="address">Address</label>
        <textarea id="address" bind:value={address} rows="3"></textarea>
      </div>
      <div class="field">
        <label for="tax">Tax ID</label>
        <input id="tax" bind:value={taxId} />
      </div>
      <div class="field">
        <label for="branch">Branch</label>
        <input id="branch" bind:value={branch} />
      </div>
      <button class="btn" type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save'}</button>
    </form>
  {/if}
</div>

<style>
  .field {
    margin-bottom: 14px;
  }
  .field label {
    display: block;
    font-size: 13px;
    color: var(--muted);
    margin-bottom: 6px;
  }
  .field input,
  .field textarea {
    width: 100%;
    box-sizing: border-box;
  }
</style>
