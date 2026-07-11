<script lang="ts">
  import { base } from '$app/paths';
  import { get } from 'svelte/store';
  import { page } from '$app/stores';
  import { onDestroy, onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import {
    getCheckoutOrderId,
    getCheckoutOrderNo,
    hasRegistrationSession
  } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';
  import { getPaymentOrder, openDocumentHTML, type PaymentOrder } from '$lib/api/billing';

  let order = $state<PaymentOrder | null>(null);
  let loading = $state(true);
  let openingDoc = $state<string | null>(null);
  let pollTimer: ReturnType<typeof setInterval> | undefined;

  function orderId(): string {
    const params = get(page).url.searchParams;
    // After ChillPay: server bridge redirects with order_id + order_no (+ status/txn_id).
    // API accepts either internal id or ChillPay order_no.
    const fromQuery =
      params.get('order_id') ||
      params.get('OrderNo') ||
      params.get('order_no') ||
      params.get('orderNo');
    if (fromQuery) return fromQuery;
    return getCheckoutOrderId() || getCheckoutOrderNo() || '';
  }

  function loginNextPath(): string {
    // Preserve full return query (order_id, order_no, status) across login.
    const qs = get(page).url.search || '';
    const id = orderId();
    if (qs) return `${base}/billing/return${qs}`;
    if (id) return `${base}/billing/return?order_id=${encodeURIComponent(id)}`;
    return `${base}/billing/return`;
  }

  function formatAmount(cents: number, currency?: string): string {
    const amount = (cents / 100).toLocaleString(undefined, {
      minimumFractionDigits: 2,
      maximumFractionDigits: 2
    });
    if (currency === 'THB' || currency === '764') return `฿${amount}`;
    if (currency === 'USD') return `$${amount}`;
    return `${amount} ${currency ?? ''}`.trim();
  }

  function methodLabel(method?: string): string {
    if (method === 'qr_promptpay') return 'QR PromptPay';
    if (method === 'credit_card' || !method) return 'Credit Card';
    return method;
  }

  async function poll() {
    const id = orderId();
    if (!id) {
      feedback.error('Missing order reference');
      loading = false;
      return;
    }
    try {
      order = await getPaymentOrder(id);
      if (order.status !== 'pending' && pollTimer) {
        clearInterval(pollTimer);
        pollTimer = undefined;
      }
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Failed to load order');
    } finally {
      loading = false;
    }
  }

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(loginNextPath())}`);
      return;
    }
    await poll();
    // Poll while pending (callback may lag behind browser return).
    pollTimer = setInterval(poll, 2000);
  });

  onDestroy(() => {
    if (pollTimer) clearInterval(pollTimer);
  });

  function statusLabel(status: string): { text: string; tone: 'pending' | 'success' | 'danger' } {
    switch (status) {
      case 'paid':
        return { text: 'Payment completed', tone: 'success' };
      case 'failed':
        return { text: 'Payment failed', tone: 'danger' };
      case 'cancelled':
        return { text: 'Payment cancelled', tone: 'danger' };
      default:
        return { text: 'Processing payment…', tone: 'pending' };
    }
  }

  async function openDoc(docType: string) {
    if (!order) return;
    openingDoc = docType;
    try {
      await openDocumentHTML(order.id, docType);
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Could not open document');
    } finally {
      openingDoc = null;
    }
  }
</script>

<div class="page-wrap" style="display:flex;justify-content:center;padding-top:48px">
  <div class="card" style="width:min(480px,100%);text-align:center">
    <h1 style="margin:0 0 16px;font-size:20px">Payment status</h1>

    {#if loading && !order}
      <p style="color:var(--muted)">Loading…</p>
    {:else if order}
      {@const st = statusLabel(order.status)}
      <p
        class="status"
        style="color:{st.tone === 'success'
          ? 'var(--success)'
          : st.tone === 'danger'
            ? 'var(--danger)'
            : 'var(--muted)'}"
      >
        {st.text}
      </p>
      <p style="margin:0 0 4px;font-size:14px">
        Package: <strong>{order.package_id}</strong>
      </p>
      <p style="margin:0 0 4px;font-size:14px">
        Amount: <strong>{formatAmount(order.amount_cents, order.currency)}</strong>
      </p>
      <p style="margin:0 0 4px;font-size:13px;color:var(--muted)">
        Method: {methodLabel(order.payment_method)}
      </p>
      <p style="margin:0 0 20px;font-size:13px;color:var(--muted)">Order: {order.order_no}</p>

      {#if order.status === 'pending'}
        <p style="font-size:12px;color:var(--muted);margin:0 0 16px">
          Waiting for ChillPay confirmation. This page updates automatically.
        </p>
      {/if}

      {#if order.status === 'failed'}
        <p style="font-size:13px;color:var(--danger);margin:0 0 16px">
          Payment was not completed. You can try buying the package again.
        </p>
      {/if}

      {#if order.status === 'paid'}
        <div class="docs">
          <p style="margin:0 0 10px;font-size:13px;font-weight:600">Documents issued</p>
          <div style="display:flex;flex-wrap:wrap;gap:8px;justify-content:center">
            <button
              class="btn ghost"
              type="button"
              disabled={openingDoc === 'receipt'}
              onclick={() => openDoc('receipt')}
            >
              {openingDoc === 'receipt' ? 'Opening…' : 'View receipt'}
            </button>
            <button
              class="btn ghost"
              type="button"
              disabled={openingDoc === 'tax_invoice'}
              onclick={() => openDoc('tax_invoice')}
            >
              {openingDoc === 'tax_invoice' ? 'Opening…' : 'View tax invoice'}
            </button>
          </div>
          {#if order.documents?.length}
            <ul class="doc-list">
              {#each order.documents as d (d.id)}
                <li>
                  {d.doc_type === 'tax_invoice' ? 'Tax invoice' : 'Receipt'}
                  <span class="muted">{d.doc_number}</span>
                </li>
              {/each}
            </ul>
          {/if}
        </div>
      {/if}
    {:else}
      <p style="color:var(--muted)">Order not found.</p>
    {/if}

    <a class="btn" href="{base}/billing" style="display:inline-block;text-decoration:none;margin-top:20px"
      >Back to billing</a
    >
  </div>
</div>

<style>
  .status {
    font-size: 18px;
    margin: 0 0 12px;
    font-weight: 600;
  }
  .docs {
    margin: 8px 0 0;
    padding: 14px;
    border: 1px solid var(--border, #333);
    border-radius: 10px;
    text-align: center;
  }
  .doc-list {
    list-style: none;
    margin: 12px 0 0;
    padding: 0;
    font-size: 12px;
    color: var(--muted);
  }
  .doc-list li {
    margin: 4px 0;
  }
  .muted {
    opacity: 0.85;
    margin-left: 6px;
  }
</style>
