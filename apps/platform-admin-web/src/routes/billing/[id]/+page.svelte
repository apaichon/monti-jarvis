<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { page } from '$app/stores';
  import { getBillingOrder, openPlatformDocumentHTML, type BillingOrder } from '$lib/api/billing';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  let order = $state<BillingOrder | null>(null);
  let loading = $state(true);

  onMount(async () => {
    const id = $page.params.id;
    try {
      order = await getBillingOrder(id);
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load order');
    } finally {
      loading = false;
    }
  });

  async function openDoc(id: string) {
    try {
      await openPlatformDocumentHTML(id);
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Open failed');
    }
  }
</script>

<div style="margin-bottom:16px">
  <a href="{base}/billing" style="font-size:13px;color:var(--muted)">← Billing</a>
</div>

{#if loading}
  <p style="color:var(--muted)">Loading…</p>
{:else if !order}
  <p style="color:var(--muted)">Order not found.</p>
{:else}
  <div class="card" style="margin-bottom:16px">
    <h1 style="margin:0 0 12px;font-size:22px">Order {order.order_no}</h1>
    <p style="margin:0 0 4px">Status: <strong>{order.status}</strong></p>
    <p style="margin:0 0 4px">Package: {order.package_id}</p>
    <p style="margin:0 0 4px">Amount: {(order.amount_cents / 100).toFixed(2)} {order.currency}</p>
    <p style="margin:0 0 4px">Method: {order.payment_method || '—'}</p>
    <p style="margin:0;color:var(--muted);font-size:13px">Txn: {order.transaction_id || '—'}</p>
  </div>

  <div class="card">
    <h2 style="margin:0 0 12px;font-size:16px">Documents</h2>
    {#if !order.documents?.length}
      <p style="color:var(--muted);margin:0">No documents (issued on paid).</p>
    {:else}
      <ul style="margin:0;padding-left:18px">
        {#each order.documents as d (d.id)}
          <li style="margin-bottom:8px">
            {d.doc_type} · {d.doc_number} · {d.status}
            <button class="btn ghost" type="button" style="margin-left:8px" onclick={() => openDoc(d.id)}>
              View
            </button>
          </li>
        {/each}
      </ul>
    {/if}
    <p style="margin:16px 0 0">
      <a href="{base}/billing/receipts">Open receipt console →</a>
    </p>
  </div>
{/if}
