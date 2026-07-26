<script lang="ts">
  import { base } from '$app/paths';
  import { page } from '$app/stores';
  import { demoHref, registerHref } from '$lib/attribution';
  import { postFunnelEvent } from '$lib/api';
  import { t } from '$lib/i18n';

  function track(ctaId: string, event: 'cta_click' | 'demo_start' | 'register_start' = 'cta_click') {
    void postFunnelEvent({
      event_name: event,
      page_path: $page.url.pathname,
      cta_id: ctaId
    }).catch(() => {});
  }

  const productNav = $derived([
    { icon: '⌂', label: $t.product_nav_overview, href: `${base}/product`, active: true },
    { icon: '◉', label: $t.product_nav_voice, href: `${base}/product#ai-voice-agents` },
    { icon: '⌁', label: $t.product_nav_omni, href: `${base}/product#omnichannel` },
    { icon: '▣', label: $t.product_nav_km, href: `${base}/product#knowledge-hub` },
    { icon: '↗', label: $t.product_nav_handover, href: `${base}/product#live-handover` },
    { icon: '◌', label: $t.product_nav_analytics, href: `${base}/product#analytics` },
    { icon: '♢', label: $t.product_nav_security, href: `${base}/product#security` },
    { icon: '⌘', label: $t.product_nav_integrations, href: `${base}/product#integrations` }
  ]);

  const knowledgeItems = $derived([$t.product_km_1, $t.product_km_2, $t.product_km_3]);

  const benefits = $derived([
    { icon: '👥', title: $t.product_ben_easy_t, body: $t.product_ben_easy_b },
    { icon: '♟', title: $t.product_ben_scale_t, body: $t.product_ben_scale_b },
    { icon: '⧉', title: $t.product_ben_flex_t, body: $t.product_ben_flex_b },
    { icon: '🛡', title: $t.product_ben_secure_t, body: $t.product_ben_secure_b }
  ]);
</script>

<svelte:head>
  <title>{$t.product_title}</title>
  <meta
    name="description"
    content="Monti Product Suite: AI voice agents, knowledge hub, live handover, analytics, and secure integrations for modern call centers."
  />
</svelte:head>

