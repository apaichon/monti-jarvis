<script lang="ts">
  import '../app.css';
  import FeedbackDialog from '$lib/components/FeedbackDialog.svelte';
  import { base } from '$app/paths';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { browser } from '$app/environment';
  import {
    clearSession,
    getRefreshToken,
    getStoredUser,
    isPlatformAdmin,
    loginPath
  } from '$lib/auth/session';
  import { logout } from '$lib/api/auth';

  let { children } = $props();

  const user = $derived(getStoredUser());
  const onLoginPage = $derived($page.url.pathname === `${base}/login` || $page.url.pathname === `${base}/login/`);

  $effect(() => {
    if (!browser || onLoginPage) return;
    if (!isPlatformAdmin()) {
      goto(loginPath($page.url.pathname));
    }
  });

  async function handleLogout() {
    const refresh = getRefreshToken();
    try {
      if (refresh) await logout(refresh);
    } catch {
      // ignore
    }
    clearSession();
    goto(`${base}/login`);
  }
</script>

<FeedbackDialog />

{#if onLoginPage}
  {@render children()}
{:else}
  <div class="shell">
    <aside class="admin-sidebar">
      <a class="brand" href="{base}/packages">
        <img src="{base}/images/monti-logo.png" alt="Monti" />
        <span><strong>MONTI</strong><small>PLATFORM ADMIN</small></span>
      </a>
      <nav class="nav-links" aria-label="Platform navigation">
        <a class="nav-link" class:active={$page.url.pathname === `${base}/`} href="{base}/"><span>⌂</span>Overview</a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/packages`)}
          href="{base}/packages"
        >
          <span>▦</span>Packages
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/tenants`) && !$page.url.pathname.includes('/avatars') && !$page.url.pathname.includes('/entitlement')}
          href="{base}/tenants"
        >
          <span>◇</span>Tenants
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/avatars`)}
          href="{base}/avatars"
        >
          <span>◉</span>Avatars
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/billing`)}
          href="{base}/billing"
        >
          <span>▣</span>Billing
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/settings`)}
          href="{base}/settings/payment"
        >
          <span>⚙</span>Payment
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/profile`)}
          href="{base}/profile"
        >
          <span>◎</span>Profile
        </a>
      </nav>
      <div class="admin-sidebar-foot">
        <div class="system-card"><span><i></i>System health</span><strong>All systems operational</strong></div>
        <button class="admin-account" type="button" onclick={handleLogout}><b>AD</b><span><strong>Admin</strong><small>{user?.email ?? 'Sign out'}</small></span><em>↗</em></button>
      </div>
    </aside>
    <section class="admin-workspace">
      <header class="topnav">
        <div class="admin-context"><span>Monti Platform</span><b>/</b><strong>Administration</strong></div>
        <div class="nav-right"><button aria-label="Search">⌕</button><button aria-label="Notifications">♢</button><span class="role-badge">SUPER ADMIN</span></div>
      </header>
      <main class="main">{@render children()}</main>
    </section>
  </div>
{/if}
