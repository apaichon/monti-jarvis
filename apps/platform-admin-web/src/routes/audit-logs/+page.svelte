<script lang="ts">
  import { onMount } from 'svelte';
  import { listAuditEvents, getAuditHealth, type AuditEvent, type AuditHealth } from '$lib/api/audit';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  const today = new Date().toISOString().slice(0, 10);
  const weekAgo = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000).toISOString().slice(0, 10);
  let startDate = $state(weekAgo);
  let endDate = $state(today);
  let tenantID = $state('');
  let actorID = $state('');
  let action = $state('');
  let resourceType = $state('');
  let outcome = $state('');
  let events = $state<AuditEvent[]>([]);
  let health = $state<AuditHealth | null>(null);
  let nextCursor = $state('');
  let loading = $state(true);
  let expandedID = $state('');

  async function load(cursor = '') {
    loading = true;
    try {
      const params = {
        start_date: startDate,
        end_date: endDate,
        tenant_id: tenantID.trim(),
        actor_id: actorID.trim(),
        action: action.trim(),
        resource_type: resourceType.trim(),
        outcome,
        limit: '50',
        cursor
      };
      const [page, status] = await Promise.all([listAuditEvents(params), getAuditHealth()]);
      events = page.events ?? [];
      nextCursor = page.next_cursor ?? '';
      health = status;
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Unable to load audit log');
      events = [];
    } finally {
      loading = false;
    }
  }

  function applyFilters() {
    load();
  }

  function formatDate(value: string) {
    return new Date(value).toLocaleString();
  }

  function healthLabel(value: AuditHealth['clickhouse']) {
    return value === 'operational' ? 'Operational' : value === 'degraded' ? 'Degraded' : value === 'disabled' ? 'Disabled' : 'Unavailable';
  }

  onMount(() => load());
</script>

<div class="page-head">
  <div>
    <p class="eyebrow">PLATFORM OPERATIONS</p>
    <h1>Audit log</h1>
    <p class="lede">Cross-tenant security and operational history.</p>
  </div>
</div>

<section class="filters card">
  <label>Start date<input type="date" bind:value={startDate} /></label>
  <label>End date<input type="date" bind:value={endDate} /></label>
  <label>Tenant ID<input bind:value={tenantID} placeholder="All tenants" /></label>
  <label>Actor ID<input bind:value={actorID} placeholder="Any actor" /></label>
  <label>Action<input bind:value={action} placeholder="Any action" /></label>
  <label>Resource<input bind:value={resourceType} placeholder="Any resource" /></label>
  <label>Outcome<select bind:value={outcome}><option value="">All outcomes</option><option value="success">Success</option><option value="denied">Denied</option><option value="failure">Failure</option></select></label>
  <div class="filter-actions"><button class="btn" type="button" onclick={applyFilters} disabled={loading}>Apply</button><button class="btn ghost" type="button" onclick={() => load()} disabled={loading}>Reset view</button></div>
</section>

<section class="health card" aria-live="polite">
  <div><span class="status-dot" class:warn={health?.clickhouse !== 'operational'}></span><strong>ClickHouse {health ? healthLabel(health.clickhouse) : 'Checking'}</strong></div>
  <span>{health?.mode ?? 'spool'} mode</span>
  <span>{health?.pending_files ?? 0} pending files</span>
  <span>{health?.queue_depth ?? 0} queued</span>
  <span>Last transfer {health?.last_successful_transfer ? formatDate(health.last_successful_transfer) : 'not recorded'}</span>
</section>

