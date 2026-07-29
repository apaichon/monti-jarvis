<script lang="ts">
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';
  import { currentPlan } from '$lib/currentPlan.svelte';
  import {
    calculatePackage,
    checkoutPackage,
    createDedicatedQuote,
    getTenantPackages,
    listDedicatedQuotes,
    type DedicatedQuote,
    type DedicatedQuoteInput,
    type PackageSummary,
    type PaymentMethodOption,
    type PriceCalculation
  } from '$lib/api/billing';

  let packages = $state<PackageSummary[]>([]);
  let quotes = $state<DedicatedQuote[]>([]);
  let paymentMethods = $state<PaymentMethodOption[]>([
    { id: 'credit_card', label: 'Credit Card', channel_code: 'creditcard' },
    { id: 'qr_promptpay', label: 'QR PromptPay', channel_code: 'bank_qrcode' }
  ]);
  let loading = $state(true);
  let buying = $state(false);
  let quoteSubmitting = $state(false);

  let checkoutPkg = $state<PackageSummary | null>(null);
  let selectedMethod = $state('credit_card');
  let billingInterval = $state<'monthly' | 'annual'>('monthly');
  let calculation = $state<PriceCalculation | null>(null);
  let calculationLoading = $state(false);

  let quotePkg = $state<PackageSummary | null>(null);
  let quoteForm = $state<DedicatedQuoteInput>(emptyQuoteForm());

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/billing`)}`);
      return;
    }
    await load();
  });

  function emptyQuoteForm(): DedicatedQuoteInput {
    return {
      package_id: '',
      company_legal_name: '',
      contact_name: '',
      contact_email: '',
      contact_phone: '',
      tax_registration_id: '',
      company_size: '',
      expected_concurrency: 1,
      preferred_region: '',
      notes: '',
      idempotency_key: ''
    };
  }

  async function load() {
    loading = true;
    try {
      const [catalog, quoteHistory] = await Promise.all([
        getTenantPackages(),
        listDedicatedQuotes()
      ]);
      packages = catalog.packages;
      quotes = quoteHistory.quotes;
      if (catalog.payment_methods?.length) {
        paymentMethods = catalog.payment_methods;
        selectedMethod = catalog.payment_methods[0].id;
      }
      await currentPlan.load(true);
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Failed to load commercial plans');
    } finally {
      loading = false;
    }
  }

  function formatMoney(cents: number, currency: string): string {
    const amount = cents / 100;
    const n = amount.toLocaleString(undefined, { maximumFractionDigits: 2 });
    switch (currency) {
      case 'THB':
      case '764':
        return `฿${n}`;
      case 'USD':
        return `$${n}`;
      case 'JPY':
      case 'CNY':
        return `¥${n}`;
      case 'KRW':
        return `₩${n}`;
      default:
        return `${n} ${currency}`;
    }
  }

  function formatPackagePrice(pkg: PackageSummary): string {
    if (pkg.price_cents <= 0) return 'Free';
    return `${formatMoney(pkg.price_cents, pkg.currency)} / mo`;
  }

  function formatDate(value: string | null | undefined): string {
    if (!value) return '—';
    const parsed = new Date(value);
    if (Number.isNaN(parsed.getTime())) return '—';
    return parsed.toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric'
    });
  }

  function formatQuotaValue(value: number | null, unit: string, unlimited = false): string {
    if (unlimited) return 'Unlimited';
    if (value == null) return 'Unavailable';
    if (unit === 'bytes') {
      if (value >= 1024 ** 3) return `${(value / 1024 ** 3).toFixed(1)} GB`;
      if (value >= 1024 ** 2) return `${(value / 1024 ** 2).toFixed(1)} MB`;
    }
    return value.toLocaleString();
  }

  function dimensionLabel(dimension: string): string {
    const labels: Record<string, string> = {
      ai_employees: 'AI employees',
      monthly_call_minutes: 'Web voice minutes',
      mobile_call_minutes: 'Mobile minutes',
      km_documents: 'Knowledge documents',
      storage_bytes: 'Knowledge storage',
      concurrent_calls: 'Concurrent calls'
    };
    return labels[dimension] ?? dimension.replaceAll('_', ' ');
  }

  function packageBlurb(pkg: PackageSummary): string {
    return (pkg.description || '').trim() || 'Commercial plan';
  }

  function isCurrent(pkg: PackageSummary): boolean {
    return currentPlan.data?.package?.id === pkg.id;
  }

  async function openCheckout(pkg: PackageSummary) {
    if (isCurrent(pkg) || pkg.quote_only || pkg.purchase_mode !== 'self_serve') return;
    checkoutPkg = pkg;
    billingInterval = 'monthly';
    selectedMethod = paymentMethods[0]?.id ?? 'credit_card';
    await refreshCalculation();
  }

  async function refreshCalculation() {
    if (!checkoutPkg) return;
    calculationLoading = true;
    calculation = null;
    try {
      calculation = await calculatePackage(checkoutPkg.id, billingInterval);
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Unable to calculate plan price');
    } finally {
      calculationLoading = false;
    }
  }

  function closeCheckout() {
    if (buying) return;
    checkoutPkg = null;
    calculation = null;
  }

  async function confirmBuy() {
    if (!checkoutPkg || !calculation?.checkout_eligible) return;
    buying = true;
    try {
      const res = await checkoutPackage(checkoutPkg.id, selectedMethod, billingInterval);
      const { saveCheckoutOrder } = await import('$lib/auth/session');
      saveCheckoutOrder(res.order_id, res.order_no);
      window.location.href = res.payment_url;
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Checkout failed');
      buying = false;
    }
  }

  function openQuote(pkg: PackageSummary) {
    if (!pkg.quote_only && pkg.purchase_mode !== 'quote') return;
    const suggested = pkg.rules_summary.max_concurrent_calls;
    quoteForm = {
      ...emptyQuoteForm(),
      package_id: pkg.id,
      expected_concurrency: typeof suggested === 'number' ? Math.max(1, suggested) : 1,
      idempotency_key: crypto.randomUUID()
    };
    quotePkg = pkg;
  }

  function closeQuote() {
    if (quoteSubmitting) return;
    quotePkg = null;
    quoteForm = emptyQuoteForm();
  }

  async function submitQuote() {
    if (!quotePkg) return;
    quoteSubmitting = true;
    try {
      const result = await createDedicatedQuote({ ...quoteForm, package_id: quotePkg.id });
      quotes = [result.quote, ...quotes.filter((item) => item.id !== result.quote.id)];
      await currentPlan.load(true);
      quotePkg = null;
      quoteForm = emptyQuoteForm();
      feedback.success('Quotation request submitted. Our platform team will review capacity and contact you.');
    } catch (err) {
      feedback.error(err instanceof Error ? err.message : 'Quotation request failed');
    } finally {
      quoteSubmitting = false;
    }
  }
