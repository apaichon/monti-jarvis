<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import {
    assignTenantEntitlement,
    getTenantEntitlement,
    listPackages,
    revokeTenantEntitlement,
    type Entitlement,
    type Package
  } from '$lib/api/packages';
  import { ApiError } from '$lib/api/http';

  const tenantId = $derived($page.params.id);

  let entitlement = $state<Entitlement | null>(null);
  let packages = $state<Package[]>([]);
  let selectedPackage = $state('');
  let error = $state('');
  let loading = $state(true);
  let assigning = $state(false);
  let revoking = $state(false);
  let showRevoke = $state(false);
  let noEntitlement = $state(false);

  async function load() {
    loading = true;
    error = '';
    noEntitlement = false;
    try {
      const pkgRes = await listPackages('active');
      packages = pkgRes.packages;
      if (!selectedPackage && packages[0]) selectedPackage = packages[0].id;
      entitlement = await getTenantEntitlement(tenantId);
    } catch (err) {
      if (err instanceof ApiError && err.status === 404) {
        noEntitlement = true;
      } else {
        error = err instanceof ApiError ? err.message : 'Failed to load entitlement';
      }
    } finally {
      loading = false;
    }
  }

  onMount(load);

  async function assign() {
    if (!selectedPackage) return;
    assigning = true;
    error = '';
    try {
      entitlement = await assignTenantEntitlement(tenantId, selectedPackage);
      noEntitlement = false;
    } catch (err) {
      error = err instanceof ApiError ? err.message : 'Assign failed';
    } finally {
      assigning = false;
    }
  }

  async function revoke() {
    revoking = true;
    error = '';
    try {
      await revokeTenantEntitlement(tenantId);
      entitlement = null;
      noEntitlement = true;
      showRevoke = false;
    } catch (err) {
      error = err instanceof ApiError ? err.message : 'Revoke failed';
    } finally {
      revoking = false;
    }
  }

  function rulesSummary(rules: Record<string, boolean | number>) {
    return Object.entries(rules)
      .map(([k, v]) => `${k}: ${v}`)
      .join(' · ');
  }
</script>

<h1 style="margin:0 0 20px;font-size:24px">Tenant entitlement — {tenantId}</h1>

{#if error}
  <p class="error" style="margin-bottom:12px">{error}</p>
{/if}

{#if loading}
  <p style="color:var(--muted)">Loading…</p>
{:else}
  <div class="card" style="margin-bottom:16px">
    <h2 style="margin:0 0 16px;font-size:16px">Current</h2>
    {#if entitlement}
      <div class="field">
        <label>Package</label>
        <div>{entitlement.package.name} ({entitlement.package.id})</div>
      </div>
      <div class="field">
        <label>Status</label>
        <span class="badge success">{entitlement.status}</span>
      </div>
      <div class="field">
        <label>Schema</label>
        <div>{entitlement.rules_schema_id}</div>
      </div>
      <div class="field">
        <label>Rules</label>
        <div style="font-size:13px;color:var(--muted)">{rulesSummary(entitlement.rules)}</div>
      </div>
    {:else if noEntitlement}
      <p style="color:var(--muted);margin:0">No active entitlement for this tenant.</p>
    {/if}
  </div>

  <div class="card" style="margin-bottom:16px">
    <h2 style="margin:0 0 16px;font-size:16px">Assign package</h2>
    <div class="field">
      <label for="pkg">Package</label>
      <select id="pkg" bind:value={selectedPackage}>
        {#each packages as pkg (pkg.id)}
          <option value={pkg.id}>{pkg.name} ({pkg.id})</option>
        {/each}
      </select>
    </div>
    <button class="btn" type="button" disabled={assigning || !selectedPackage} onclick={assign}>
      {assigning ? 'Assigning…' : 'Assign to tenant'}
    </button>
  </div>

  {#if entitlement}
    <button class="btn danger" type="button" onclick={() => (showRevoke = true)}>Revoke entitlement</button>
  {/if}
{/if}

{#if showRevoke}
  <div class="modal-backdrop" role="presentation" onclick={() => (showRevoke = false)}>
    <div class="card modal" role="dialog" onclick={(e) => e.stopPropagation()}>
      <h3 style="margin:0 0 12px">Revoke entitlement?</h3>
      <p style="color:var(--muted);font-size:14px">This sets the active entitlement to revoked.</p>
      <div style="display:flex;gap:10px;justify-content:flex-end;margin-top:16px">
        <button class="btn ghost" type="button" onclick={() => (showRevoke = false)}>Cancel</button>
        <button class="btn danger" type="button" disabled={revoking} onclick={revoke}>
          {revoking ? 'Revoking…' : 'Revoke'}
        </button>
      </div>
    </div>
  </div>
{/if}