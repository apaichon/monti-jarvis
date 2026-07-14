<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';
  import { getSatisfactionStatistics, type SatisfactionStatistics } from '$lib/api/operations';

  let startDate = $state(localISODate());
  let endDate = $state(localISODate());
  let avatarId = $state('');
  let channel = $state<'' | 'chat' | 'voice'>('');
  let stats = $state<SatisfactionStatistics | null>(null);
  let loading = $state(true);
  let applying = $state(false);
  let loadError = $state('');

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/satisfaction`)}`);
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
      stats = await getSatisfactionStatistics({ startDate, endDate, avatarId, channel });
    } catch (err) {
      loadError = err instanceof ApiError ? err.message : 'Failed to load satisfaction statistics';
      feedback.error(loadError);
    } finally {
      loading = false;
    }
  }

  async function applyFilters(event: SubmitEvent) {
    event.preventDefault();
    applying = true;
    try {
      await load();
    } finally {
      applying = false;
    }
  }

  async function resetToday() {
    const today = localISODate();
    startDate = today;
    endDate = today;
    avatarId = '';
    channel = '';
    await load();
  }

  function score(value: number) {
    return value > 0 ? value.toFixed(2) : '—';
  }

  function percent(value: number) {
    return `${value.toFixed(0)}%`;
  }
</script>

<svelte:head><title>Satisfaction | Monti Tenant</title></svelte:head>

