<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';
  import {
    getSystemPerformance,
    type AnalyticsHealth,
    type MonitoringComponent,
    type SystemPerformanceSnapshot
  } from '$lib/api/monitoring';

  let snapshot = $state<SystemPerformanceSnapshot | null>(null);
  let loading = $state(true);
  let loadError = $state('');

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/monitoring`)}`);
      return;
    }
    await load();
  });

  async function load() {
    loading = true;
    loadError = '';
    try {
      snapshot = await getSystemPerformance();
    } catch (err) {
      loadError = err instanceof ApiError ? err.message : 'Failed to load system performance';
      feedback.error(loadError);
    } finally {
      loading = false;
    }
  }

  function titleCase(value: string) {
    return value.replaceAll('_', ' ').replace(/\b\w/g, (letter) => letter.toUpperCase());
  }

  function statusLabel(value: string) {
    const labels: Record<string, string> = {
      operational: 'Operational',
      degraded: 'Degraded',
      unavailable: 'Unavailable',
      disabled: 'Not configured',
      current: 'Current',
      stale: 'Stale'
    };
    return labels[value] ?? titleCase(value);
  }

  function statusHint(value: string) {
    const hints: Record<string, string> = {
      operational: 'ทำงานปกติ',
      degraded: 'ทำงานช้าบางส่วน',
      unavailable: 'ไม่พร้อมใช้งาน',
      disabled: 'ยังไม่ได้ตั้งค่า',
      current: 'ข้อมูลเป็นปัจจุบัน',
      stale: 'ข้อมูลล่าช้า'
    };
    return hints[value] ?? '';
  }

  function checkedAt(value?: string) {
    if (!value) return 'Not available';
    return new Date(value).toLocaleString();
  }

  function latency(component: MonitoringComponent) {
    return component.latency_ms === null ? 'Configured' : `${component.latency_ms} ms`;
  }

  function componentIcon(status: MonitoringComponent['status']) {
    return status === 'operational' ? '✓' : status === 'disabled' ? '−' : '!';
  }

  function overallIcon(status: SystemPerformanceSnapshot['overall_status']) {
    return status === 'operational' ? '✓' : status === 'degraded' ? '!' : '×';
  }

  function analyticsClass(analytics: AnalyticsHealth) {
    return `status-${analytics.status}`;
  }
</script>

<svelte:head><title>System performance | Monti Tenant</title></svelte:head>

