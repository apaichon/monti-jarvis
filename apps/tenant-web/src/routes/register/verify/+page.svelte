<script lang="ts">
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { verifyEmail } from '$lib/api/register';
  import { saveSession } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';

  let status = $state<'loading' | 'ok' | 'error'>('loading');
  let message = $state('Verifying your email…');

  function failVerification(text: string) {
    status = 'error';
    message = text;
    feedback.error(text);
  }

  onMount(async () => {
    const token = $page.url.searchParams.get('token');
    if (!token) {
      failVerification('Missing verification token');
      return;
    }
    try {
      const res = await verifyEmail(token);
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
        status = 'ok';
        message = res.message ?? 'Email verified. Redirecting…';
        setTimeout(() => goto(`${base}/register/success`), 800);
        return;
      }
      failVerification('Verification did not return a session');
    } catch (err) {
      failVerification(err instanceof Error ? err.message : 'Verification failed');
    }
  });
</script>

<div class="login-wrap">
  <div class="card login-card" style="text-align:center">
    <h1 style="margin:0 0 8px;font-size:22px">Email verification</h1>
    <p style="color:var(--muted);font-size:14px;margin:0">{message}</p>
    {#if status === 'error'}
      <p style="margin-top:16px"><a class="link" href="{base}/register">Register again</a></p>
    {/if}
  </div>
</div>