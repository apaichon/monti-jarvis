<script lang="ts">
  import '../app.css';
  import { base } from '$app/paths';
  import { page } from '$app/stores';
  import { browser } from '$app/environment';
  import { captureAttributionFromSearch, demoHref, registerHref } from '$lib/attribution';
  import { postFunnelEvent } from '$lib/api';
  import { initLangFromUrl, lang, setLang, t, type Lang } from '$lib/i18n';

  let { children } = $props();

  let mobileOpen = $state(false);
  let solutionsOpen = $state(false);
  let solutionsCloseTimer: ReturnType<typeof setTimeout> | null = null;

  function isActive(match: string) {
    const path = $page.url.pathname;
    if (match === '/product/product') {
      return path === `${base}/product` || path === `${base}/product/`;
    }
    return path.includes(match);
  }

  function clearSolutionsCloseTimer() {
    if (solutionsCloseTimer != null) {
      clearTimeout(solutionsCloseTimer);
      solutionsCloseTimer = null;
    }
  }

  function openSolutions() {
    clearSolutionsCloseTimer();
    solutionsOpen = true;
  }

  function scheduleCloseSolutions() {
    clearSolutionsCloseTimer();
    // Small delay so the cursor can cross the gap into the panel without flicker.
    solutionsCloseTimer = setTimeout(() => {
      solutionsOpen = false;
      solutionsCloseTimer = null;
    }, 180);
  }

  function closeMenus() {
    clearSolutionsCloseTimer();
    mobileOpen = false;
    solutionsOpen = false;
  }

  function trackCta(ctaId: string) {
    if (!browser) return;
    void postFunnelEvent({
      event_name: 'cta_click',
      page_path: $page.url.pathname,
      cta_id: ctaId
    }).catch(() => {});
  }

  function switchLang(next: Lang) {
    setLang(next);
  }

  $effect(() => {
    if (!browser) return;
    captureAttributionFromSearch($page.url.searchParams);
    initLangFromUrl($page.url.searchParams);

    void postFunnelEvent({
      event_name: 'page_view',
      page_path: $page.url.pathname
    }).catch(() => {});
  });

  // Close menus on route change so the panel never stays open after navigation.
  $effect(() => {
    void $page.url.pathname;
    closeMenus();
  });
</script>

