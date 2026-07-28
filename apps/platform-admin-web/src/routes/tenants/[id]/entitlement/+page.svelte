<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { page } from '$app/stores';
  import {
    assignTenantEntitlement,
    createPromotionGrant,
    getTenantEntitlement,
    listPackages,
    listPromotionGrants,
    revokeTenantEntitlement,
    type Entitlement,
    type Package,
    type PromotionGrant
  } from '$lib/api/packages';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  const tenantId = $derived($page.params.id ?? '');

  let entitlement = $state<Entitlement | null>(null);
  let packages = $state<Package[]>([]);
  let selectedPackage = $state('');
  let promoPackage = $state('');
  let promoReason = $state('');
  let promoValidUntil = $state('');
  let promoAmount = $state(0);
  let promoIdempotency = $state('');
  let grants = $state<PromotionGrant[]>([]);
  let lastGrant = $state<PromotionGrant | null>(null);
  let loading = $state(true);
  let assigning = $state(false);
  let granting = $state(false);
  let revoking = $state(false);
  let showRevoke = $state(false);
  let noEntitlement = $state(false);

  async function load() {
    loading = true;
    noEntitlement = false;
    try {
      const pkgRes = await listPackages('active');
      packages = pkgRes.packages;
      if (!selectedPackage && packages[0]) selectedPackage = packages[0].id;
      if (!promoPackage && packages[0]) promoPackage = packages[0].id;
      try {
        entitlement = await getTenantEntitlement(tenantId);
      } catch (err) {
        if (err instanceof ApiError && err.status === 404) {
          noEntitlement = true;
          entitlement = null;
        } else {
          throw err;
        }
      }
      try {
        const grantRes = await listPromotionGrants(tenantId, 20);
        grants = grantRes.grants ?? [];
      } catch {
        grants = [];
      }
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load entitlement');
    } finally {
      loading = false;
    }
  }

  onMount(load);

  async function assign() {
    if (!selectedPackage) return;
    assigning = true;
    try {
      entitlement = await assignTenantEntitlement(tenantId, selectedPackage);
      noEntitlement = false;
      feedback.success('Package assigned to tenant (no tax invoice)');
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Assign failed');
    } finally {
      assigning = false;
    }
  }

  async function grantPromotion() {
    if (!promoPackage || !promoReason.trim()) {
      feedback.error('Package and reason are required for promotion grant');
      return;
    }
    granting = true;
    lastGrant = null;
    try {
      const body: {
        package_id: string;
        reason: string;
        valid_until?: string;
        amount_cents?: number;
        idempotency_key?: string;
      } = {
        package_id: promoPackage,
        reason: promoReason.trim(),
        amount_cents: Number.isFinite(promoAmount) ? Math.max(0, Math.floor(promoAmount)) : 0
      };
      if (promoValidUntil.trim()) body.valid_until = promoValidUntil.trim();
      if (promoIdempotency.trim()) body.idempotency_key = promoIdempotency.trim();

      const result = await createPromotionGrant(tenantId, body);
      lastGrant = result;
      if (result.entitlement) {
        entitlement = result.entitlement;
        noEntitlement = false;
      } else {
        entitlement = await getTenantEntitlement(tenantId);
        noEntitlement = false;
      }
      const grantRes = await listPromotionGrants(tenantId, 20);
      grants = grantRes.grants ?? [];
      const inv = result.tax_invoice?.doc_number ?? 'issued';
      feedback.success(
        result.replayed
          ? `Promotion grant replayed — tax invoice ${inv}`
          : `Promotion granted — active plan set, tax invoice ${inv}`
      );
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Promotion grant failed');
    } finally {
      granting = false;
    }
  }

  async function revoke() {
    revoking = true;
    try {
      await revokeTenantEntitlement(tenantId);
      entitlement = null;
      noEntitlement = true;
      showRevoke = false;
      feedback.success('Entitlement revoked');
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Revoke failed');
    } finally {
      revoking = false;
    }
  }

  function rulesSummary(rules: Record<string, boolean | number>) {
    return Object.entries(rules)
      .map(([k, v]) => `${k}: ${v}`)
      .join(' · ');
  }

  function fmtWhen(iso: string) {
    try {
      return new Date(iso).toLocaleString();
    } catch {
      return iso;
    }
  }
</script>

<h1 style="margin:0 0 20px;font-size:24px">Tenant entitlement — {tenantId}</h1>

