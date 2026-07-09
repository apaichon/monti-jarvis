<script lang="ts">
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { login } from '$lib/api/auth';
  import { ApiError } from '$lib/api/http';
  import { onMount } from 'svelte';
  import { getStoredUser, hasRegistrationSession, saveSession } from '$lib/auth/session';
  import { fetchOAuthProviders, oauthStartURL } from '$lib/api/register';
  import OAuthButton from '$lib/components/OAuthButton.svelte';
  import { feedback } from '$lib/feedback.svelte';

  let email = $state('');
  let password = $state('');
  let loading = $state(false);
  let user = $state(getStoredUser());
  let providers = $state<string[]>([]);

  onMount(async () => {
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
      saveSession(pair, pair.user.tenant_id);
      goto(`${base}/backoffice`);
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

    {#if hasRegistrationSession() && user}
      <p style="margin:0 0 12px">You're already signed in as <strong>{user.display_name}</strong>.</p>
      <a class="btn" href="{base}/backoffice" style="display:inline-block;text-decoration:none;margin-bottom:12px">Open backoffice</a>
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
</style>