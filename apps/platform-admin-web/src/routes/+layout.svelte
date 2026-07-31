<script lang="ts">
  import '../app.css';
  import FeedbackDialog from '$lib/components/FeedbackDialog.svelte';
  import LanguageSelector from '$lib/components/LanguageSelector.svelte';
  import { initLangFromUrl, t } from '$lib/i18n';
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
  import { onMount } from 'svelte';

  let { children } = $props();

  const user = $derived(getStoredUser());
  const onLoginPage = $derived($page.url.pathname === `${base}/login` || $page.url.pathname === `${base}/login/`);
  let appVersion = $state('');

  onMount(() => {
    initLangFromUrl(new URLSearchParams(window.location.search));
    void fetch('/api/version')
      .then((r) => (r.ok ? r.json() : null))
      .then((data) => {
        if (data?.version) appVersion = String(data.version);
      })
      .catch(() => {});
  });

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
        <span><strong>MONTI</strong><small>{$t.brand_admin}</small></span>
      </a>
      <nav class="nav-links" aria-label={$t.nav_aria}>
        <a class="nav-link" class:active={$page.url.pathname === `${base}/`} href="{base}/"><span>⌂</span>{$t.nav_overview}</a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/packages`)}
          href="{base}/packages"
        >
          <span>▦</span>{$t.nav_packages}
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/tenants`) && !$page.url.pathname.includes('/avatars') && !$page.url.pathname.includes('/entitlement')}
          href="{base}/tenants"
        >
          <span>◇</span>{$t.nav_tenants}
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/avatars`)}
          href="{base}/avatars"
        >
          <span>◉</span>{$t.nav_avatars}
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/billing`) && !$page.url.pathname.startsWith(`${base}/billing/quotes`)}
          href="{base}/billing"
        >
          <span>▣</span>{$t.nav_billing}
        </a>
        <a
          class="nav-link sub-link"
          class:active={$page.url.pathname.startsWith(`${base}/billing/quotes`)}
          href="{base}/billing/quotes"
        >
          <span>▥</span>{$t.nav_quotes}
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/leads`)}
          href="{base}/leads"
        >
          <span>✉</span>{$t.nav_leads}
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/referrals`)}
          href="{base}/referrals"
        >
          <span>↗</span>{$t.nav_referrals}
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/audit-logs`)}
          href="{base}/audit-logs"
        >
          <span>▤</span>{$t.nav_audit}
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/monitoring`)}
          href="{base}/monitoring"
        >
          <span>◌</span>{$t.nav_monitoring}
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/call-center`)}
          href="{base}/call-center"
        >
          <span>◍</span>{$t.nav_call_center}
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/settings`)}
          href="{base}/settings/payment"
        >
          <span>⚙</span>{$t.nav_payment}
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/profile`)}
          href="{base}/profile"
        >
          <span>◎</span>{$t.nav_profile}
        </a>
      </nav>
      <div class="admin-sidebar-foot">
        <div class="system-card"><span><i></i>{$t.system_health}</span><strong>{$t.system_all_ok}</strong></div>
        {#if appVersion}
          <div style="padding:2px 0 8px;font-size:11px;color:var(--muted)">{$t.admin_version} · {appVersion}</div>
        {/if}
        <LanguageSelector compact />
        <button class="admin-account" type="button" onclick={handleLogout}><b>AD</b><span><strong>{$t.account_admin}</strong><small>{user?.email ?? $t.account_sign_out}</small></span><em>↗</em></button>
      </div>
    </aside>
    <section class="admin-workspace">
      <header class="topnav">
        <div class="admin-context"><span>{$t.topbar_platform}</span><b>/</b><strong>{$t.topbar_admin}</strong></div>
        <div class="nav-right"><button aria-label={$t.topbar_search}>⌕</button><button aria-label={$t.topbar_notifications}>♢</button><span class="role-badge">{$t.role_super}</span></div>
      </header>
      <main class="main">{@render children()}</main>
    </section>
  </div>
{/if}
