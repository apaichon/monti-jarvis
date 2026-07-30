<script lang="ts">
  import '../app.css';
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { page } from '$app/stores';
  import FeedbackDialog from '$lib/components/FeedbackDialog.svelte';
  import LanguageSelector from '$lib/components/LanguageSelector.svelte';
  import { initLangFromUrl, t } from '$lib/i18n';
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
  let appVersion = $state('');

  onMount(() => {
    initLangFromUrl(new URLSearchParams(window.location.search));
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
    void fetch('/api/version')
      .then((r) => (r.ok ? r.json() : null))
      .then((data) => {
        if (data?.version) appVersion = String(data.version);
      })
      .catch(() => {});
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
    // Use get-style via store snapshot at call time is hard; keep English numbers, localized phrase from $t in template.
    if (currentPlan.loading && !currentPlan.data) return '';
    if (currentPlan.data?.compact_utilization == null) return '';
    return `${planUsagePercent()}`;
  }
</script>

{#if !sessionReady}
  <div style="min-height:100vh;display:grid;place-items:center;color:var(--muted)">{$t.status_loading_session}</div>
{:else if showShell}
  <div class="tenant-app-shell">
    <aside class="tenant-sidebar">
      <div class="tenant-sidebar-top">
        <a class="brand tenant-brand" href="{base}/backoffice">
          <img src="{base}/images/monti-logo.png" alt="" />
          <span><strong>MONTI</strong><small>{$t.brand_console}</small></span>
        </a>
        <div class="workspace-switcher">
          <span class="workspace-avatar">M</span>
          <span><small>{$t.workspace}</small><strong>Monti AI</strong></span><b>⌄</b>
        </div>
      </div>

      <nav class="tenant-nav" aria-label={$t.nav_aria}>
        <div class="nav-group">
          <div class="nav-group-label">{$t.nav_group_operations}</div>
          <a class="nav-link" href="{base}/backoffice" class:active={active('/backoffice')}
            ><span>⌂</span>{$t.nav_overview}</a
          >
          <a class="nav-link" href="{base}/dashboard" class:active={active('/dashboard')}
            ><span>▦</span>{$t.nav_call_center}</a
          >
          <a class="nav-link" href="{base}/monitoring" class:active={active('/monitoring')}
            ><span>◌</span>{$t.nav_monitoring}</a
          >
          <a class="nav-link" href="{base}/tickets" class:active={active('/tickets')}
            ><span>▱</span>{$t.nav_tickets}</a
          >
          <a class="nav-link" href="{base}/satisfaction" class:active={active('/satisfaction')}
            ><span>★</span>{$t.nav_satisfaction}</a
          >
          <a class="nav-link" href="{base}/preview" class:active={active('/preview')}
            ><span>◉</span>{$t.nav_preview} <em>{$t.nav_live}</em></a
          >
        </div>

        <div class="nav-group">
          <div class="nav-group-label">{$t.nav_group_knowledge}</div>
          <a class="nav-link" href="{base}/km" class:active={active('/km') && !active('/knowledge-gaps')}
            ><span>◫</span>{$t.nav_knowledge}</a
          >
          <a class="nav-link" href="{base}/knowledge-gaps" class:active={active('/knowledge-gaps')}
            ><span>△</span>{$t.nav_gaps}</a
          >
          <a
            class="nav-link"
            href="{base}/conversation-records"
            class:active={active('/conversation-records')}><span>▥</span>{$t.nav_records}</a
          >
        </div>

        <div class="nav-group">
          <div class="nav-group-label">{$t.nav_group_commerce}</div>
          <a
            class="nav-link"
            href="{base}/billing"
            class:active={active('/billing') && !active('/documents') && !active('/tax')}
            ><span>▣</span>{$t.nav_billing}</a
          >
          <a class="nav-link" href="{base}/billing/documents" class:active={active('/documents')}
            ><span>▤</span>{$t.nav_documents}</a
          >
          <a class="nav-link" href="{base}/billing/tax" class:active={active('/tax')}
            ><span>◇</span>{$t.nav_tax}</a
          >
        </div>

        <div class="nav-group">
          <div class="nav-group-label">{$t.nav_group_channels}</div>
          <a class="nav-link" href="{base}/avatars" class:active={active('/avatars')}
            ><span>☺</span>{$t.nav_avatars}</a
          >
          <a class="nav-link" href="{base}/embed" class:active={active('/embed')}
            ><span>⌘</span>{$t.nav_embed}</a
          >
          <a class="nav-link" href="{base}/theme" class:active={active('/theme')}
            ><span>◈</span>{$t.nav_theme}</a
          >
          <a class="nav-link" href="{base}/ai" class:active={active('/ai')}
            ><span>✦</span>{$t.nav_ai_config}</a
          >
        </div>

        <div class="nav-group">
          <div class="nav-group-label">{$t.nav_group_directory}</div>
          <a class="nav-link" href="{base}/customers" class:active={active('/customers')}
            ><span>♙</span>{$t.nav_customers}</a
          >
          <a class="nav-link" href="{base}/tiers" class:active={active('/tiers')}
            ><span>◆</span>{$t.nav_tiers}</a
          >
        </div>

        <div class="nav-group">
          <div class="nav-group-label">{$t.nav_group_growth}</div>
          <a class="nav-link" href="{base}/referrals" class:active={active('/referrals')}><span>↗</span>{$t.nav_referrals}</a>
        </div>

        <div class="nav-group">
          <div class="nav-group-label">{$t.nav_group_settings}</div>
          <a class="nav-link" href="{base}/settings" class:active={active('/settings')}
            ><span>⚙</span>{$t.nav_settings}</a
          >
        </div>
      </nav>

      <div class="tenant-sidebar-foot">
        <a class="plan-card plan-card-link" href="{base}/billing">
          <small>{$t.plan_current}</small>
          <strong>{currentPlan.data?.package?.name ?? (currentPlan.loading ? $t.plan_loading : $t.plan_none)}</strong>
          <span><i style={`width:${planUsagePercent()}%`}></i></span>
          <small>{#if currentPlan.loading && !currentPlan.data}{$t.plan_loading_allowance}{:else if currentPlan.data?.compact_utilization == null}{$t.plan_usage_unavailable}{:else}{planUsageLabel()}{$t.plan_usage_pct}{/if}</small>
        </a>
        {#if appVersion}
          <div style="padding:4px 12px 8px;font-size:11px;color:var(--muted)">{$t.app_version} {appVersion}</div>
        {/if}
        <LanguageSelector compact />
        <button class="account-button" type="button" onclick={logout}
          ><span class="workspace-avatar">AD</span><span
            ><strong>{$t.account_admin}</strong><small>{$t.account_sign_out}</small></span
          ><b>↗</b></button
        >
      </div>
    </aside>
    <section class="tenant-workspace">
      <header class="tenant-topbar">
        <div><span class="status-dot"></span> {$t.status_all_systems}</div>
        <div class="topbar-actions">
          <button aria-label={$t.topbar_search}>⌕</button><button aria-label={$t.topbar_notifications}>♢</button><a
            href="{base}/login"
            aria-label={$t.topbar_profile}>AD</a
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
