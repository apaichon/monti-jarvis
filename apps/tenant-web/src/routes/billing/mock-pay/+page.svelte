<script lang="ts">
  import { base } from '$app/paths';
  import { get } from 'svelte/store';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';
  import { completeMockPayment, getPaymentOrder, type PaymentOrder } from '$lib/api/billing';

  let order = $state<PaymentOrder | null>(null);
  let loading = $state(true);
  let paying = $state(false);

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/billing/mock-pay`)}`);
      return;
    }
    const orderId = get(page).url.searchParams.get('order_id');
    if (!orderId) {
      feedback.error('Missing order_id');
      loading = false;
      return;
    }
    try {
      order = await getPaymentOrder(orderId);
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Failed to load order');
    } finally {
      loading = false;
    }
  });

  function formatAmount(cents: number): string {
    if (cents <= 0) return 'Free';
    return `฿${(cents / 100).toLocaleString()}`;
  }

  function methodLabel(method?: string): string {
    if (method === 'qr_promptpay') return 'QR PromptPay';
    return 'Credit Card';
  }

  async function complete(result: 'paid' | 'failed') {
    if (!order) return;
    paying = true;
    try {
      await completeMockPayment(order.id, result);
      sessionStorage.setItem('monti_checkout_order_id', order.id);
      goto(`${base}/billing/return?order_id=${order.id}`);
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Mock payment failed');
      paying = false;
    }
  }
</script>

<div class="page-wrap" style="display:flex;justify-content:center;padding-top:48px">
  <div class="card" style="width:min(420px,100%);text-align:center">
    <h1 style="margin:0 0 8px;font-size:20px">Mock payment (dev only)</h1>
    <p style="margin:0 0 24px;color:var(--muted);font-size:13px">
      Simulates ChillPay for local testing. Choose success or failure, then return to payment status.
    </p>

    {#if loading}
      <p style="color:var(--muted)">Loading…</p>
    {:else if order}
      <p style="margin:0 0 8px;font-size:15px">
        Order <strong>{order.id}</strong>
      </p>
      <p style="margin:0 0 4px;color:var(--muted);font-size:14px">
        {order.package_id} · {formatAmount(order.amount_cents)}
      </p>
      <p style="margin:0 0 24px;color:var(--muted);font-size:13px">
        Method: {methodLabel(order.payment_method)}
      </p>
      <div style="display:flex;flex-direction:column;gap:10px">
        <button class="btn" type="button" disabled={paying} onclick={() => complete('paid')}>
          {paying ? 'Completing…' : 'Complete payment (success)'}
        </button>
        <button class="btn ghost" type="button" disabled={paying} onclick={() => complete('failed')}>
          Simulate failed payment
        </button>
      </div>
    {:else}
      <p style="color:var(--muted)">Order not found.</p>
    {/if}
  </div>
</div>
