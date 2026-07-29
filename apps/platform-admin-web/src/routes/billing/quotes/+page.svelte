<script lang="ts">
  import { onMount } from 'svelte';
  import { feedback } from '$lib/feedback.svelte';
  import {
    listDedicatedQuotes,
    transitionDedicatedQuote,
    type DedicatedQuote,
    type QuoteTransition
  } from '$lib/api/commercial';

  let quotes = $state<DedicatedQuote[]>([]);
  let selected = $state<DedicatedQuote | null>(null);
  let loading = $state(true);
  let saving = $state(false);
  let statusFilter = $state('');
  let capacityNotes = $state('');
  let quoteAmount = $state(0);
  let quoteCurrency = $state('THB');
  let quoteValidDays = $state(30);

  onMount(load);

  async function load() {
    loading = true;
    try {
      const result = await listDedicatedQuotes({ status: statusFilter });
      quotes = result.quotes;
      if (selected) selected = quotes.find((item) => item.id === selected?.id) ?? null;
    } catch (error) {
      feedback.error(error instanceof Error ? error.message : 'Unable to load quotation requests');
    } finally {
      loading = false;
    }
  }

  function selectQuote(quote: DedicatedQuote) {
    selected = quote;
    capacityNotes =
      typeof quote.capacity_snapshot?.review_notes === 'string'
        ? quote.capacity_snapshot.review_notes
        : '';
    quoteAmount = quote.quoted_amount_cents ? quote.quoted_amount_cents / 100 : 0;
    quoteCurrency = quote.currency || 'THB';
  }

  function allowedActions(status: string): Array<{ status: string; label: string; danger?: boolean }> {
    switch (status) {
      case 'submitted':
        return [{ status: 'under_review', label: 'Begin review' }, { status: 'withdrawn', label: 'Withdraw', danger: true }];
      case 'under_review':
        return [{ status: 'capacity_confirmed', label: 'Confirm capacity' }, { status: 'rejected', label: 'Reject', danger: true }];
      case 'capacity_confirmed':
        return [{ status: 'quoted', label: 'Issue quotation' }, { status: 'rejected', label: 'Reject', danger: true }];
      case 'quoted':
        return [
          { status: 'accepted', label: 'Record acceptance' },
          { status: 'expired', label: 'Expire', danger: true },
          { status: 'rejected', label: 'Reject', danger: true }
        ];
      case 'accepted':
        return [{ status: 'provisioning', label: 'Start provisioning' }];
      case 'provisioning':
        return [{ status: 'active', label: 'Activate service' }, { status: 'rejected', label: 'Reject', danger: true }];
      default:
        return [];
    }
  }

  async function applyTransition(target: string) {
    if (!selected) return;
    const input: QuoteTransition = { status: target };
    if (target === 'capacity_confirmed') {
      if (!capacityNotes.trim()) {
        feedback.error('Capacity review notes are required.');
        return;
      }
      input.capacity_snapshot = {
        review_notes: capacityNotes.trim(),
        preferred_region: selected.preferred_region,
        expected_concurrency: selected.expected_concurrency,
        confirmed_at: new Date().toISOString()
      };
    }
    if (target === 'quoted') {
      if (!Number.isFinite(quoteAmount) || quoteAmount <= 0) {
        feedback.error('Enter a positive quotation amount.');
        return;
      }
      const expiry = new Date();
      expiry.setUTCDate(expiry.getUTCDate() + Math.max(1, quoteValidDays));
      input.quoted_amount_cents = Math.round(quoteAmount * 100);
      input.currency = quoteCurrency.trim() || 'THB';
      input.quote_expires_at = expiry.toISOString();
    }
    saving = true;
    try {
      const updated = await transitionDedicatedQuote(selected.id, input);
      quotes = quotes.map((item) => (item.id === updated.id ? updated : item));
      selected = updated;
      feedback.success(`Quotation moved to ${updated.status.replaceAll('_', ' ')}.`);
    } catch (error) {
      feedback.error(error instanceof Error ? error.message : 'Unable to update quotation');
    } finally {
      saving = false;
    }
  }

  function money(cents: number | undefined, currency: string): string {
    if (cents == null) return '—';
    return `${(cents / 100).toLocaleString(undefined, { maximumFractionDigits: 2 })} ${currency}`;
  }

  function date(value: string | undefined): string {
    return value ? new Date(value).toLocaleString() : '—';
  }