<div class="monitoring-page">
  <div class="page-head">
    <div>
      <p class="eyebrow">Tenant operations</p>
      <h1>System performance</h1>
      <p class="muted">Service health and analytics freshness for this tenant.</p>
    </div>
    <span class="scope-badge">Tenant scoped</span>
  </div>

  {#if loading}
    <section class="card loading-state" aria-live="polite">
      <span class="loading-orb"></span>
      <div><strong>Checking service health...</strong><small>กำลังตรวจสอบสถานะระบบ</small></div>
    </section>
  {:else if loadError}
    <section class="card error-state" aria-live="assertive">
      <strong>Unable to load system performance.</strong>
      <span>{loadError}</span>
      <button class="btn ghost" type="button" onclick={load}>Retry</button>
    </section>
  {:else if snapshot}
    <section class="card overall-card">
      <div class={`overall-icon status-${snapshot.overall_status}`}>{overallIcon(snapshot.overall_status)}</div>
      <div class="overall-copy">
        <span class="section-label">Overall status</span>
        <strong>{statusLabel(snapshot.overall_status)}</strong>
        <small>{statusHint(snapshot.overall_status)} · Checked {checkedAt(snapshot.checked_at)}</small>
      </div>
      <button class="btn ghost retry" type="button" onclick={load} disabled={loading} aria-label="Retry system health check">
        ↻ Retry
      </button>
    </section>

    <section class="component-grid" aria-label="Dependency health">
      {#each snapshot.components as component (component.name)}
        <article class="card component-card">
          <div class={`component-icon status-${component.status}`}>{componentIcon(component.status)}</div>
          <div class="component-copy">
            <strong>{titleCase(component.name)}</strong>
            <span class={`status-text status-${component.status}`}>{statusLabel(component.status)}</span>
            <small>{latency(component)} · {checkedAt(component.checked_at)}</small>
          </div>
        </article>
      {/each}
    </section>

    <section class="card analytics-card">
      <div class={`analytics-icon ${analyticsClass(snapshot.analytics)}`}>◌</div>
      <div class="analytics-copy">
        <span class="section-label">Analytics freshness</span>
        <strong>{statusLabel(snapshot.analytics.status)}</strong>
        <small>{statusHint(snapshot.analytics.status)}</small>
      </div>
      <div class="analytics-times">
        <span>Last projected</span>
        <strong>{checkedAt(snapshot.analytics.last_projected_at)}</strong>
        <small>Generated {checkedAt(snapshot.analytics.generated_at)}</small>
      </div>
    </section>

    <p class="privacy-note">Only normalized status and latency are shown. Provider errors, credentials, customer data, transcripts, and audio paths are hidden.</p>
  {/if}
</div>

<style>
  .monitoring-page { display: grid; gap: 18px; max-width: 1240px; margin: 0 auto; }
  .page-head { display: flex; align-items: start; justify-content: space-between; gap: 16px; }
  .eyebrow, h1, p { margin: 0; }
  .eyebrow { margin-bottom: 6px; color: var(--cyan); font-size: 11px; letter-spacing: .12em; text-transform: uppercase; }
  h1 { font-size: 34px; letter-spacing: -.02em; }
  .muted, .privacy-note { color: var(--muted); }
  .scope-badge { border: 1px solid var(--line); border-radius: 999px; padding: 6px 10px; color: var(--muted); font-size: 11px; white-space: nowrap; }
  .overall-card, .analytics-card { display: flex; align-items: center; gap: 15px; }
  .overall-icon, .component-icon, .analytics-icon { display: grid; place-items: center; flex: 0 0 auto; border: 1px solid currentColor; border-radius: 12px; font-weight: 800; }
  .overall-icon { width: 46px; height: 46px; font-size: 22px; }
  .overall-copy, .component-copy, .analytics-copy { display: grid; gap: 4px; min-width: 0; }
  .overall-copy strong, .analytics-copy strong { font-size: 22px; }
  .overall-copy small, .component-copy small, .analytics-copy small, .analytics-times small { color: var(--muted); font-size: 11px; }
  .section-label { color: var(--muted); font-size: 11px; text-transform: uppercase; letter-spacing: .08em; }
  .retry { margin-left: auto; }
  .component-grid { display: grid; grid-template-columns: repeat(4, minmax(0, 1fr)); gap: 12px; }
  .component-card { display: flex; align-items: center; gap: 12px; min-width: 0; padding: 16px; }
  .component-icon { width: 32px; height: 32px; border-radius: 9px; font-size: 14px; }
  .component-copy strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .status-text { font-size: 12px; }
  .status-operational { color: var(--success); }
  .status-degraded, .status-stale { color: var(--warn); }
  .status-unavailable { color: var(--danger); }
  .status-disabled { color: var(--muted); }
  .analytics-icon { width: 40px; height: 40px; font-size: 20px; }
  .analytics-times { display: grid; gap: 3px; margin-left: auto; text-align: right; }
  .analytics-times span { color: var(--muted); font-size: 11px; }
  .analytics-times strong { font-size: 13px; }
  .privacy-note { font-size: 11px; text-align: right; }
  .loading-state { display: flex; align-items: center; gap: 12px; color: var(--muted); }
  .loading-state div { display: grid; gap: 4px; }
  .loading-state small { font-size: 11px; }
  .loading-orb { width: 12px; height: 12px; border: 2px solid var(--cyan); border-top-color: transparent; border-radius: 50%; animation: spin .8s linear infinite; }
  .error-state { display: grid; gap: 8px; padding: 32px; color: var(--muted); }
  .error-state strong { color: var(--ink); }
  .error-state .btn { justify-self: start; }
  @keyframes spin { to { transform: rotate(360deg); } }
  @media (max-width: 980px) { .component-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); } }
  @media (max-width: 640px) {
    .page-head { flex-direction: column; }
    .overall-card, .analytics-card { align-items: flex-start; flex-wrap: wrap; }
    .overall-copy { flex: 1 1 calc(100% - 64px); }
    .retry { flex: 1 1 100%; margin-left: 0; }
    .component-grid { grid-template-columns: 1fr; }
    .analytics-times { flex: 1 1 100%; margin-left: 0; padding-left: 55px; text-align: left; }
    .privacy-note { text-align: left; }
  }
</style>
