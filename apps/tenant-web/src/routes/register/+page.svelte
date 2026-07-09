<script lang="ts">
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { registerTenant, fetchOAuthProviders, oauthStartURL } from '$lib/api/register';
  import { ApiError } from '$lib/api/http';
  import { saveSession } from '$lib/auth/session';
  import { suggestSlug } from '$lib/utils/slug';
  import OAuthButton from '$lib/components/OAuthButton.svelte';
  import { feedback } from '$lib/feedback.svelte';

  let companyName = $state('');
  let slug = $state('');
  let slugTouched = $state(false);
  let adminEmail = $state('');
  let adminDisplayName = $state('');
  let password = $state('');
  let confirmPassword = $state('');
  let loading = $state(false);
  let providers = $state<string[]>([]);
  let mode = $state<'choose' | 'email'>('choose');

  onMount(async () => {
    const err = $page.url.searchParams.get('error');
    if (err) feedback.error(err);
    try {
      const res = await fetchOAuthProviders();
      providers = res.providers;
    } catch {
      providers = [];
    }
  });

  function onCompanyInput() {
    if (!slugTouched) slug = suggestSlug(companyName);
  }

  function onSlugInput() {
    slugTouched = true;
    slug = slug.trim().toLowerCase();
  }

  function startOAuth(provider: string) {
    window.location.href = oauthStartURL(provider, {
      company_name: companyName.trim(),
      slug: slug.trim(),
      display_name: adminDisplayName.trim()
    });
  }

  async function submit(e: Event) {
    e.preventDefault();
    if (!companyName.trim() || !slug.trim() || !adminEmail.trim() || !adminDisplayName.trim()) {
      feedback.error('All fields are required');
      return;
    }
    if (password.length < 8) {
      feedback.error('Password must be at least 8 characters');
      return;
    }
    if (password !== confirmPassword) {
      feedback.error('Passwords do not match');
      return;
    }
    loading = true;
    try {
      const res = await registerTenant({
        company_name: companyName.trim(),
        slug: slug.trim(),
        admin_email: adminEmail.trim(),
        admin_password: password,
        admin_display_name: adminDisplayName.trim()
      });
      if (res.verification_required) {
        goto(`${base}/register/check-email?email=${encodeURIComponent(adminEmail.trim())}`);
        return;
      }
      if (res.access_token && res.refresh_token && res.user) {
        saveSession(
          {
            access_token: res.access_token,
            refresh_token: res.refresh_token,
            expires_in: res.expires_in ?? 0,
            token_type: res.token_type ?? 'Bearer',
            user: res.user
          },
          res.tenant_id,
          res.registration_id
        );
        goto(`${base}/register/success`);
      }
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Registration failed');
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
        <div style="color:var(--muted);font-size:13px">Start your AI call center workspace</div>
      </div>
    </div>

    {#if mode === 'choose'}
      <div class="field">
        <label for="company-quick">Company name</label>
        <input id="company-quick" bind:value={companyName} oninput={onCompanyInput} placeholder="Acme Corp" />
      </div>
      <div class="field">
        <label for="slug-quick">Workspace URL</label>
        <div class="slug-row">
          <span class="slug-prefix">monti.app/</span>
          <input id="slug-quick" bind:value={slug} oninput={onSlugInput} />
        </div>
      </div>
      <div class="field">
        <label for="display-quick">Your name</label>
        <input id="display-quick" bind:value={adminDisplayName} />
      </div>

      <div class="oauth-stack">
        {#if providers.includes('google')}
          <OAuthButton provider="google" label="Continue with Google" onclick={() => startOAuth('google')} />
        {/if}
        {#if providers.includes('github')}
          <OAuthButton provider="github" label="Continue with GitHub" onclick={() => startOAuth('github')} />
        {/if}
        <button class="btn email-btn" type="button" onclick={() => (mode = 'email')}>
          <span class="email-icon" aria-hidden="true">✉</span>
          Continue with email
        </button>
      </div>
    {:else}
      <form onsubmit={submit}>
        <div class="field">
          <label for="company">Company name</label>
          <input id="company" type="text" bind:value={companyName} oninput={onCompanyInput} disabled={loading} />
        </div>
        <div class="field">
          <label for="slug">Workspace URL</label>
          <div class="slug-row">
            <span class="slug-prefix">monti.app/</span>
            <input id="slug" type="text" bind:value={slug} oninput={onSlugInput} disabled={loading} />
          </div>
        </div>
        <div class="field">
          <label for="email">Admin email</label>
          <input id="email" type="email" bind:value={adminEmail} disabled={loading} />
        </div>
        <div class="field">
          <label for="display">Your name</label>
          <input id="display" type="text" bind:value={adminDisplayName} disabled={loading} />
        </div>
        <div class="field">
          <label for="password">Password</label>
          <input id="password" type="password" bind:value={password} disabled={loading} />
        </div>
        <div class="field">
          <label for="confirm">Confirm password</label>
          <input id="confirm" type="password" bind:value={confirmPassword} disabled={loading} />
        </div>
        <button class="btn" type="submit" disabled={loading} style="width:100%;margin-top:8px">
          {loading ? 'Creating account…' : 'Create account with email'}
        </button>
        <button class="btn ghost" type="button" style="width:100%;margin-top:8px" onclick={() => (mode = 'choose')}>
          Back to sign-up options
        </button>
      </form>
    {/if}

    <p style="color:var(--muted);font-size:12px;margin-top:16px;text-align:center">
      Already have an account? <a class="link" href="{base}/login">Sign in</a> ·
      <a class="link" href="/">Caller portal</a>
    </p>
  </div>
</div>

<style>
  .oauth-stack {
    display: grid;
    gap: 10px;
    margin-top: 8px;
  }

  .email-btn {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    width: 100%;
  }

  .email-icon {
    font-size: 16px;
    line-height: 1;
    opacity: 0.9;
  }
</style>