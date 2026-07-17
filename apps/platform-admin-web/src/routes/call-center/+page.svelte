<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { clearSession, loginPath } from '$lib/auth/session';
  import { ApiError } from '$lib/api/http';
  import {
    getPlatformCallCenterStatistics,
    type CallCenterDimension,
    type PlatformCallCenterStatistics,
    type PlatformCallCenterTenant
  } from '$lib/api/callCenter';

  const now = new Date();
  const today = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}-${String(now.getDate()).padStart(2, '0')}`;
  let startDate = $state(today);
  let endDate = $state(today);
  let tenantFilter = $state('');
  let offset = $state(0);
  let data = $state<PlatformCallCenterStatistics | null>(null);
  let loading = $state(true);
  let error = $state('');
  let unavailable = $state(false);

  const statusNames: Record<string, string> = {
    current: 'Current',
    stale: 'Stale',
    empty: 'No activity',
    unavailable: 'Unavailable'
  };

  function statusLabel(value: string) {
    return statusNames[value] ?? value;
  }

  function formatNumber(value: number) {
    return new Intl.NumberFormat().format(value || 0);
  }

  function formatMinutes(value: number) {
    return `${formatNumber(value)} min`;
  }

  function formatDuration(seconds: number) {
    if (!seconds) return '0 sec';
    const minutes = Math.floor(seconds / 60);
    const remaining = Math.round(seconds % 60);
    return minutes ? `${minutes}m ${remaining}s` : `${remaining}s`;
  }

  function formatAverage(seconds: number) {
    return `${Math.round(seconds || 0)} sec`;
  }

  function formatTime(value?: string) {
    return value ? new Date(value).toLocaleString() : 'Not recorded';
  }

  async function load() {
    loading = true;
    error = '';
    unavailable = false;
    try {
      data = await getPlatformCallCenterStatistics({
        start_date: startDate,
        end_date: endDate,
        tenant_id: tenantFilter.trim(),
        limit: '50',
        offset: String(offset)
      });
    } catch (err) {
      if (err instanceof ApiError && err.status === 401) {
        clearSession();
        await goto(loginPath($page.url.pathname));
        return;
      }
      unavailable = err instanceof ApiError && err.status === 503;
      error = err instanceof ApiError ? err.message : 'Unable to load call-center statistics';
    } finally {
      loading = false;
    }
  }

  function applyFilters() {
    offset = 0;
    load();
  }

  function showToday() {
    startDate = today;
    endDate = today;
    offset = 0;
    load();
  }

  function nextPage() {
    if (!data || offset + data.tenants.length >= data.pagination.total) return;
    offset += data.pagination.limit;
    load();
  }

  function previousPage() {
    offset = Math.max(0, offset - (data?.pagination.limit || 50));
    load();
  }

  function dimensionLabel(item: CallCenterDimension) {
    return item.channel || item.name || item.id || 'Unknown';
  }

  function tenantRangeLabel(tenants: PlatformCallCenterTenant[]) {
    if (!tenants.length) return 'No active tenants match this filter.';
    return `${offset + 1}–${offset + tenants.length} of ${data?.pagination.total ?? tenants.length} active tenants`;
  }

  onMount(() => load());
</script>

<div class="page-head">
  <div>
    <p class="eyebrow">PLATFORM ANALYTICS</p>
    <h1>Call-center statistics</h1>
    <p class="lede">Completed AI conversations across active tenants.</p>
  </div>
  <span class="scope-badge">Platform scoped</span>
</div>

<section class="filters card" aria-label="Call-center statistics filters">
  <label>Tenant ID<input bind:value={tenantFilter} placeholder="All active tenants" maxlength="128" /></label>
  <label>Start date<input type="date" bind:value={startDate} /></label>
  <label>End date<input type="date" bind:value={endDate} /></label>
  <div class="filter-actions"><button class="btn" type="button" onclick={applyFilters} disabled={loading}>Apply</button><button class="btn ghost" type="button" onclick={showToday} disabled={loading}>Today</button></div>
</section>

{#if error}
  <section class="error card" role="alert">
    <strong>{unavailable ? 'Analytics are unavailable.' : 'Unable to load statistics.'}</strong>
    <span>{unavailable ? 'ClickHouse did not provide a safe result. Activity is not being shown as zero.' : error}</span>
    <button class="btn ghost" type="button" onclick={load} disabled={loading}>Retry</button>
  </section>
{/if}

{#if loading && !data}
  <section class="loading-grid" aria-label="Loading call-center statistics">
    <div class="skeleton"></div><div class="skeleton"></div><div class="skeleton wide"></div>
  </section>
{:else if data}
  <section class="freshness card" aria-live="polite">
    <div><span class="status-dot" class:current={data.freshness.status === 'current'} class:stale={data.freshness.status === 'stale'} class:unavailable={data.freshness.status === 'unavailable'}></span><strong>Analytics {statusLabel(data.freshness.status)}</strong></div>
    <span>{data.range.start_date} to {data.range.end_date} · {data.range.timezone}</span>
    <span>Last projected {formatTime(data.freshness.last_projected_at)}</span>
  </section>

  <section class="summary-grid" aria-label="Call-center summary">
    <div class="metric card"><span>Completed conversations</span><strong>{formatNumber(data.totals.completed_conversations)}</strong><small>{formatMinutes(data.totals.range_call_minutes)} of completed activity</small></div>
    <div class="metric card"><span>Total talk time</span><strong>{formatDuration(data.totals.total_duration_seconds)}</strong><small>Average {formatAverage(data.totals.average_duration_seconds)}</small></div>
    <div class="metric card"><span>Channels</span><strong>{formatNumber(data.totals.chat_conversations)} chat</strong><small>{formatNumber(data.totals.voice_conversations)} voice</small></div>
    <div class="metric card"><span>Satisfaction</span><strong>{data.totals.average_satisfaction ? `${data.totals.average_satisfaction.toFixed(2)} / 5` : '—'}</strong><small>{formatNumber(data.totals.reviewed_conversations)} reviewed · {data.totals.review_completion_rate.toFixed(1)}%</small></div>
  </section>

  <section class="split-grid">
    <div class="card panel">
      <div class="section-head"><div><h2>By channel</h2><p>Completed activity and duration</p></div><span class="count">{data.by_channel.length}</span></div>
      {#if data.by_channel.length}
        <div class="dimension-list">
          {#each data.by_channel as item (item.channel)}
            <div class="dimension-row"><strong>{dimensionLabel(item)}</strong><span>{formatNumber(item.completed)} conversations</span><small>{formatDuration(item.total_duration_seconds)} · avg {formatAverage(item.average_duration_seconds)}</small></div>
          {/each}
        </div>
      {:else}<div class="empty compact"><strong>No channel activity.</strong><span>This valid range contains no completed calls.</span></div>{/if}
    </div>
    <div class="card panel">
      <div class="section-head"><div><h2>By AI avatar</h2><p>Aggregate usage only</p></div><span class="count">{data.by_avatar.length}</span></div>
      {#if data.by_avatar.length}
        <div class="dimension-list">
          {#each data.by_avatar as item (item.id)}
            <div class="dimension-row"><strong>{dimensionLabel(item)}</strong><span>{formatNumber(item.completed)} conversations</span><small>{formatMinutes(Math.ceil(item.total_duration_seconds / 60))} · avg {formatAverage(item.average_duration_seconds)}</small></div>
          {/each}
        </div>
      {:else}<div class="empty compact"><strong>No avatar activity.</strong><span>Avatar breakdowns appear when calls are completed.</span></div>{/if}
    </div>
  </section>

  <section class="insight-grid">
    <div class="card insight"><span>Package usage</span><strong>{formatMinutes(data.package_usage.range_call_minutes)}</strong><small>{formatNumber(data.package_usage.active_package_tenants)} active package tenants</small><em>Live enforcement counters are {data.package_usage.enforcement_counters.replace('_', ' ')}.</em></div>
    <div class="card insight"><span>Rating distribution</span><strong>{formatNumber(data.totals.reviewed_conversations)} reviewed</strong><div class="distribution">{#each ['5', '4', '3', '2', '1'] as score}<span>{score}★ {formatNumber(data.totals.satisfaction_distribution[score] ?? 0)}</span>{/each}</div><small class:unavailable={data.enrichment.satisfaction === 'unavailable'}>{data.enrichment.satisfaction === 'available' ? 'Satisfaction enrichment available' : 'Satisfaction enrichment unavailable'}</small></div>
  </section>

  <section class="card tenant-card">
    <div class="section-head"><div><h2>Tenant breakdown</h2><p>{tenantRangeLabel(data.tenants)}</p></div><button class="btn ghost" type="button" onclick={load} disabled={loading}>Refresh</button></div>
    {#if !data.tenants.length}
      <div class="empty"><strong>No active tenants match this view.</strong><span>Try removing the tenant filter or choose another date range.</span></div>
    {:else}
      <div class="tenant-list">
        {#each data.tenants as tenant (tenant.tenant_id)}
          <article class="tenant-row">
            <div class="tenant-identity"><strong>{tenant.name}</strong><small>{tenant.slug} · {tenant.tenant_id}</small><span class="state" class:current={tenant.analytics_status === 'current'} class:stale={tenant.analytics_status === 'stale'}>{statusLabel(tenant.analytics_status)}</span></div>
            <div><span class="row-label">Conversations</span><strong>{formatNumber(tenant.completed_conversations)}</strong><small>{formatMinutes(tenant.range_call_minutes)} · avg {formatAverage(tenant.average_duration_seconds)}</small></div>
            <div><span class="row-label">Satisfaction</span><strong>{tenant.average_satisfaction ? `${tenant.average_satisfaction.toFixed(2)} / 5` : '—'}</strong><small>{formatNumber(tenant.reviewed_conversations)} reviewed · {tenant.review_completion_rate.toFixed(1)}%</small></div>
            <div><span class="row-label">Package</span><strong>{tenant.package.name}</strong><small>{tenant.package.status}</small></div>
          </article>
        {/each}
      </div>
      <div class="pager"><button class="btn ghost" type="button" onclick={previousPage} disabled={loading || offset === 0}>Previous</button><button class="btn ghost" type="button" onclick={nextPage} disabled={loading || offset + data.tenants.length >= data.pagination.total}>Next</button></div>
    {/if}
  </section>
{/if}

<style>
  .page-head { display: flex; align-items: end; justify-content: space-between; gap: 18px; margin-bottom: 22px; }
  .eyebrow { margin: 0 0 8px; color: var(--cyan); font-size: 11px; letter-spacing: .16em; font-weight: 700; }
  h1 { margin: 0; font-size: 32px; }.lede { margin: 7px 0 0; color: var(--muted); }.scope-badge, .count { padding: 8px 12px; border: 1px solid var(--line); border-radius: 18px; color: var(--muted); font-size: 12px; white-space: nowrap; }
  .filters { display: grid; grid-template-columns: minmax(220px, 1.4fr) repeat(2, minmax(130px, .7fr)) auto; gap: 14px; margin-bottom: 14px; } label { display: grid; gap: 6px; color: var(--muted); font-size: 12px; } input { width: 100%; min-height: 40px; border: 1px solid var(--line); border-radius: 8px; background: rgb(4 10 22 / 74%); color: var(--ink); padding: 8px 10px; }.filter-actions { display: flex; align-items: end; gap: 8px; }
  .freshness { display: flex; align-items: center; gap: 18px; flex-wrap: wrap; margin-bottom: 14px; color: var(--muted); font-size: 12px; }.freshness div { display: flex; align-items: center; gap: 8px; color: var(--ink); }.status-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--muted); }.status-dot.current { background: var(--success); box-shadow: 0 0 9px var(--success); }.status-dot.stale { background: #f5b94c; }.status-dot.unavailable { background: var(--danger); }
  .summary-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 14px; margin-bottom: 14px; }.metric { display: grid; gap: 8px; min-height: 124px; }.metric span, .metric small, .section-head p, .row-label { color: var(--muted); font-size: 12px; }.metric strong { font-size: 25px; }.metric small { margin-top: auto; }.current { color: var(--success); }.stale { color: #f5b94c; }.unavailable { color: var(--danger); }
  .split-grid, .insight-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; margin-bottom: 14px; }.panel, .tenant-card { min-width: 0; }.section-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 14px; }.section-head p { margin: 5px 0 0; } h2 { margin: 0; font-size: 18px; }
  .dimension-list { display: grid; gap: 8px; }.dimension-row { display: grid; grid-template-columns: minmax(90px, 1fr) auto; gap: 4px 12px; padding: 10px 0; border-bottom: 1px solid var(--line); font-size: 13px; }.dimension-row:last-child { border-bottom: 0; }.dimension-row span, .dimension-row small { color: var(--muted); font-size: 11px; text-align: right; }.dimension-row small { grid-column: 1 / -1; text-align: left; }
  .insight { display: grid; gap: 8px; min-height: 145px; }.insight > span, .insight small, .insight em { color: var(--muted); font-size: 12px; }.insight strong { font-size: 23px; }.insight em { font-style: normal; margin-top: auto; }.distribution { display: flex; gap: 10px; flex-wrap: wrap; color: #f5b94c; font-size: 12px; }
  .tenant-card { overflow: hidden; }.tenant-list { display: grid; }.tenant-row { display: grid; grid-template-columns: minmax(190px, 1.5fr) repeat(3, minmax(125px, 1fr)); gap: 18px; align-items: center; padding: 15px 0; border-bottom: 1px solid var(--line); }.tenant-row > div { min-width: 0; display: grid; gap: 4px; }.tenant-row strong { font-size: 13px; }.tenant-row small { color: var(--muted); font-size: 11px; }.tenant-identity { align-content: center; }.tenant-identity .state { width: max-content; font-size: 11px; }.pager { display: flex; justify-content: flex-end; gap: 8px; padding-top: 14px; }.empty, .error { display: grid; gap: 8px; color: var(--muted); }.empty { padding: 30px 0; }.empty.compact { padding: 12px 0; }.empty strong, .error strong { color: var(--ink); }.error { margin-bottom: 14px; border-color: rgb(255 92 122 / 35%); }.error .btn { width: max-content; }.loading-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }.skeleton { min-height: 180px; border-radius: var(--radius); background: linear-gradient(90deg, rgb(16 26 46 / 80%), rgb(28 42 68 / 65%), rgb(16 26 46 / 80%)); background-size: 220% 100%; animation: pulse 1.6s ease-in-out infinite; }.skeleton.wide { grid-column: 1 / -1; min-height: 300px; } @keyframes pulse { 0% { background-position: 0 0; } 100% { background-position: -220% 0; } }
  @media (max-width: 900px) { .summary-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }.filters { grid-template-columns: repeat(2, minmax(0, 1fr)); }.split-grid, .insight-grid { grid-template-columns: 1fr; }.tenant-row { grid-template-columns: minmax(180px, 1.4fr) repeat(3, minmax(100px, 1fr)); gap: 10px; } }
  @media (max-width: 700px) { .page-head { align-items: start; flex-direction: column; }.scope-badge { align-self: start; }.filters { grid-template-columns: 1fr; }.filter-actions { align-items: stretch; }.freshness { display: grid; gap: 8px; }.tenant-row { grid-template-columns: 1fr 1fr; gap: 13px; }.tenant-identity { grid-column: 1 / -1; }.loading-grid { grid-template-columns: 1fr; }.skeleton.wide { grid-column: auto; } }
  @media (max-width: 420px) { .summary-grid { grid-template-columns: 1fr; }.metric { min-height: 105px; }.tenant-row { grid-template-columns: 1fr; }.tenant-identity { grid-column: auto; }.dimension-row { grid-template-columns: 1fr; }.dimension-row span { text-align: left; } }
</style>
