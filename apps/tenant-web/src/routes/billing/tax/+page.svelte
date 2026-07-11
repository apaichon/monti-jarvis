<script lang="ts">
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';
  import { getTaxProfile, putTaxProfile } from '$lib/api/tax';

  let companyName = $state('');
  let taxId = $state('');
  let branch = $state('00000');
  let address = $state('');
  let refreshInvoices = $state(true);
  let loading = $state(true);
  let saving = $state(false);

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/billing/tax`)}`);
      return;
    }
    try {
      const p = await getTaxProfile();
      companyName = p.company_name || '';
      taxId = p.tax_id || '';
      branch = p.branch || '00000';
      address = p.address || '';
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Failed to load tax profile');
    } finally {
      loading = false;
    }
  });

  async function save(e: Event) {
    e.preventDefault();
    saving = true;
    try {
      const res = await putTaxProfile({
        company_name: companyName.trim(),
        tax_id: taxId.trim(),
        branch: branch.trim() || '00000',
        address: address.trim(),
        refresh_invoices: refreshInvoices
      });
      const n = res.invoices_refreshed ?? 0;
      feedback.success(
        n > 0 ? `Tax profile saved; reissued ${n} tax invoice(s)` : 'Tax profile saved'
      );
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Save failed');
    } finally {
      saving = false;
    }
  }
</script>

<div class="page-wrap">
  <div style="margin-bottom:16px">
    <a href="{base}/billing" style="font-size:13px;color:var(--muted)">← Billing</a>
  </div>

  <div class="card" style="max-width:520px">
    <h1 style="margin:0 0 8px;font-size:22px">Tax invoice profile</h1>
    <p style="margin:0 0 20px;color:var(--muted);font-size:13px">
      Buyer details for tax invoices (Sprint 12). Optional: reissue active tax invoices after save.
    </p>

    {#if loading}
      <p style="color:var(--muted)">Loading…</p>
    {:else}
      <form onsubmit={save}>
        <div class="field">
          <label for="company">Company name</label>
          <input id="company" bind:value={companyName} required />
        </div>
        <div class="field">
          <label for="tax">Tax ID</label>
          <input id="tax" bind:value={taxId} placeholder="0-0000-00000-00-0" />
        </div>
        <div class="field">
          <label for="branch">Branch</label>
          <input id="branch" bind:value={branch} />
        </div>
        <div class="field">
          <label for="address">Address</label>
          <textarea id="address" bind:value={address} rows="3"></textarea>
        </div>
        <label style="display:flex;align-items:center;gap:8px;margin-bottom:16px;font-size:13px">
          <input type="checkbox" bind:checked={refreshInvoices} />
          Reissue active tax invoices with these details
        </label>
        <button class="btn" type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save'}</button>
        <a class="btn ghost" href="{base}/billing/documents" style="margin-left:8px;text-decoration:none"
          >Documents</a
        >
      </form>
    {/if}
  </div>
</div>

<style>
  .page-wrap {
    max-width: 960px;
    margin: 0 auto;
    padding: 32px 20px 48px;
  }
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
