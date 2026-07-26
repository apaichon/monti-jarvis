<script lang="ts">
  import { base } from '$app/paths';
  import { page } from '$app/stores';
  import { demoHref, registerHref } from '$lib/attribution';
  import { postFunnelEvent } from '$lib/api';
  import { t } from '$lib/i18n';

  function startDemo() {
    void postFunnelEvent({
      event_name: 'demo_start',
      page_path: $page.url.pathname,
      cta_id: 'demo_page_start'
    }).catch(() => {});
  }
</script>

<svelte:head>
  <title>{$t.demo_title}</title>
</svelte:head>

<section class="section">
  <span class="badge">{$t.demo_badge}</span>
  <h1 class="section-title" style="margin-top:12px">{$t.demo_h1}</h1>
  <p class="section-lede">{$t.demo_lede}</p>

  <div class="grid-2">
    <article class="card">
      <h2>{$t.demo_try_h2}</h2>
      <ul>
        <li>{$t.demo_try_1}</li>
        <li>{$t.demo_try_2}</li>
        <li>{$t.demo_try_3}</li>
        <li>{$t.demo_try_4}</li>
      </ul>
    </article>
    <article class="card">
      <h2>{$t.demo_after_h2}</h2>
      <ul>
        <li>{$t.demo_after_1}</li>
        <li>{$t.demo_after_2}</li>
        <li>{$t.demo_after_3}</li>
        <li>{$t.demo_after_4}</li>
      </ul>
    </article>
  </div>

  <div class="cta-band card">
    <div>
      <h2>{$t.demo_open_h2}</h2>
      <p class="muted">{$t.demo_open_p}</p>
    </div>
    <div class="cta-row">
      <a class="btn cyan" href={demoHref()} onclick={startDemo}>{$t.demo_launch}</a>
      <a class="btn ghost" href="{base}/contact?kind=book_demo">{$t.demo_book}</a>
      <a class="btn ghost" href={registerHref()}>{$t.demo_register}</a>
    </div>
  </div>
</section>

<style>
  h2 {
    margin: 0 0 10px;
    font-size: 17px;
  }

  ul {
    margin: 0;
    padding-left: 18px;
    display: grid;
    gap: 8px;
    color: #c4d0e4;
    font-size: 14px;
    line-height: 1.45;
  }

  .cta-band {
    margin-top: 24px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 20px;
    flex-wrap: wrap;
  }

  .cta-band h2 {
    margin: 0 0 6px;
    font-size: 1.35rem;
  }

  .cta-band p {
    margin: 0;
    font-size: 13px;
  }

  .cta-row {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }
</style>