{#if loading}
  <p style="color:var(--muted)">Loading…</p>
{:else}
  <div class="card" style="margin-bottom:16px">
    <h2 style="margin:0 0 16px;font-size:16px">Current plan</h2>
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
      {#if entitlement.valid_until}
        <div class="field">
          <label>Valid until</label>
          <div>{entitlement.valid_until}</div>
        </div>
      {/if}
    {:else if noEntitlement}
      <p style="color:var(--muted);margin:0">No active entitlement for this tenant.</p>
    {/if}
  </div>

  <div class="card" style="margin-bottom:16px;border-color:var(--accent, #38bdf8)">
    <h2 style="margin:0 0 8px;font-size:16px">Promotion grant</h2>
    <p style="margin:0 0 16px;font-size:13px;color:var(--muted)">
      Sets the selected package as the tenant <strong>active plan</strong> and
      <strong>issues a tax invoice</strong> in one atomic grant. Use this for complimentary or sales-approved packages.
    </p>
    <div class="field">
      <label for="promo-pkg">Package *</label>
      <select id="promo-pkg" bind:value={promoPackage}>
        {#each packages as pkg (pkg.id)}
          <option value={pkg.id}>{pkg.name} ({pkg.id})</option>
        {/each}
      </select>
    </div>
    <div class="field">
      <label for="promo-reason">Reason *</label>
      <input
        id="promo-reason"
        type="text"
        bind:value={promoReason}
        placeholder="e.g. Q3 sales complimentary trial"
        maxlength="500"
      />
    </div>
    <div class="field">
      <label for="promo-until">Valid until (optional)</label>
      <input id="promo-until" type="date" bind:value={promoValidUntil} />
    </div>
    <div class="field">
      <label for="promo-amount">Amount (cents, default 0)</label>
      <input id="promo-amount" type="number" min="0" step="1" bind:value={promoAmount} />
    </div>
    <div class="field">
      <label for="promo-idem">Idempotency key (optional)</label>
      <input
        id="promo-idem"
        type="text"
        bind:value={promoIdempotency}
        placeholder="promo-tenant-campaign"
      />
    </div>
    <p style="font-size:12px;color:var(--muted);margin:0 0 12px">
      ⚠ This action will (1) set the selected package as the active plan and (2) issue a tax invoice for this tenant.
    </p>
    <button
      class="btn"
      type="button"
      disabled={granting || !promoPackage || !promoReason.trim()}
      onclick={grantPromotion}
    >
      {granting ? 'Granting…' : 'Grant promotion'}
    </button>

    {#if lastGrant}
      <div
        style="margin-top:16px;padding:12px;border-radius:8px;background:var(--surface-2, #0f172a);font-size:14px"
      >
        <div>
          <strong>Active plan:</strong>
          {lastGrant.entitlement?.package?.name ?? lastGrant.package_id}
        </div>
        <div>
          <strong>Tax invoice:</strong>
          {lastGrant.tax_invoice?.doc_number ?? lastGrant.tax_invoice_number ?? '—'}
        </div>
        <div style="margin-top:8px">
          <a class="link" href="{base}/billing/receipts">Open receipts & tax invoices →</a>
        </div>
      </div>
    {/if}
  </div>

  {#if grants.length}
    <div class="card" style="margin-bottom:16px">
      <h2 style="margin:0 0 16px;font-size:16px">Recent promotion grants</h2>
      <div style="overflow-x:auto">
        <table style="width:100%;border-collapse:collapse;font-size:13px">
          <thead>
            <tr style="text-align:left;color:var(--muted)">
              <th style="padding:6px 8px">When</th>
              <th style="padding:6px 8px">Package</th>
              <th style="padding:6px 8px">Amount</th>
              <th style="padding:6px 8px">Tax invoice</th>
              <th style="padding:6px 8px">Reason</th>
            </tr>
          </thead>
          <tbody>
            {#each grants as g (g.id)}
              <tr style="border-top:1px solid var(--border, #1e293b)">
                <td style="padding:6px 8px">{fmtWhen(g.created_at)}</td>
                <td style="padding:6px 8px">{g.package_id}</td>
                <td style="padding:6px 8px">{g.amount_cents}</td>
                <td style="padding:6px 8px">{g.tax_invoice_number ?? g.tax_invoice_id ?? '—'}</td>
                <td style="padding:6px 8px">{g.reason}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    </div>
  {/if}

  <div class="card" style="margin-bottom:16px">
    <h2 style="margin:0 0 8px;font-size:16px">Assign package only</h2>
    <p style="margin:0 0 16px;font-size:13px;color:var(--muted)">
      Ops assign without a tax invoice. Prefer <strong>Promotion grant</strong> for commercial promotions.
    </p>
    <div class="field">
      <label for="pkg">Package</label>
      <select id="pkg" bind:value={selectedPackage}>
        {#each packages as pkg (pkg.id)}
          <option value={pkg.id}>{pkg.name} ({pkg.id})</option>
        {/each}
      </select>
    </div>
    <button class="btn ghost" type="button" disabled={assigning || !selectedPackage} onclick={assign}>
      {assigning ? 'Assigning…' : 'Assign to tenant (no invoice)'}
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
