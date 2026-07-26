<script lang="ts">
  import { base } from '$app/paths';
  import { demoHref, registerHref } from '$lib/attribution';
  import { postFunnelEvent } from '$lib/api';
  import { page } from '$app/stores';
  import { t } from '$lib/i18n';

  function track(ctaId: string, event: 'cta_click' | 'demo_start' | 'register_start' = 'cta_click') {
    void postFunnelEvent({
      event_name: event,
      page_path: $page.url.pathname,
      cta_id: ctaId
    }).catch(() => {});
  }

  const featureChips = $derived([
    { icon: '☎', label: $t.home_chip_voice },
    { icon: '🌐', label: $t.home_chip_lang },
    { icon: '📘', label: $t.home_chip_km },
    { icon: '🛡', label: $t.home_chip_secure }
  ]);

  const stats = $derived([
    { value: '60%', label: $t.home_stat_60 },
    { value: '1000s', label: $t.home_stat_1000s },
    { value: '24/7', label: $t.home_stat_247 },
    { value: 'Higher CSAT', label: $t.home_stat_csat }
  ]);

  const capabilities = $derived([
    { title: $t.home_cap_voice_t, body: $t.home_cap_voice_b },
    { title: $t.home_cap_km_t, body: $t.home_cap_km_b },
    { title: $t.home_cap_omni_t, body: $t.home_cap_omni_b },
    { title: $t.home_cap_workflow_t, body: $t.home_cap_workflow_b },
    { title: $t.home_cap_handover_t, body: $t.home_cap_handover_b },
    { title: $t.home_cap_insights_t, body: $t.home_cap_insights_b }
  ]);

  const useCases = $derived([
    $t.home_use_1,
    $t.home_use_2,
    $t.home_use_3,
    $t.home_use_4,
    $t.home_use_5,
    $t.home_use_6
  ]);
</script>

<svelte:head>
  <title>{$t.home_title}</title>
</svelte:head>

