<script lang="ts">
  import '../app.css';
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

{#if onLoginPage}
  {@render children()}
{:else}
  <div class="shell">
    <header class="topnav">
      <a class="brand" href="{base}/packages">
        <img src="{base}/images/monti-logo.png" alt="Monti" />
        <strong>MONTI ADMIN</strong>
      </a>
      <nav class="nav-links">
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/packages`)}
          href="{base}/packages"
        >
          Packages
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/avatars`)}
          href="{base}/avatars"
        >
          Avatars
        </a>
        <a
          class="nav-link"
          class:active={$page.url.pathname.startsWith(`${base}/profile`)}
          href="{base}/profile"
        >
          Profile
        </a>
      </nav>
      <div class="nav-right">
        <span>{user?.email ?? '—'}</span>
        <button class="btn ghost" type="button" onclick={handleLogout}>Logout</button>
      </div>
    </header>
    <main class="main">
      {@render children()}
    </main>
  </div>
{/if}