<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';
  import { getCallCenterStatistics, type CallCenterStatistics } from '$lib/api/callCenter';

  let startDate = $state(localISODate());
  let endDate = $state(localISODate());
  let stats = $state<CallCenterStatistics | null>(null);
  let loading = $state(true);
  let applying = $state(false);
  let loadError = $state('');

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/dashboard`)}`);
      return;
    }
    await load();
  });

  function localISODate(date = new Date()) {
    const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
    return local.toISOString().slice(0, 10);
  }

  async function load() {
    loading = true;
    loadError = '';
    try {
      stats = await getCallCenterStatistics({ startDate, endDate });
    } catch (err) {
      loadError = err instanceof ApiError ? err.message : 'Failed to load call-center statistics';
      feedback.error(loadError);
    } finally {
      loading = false;
    }
  }

  async function applyFilters(event: SubmitEvent) {
    event.preventDefault();
    applying = true;
    try { await load(); } finally { applying = false; }
  }

  async function resetToday() {
    startDate = localISODate();
    endDate = startDate;
    await load();
  }

  function duration(seconds: number) {
    const minutes = Math.floor(seconds / 60);
    const remaining = Math.round(seconds % 60);
    return `${minutes}m ${String(remaining).padStart(2, '0')}s`;
  }

  function percent(value: number, total: number) {
    return total > 0 ? Math.min(100, (value / total) * 100) : 0;
  }
</script>

<svelte:head><title>Call center | Monti Tenant</title></svelte:head>

