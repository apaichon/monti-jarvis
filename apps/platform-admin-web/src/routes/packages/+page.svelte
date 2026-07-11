<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { archivePackage, listPackages, type Package } from '$lib/api/packages';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  let packages = $state<Package[]>([]);
  let statusFilter = $state('active');
  let loading = $state(true);
  let archiveTarget = $state<Package | null>(null);
  let archiving = $state(false);

  async function load() {
    loading = true;
    try {
      const res = await listPackages(statusFilter);
      packages = res.packages;
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load packages');
    } finally {
      loading = false;
    }
  }

  onMount(load);

  async function confirmArchive() {
    if (!archiveTarget) return;
    archiving = true;
    try {
      await archivePackage(archiveTarget.id);
      archiveTarget = null;
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Archive failed');
    } finally {
      archiving = false;
    }
  }

  function formatPrice(pkg: Package) {
    if (!pkg.price_cents) return '—';
    const amount = (pkg.price_cents / 100).toFixed(0);
    const period = pkg.billing_period === 'monthly' ? 'mo' : pkg.billing_period;
    switch (pkg.currency) {
      case 'THB':
      case '764':
        return `฿${amount}/${period}`;
      case 'USD':
        return `$${amount}/${period}`;
      case 'JPY':
        return `¥${amount}/${period}`;
      case 'KRW':
        return `₩${amount}/${period}`;
      case 'CNY':
        return `¥${amount}/${period}`;
      default:
        return `${amount} ${pkg.currency}/${period}`;
    }
  }
</script>

<div style="display:flex;justify-content:space-between;align-items:center;gap:12px;margin-bottom:20px;flex-wrap:wrap">
  <h1 style="margin:0;font-size:24px">Packages</h1>
  <div style="display:flex;gap:10px;align-items:center">
    <label style="font-size:13px;color:var(--muted)">
      Status
      <select bind:value={statusFilter} onchange={load} style="margin-left:8px">
        <option value="">all</option>
        <option value="active">active</option>
        <option value="draft">draft</option>
        <option value="archived">archived</option>
      </select>
    </label>
    <a class="btn" href="{base}/packages/new">+ New</a>
  </div>
</div>

<div class="card">
  {#if loading}
    <p style="color:var(--muted)">Loading…</p>
  {:else if packages.length === 0}
    <p style="color:var(--muted)">No packages found. <a class="link" href="{base}/packages/new">Create one</a></p>
  {:else}
    <table class="table">
      <thead>
        <tr>
          <th>slug</th>
          <th>name</th>
          <th>status</th>
          <th>schema</th>
          <th>price</th>
          <th>actions</th>
        </tr>
      </thead>
      <tbody>
        {#each packages as pkg (pkg.id)}
          <tr>
            <td>{pkg.slug}</td>
            <td>{pkg.name}</td>
            <td><span class="badge">{pkg.status}</span></td>
            <td>{pkg.rules_schema_id}</td>
            <td>{formatPrice(pkg)}</td>
            <td>
              <div class="row-actions">
                <a class="link" href="{base}/packages/{pkg.id}">Edit</a>
                <a class="link" href="{base}/tenants/demo/entitlement">Assign demo</a>
                {#if pkg.status !== 'archived'}
                  <button class="link" type="button" style="background:none;border:none;padding:0;color:var(--danger)" onclick={() => (archiveTarget = pkg)}>
                    Archive
                  </button>
                {/if}
              </div>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
  {/if}
</div>

{#if archiveTarget}
  <div class="modal-backdrop" role="presentation" onclick={() => (archiveTarget = null)}>
    <div class="card modal" role="dialog" onclick={(e) => e.stopPropagation()}>
      <h3 style="margin:0 0 12px">Archive "{archiveTarget.name}"?</h3>
      <p style="color:var(--muted);font-size:14px">
        Active entitlements block archive (409).
      </p>
      <div style="display:flex;gap:10px;justify-content:flex-end;margin-top:16px">
        <button class="btn ghost" type="button" onclick={() => (archiveTarget = null)}>Cancel</button>
        <button class="btn danger" type="button" disabled={archiving} onclick={confirmArchive}>
          {archiving ? 'Archiving…' : 'Archive'}
        </button>
      </div>
    </div>
  </div>
{/if}