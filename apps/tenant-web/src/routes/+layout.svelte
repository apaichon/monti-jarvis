<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { page } from '$app/stores';
  import FeedbackDialog from '$lib/components/FeedbackDialog.svelte';
  import {
    clearSession,
    hasRegistrationSession,
    subscribeSession
  } from '$lib/auth/session';

  let { children } = $props();

  // Reactive tick so first login re-renders shell without hard refresh (SPRINT-042).
  let sessionTick = $state(0);

  onMount(() => {
    return subscribeSession(() => {
      sessionTick += 1;
    });
  });

  const showShell = $derived(
    sessionTick >= 0 &&
      hasRegistrationSession() &&
      !$page.url.pathname.endsWith('/login') &&
      !$page.url.pathname.includes('/register')
  );

  function logout() {
    clearSession();
    window.location.href = `${base}/login`;
  }

  function active(pathPart: string) {
    return $page.url.pathname.includes(pathPart);
  }
</script>

{#if showShell}
  <div class="tenant-app-shell">
    <aside class="tenant-sidebar">
      <div class="tenant-sidebar-top">
        <a class="brand tenant-brand" href="{base}/backoffice">
          <img src="{base}/images/monti-logo.png" alt="" />
          <span><strong>MONTI</strong><small>TENANT CONSOLE</small></span>
        </a>
        <div class="workspace-switcher">
          <span class="workspace-avatar">M</span>
          <span><small>Workspace</small><strong>Monti AI</strong></span><b>⌄</b>
        </div>
      </div>

      <nav class="tenant-nav" aria-label="Tenant navigation">
        <div class="nav-group">
          <div class="nav-group-label">Operations</div>
          <a class="nav-link" href="{base}/backoffice" class:active={active('/backoffice')}
            ><span>⌂</span>Overview</a
          >
          <a class="nav-link" href="{base}/dashboard" class:active={active('/dashboard')}
            ><span>▦</span>Call center</a
          >
          <a class="nav-link" href="{base}/monitoring" class:active={active('/monitoring')}
            ><span>◌</span>Monitoring</a
          >
          <a class="nav-link" href="{base}/tickets" class:active={active('/tickets')}
            ><span>▱</span>Tickets</a
          >
          <a class="nav-link" href="{base}/satisfaction" class:active={active('/satisfaction')}
            ><span>★</span>Satisfaction</a
          >
          <a class="nav-link" href="{base}/preview" class:active={active('/preview')}
            ><span>◉</span>Preview <em>LIVE</em></a
          >
        </div>

        <div class="nav-group">
          <div class="nav-group-label">Knowledge</div>
          <a class="nav-link" href="{base}/km" class:active={active('/km') && !active('/knowledge-gaps')}
            ><span>◫</span>Knowledge</a
          >
          <a class="nav-link" href="{base}/knowledge-gaps" class:active={active('/knowledge-gaps')}
            ><span>△</span>Gaps</a
          >
          <a
            class="nav-link"
            href="{base}/conversation-records"
            class:active={active('/conversation-records')}><span>▥</span>Records</a
          >
        </div>

        <div class="nav-group">
          <div class="nav-group-label">Commerce</div>
          <a
            class="nav-link"
            href="{base}/billing"
            class:active={active('/billing') && !active('/documents') && !active('/tax')}
            ><span>▣</span>Billing</a
          >
          <a class="nav-link" href="{base}/billing/documents" class:active={active('/documents')}
            ><span>▤</span>Documents</a
          >
          <a class="nav-link" href="{base}/billing/tax" class:active={active('/tax')}
            ><span>◇</span>Tax</a
          >
        </div>

        <div class="nav-group">
          <div class="nav-group-label">Channels</div>
          <a class="nav-link" href="{base}/embed" class:active={active('/embed')}
            ><span>⌘</span>Embed</a
          >
          <a class="nav-link" href="{base}/theme" class:active={active('/theme')}
            ><span>◈</span>Theme</a
          >
        </div>

        <div class="nav-group">
          <div class="nav-group-label">Directory</div>
          <a class="nav-link" href="{base}/customers" class:active={active('/customers')}
            ><span>♙</span>Customers</a
          >
          <a class="nav-link" href="{base}/tiers" class:active={active('/tiers')}
            ><span>◆</span>Tiers</a
          >
        </div>

        <div class="nav-group">
          <div class="nav-group-label">Settings</div>
          <a class="nav-link" href="{base}/settings" class:active={active('/settings')}
            ><span>⚙</span>Settings</a
          >
        </div>
      </nav>

      <div class="tenant-sidebar-foot">
        <div class="plan-card">
          <small>CURRENT PLAN</small><strong>Enterprise</strong><span><i></i></span><small
            >68% monthly allowance</small
          >
        </div>
        <button class="account-button" type="button" onclick={logout}
          ><span class="workspace-avatar">AD</span><span
            ><strong>Admin</strong><small>Sign out</small></span
          ><b>↗</b></button
        >
      </div>
    </aside>
    <section class="tenant-workspace">
      <header class="tenant-topbar">
        <div><span class="status-dot"></span> All systems operational</div>
        <div class="topbar-actions">
          <button aria-label="Search">⌕</button><button aria-label="Notifications">♢</button><a
            href="{base}/login"
            aria-label="Profile">AD</a
          >
        </div>
      </header>
      <main class="tenant-main">{@render children()}</main>
    </section>
  </div>
{:else}
  {@render children()}
{/if}
<FeedbackDialog />