<div class="call-center-page">
  <div class="page-head">
    <div><p class="eyebrow">Tenant operations</p><h1>Call center</h1><p class="muted">Completed AI conversations, duration, and package usage.</p></div>
    <span class="scope-badge">Tenant scoped</span>
  </div>

  <form class="filters card" onsubmit={applyFilters}>
    <label>Start date<input type="date" bind:value={startDate} /></label>
    <label>End date<input type="date" bind:value={endDate} /></label>
    <div class="filter-actions"><button class="btn" type="submit" disabled={applying}>{applying ? 'Loading...' : 'Apply'}</button><button class="btn ghost" type="button" onclick={resetToday}>Today</button></div>
  </form>

  {#if loading}
    <p class="muted loading">Loading call-center statistics...</p>
  {:else if loadError}
    <section class="card error-state"><strong>Unable to load call-center statistics.</strong><span>{loadError}</span><button class="btn ghost" type="button" onclick={load}>Retry</button></section>
  {:else if stats}
    <section class="kpi-grid" aria-label="Call-center summary">
      <article class="card kpi"><span>Completed conversations</span><strong>{stats.total_completed_conversations}</strong><small>{stats.range.start_date} to {stats.range.end_date}</small></article>
      <article class="card kpi"><span>Total talk time</span><strong>{duration(stats.total_duration_seconds)}</strong><small>{stats.timezone}</small></article>
      <article class="card kpi"><span>Average conversation</span><strong>{duration(stats.average_duration_seconds)}</strong><small>Across archived records</small></article>
      <article class="card kpi"><span>Daily package usage</span><strong>{stats.daily_usage.call_minutes} min</strong><small>{stats.call_limits?.max_call_minutes_per_day ? `Daily cap ${stats.call_limits.max_call_minutes_per_day} min` : 'No daily cap set'}</small></article>
    </section>

    {#if stats.total_completed_conversations === 0}
      <section class="card empty"><strong>No completed conversations in this range.</strong><span>Try another date range after a call has been archived.</span></section>
    {:else}
      <div class="dashboard-grid">
        <section class="card usage-card">
          <div class="section-head"><div><h2>Monthly call minutes</h2><p class="muted">Package allowance</p></div><strong>{stats.quota?.usage?.monthly_call_minutes ?? 0} / {stats.quota?.limits?.max_monthly_call_minutes ?? 'unlimited'}</strong></div>
          <div class="progress"><i style={`width: ${percent(stats.quota?.usage?.monthly_call_minutes ?? 0, stats.quota?.limits?.max_monthly_call_minutes ?? 0)}%`}></i></div>
          <div class="usage-meta"><span>{stats.quota?.package?.name ?? 'No package assigned'}</span><span>{stats.quota?.period ?? 'Current period'}</span></div>
        </section>
        <section class="card breakdown"><div class="section-head"><h2>By channel</h2><span>{stats.by_channel.length}</span></div>{#each stats.by_channel as bucket (bucket.channel)}<div class="breakdown-row"><span><strong>{bucket.channel || 'Unknown'}</strong><small>{bucket.completed} completed</small></span><b>{duration(bucket.total_duration_seconds)}</b></div>{/each}</section>
        <section class="card breakdown"><div class="section-head"><h2>By AI employee</h2><span>{stats.by_avatar.length}</span></div>{#each stats.by_avatar as bucket (bucket.id)}<div class="breakdown-row"><span><strong>{bucket.name || bucket.id || 'Unknown avatar'}</strong><small>{bucket.completed} completed</small></span><b>{duration(bucket.total_duration_seconds)}</b></div>{/each}</section>
      </div>
    {/if}
    {#if stats.freshness}<p class="freshness">Analytics updated {new Date(stats.freshness).toLocaleString()}</p>{/if}
  {/if}
</div>

<style>
  .call-center-page { display: grid; gap: 18px; max-width: 1240px; margin: 0 auto; }
  .page-head, .section-head { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
  .page-head { align-items: start; }
  .eyebrow { margin: 0 0 6px; color: var(--cyan); font-size: 11px; letter-spacing: .12em; text-transform: uppercase; }
  h1, h2, p { margin: 0; }
  h1 { font-size: 34px; }
  h2 { font-size: 16px; }
  .muted, .kpi span, .kpi small, .usage-meta, .freshness { color: var(--muted); }
  .scope-badge, .section-head > span { border: 1px solid var(--line); border-radius: 999px; padding: 6px 10px; color: var(--muted); font-size: 11px; white-space: nowrap; }
  .filters { display: grid; grid-template-columns: repeat(2, minmax(150px, 220px)) auto; gap: 12px; align-items: end; }
  label { display: grid; gap: 6px; color: var(--muted); font-size: 11px; }
  input { width: 100%; border: 1px solid var(--line); border-radius: 8px; padding: 9px 10px; color: var(--ink); background: rgb(4 9 20 / 74%); }
  .filter-actions { display: flex; gap: 8px; }
  .btn { border: 1px solid rgb(74 135 255 / 46%); border-radius: 9px; padding: 9px 13px; background: linear-gradient(100deg, var(--blue), var(--violet)); color: var(--ink); font-weight: 650; white-space: nowrap; }
  .btn.ghost { background: rgb(13 23 42 / 62%); }
  .btn:disabled { opacity: .55; cursor: not-allowed; }
  .loading { padding: 24px 0; }
  .kpi-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; }
  .kpi { display: grid; gap: 7px; min-height: 112px; }
  .kpi strong { font-size: 27px; }
  .kpi small { font-size: 11px; }
  .empty, .error-state { display: grid; gap: 8px; padding: 32px; color: var(--muted); }
  .error-state strong { color: var(--ink); }
  .error-state .btn { justify-self: start; }
  .dashboard-grid { display: grid; grid-template-columns: 1.25fr 1fr 1fr; gap: 12px; align-items: start; }
  .usage-card { grid-column: 1 / -1; }
  .usage-card .section-head > strong { font-size: 20px; }
  .usage-card .section-head p { margin-top: 5px; font-size: 12px; }
  .progress { height: 10px; margin-top: 18px; overflow: hidden; border-radius: 999px; background: rgb(104 127 166 / 18%); }
  .progress i { display: block; height: 100%; border-radius: inherit; background: linear-gradient(90deg, var(--cyan), var(--blue)); }
  .usage-meta { display: flex; justify-content: space-between; gap: 12px; margin-top: 10px; font-size: 11px; }
  .breakdown-row { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 13px 0; border-bottom: 1px solid var(--line); }
  .breakdown-row:last-child { border-bottom: 0; }
  .breakdown-row span { display: grid; gap: 4px; min-width: 0; }
  .breakdown-row strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .breakdown-row small { color: var(--muted); font-size: 11px; }
  .breakdown-row b { color: var(--cyan); white-space: nowrap; }
  .freshness { font-size: 11px; text-align: right; }
  @media (max-width: 980px) { .kpi-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } .dashboard-grid { grid-template-columns: 1fr 1fr; } .usage-card { grid-column: 1 / -1; } }
  @media (max-width: 640px) { .page-head { flex-direction: column; } .filters { grid-template-columns: 1fr 1fr; } .filter-actions { grid-column: 1 / -1; } .kpi-grid, .dashboard-grid { grid-template-columns: 1fr; } .usage-card { grid-column: auto; } }
</style>
