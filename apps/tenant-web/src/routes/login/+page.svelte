<script lang="ts">
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { login } from '$lib/api/auth';
  import { ApiError } from '$lib/api/http';
  import { onMount } from 'svelte';
  import {
    getStoredUser,
    hasRegistrationSession,
    safeTenantNextPath,
    saveSession,
    type TokenPair
  } from '$lib/auth/session';
  import { fetchOAuthProviders, oauthStartURL } from '$lib/api/register';
  import OAuthButton from '$lib/components/OAuthButton.svelte';
  import { feedback } from '$lib/feedback.svelte';

  let email = $state('');
  let password = $state('');
  let loading = $state(false);
  let user = $state(getStoredUser());
  let providers = $state<string[]>([]);
  let sessionExpired = $state(false);

  function safeNextPath(): string {
    return safeTenantNextPath($page.url.searchParams.get('next'), base, `${base}/backoffice`);
  }

  function consumeOAuthCallback(): boolean {
    const params = $page.url.searchParams;
    const access = params.get('access_token');
    const refresh = params.get('refresh_token');
    if (!access || !refresh) {
      const err = params.get('error');
      if (err) feedback.error(err);
      return false;
    }
    const tenantId = params.get('tenant_id') ?? undefined;
    const role = params.get('role') || 'tenant_admin';
    if (role !== 'tenant_admin') {
      feedback.error('This portal is for tenant administrators');
      return true;
    }
    const pair: TokenPair = {
      access_token: access,
      refresh_token: refresh,
      expires_in: Number(params.get('expires_in') || 0),
      token_type: 'Bearer',
      user: {
        id: params.get('user_id') || '',
        email: params.get('email') || '',
        display_name: params.get('display_name') || '',
        role,
        tenant_id: tenantId
      }
    };
    saveSession(pair, tenantId);
    user = pair.user;
    const dest = safeNextPath();
    // Drop tokens from the URL so they are not left in history.
    history.replaceState({}, '', `${base}/login`);
    void goto(dest, { invalidateAll: true });
    return true;
  }

  onMount(async () => {
    if (consumeOAuthCallback()) return;
    if ($page.url.searchParams.get('reason') === 'session_expired') {
      sessionExpired = true;
      feedback.info('Your session expired. Please sign in again.', 'Session expired');
    }
    // Already signed in + return from payment / deep link → continue to next.
    if (hasRegistrationSession()) {
      const next = $page.url.searchParams.get('next');
      if (next || !sessionExpired) {
        if (next) {
          goto(safeNextPath());
          return;
        }
      }
    }
    try {
      const res = await fetchOAuthProviders();
      providers = res.providers;
    } catch {
      providers = [];
    }
  });

  function startOAuth(provider: string) {
    window.location.href = oauthStartURL(provider, {});
  }

  async function submit(e: Event) {
    e.preventDefault();
    loading = true;
    try {
      const pair = await login(email.trim(), password);
      if (pair.user.role !== 'tenant_admin') {
        feedback.error('This portal is for tenant administrators');
        return;
      }
      // Persist session before navigation so layout shell paints on first tick.
      saveSession(pair, pair.user.tenant_id);
      user = pair.user;
      sessionExpired = false;
      await goto(safeNextPath(), { invalidateAll: true });
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Sign in failed');
    } finally {
      loading = false;
    }
  }
</script>

<div class="login-wrap">
  <div class="card login-card">
    <div class="brand" style="margin-bottom:16px">
      <img src="{base}/images/monti-logo.png" alt="Monti" />
      <div>
        <strong>MONTI</strong>
        <div style="color:var(--muted);font-size:13px">Tenant portal</div>
      </div>
    </div>

    {#if sessionExpired}
      <div class="session-banner" role="status">
        Your session expired. Sign in again to continue.
      </div>
    {/if}

    {#if hasRegistrationSession() && user}
      <p style="margin:0 0 12px">You're already signed in as <strong>{user.display_name}</strong>.</p>
      <a class="btn" href={safeNextPath()} style="display:inline-block;text-decoration:none;margin-bottom:12px">Continue</a>
    {/if}

    {#if providers.length > 0}
      <div class="oauth-stack">
        {#if providers.includes('google')}
          <OAuthButton provider="google" label="Sign in with Google" disabled={loading} onclick={() => startOAuth('google')} />
        {/if}
        {#if providers.includes('github')}
          <OAuthButton provider="github" label="Sign in with GitHub" disabled={loading} onclick={() => startOAuth('github')} />
        {/if}
      </div>
      <p class="divider"><span>or sign in with email</span></p>
    {/if}

    <form onsubmit={submit}>
      <div class="field">
        <label for="email">Email</label>
        <input id="email" type="email" bind:value={email} disabled={loading} autocomplete="username" />
      </div>
      <div class="field">
        <label for="password">Password</label>
        <input id="password" type="password" bind:value={password} disabled={loading} autocomplete="current-password" />
      </div>
      <button class="btn" type="submit" disabled={loading} style="width:100%;margin-top:8px">
        {loading ? 'Signing in…' : 'Sign in'}
      </button>
    </form>

    <p style="color:var(--muted);font-size:12px;margin-top:16px;text-align:center">
      <a class="link" href="{base}/register">Create an account</a>
    </p>
  </div>
</div>

<style>
  .oauth-stack {
    display: grid;
    gap: 10px;
    margin-bottom: 16px;
  }

  .divider {
    display: flex;
    align-items: center;
    gap: 12px;
    margin: 0 0 16px;
    color: var(--muted);
    font-size: 12px;
  }

  .divider::before,
  .divider::after {
    content: '';
    flex: 1;
    height: 1px;
    background: rgb(70 132 190 / 22%);
  }

  .divider span {
    white-space: nowrap;
  }

  .session-banner {
    margin: 0 0 14px;
    padding: 10px 12px;
    border-radius: 10px;
    border: 1px solid rgb(240 184 63 / 35%);
    background: rgb(240 184 63 / 12%);
    color: #f0d9a0;
    font-size: 13px;
    line-height: 1.4;
  }
</style>
