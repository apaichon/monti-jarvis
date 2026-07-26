<script lang="ts">
  import { base } from '$app/paths';
  import { page } from '$app/stores';
  import { postFunnelEvent } from '$lib/api';
  import { t } from '$lib/i18n';

  function track(ctaId: string) {
    void postFunnelEvent({
      event_name: 'cta_click',
      page_path: $page.url.pathname,
      cta_id: ctaId
    }).catch(() => {});
  }

  const industries = $derived([
    { id: 'customer-support', icon: '👥', title: $t.solutions_ind_cs_t, body: $t.solutions_ind_cs_b },
    { id: 'financial-services', icon: '🏛', title: $t.solutions_ind_fin_t, body: $t.solutions_ind_fin_b },
    { id: 'e-commerce', icon: '🛒', title: $t.solutions_ind_ecom_t, body: $t.solutions_ind_ecom_b },
    { id: 'healthcare', icon: '✚', title: $t.solutions_ind_health_t, body: $t.solutions_ind_health_b },
    { id: 'travel-hospitality', icon: '📅', title: $t.solutions_ind_travel_t, body: $t.solutions_ind_travel_b },
    { id: 'telecom', icon: '↻', title: $t.solutions_ind_tel_t, body: $t.solutions_ind_tel_b },
    { id: 'education', icon: '⌂', title: $t.solutions_ind_edu_t, body: $t.solutions_ind_edu_b },
    { id: 'public-sector', icon: '🏛', title: $t.solutions_ind_gov_t, body: $t.solutions_ind_gov_b }
  ]);
</script>

<svelte:head>
  <title>{$t.solutions_title}</title>
  <meta
    name="description"
    content={$t.solutions_p}
  />
</svelte:head>

