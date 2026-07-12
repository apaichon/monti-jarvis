<script lang="ts">
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';
  import {
    checkoutPackage,
    getTenantPackages,
    type CurrentEntitlement,
    type PackageSummary,
    type PaymentMethodOption
  } from '$lib/api/billing';

  let packages = $state<PackageSummary[]>([]);
  let current = $state<CurrentEntitlement>(null);
  let paymentMethods = $state<PaymentMethodOption[]>([
    { id: 'credit_card', label: 'Credit Card', channel_code: 'creditcard' },
    { id: 'qr_promptpay', label: 'QR PromptPay', channel_code: 'bank_qrcode' }
  ]);
  let loading = $state(true);
  let buying = $state(false);

  let checkoutPkg = $state<PackageSummary | null>(null);
  let selectedMethod = $state('credit_card');

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/billing`)}`);
      return;
    }
    await load();
  });

  async function load() {
    loading = true;
    try {
      const res = await getTenantPackages();
      packages = res.packages;
      current = res.current_entitlement;
      if (res.payment_methods?.length) {
        paymentMethods = res.payment_methods;
        selectedMethod = res.payment_methods[0].id;
      }
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Failed to load packages');
    } finally {
      loading = false;
    }
  }

  function formatPrice(cents: number, currency: string): string {
    if (cents <= 0) return 'Free';
    const amount = cents / 100;
    const n = amount.toLocaleString();
    switch (currency) {
      case 'THB':
      case '764':
        return `฿${n} / mo`;
      case 'USD':
        return `$${n} / mo`;
      case 'JPY':
        return `¥${n} / mo`;
      case 'KRW':
        return `₩${n} / mo`;
      case 'CNY':
        return `¥${n} / mo`;
      default:
        return `${n} ${currency} / mo`;
    }
  }

  function ruleLine(summary: Record<string, number | boolean> | undefined): string {
    if (!summary) return '—';
    const avatars = summary.max_ai_employees;
    const minutes = summary.max_monthly_call_minutes;
    const parts: string[] = [];
    if (typeof avatars === 'number') parts.push(`${avatars} avatars`);
    if (typeof minutes === 'number') parts.push(`${minutes} min/mo`);
    return parts.join(' · ') || '—';
  }

  /** Avoid showing the same avatar/minutes line twice when description mirrors rules. */
  function packageBlurb(pkg: PackageSummary): string {
    const rules = ruleLine(pkg.rules_summary);
    const desc = (pkg.description || '').trim();
    if (!desc || desc === rules) return rules;
    return desc;
  }

  function openCheckout(pkg: PackageSummary) {
    if (current?.package_id === pkg.id) return;
    checkoutPkg = pkg;
    selectedMethod = paymentMethods[0]?.id ?? 'credit_card';
  }

  function closeCheckout() {
    if (buying) return;
    checkoutPkg = null;
  }

  async function confirmBuy() {
    if (!checkoutPkg) return;
    buying = true;
    try {
      const res = await checkoutPackage(checkoutPkg.id, selectedMethod);
      const { saveCheckoutOrder } = await import('$lib/auth/session');
      saveCheckoutOrder(res.order_id, res.order_no);
      window.location.href = res.payment_url;
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Checkout failed');
      buying = false;
    }
  }
</script>