<div class="site" data-lang={$lang}>
  <header class="topbar">
    <div class="topbar-inner">
      <a class="brand" href="{base}/" aria-label="Monti home">
        <img src="{base}/images/monti-logo.png" alt="" width="36" height="36" />
        <span>
          <strong>MONTI</strong>
          <small class:th={$lang === 'th'} class:ja={$lang === 'ja'}>{$t.brand_tagline}</small>
        </span>
      </a>

      <nav class="nav-desktop" aria-label="Primary">
        <a class="nav-link" class:active={isActive('/product/product')} href="{base}/product">{$t.nav_product}</a>

        <div
          class="nav-item has-children"
          class:open={solutionsOpen}
          onmouseenter={openSolutions}
          onmouseleave={scheduleCloseSolutions}
        >
          <button
            type="button"
            class="nav-link nav-parent"
            class:active={isActive('/solutions')}
            aria-expanded={solutionsOpen}
            aria-haspopup="true"
            onclick={() => {
              clearSolutionsCloseTimer();
              solutionsOpen = !solutionsOpen;
            }}
          >
            {$t.nav_solutions}
            <span class="caret" aria-hidden="true">▾</span>
          </button>
          <!-- Always mounted so hover can travel into the panel without unmount flicker. -->
          <div class="nav-dropdown" role="menu" class:visible={solutionsOpen} aria-hidden={!solutionsOpen}>
            <a
              role="menuitem"
              class="nav-drop-link"
              class:active={$page.url.pathname === `${base}/solutions` ||
                $page.url.pathname === `${base}/solutions/`}
              href="{base}/solutions"
              tabindex={solutionsOpen ? 0 : -1}
              onclick={closeMenus}
            >
              {$t.nav_solutions_industry}
            </a>
            <a
              role="menuitem"
              class="nav-drop-link"
              class:active={isActive('/solutions/enterprise')}
              href="{base}/solutions/enterprise"
              tabindex={solutionsOpen ? 0 : -1}
              onclick={closeMenus}
            >
              {$t.nav_solutions_enterprise}
            </a>
          </div>
        </div>

        <a class="nav-link" class:active={isActive('/resources')} href="{base}/resources">{$t.nav_resources}</a>
        <a class="nav-link" class:active={isActive('/pricing')} href="{base}/pricing">{$t.nav_pricing}</a>
        <a class="nav-link" class:active={isActive('/about')} href="{base}/about">{$t.nav_about}</a>
      </nav>

      <div class="top-actions">
        <div class="lang-toggle" role="group" aria-label={$t.lang_label}>
          <button
            type="button"
            class:active={$lang === 'en'}
            onclick={() => switchLang('en')}
            aria-pressed={$lang === 'en'}>EN</button
          >
          <button
            type="button"
            class:active={$lang === 'th'}
            onclick={() => switchLang('th')}
            aria-pressed={$lang === 'th'}>TH</button
          >
          <button
            type="button"
            class:active={$lang === 'ja'}
            onclick={() => switchLang('ja')}
            aria-pressed={$lang === 'ja'}>JA</button
          >
        </div>
        <a class="nav-login" href="/tenant/login" onclick={() => trackCta('nav_login')}>{$t.nav_login}</a>
        <a
          class="btn sm book"
          href="{base}/contact?kind=book_demo"
          onclick={() => trackCta('nav_book_demo')}>{$t.nav_book_demo}</a
        >
        <button
          class="menu-btn"
          type="button"
          aria-label={$t.nav_open_menu}
          aria-expanded={mobileOpen}
          onclick={() => (mobileOpen = !mobileOpen)}
        >
          ☰
        </button>
      </div>
    </div>

    {#if mobileOpen}
      <div class="nav-mobile" role="navigation" aria-label="Mobile">
        <a class="nav-link" href="{base}/product" onclick={closeMenus}>{$t.nav_product}</a>
        <div class="nav-mobile-group">
          <span class="nav-mobile-label">{$t.nav_solutions}</span>
          <a class="nav-link nested" href="{base}/solutions" onclick={closeMenus}
            >{$t.nav_solutions_industry}</a
          >
          <a class="nav-link nested" href="{base}/solutions/enterprise" onclick={closeMenus}
            >{$t.nav_solutions_enterprise}</a
          >
        </div>
        <a class="nav-link" href="{base}/resources" onclick={closeMenus}>{$t.nav_resources}</a>
        <a class="nav-link" href="{base}/pricing" onclick={closeMenus}>{$t.nav_pricing}</a>
        <a class="nav-link" href="{base}/about" onclick={closeMenus}>{$t.nav_about}</a>
        <a href="{base}/contact" onclick={closeMenus}>{$t.nav_contact}</a>
        <a href="{base}/demo" onclick={closeMenus}>{$t.nav_demo}</a>
      </div>
    {/if}
  </header>

  <main class="main">
    {@render children()}
  </main>

  <footer class="footer">
    <div class="footer-inner">
      <div class="footer-brand">
        <img src="{base}/images/monti-logo.png" alt="" width="28" height="28" />
        <div>
          <strong>Monti</strong>
          <p class="muted">{$t.footer_blurb}</p>
        </div>
      </div>
      <div class="footer-cols">
        <div>
          <h3>{$t.footer_product}</h3>
          <a href="{base}/product">{$t.footer_overview}</a>
          <a href="{base}/solutions">{$t.nav_solutions_industry}</a>
          <a href="{base}/solutions/enterprise">{$t.nav_solutions_enterprise}</a>
          <a href="{base}/pricing">{$t.nav_pricing}</a>
          <a href="{base}/resources">{$t.nav_resources}</a>
        </div>
        <div>
          <h3>{$t.footer_get_started}</h3>
          <a href={demoHref()} onclick={() => trackCta('footer_demo')}>{$t.footer_live_demo}</a>
          <a href="{base}/contact?kind=book_demo">{$t.nav_book_demo}</a>
          <a href={registerHref()} onclick={() => trackCta('footer_register')}>{$t.footer_register}</a>
          <a href="{base}/contact">{$t.footer_contact_sales}</a>
        </div>
        <div>
          <h3>{$t.footer_company}</h3>
          <a href="{base}/about">{$t.nav_about}</a>
          <a href="{base}/demo">{$t.footer_demo_guide}</a>
        </div>
      </div>
    </div>
    <div class="footer-bottom">
      <span class="muted">© {new Date().getFullYear()} {$t.footer_rights}</span>
      <span class="muted">{$t.footer_care}</span>
    </div>
  </footer>
</div>

<style>
  .site {
    min-height: 100vh;
    display: grid;
    grid-template-rows: auto 1fr auto;
  }

  .topbar {
    position: sticky;
    top: 0;
    z-index: 40;
    border-bottom: 1px solid rgb(60 110 200 / 14%);
    background: rgb(2 6 15 / 86%);
    backdrop-filter: blur(18px);
  }

  .topbar-inner {
    width: min(1180px, 100%);
    margin: 0 auto;
    padding: 12px 24px;
    display: flex;
    align-items: center;
    gap: 18px;
  }

  .brand {
    display: flex;
    align-items: center;
    gap: 10px;
    text-decoration: none;
    color: var(--ink);
    flex-shrink: 0;
  }

  .brand img {
    width: 38px;
    height: 38px;
    border-radius: 50%;
    object-fit: cover;
    box-shadow: 0 0 0 2px rgb(40 110 230 / 35%), 0 0 22px rgb(35 117 255 / 36%);
  }

  .brand span {
    display: grid;
    gap: 1px;
  }

  .brand strong {
    letter-spacing: 0.22em;
    font-size: 14px;
    font-weight: 750;
  }

  .brand small {
    color: #6f8ab5;
    font-size: 8px;
    letter-spacing: 0.14em;
  }

  .brand small.th,
  .brand small.ja {
    font-size: 10px;
    letter-spacing: 0.04em;
  }

  .nav-desktop {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 2px;
    flex: 1;
  }

  .nav-item {
    position: relative;
    /* Keep hover hitbox continuous with the dropdown (no dead gap). */
    padding-bottom: 0;
  }

  .nav-item.has-children {
    /* Extends the hover area slightly under the trigger toward the panel. */
    padding-bottom: 10px;
    margin-bottom: -10px;
  }

  .nav-link {
    color: #a7b6cf;
    text-decoration: none;
    font-size: 13px;
    font-weight: 500;
    padding: 8px 14px;
    border-radius: 8px;
    border: 1px solid transparent;
    transition: 0.15s ease;
    background: transparent;
    font-family: inherit;
    cursor: pointer;
    display: inline-flex;
    align-items: center;
    gap: 4px;
  }

  .nav-link:hover,
  .nav-link.active,
  .nav-item.open > .nav-parent {
    color: white;
    background: rgb(30 70 150 / 18%);
  }

  .caret {
    font-size: 10px;
    opacity: 0.75;
    transform: translateY(1px);
  }

  .nav-dropdown {
    position: absolute;
    top: 100%;
    left: 0;
    min-width: 200px;
    /* Invisible bridge above the panel so the cursor never leaves the hover zone. */
    padding: 10px 0 0;
    margin: 0;
    border: 0;
    background: transparent;
    box-shadow: none;
    display: grid;
    gap: 2px;
    z-index: 50;
    opacity: 0;
    visibility: hidden;
    pointer-events: none;
    transform: translateY(4px);
    transition:
      opacity 0.12s ease,
      transform 0.12s ease,
      visibility 0.12s ease;
  }

  .nav-dropdown.visible {
    opacity: 1;
    visibility: visible;
    pointer-events: auto;
    transform: translateY(0);
  }

  /* Visible card sits inside the bridged padding area. */
  .nav-dropdown::before {
    content: '';
    position: absolute;
    inset: 10px 0 0;
    border-radius: 12px;
    border: 1px solid rgb(70 120 200 / 28%);
    background: rgb(6 12 28 / 98%);
    box-shadow: 0 18px 40px rgb(0 8 24 / 45%);
    backdrop-filter: blur(16px);
    z-index: -1;
  }

  .nav-drop-link {
    position: relative;
    z-index: 1;
    display: block;
    margin: 0 8px;
    padding: 10px 12px;
    border-radius: 8px;
    color: #c4d2ea;
    text-decoration: none;
    font-size: 13px;
    font-weight: 550;
  }

  .nav-drop-link:first-child {
    margin-top: 8px;
  }

  .nav-drop-link:last-child {
    margin-bottom: 8px;
  }

  .nav-drop-link:hover,
  .nav-drop-link.active {
    color: #fff;
    background: rgb(30 70 150 / 28%);
  }

  .nav-mobile-group {
    display: grid;
    gap: 2px;
  }

  .nav-mobile-label {
    padding: 8px 12px 4px;
    font-size: 11px;
    letter-spacing: 0.1em;
    text-transform: uppercase;
    color: #6d7c99;
    font-weight: 700;
  }

  .nav-mobile .nav-link.nested {
    padding-left: 22px;
    color: #a7b6cf;
  }

  .nav-login {
    color: #c4d2ea;
    text-decoration: none;
    font-size: 13px;
    font-weight: 600;
    padding: 8px 10px;
  }

  .nav-login:hover {
    color: #fff;
  }

  .btn.book {
    border: 0;
    background: linear-gradient(100deg, #1f7bff, #2f6dff);
    color: #fff;
    box-shadow: 0 8px 22px rgb(31 123 255 / 28%);
  }

  .top-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    flex-shrink: 0;
  }

  .lang-toggle {
    display: inline-flex;
    border: 1px solid var(--line);
    border-radius: 8px;
    overflow: hidden;
  }

  .lang-toggle button {
    border: 0;
    background: transparent;
    color: var(--muted);
    padding: 6px 8px;
    font-size: 11px;
    font-weight: 700;
    min-width: 32px;
  }

  .lang-toggle button.active {
    background: rgb(35 117 255 / 22%);
    color: var(--ink);
  }

  .menu-btn {
    display: none;
    width: 36px;
    height: 36px;
    border: 1px solid var(--line);
    border-radius: 9px;
    background: rgb(15 25 43 / 60%);
    color: var(--ink);
  }

  .nav-mobile {
    display: none;
    flex-direction: column;
    gap: 4px;
    padding: 8px 24px 16px;
    border-top: 1px solid var(--line);
  }

  .nav-mobile a {
    color: #c4d0e4;
    text-decoration: none;
    padding: 10px 12px;
    border-radius: 8px;
  }

  .main {
    min-width: 0;
  }

  .footer {
    border-top: 1px solid var(--line);
    background: linear-gradient(180deg, rgb(4 8 18 / 40%), rgb(2 4 12 / 90%));
    margin-top: 24px;
  }

  .footer-inner {
    width: min(1180px, 100%);
    margin: 0 auto;
    padding: 40px 24px 24px;
    display: grid;
    grid-template-columns: 1.2fr 2fr;
    gap: 32px;
  }

  .footer-brand {
    display: flex;
    gap: 12px;
    align-items: flex-start;
  }

  .footer-brand img {
    border-radius: 50%;
  }

  .footer-brand strong {
    letter-spacing: 0.16em;
    font-size: 13px;
  }

  .footer-brand p {
    margin: 6px 0 0;
    font-size: 13px;
    line-height: 1.5;
    max-width: 28ch;
  }

  .footer-cols {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 18px;
  }

  .footer-cols h3 {
    margin: 0 0 10px;
    font-size: 12px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: #6d7c99;
  }

  .footer-cols a {
    display: block;
    color: #b4c1d8;
    text-decoration: none;
    font-size: 13px;
    margin-bottom: 8px;
  }

  .footer-cols a:hover {
    color: var(--cyan);
  }

  .footer-bottom {
    width: min(1180px, 100%);
    margin: 0 auto;
    padding: 14px 24px 28px;
    display: flex;
    justify-content: space-between;
    gap: 12px;
    flex-wrap: wrap;
    border-top: 1px solid rgb(70 100 150 / 14%);
    font-size: 12px;
  }

  @media (max-width: 980px) {
    .nav-desktop {
      display: none;
    }

    .menu-btn {
      display: grid;
      place-items: center;
    }

    .nav-mobile {
      display: flex;
    }

    .footer-inner {
      grid-template-columns: 1fr;
    }

    .footer-cols {
      grid-template-columns: 1fr 1fr;
    }
  }

  @media (max-width: 560px) {
    .topbar-inner {
      padding: 12px 16px;
    }

    .nav-login {
      display: none;
    }

    .footer-cols {
      grid-template-columns: 1fr;
    }
  }
</style>
