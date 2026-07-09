<script lang="ts">
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { isPlatformAdmin, saveSession } from '$lib/auth/session';
  import { login } from '$lib/api/auth';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  let email = $state('platform@monti.local');
  let password = $state('');
  let loading = $state(false);

  onMount(() => {
    if (isPlatformAdmin()) {
      goto(`${base}/packages`);
    }
  });

  async function submit(e: Event) {
    e.preventDefault();
    if (!email.includes('@') || !password) {
      feedback.error('Email and password are required');
      return;
    }
    loading = true;
    try {
      const pair = await login(email.trim(), password);
      if (pair.user.role !== 'platform_admin') {
        feedback.error('This portal is for platform administrators only');
        return;
      }
      saveSession(pair);
      const next = $page.url.searchParams.get('next');
      goto(next && next.startsWith(base) ? next : `${base}/packages`);
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
        <div style="color:var(--muted);font-size:13px">Platform Admin</div>
      </div>
    </div>
    <p style="color:var(--muted);font-size:14px;margin:0 0 20px">
      Sign in to manage packages and tenant entitlements.
    </p>
    <form onsubmit={submit}>
      <div class="field">
        <label for="email">Email</label>
        <input id="email" type="email" bind:value={email} disabled={loading} autocomplete="username" />
      </div>
      <div class="field">
        <label for="password">Password</label>
        <input
          id="password"
          type="password"
          bind:value={password}
          disabled={loading}
          autocomplete="current-password"
        />
      </div>
      <button class="btn" type="submit" disabled={loading} style="width:100%;margin-top:8px">
        {loading ? 'Signing in…' : 'Sign in'}
      </button>
    </form>
    <p style="color:var(--muted);font-size:12px;margin-top:16px">
      Tenant admins: use API or tenant portal (Sprint 15+).
    </p>
  </div>
</div>