<div class="page-wrap">
  <div style="display:flex;justify-content:space-between;align-items:flex-start;gap:12px;flex-wrap:wrap;margin-bottom:24px">
    <div>
      <h1 style="margin:0 0 8px;font-size:24px">Billing & packages</h1>
      <p style="margin:0;color:var(--muted);font-size:14px">
        Choose a plan, select Credit Card or QR PromptPay, then complete payment. Status, receipt, and tax invoice follow.
      </p>
    </div>
    <div style="display:flex;gap:8px">
      <a class="btn ghost" href="{base}/billing/documents" style="text-decoration:none">Documents</a>
      <a class="btn ghost" href="{base}/billing/tax" style="text-decoration:none">Tax profile</a>
    </div>
  </div>

  {#if loading}
    <p style="color:var(--muted)">Loading…</p>
  {:else}
    <div class="card" style="margin-bottom:24px">
      <h2 style="margin:0 0 8px;font-size:16px">Current plan</h2>
      {#if current}
        <p style="margin:0;font-size:15px">
          <strong>{current.package_name}</strong>
          <span style="color:var(--muted)"> · {current.status}</span>
        </p>
      {:else}
        <p style="margin:0;color:var(--muted)">No active package assigned.</p>
      {/if}
    </div>

    <h2 style="margin:0 0 16px;font-size:16px">Available packages</h2>
    <div class="pkg-grid">
      {#each packages as pkg (pkg.id)}
        <div class="card pkg-card">
          <h3 style="margin:0 0 4px;font-size:18px">{pkg.name}</h3>
          <p style="margin:0 0 8px;color:var(--cyan);font-size:15px">
            {formatPrice(pkg.price_cents, pkg.currency)}
          </p>
          <p style="margin:0 0 16px;color:var(--muted);font-size:13px">{packageBlurb(pkg)}</p>
          {#if current?.package_id === pkg.id}
            <button class="btn ghost" type="button" disabled>Current</button>
          {:else}
            <button class="btn" type="button" onclick={() => openCheckout(pkg)}>
              Buy {pkg.name}
            </button>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

{#if checkoutPkg}
  <div
    class="modal-backdrop"
    role="presentation"
    onclick={closeCheckout}
    onkeydown={(e) => e.key === 'Escape' && closeCheckout()}
  >
    <div
      class="card modal"
      role="dialog"
      aria-modal="true"
      aria-labelledby="checkout-title"
      tabindex="-1"
      onclick={(e) => e.stopPropagation()}
      onkeydown={(e) => e.stopPropagation()}
    >
      <h2 id="checkout-title" style="margin:0 0 8px;font-size:18px">Pay for {checkoutPkg.name}</h2>
      <p style="margin:0 0 16px;color:var(--muted);font-size:14px">
        {formatPrice(checkoutPkg.price_cents, checkoutPkg.currency)}
      </p>

      <p style="margin:0 0 10px;font-size:13px;font-weight:600">Payment method</p>
      <div class="method-list">
        {#each paymentMethods as m (m.id)}
          <label class="method-option" class:selected={selectedMethod === m.id}>
            <input type="radio" name="pay-method" value={m.id} bind:group={selectedMethod} />
            <span class="method-icon" aria-hidden="true">
              {#if m.id === 'qr_promptpay'}QR{:else}💳{/if}
            </span>
            <span>
              <strong style="display:block;font-size:14px">{m.label}</strong>
              <span style="font-size:12px;color:var(--muted)">
                {#if m.id === 'qr_promptpay'}
                  Scan PromptPay QR via mobile banking
                {:else}
                  Visa / Mastercard via ChillPay
                {/if}
              </span>
            </span>
          </label>
        {/each}
      </div>

      <div style="display:flex;gap:10px;justify-content:flex-end;margin-top:20px">
        <button class="btn ghost" type="button" disabled={buying} onclick={closeCheckout}>Cancel</button>
        <button class="btn" type="button" disabled={buying} onclick={confirmBuy}>
          {buying ? 'Redirecting…' : 'Continue to payment'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .page-wrap {
    max-width: 960px;
    margin: 0 auto;
    padding: 32px 20px 48px;
  }
  .pkg-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
    gap: 16px;
  }
  .pkg-card {
    display: flex;
    flex-direction: column;
  }
  .pkg-card .btn {
    margin-top: auto;
  }
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    display: flex;
    align-items: center;
    justify-content: center;
    z-index: 100;
    padding: 16px;
  }
  .modal {
    width: min(420px, 100%);
    outline: none;
  }
  .method-list {
    display: flex;
    flex-direction: column;
    gap: 10px;
  }
  .method-option {
    display: flex;
    align-items: flex-start;
    gap: 12px;
    padding: 12px;
    border: 1px solid var(--border, #333);
    border-radius: 10px;
    cursor: pointer;
  }
  .method-option.selected {
    border-color: var(--cyan, #22d3ee);
    box-shadow: 0 0 0 1px var(--cyan, #22d3ee);
  }
  .method-option input {
    margin-top: 4px;
  }
  .method-icon {
    width: 36px;
    height: 36px;
    border-radius: 8px;
    background: rgba(34, 211, 238, 0.12);
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 14px;
    flex-shrink: 0;
  }
</style>