<section class="solutions-page">
  <div class="glow" aria-hidden="true"></div>

  <div class="intro">
    <h1>{$t.solutions_h1}</h1>
    <p>{$t.solutions_p}</p>
  </div>

  <div class="industry-grid">
    {#each industries as item}
      <article class="industry-card" id={item.id}>
        <span class="industry-icon" aria-hidden="true">{item.icon}</span>
        <h2>{item.title}</h2>
        <p>{item.body}</p>
        <a
          class="learn-more"
          href="{base}/contact?kind=book_demo&use_case={encodeURIComponent(item.title)}"
          onclick={() => track(`solution_${item.id}`)}
        >
          {$t.solutions_learn} <span aria-hidden="true">→</span>
        </a>
      </article>
    {/each}
  </div>

  <aside class="expert-banner">
    <div class="expert-mascot">
      <img src="{base}/images/monti-logo.png" alt="Monti" width="88" height="88" />
    </div>
    <div class="expert-copy">
      <h2>{$t.solutions_expert_h2}</h2>
      <p>{$t.solutions_expert_p}</p>
    </div>
    <a
      class="btn-expert"
      href="{base}/contact?kind=book_demo"
      onclick={() => track('solutions_talk_expert')}
    >
      {$t.solutions_expert_cta}
    </a>
    <div class="banner-wave" aria-hidden="true"></div>
  </aside>
</section>

<style>
  .solutions-page {
    position: relative;
    overflow: hidden;
    padding: 52px 24px 72px;
    background:
      radial-gradient(circle at 70% 8%, rgb(25 90 220 / 12%), transparent 28%),
      radial-gradient(circle at 20% 60%, rgb(15 50 140 / 10%), transparent 30%),
      linear-gradient(180deg, #020713 0%, #031025 55%, #020713 100%);
  }

  .glow {
    position: absolute;
    inset: 0;
    pointer-events: none;
    background-image:
      linear-gradient(rgb(30 90 180 / 10%) 1px, transparent 1px),
      linear-gradient(90deg, rgb(30 90 180 / 10%) 1px, transparent 1px);
    background-size: 56px 56px;
    mask-image: linear-gradient(180deg, #000 0%, transparent 70%);
    opacity: 0.35;
  }

  .intro {
    position: relative;
    z-index: 1;
    width: min(920px, 100%);
    margin: 0 auto 36px;
    text-align: left;
  }

  .intro h1 {
    margin: 0 0 14px;
    font-size: clamp(2rem, 4vw, 2.85rem);
    line-height: 1.1;
    letter-spacing: -0.03em;
    font-weight: 750;
  }

  .intro p {
    margin: 0;
    color: #93a6c4;
    font-size: 1.05rem;
    line-height: 1.55;
  }

  .industry-grid {
    position: relative;
    z-index: 1;
    width: min(1100px, 100%);
    margin: 0 auto;
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 14px;
  }

  .industry-card {
    display: flex;
    flex-direction: column;
    min-height: 210px;
    padding: 22px 18px 18px;
    border-radius: 16px;
    border: 1px solid rgb(70 120 200 / 22%);
    background: linear-gradient(165deg, rgb(10 22 48 / 88%), rgb(5 12 28 / 92%));
    box-shadow:
      0 16px 40px rgb(0 10 30 / 28%),
      inset 0 1px 0 rgb(255 255 255 / 3%);
    transition:
      border-color 0.15s ease,
      transform 0.15s ease,
      box-shadow 0.15s ease;
  }

  .industry-card:hover {
    border-color: rgb(70 140 255 / 40%);
    transform: translateY(-2px);
    box-shadow: 0 18px 44px rgb(20 60 160 / 18%);
  }

  .industry-icon {
    width: 42px;
    height: 42px;
    border-radius: 12px;
    display: grid;
    place-items: center;
    margin-bottom: 16px;
    font-size: 18px;
    color: #5aa8ff;
    background: rgb(20 60 150 / 30%);
    border: 1px solid rgb(70 140 240 / 28%);
    box-shadow: 0 0 18px rgb(40 120 255 / 12%);
  }

  .industry-card h2 {
    margin: 0 0 10px;
    font-size: 1.05rem;
    font-weight: 700;
    letter-spacing: -0.01em;
  }

  .industry-card p {
    margin: 0 0 16px;
    color: #8ea0bd;
    font-size: 13px;
    line-height: 1.55;
    flex: 1;
  }

  .learn-more {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: #4ea3ff;
    font-size: 13px;
    font-weight: 650;
    text-decoration: none;
  }

  .learn-more:hover {
    color: #7ec0ff;
  }

  .expert-banner {
    position: relative;
    z-index: 1;
    width: min(1100px, 100%);
    margin: 28px auto 0;
    padding: 22px 28px;
    border-radius: 18px;
    border: 1px solid rgb(70 130 220 / 28%);
    background: linear-gradient(120deg, rgb(8 20 48 / 92%), rgb(6 14 34 / 95%) 55%, rgb(10 28 70 / 70%));
    box-shadow:
      0 20px 50px rgb(0 12 40 / 35%),
      inset 0 1px 0 rgb(255 255 255 / 4%);
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 22px;
    align-items: center;
    overflow: hidden;
  }

  .expert-mascot {
    width: 96px;
    height: 96px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background:
      radial-gradient(circle at 35% 30%, rgb(80 160 255 / 25%), transparent 55%),
      rgb(8 18 40 / 80%);
    border: 1px solid rgb(70 140 240 / 30%);
    box-shadow: 0 0 30px rgb(40 120 255 / 18%);
    flex-shrink: 0;
  }

  .expert-mascot img {
    width: 72px;
    height: 72px;
    border-radius: 50%;
    object-fit: cover;
  }

  .expert-copy h2 {
    margin: 0 0 6px;
    font-size: 1.25rem;
    letter-spacing: -0.02em;
  }

  .expert-copy p {
    margin: 0;
    color: #93a6c4;
    font-size: 14px;
    line-height: 1.5;
    max-width: 42ch;
  }

  .btn-expert {
    position: relative;
    z-index: 1;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    border-radius: 10px;
    padding: 12px 20px;
    background: linear-gradient(100deg, #1f7bff, #2f6dff);
    color: #fff;
    font-weight: 700;
    font-size: 14px;
    text-decoration: none;
    box-shadow: 0 10px 28px rgb(31 123 255 / 30%);
    white-space: nowrap;
  }

  .btn-expert:hover {
    filter: brightness(1.06);
  }

  .banner-wave {
    position: absolute;
    right: -40px;
    top: 0;
    bottom: 0;
    width: 42%;
    pointer-events: none;
    background:
      radial-gradient(ellipse at 70% 50%, rgb(30 100 255 / 18%), transparent 55%),
      repeating-linear-gradient(
        100deg,
        transparent 0 18px,
        rgb(40 120 255 / 8%) 18px 19px
      );
    mask-image: linear-gradient(90deg, transparent, #000 40%);
    opacity: 0.85;
  }

  @media (max-width: 1020px) {
    .industry-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }
  }

  @media (max-width: 720px) {
    .solutions-page {
      padding: 32px 16px 48px;
    }

    .intro {
      text-align: left;
    }

    .industry-grid {
      grid-template-columns: 1fr;
    }

    .expert-banner {
      grid-template-columns: 1fr;
      text-align: center;
      justify-items: center;
      padding: 24px 20px;
    }

    .expert-copy p {
      max-width: none;
    }

    .banner-wave {
      width: 100%;
      opacity: 0.35;
    }
  }
</style>
