<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { listTenants, type TenantListItem } from '$lib/api/tenants';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  let tenants = $state<TenantListItem[]>([]);
  let statusFilter = $state('pending_kyc');
  let total = $state(0);
  let loading = $state(true);

  async function load() {
    loading = true;
    try {
      const res = await listTenants(statusFilter);
      tenants = res.tenants;
      total = res.total;
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load tenants');
    } finally {
      loading = false;
    }
  }

  onMount(load);

  function formatDate(value: string) {
    try {
      return new Date(value).toLocaleString();
    } catch {
      return value;
    }
  }
</script>

<div style="display:flex;justify-content:space-between;align-items:center;gap:12px;margin-bottom:20px;flex-wrap:wrap">
  <h1 style="margin:0;font-size:24px">Tenants</h1>
  <label style="font-size:13px;color:var(--muted)">
    Status
    <select bind:value={statusFilter} onchange={load} style="margin-left:8px">
      <option value="">all</option>
      <option value="pending_kyc">pending_kyc</option>
      <option value="active">active</option>
      <option value="suspended">suspended</option>
    </select>
  </label>
</div>

<div class="card">
  {#if loading}
    <p style="color:var(--muted)">Loading…</p>
  {:else if tenants.length === 0}
    <p style="color:var(--muted)">No tenants found.</p>
  {:else}
    <table class="table">
      <thead>
        <tr>
          <th>Workspace</th>
          <th>Company</th>
          <th>Status</th>
          <th>Admin email</th>
          <th>Created</th>
          <th></th>
        </tr>
      </thead>
      <tbody>
        {#each tenants as tenant}
          <tr>
            <td><code>{tenant.slug}</code></td>
            <td>{tenant.name}</td>
            <td><span class="badge">{tenant.status}</span></td>
            <td>{tenant.admin_email}</td>
            <td>{formatDate(tenant.created_at)}</td>
            <td>
              <a class="link" href="{base}/tenants/{tenant.id}/avatars">Avatars</a>
            </td>
          </tr>
        {/each}
      </tbody>
    </table>
    <p style="color:var(--muted);font-size:12px;margin:12px 0 0">{total} total</p>
  {/if}
</div>