</script>

<div class="page">
  <header>
    <div>
      <p class="eyebrow">COMMERCIAL OPERATIONS</p>
      <h1>Dedicated VM quotations</h1>
      <p class="muted">Review requirements, verify capacity, issue terms, and explicitly activate provisioned service.</p>
    </div>
    <label>
      Status
      <select bind:value={statusFilter} onchange={load}>
        <option value="">All</option>
        <option value="submitted">Submitted</option>
        <option value="under_review">Under review</option>
        <option value="capacity_confirmed">Capacity confirmed</option>
        <option value="quoted">Quoted</option>
        <option value="accepted">Accepted</option>
        <option value="provisioning">Provisioning</option>
        <option value="active">Active</option>
        <option value="rejected">Rejected</option>
        <option value="expired">Expired</option>
      </select>
    </label>
  </header>

  <div class="workspace">
    <section class="card queue">
      {#if loading}
        <p class="muted">Loading requests…</p>
      {:else if !quotes.length}
        <p class="muted">No quotation requests match this filter.</p>
      {:else}
        {#each quotes as quote (quote.id)}
          <button type="button" class:selected={selected?.id === quote.id} onclick={() => selectQuote(quote)}>
            <span>
              <strong>{quote.company_legal_name}</strong>
              <small>{quote.package_name} · {quote.tenant_id}</small>
            </span>
            <span>
              <em>{quote.status.replaceAll('_', ' ')}</em>
              <small>{quote.expected_concurrency.toLocaleString()} concurrent</small>
            </span>
          </button>
        {/each}
      {/if}
    </section>

    <section class="card detail">
      {#if !selected}
        <p class="muted">Select a quotation request to review it.</p>
      {:else}
        <div class="detail-head">
          <div>
            <p class="eyebrow">{selected.id}</p>
            <h2>{selected.company_legal_name}</h2>
            <p class="muted">{selected.package_name} · {selected.status.replaceAll('_', ' ')}</p>
          </div>
          <span class="status">{selected.status.replaceAll('_', ' ')}</span>
        </div>

        <div class="facts">
          <div><span>Tenant</span><strong>{selected.tenant_id}</strong></div>
          <div><span>Expected concurrency</span><strong>{selected.expected_concurrency.toLocaleString()}</strong></div>
          <div><span>Region</span><strong>{selected.preferred_region || 'No preference'}</strong></div>
          <div><span>Company size</span><strong>{selected.company_size || '—'}</strong></div>
          <div><span>Contact</span><strong>{selected.contact_name}</strong><small>{selected.contact_email} · {selected.contact_phone}</small></div>
          <div><span>Tax / registration</span><strong>{selected.tax_registration_id || '—'}</strong></div>
          <div><span>Quoted amount</span><strong>{money(selected.quoted_amount_cents, selected.currency)}</strong></div>
          <div><span>Quote expires</span><strong>{date(selected.quote_expires_at)}</strong></div>
        </div>

        {#if selected.notes}
          <div class="notes"><span>Tenant notes</span><p>{selected.notes}</p></div>
        {/if}

        {#if selected.status === 'under_review'}
          <label class="field">
            Capacity review notes
            <textarea rows="4" bind:value={capacityNotes} placeholder="Host pool, region, resources, capacity test, and operator notes"></textarea>
          </label>
        {/if}

        {#if selected.status === 'capacity_confirmed'}
          <div class="quote-fields">
            <label>Amount <input type="number" min="0.01" step="0.01" bind:value={quoteAmount} /></label>
            <label>Currency <input maxlength="8" bind:value={quoteCurrency} /></label>
            <label>Valid days <input type="number" min="1" max="365" bind:value={quoteValidDays} /></label>
          </div>
        {/if}

        {#if allowedActions(selected.status).length}
          <div class="actions">
            {#each allowedActions(selected.status) as action}
              <button class:danger={action.danger} class="btn" type="button" disabled={saving} onclick={() => applyTransition(action.status)}>
                {saving ? 'Saving…' : action.label}
              </button>
            {/each}
          </div>
        {/if}

        <p class="invariant">No tenant payment gateway is called by a quotation transition. Activation is an explicit post-provisioning action.</p>
      {/if}
    </section>
  </div>
</div>

<style>
  .page { max-width: 1240px; margin: 0 auto; padding: 28px; }
  header { display: flex; justify-content: space-between; align-items: end; gap: 20px; margin-bottom: 20px; }
  h1, h2, p { margin-top: 0; }
  h1 { margin-bottom: 6px; font-size: 25px; }
  h2 { margin-bottom: 5px; font-size: 20px; }
  .eyebrow { margin-bottom: 6px; color: var(--muted); font-size: 10px; font-weight: 700; letter-spacing: .12em; }
  .muted { color: var(--muted); font-size: 12px; }
  header label, .field, .quote-fields label { display: grid; gap: 6px; color: var(--muted); font-size: 11px; }
  select, input, textarea { border: 1px solid var(--line); border-radius: 8px; padding: 9px 10px; color: var(--ink); background: rgb(4 10 21 / 80%); }
  .workspace { display: grid; grid-template-columns: minmax(300px, .8fr) minmax(480px, 1.4fr); gap: 16px; align-items: start; }
  .queue { display: grid; gap: 7px; padding: 10px; }
  .queue button { display: flex; justify-content: space-between; gap: 12px; border: 1px solid transparent; border-radius: 9px; padding: 12px; color: var(--ink); text-align: left; background: rgb(6 13 27 / 58%); }
  .queue button:hover, .queue button.selected { border-color: var(--cyan); background: rgb(22 199 255 / 7%); }
  .queue button > span { display: grid; gap: 4px; }
  .queue button > span:last-child { justify-items: end; text-align: right; }
  .queue small { color: var(--muted); font-size: 9px; }
  .queue em, .status { color: var(--cyan); font-size: 10px; font-style: normal; text-transform: capitalize; }
  .detail { min-height: 380px; }
  .detail-head { display: flex; justify-content: space-between; gap: 14px; }
  .status { height: fit-content; border: 1px solid rgb(22 199 255 / 28%); border-radius: 999px; padding: 6px 9px; background: rgb(22 199 255 / 7%); }
  .facts { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin: 18px 0; }
  .facts > div { display: grid; gap: 4px; border: 1px solid var(--line); border-radius: 9px; padding: 10px; }
  .facts span, .notes span { color: var(--muted); font-size: 9px; text-transform: uppercase; }
  .facts strong { font-size: 12px; }
  .facts small { color: var(--muted); font-size: 10px; }
  .notes { margin-bottom: 16px; border-left: 2px solid var(--cyan); padding-left: 11px; }
  .notes p { margin: 5px 0 0; color: var(--muted); font-size: 12px; white-space: pre-wrap; }
  .field textarea { width: 100%; }
  .quote-fields { display: grid; grid-template-columns: 1fr .6fr .6fr; gap: 10px; }
  .actions { display: flex; flex-wrap: wrap; justify-content: flex-end; gap: 8px; margin-top: 18px; }
  .actions .danger { border-color: rgb(255 92 122 / 45%); background: rgb(255 92 122 / 12%); color: var(--danger); box-shadow: none; }
  .invariant { margin: 18px 0 0; border-top: 1px solid var(--line); padding-top: 12px; color: var(--muted); font-size: 9px; line-height: 1.5; }
  @media (max-width: 900px) { .workspace { grid-template-columns: 1fr; } }
  @media (max-width: 620px) {
    .page { padding: 18px 12px; }
    header { display: grid; align-items: start; }
    .facts, .quote-fields { grid-template-columns: 1fr; }
    .queue button { display: grid; }
    .queue button > span:last-child { justify-items: start; text-align: left; }
  }
</style>
