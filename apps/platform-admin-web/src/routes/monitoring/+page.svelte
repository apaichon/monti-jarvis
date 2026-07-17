<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { base } from '$app/paths';
  import { clearSession, loginPath } from '$lib/auth/session';
  import { ApiError } from '$lib/api/http';
  import {
    getPlatformSystemPerformance,
    type MonitoringAnalytics,
    type MonitoringComponent,
    type MonitoringStatus,
    type PlatformPerformance
  } from '$lib/api/monitoring';

  let tenantFilter = $state('');
  let statusFilter = $state<'' | Exclude<MonitoringStatus, 'disabled'>>('');
  let pageSize = $state('50');
  let offset = $state(0);
  let data = $state<PlatformPerformance | null>(null);
  let loading = $state(true);
  let error = $state('');

  const statusNames: Record<string, string> = {
    operational: 'Operational',
    degraded: 'Degraded',
    unavailable: 'Unavailable',
    disabled: 'Disabled',
    current: 'Current',
    stale: 'Stale'
  };

  function statusLabel(value: string) {
    return statusNames[value] ?? value;
  }

  function formatTime(value?: string | null) {
    if (!value) return 'Not recorded';
    return new Date(value).toLocaleString();
  }

  function formatAge(seconds: number) {
    if (!seconds) return '0s';
    if (seconds < 60) return `${seconds}s`;
    return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
  }

  function analyticsDetail(analytics: MonitoringAnalytics) {
    if (analytics.status === 'current') return analytics.last_projected_at ? `Projected ${formatTime(analytics.last_projected_at)}` : 'Freshness current';
    if (analytics.status === 'stale') return analytics.last_projected_at ? `Last projected ${formatTime(analytics.last_projected_at)}` : 'Projection is stale';
    return statusLabel(analytics.status);
  }

  function componentDetail(component: MonitoringComponent) {
    return component.latency_ms === null ? statusLabel(component.status) : `${statusLabel(component.status)} · ${component.latency_ms}ms`;
  }

  async function load() {
    loading = true;
    error = '';
    try {
      data = await getPlatformSystemPerformance({
        tenant_id: tenantFilter.trim(),
        status: statusFilter,
        limit: pageSize,
        offset: String(offset)
      });
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearSession();
        await goto(loginPath($page.url.pathname));
        return;
      }
      error = err instanceof ApiError ? err.message : 'Unable to load system performance';
    } finally {
      loading = false;
    }
  }

  function applyFilters() {
    offset = 0;
    load();
  }

  function resetFilters() {
    tenantFilter = '';
    statusFilter = '';
    pageSize = '50';
    offset = 0;
    load();
  }

  function nextPage() {
    if (!data || offset + data.tenants.length >= data.total) return;
    offset += Number(pageSize);
    load();
  }

  function previousPage() {
    offset = Math.max(0, offset - Number(pageSize));
    load();
  }

  onMount(() => load());
</script>

<div class="page-head">
  <div>
    <p class="eyebrow">TENANT OPERATIONS</p>
    <h1>System performance</h1>
    <p class="lede">Service health and analytics freshness across active tenants.</p>
  </div>
  <span class="scope-badge">Platform scoped</span>
</div>

<section class="filters card" aria-label="Monitoring filters">
  <label>Tenant ID<input bind:value={tenantFilter} placeholder="All tenants" /></label>
  <label>Status<select bind:value={statusFilter}><option value="">All statuses</option><option value="operational">Operational</option><option value="degraded">Degraded</option><option value="unavailable">Unavailable</option></select></label>
  <label>Rows<select bind:value={pageSize}><option value="25">25</option><option value="50">50</option><option value="100">100</option></select></label>
  <div class="filter-actions"><button class="btn" type="button" onclick={applyFilters} disabled={loading}>Apply</button><button class="btn ghost" type="button" onclick={resetFilters} disabled={loading}>Reset</button></div>
</section>

