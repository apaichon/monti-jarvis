<script lang="ts">
  import { onMount } from 'svelte';
  import {
    getPaymentGateway,
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

  let provider = $state('mock');
  let mode = $state('test');
  let merchantCode = $state('');
  let apiKey = $state('');
  let md5Key = $state('');
  let baseURL = $state('https://sandbox-appsrv2.chillpay.co/api/v2/Payment');
  let routeNo = $state(1);
  let currency = $state('764');
  let returnURL = $state('');

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
        return_url: returnURL
      } as Parameters<typeof updatePaymentGateway>[0];
      if (apiKey.trim()) body.api_key = apiKey.trim();
      if (md5Key.trim()) body.md5_key = md5Key.trim();
      config = await updatePaymentGateway(body);
      apiKey = '';
      md5Key = '';
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
</script>

<h1 style="margin:0 0 4px;font-size:24px">Payment gateway</h1>
<p style="color:var(--muted);font-size:14px;margin:0 0 20px">
  Configure ChillPay (or mock) for tenant checkout in Sprint 9.
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
      </select>
    </div>
    <div class="field">
      <label for="mode">Mode</label>
      <select id="mode" bind:value={mode}>
        <option value="test">test</option>
        <option value="live">live</option>
      </select>
    </div>
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
    <div class="field">
      <label for="callback">Callback URL (read-only)</label>
      <input id="callback" readonly value={config?.callback_url ?? ''} style="opacity:0.85" />
    </div>
    {#if config?.last_callback_at}
      <p style="color:var(--muted);font-size:13px;margin:0 0 12px">
        Last callback: {config.last_callback_at}
      </p>
    {/if}
    <div style="display:flex;gap:12px;margin-top:8px">
      <button class="btn ghost" type="button" disabled={testing} onclick={testConnection}>
        {testing ? 'Testing…' : 'Test connection'}
      </button>
      <button class="btn primary" type="button" disabled={saving} onclick={save}>
        {saving ? 'Saving…' : 'Save'}
      </button>
    </div>
  </div>
{/if}