<section class="card table-card">
  <div class="section-head"><div><h2>Events</h2><p>{events.length ? `${events.length} events in this page` : 'No events in this range.'}</p></div><button class="btn ghost" type="button" onclick={() => load()} disabled={loading}>Refresh</button></div>
  {#if loading}
    <p class="muted">Loading audit events...</p>
  {:else if !events.length}
    <div class="empty"><strong>No audit events found.</strong><span>Try a wider date range or fewer filters.</span></div>
  {:else}
    <div class="table-wrap">
      <table>
        <thead><tr><th>Time</th><th>Tenant</th><th>Actor</th><th>Action</th><th>Resource</th><th>Outcome</th><th></th></tr></thead>
        <tbody>
          {#each events as event (event.event_id)}
            <tr>
              <td class="time">{formatDate(event.occurred_at)}</td>
              <td>{event.tenant_id || 'Platform'}</td>
              <td><strong>{event.actor_id}</strong><small>{event.actor_type}</small></td>
              <td><code>{event.action}</code></td>
              <td>{event.resource_type}{event.resource_id ? ` · ${event.resource_id}` : ''}</td>
              <td><span class:success={event.outcome === 'success'} class:denied={event.outcome === 'denied'} class="outcome">{event.outcome}</span></td>
              <td><button class="detail" type="button" onclick={() => (expandedID = expandedID === event.event_id ? '' : event.event_id)} aria-label="Toggle event metadata">{expandedID === event.event_id ? '−' : '+'}</button></td>
            </tr>
            {#if expandedID === event.event_id}
              <tr class="detail-row"><td colspan="7"><pre>{JSON.stringify(event.metadata ?? {}, null, 2)}</pre><small>Request {event.request_id} · {event.source}</small></td></tr>
            {/if}
          {/each}
        </tbody>
      </table>
    </div>
    {#if nextCursor}<div class="pager"><button class="btn ghost" type="button" onclick={() => load(nextCursor)} disabled={loading}>Next page</button></div>{/if}
  {/if}
</section>

<style>
  .page-head { margin-bottom: 22px; }
  .eyebrow { margin: 0 0 8px; color: var(--cyan); font-size: 11px; letter-spacing: .16em; font-weight: 700; }
  h1 { margin: 0; font-size: 32px; letter-spacing: 0; }
  .lede { margin: 7px 0 0; color: var(--muted); }
  .filters { display: grid; grid-template-columns: repeat(4, minmax(130px, 1fr)); gap: 14px; margin-bottom: 14px; }
  label { display: grid; gap: 6px; color: var(--muted); font-size: 12px; }
  input, select { width: 100%; min-height: 38px; border: 1px solid var(--line); border-radius: 8px; background: rgb(4 10 22 / 74%); color: var(--ink); padding: 8px 10px; }
  .filter-actions { display: flex; align-items: end; gap: 8px; }
  .health { display: flex; align-items: center; gap: 18px; flex-wrap: wrap; margin-bottom: 14px; color: var(--muted); font-size: 12px; }
  .health div { display: flex; align-items: center; gap: 8px; color: var(--ink); }
  .status-dot { width: 8px; height: 8px; border-radius: 50%; background: var(--success); box-shadow: 0 0 10px var(--success); }
  .status-dot.warn { background: #f5b94c; box-shadow: 0 0 10px #f5b94c; }
  .table-card { overflow: hidden; }
  .section-head { display: flex; align-items: center; justify-content: space-between; gap: 12px; margin-bottom: 14px; }
  h2 { margin: 0; font-size: 18px; }
  .section-head p, .muted { margin: 5px 0 0; color: var(--muted); font-size: 12px; }
  .table-wrap { overflow-x: auto; }
  table { width: 100%; border-collapse: collapse; min-width: 880px; font-size: 12px; }
  th, td { padding: 11px 9px; border-bottom: 1px solid var(--line); text-align: left; vertical-align: top; }
  th { color: var(--muted); font-size: 11px; font-weight: 600; }
  td small { display: block; color: var(--muted); margin-top: 3px; }
  .time { color: var(--muted); white-space: nowrap; }
  code { color: #a9c9ff; font-size: 11px; }
  .outcome { color: #f5b94c; }
  .outcome.success { color: var(--success); }
  .outcome.denied { color: var(--danger); }
  .detail { width: 25px; height: 25px; border: 1px solid var(--line); border-radius: 6px; color: var(--ink); background: transparent; }
  .detail-row td { background: rgb(9 17 32 / 70%); }
  pre { margin: 0 0 7px; color: #b8c7e2; white-space: pre-wrap; font-size: 11px; }
  .empty { display: grid; gap: 7px; padding: 35px 0; color: var(--muted); }
  .empty strong { color: var(--ink); }
  .pager { display: flex; justify-content: flex-end; padding-top: 14px; }
  @media (max-width: 820px) { .filters { grid-template-columns: repeat(2, minmax(130px, 1fr)); } }
  @media (max-width: 520px) { .filters { grid-template-columns: 1fr; } .filter-actions { align-items: stretch; } }
</style>
