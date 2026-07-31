<script lang="ts">
  import { onMount } from 'svelte';
  import { ApiError } from '$lib/api/http';
  import {
    listReferralRedemptions,
    reverseReferralRedemption,
    type ReferralRedemption
  } from '$lib/api/referrals';
  import { feedback } from '$lib/feedback.svelte';

  let items = $state<ReferralRedemption[]>([]);
  let tenantFilter = $state('');
  let codeFilter = $state('');
  let statusFilter = $state('');
  let loading = $state(true);
  let reversing = $state('');

  async function load() {
    loading = true;
    try {
      const response = await listReferralRedemptions({
        tenantId: tenantFilter.trim(),
        code: codeFilter.trim(),
        status: statusFilter
      });
      items = response.redemptions;
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load referral redemptions');
    } finally {
      loading = false;
    }
  }

  async function reverse(item: ReferralRedemption) {
    const reason = window.prompt('Reason for reversing this bonus grant?');
    if (reason === null) return;
    reversing = item.id;
    try {
      await reverseReferralRedemption(item.id, reason.trim() || 'platform reversal');
      feedback.success('Referral bonus reversed');
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to reverse referral bonus');
    } finally {
      reversing = '';
    }
  }

  onMount(load);
</script>

<svelte:head><title>Referral redemptions | Monti Admin</title></svelte:head>

<div class="page-head">
  <div>
    <p>Growth</p>
    <h1>Referral redemptions</h1>
  </div>
  <button class="btn ghost" type="button" onclick={load}>Refresh</button>
</div>

<section class="card filters">
  <label>
    Tenant ID
    <input bind:value={tenantFilter} placeholder="Redeemer or referrer tenant" />
  </label>
  <label>
    Referral code
    <input bind:value={codeFilter} placeholder="ref_..." />
  </label>
  <label>
    Status
    <select bind:value={statusFilter}>
      <option value="">All</option>
      <option value="applied">Applied</option>
      <option value="reversed">Reversed</option>
    </select>
  </label>
  <button class="btn" type="button" onclick={load}>Apply filters</button>
</section>

<section class="card">
  {#if loading}
    <p class="muted">Loading referral redemptions...</p>
  {:else if items.length === 0}
    <p class="muted">No referral redemptions found.</p>
  {:else}
    <div class="table-wrap">
      <table class="table">
        <thead>
          <tr>
            <th>Code</th>
            <th>Redeemer</th>
            <th>Referrer</th>
            <th>Applied</th>
            <th>Status</th>
            <th>Action</th>
          </tr>
        </thead>
        <tbody>
          {#each items as item (item.id)}
            <tr>
              <td>{item.referral_code}</td>
              <td>{item.redeemer_tenant_id}</td>
              <td>{item.referrer_tenant_id}</td>
              <td>{new Date(item.applied_at).toLocaleString()}</td>
              <td><span class:reversed={item.status === 'reversed'}>{item.status}</span></td>
              <td>
                {#if item.status === 'applied'}
                  <button
                    class="btn danger compact"
                    type="button"
                    disabled={reversing === item.id}
                    onclick={() => void reverse(item)}
                  >
                    {reversing === item.id ? 'Reversing...' : 'Reverse'}
                  </button>
                {:else}
                  <span class="muted">Completed</span>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</section>

<style>
  .page-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 16px;
    margin-bottom: 20px;
  }
  h1, p { margin: 0; }
  h1 { font-size: 24px; }
  .page-head p { margin-bottom: 5px; color: var(--cyan); font-size: 11px; text-transform: uppercase; }
  .filters {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr)) auto;
    gap: 12px;
    align-items: end;
    margin-bottom: 16px;
  }
  label { display: grid; gap: 6px; color: var(--muted); font-size: 11px; }
  input, select {
    min-width: 0;
    padding: 10px 11px;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: rgb(7 13 26 / 72%);
    color: var(--ink);
  }
  .table-wrap { overflow-x: auto; }
  .muted { color: var(--muted); }
  .reversed { color: var(--muted); }
  .compact { padding: 7px 10px; font-size: 11px; }
  @media (max-width: 900px) {
    .filters { grid-template-columns: 1fr 1fr; }
  }
  @media (max-width: 620px) {
    .filters { grid-template-columns: 1fr; }
  }
</style>
