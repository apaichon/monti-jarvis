<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import {
    listBillingDocuments,
    openPlatformDocumentHTML,
    reissueDocument,
    voidDocument,
    type BillingDocument
  } from '$lib/api/billing';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  let docs = $state<BillingDocument[]>([]);
  let docType = $state('');
  let status = $state('issued');
  let loading = $state(true);
  let busyId = $state<string | null>(null);

  async function load() {
    loading = true;
    try {
      const res = await listBillingDocuments({
        doc_type: docType || undefined,
        status: status || undefined
      });
      docs = res.documents ?? [];
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load documents');
    } finally {
      loading = false;
    }
  }

  onMount(load);

  async function doVoid(d: BillingDocument) {
    const reason = prompt('Void reason?', 'voided by platform admin');
    if (reason === null) return;
    busyId = d.id;
    try {
      await voidDocument(d.id, reason);
      feedback.success('Document voided');
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Void failed');
    } finally {
      busyId = null;
    }
  }

  async function doReissue(d: BillingDocument) {
    busyId = d.id;
    try {
      await reissueDocument(d.id);
      feedback.success('Document reissued');
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Reissue failed');
    } finally {
      busyId = null;
    }
  }

  async function openDoc(id: string) {
    try {
      await openPlatformDocumentHTML(id);
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Open failed');
    }
  }
</script>

<div style="display:flex;justify-content:space-between;align-items:center;gap:12px;margin-bottom:20px;flex-wrap:wrap">
  <div>
    <h1 style="margin:0;font-size:24px">Receipts & tax invoices</h1>
    <p style="margin:6px 0 0;color:var(--muted);font-size:13px">Void / reissue / preview (Sprint 11)</p>
  </div>
  <div style="display:flex;gap:10px;align-items:center;flex-wrap:wrap">
    <a class="btn ghost" href="{base}/billing">← Ledger</a>
    <a class="btn ghost" href="{base}/billing/seller">Seller branding</a>
    <label style="font-size:13px;color:var(--muted)">
      Type
      <select bind:value={docType} onchange={load} style="margin-left:8px">
        <option value="">all</option>
        <option value="receipt">receipt</option>
        <option value="tax_invoice">tax_invoice</option>
      </select>
    </label>
    <label style="font-size:13px;color:var(--muted)">
      Status
      <select bind:value={status} onchange={load} style="margin-left:8px">
        <option value="">all</option>
        <option value="issued">issued</option>
        <option value="voided">voided</option>
      </select>
    </label>
  </div>
</div>

<div class="card">
  {#if loading}
    <p style="color:var(--muted)">Loading…</p>
  {:else if !docs.length}
    <p style="color:var(--muted);margin:0">No documents.</p>
  {:else}
    <table class="table">
      <thead>
        <tr>
          <th>Issued</th>
          <th>Number</th>
          <th>Type</th>
          <th>Buyer</th>
          <th>Package</th>
          <th>Status</th>
          <th>Actions</th>
        </tr>
      </thead>
      <tbody>
        {#each docs as d (d.id)}
          <tr>
            <td style="font-size:12px;color:var(--muted)">{new Date(d.issued_at).toLocaleString()}</td>
            <td>{d.doc_number}</td>
            <td>{d.doc_type}</td>
            <td>
              {d.buyer_name}
              {#if d.buyer_tax_id}<div style="font-size:11px;color:var(--muted)">{d.buyer_tax_id}</div>{/if}
            </td>
            <td>{d.package_name}</td>
            <td>{d.status}</td>
            <td style="white-space:nowrap">
              <button class="btn ghost" type="button" onclick={() => openDoc(d.id)}>View</button>
              {#if d.status === 'issued'}
                <button class="btn ghost" type="button" disabled={busyId === d.id} onclick={() => doVoid(d)}>
                  Void
                </button>
                <button class="btn ghost" type="button" disabled={busyId === d.id} onclick={() => doReissue(d)}>
                  Reissue
                </button>
              {/if}
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

<style>
  .table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
  }
  .table th,
  .table td {
    text-align: left;
    padding: 8px 6px;
    border-bottom: 1px solid var(--border, #2a3344);
    vertical-align: top;
  }
  .table th {
    color: var(--muted);
    font-weight: 500;
    font-size: 12px;
  }
</style>
