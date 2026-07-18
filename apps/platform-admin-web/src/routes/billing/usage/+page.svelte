<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { base } from '$app/paths';
  import { getBillingUsage, type BillingUsageResponse } from '$lib/api/billingUsage';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';
  import { clearSession, loginPath } from '$lib/auth/session';

  const today = new Date().toISOString().slice(0, 10);
  let startDate = $state(today);
  let endDate = $state(today);
  let tenantFilter = $state('');
  let offset = $state(0);
  let loading = $state(true);
  let error = $state('');
  let data = $state<BillingUsageResponse | null>(null);

  async function load(reset = false) {
    if (reset) offset = 0;
    loading = true;
    error = '';
    try {
      data = await getBillingUsage({ start_date: startDate, end_date: endDate, tenant_id: tenantFilter.trim() || undefined, limit: 50, offset });
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearSession();
        await goto(loginPath($page.url.pathname));
        return;
      }
      error = err instanceof ApiError ? err.message : 'Unable to load billing usage';
      if (err instanceof ApiError && err.status >= 500) feedback.error(error);
    } finally {
      loading = false;
    }
  }

  function setToday() {
    startDate = today;
    endDate = today;
    load(true);
  }

  function formatAmount(minor: number, currency: string): string {
    const amount = (minor / 100).toLocaleString(undefined, { minimumFractionDigits: 2, maximumFractionDigits: 2 });
    if (currency === 'THB' || currency === '764') return `฿${amount}`;
    if (currency === 'USD') return `$${amount}`;
    return `${amount} ${currency || ''}`.trim();
  }

  function formatCost(microunits: number, currency: string): string {
    const amount = (microunits / 1_000_000).toLocaleString(undefined, { minimumFractionDigits: 4, maximumFractionDigits: 4 });
    return `${currency === 'USD' ? '$' : ''}${amount} ${currency === 'USD' ? '' : currency}`.trim();
  }

  function formatAICost(microunits: number, currency: string, status: string): string {
    if (status === 'empty' || status === 'unavailable') return 'Unavailable';
    const value = formatCost(microunits, currency);
    return status === 'warning' ? `≈ ${value}` : value;
  }

  function stateLabel(value: string): string {
    return value.replaceAll('_', ' ');
  }

  function canNext(): boolean {
    return !!data && offset + data.tenants.length < data.pagination.total;
  }

  function previous() {
    offset = Math.max(0, offset - 50);
    load();
  }

  function next() {
    if (!canNext()) return;
    offset += 50;
    load();
  }

  onMount(() => load());
</script>

<div class="page-head">
  <div>
    <h1>Billing &amp; usage</h1>
    <p>Paid package value, quota enforcement, and AI infrastructure cost coverage.</p>
  </div>
  <a class="btn ghost" href="{base}/billing">Orders ledger</a>
</div>

<section class="card filters" aria-label="Billing usage filters">
  <label>Start date<input type="date" bind:value={startDate} /></label>
  <label>End date<input type="date" bind:value={endDate} /></label>
  <label class="tenant">Tenant id<input bind:value={tenantFilter} placeholder="All tenants" /></label>
  <button class="btn" type="button" onclick={() => load(true)}>Apply</button>
  <button class="btn ghost" type="button" onclick={setToday}>Today</button>
</section>

