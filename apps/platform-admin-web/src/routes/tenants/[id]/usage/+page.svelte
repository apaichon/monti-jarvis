<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { page } from '$app/stores';
  import { getTenantUsage, type TenantUsage } from '$lib/api/usage';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  const tenantId = $derived($page.params.id);

  let data = $state<TenantUsage | null>(null);
  let loading = $state(true);

  async function load() {
    loading = true;
    try {
      data = await getTenantUsage(tenantId);
    } catch (err) {
      data = null;
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load usage');
    } finally {
      loading = false;
    }
  }

  onMount(load);

  function pct(usage: number, limit: number) {
    if (limit <= 0) return usage > 0 ? 100 : 0;
    return Math.min(100, Math.round((usage / limit) * 100));
  }

  function barColor(usage: number, limit: number) {
    const p = pct(usage, limit);
    if (p >= 100) return 'var(--danger, #c0392b)';
    if (p >= 80) return 'var(--warn, #d68910)';
    return 'var(--accent, #2a6df4)';
  }

  const rows = $derived(() => {
    if (!data?.limits) return [];
    const l = data.limits;
    const u = data.usage;
    return [
      { label: 'AI employees', usage: u.ai_employees, limit: l.max_ai_employees },
      { label: 'Call minutes (month)', usage: u.monthly_call_minutes, limit: l.max_monthly_call_minutes },
      { label: 'KM documents', usage: u.km_documents, limit: l.max_km_documents },
      { label: 'Concurrent calls', usage: u.concurrent_calls, limit: l.max_concurrent_calls }
    ];
  });
</script>

<div style="display:flex;justify-content:space-between;align-items:center;gap:12px;margin-bottom:20px;flex-wrap:wrap">
  <div>
    <h1 style="margin:0;font-size:24px">Usage — {tenantId}</h1>
    <p style="margin:6px 0 0;color:var(--muted);font-size:13px">
      <a class="link" href="{base}/tenants">← Tenants</a>
      ·
      <a class="link" href="{base}/tenants/{tenantId}/entitlement">Entitlement</a>
      ·
      <a class="link" href="{base}/tenants/{tenantId}/avatars">Avatars</a>
      ·
      <a class="link" href="{base}/tenants/{tenantId}/kyc">KYC</a>
    </p>
  </div>
  <button class="btn secondary" type="button" onclick={load} disabled={loading}>
    {loading ? 'Loading…' : 'Refresh'}
  </button>
</div>

{#if loading && !data}
  <p style="color:var(--muted)">Loading…</p>
{:else if data}
  <div class="card" style="margin-bottom:16px">
    <div style="display:flex;flex-wrap:wrap;gap:16px;align-items:center">
      <div>
        <div style="font-size:12px;color:var(--muted)">Package</div>
        <strong>
          {#if data.package}
            {data.package.name}
            <span style="color:var(--muted);font-weight:400">({data.package.slug})</span>
          {:else}
            —
          {/if}
        </strong>
      </div>
      <div>
        <div style="font-size:12px;color:var(--muted)">Status</div>
        <span class="badge" class:success={data.status === 'active'}>{data.status}</span>
      </div>
      <div>
        <div style="font-size:12px;color:var(--muted)">Period (UTC)</div>
        <strong>{data.period}</strong>
      </div>
    </div>
  </div>

  {#if !data.limits || data.status === 'none'}
    <div class="card">
      <p style="margin:0;color:var(--muted)">No active package — usage meters unavailable.</p>
      <p style="margin:12px 0 0">
        <a class="link" href="{base}/tenants/{tenantId}/entitlement">Assign a package</a>
      </p>
    </div>
  {:else}
    <div class="card">
      <h2 style="margin:0 0 16px;font-size:16px">Limits vs usage</h2>
      <div style="display:flex;flex-direction:column;gap:16px">
        {#each rows() as row}
          <div>
            <div style="display:flex;justify-content:space-between;font-size:13px;margin-bottom:6px">
              <span>{row.label}</span>
              <span style="color:var(--muted)">{row.usage} / {row.limit}</span>
            </div>
            <div
              style="height:10px;background:var(--border,#e5e7eb);border-radius:999px;overflow:hidden"
            >
              <div
                style="height:100%;width:{pct(row.usage, row.limit)}%;background:{barColor(
                  row.usage,
                  row.limit
                )};transition:width .2s"
              ></div>
            </div>
          </div>
        {/each}
      </div>
      <div style="margin-top:20px;display:flex;gap:24px;flex-wrap:wrap;font-size:13px">
        <div>
          <span style="color:var(--muted)">voice_enabled</span>
          <strong style="margin-left:8px">{data.limits.voice_enabled ? 'yes' : 'no'}</strong>
        </div>
        <div>
          <span style="color:var(--muted)">rag_enabled</span>
          <strong style="margin-left:8px">{data.limits.rag_enabled ? 'yes' : 'no'}</strong>
        </div>
      </div>
    </div>
  {/if}
{:else}
  <p style="color:var(--muted)">Could not load usage.</p>
{/if}
