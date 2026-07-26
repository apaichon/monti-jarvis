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

  const aiItems = $derived([
    { icon: '▢', label: $t.ent_arch_web },
    { icon: '🎙', label: $t.ent_arch_voice },
    { icon: '💬', label: $t.ent_arch_chat },
    { icon: '⚭', label: $t.ent_arch_orch },
    { icon: '🛡', label: $t.ent_arch_guard },
    { icon: '▤', label: $t.ent_arch_analytics }
  ]);

  const dataItems = $derived([
    { icon: '☎', label: $t.ent_arch_calls },
    { icon: '📚', label: $t.ent_arch_km },
    { icon: '👤', label: $t.ent_arch_customer },
    { icon: '🗄', label: $t.ent_arch_storage },
    { icon: '🗂', label: $t.ent_arch_db },
    { icon: '📋', label: $t.ent_arch_audit }
  ]);

  const how = $derived([
    { n: '1', icon: '🧠', t: $t.ent_how_1_t, b: $t.ent_how_1_b },
    { n: '2', icon: '🛡', t: $t.ent_how_2_t, b: $t.ent_how_2_b },
    { n: '3', icon: '🗄', t: $t.ent_how_3_t, b: $t.ent_how_3_b },
    { n: '4', icon: '☁', t: $t.ent_how_4_t, b: $t.ent_how_4_b }
  ]);

  const managed = $derived([
    $t.ent_managed_1,
    $t.ent_managed_2,
    $t.ent_managed_3,
    $t.ent_managed_4,
    $t.ent_managed_5
  ]);

  const controlled = $derived([
    $t.ent_controlled_1,
    $t.ent_controlled_2,
    $t.ent_controlled_3,
    $t.ent_controlled_4,
    $t.ent_controlled_5
  ]);

  const deploy = $derived([
    { icon: '🏢', t: $t.ent_deploy_1_t, b: $t.ent_deploy_1_b },
    { icon: '☁', t: $t.ent_deploy_2_t, b: $t.ent_deploy_2_b },
    { icon: '⇄', t: $t.ent_deploy_3_t, b: $t.ent_deploy_3_b }
  ]);

  const benefits = $derived([
    { icon: '🛡', t: $t.ent_ben_1_t, b: $t.ent_ben_1_b },
    { icon: '✔', t: $t.ent_ben_2_t, b: $t.ent_ben_2_b },
    { icon: '⚡', t: $t.ent_ben_3_t, b: $t.ent_ben_3_b },
    { icon: '◎', t: $t.ent_ben_4_t, b: $t.ent_ben_4_b },
    { icon: '⧉', t: $t.ent_ben_5_t, b: $t.ent_ben_5_b },
    { icon: '👁', t: $t.ent_ben_6_t, b: $t.ent_ben_6_b }
  ]);

  const trusts = $derived([$t.ent_trust_1, $t.ent_trust_2, $t.ent_trust_3, $t.ent_trust_4]);
</script>

<svelte:head>
  <title>{$t.ent_title}</title>
  <meta name="description" content={$t.ent_lede} />
</svelte:head>

