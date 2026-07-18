<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { listBillingOrders, type BillingOrder } from '$lib/api/billing';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  let orders = $state<BillingOrder[]>([]);
  let statusFilter = $state('');
  let tenantFilter = $state('');
  let loading = $state(true);

  async function load() {
    loading = true;
    try {
      const res = await listBillingOrders({
        status: statusFilter || undefined,
        tenant_id: tenantFilter.trim() || undefined
      });
      orders = res.orders ?? [];
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load billing orders');
    } finally {
      loading = false;
    }
  }

  onMount(load);

  function formatAmount(cents: number, currency?: string): string {
    const n = (cents / 100).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
    if (currency === 'THB' || currency === '764') return `฿${n}`;
    if (currency === 'USD') return `$${n}`;
    return `${n} ${currency ?? ''}`.trim();
  }
</script>

<div style="display:flex;justify-content:space-between;align-items:center;gap:12px;margin-bottom:20px;flex-wrap:wrap">
  <div>
    <h1 style="margin:0;font-size:24px">Billing ledger</h1>
    <p style="margin:6px 0 0;color:var(--muted);font-size:13px">Cross-tenant payment orders (Sprint 10)</p>
  </div>
  <div style="display:flex;gap:10px;align-items:center;flex-wrap:wrap">
    <a class="btn ghost" href="{base}/billing/usage">Usage</a>
    <a class="btn ghost" href="{base}/billing/receipts">Receipts</a>
    <a class="btn ghost" href="{base}/billing/seller">Seller branding</a>
    <label style="font-size:13px;color:var(--muted)">
      Status
      <select bind:value={statusFilter} onchange={load} style="margin-left:8px">
        <option value="">all</option>
        <option value="pending">pending</option>
        <option value="paid">paid</option>
        <option value="failed">failed</option>
        <option value="cancelled">cancelled</option>
      </select>
    </label>
    <label style="font-size:13px;color:var(--muted)">
      Tenant id
      <input bind:value={tenantFilter} onchange={load} placeholder="optional" style="margin-left:8px;width:120px" />
    </label>
    <button class="btn ghost" type="button" onclick={load}>Refresh</button>
  </div>
</div>

<div class="card">
  {#if loading}
    <p style="color:var(--muted)">Loading…</p>
  {:else if !orders.length}
    <p style="color:var(--muted);margin:0">No payment orders yet. Complete a tenant checkout (mock or ChillPay) first.</p>
  {:else}
    <table class="table">
      <thead>
        <tr>
          <th>Created</th>
          <th>Tenant</th>
          <th>Package</th>
          <th>Amount</th>
          <th>Method</th>
          <th>Status</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#each orders as o (o.id)}
          <tr>
            <td style="font-size:12px;color:var(--muted)">{new Date(o.created_at).toLocaleString()}</td>
            <td>
              <div>{o.tenant_name || o.tenant_id}</div>
              <div style="font-size:11px;color:var(--muted)">{o.order_no}</div>
            </td>
            <td>{o.package_name || o.package_id}</td>
            <td>{formatAmount(o.amount_cents, o.currency)}</td>
            <td style="font-size:12px">{o.payment_method || '—'}</td>
            <td><span class="badge">{o.status}</span></td>
            <td><a href="{base}/billing/{o.id}">Detail</a></td>
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
    font-size: 14px;
  }
  .table th,
  .table td {
    text-align: left;
    padding: 10px 8px;
    border-bottom: 1px solid var(--border, #2a3344);
  }
  .table th {
    color: var(--muted);
    font-weight: 500;
    font-size: 12px;
  }
  .badge {
    display: inline-block;
    padding: 2px 8px;
    border-radius: 999px;
    background: rgba(34, 211, 238, 0.12);
    font-size: 12px;
  }
</style>