<div class="satisfaction-page">
  <div class="page-head">
    <div>
      <p class="eyebrow">Tenant operations</p>
      <h1>Customer satisfaction</h1>
      <p class="muted">Review how customers rate completed AI conversations.</p>
    </div>
    <span class="scope-badge">Tenant scoped</span>
  </div>

  <form class="filters card" onsubmit={applyFilters}>
    <label>Start date<input type="date" bind:value={startDate} /></label>
    <label>End date<input type="date" bind:value={endDate} /></label>
    <label>Avatar ID<input bind:value={avatarId} placeholder="Any avatar" /></label>
    <label>Channel
      <select bind:value={channel}>
        <option value="">All channels</option>
        <option value="voice">Voice</option>
        <option value="chat">Chat</option>
      </select>
    </label>
    <div class="filter-actions">
      <button class="btn" type="submit" disabled={applying}>{applying ? 'Loading…' : 'Apply'}</button>
      <button class="btn ghost" type="button" onclick={resetToday}>Today</button>
    </div>
  </form>

  {#if loading}
    <p class="muted loading">Loading satisfaction statistics…</p>
  {:else if loadError}
    <section class="card error-state">
      <strong>Unable to load satisfaction statistics.</strong>
      <span>{loadError}</span>
      <button class="btn ghost" type="button" onclick={load}>Retry</button>
    </section>
  {:else if stats}
    <section class="kpi-grid" aria-label="Satisfaction summary">
      <article class="card kpi"><span>Completed conversations</span><strong>{stats.total_completed_conversations}</strong></article>
      <article class="card kpi"><span>Reviewed</span><strong>{stats.reviewed_conversations}</strong><small>{stats.unrated_conversations} unrated</small></article>
      <article class="card kpi"><span>Review completion</span><strong>{percent(stats.review_completion_rate)}</strong></article>
      <article class="card kpi"><span>Average rating</span><strong class="stars">★ {score(stats.average_score)}</strong></article>
    </section>

    {#if stats.total_completed_conversations === 0}
      <section class="card empty"><strong>No completed conversations in this range.</strong><span>Try another date range or remove the dimension filters.</span></section>
    {:else}
      <div class="stats-grid">
        <section class="card distribution">
          <div class="section-head"><h2>Rating distribution</h2><span>{stats.reviewed_conversations} reviews</span></div>
          {#each [5, 4, 3, 2, 1] as rating}
            {@const count = stats.distribution[String(rating)] ?? 0}
            <div class="bar-row">
              <span class="rating-label">★ {rating}</span>
              <div class="bar-track"><i style={`width: ${stats.reviewed_conversations ? (count / stats.reviewed_conversations) * 100 : 0}%`}></i></div>
              <strong>{count}</strong>
            </div>
          {/each}
        </section>

        <section class="card breakdown">
          <div class="section-head"><h2>By AI employee</h2><span>{stats.by_avatar.length}</span></div>
          {#if stats.by_avatar.length === 0}
            <p class="muted">No avatar breakdown available.</p>
          {:else}
            {#each stats.by_avatar as bucket (bucket.id)}
              <div class="breakdown-row"><span><strong>{bucket.name || bucket.id || 'Unknown avatar'}</strong><small>{bucket.completed} completed · {bucket.reviewed} reviewed</small></span><b>★ {score(bucket.average_score)}</b></div>
            {/each}
          {/if}
        </section>

        <section class="card breakdown">
          <div class="section-head"><h2>By channel</h2><span>{stats.by_channel.length}</span></div>
          {#each stats.by_channel as bucket (bucket.channel)}
            <div class="breakdown-row"><span><strong>{bucket.channel}</strong><small>{bucket.completed} completed · {bucket.reviewed} reviewed</small></span><b>★ {score(bucket.average_score)}</b></div>
          {/each}
        </section>
      </div>
    {/if}
  {/if}
</div>

<style>
  .satisfaction-page { display: grid; gap: 18px; max-width: 1240px; margin: 0 auto; }
  .page-head, .section-head { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
  .page-head { align-items: start; }
  .eyebrow { margin: 0 0 6px; color: var(--cyan); font-size: 11px; letter-spacing: .12em; text-transform: uppercase; }
  h1, h2 { margin: 0; }
  h1 { font-size: 34px; }
  h2 { font-size: 16px; }
  .muted { color: var(--muted); }
  .scope-badge, .section-head > span { border: 1px solid var(--line); border-radius: 999px; padding: 6px 10px; color: var(--muted); font-size: 11px; white-space: nowrap; }
  .filters { display: grid; grid-template-columns: repeat(4, minmax(130px, 1fr)) auto; gap: 12px; align-items: end; }
  label { display: grid; gap: 6px; color: var(--muted); font-size: 11px; }
  input, select { width: 100%; border: 1px solid var(--line); border-radius: 8px; padding: 9px 10px; color: var(--ink); background: rgb(4 9 20 / 74%); }
  .filter-actions { display: flex; gap: 8px; }
  .btn { border: 1px solid rgb(74 135 255 / 46%); border-radius: 9px; padding: 9px 13px; background: linear-gradient(100deg, var(--blue), var(--violet)); color: var(--ink); font-weight: 650; white-space: nowrap; }
  .btn.ghost { background: rgb(13 23 42 / 62%); }
  .btn:disabled { opacity: .55; cursor: not-allowed; }
  .loading { padding: 24px 0; }
  .kpi-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; }
  .kpi { display: grid; gap: 7px; min-height: 112px; }
  .kpi span, .kpi small { color: var(--muted); font-size: 11px; }
  .kpi strong { font-size: 28px; }
  .stars, .rating-label { color: #ffd166; }
  .empty { display: grid; gap: 8px; padding: 32px; color: var(--muted); }
  .error-state { display: grid; gap: 10px; padding: 28px; color: var(--muted); }
  .error-state strong { color: var(--ink); }
  .error-state .btn { justify-self: start; }
  .stats-grid { display: grid; grid-template-columns: 1.1fr 1fr 1fr; gap: 12px; align-items: start; }
  .distribution, .breakdown { min-width: 0; }
  .bar-row { display: grid; grid-template-columns: 42px 1fr 24px; gap: 9px; align-items: center; margin-top: 14px; font-size: 12px; }
  .bar-track { height: 8px; overflow: hidden; border-radius: 999px; background: rgb(104 127 166 / 18%); }
  .bar-track i { display: block; height: 100%; border-radius: inherit; background: linear-gradient(90deg, var(--cyan), var(--blue)); }
  .bar-row strong { color: var(--muted); text-align: right; }
  .breakdown-row { display: flex; align-items: center; justify-content: space-between; gap: 12px; padding: 13px 0; border-bottom: 1px solid var(--line); }
  .breakdown-row:last-child { border-bottom: 0; }
  .breakdown-row span { display: grid; gap: 4px; min-width: 0; }
  .breakdown-row strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .breakdown-row small { color: var(--muted); font-size: 11px; }
  .breakdown-row b { color: #ffd166; white-space: nowrap; }
  @media (max-width: 980px) { .filters { grid-template-columns: repeat(2, minmax(130px, 1fr)); } .filter-actions { grid-column: 1 / -1; } .stats-grid { grid-template-columns: 1fr 1fr; } .distribution { grid-column: 1 / -1; } }
  @media (max-width: 640px) { .page-head { flex-direction: column; } .kpi-grid { grid-template-columns: 1fr 1fr; } .stats-grid { grid-template-columns: 1fr; } .distribution { grid-column: auto; } }
</style>