{#if error}
  <section class="error card" role="alert">
    <strong>Unable to load system performance.</strong>
    <span>{error}</span>
    <button class="btn ghost" type="button" onclick={load} disabled={loading}>Retry</button>
  </section>
{/if}

{#if loading && !data}
  <section class="loading-grid" aria-label="Loading system performance">
    <div class="skeleton wide"></div><div class="skeleton"></div><div class="skeleton table-skeleton"></div>
  </section>
{:else if data}
  <section class="summary-grid" aria-label="System performance summary">
    <div class="metric card"><span>Overall status</span><strong class:operational={data.overall_status === 'operational'} class:degraded={data.overall_status === 'degraded'} class:unavailable={data.overall_status === 'unavailable'}>{statusLabel(data.overall_status)}</strong><small>Checked {formatTime(data.checked_at)}</small></div>
    <div class="metric card"><span>Operational tenants</span><strong class="operational">{data.summary.operational}</strong><small>{data.summary.tenants_total} tenants in filtered view</small></div>
    <div class="metric card"><span>Degraded tenants</span><strong class="degraded">{data.summary.degraded}</strong><small>Includes stale or disabled dependencies</small></div>
    <div class="metric card"><span>Unavailable tenants</span><strong class="unavailable">{data.summary.unavailable}</strong><small>Requires operator attention</small></div>
  </section>

  <section class="split-grid">
    <div class="card panel">
      <div class="section-head"><div><h2>Shared dependencies</h2><p>Normalized request-time probes</p></div><span class="count">{data.components.length}</span></div>
      <div class="dependency-list">
        {#each data.components as component (component.name)}
          <div class="dependency-row"><span class="state-dot" class:operational={component.status === 'operational'} class:degraded={component.status === 'degraded'} class:unavailable={component.status === 'unavailable'}></span><strong>{component.name}</strong><span class="state" class:operational={component.status === 'operational'} class:degraded={component.status === 'degraded'} class:unavailable={component.status === 'unavailable'}>{componentDetail(component)}</span></div>
        {/each}
      </div>
    </div>
    <div class="card panel">
      <div class="section-head"><div><h2>Audit delivery</h2><p>Safe spool and ClickHouse health</p></div><span class="state" class:operational={data.audit.status === 'operational'} class:degraded={data.audit.status === 'degraded'} class:unavailable={data.audit.status === 'unavailable'}>{statusLabel(data.audit.status)}</span></div>
      <div class="audit-grid"><span>Mode<strong>{data.audit.mode || 'Not configured'}</strong></span><span>Pending files<strong>{data.audit.pending_files}</strong></span><span>Failed files<strong>{data.audit.failed_files}</strong></span><span>Oldest pending<strong>{formatAge(data.audit.oldest_pending_file_age_seconds)}</strong></span></div>
      <small class="muted">Last transfer {formatTime(data.audit.last_successful_transfer)}</small>
    </div>
  </section>

  <section class="card table-card">
    <div class="section-head"><div><h2>Tenant health</h2><p>{data.tenants.length ? `${offset + 1}-${offset + data.tenants.length} of ${data.total} matching tenants` : 'No tenants match these filters.'}</p></div><button class="btn ghost" type="button" onclick={load} disabled={loading}>Refresh</button></div>
    {#if !data.tenants.length}
      <div class="empty"><strong>No tenants match these filters.</strong><span>Reset filters or choose another status.</span></div>
    {:else}
      <div class="table-wrap">
        <table>
          <thead><tr><th>Tenant</th><th>Status</th><th>Analytics freshness</th><th>Audit</th></tr></thead>
          <tbody>
            {#each data.tenants as tenant (tenant.tenant_id)}
              <tr>
                <td><strong>{tenant.name}</strong><small>{tenant.slug} · {tenant.tenant_id}</small></td>
                <td><span class="state" class:operational={tenant.status === 'operational'} class:degraded={tenant.status === 'degraded'} class:unavailable={tenant.status === 'unavailable'}>{statusLabel(tenant.status)}</span></td>
                <td><span class="state" class:operational={tenant.analytics.status === 'current'} class:degraded={tenant.analytics.status === 'stale'} class:unavailable={tenant.analytics.status === 'unavailable'}>{statusLabel(tenant.analytics.status)}</span><small>{analyticsDetail(tenant.analytics)}</small></td>
                <td><span class="state" class:operational={tenant.audit_status === 'operational'} class:degraded={tenant.audit_status === 'degraded' || tenant.audit_status === 'disabled'} class:unavailable={tenant.audit_status === 'unavailable'}>{statusLabel(tenant.audit_status)}</span></td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <div class="pager"><button class="btn ghost" type="button" onclick={previousPage} disabled={loading || offset === 0}>Previous</button><button class="btn ghost" type="button" onclick={nextPage} disabled={loading || offset + data.tenants.length >= data.total}>Next</button></div>
    {/if}
  </section>
{/if}

<style>
  .page-head { display: flex; align-items: end; justify-content: space-between; gap: 18px; margin-bottom: 22px; }
  .eyebrow { margin: 0 0 8px; color: var(--cyan); font-size: 11px; letter-spacing: .16em; font-weight: 700; }
  h1 { margin: 0; font-size: 32px; letter-spacing: 0; }
  .lede { margin: 7px 0 0; color: var(--muted); }
  .scope-badge, .count { padding: 8px 12px; border: 1px solid var(--line); border-radius: 18px; color: var(--muted); font-size: 12px; white-space: nowrap; }
  .filters { display: grid; grid-template-columns: minmax(220px, 1.4fr) repeat(2, minmax(130px, .7fr)) auto; gap: 14px; margin-bottom: 14px; }
  label { display: grid; gap: 6px; color: var(--muted); font-size: 12px; }
  input, select { width: 100%; min-height: 40px; border: 1px solid var(--line); border-radius: 8px; background: rgb(4 10 22 / 74%); color: var(--ink); padding: 8px 10px; }
  .filter-actions { display: flex; align-items: end; gap: 8px; }
  .summary-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 14px; margin-bottom: 14px; }
  .metric { display: grid; gap: 8px; min-height: 124px; }
  .metric span, .metric small, .section-head p, .muted { color: var(--muted); font-size: 12px; }
  .metric strong { font-size: 26px; }
  .metric small { margin-top: auto; }
  .operational { color: var(--success); }.degraded { color: #f5b94c; }.unavailable { color: var(--danger); }
  .split-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; margin-bottom: 14px; }
  .panel { min-width: 0; }
  .section-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
  h2 { margin: 0; font-size: 18px; }.section-head p { margin: 5px 0 0; }
  .dependency-list { display: grid; gap: 8px; }.dependency-row { display: grid; grid-template-columns: 8px minmax(80px, 1fr) auto; gap: 10px; align-items: center; padding: 10px 0; border-bottom: 1px solid var(--line); font-size: 13px; }.dependency-row:last-child { border-bottom: 0; }.state-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--muted); }.state-dot.operational { background: var(--success); box-shadow: 0 0 9px var(--success); }.state-dot.degraded { background: #f5b94c; }.state-dot.unavailable { background: var(--danger); }
  .state { color: var(--muted); font-size: 12px; }.audit-grid { display: grid; grid-template-columns: repeat(2, 1fr); gap: 14px; margin-bottom: 18px; }.audit-grid span { color: var(--muted); font-size: 12px; }.audit-grid strong { display: block; color: var(--ink); font-size: 15px; margin-top: 5px; }.table-card { overflow: hidden; }.table-wrap { overflow-x: auto; } table { width: 100%; border-collapse: collapse; min-width: 620px; font-size: 13px; } th, td { padding: 12px 9px; border-bottom: 1px solid var(--line); text-align: left; vertical-align: top; } th { color: var(--muted); font-size: 11px; font-weight: 600; } td small { display: block; color: var(--muted); margin-top: 4px; font-size: 11px; }.pager { display: flex; justify-content: flex-end; gap: 8px; padding-top: 14px; }.empty, .error { display: grid; gap: 8px; color: var(--muted); }.empty { padding: 30px 0; }.empty strong, .error strong { color: var(--ink); }.error { margin-bottom: 14px; border-color: rgb(255 92 122 / 35%); }.error .btn { width: max-content; }.loading-grid { display: grid; gap: 14px; }.skeleton { min-height: 124px; border-radius: var(--radius); background: linear-gradient(90deg, rgb(16 26 46 / 80%), rgb(28 42 68 / 65%), rgb(16 26 46 / 80%)); background-size: 220% 100%; animation: pulse 1.6s ease-in-out infinite; }.skeleton.wide { min-height: 260px; }.skeleton.table-skeleton { min-height: 360px; } @keyframes pulse { 0% { background-position: 0 0; } 100% { background-position: -220% 0; } }
  @media (max-width: 900px) { .summary-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }.filters { grid-template-columns: repeat(2, minmax(0, 1fr)); }.filter-actions { align-items: end; }.split-grid { grid-template-columns: 1fr; } }
  @media (max-width: 520px) { .page-head { align-items: start; flex-direction: column; }.scope-badge { align-self: start; }.filters { grid-template-columns: 1fr; }.filter-actions { align-items: stretch; }.summary-grid { grid-template-columns: 1fr 1fr; }.metric { min-height: 112px; padding: 15px; }.metric strong { font-size: 21px; }.audit-grid { gap: 10px; } }
</style>
