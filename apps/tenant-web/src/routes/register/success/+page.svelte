<script lang="ts">
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { getStoredTenantId, getStoredUser, hasRegistrationSession, saveSession } from '$lib/auth/session';

  let tenantId = $state('');
  let displayName = $state('');

  onMount(() => {
    const access = $page.url.searchParams.get('access_token');
    const refresh = $page.url.searchParams.get('refresh_token');
    const tid = $page.url.searchParams.get('tenant_id');
    if (access && refresh) {
      saveSession(
        {
          access_token: access,
          refresh_token: refresh,
          expires_in: 0,
          token_type: 'Bearer',
          user: {
            id: '',
            email: '',
            display_name: '',
            role: 'tenant_admin',
            tenant_id: tid ?? undefined
          }
        },
        tid ?? undefined
      );
    }
    if (!hasRegistrationSession()) {
      goto(`${base}/register`);
      return;
    }
    tenantId = getStoredTenantId() ?? tid ?? '';
    displayName = getStoredUser()?.display_name ?? '';
  });
</script>

<div class="login-wrap">
  <div class="card login-card" style="text-align:center">
    <div class="success-icon" style="margin-bottom:12px">✓</div>
    <h1 style="margin:0 0 8px;font-size:22px">Registration complete</h1>
    <p style="color:var(--muted);font-size:14px;margin:0 0 20px">
      You can sign in now. Platform KYC review is pending — submit your business details in the backoffice.
    </p>

    <div style="text-align:left;margin-bottom:20px">
      <p style="margin:0 0 6px;font-size:13px;color:var(--muted)">Workspace</p>
      <p style="margin:0;font-size:18px;font-weight:600">{tenantId || '—'}</p>
      {#if displayName}
        <p style="margin:8px 0 0;font-size:13px;color:var(--muted)">Signed in as {displayName}</p>
      {/if}
    </div>

    <a class="btn" href="{base}/backoffice" style="display:block;text-decoration:none;margin-bottom:10px">Open tenant backoffice</a>
    <a class="link" href="/">Try caller demo →</a>
  </div>
</div>