{#if loading && !data}
  <section class="card state">Loading usage…</section>
{:else if error && !data}
  <section class="card state error-state"><strong>Usage unavailable</strong><span>{error}</span><button class="btn" type="button" onclick={() => load()}>Retry</button></section>
{:else if data}
  <div class="freshness {data.freshness.status}">
    Data: <strong>{stateLabel(data.freshness.status)}</strong> · {data.range.start_date} to {data.range.end_date} · generated {new Date(data.freshness.generated_at).toLocaleTimeString()}
  </div>

  <section class="kpis">
    <article class="card kpi"><span>Paid value</span><strong>{formatAmount(data.billing.paid_amount_minor, data.billing.currency)}</strong><small>{data.billing.paid_orders} paid orders</small></article>
    <article class="card kpi"><span>Reporting minutes</span><strong>{data.quota.reporting_minutes.toLocaleString()}</strong><small>Historical range usage</small></article>
    <article class="card kpi"><span>Quota enforcement</span><strong>{data.quota.enforcement.monthly_used.toLocaleString()} / {data.quota.enforcement.monthly_limit.toLocaleString()}</strong><small>{stateLabel(data.quota.enforcement.status)} · current snapshot</small></article>
    <article class="card kpi"><span>AI cost</span><strong>{formatAICost(data.ai_cost.observed_cost_microunits + data.ai_cost.estimated_cost_microunits, data.ai_cost.currency, data.ai_cost.status)}</strong><small>{data.ai_cost.coverage_percent.toFixed(1)}% measured coverage</small></article>
  </section>

  <section class="card coverage">
    <div><strong>AI usage coverage</strong><span class="badge {data.ai_cost.status}">{stateLabel(data.ai_cost.status)}</span></div>
    <div class="coverage-grid">
      <span><b>{data.ai_cost.observed_events}</b> observed</span>
      <span><b>{data.ai_cost.estimated_events}</b> estimated</span>
      <span><b>{data.ai_cost.unavailable_events}</b> unavailable</span>
      <span><b>{data.reconciliation.orders_entitlements}</b> orders / entitlements</span>
    </div>
  </section>

  <section class="card table-card">
    <div class="section-head"><div><h2>Tenant usage</h2><small>{data.pagination.total} active tenants</small></div><span>{data.reconciliation.ai_coverage} AI reconciliation</span></div>
    {#if !data.tenants.length}
      <div class="state">No active tenants in this view.</div>
    {:else}
      <div class="table-wrap"><table>
        <thead><tr><th>Tenant</th><th>Package</th><th>Paid</th><th>Quota</th><th>AI cost</th><th>State</th></tr></thead>
        <tbody>
          {#each data.tenants as tenant (tenant.tenant_id)}
            <tr>
              <td><strong>{tenant.name || tenant.slug}</strong><small>{tenant.slug}</small></td>
              <td>{tenant.package.name}<small>{tenant.package.status}</small></td>
              <td>{formatAmount(tenant.paid_amount_minor, tenant.currency || data.billing.currency)}<small>{tenant.paid_orders} orders</small></td>
              <td>{tenant.quota.monthly_used} / {tenant.quota.monthly_limit}<small>{tenant.reporting_minutes} range min</small></td>
              <td>{formatAICost(tenant.ai_observed_cost_microunits + tenant.ai_estimated_cost_microunits, data.ai_cost.currency, tenant.ai_coverage_percent > 0 ? data.ai_cost.status : 'unavailable')}<small>{tenant.ai_coverage_percent.toFixed(1)}% coverage</small></td>
              <td><span class="badge {tenant.status}">{stateLabel(tenant.status)}</span></td>
            </tr>
          {/each}
        </tbody>
      </table></div>
    {/if}
    <div class="pager"><span>Showing {data.pagination.offset + 1}–{Math.min(data.pagination.offset + data.tenants.length, data.pagination.total)} of {data.pagination.total}</span><div><button class="btn ghost" type="button" disabled={offset === 0 || loading} onclick={previous}>Previous</button><button class="btn ghost" type="button" disabled={!canNext() || loading} onclick={next}>Next</button></div></div>
  </section>
{:else}
  <section class="card state">No usage data.</section>
{/if}

<style>
  .page-head,.filters,.section-head,.pager,.coverage>div:first-child { display:flex; align-items:center; justify-content:space-between; gap:12px; flex-wrap:wrap; }
  .page-head { margin-bottom:20px; }
  h1 { margin:0; font-size:24px; }
  h2 { margin:0; font-size:17px; }
  p,small { color:var(--muted); }
  p { margin:6px 0 0; font-size:13px; }
  .filters { justify-content:flex-start; margin-bottom:14px; }
  label { color:var(--muted); font-size:12px; }
  input { display:block; margin-top:5px; min-height:36px; padding:7px 9px; border:1px solid var(--border); border-radius:7px; background:var(--panel); color:var(--text); }
  .tenant input { min-width:170px; }
  .freshness { margin:12px 0; font-size:12px; color:var(--muted); }
  .freshness.stale,.freshness.warning { color:#f59e0b; }
  .freshness.unavailable { color:#f87171; }
  .kpis { display:grid; grid-template-columns:repeat(4,minmax(0,1fr)); gap:12px; margin-bottom:12px; }
  .kpi { padding:16px; }
  .kpi span,.kpi small,.kpi strong { display:block; }
  .kpi span { color:var(--muted); font-size:12px; }
  .kpi strong { margin:8px 0 5px; font-size:22px; }
  .kpi small,td small { font-size:11px; }
  .coverage,.table-card { padding:16px; margin-bottom:12px; }
  .coverage-grid { display:grid; grid-template-columns:repeat(4,minmax(0,1fr)); gap:8px; margin-top:14px; }
  .coverage-grid span { padding:10px; border-radius:8px; background:var(--panel); color:var(--muted); font-size:12px; }
  .coverage-grid b { display:block; color:var(--text); font-size:18px; margin-bottom:4px; }
  .badge { display:inline-block; border-radius:999px; padding:3px 8px; background:rgba(34,211,238,.12); font-size:11px; text-transform:capitalize; }
  .badge.warning,.badge.degraded { background:rgba(245,158,11,.14); color:#fbbf24; }
  .badge.unavailable { background:rgba(248,113,113,.14); color:#fca5a5; }
  .table-wrap { overflow-x:auto; }
  table { width:100%; border-collapse:collapse; font-size:13px; }
  th,td { padding:11px 8px; text-align:left; border-bottom:1px solid var(--border); white-space:nowrap; }
  th { color:var(--muted); font-size:11px; font-weight:500; }
  td small { display:block; color:var(--muted); margin-top:3px; }
  .pager { margin-top:12px; color:var(--muted); font-size:12px; }
  .pager div { display:flex; gap:8px; }
  .state { padding:24px; color:var(--muted); display:flex; align-items:center; gap:12px; flex-wrap:wrap; }
  .error-state { color:#fca5a5; }
  @media (max-width:900px) { .kpis { grid-template-columns:repeat(2,minmax(0,1fr)); } .coverage-grid { grid-template-columns:repeat(2,minmax(0,1fr)); } }
  @media (max-width:700px) { .kpis { grid-template-columns:repeat(2,minmax(0,1fr)); } .filters { align-items:stretch; } .filters label,.filters input,.filters button { width:100%; } .tenant input { min-width:0; } .page-head { align-items:flex-start; } th,td { white-space:normal; min-width:120px; } }
</style>
