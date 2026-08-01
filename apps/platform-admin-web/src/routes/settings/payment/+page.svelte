<script lang="ts">
  import { onMount } from 'svelte';
  import {
    getPaymentGateway,
    reconcilePaymentGateway,
    testPaymentGateway,
    updatePaymentGateway,
    type PaymentGatewayConfig
  } from '$lib/api/payment';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  let config = $state<PaymentGatewayConfig | null>(null);
  let loading = $state(true);
  let saving = $state(false);
  let testing = $state(false);
  let reconciling = $state(false);
  let reconcileMessage = $state('');

  let provider = $state('mock');
  let mode = $state('test');
  let merchantCode = $state('');
  let apiKey = $state('');
  let md5Key = $state('');
  let baseURL = $state('https://sandbox-appsrv2.chillpay.co/api/v2/Payment');
  let routeNo = $state(1);
  let currency = $state('764');
  let returnURL = $state('');
  let stripePublishableKey = $state('');
  let stripeSecretKey = $state('');
  let stripeWebhookSecret = $state('');
  let stripeAPIBaseURL = $state('https://api.stripe.com');
  let stripeSuccessURL = $state('');
  let stripeCancelURL = $state('');

  onMount(load);

  async function load() {
    loading = true;
    try {
      config = await getPaymentGateway();
      provider = config.provider || 'mock';
      mode = config.mode || 'test';
      merchantCode = config.merchant_code || '';
      baseURL = config.base_url || baseURL;
      routeNo = config.route_no || 1;
      currency = config.currency || '764';
      returnURL = config.return_url || '';
      stripePublishableKey = config.stripe?.publishable_key || '';
      stripeAPIBaseURL = config.stripe?.api_base_url || 'https://api.stripe.com';
      stripeSuccessURL = config.stripe?.success_url || '';
      stripeCancelURL = config.stripe?.cancel_url || '';
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load payment settings');
    } finally {
      loading = false;
    }
  }

  async function save() {
    saving = true;
    try {
      const body = {
        provider,
        mode,
        merchant_code: merchantCode,
        base_url: baseURL,
        route_no: routeNo,
        currency,
        return_url: returnURL,
        stripe: {
          publishable_key: stripePublishableKey,
          api_base_url: stripeAPIBaseURL,
          success_url: stripeSuccessURL,
          cancel_url: stripeCancelURL
        }
      } as Parameters<typeof updatePaymentGateway>[0];
      if (apiKey.trim()) body.api_key = apiKey.trim();
      if (md5Key.trim()) body.md5_key = md5Key.trim();
      if (stripeSecretKey.trim()) body.stripe!.secret_key = stripeSecretKey.trim();
      if (stripeWebhookSecret.trim()) body.stripe!.webhook_secret = stripeWebhookSecret.trim();
      config = await updatePaymentGateway(body);
      apiKey = '';
      md5Key = '';
      stripeSecretKey = '';
      stripeWebhookSecret = '';
      feedback.success('Payment gateway saved');
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Save failed');
    } finally {
      saving = false;
    }
  }

  async function testConnection() {
    testing = true;
    try {
      const res = await testPaymentGateway();
      if (res.ok) {
        feedback.success(res.message || 'Connection OK');
      } else {
        feedback.error(res.message || 'Connection failed');
      }
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Test failed');
    } finally {
      testing = false;
    }
  }

  async function reconcile() {
    reconciling = true;
    reconcileMessage = '';
    try {
      const res = await reconcilePaymentGateway({ provider, limit: 50, dry_run: true });
      reconcileMessage = `${res.count} order${res.count === 1 ? '' : 's'} checked`;
      feedback.success('Reconciliation report ready');
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Reconcile failed');
    } finally {
      reconciling = false;
    }
  }
</script>

<h1 style="margin:0 0 4px;font-size:24px">Payment gateway</h1>
<p style="color:var(--muted);font-size:14px;margin:0 0 20px">
  Configure the active tenant checkout provider.
</p>

