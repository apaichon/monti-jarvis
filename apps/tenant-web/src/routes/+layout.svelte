<script lang="ts">
  import '../app.css';
  import { base } from '$app/paths';
  import { page } from '$app/stores';
  import FeedbackDialog from '$lib/components/FeedbackDialog.svelte';
  import { clearSession, hasRegistrationSession } from '$lib/auth/session';

  let { children } = $props();

  const showShell = $derived(
    hasRegistrationSession() &&
      !$page.url.pathname.endsWith('/login') &&
      !$page.url.pathname.includes('/register')
  );

  function logout() {
    clearSession();
    window.location.href = `${base}/login`;
  }
</script>

{#if showShell}
  <div class="tenant-app-shell">
    <aside class="tenant-sidebar">
      <a class="brand tenant-brand" href="{base}/backoffice">
        <img src="{base}/images/monti-logo.png" alt="" />
        <span><strong>MONTI</strong><small>TENANT CONSOLE</small></span>
      </a>
      <div class="workspace-switcher"><span class="workspace-avatar">M</span><span><small>Workspace</small><strong>Monti AI</strong></span><b>⌄</b></div>
      <nav class="tenant-nav" aria-label="Tenant navigation">
        <a class="nav-link" href="{base}/backoffice" class:active={$page.url.pathname.includes('/backoffice')}><span>⌂</span>Overview</a>
        <a class="nav-link" href="{base}/dashboard" class:active={$page.url.pathname.includes('/dashboard')}><span>▦</span>Call center</a>
        <a
          class="nav-link"
          href="{base}/billing"
          class:active={$page.url.pathname.includes('/billing') &&
            !$page.url.pathname.includes('/documents') &&
            !$page.url.pathname.includes('/tax')}><span>▣</span>Billing</a
        >
        <a
          class="nav-link"
          href="{base}/billing/documents"
          class:active={$page.url.pathname.includes('/documents')}><span>▤</span>Documents</a
        >
        <a class="nav-link" href="{base}/billing/tax" class:active={$page.url.pathname.includes('/tax')}
          ><span>◇</span>Tax</a
        >
        <a
          class="nav-link"
          href="{base}/embed"
          class:active={$page.url.pathname.includes('/embed')}><span>⌘</span>Embed</a
        >
        <a
          class="nav-link"
          href="{base}/km"
          class:active={$page.url.pathname.includes('/km')}><span>◫</span>Knowledge</a
        >
        <a
          class="nav-link"
          href="{base}/settings"
          class:active={$page.url.pathname.includes('/settings')}><span>⚙</span>Settings</a
        >
        <a
          class="nav-link"
          href="{base}/tiers"
          class:active={$page.url.pathname.includes('/tiers')}><span>◆</span>Tiers</a
        >
        <a
          class="nav-link"
          href="{base}/customers"
          class:active={$page.url.pathname.includes('/customers')}><span>♙</span>Customers</a
        >
        <a
          class="nav-link"
          href="{base}/conversation-records"
          class:active={$page.url.pathname.includes('/conversation-records')}><span>▥</span>Records</a
        >
        <a
          class="nav-link"
          href="{base}/knowledge-gaps"
          class:active={$page.url.pathname.includes('/knowledge-gaps')}><span>△</span>Gaps</a
        >
        <a
          class="nav-link"
          href="{base}/tickets"
          class:active={$page.url.pathname.includes('/tickets')}><span>▱</span>Tickets</a
        >
        <a
          class="nav-link"
          href="{base}/satisfaction"
          class:active={$page.url.pathname.includes('/satisfaction')}><span>★</span>Satisfaction</a
        >
        <a
          class="nav-link"
          href="{base}/preview"
          class:active={$page.url.pathname.includes('/preview')}><span>◉</span>Preview <em>LIVE</em></a
        >
      </nav>
      <div class="tenant-sidebar-foot">
        <div class="plan-card"><small>CURRENT PLAN</small><strong>Enterprise</strong><span><i></i></span><small>68% monthly allowance</small></div>
        <button class="account-button" type="button" onclick={logout}><span class="workspace-avatar">AD</span><span><strong>Admin</strong><small>Sign out</small></span><b>↗</b></button>
      </div>
    </aside>
    <section class="tenant-workspace">
      <header class="tenant-topbar">
        <div><span class="status-dot"></span> All systems operational</div>
        <div class="topbar-actions"><button aria-label="Search">⌕</button><button aria-label="Notifications">♢</button><a href="{base}/login" aria-label="Profile">AD</a></div>
      </header>
      <main class="tenant-main">{@render children()}</main>
    </section>
  </div>
{:else}
  {@render children()}
{/if}
<FeedbackDialog />