</script>

<div class="page-wrap">
  <header class="page-header">
    <div>
      <h1>Billing & packages</h1>
      <p>Shared Cloud plans use secure checkout. Dedicated VM plans start with capacity and company review.</p>
    </div>
    <div class="header-actions">
      <a class="btn ghost" href="{base}/billing/documents">Documents</a>
      <a class="btn ghost" href="{base}/billing/tax">Tax profile</a>
    </div>
  </header>

  {#if loading}
    <p class="muted">Loading commercial plans…</p>
  {:else}
    <section class="card current-plan" aria-labelledby="current-plan-title">
      <div class="section-heading">
        <div>
          <span class="eyebrow">CURRENT PLAN</span>
          <h2 id="current-plan-title">{currentPlan.data?.package?.name ?? 'No active package'}</h2>
          {#if currentPlan.data?.package}
            <p>
              {currentPlan.data.package.deployment_mode === 'dedicated_vm' ? 'Dedicated VM' : 'Shared Cloud'}
              · {currentPlan.data.package.status}
              {#if currentPlan.data.subscription}
                · {currentPlan.data.subscription.billing_interval}
              {/if}
            </p>
          {:else}
            <p>Select a Shared plan or request a Dedicated quotation below.</p>
          {/if}
        </div>
        {#if currentPlan.data?.next_bill}
          <div class="next-bill">
            <span>Next bill</span>
            <strong>{formatMoney(currentPlan.data.next_bill.amount_cents, currentPlan.data.next_bill.currency)}</strong>
            <small>{formatDate(currentPlan.data.next_bill.at)}</small>
          </div>
        {:else if currentPlan.data?.billing_state === 'quote_pending'}
          <div class="status-pill quote">Quotation in review</div>
        {:else if currentPlan.data?.package}
          <div class="status-pill">No scheduled bill</div>
        {/if}
      </div>

      {#if currentPlan.data?.subscription}
        <div class="period-row">
          <span>
            <small>Current period</small>
            {formatDate(currentPlan.data.subscription.current_period_start)}
            – {formatDate(currentPlan.data.subscription.current_period_end)}
          </span>
          <span>
            <small>Billing status</small>
            {currentPlan.data.billing_state.replaceAll('_', ' ')}
          </span>
        </div>
      {/if}

      {#if currentPlan.data?.quota?.length}
        <div class="quota-grid">
          {#each currentPlan.data.quota as item (item.dimension)}
            <article class="quota-card">
              <div>
                <span>{dimensionLabel(item.dimension)}</span>
                <em class:unavailable={item.freshness === 'unavailable'}>{item.freshness}</em>
              </div>
              <strong>
                {formatQuotaValue(item.used, item.unit)}
                <small>/ {formatQuotaValue(item.limit, item.unit, item.unlimited)}</small>
              </strong>
              <p>
                {#if item.freshness === 'unavailable'}
                  Usage is temporarily unavailable
                {:else if item.unlimited}
                  No package ceiling
                {:else}
                  {formatQuotaValue(item.remaining, item.unit)} remaining
                {/if}
              </p>
              {#if item.utilization != null}
                <span class="quota-progress">
                  <i style={`width:${Math.min(100, Math.max(0, item.utilization * 100))}%`}></i>
                </span>
              {/if}
            </article>
          {/each}
        </div>
      {/if}
    </section>

    {#if currentPlan.data?.quote}
      <section class="card quote-summary">
        <div>
          <span class="eyebrow">DEDICATED REQUEST</span>
          <h2>{currentPlan.data.quote.package_name}</h2>
          <p>{currentPlan.data.quote.company_legal_name} · {currentPlan.data.quote.expected_concurrency.toLocaleString()} expected concurrent calls</p>
        </div>
        <div class="status-pill quote">{currentPlan.data.quote.status.replaceAll('_', ' ')}</div>
      </section>
    {/if}

    <h2 class="available-title">Available packages</h2>
    <div class="pkg-grid">
      {#each packages as pkg (pkg.id)}
        <article class="card pkg-card" class:dedicated={pkg.deployment === 'dedicated_vm'}>
          <div class="package-mode">
            {pkg.deployment === 'dedicated_vm' ? 'Dedicated VM' : 'Shared Cloud'}
          </div>
          <h3>{pkg.name}</h3>
          <p class="package-price">{formatPackagePrice(pkg)}</p>
          <p class="package-description">{packageBlurb(pkg)}</p>
          {#if isCurrent(pkg)}
            <button class="btn ghost" type="button" disabled>Current plan</button>
          {:else if pkg.quote_only || pkg.purchase_mode === 'quote'}
            <button class="btn quote-button" type="button" onclick={() => openQuote(pkg)}>
              Request quotation
            </button>
          {:else}
            <button class="btn" type="button" onclick={() => openCheckout(pkg)}>
              Buy {pkg.name}
            </button>
          {/if}
        </article>
      {/each}
    </div>

    {#if quotes.length}
      <section class="quote-history">
        <h2>Quotation history</h2>
        <div class="history-list">
          {#each quotes.slice(0, 6) as quote (quote.id)}
            <article class="card">
              <div>
                <strong>{quote.package_name}</strong>
                <span>{quote.company_legal_name}</span>
              </div>
              <div>
                <span class="status-pill quote">{quote.status.replaceAll('_', ' ')}</span>
                <small>{formatDate(quote.created_at)}</small>
              </div>
            </article>
          {/each}
        </div>
      </section>
    {/if}
  {/if}
</div>

{#if checkoutPkg}
  <div class="modal-backdrop" role="presentation" onclick={closeCheckout} onkeydown={(event) => event.key === 'Escape' && closeCheckout()}>
    <div class="card modal checkout-modal" role="dialog" aria-modal="true" aria-labelledby="checkout-title" tabindex="-1" onclick={(event) => event.stopPropagation()} onkeydown={(event) => event.stopPropagation()}>
      <div class="modal-heading">
        <div>
          <span class="eyebrow">SHARED CLOUD CHECKOUT</span>
          <h2 id="checkout-title">{checkoutPkg.name}</h2>
        </div>
        <button class="close-button" type="button" aria-label="Close" disabled={buying} onclick={closeCheckout}>×</button>
      </div>

      <label class="field">
        <span>Billing interval</span>
        <select bind:value={billingInterval} onchange={() => refreshCalculation()}>
          <option value="monthly">Monthly</option>
          <option value="annual">Annual · 20% catalog discount</option>
        </select>
      </label>

      {#if calculationLoading}
        <div class="calculation-card muted">Calculating server-authoritative price…</div>
      {:else if calculation}
        <div class="calculation-card">
          <div><span>Base price</span><strong>{formatMoney(calculation.base_price_cents, calculation.currency)}</strong></div>
          {#if calculation.discount_cents > 0}
            <div><span>Annual discount</span><strong>−{formatMoney(calculation.discount_cents, calculation.currency)}</strong></div>
          {/if}
          <div><span>Tax ({(calculation.tax_rate_bps / 100).toFixed(0)}%)</span><strong>{formatMoney(calculation.tax_cents, calculation.currency)}</strong></div>
          <div class="total"><span>Amount due</span><strong>{formatMoney(calculation.amount_due_cents, calculation.currency)}</strong></div>
        </div>
      {/if}

      <p class="field-label">Payment method</p>
      <div class="method-list">
        {#each paymentMethods as method (method.id)}
          <label class="method-option" class:selected={selectedMethod === method.id}>
            <input type="radio" name="pay-method" value={method.id} bind:group={selectedMethod} />
            <span class="method-icon" aria-hidden="true">{method.id === 'qr_promptpay' ? 'QR' : '💳'}</span>
            <span>
              <strong>{method.label}</strong>
              <small>{method.id === 'qr_promptpay' ? 'Scan with mobile banking' : 'Visa / Mastercard via ChillPay'}</small>
            </span>
          </label>
        {/each}
      </div>

      <div class="modal-actions">
        <button class="btn ghost" type="button" disabled={buying} onclick={closeCheckout}>Cancel</button>
        <button class="btn" type="button" disabled={buying || !calculation?.checkout_eligible} onclick={confirmBuy}>
          {buying ? 'Redirecting…' : 'Continue to payment'}
        </button>
      </div>
    </div>
  </div>
{/if}

{#if quotePkg}
  <div class="modal-backdrop" role="presentation" onclick={closeQuote} onkeydown={(event) => event.key === 'Escape' && closeQuote()}>
    <div class="card modal quote-modal" role="dialog" aria-modal="true" aria-labelledby="quote-title" tabindex="-1" onclick={(event) => event.stopPropagation()} onkeydown={(event) => event.stopPropagation()}>
    <form onsubmit={(event) => { event.preventDefault(); void submitQuote(); }}>
      <div class="modal-heading">
        <div>
          <span class="eyebrow">DEDICATED VM</span>
          <h2 id="quote-title">Request {quotePkg.name}</h2>
          <p>No payment is taken. We verify capacity and send commercial terms.</p>
        </div>
        <button class="close-button" type="button" aria-label="Close" disabled={quoteSubmitting} onclick={closeQuote}>×</button>
      </div>

      <div class="form-grid">
        <label class="field span-2">
          <span>Company legal name *</span>
          <input required maxlength="200" autocomplete="organization" bind:value={quoteForm.company_legal_name} />
        </label>
        <label class="field">
          <span>Contact name *</span>
          <input required maxlength="120" autocomplete="name" bind:value={quoteForm.contact_name} />
        </label>
        <label class="field">
          <span>Contact email *</span>
          <input required type="email" autocomplete="email" bind:value={quoteForm.contact_email} />
        </label>
        <label class="field">
          <span>Contact phone *</span>
          <input required maxlength="40" autocomplete="tel" bind:value={quoteForm.contact_phone} />
        </label>
        <label class="field">
          <span>Tax / registration ID</span>
          <input maxlength="80" bind:value={quoteForm.tax_registration_id} />
        </label>
        <label class="field">
          <span>Company size</span>
          <select bind:value={quoteForm.company_size}>
            <option value="">Select</option>
            <option value="1-10">1–10</option>
            <option value="11-50">11–50</option>
            <option value="51-200">51–200</option>
            <option value="201-500">201–500</option>
            <option value="501+">501+</option>
          </select>
        </label>
        <label class="field">
          <span>Expected concurrency *</span>
          <input required type="number" min="1" max="1000000" bind:value={quoteForm.expected_concurrency} />
        </label>
        <label class="field">
          <span>Preferred region</span>
          <select bind:value={quoteForm.preferred_region}>
            <option value="">No preference</option>
            <option value="th-bangkok">Bangkok</option>
            <option value="sg-singapore">Singapore</option>
            <option value="jp-tokyo">Tokyo</option>
            <option value="eu-frankfurt">Frankfurt</option>
            <option value="other">Other</option>
          </select>
        </label>
        <label class="field span-2">
          <span>Notes</span>
          <textarea rows="4" maxlength="2000" placeholder="Compliance, integration, rollout, or capacity details" bind:value={quoteForm.notes}></textarea>
        </label>
      </div>

      <div class="quote-invariant">
        <strong>Quotation only</strong>
        <span>This request creates no payment order, subscription, entitlement, receipt, or tax invoice.</span>
      </div>

      <div class="modal-actions">
        <button class="btn ghost" type="button" disabled={quoteSubmitting} onclick={closeQuote}>Cancel</button>
        <button class="btn quote-button" type="submit" disabled={quoteSubmitting}>
          {quoteSubmitting ? 'Submitting…' : 'Submit quotation request'}
        </button>
      </div>
    </form>
    </div>
  </div>
{/if}

<style>
  .page-wrap { max-width: 1180px; margin: 0 auto; padding: 32px 20px 56px; }
  .page-header, .section-heading, .quote-summary, .modal-heading { display: flex; justify-content: space-between; align-items: flex-start; gap: 16px; }
  .page-header { margin-bottom: 24px; }
  .page-header h1 { margin: 0 0 8px; font-size: 26px; }
  .page-header p, .current-plan p, .quote-summary p, .modal-heading p { margin: 0; color: var(--muted); font-size: 13px; line-height: 1.55; }
  .header-actions, .modal-actions { display: flex; gap: 10px; }
  .header-actions a { text-decoration: none; }
  .muted { color: var(--muted); }
  .eyebrow { display: block; margin-bottom: 7px; color: var(--muted); font-size: 10px; font-weight: 700; letter-spacing: .12em; }
  .current-plan { margin-bottom: 24px; }
  .current-plan h2, .quote-summary h2 { margin: 0 0 5px; font-size: 21px; }
  .next-bill { min-width: 150px; display: grid; gap: 3px; text-align: right; }
  .next-bill span, .next-bill small { color: var(--muted); font-size: 11px; }
  .next-bill strong { font-size: 20px; color: var(--cyan); }
  .status-pill { width: fit-content; border: 1px solid rgb(61 214 140 / 32%); border-radius: 999px; padding: 6px 10px; color: var(--success); background: rgb(61 214 140 / 8%); font-size: 11px; text-transform: capitalize; white-space: nowrap; }
  .status-pill.quote { border-color: rgb(22 199 255 / 32%); color: var(--cyan); background: rgb(22 199 255 / 8%); }
  .period-row { display: flex; gap: 32px; padding: 15px 0 2px; color: var(--ink); font-size: 12px; }
  .period-row span { display: grid; gap: 4px; }
  .period-row small { color: var(--muted); }
  .quota-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 12px; margin-top: 18px; }
  .quota-card { min-width: 0; border: 1px solid var(--line); border-radius: 12px; padding: 13px; background: rgb(5 12 25 / 58%); }
  .quota-card > div { display: flex; justify-content: space-between; gap: 8px; color: var(--muted); font-size: 11px; }
  .quota-card em { color: var(--success); font-size: 9px; font-style: normal; text-transform: uppercase; }
  .quota-card em.unavailable { color: var(--warn); }
  .quota-card strong { display: block; margin-top: 8px; font-size: 16px; }
  .quota-card strong small { color: var(--muted); font-size: 11px; font-weight: 500; }
  .quota-card p { min-height: 17px; margin-top: 4px; font-size: 10px; }
  .quota-progress { display: block; height: 4px; margin-top: 9px; overflow: hidden; border-radius: 999px; background: rgb(90 110 150 / 20%); }
  .quota-progress i { display: block; height: 100%; background: linear-gradient(90deg, var(--cyan), var(--violet)); }
  .quote-summary { align-items: center; margin-bottom: 24px; }
  .available-title, .quote-history h2 { margin: 0 0 16px; font-size: 17px; }
  .pkg-grid { display: grid; grid-template-columns: repeat(3, minmax(0, 1fr)); gap: 16px; }
  .pkg-card { display: flex; min-height: 300px; flex-direction: column; }
  .pkg-card.dedicated { border-color: rgb(141 57 255 / 30%); }
  .package-mode { width: fit-content; margin-bottom: 16px; border-radius: 999px; padding: 4px 8px; color: var(--cyan); background: rgb(22 199 255 / 9%); font-size: 9px; font-weight: 700; letter-spacing: .08em; text-transform: uppercase; }
  .dedicated .package-mode { color: #b885ff; background: rgb(141 57 255 / 10%); }
  .pkg-card h3 { margin: 0 0 7px; font-size: 19px; }
  .package-price { margin: 0 0 13px; color: var(--cyan); font-size: 16px; }
  .package-description { flex: 1; margin: 0 0 20px; color: var(--muted); font-size: 12px; line-height: 1.55; }
  .quote-button { background: linear-gradient(100deg, #5967ff, var(--violet)); }
  .quote-history { margin-top: 32px; }
  .history-list { display: grid; gap: 9px; }
  .history-list article { display: flex; justify-content: space-between; align-items: center; gap: 14px; padding: 14px 16px; }
  .history-list article > div { display: grid; gap: 4px; }
  .history-list article > div:last-child { justify-items: end; }
  .history-list span, .history-list small { color: var(--muted); font-size: 11px; }
  .modal-backdrop { position: fixed; inset: 0; z-index: 100; display: flex; align-items: center; justify-content: center; overflow: auto; padding: 20px; background: rgb(0 0 0 / 68%); }
  .modal { width: min(480px, 100%); max-height: calc(100vh - 40px); overflow: auto; outline: none; }
  .quote-modal { width: min(760px, 100%); }
  .modal-heading { margin-bottom: 18px; }
  .modal-heading h2 { margin: 0 0 4px; font-size: 20px; }
  .close-button { border: 0; color: var(--muted); background: transparent; font-size: 25px; line-height: 1; }
  .field { display: grid; gap: 6px; margin: 0; }
  .field > span, .field-label { color: var(--muted); font-size: 11px; font-weight: 600; }
  .field input, .field select, .field textarea { width: 100%; border: 1px solid var(--line); border-radius: 9px; padding: 10px 11px; }
  .form-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; }
  .span-2 { grid-column: span 2; }
  .calculation-card { display: grid; gap: 8px; margin: 14px 0 18px; border: 1px solid var(--line); border-radius: 11px; padding: 13px; background: rgb(4 10 21 / 68%); }
  .calculation-card > div { display: flex; justify-content: space-between; gap: 12px; font-size: 12px; }
  .calculation-card span { color: var(--muted); }
  .calculation-card .total { margin-top: 4px; border-top: 1px solid var(--line); padding-top: 10px; font-size: 15px; }
  .calculation-card .total strong { color: var(--cyan); }
  .method-list { display: grid; gap: 9px; margin-top: 8px; }
  .method-option { display: flex; align-items: flex-start; gap: 11px; border: 1px solid var(--line); border-radius: 10px; padding: 11px; cursor: pointer; }
  .method-option.selected { border-color: var(--cyan); box-shadow: 0 0 0 1px var(--cyan); }
  .method-option input { margin-top: 5px; }
  .method-option > span:last-child { display: grid; gap: 3px; }
  .method-option strong { font-size: 13px; }
  .method-option small { color: var(--muted); font-size: 10px; }
  .method-icon { width: 34px; height: 34px; flex-shrink: 0; display: grid; place-items: center; border-radius: 8px; background: rgb(22 199 255 / 10%); font-size: 12px; }
  .modal-actions { justify-content: flex-end; margin-top: 20px; }
  .quote-invariant { display: grid; gap: 4px; margin-top: 16px; border: 1px solid rgb(22 199 255 / 20%); border-radius: 10px; padding: 12px; background: rgb(22 199 255 / 6%); }
  .quote-invariant strong { color: var(--cyan); font-size: 12px; }
  .quote-invariant span { color: var(--muted); font-size: 10px; line-height: 1.45; }

  @media (max-width: 900px) {
    .pkg-grid, .quota-grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  }

  @media (max-width: 620px) {
    .page-wrap { padding: 22px 14px 88px; }
    .page-header, .section-heading, .quote-summary { display: grid; }
    .header-actions { width: 100%; }
    .header-actions a { flex: 1; text-align: center; }
    .next-bill { text-align: left; }
    .period-row { display: grid; gap: 12px; }
    .pkg-grid, .quota-grid, .form-grid { grid-template-columns: 1fr; }
    .span-2 { grid-column: span 1; }
    .pkg-card { min-height: auto; }
    .modal-backdrop { align-items: flex-start; padding: 10px; }
    .modal { max-height: calc(100vh - 20px); }
    .modal-actions { display: grid; grid-template-columns: 1fr; }
    .modal-actions button { width: 100%; }
  }
</style>