<section class="hero-bleed">
  <div class="hero-bg" aria-hidden="true">
    <img src="{base}/images/home-hero-ref.png" alt="" class="hero-bg-img" />
    <div class="hero-bg-fade"></div>
  </div>

  <div class="hero-inner">
    <div class="hero-left">
      <span class="pill">{$t.home_pill}</span>
      <h1>
        {$t.home_h1_1}<br />
        {$t.home_h1_2}<br />
        {$t.home_h1_3}
      </h1>
      <p class="lede">{$t.home_lede}</p>
      <div class="hero-ctas">
        <a
          class="btn primary"
          href={demoHref()}
          onclick={() => track('hero_try_demo', 'demo_start')}
        >
          {$t.home_cta_demo}
          <span class="arrow">→</span>
        </a>
        <a class="btn outline" href="{base}/demo" onclick={() => track('hero_watch_video')}>
          <span class="play">▶</span>
          {$t.home_cta_video}
        </a>
      </div>
      <ul class="chips">
        {#each featureChips as chip}
          <li>
            <span class="chip-icon" aria-hidden="true">{chip.icon}</span>
            {chip.label}
          </li>
        {/each}
      </ul>
    </div>

    <div class="hero-center">
      <div class="avatar-stage">
        <div class="ring ring-outer"></div>
        <div class="ring ring-mid"></div>
        <div class="wave" aria-hidden="true"></div>
        <img
          class="avatar-photo"
          src="{base}/images/ava.jpg"
          alt="Ava, Monti AI support agent"
          width="280"
          height="280"
        />
        <div class="avatar-badge">
          <strong>Ava</strong>
          <span>{$t.home_ava_role}</span>
          <small>{$t.home_ava_tone}</small>
        </div>
      </div>
    </div>

    <aside class="caller-desk card-glass" aria-label="Monti Caller Desk preview">
      <header class="desk-head">
        <div>
          <strong>{$t.home_desk_title}</strong>
          <span class="live"><i></i> {$t.home_desk_live}</span>
        </div>
        <div class="tabs">
          <span class="tab active">{$t.home_desk_general}</span>
          <span class="tab">{$t.home_desk_billing}</span>
          <span class="tab">{$t.home_desk_tech}</span>
        </div>
      </header>
      <div class="desk-body">
        <div class="msg agent">
          <img src="{base}/images/ava.jpg" alt="" width="28" height="28" />
          <div>
            <strong>Ava</strong>
            <p>{$t.home_desk_welcome}</p>
            <div class="waveform" aria-hidden="true">
              {#each Array(18) as _, i}
                <i style="--h:{(i % 5) + 2}"></i>
              {/each}
            </div>
          </div>
        </div>
        <div class="msg user">
          <div>
            <strong>{$t.home_desk_you}</strong>
            <p>{$t.home_desk_user_msg}</p>
            <time>{$t.home_desk_just_now}</time>
          </div>
        </div>
      </div>
      <footer class="desk-foot">
        <input type="text" placeholder={$t.home_desk_placeholder} disabled aria-disabled="true" />
        <button type="button" class="send" disabled>{$t.home_desk_send}</button>
      </footer>
    </aside>
  </div>

  <div class="stats-bar">
    {#each stats as s}
      <div class="stat">
        <strong>{s.value}</strong>
        <span>{s.label}</span>
      </div>
    {/each}
  </div>
</section>

<section class="section lower">
  <div class="lower-grid">
    <div>
      <p class="eyebrow">{$t.home_built_eyebrow}</p>
      <h2>{$t.home_built_h2}</h2>
      <div class="cap-grid">
        {#each capabilities as item}
          <article class="cap">
            <h3>{item.title}</h3>
            <p>{item.body}</p>
          </article>
        {/each}
      </div>
    </div>

    <div>
      <p class="eyebrow">{$t.home_use_eyebrow}</p>
      <h2>{$t.home_use_h2}</h2>
      <div class="use-grid">
        {#each useCases as label}
          <a class="use-card" href="{base}/solutions" onclick={() => track('home_use_case')}>{label}</a>
        {/each}
      </div>
    </div>

    <aside class="qr-card">
      <div class="qr-copy">
        <h3>{$t.home_qr_h3}</h3>
        <p>{$t.home_qr_p}</p>
        <a
          class="btn primary sm"
          href={demoHref()}
          onclick={() => track('home_qr_demo', 'demo_start')}>{$t.home_qr_cta}</a
        >
      </div>
      <a class="qr-box" href={demoHref()} aria-label="Open live demo" onclick={() => track('home_qr', 'demo_start')}>
        <svg viewBox="0 0 100 100" width="120" height="120" role="img" aria-label="Demo QR placeholder">
          <rect width="100" height="100" fill="#fff" rx="4" />
          <g fill="#0a1628">
            <rect x="8" y="8" width="28" height="28" rx="2" />
            <rect x="14" y="14" width="16" height="16" fill="#fff" />
            <rect x="18" y="18" width="8" height="8" />
            <rect x="64" y="8" width="28" height="28" rx="2" />
            <rect x="70" y="14" width="16" height="16" fill="#fff" />
            <rect x="74" y="18" width="8" height="8" />
            <rect x="8" y="64" width="28" height="28" rx="2" />
            <rect x="14" y="70" width="16" height="16" fill="#fff" />
            <rect x="18" y="74" width="8" height="8" />
            <rect x="44" y="8" width="6" height="6" />
            <rect x="52" y="8" width="6" height="6" />
            <rect x="44" y="16" width="6" height="6" />
            <rect x="52" y="24" width="6" height="6" />
            <rect x="44" y="44" width="6" height="6" />
            <rect x="52" y="44" width="6" height="6" />
            <rect x="60" y="44" width="6" height="6" />
            <rect x="68" y="44" width="6" height="6" />
            <rect x="76" y="44" width="6" height="6" />
            <rect x="84" y="44" width="6" height="6" />
            <rect x="44" y="52" width="6" height="6" />
            <rect x="60" y="52" width="6" height="6" />
            <rect x="76" y="52" width="6" height="6" />
            <rect x="44" y="60" width="6" height="6" />
            <rect x="52" y="60" width="6" height="6" />
            <rect x="68" y="60" width="6" height="6" />
            <rect x="84" y="60" width="6" height="6" />
            <rect x="44" y="68" width="6" height="6" />
            <rect x="60" y="68" width="6" height="6" />
            <rect x="76" y="76" width="6" height="6" />
            <rect x="84" y="68" width="6" height="6" />
            <rect x="68" y="84" width="6" height="6" />
            <rect x="84" y="84" width="6" height="6" />
            <rect x="52" y="84" width="6" height="6" />
            <rect x="44" y="84" width="6" height="6" />
          </g>
        </svg>
      </a>
    </aside>
  </div>
</section>

<section class="section bottom-cta">
  <div class="cta-band">
    <div>
      <h2>{$t.home_ready_h2}</h2>
      <p class="muted">{$t.home_ready_p}</p>
    </div>
    <div class="hero-ctas">
      <a class="btn primary" href={demoHref()} onclick={() => track('bottom_demo', 'demo_start')}>{$t.home_ready_demo}</a>
      <a class="btn outline" href="{base}/contact?kind=book_demo">{$t.home_ready_book}</a>
      <a
        class="btn outline"
        href={registerHref()}
        onclick={() => track('bottom_register', 'register_start')}>{$t.home_ready_register}</a
      >
    </div>
  </div>
</section>

<style>
  .hero-bleed {
    position: relative;
    overflow: hidden;
    border-bottom: 1px solid rgb(60 110 200 / 16%);
    background: radial-gradient(ellipse 80% 60% at 70% 40%, rgb(20 70 180 / 22%), transparent 55%),
      radial-gradient(circle at 15% 20%, rgb(22 120 255 / 12%), transparent 40%), #02060f;
  }

  .hero-bg {
    position: absolute;
    inset: 0;
    pointer-events: none;
  }

  .hero-bg-img {
    position: absolute;
    right: -8%;
    top: -6%;
    width: min(920px, 78vw);
    height: auto;
    opacity: 0.22;
    filter: saturate(1.1) blur(0.2px);
    mask-image: linear-gradient(90deg, transparent 0%, #000 28%, #000 78%, transparent 100%);
  }

  .hero-bg-fade {
    position: absolute;
    inset: 0;
    background: linear-gradient(100deg, #02060f 18%, rgb(2 6 15 / 72%) 48%, rgb(2 6 15 / 55%) 100%);
  }

  .hero-inner {
    position: relative;
    z-index: 1;
    width: min(1200px, 100%);
    margin: 0 auto;
    padding: 48px 24px 28px;
    display: grid;
    grid-template-columns: 1.15fr 0.95fr 0.95fr;
    gap: 20px;
    align-items: center;
  }

  .pill {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    border: 1px solid rgb(55 140 255 / 35%);
    border-radius: 999px;
    padding: 6px 12px;
    font-size: 12px;
    font-weight: 600;
    color: #9ec8ff;
    background: rgb(18 48 100 / 35%);
    margin-bottom: 18px;
  }

  .hero-left h1 {
    margin: 0 0 16px;
    font-size: clamp(2.1rem, 4.2vw, 3.15rem);
    line-height: 1.08;
    letter-spacing: -0.035em;
    font-weight: 750;
  }

  .hero-left h1 em {
    font-style: normal;
    background: linear-gradient(90deg, #3b9bff, #1f6fff);
    -webkit-background-clip: text;
    background-clip: text;
    color: transparent;
  }

  .lede {
    margin: 0 0 22px;
    color: #93a4c0;
    font-size: 1.02rem;
    line-height: 1.65;
    max-width: 40ch;
  }

  .hero-ctas {
    display: flex;
    flex-wrap: wrap;
    gap: 12px;
    margin-bottom: 22px;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    border-radius: 10px;
    padding: 12px 18px;
    font-weight: 650;
    font-size: 14px;
    text-decoration: none;
    border: 1px solid transparent;
    transition:
      transform 0.15s ease,
      box-shadow 0.15s ease;
  }

  .btn:hover {
    transform: translateY(-1px);
  }

  .btn.primary {
    background: linear-gradient(100deg, #1f7bff, #2f6dff);
    color: #fff;
    box-shadow: 0 10px 30px rgb(31 123 255 / 32%);
  }

  .btn.outline {
    background: rgb(8 16 32 / 55%);
    border-color: rgb(90 130 190 / 35%);
    color: #e8eefc;
  }

  .btn.sm {
    padding: 9px 14px;
    font-size: 13px;
  }

  .arrow,
  .play {
    font-size: 12px;
    opacity: 0.9;
  }

  .chips {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 10px 14px;
  }

  .chips li {
    display: inline-flex;
    align-items: center;
    gap: 7px;
    color: #a9b8d0;
    font-size: 12px;
    font-weight: 500;
  }

  .chip-icon {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: rgb(20 50 110 / 45%);
    border: 1px solid rgb(70 130 220 / 28%);
    font-size: 12px;
  }

  .hero-center {
    display: grid;
    place-items: center;
    min-height: 360px;
  }

  .avatar-stage {
    position: relative;
    width: min(320px, 100%);
    aspect-ratio: 1;
    display: grid;
    place-items: center;
  }

  .ring {
    position: absolute;
    border-radius: 50%;
    border: 1px solid rgb(50 140 255 / 28%);
    box-shadow:
      0 0 40px rgb(30 110 255 / 18%),
      inset 0 0 40px rgb(20 90 220 / 10%);
  }

  .ring-outer {
    inset: 0;
  }

  .ring-mid {
    inset: 10%;
    border-color: rgb(60 160 255 / 35%);
  }

  .wave {
    position: absolute;
    inset: 18%;
    border-radius: 50%;
    background:
      radial-gradient(circle at 50% 55%, transparent 42%, rgb(30 100 255 / 8%) 43%, transparent 60%),
      repeating-radial-gradient(circle at 50% 50%, transparent 0 8px, rgb(40 130 255 / 6%) 8px 9px);
    animation: pulse 4.5s ease-in-out infinite;
  }

  @keyframes pulse {
    0%,
    100% {
      opacity: 0.55;
      transform: scale(1);
    }
    50% {
      opacity: 1;
      transform: scale(1.03);
    }
  }

  .avatar-photo {
    position: relative;
    z-index: 2;
    width: 62%;
    height: 62%;
    object-fit: cover;
    object-position: top center;
    border-radius: 50%;
    border: 3px solid rgb(80 160 255 / 45%);
    box-shadow:
      0 0 0 8px rgb(20 60 140 / 25%),
      0 18px 50px rgb(0 40 120 / 45%);
  }

  .avatar-badge {
    position: absolute;
    z-index: 3;
    bottom: 6%;
    left: 50%;
    transform: translateX(-50%);
    min-width: 150px;
    text-align: center;
    padding: 10px 14px;
    border-radius: 12px;
    background: linear-gradient(160deg, rgb(12 28 58 / 92%), rgb(6 14 32 / 94%));
    border: 1px solid rgb(70 140 230 / 35%);
    box-shadow: 0 12px 30px rgb(0 0 0 / 35%);
  }

  .avatar-badge strong {
    display: block;
    font-size: 15px;
  }

  .avatar-badge span {
    display: block;
    font-size: 12px;
    color: #8eb6f0;
    margin-top: 2px;
  }

  .avatar-badge small {
    display: block;
    font-size: 11px;
    color: #6f84a8;
    margin-top: 2px;
  }

  .card-glass {
    border: 1px solid rgb(70 130 220 / 28%);
    border-radius: 16px;
    background: linear-gradient(165deg, rgb(10 22 48 / 88%), rgb(5 12 28 / 92%));
    box-shadow:
      0 24px 60px rgb(0 10 40 / 45%),
      inset 0 1px 0 rgb(255 255 255 / 4%);
    backdrop-filter: blur(16px);
    overflow: hidden;
  }

  .caller-desk {
    padding: 0;
    max-width: 340px;
    justify-self: end;
  }

  .desk-head {
    padding: 14px 14px 10px;
    border-bottom: 1px solid rgb(70 120 200 / 18%);
  }

  .desk-head > div:first-child {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 10px;
  }

  .desk-head strong {
    font-size: 13px;
  }

  .live {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: #3dd68c;
    font-size: 11px;
    font-weight: 700;
  }

  .live i {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: #3dd68c;
    box-shadow: 0 0 8px #3dd68c;
  }

  .tabs {
    display: flex;
    gap: 6px;
  }

  .tab {
    font-size: 11px;
    padding: 4px 9px;
    border-radius: 7px;
    color: #8ea0bd;
    background: rgb(15 28 52 / 70%);
    border: 1px solid transparent;
  }

  .tab.active {
    color: #fff;
    background: linear-gradient(100deg, #1f7bff, #2a62e8);
  }

  .desk-body {
    padding: 14px;
    display: grid;
    gap: 12px;
    min-height: 210px;
  }

  .msg {
    display: flex;
    gap: 10px;
    font-size: 12px;
  }

  .msg img {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    object-fit: cover;
    flex-shrink: 0;
  }

  .msg strong {
    display: block;
    font-size: 11px;
    margin-bottom: 4px;
    color: #b7c8e4;
  }

  .msg p {
    margin: 0;
    line-height: 1.45;
    color: #d5e0f4;
  }

  .msg.agent > div {
    background: rgb(18 40 80 / 55%);
    border: 1px solid rgb(60 120 220 / 22%);
    border-radius: 12px;
    padding: 10px 12px;
  }

  .msg.user {
    justify-content: flex-end;
  }

  .msg.user > div {
    background: rgb(12 30 58 / 80%);
    border: 1px solid rgb(50 100 180 / 25%);
    border-radius: 12px;
    padding: 10px 12px;
    max-width: 85%;
  }

  .msg time {
    display: block;
    margin-top: 6px;
    font-size: 10px;
    color: #6a7d9c;
    text-align: right;
  }

  .waveform {
    display: flex;
    align-items: flex-end;
    gap: 2px;
    height: 22px;
    margin-top: 10px;
  }

  .waveform i {
    width: 3px;
    height: calc(var(--h) * 3px);
    border-radius: 2px;
    background: linear-gradient(180deg, #4db3ff, #1f6fff);
    opacity: 0.85;
  }

  .desk-foot {
    display: flex;
    gap: 8px;
    padding: 0 14px 14px;
  }

  .desk-foot input {
    flex: 1;
    border: 1px solid rgb(70 120 190 / 28%);
    border-radius: 9px;
    background: rgb(4 10 22 / 70%);
    color: var(--ink);
    padding: 9px 11px;
    font-size: 12px;
  }

  .desk-foot .send {
    border: 0;
    border-radius: 9px;
    background: #1f7bff;
    color: #fff;
    font-weight: 700;
    font-size: 12px;
    padding: 0 14px;
    opacity: 0.75;
  }

  .stats-bar {
    position: relative;
    z-index: 1;
    width: min(1200px, 100%);
    margin: 8px auto 0;
    padding: 18px 24px 28px;
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 12px;
  }

  .stat {
    display: grid;
    gap: 4px;
    padding: 14px 12px;
    border-radius: 12px;
    border: 1px solid rgb(60 120 210 / 18%);
    background: rgb(8 18 40 / 45%);
    text-align: center;
  }

  .stat strong {
    font-size: clamp(1.15rem, 2vw, 1.45rem);
    color: #4ea3ff;
    letter-spacing: -0.02em;
  }

  .stat span {
    font-size: 12px;
    color: #8ea0bd;
    line-height: 1.35;
  }

  .lower {
    padding-top: 48px;
  }

  .lower-grid {
    display: grid;
    grid-template-columns: 1.2fr 1fr 0.85fr;
    gap: 28px;
    align-items: start;
  }

  .eyebrow {
    margin: 0 0 8px;
    font-size: 11px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: #5f8fd4;
    font-weight: 700;
  }

  .eyebrow.center {
    text-align: center;
  }

  .lower h2 {
    margin: 0 0 18px;
    font-size: clamp(1.25rem, 2vw, 1.55rem);
    letter-spacing: -0.02em;
    line-height: 1.25;
  }

  .cap-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }

  .cap {
    padding: 14px;
    border-radius: 12px;
    border: 1px solid rgb(70 120 190 / 18%);
    background: rgb(8 16 34 / 55%);
  }

  .cap h3 {
    margin: 0 0 6px;
    font-size: 13px;
  }

  .cap p {
    margin: 0;
    font-size: 12px;
    color: #8ea0bd;
    line-height: 1.45;
  }

  .use-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 10px;
  }

  .use-card {
    display: grid;
    place-items: center;
    text-align: center;
    min-height: 64px;
    padding: 12px;
    border-radius: 12px;
    border: 1px solid rgb(70 120 190 / 18%);
    background: rgb(8 16 34 / 55%);
    color: #d5e0f4;
    text-decoration: none;
    font-size: 13px;
    font-weight: 600;
    transition: border-color 0.15s ease, background 0.15s ease;
  }

  .use-card:hover {
    border-color: rgb(70 150 255 / 45%);
    background: rgb(16 40 90 / 45%);
  }

  .qr-card {
    border-radius: 16px;
    border: 1px solid rgb(80 130 255 / 40%);
    background: linear-gradient(160deg, rgb(18 40 100 / 55%), rgb(8 16 40 / 80%));
    padding: 18px;
    display: grid;
    gap: 14px;
    justify-items: center;
    text-align: center;
    box-shadow: 0 16px 40px rgb(20 60 160 / 18%);
  }

  .qr-copy h3 {
    margin: 0 0 8px;
    font-size: 1.1rem;
  }

  .qr-copy p {
    margin: 0 0 14px;
    color: #9aacc8;
    font-size: 13px;
    line-height: 1.45;
  }

  .qr-box {
    display: grid;
    place-items: center;
    padding: 10px;
    border-radius: 12px;
    background: #fff;
    box-shadow: 0 8px 24px rgb(0 0 0 / 25%);
  }

  .bottom-cta {
    padding-top: 8px;
    padding-bottom: 48px;
  }

  .cta-band {
    display: flex;
    justify-content: space-between;
    align-items: center;
    gap: 24px;
    flex-wrap: wrap;
    padding: 24px;
    border-radius: 16px;
    border: 1px solid rgb(70 120 190 / 22%);
    background: linear-gradient(145deg, rgb(12 24 50 / 80%), rgb(6 12 28 / 90%));
  }

  .cta-band h2 {
    margin: 0 0 8px;
    font-size: 1.4rem;
  }

  .cta-band p {
    margin: 0;
    max-width: 48ch;
  }

  .cta-band .hero-ctas {
    margin: 0;
  }

  @media (max-width: 1100px) {
    .hero-inner {
      grid-template-columns: 1fr 1fr;
    }

    .caller-desk {
      grid-column: 1 / -1;
      justify-self: stretch;
      max-width: none;
    }

    .lower-grid {
      grid-template-columns: 1fr 1fr;
    }

    .qr-card {
      grid-column: 1 / -1;
      grid-template-columns: 1fr auto;
      text-align: left;
      justify-items: start;
      align-items: center;
    }
  }

  @media (max-width: 780px) {
    .hero-inner {
      grid-template-columns: 1fr;
      padding-top: 28px;
    }

    .hero-bg-img {
      opacity: 0.12;
      right: -20%;
      width: 120%;
    }

    .stats-bar {
      grid-template-columns: 1fr 1fr;
    }

    .lower-grid {
      grid-template-columns: 1fr;
    }

    .qr-card {
      grid-template-columns: 1fr;
      justify-items: center;
      text-align: center;
    }
  }

  @media (max-width: 480px) {
    .stats-bar,
    .cap-grid,
    .use-grid {
      grid-template-columns: 1fr;
    }
  }
</style>
