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

  const pillars = $derived([
    { symbol: '〰', title: $t.about_val_customer_t, body: $t.about_val_customer_b },
    { symbol: '⚡', title: $t.about_val_innov_t, body: $t.about_val_innov_b },
    { symbol: '👥', title: $t.about_val_integ_t, body: $t.about_val_integ_b },
    { symbol: '🛡', title: $t.about_val_impact_t, body: $t.about_val_impact_b }
  ]);
</script>

<svelte:head>
  <title>{$t.about_title}</title>
  <meta name="description" content={$t.about_p1} />
</svelte:head>

<section class="about-page">
  <div class="glow" aria-hidden="true"></div>

  <div class="intro">
    <h1>{$t.about_h1}</h1>
    <p class="lede">{$t.about_p1}</p>
  </div>

  <!-- Design hero: conversation visual (image cropped to omit Trusted by logos). -->
  <div class="hero-frame">
    <a
      class="hero-art"
      href="{base}/demo"
      onclick={() => track('about_hero_demo')}
      aria-label={$t.about_story}
    >
      <img
        src="{base}/images/about-hero.png"
        alt="Monti AI call center — natural voice conversation between agent and customer"
        width="1672"
        height="941"
        decoding="async"
      />
      <span class="play-hint" aria-hidden="true">
        <span class="play-ring">
          <span class="play-tri">▶</span>
        </span>
      </span>
    </a>
  </div>

  <div class="pillars">
    {#each pillars as item}
      <article class="pillar">
        <span class="pillar-icon" aria-hidden="true">{item.symbol}</span>
        <h2>{item.title}</h2>
        <p>{item.body}</p>
      </article>
    {/each}
  </div>

  <aside class="cta-banner">
    <div class="cta-mascot">
      <img src="{base}/images/monti-logo.png" alt="" width="72" height="72" />
    </div>
    <div class="cta-copy">
      <h2>{$t.about_cta_h2}</h2>
      <p>{$t.about_cta_p}</p>
    </div>
    <a
      class="btn-demo"
      href="{base}/contact?kind=book_demo"
      onclick={() => track('about_book_demo')}
    >
      {$t.about_cta_btn}
    </a>
  </aside>
</section>

<style>
  .about-page {
    position: relative;
    overflow: hidden;
    padding: 48px 24px 72px;
    background:
      radial-gradient(circle at 50% 18%, rgb(30 110 255 / 16%), transparent 32%),
      radial-gradient(circle at 15% 70%, rgb(20 60 160 / 10%), transparent 28%),
      linear-gradient(180deg, #020713 0%, #031025 52%, #020713 100%);
  }

  .glow {
    position: absolute;
    inset: 0;
    pointer-events: none;
    opacity: 0.28;
    background-image:
      linear-gradient(rgb(30 90 180 / 12%) 1px, transparent 1px),
      linear-gradient(90deg, rgb(30 90 180 / 12%) 1px, transparent 1px);
    background-size: 56px 56px;
    mask-image: linear-gradient(180deg, #000, transparent 72%);
  }

  .intro {
    position: relative;
    z-index: 1;
    width: min(860px, 100%);
    margin: 0 auto 28px;
    text-align: center;
  }

  .intro h1 {
    margin: 0 0 14px;
    font-size: clamp(2.1rem, 4.4vw, 3.15rem);
    line-height: 1.08;
    letter-spacing: -0.035em;
    font-weight: 750;
  }

  .lede {
    margin: 0 auto;
    color: #93a6c4;
    font-size: clamp(1rem, 1.6vw, 1.15rem);
    line-height: 1.55;
    max-width: 40ch;
  }

  .hero-frame {
    position: relative;
    z-index: 1;
    width: min(1100px, 100%);
    margin: 0 auto 36px;
  }

  .hero-art {
    position: relative;
    display: block;
    border-radius: 20px;
    overflow: hidden;
    border: 1px solid rgb(70 130 220 / 28%);
    box-shadow:
      0 28px 70px rgb(0 12 40 / 45%),
      0 0 0 1px rgb(40 100 255 / 8%),
      inset 0 1px 0 rgb(255 255 255 / 4%);
    text-decoration: none;
    background: #040a16;
    /* Crop bottom strip that contains “Trusted by …” logos from the source art. */
    aspect-ratio: 16 / 8.2;
  }

  .hero-art img {
    width: 100%;
    height: 118%;
    object-fit: cover;
    object-position: center 38%;
    display: block;
    transform: scale(1.02);
    transition: transform 0.35s ease;
  }

  .hero-art:hover img {
    transform: scale(1.04);
  }

  .play-hint {
    position: absolute;
    left: 50%;
    top: 52%;
    transform: translate(-50%, -50%);
    pointer-events: none;
    opacity: 0;
    transition: opacity 0.2s ease;
  }

  .hero-art:hover .play-hint,
  .hero-art:focus-visible .play-hint {
    opacity: 1;
  }

  .play-ring {
    width: 72px;
    height: 72px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: rgb(15 50 140 / 55%);
    border: 2px solid rgb(120 190 255 / 70%);
    box-shadow:
      0 0 0 10px rgb(40 120 255 / 18%),
      0 12px 36px rgb(0 30 80 / 45%);
    backdrop-filter: blur(8px);
  }

  .play-tri {
    color: #fff;
    font-size: 20px;
    margin-left: 3px;
  }

  .pillars {
    position: relative;
    z-index: 1;
    width: min(1100px, 100%);
    margin: 0 auto 36px;
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 14px;
  }

  .pillar {
    text-align: center;
    padding: 22px 16px 20px;
    border-radius: 16px;
    border: 1px solid rgb(70 120 200 / 22%);
    background: linear-gradient(165deg, rgb(10 22 48 / 88%), rgb(5 12 28 / 92%));
    box-shadow: 0 14px 36px rgb(0 10 30 / 22%);
  }

  .pillar-icon {
    width: 48px;
    height: 48px;
    margin: 0 auto 14px;
    border-radius: 14px;
    display: grid;
    place-items: center;
    font-size: 20px;
    color: #7ec0ff;
    background: rgb(20 60 150 / 30%);
    border: 1px solid rgb(70 140 240 / 30%);
    box-shadow: 0 0 22px rgb(40 120 255 / 14%);
  }

  .pillar h2 {
    margin: 0 0 8px;
    font-size: 15px;
    font-weight: 700;
    letter-spacing: -0.01em;
  }

  .pillar p {
    margin: 0;
    color: #8ea0bd;
    font-size: 13px;
    line-height: 1.5;
  }

  .cta-banner {
    position: relative;
    z-index: 1;
    width: min(1100px, 100%);
    margin: 0 auto;
    padding: 22px 26px;
    border-radius: 18px;
    border: 1px solid rgb(70 130 220 / 28%);
    background: linear-gradient(120deg, rgb(8 20 48 / 92%), rgb(6 14 34 / 95%));
    box-shadow: 0 20px 50px rgb(0 12 40 / 30%);
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 20px;
    align-items: center;
  }

  .cta-mascot {
    width: 80px;
    height: 80px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background:
      radial-gradient(circle at 35% 30%, rgb(80 160 255 / 22%), transparent 55%),
      rgb(8 18 40 / 80%);
    border: 1px solid rgb(70 140 240 / 30%);
    box-shadow: 0 0 28px rgb(40 120 255 / 16%);
  }

  .cta-mascot img {
    width: 60px;
    height: 60px;
    border-radius: 50%;
    object-fit: cover;
  }

  .cta-copy h2 {
    margin: 0 0 6px;
    font-size: 1.2rem;
    letter-spacing: -0.02em;
  }

  .cta-copy p {
    margin: 0;
    color: #93a6c4;
    font-size: 14px;
    line-height: 1.5;
    max-width: 48ch;
  }

  .btn-demo {
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

  .btn-demo:hover {
    filter: brightness(1.06);
  }

  @media (max-width: 960px) {
    .pillars {
      grid-template-columns: 1fr 1fr;
    }

    .hero-art {
      aspect-ratio: 16 / 9;
    }
  }

  @media (max-width: 600px) {
    .about-page {
      padding: 32px 16px 48px;
    }

    .pillars {
      grid-template-columns: 1fr;
    }

    .cta-banner {
      grid-template-columns: 1fr;
      text-align: center;
      justify-items: center;
    }

    .cta-copy p {
      max-width: none;
    }

    .hero-art {
      aspect-ratio: 4 / 3;
    }

    .hero-art img {
      height: 130%;
      object-position: center 42%;
    }
  }
</style>
