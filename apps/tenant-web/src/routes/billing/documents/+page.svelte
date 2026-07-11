<script lang="ts">
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';
  import type { PaymentDocument } from '$lib/api/billing';
  import { listBillingDocuments, openTenantDocumentHTML } from '$lib/api/tax';

  let docs = $state<PaymentDocument[]>([]);
  let loading = $state(true);

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/billing/documents`)}`);
      return;
    }
    try {
      const res = await listBillingDocuments();
      docs = res.documents ?? [];
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Failed to load documents');
    } finally {
      loading = false;
    }
  });

  async function openDoc(id: string) {
    try {
      await openTenantDocumentHTML(id);
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Open failed');
    }
  }

  function formatAmount(cents: number, currency: string): string {
    const n = (cents / 100).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
    if (currency === 'THB' || currency === '764') return `฿${n}`;
    return `${n} ${currency}`;
  }
</script>

<div class="page-wrap">
  <div style="display:flex;justify-content:space-between;align-items:center;gap:12px;flex-wrap:wrap;margin-bottom:20px">
    <div>
      <h1 style="margin:0;font-size:22px">Documents</h1>
      <p style="margin:6px 0 0;color:var(--muted);font-size:13px">Receipts and tax invoices (Sprint 12 vault)</p>
    </div>
    <div style="display:flex;gap:10px">
      <a class="btn ghost" href="{base}/billing">Packages</a>
      <a class="btn ghost" href="{base}/billing/tax">Tax profile</a>
    </div>
  </div>

  <div class="card">
    {#if loading}
      <p style="color:var(--muted)">Loading…</p>
    {:else if !docs.length}
      <p style="margin:0;color:var(--muted)">No documents yet. Complete a package purchase first.</p>
    {:else}
      <table class="table">
        <thead>
          <tr>
            <th>Issued</th>
            <th>Type</th>
            <th>Number</th>
            <th>Amount</th>
            <th>Status</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {#each docs as d (d.id)}
            <tr>
              <td style="font-size:12px;color:var(--muted)">{new Date(d.issued_at).toLocaleString()}</td>
              <td>{d.doc_type === 'tax_invoice' ? 'Tax invoice' : 'Receipt'}</td>
              <td>{d.doc_number}</td>
              <td>{formatAmount(d.amount_cents, d.currency)}</td>
              <td>{d.status}</td>
              <td>
                <button class="btn ghost" type="button" onclick={() => openDoc(d.id)}>View / print</button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>
</div>

<style>
  .page-wrap {
    max-width: 960px;
    margin: 0 auto;
    padding: 32px 20px 48px;
  }
  .table {
    width: 100%;
    border-collapse: collapse;
    font-size: 14px;
  }
  .table th,
  .table td {
    text-align: left;
    padding: 10px 8px;
    border-bottom: 1px solid var(--line, #2a3344);
  }
  .table th {
    color: var(--muted);
    font-size: 12px;
    font-weight: 500;
  }
</style>
