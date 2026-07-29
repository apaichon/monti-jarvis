<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { page } from '$app/stores';
  import FeedbackDialog from '$lib/components/FeedbackDialog.svelte';
  import { currentPlan } from '$lib/currentPlan.svelte';
  import {
    clearSession,
    bootstrapSession,
    hasRegistrationSession,
    subscribeSession
  } from '$lib/auth/session';

  let { children } = $props();

  // Reactive tick so first login re-renders shell without hard refresh (SPRINT-042).
  let sessionTick = $state(0);
  let sessionReady = $state(false);

  onMount(() => {
    const unsubscribe = subscribeSession(() => {
      sessionTick += 1;
    });
    void bootstrapSession()
      .then(() => {
        if (hasRegistrationSession()) {
          void currentPlan.load().catch(() => {});
        }
      })
      .finally(() => {
        sessionReady = true;
      });
    return unsubscribe;
  });

  const showShell = $derived(
    sessionReady &&
      hasRegistrationSession() &&
      !$page.url.pathname.endsWith('/login') &&
      !$page.url.pathname.includes('/register')
  );

  function logout() {
    void fetch('/api/auth/logout', { method: 'POST', credentials: 'include' }).catch(() => {});
    clearSession();
    currentPlan.clear();
    window.location.href = `${base}/login`;
  }

  function active(pathPart: string) {
    return $page.url.pathname.includes(pathPart);
  }

  function planUsagePercent(): number {
    const value = currentPlan.data?.compact_utilization;
    if (typeof value !== 'number' || !Number.isFinite(value)) return 0;
    return Math.max(0, Math.min(100, Math.round(value * 100)));
  }

  function planUsageLabel(): string {
    if (currentPlan.loading && !currentPlan.data) return 'Loading allowance…';
    if (currentPlan.data?.compact_utilization == null) return 'Usage unavailable';
    return `${planUsagePercent()}% highest quota usage`;
  }
</script>

{#if !sessionReady}
  <div style="min-height:100vh;display:grid;place-items:center;color:var(--muted)">Loading session…</div>
{:else if showShell}
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
          <a class="nav-link" href="{base}/avatars" class:active={active('/avatars')}
            ><span>☺</span>Avatars</a
          >
          <a class="nav-link" href="{base}/embed" class:active={active('/embed')}
            ><span>⌘</span>Embed</a
          >
          <a class="nav-link" href="{base}/theme" class:active={active('/theme')}
            ><span>◈</span>Theme</a
          >
          <a class="nav-link" href="{base}/ai" class:active={active('/ai')}
            ><span>✦</span>AI config</a
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
          <div class="nav-group-label">Growth</div>
          <a class="nav-link" href="{base}/referrals" class:active={active('/referrals')}><span>↗</span>Referrals</a>
        </div>

        <div class="nav-group">
          <div class="nav-group-label">Settings</div>
          <a class="nav-link" href="{base}/settings" class:active={active('/settings')}
            ><span>⚙</span>Settings</a
          >
        </div>
      </nav>

      <div class="tenant-sidebar-foot">
        <a class="plan-card plan-card-link" href="{base}/billing">
          <small>CURRENT PLAN</small>
          <strong>{currentPlan.data?.package?.name ?? (currentPlan.loading ? 'Loading…' : 'No active plan')}</strong>
          <span><i style={`width:${planUsagePercent()}%`}></i></span>
          <small>{planUsageLabel()}</small>
        </a>
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
