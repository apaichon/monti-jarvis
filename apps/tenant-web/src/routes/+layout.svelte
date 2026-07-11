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
  <header class="tenant-shell">
    <div class="tenant-shell-inner">
      <a class="brand link" href="{base}/billing">
        <img src="{base}/images/monti-logo.png" alt="Monti" />
        <span>MONTI TENANT</span>
      </a>
      <nav class="tenant-nav">
        <a class="link" href="{base}/backoffice">Backoffice</a>
        <a
          class="link"
          href="{base}/billing"
          class:active={$page.url.pathname.includes('/billing') &&
            !$page.url.pathname.includes('/documents') &&
            !$page.url.pathname.includes('/tax')}>Billing</a
        >
        <a
          class="link"
          href="{base}/billing/documents"
          class:active={$page.url.pathname.includes('/documents')}>Documents</a
        >
        <a class="link" href="{base}/billing/tax" class:active={$page.url.pathname.includes('/tax')}
          >Tax</a
        >
        <a class="link" href="{base}/login">Profile</a>
      </nav>
      <button class="btn ghost" type="button" onclick={logout}>Logout</button>
    </div>
  </header>
{/if}

{@render children()}
<FeedbackDialog />

<style>
  .tenant-shell {
    border-bottom: 1px solid var(--line);
    background: rgb(8 14 28 / 92%);
    position: sticky;
    top: 0;
    z-index: 10;
  }
  .tenant-shell-inner {
    max-width: 960px;
    margin: 0 auto;
    padding: 12px 20px;
    display: flex;
    align-items: center;
    gap: 16px;
    flex-wrap: wrap;
  }
  .tenant-nav {
    display: flex;
    gap: 16px;
    flex: 1;
    font-size: 14px;
  }
  .tenant-nav a.active {
    color: var(--cyan);
  }
</style>