<section class="ent-page">
  <div class="glow" aria-hidden="true"></div>

  <!-- Hero -->
  <div class="hero">
    <div class="hero-copy">
      <span class="pill">{$t.ent_badge}</span>
      <h1>
        {$t.ent_h1_1}
        <span class="accent">{$t.ent_h1_2}</span>
        {$t.ent_h1_3}
      </h1>
      <p class="lede">{$t.ent_lede}</p>
      <div class="hero-ctas">
        <a
          class="btn primary"
          href="{base}/contact?kind=contact&use_case=Enterprise%20sales"
          onclick={() => track('ent_talk_sales')}
        >
          {$t.ent_cta_sales} <span aria-hidden="true">›</span>
        </a>
        <a class="btn ghost" href="#architecture" onclick={() => track('ent_see_arch')}>
          {$t.ent_cta_arch}
        </a>
      </div>
    </div>

    <div class="arch" id="architecture">
      <div class="arch-main">
        <div class="layer layer-ai">
          <div class="layer-head">{$t.ent_layer_ai}</div>
          <div class="layer-grid">
            {#each aiItems as item}
              <div class="chip">
                <span class="chip-ico" aria-hidden="true">{item.icon}</span>
                <span>{item.label}</span>
              </div>
            {/each}
          </div>
          <div class="layer-foot">{$t.ent_layer_ai_sub}</div>
        </div>

        <div class="connector" aria-hidden="true">
          <span class="conn-line"></span>
          <span class="conn-label">{$t.ent_secure_link}</span>
          <span class="conn-line"></span>
        </div>

        <div class="layer layer-data">
          <div class="layer-head">{$t.ent_layer_data}</div>
          <div class="layer-grid">
            {#each dataItems as item}
              <div class="chip">
                <span class="chip-ico" aria-hidden="true">{item.icon}</span>
                <span>{item.label}</span>
              </div>
            {/each}
          </div>
          <div class="layer-foot">{$t.ent_layer_data_sub}</div>
          <div class="env-badge">{$t.ent_layer_env}</div>
        </div>
      </div>

      <aside class="ownership" aria-label="Ownership">
        <div class="own-label">OWNERSHIP</div>
        <div class="own-card monti">
          <img src="{base}/images/monti-logo.png" alt="" width="36" height="36" />
          <span>{$t.ent_own_monti}</span>
        </div>
        <div class="own-divider" aria-hidden="true"></div>
        <div class="own-card you">
          <span class="you-ico" aria-hidden="true">👤</span>
          <span>{$t.ent_own_you}</span>
        </div>
      </aside>
    </div>
  </div>

  <!-- Trust bar -->
  <div class="trust-bar">
    <div class="trust-title">
      <span class="shield" aria-hidden="true">🛡</span>
      <strong>{$t.ent_trust_title}</strong>
    </div>
    <ul>
      {#each trusts as item}
        <li><span aria-hidden="true">✓</span> {item}</li>
      {/each}
    </ul>
  </div>

  <!-- How it works -->
  <div class="section">
    <h2 class="section-label">{$t.ent_how_h2}</h2>
    <div class="how-grid">
      {#each how as step}
        <article class="how-card">
          <span class="how-ico" aria-hidden="true">{step.icon}</span>
          <h3>{step.t}</h3>
          <p>{step.b}</p>
        </article>
      {/each}
    </div>
  </div>

  <!-- Managed / Controlled -->
  <div class="split">
    <div class="panel panel-monti">
      <div class="panel-head">
        <img src="{base}/images/monti-logo.png" alt="" width="28" height="28" />
        <h2>{$t.ent_managed_h2}</h2>
      </div>
      <div class="panel-items">
        {#each managed as label}
          <div class="panel-item">
            <span class="dot blue" aria-hidden="true"></span>
            {label}
          </div>
        {/each}
      </div>
      <p class="panel-foot">{$t.ent_managed_foot}</p>
    </div>
    <div class="panel panel-you">
      <div class="panel-head">
        <span class="you-badge" aria-hidden="true">👤</span>
        <h2>{$t.ent_controlled_h2}</h2>
      </div>
      <div class="panel-items">
        {#each controlled as label}
          <div class="panel-item">
            <span class="dot green" aria-hidden="true"></span>
            {label}
          </div>
        {/each}
      </div>
      <p class="panel-foot">{$t.ent_controlled_foot}</p>
    </div>
  </div>

  <!-- Deploy + benefits -->
  <div class="bottom-grid">
    <div class="section tight">
      <h2 class="section-label">{$t.ent_deploy_h2}</h2>
      <div class="deploy-grid">
        {#each deploy as d}
          <article class="mini-card">
            <span class="mini-ico" aria-hidden="true">{d.icon}</span>
            <h3>{d.t}</h3>
            <p>{d.b}</p>
          </article>
        {/each}
      </div>
    </div>
    <div class="section tight">
      <h2 class="section-label">{$t.ent_ben_h2}</h2>
      <div class="ben-grid">
        {#each benefits as b}
          <article class="ben-card">
            <span class="ben-ico" aria-hidden="true">{b.icon}</span>
            <div>
              <h3>{b.t}</h3>
              <p>{b.b}</p>
            </div>
          </article>
        {/each}
      </div>
    </div>
  </div>

  <!-- CTA -->
  <aside class="cta-banner">
    <div class="cta-copy">
      <h2>{$t.ent_cta_h2}</h2>
      <p>{$t.ent_cta_p}</p>
    </div>
    <div class="cta-actions">
      <a
        class="btn primary"
        href="{base}/contact?kind=book_demo&use_case=Enterprise%20demo"
        onclick={() => track('ent_book_demo')}
      >
        {$t.ent_cta_demo} <span aria-hidden="true">›</span>
      </a>
      <a
        class="btn ghost"
        href="{base}/contact?kind=contact&use_case=Talk%20to%20architect"
        onclick={() => track('ent_talk_architect')}
      >
        {$t.ent_cta_architect}
      </a>
    </div>
    <div class="cta-wave" aria-hidden="true"></div>
  </aside>
</section>

<style>
  .ent-page {
    position: relative;
    overflow: hidden;
    padding: 40px 24px 72px;
    background:
      radial-gradient(circle at 78% 6%, rgb(30 100 255 / 14%), transparent 26%),
      radial-gradient(circle at 12% 40%, rgb(20 80 180 / 10%), transparent 28%),
      linear-gradient(180deg, #020713 0%, #031025 50%, #020713 100%);
  }

  .glow {
    position: absolute;
    inset: 0;
    pointer-events: none;
    background-image:
      linear-gradient(rgb(30 90 180 / 9%) 1px, transparent 1px),
      linear-gradient(90deg, rgb(30 90 180 / 9%) 1px, transparent 1px);
    background-size: 52px 52px;
    mask-image: linear-gradient(180deg, #000 0%, transparent 72%);
    opacity: 0.4;
  }

  .hero {
    position: relative;
    z-index: 1;
    width: min(1180px, 100%);
    margin: 0 auto 22px;
    display: grid;
    grid-template-columns: minmax(0, 0.95fr) minmax(0, 1.15fr);
    gap: 28px;
    align-items: start;
  }

  .pill {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    padding: 6px 12px;
    border-radius: 999px;
    border: 1px solid rgb(70 140 255 / 35%);
    background: rgb(20 50 120 / 35%);
    color: #9ec5ff;
    font-size: 12px;
    font-weight: 650;
    margin-bottom: 16px;
  }

  .pill::before {
    content: '';
    width: 7px;
    height: 7px;
    border-radius: 50%;
    background: #4ea3ff;
    box-shadow: 0 0 10px #4ea3ff;
  }

  .hero-copy h1 {
    margin: 0 0 16px;
    font-size: clamp(2rem, 3.6vw, 2.9rem);
    line-height: 1.12;
    letter-spacing: -0.03em;
    font-weight: 750;
  }

  .accent {
    color: #3d9bff;
  }

  .lede {
    margin: 0 0 22px;
    color: #93a6c4;
    font-size: 1.02rem;
    line-height: 1.6;
    max-width: 42ch;
  }

  .hero-ctas,
  .cta-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
  }

  .btn {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    border-radius: 10px;
    padding: 11px 18px;
    font-size: 14px;
    font-weight: 700;
    text-decoration: none;
    border: 1px solid transparent;
    transition: filter 0.15s ease, border-color 0.15s ease;
  }

  .btn.primary {
    background: linear-gradient(100deg, #1f7bff, #2f6dff);
    color: #fff;
    box-shadow: 0 10px 28px rgb(31 123 255 / 30%);
  }

  .btn.primary:hover {
    filter: brightness(1.06);
  }

  .btn.ghost {
    background: rgb(10 22 48 / 70%);
    border-color: rgb(80 130 220 / 35%);
    color: #d6e4ff;
  }

  .btn.ghost:hover {
    border-color: rgb(100 160 255 / 55%);
  }

  /* Architecture diagram */
  .arch {
    display: grid;
    grid-template-columns: 1fr auto;
    gap: 12px;
    align-items: stretch;
  }

  .arch-main {
    display: flex;
    flex-direction: column;
    gap: 0;
    min-width: 0;
  }

  .layer {
    border-radius: 14px;
    padding: 14px 14px 12px;
    border: 1px solid transparent;
  }

  .layer-ai {
    border-color: rgb(70 140 255 / 45%);
    background: linear-gradient(165deg, rgb(12 30 70 / 90%), rgb(8 18 45 / 92%));
    box-shadow: 0 0 0 1px rgb(40 100 255 / 8%), 0 16px 40px rgb(10 40 120 / 25%);
  }

  .layer-data {
    border-color: rgb(40 180 140 / 40%);
    background: linear-gradient(165deg, rgb(8 35 32 / 88%), rgb(6 22 28 / 92%));
    box-shadow: 0 0 0 1px rgb(30 160 120 / 8%), 0 16px 40px rgb(0 40 30 / 22%);
  }

  .layer-head {
    font-size: 10px;
    letter-spacing: 0.08em;
    font-weight: 700;
    color: #9eb8e8;
    margin-bottom: 12px;
    text-align: center;
  }

  .layer-data .layer-head {
    color: #8fd9c0;
  }

  .layer-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 8px;
  }

  .chip {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 6px;
    padding: 10px 6px;
    border-radius: 10px;
    border: 1px solid rgb(80 130 220 / 22%);
    background: rgb(8 16 36 / 55%);
    font-size: 11px;
    text-align: center;
    color: #c8d6ee;
    line-height: 1.25;
    min-height: 72px;
    justify-content: center;
  }

  .layer-data .chip {
    border-color: rgb(50 160 130 / 25%);
    background: rgb(6 22 24 / 55%);
    color: #c5e8dc;
  }

  .chip-ico {
    font-size: 16px;
    opacity: 0.9;
  }

  .layer-foot {
    margin-top: 10px;
    text-align: center;
    font-size: 11px;
    color: #7f96bb;
  }

  .layer-data .layer-foot {
    color: #7db5a3;
  }

  .env-badge {
    margin: 10px auto 0;
    width: fit-content;
    padding: 5px 12px;
    border-radius: 999px;
    border: 1px solid rgb(50 170 140 / 35%);
    background: rgb(10 40 36 / 70%);
    color: #8fd9c0;
    font-size: 10px;
    font-weight: 700;
    letter-spacing: 0.04em;
  }

  .connector {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 4px;
    padding: 6px 0;
    color: #6f8ab5;
    font-size: 10px;
  }

  .conn-line {
    width: 1px;
    height: 10px;
    background: linear-gradient(180deg, rgb(80 140 255 / 10%), rgb(80 140 255 / 55%), rgb(60 200 160 / 55%));
  }

  .conn-label {
    padding: 3px 10px;
    border-radius: 999px;
    border: 1px dashed rgb(90 140 220 / 35%);
    background: rgb(8 16 36 / 70%);
  }

  .ownership {
    width: 118px;
    border-radius: 14px;
    border: 1px solid rgb(70 120 200 / 22%);
    background: linear-gradient(180deg, rgb(10 20 42 / 90%), rgb(6 12 28 / 92%));
    padding: 14px 10px;
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 12px;
  }

  .own-label {
    font-size: 9px;
    letter-spacing: 0.14em;
    color: #6d7c99;
    font-weight: 700;
  }

  .own-card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 8px;
    text-align: center;
    font-size: 11px;
    color: #c4d2ea;
    font-weight: 650;
    line-height: 1.3;
  }

  .own-card img,
  .you-ico {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
  }

  .you-ico {
    display: grid;
    place-items: center;
    background: rgb(20 80 70 / 40%);
    border: 1px solid rgb(60 180 140 / 35%);
    font-size: 18px;
  }

  .own-card.monti img {
    box-shadow: 0 0 0 2px rgb(40 120 255 / 35%), 0 0 16px rgb(40 120 255 / 25%);
  }

  .own-divider {
    width: 1px;
    flex: 1;
    min-height: 40px;
    border-left: 1px dashed rgb(100 140 200 / 35%);
  }

  /* Trust */
  .trust-bar {
    position: relative;
    z-index: 1;
    width: min(1180px, 100%);
    margin: 0 auto 28px;
    padding: 14px 20px;
    border-radius: 14px;
    border: 1px solid rgb(70 130 220 / 28%);
    background: linear-gradient(100deg, rgb(10 24 52 / 90%), rgb(8 18 40 / 92%));
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    gap: 14px 22px;
  }

  .trust-title {
    display: inline-flex;
    align-items: center;
    gap: 8px;
    white-space: nowrap;
  }

  .trust-title strong {
    font-size: 14px;
  }

  .shield {
    font-size: 16px;
  }

  .trust-bar ul {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-wrap: wrap;
    gap: 8px 16px;
  }

  .trust-bar li {
    display: inline-flex;
    align-items: center;
    gap: 6px;
    color: #9eb0cc;
    font-size: 12.5px;
  }

  .trust-bar li span {
    color: #4fd4a0;
    font-weight: 700;
  }

  .section {
    position: relative;
    z-index: 1;
    width: min(1180px, 100%);
    margin: 0 auto 26px;
  }

  .section.tight {
    margin: 0;
  }

  .section-label {
    margin: 0 0 12px;
    font-size: 11px;
    letter-spacing: 0.14em;
    text-transform: uppercase;
    color: #6d7c99;
    font-weight: 700;
  }

  .how-grid {
    display: grid;
    grid-template-columns: repeat(4, minmax(0, 1fr));
    gap: 12px;
  }

  .how-card {
    padding: 18px 14px;
    border-radius: 14px;
    border: 1px solid rgb(70 120 200 / 22%);
    background: linear-gradient(165deg, rgb(10 22 48 / 88%), rgb(5 12 28 / 92%));
    min-height: 168px;
  }

  .how-ico {
    display: grid;
    place-items: center;
    width: 36px;
    height: 36px;
    border-radius: 10px;
    margin-bottom: 12px;
    background: rgb(20 60 150 / 30%);
    border: 1px solid rgb(70 140 240 / 28%);
    font-size: 16px;
  }

  .how-card h3 {
    margin: 0 0 8px;
    font-size: 13.5px;
    font-weight: 700;
  }

  .how-card p {
    margin: 0;
    color: #8ea0bd;
    font-size: 12.5px;
    line-height: 1.5;
  }

  .split {
    position: relative;
    z-index: 1;
    width: min(1180px, 100%);
    margin: 0 auto 26px;
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }

  .panel {
    border-radius: 16px;
    padding: 18px 18px 16px;
    border: 1px solid transparent;
  }

  .panel-monti {
    border-color: rgb(70 140 255 / 35%);
    background: linear-gradient(165deg, rgb(12 28 68 / 90%), rgb(8 16 40 / 92%));
  }

  .panel-you {
    border-color: rgb(40 180 140 / 35%);
    background: linear-gradient(165deg, rgb(8 36 34 / 88%), rgb(6 20 28 / 92%));
  }

  .panel-head {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-bottom: 14px;
  }

  .panel-head img {
    border-radius: 50%;
    box-shadow: 0 0 0 2px rgb(40 120 255 / 30%);
  }

  .you-badge {
    width: 28px;
    height: 28px;
    border-radius: 50%;
    display: grid;
    place-items: center;
    background: rgb(20 80 70 / 45%);
    border: 1px solid rgb(60 180 140 / 35%);
    font-size: 14px;
  }

  .panel-head h2 {
    margin: 0;
    font-size: 13px;
    letter-spacing: 0.06em;
    font-weight: 750;
  }

  .panel-items {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
    margin-bottom: 14px;
  }

  .panel-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 10px;
    border-radius: 10px;
    border: 1px solid rgb(80 120 200 / 18%);
    background: rgb(6 14 32 / 45%);
    font-size: 12.5px;
    color: #c8d4ea;
    font-weight: 600;
  }

  .panel-you .panel-item {
    border-color: rgb(50 150 120 / 20%);
    background: rgb(6 22 24 / 45%);
    color: #c5e8dc;
  }

  .dot {
    width: 7px;
    height: 7px;
    border-radius: 50%;
    flex-shrink: 0;
  }

  .dot.blue {
    background: #4ea3ff;
    box-shadow: 0 0 8px rgb(78 163 255 / 50%);
  }

  .dot.green {
    background: #4fd4a0;
    box-shadow: 0 0 8px rgb(79 212 160 / 45%);
  }

  .panel-foot {
    margin: 0;
    font-size: 12px;
    color: #8ea0bd;
    line-height: 1.45;
  }

  .panel-you .panel-foot {
    color: #7db5a3;
  }

  .bottom-grid {
    position: relative;
    z-index: 1;
    width: min(1180px, 100%);
    margin: 0 auto 26px;
    display: grid;
    grid-template-columns: 1fr 1.2fr;
    gap: 16px;
  }

  .deploy-grid {
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 10px;
  }

  .mini-card {
    padding: 14px 12px;
    border-radius: 12px;
    border: 1px solid rgb(70 120 200 / 22%);
    background: linear-gradient(165deg, rgb(10 22 48 / 88%), rgb(5 12 28 / 92%));
  }

  .mini-ico {
    display: grid;
    place-items: center;
    width: 32px;
    height: 32px;
    border-radius: 9px;
    margin-bottom: 10px;
    background: rgb(20 60 150 / 28%);
    border: 1px solid rgb(70 140 240 / 25%);
    font-size: 14px;
  }

  .mini-card h3 {
    margin: 0 0 6px;
    font-size: 13px;
  }

  .mini-card p {
    margin: 0;
    color: #8ea0bd;
    font-size: 12px;
    line-height: 1.45;
  }

  .ben-grid {
    display: grid;
    grid-template-columns: repeat(2, minmax(0, 1fr));
    gap: 8px;
  }

  .ben-card {
    display: flex;
    gap: 10px;
    align-items: flex-start;
    padding: 12px;
    border-radius: 12px;
    border: 1px solid rgb(70 120 200 / 18%);
    background: rgb(8 16 36 / 55%);
  }

  .ben-ico {
    width: 28px;
    height: 28px;
    border-radius: 8px;
    display: grid;
    place-items: center;
    flex-shrink: 0;
    background: rgb(20 60 150 / 28%);
    border: 1px solid rgb(70 140 240 / 22%);
    font-size: 13px;
  }

  .ben-card h3 {
    margin: 0 0 3px;
    font-size: 12.5px;
  }

  .ben-card p {
    margin: 0;
    color: #8ea0bd;
    font-size: 11.5px;
    line-height: 1.4;
  }

  .cta-banner {
    position: relative;
    z-index: 1;
    width: min(1180px, 100%);
    margin: 0 auto;
    padding: 26px 28px;
    border-radius: 18px;
    border: 1px solid rgb(70 130 220 / 30%);
    background: linear-gradient(110deg, rgb(8 20 48 / 94%), rgb(6 14 34 / 95%) 50%, rgb(12 32 80 / 75%));
    box-shadow: 0 20px 50px rgb(0 12 40 / 35%);
    display: grid;
    grid-template-columns: 1.4fr auto;
    gap: 20px;
    align-items: center;
    overflow: hidden;
  }

  .cta-copy {
    position: relative;
    z-index: 1;
  }

  .cta-copy h2 {
    margin: 0 0 8px;
    font-size: clamp(1.15rem, 2vw, 1.4rem);
    letter-spacing: -0.02em;
    max-width: 28ch;
  }

  .cta-copy p {
    margin: 0;
    color: #93a6c4;
    font-size: 13.5px;
    line-height: 1.55;
    max-width: 52ch;
  }

  .cta-actions {
    position: relative;
    z-index: 1;
  }

  .cta-wave {
    position: absolute;
    right: -30px;
    top: 0;
    bottom: 0;
    width: 40%;
    pointer-events: none;
    background:
      radial-gradient(ellipse at 70% 50%, rgb(30 100 255 / 20%), transparent 55%),
      repeating-linear-gradient(100deg, transparent 0 18px, rgb(40 120 255 / 8%) 18px 19px);
    mask-image: linear-gradient(90deg, transparent, #000 40%);
    opacity: 0.85;
  }

  @media (max-width: 1100px) {
    .hero {
      grid-template-columns: 1fr;
    }

    .how-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .bottom-grid {
      grid-template-columns: 1fr;
    }
  }

  @media (max-width: 720px) {
    .ent-page {
      padding: 28px 16px 48px;
    }

    .arch {
      grid-template-columns: 1fr;
    }

    .ownership {
      width: 100%;
      flex-direction: row;
      justify-content: space-around;
      min-height: auto;
    }

    .own-divider {
      width: 40px;
      min-height: 1px;
      border-left: 0;
      border-top: 1px dashed rgb(100 140 200 / 35%);
      flex: 0 0 auto;
      align-self: center;
    }

    .layer-grid {
      grid-template-columns: repeat(2, minmax(0, 1fr));
    }

    .how-grid,
    .split,
    .deploy-grid,
    .ben-grid,
    .panel-items {
      grid-template-columns: 1fr;
    }

    .cta-banner {
      grid-template-columns: 1fr;
      text-align: left;
      padding: 22px 18px;
    }
  }
</style>