<section class="suite">
  <div class="suite-glow" aria-hidden="true"></div>
  <div class="suite-grid" aria-hidden="true"></div>

  <div class="suite-hero">
    <div class="suite-copy">
      <h1>{$t.product_h1}</h1>
      <p>{$t.product_lede}</p>
      <div class="suite-ctas">
        <a class="btn primary" href={demoHref()} onclick={() => track('product_live_demo', 'demo_start')}>
          {$t.product_cta_demo} <span aria-hidden="true">→</span>
        </a>
        <a class="link-quiet" href={registerHref()} onclick={() => track('product_register', 'register_start')}>
          {$t.product_cta_register}
        </a>
      </div>
    </div>

    <div class="stage" aria-label="Monti product suite preview">
      <!-- Decorative network nodes -->
      <span class="node n1" aria-hidden="true">◉</span>
      <span class="node n2" aria-hidden="true">▣</span>
      <span class="node n3" aria-hidden="true">↗</span>
      <span class="node n4" aria-hidden="true">◇</span>
      <span class="wire w1" aria-hidden="true"></span>
      <span class="wire w2" aria-hidden="true"></span>
      <span class="wire w3" aria-hidden="true"></span>
      <span class="wire w4" aria-hidden="true"></span>

      <aside class="panel nav-panel">
        <div class="panel-kicker">{$t.product_suite_kicker}</div>
        <nav aria-label="Product capabilities">
          {#each productNav as item}
            <a
              class:active={item.active}
              href={item.href}
              onclick={() => track(`product_nav_${item.label.toLowerCase().replaceAll(/[\s&]+/g, '_')}`)}
            >
              <span class="nav-ico" aria-hidden="true">{item.icon}</span>
              <span>{item.label}</span>
            </a>
          {/each}
        </nav>
      </aside>

      <article class="panel knowledge-panel" id="knowledge-hub">
        <div class="kh-head">
          <span class="kh-ico" aria-hidden="true">▣</span>
          <div>
            <strong>{$t.product_km_title}</strong>
            <small>{$t.product_km_sub}</small>
          </div>
        </div>
        <div class="kh-search" aria-hidden="true">
          <span>⌕</span>
          {$t.product_km_search}
        </div>
        <ul class="kh-list">
          {#each knowledgeItems as item}
            <li>
              <span class="doc" aria-hidden="true">▤</span>
              <span>{item}</span>
              <span class="arrow" aria-hidden="true">›</span>
            </li>
          {/each}
        </ul>
      </article>

      <article class="panel voice-panel" id="ai-voice-agents">
        <header class="voice-head">
          <span>{$t.product_voice_talking}</span>
          <span class="dots" aria-hidden="true">•••</span>
        </header>
        <div class="avatar-wrap">
          <span class="ring r1" aria-hidden="true"></span>
          <span class="ring r2" aria-hidden="true"></span>
          <img src="{base}/images/ava.jpg" alt="Ava, Monti AI voice agent" width="120" height="120" />
        </div>
        <strong class="voice-name">Ava</strong>
        <span class="voice-role">{$t.product_voice_role}</span>
        <span class="voice-time">00:01:24</span>
        <div class="voice-wave" aria-hidden="true">
          {#each Array(28) as _, i}
            <i style="--h:{(i * 5) % 16 + 5}px"></i>
          {/each}
        </div>
        <button class="end-call" type="button" onclick={() => track('product_end_call')}>{$t.product_voice_end}</button>
      </article>
    </div>
  </div>

  <div class="pitch">
    <h2>{$t.product_pitch_h2} <em>{$t.product_pitch_em}</em></h2>
    <p>{$t.product_pitch_p}</p>
  </div>

  <div class="benefit-grid">
    {#each benefits as item}
      <article class="benefit">
        <span class="benefit-ico" aria-hidden="true">{item.icon}</span>
        <h3>{item.title}</h3>
        <p>{item.body}</p>
      </article>
    {/each}
  </div>

  <div class="footer-cta">
    <div>
      <p class="eyebrow">{$t.product_footer_eyebrow}</p>
      <h2>{$t.product_footer_h2}</h2>
    </div>
    <div class="suite-ctas">
      <a class="btn primary" href={demoHref()} onclick={() => track('product_footer_demo', 'demo_start')}>
        {$t.product_footer_demo}
      </a>
      <a class="btn ghost" href="{base}/contact?kind=book_demo" onclick={() => track('product_footer_contact')}>
        {$t.product_footer_talk}
      </a>
    </div>
  </div>
</section>

<style>
  .suite {
    position: relative;
    overflow: hidden;
    padding: 48px 24px 72px;
    background:
      radial-gradient(circle at 72% 22%, rgb(20 90 220 / 16%), transparent 28%),
      radial-gradient(circle at 28% 48%, rgb(15 50 140 / 14%), transparent 32%),
      linear-gradient(180deg, #020713 0%, #031025 50%, #020713 100%);
  }

  .suite-glow {
    position: absolute;
    inset: 8% 10% auto;
    height: 420px;
    background: radial-gradient(ellipse at center, rgb(30 100 255 / 12%), transparent 70%);
    pointer-events: none;
  }

  .suite-grid {
    position: absolute;
    inset: 0;
    opacity: 0.16;
    background-image:
      linear-gradient(rgb(30 90 180 / 18%) 1px, transparent 1px),
      linear-gradient(90deg, rgb(30 90 180 / 18%) 1px, transparent 1px);
    background-size: 52px 52px;
    mask-image: linear-gradient(180deg, #000 0%, transparent 72%);
    pointer-events: none;
  }

  .suite-hero {
    position: relative;
    z-index: 1;
    width: min(1180px, 100%);
    margin: 0 auto;
    display: grid;
    grid-template-columns: minmax(240px, 0.9fr) minmax(0, 1.4fr);
    gap: 28px;
    align-items: start;
  }

  .suite-copy h1 {
    margin: 0 0 14px;
    font-size: clamp(2rem, 3.6vw, 2.75rem);
    letter-spacing: -0.03em;
    line-height: 1.1;
    font-weight: 750;
  }

  .suite-copy > p {
    margin: 0 0 20px;
    color: #93a6c4;
    font-size: 1.05rem;
    line-height: 1.55;
  }

  .suite-ctas {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 12px;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    border-radius: 10px;
    padding: 11px 18px;
    font-weight: 650;
    font-size: 14px;
    text-decoration: none;
    border: 1px solid transparent;
    transition: transform 0.15s ease, box-shadow 0.15s ease;
  }

  .btn:hover {
    transform: translateY(-1px);
  }

  .btn.primary {
    background: linear-gradient(100deg, #1f7bff, #2f6dff);
    color: #fff;
    box-shadow: 0 10px 28px rgb(31 123 255 / 30%);
  }

  .btn.ghost {
    background: rgb(10 20 42 / 70%);
    border-color: rgb(80 130 200 / 30%);
    color: #e8eefc;
  }

  .link-quiet {
    color: #9ec0ef;
    font-size: 14px;
    font-weight: 600;
    text-decoration: none;
  }

  .link-quiet:hover {
    color: #fff;
  }

  .stage {
    position: relative;
    min-height: 420px;
    display: grid;
    grid-template-columns: 200px minmax(180px, 1fr) 210px;
    gap: 14px;
    align-items: center;
    padding: 8px 0 8px 8px;
  }

  .panel {
    position: relative;
    z-index: 2;
    border-radius: 16px;
    border: 1px solid rgb(70 130 220 / 28%);
    background: linear-gradient(165deg, rgb(10 22 48 / 90%), rgb(5 12 28 / 94%));
    box-shadow:
      0 20px 50px rgb(0 12 40 / 40%),
      inset 0 1px 0 rgb(255 255 255 / 4%);
    backdrop-filter: blur(14px);
  }

  .nav-panel {
    padding: 14px 12px;
    align-self: stretch;
  }

  .panel-kicker {
    font-size: 10px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: #6f8fc4;
    font-weight: 700;
    margin: 0 0 10px 6px;
  }

  .nav-panel nav {
    display: grid;
    gap: 4px;
  }

  .nav-panel a {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 9px 10px;
    border-radius: 10px;
    color: #b4c4dc;
    text-decoration: none;
    font-size: 13px;
    font-weight: 500;
    border: 1px solid transparent;
  }

  .nav-panel a:hover {
    background: rgb(20 50 110 / 25%);
    color: #fff;
  }

  .nav-panel a.active {
    color: #fff;
    background: linear-gradient(100deg, rgb(25 90 220 / 55%), rgb(20 60 160 / 40%));
    border-color: rgb(70 140 255 / 35%);
    box-shadow: 0 0 0 1px rgb(40 110 255 / 12%);
  }

  .nav-ico {
    width: 22px;
    height: 22px;
    display: grid;
    place-items: center;
    font-size: 12px;
    opacity: 0.9;
  }

  .knowledge-panel {
    padding: 16px;
    min-height: 250px;
  }

  .kh-head {
    display: flex;
    gap: 10px;
    align-items: center;
    margin-bottom: 14px;
  }

  .kh-ico {
    width: 34px;
    height: 34px;
    border-radius: 10px;
    display: grid;
    place-items: center;
    background: rgb(25 70 160 / 45%);
    border: 1px solid rgb(70 140 240 / 30%);
    color: #7eb6ff;
  }

  .kh-head strong {
    display: block;
    font-size: 14px;
  }

  .kh-head small {
    display: block;
    color: #7f93b3;
    font-size: 11px;
    margin-top: 2px;
  }

  .kh-search {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 12px;
    border-radius: 10px;
    border: 1px solid rgb(70 120 190 / 28%);
    background: rgb(4 10 24 / 70%);
    color: #7f93b3;
    font-size: 13px;
    margin-bottom: 12px;
  }

  .kh-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 8px;
  }

  .kh-list li {
    display: grid;
    grid-template-columns: auto 1fr auto;
    gap: 10px;
    align-items: center;
    padding: 10px 12px;
    border-radius: 10px;
    border: 1px solid rgb(60 110 190 / 18%);
    background: rgb(8 18 40 / 55%);
    color: #d3e0f4;
    font-size: 13px;
  }

  .kh-list .doc {
    color: #6ea8ff;
  }

  .kh-list .arrow {
    color: #6f84a8;
  }

  .voice-panel {
    padding: 14px 14px 16px;
    text-align: center;
    display: grid;
    justify-items: center;
    gap: 4px;
  }

  .voice-head {
    width: 100%;
    display: flex;
    justify-content: space-between;
    align-items: center;
    font-size: 12px;
    color: #a8bbd8;
    margin-bottom: 8px;
  }

  .dots {
    letter-spacing: 1px;
    opacity: 0.6;
  }

  .avatar-wrap {
    position: relative;
    width: 128px;
    height: 128px;
    display: grid;
    place-items: center;
    margin: 4px 0 8px;
  }

  .avatar-wrap img {
    width: 96px;
    height: 96px;
    border-radius: 50%;
    object-fit: cover;
    object-position: top center;
    border: 2px solid rgb(80 160 255 / 50%);
    position: relative;
    z-index: 2;
    box-shadow: 0 0 0 6px rgb(20 60 140 / 25%), 0 12px 30px rgb(0 40 120 / 40%);
  }

  .ring {
    position: absolute;
    border-radius: 50%;
    border: 1px solid rgb(50 140 255 / 30%);
  }

  .r1 {
    inset: 0;
  }

  .r2 {
    inset: 10px;
    border-color: rgb(70 160 255 / 35%);
    animation: pulse 3.5s ease-in-out infinite;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 0.5;
      transform: scale(1);
    }
    50% {
      opacity: 1;
      transform: scale(1.04);
    }
  }

  .voice-name {
    font-size: 15px;
  }

  .voice-role {
    font-size: 11px;
    color: #7f93b3;
  }

  .voice-time {
    font-size: 12px;
    color: #8eb6f0;
    font-variant-numeric: tabular-nums;
    margin-top: 2px;
  }

  .voice-wave {
    display: flex;
    align-items: flex-end;
    gap: 2px;
    height: 28px;
    margin: 10px 0 12px;
  }

  .voice-wave i {
    width: 3px;
    height: var(--h);
    border-radius: 2px;
    background: linear-gradient(180deg, #4db3ff, #1f6fff);
    opacity: 0.9;
  }

  .end-call {
    border: 0;
    border-radius: 9px;
    background: linear-gradient(100deg, #e24b5a, #c83248);
    color: #fff;
    font-weight: 700;
    font-size: 13px;
    padding: 9px 28px;
    cursor: pointer;
    box-shadow: 0 8px 20px rgb(200 50 70 / 28%);
  }

  .node {
    position: absolute;
    z-index: 1;
    width: 36px;
    height: 36px;
    border-radius: 12px;
    display: grid;
    place-items: center;
    font-size: 13px;
    color: #7eb6ff;
    background: rgb(10 28 60 / 80%);
    border: 1px solid rgb(60 130 230 / 35%);
    box-shadow: 0 0 20px rgb(30 100 255 / 18%);
  }

  .n1 {
    top: 8%;
    right: 38%;
  }
  .n2 {
    top: 18%;
    right: 6%;
  }
  .n3 {
    bottom: 28%;
    right: 4%;
  }
  .n4 {
    bottom: 12%;
    right: 34%;
  }

  .wire {
    position: absolute;
    z-index: 0;
    height: 1px;
    background: linear-gradient(90deg, transparent, rgb(60 140 255 / 35%), transparent);
    opacity: 0.7;
  }

  .w1 {
    top: 18%;
    right: 18%;
    width: 22%;
    transform: rotate(-18deg);
  }
  .w2 {
    top: 42%;
    right: 12%;
    width: 18%;
    transform: rotate(28deg);
  }
  .w3 {
    bottom: 30%;
    right: 16%;
    width: 20%;
    transform: rotate(-12deg);
  }
  .w4 {
    bottom: 18%;
    right: 28%;
    width: 16%;
    transform: rotate(20deg);
  }

  .pitch {
    position: relative;
    z-index: 1;
    width: min(720px, 100%);
    margin: 56px auto 28px;
    text-align: center;
  }

  .pitch h2 {
    margin: 0 0 12px;
    font-size: clamp(1.45rem, 2.4vw, 1.85rem);
    letter-spacing: -0.02em;
  }

  .pitch h2 em {
    font-style: normal;
    color: #4ea3ff;
  }

  .pitch p {
    margin: 0;
    color: #93a6c4;
    line-height: 1.6;
    font-size: 15px;
  }

  .benefit-grid {
    position: relative;
    z-index: 1;
    width: min(1100px, 100%);
    margin: 0 auto;
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 14px;
  }

  .benefit {
    padding: 20px 16px;
    border-radius: 14px;
    border: 1px solid rgb(70 120 190 / 20%);
    background: linear-gradient(165deg, rgb(10 22 48 / 75%), rgb(5 12 28 / 85%));
    text-align: left;
  }

  .benefit-ico {
    display: grid;
    place-items: center;
    width: 40px;
    height: 40px;
    border-radius: 12px;
    margin-bottom: 12px;
    background: rgb(20 60 140 / 35%);
    border: 1px solid rgb(70 140 240 / 28%);
    font-size: 18px;
  }

  .benefit h3 {
    margin: 0 0 8px;
    font-size: 15px;
  }

  .benefit p {
    margin: 0;
    color: #8ea0bd;
    font-size: 13px;
    line-height: 1.5;
  }

  .footer-cta {
    position: relative;
    z-index: 1;
    width: min(1100px, 100%);
    margin: 40px auto 0;
    padding: 22px 24px;
    border-radius: 16px;
    border: 1px solid rgb(70 120 190 / 22%);
    background: linear-gradient(145deg, rgb(12 24 50 / 80%), rgb(6 12 28 / 90%));
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 20px;
    flex-wrap: wrap;
  }

  .eyebrow {
    margin: 0 0 6px;
    font-size: 11px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: #5f8fd4;
    font-weight: 700;
  }

  .footer-cta h2 {
    margin: 0;
    font-size: 1.35rem;
  }

  @media (max-width: 1020px) {
    .suite-hero {
      grid-template-columns: 1fr;
    }

    .stage {
      grid-template-columns: 1fr 1fr;
      min-height: 0;
    }

    .nav-panel {
      grid-column: 1 / -1;
    }

    .node,
    .wire {
      display: none;
    }

    .benefit-grid {
      grid-template-columns: 1fr 1fr;
    }
  }

  @media (max-width: 680px) {
    .suite {
      padding: 28px 16px 48px;
    }

    .stage {
      grid-template-columns: 1fr;
    }

    .benefit-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