{#if loading}
  <p style="color:var(--muted)">Loading…</p>
{:else}
  <div class="card" style="max-width:640px">
    <div class="field">
      <label for="provider">Provider</label>
      <select id="provider" bind:value={provider}>
        <option value="mock">mock (local dev)</option>
        <option value="chillpay">chillpay</option>
        <option value="stripe">stripe</option>
      </select>
    </div>
    <div class="field">
      <label for="mode">Mode</label>
      <select id="mode" bind:value={mode}>
        <option value="test">test</option>
        <option value="live">live</option>
      </select>
    </div>
    {#if provider === 'chillpay'}
      <div class="field">
        <label for="merchant">Merchant code</label>
        <input id="merchant" bind:value={merchantCode} autocomplete="off" />
      </div>
      <div class="field">
        <label for="apikey">API key</label>
        <input id="apikey" type="password" bind:value={apiKey} placeholder={config?.api_key_masked || 'unchanged if empty'} />
      </div>
      <div class="field">
        <label for="md5">MD5 secret key</label>
        <input
          id="md5"
          type="password"
          bind:value={md5Key}
          placeholder={config?.md5_key_set ? '•••••••• (unchanged if empty)' : 'required for chillpay'}
        />
      </div>
      <div class="field">
        <label for="base">Base URL</label>
        <input id="base" bind:value={baseURL} />
      </div>
      <div class="field" style="display:grid;grid-template-columns:1fr 1fr;gap:12px">
        <div>
          <label for="route">Route no</label>
          <input id="route" type="number" min="1" bind:value={routeNo} />
        </div>
        <div>
          <label for="currency">Currency</label>
          <input id="currency" bind:value={currency} />
        </div>
      </div>
      <div class="field">
        <label for="return">Return URL</label>
        <input id="return" bind:value={returnURL} />
      </div>
    {:else if provider === 'stripe'}
      <div class="field">
        <label for="stripe-publishable">Publishable key</label>
        <input id="stripe-publishable" bind:value={stripePublishableKey} autocomplete="off" placeholder="pk_test_..." />
      </div>
      <div class="field">
        <label for="stripe-secret">Secret key</label>
        <input
          id="stripe-secret"
          type="password"
          bind:value={stripeSecretKey}
          placeholder={config?.stripe?.secret_key_set ? '•••••••• (unchanged if empty)' : 'required for stripe'}
        />
      </div>
      <div class="field">
        <label for="stripe-webhook">Webhook secret</label>
        <input
          id="stripe-webhook"
          type="password"
          bind:value={stripeWebhookSecret}
          placeholder={config?.stripe?.webhook_secret_set ? '•••••••• (unchanged if empty)' : 'required for webhook'}
        />
      </div>
      <div class="field">
        <label for="stripe-base">API base URL</label>
        <input id="stripe-base" bind:value={stripeAPIBaseURL} placeholder="https://api.stripe.com" />
      </div>
      <div class="field">
        <label for="stripe-success">Success URL</label>
        <input id="stripe-success" bind:value={stripeSuccessURL} placeholder="/tenant/billing/return" />
      </div>
      <div class="field">
        <label for="stripe-cancel">Cancel URL</label>
        <input id="stripe-cancel" bind:value={stripeCancelURL} placeholder="/tenant/billing" />
      </div>
      <div class="field">
        <label for="stripe-callback">Stripe webhook URL</label>
        <input id="stripe-callback" readonly value={config?.stripe?.callback_url ?? ''} style="opacity:0.85" />
      </div>
    {/if}
    <div class="field">
      <label for="callback">Callback URL (read-only)</label>
      <input id="callback" readonly value={config?.callback_url ?? ''} style="opacity:0.85" />
    </div>
    {#if config?.last_callback_at}
      <p style="color:var(--muted);font-size:13px;margin:0 0 12px">
        Last callback: {config.last_callback_at}
      </p>
    {/if}
    {#if config?.last_test_status}
      <p style="color:var(--muted);font-size:13px;margin:0 0 12px">
        Last test: {config.last_test_status}{config.last_tested_at ? ` · ${config.last_tested_at}` : ''}
      </p>
    {/if}
    {#if config?.last_webhook_status}
      <p style="color:var(--muted);font-size:13px;margin:0 0 12px">
        Last webhook: {config.last_webhook_status}{config.last_webhook_at ? ` · ${config.last_webhook_at}` : ''}
      </p>
    {/if}
    {#if reconcileMessage}
      <p style="color:var(--muted);font-size:13px;margin:0 0 12px">{reconcileMessage}</p>
    {/if}
    <div style="display:flex;gap:12px;margin-top:8px">
      <button class="btn ghost" type="button" disabled={testing} onclick={testConnection}>
        {testing ? 'Testing…' : 'Test connection'}
      </button>
      {#if provider === 'stripe'}
        <button class="btn ghost" type="button" disabled={reconciling} onclick={reconcile}>
          {reconciling ? 'Reconciling…' : 'Reconcile'}
        </button>
      {/if}
      <button class="btn primary" type="button" disabled={saving} onclick={save}>
        {saving ? 'Saving…' : 'Save'}
      </button>
    </div>
  </div>
{/if}
