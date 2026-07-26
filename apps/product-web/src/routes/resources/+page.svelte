<script lang="ts">
  import { base } from '$app/paths';
  import { t } from '$lib/i18n';
  import { highlightCode } from '$lib/highlight';
  import '$lib/highlight-theme.css';

  type SdkId = 'web-component' | 'vue' | 'react' | 'svelte';

  let active = $state<SdkId>('web-component');
  let copied = $state(false);
  let mobileCopied = $state(false);

  const mobileInstall = 'npm install @monti/mobile-sdk';
  const mobileCode = `import { MontiMobileClient } from "@monti/mobile-sdk";

const client = new MontiMobileClient({
  baseUrl: "https://your-monti-host.example",
  tenantId: "your-tenant-slug", // optional when already authenticated
  tokenStore, // secure keychain / keystore on device
  websocket: (url) => new WebSocket(url),
});

// Load tenant brand, avatars, locale, quota
const bootstrap = await client.getBootstrap();

// Customer picks a branded AI agent, then direct call
const call = await client.createCall({
  avatarId: bootstrap.default_avatar_id,
});
const handle = await client.connectCall(call.call_id);
handle.onEvent((event) => console.log(event));`;

  const sdks = $derived([
    {
      id: 'web-component' as SdkId,
      label: 'Web Component',
      package: '@monti/embed-web-component',
      lang: 'html',
      install: 'npm install @monti/embed-web-component @monti/embed-core',
      code:
        '<script type="module">\n' +
        '  import "@monti/embed-web-component";\n' +
        '</' +
        'script>\n\n' +
        '<monti-embed\n' +
        '  embed-key="emb_YOUR_KEY"\n' +
        '  api-base="http://localhost:8091"\n' +
        '  position="bottom-right"\n' +
        '></monti-embed>\n\n' +
        '<' +
        'script>\n' +
        '  const el = document.querySelector("monti-embed");\n' +
        '  el.addEventListener("monti-open", () => console.log("open"));\n' +
        '  el.addEventListener("monti-error", (e) => console.error(e.detail));\n' +
        '</' +
        'script>'
    },
    {
      id: 'vue' as SdkId,
      label: 'Vue 3',
      package: '@monti/embed-vue',
      lang: 'html',
      install: 'npm install @monti/embed-vue @monti/embed-core',
      code:
        '<' +
        'script setup lang="ts">\n' +
        'import { MontiEmbedVue } from "@monti/embed-vue";\n' +
        '</' +
        'script>\n\n' +
        '<template>\n' +
        '  <MontiEmbedVue\n' +
        '    embed-key="emb_YOUR_KEY"\n' +
        '    api-base="http://localhost:8091"\n' +
        '    position="bottom-right"\n' +
        '    @open="() => {}"\n' +
        '    @close="() => {}"\n' +
        '    @error="(e) => console.error(e)"\n' +
        '  />\n' +
        '</template>'
    },
    {
      id: 'react' as SdkId,
      label: 'React',
      package: '@monti/embed-react',
      lang: 'typescript',
      install: 'npm install @monti/embed-react @monti/embed-core',
      code: `import { MontiEmbedReact } from "@monti/embed-react";

export function SupportWidget() {
  return (
    <MontiEmbedReact
      embedKey="emb_YOUR_KEY"
      apiBase="http://localhost:8091"
      position="bottom-right"
      onError={(e) => console.error(e.code, e.message)}
    />
  );
}`
    },
    {
      id: 'svelte' as SdkId,
      label: 'Svelte',
      package: '@monti/embed-svelte',
      lang: 'html',
      install: 'npm install @monti/embed-svelte @monti/embed-core',
      code:
        '<' +
        'script>\n' +
        '  import MontiEmbed from "@monti/embed-svelte/MontiEmbed.svelte";\n' +
        '</' +
        'script>\n\n' +
        '<MontiEmbed\n' +
        '  embedKey="emb_YOUR_KEY"\n' +
        '  apiBase="http://localhost:8091"\n' +
        '/>'
    }
  ]);

  const current = $derived(sdks.find((s) => s.id === active) ?? sdks[0]);
  const installHtml = $derived(highlightCode(current.install, 'bash'));
  const codeHtml = $derived(highlightCode(current.code, current.lang));
  const mobileInstallHtml = $derived(highlightCode(mobileInstall, 'bash'));
  const mobileCodeHtml = $derived(highlightCode(mobileCode, 'typescript'));

  const mobileFeatures = $derived([
    { title: $t.resources_mobile_f1_t, body: $t.resources_mobile_f1_b },
    { title: $t.resources_mobile_f2_t, body: $t.resources_mobile_f2_b },
    { title: $t.resources_mobile_f3_t, body: $t.resources_mobile_f3_b }
  ]);

  const mobileShots = $derived([
    {
      src: `${base}/images/mobile/brands.png`,
      alt: $t.resources_mobile_cap_brands,
      caption: $t.resources_mobile_cap_brands
    },
    {
      src: `${base}/images/mobile/tenant.png`,
      alt: $t.resources_mobile_cap_tenant,
      caption: $t.resources_mobile_cap_tenant
    },
    {
      src: `${base}/images/mobile/call.png`,
      alt: $t.resources_mobile_cap_call,
      caption: $t.resources_mobile_cap_call
    }
  ]);

  async function copyCode() {
    try {
      await navigator.clipboard.writeText(current.code);
      copied = true;
      setTimeout(() => (copied = false), 1600);
    } catch {
      copied = false;
    }
  }

  async function copyMobileCode() {
    try {
      await navigator.clipboard.writeText(mobileCode);
      mobileCopied = true;
      setTimeout(() => (mobileCopied = false), 1600);
    } catch {
      mobileCopied = false;
    }
  }
</script>

<svelte:head>
  <title>{$t.resources_title}</title>
  <meta name="description" content={$t.resources_p} />
</svelte:head>

<section class="resources-page">
  <div class="glow" aria-hidden="true"></div>

  <div class="intro">
    <h1>{$t.resources_h1}</h1>
    <p>{$t.resources_p}</p>
  </div>

  <!-- Mobile branded direct call (first) -->
  <div class="section-label">
    <p class="eyebrow">{$t.resources_mobile_eyebrow}</p>
    <h2 class="section-h2">{$t.resources_mobile_h2}</h2>
    <p class="section-p">{$t.resources_mobile_p}</p>
  </div>

  <div class="feature-row">
    {#each mobileFeatures as item}
      <article class="feature-card">
        <h3>{item.title}</h3>
        <p>{item.body}</p>
      </article>
    {/each}
  </div>

  <div class="phone-row">
    {#each mobileShots as shot}
      <figure class="phone-shot">
        <div class="phone-frame">
          <img src={shot.src} alt={shot.alt} loading="lazy" />
        </div>
        <figcaption>{shot.caption}</figcaption>
      </figure>
    {/each}
  </div>

  <article class="sdk-panel mobile-sdk">
    <header class="sdk-head">
      <div>
        <span class="pkg">@monti/mobile-sdk</span>
        <h2>{$t.resources_mobile_sdk_h3}</h2>
        <p class="sdk-sub muted">{$t.resources_mobile_sdk_p}</p>
      </div>
      <button type="button" class="copy-btn" onclick={copyMobileCode}>
        {mobileCopied ? $t.resources_copied : $t.resources_copy}
      </button>
    </header>

    <div class="block">
      <h3>{$t.resources_install}</h3>
      <pre class="code"><code class="hljs language-bash">{@html mobileInstallHtml}</code></pre>
    </div>

    <div class="block">
      <h3>{$t.resources_example}</h3>
      <pre class="code example"><code class="hljs language-typescript">{@html mobileCodeHtml}</code></pre>
    </div>
  </article>

  <!-- Web embed SDKs -->
  <div class="section-label web-label">
    <p class="eyebrow">{$t.resources_web_eyebrow}</p>
    <h2 class="section-h2">{$t.resources_web_h2}</h2>
  </div>

  <div class="sdk-tabs" role="tablist" aria-label="Embed SDK">
    {#each sdks as sdk}
      <button
        type="button"
        role="tab"
        class:active={active === sdk.id}
        aria-selected={active === sdk.id}
        onclick={() => {
          active = sdk.id;
          copied = false;
        }}
      >
        {sdk.label}
      </button>
    {/each}
  </div>

  <article class="sdk-panel">
    <header class="sdk-head">
      <div>
        <span class="pkg">{current.package}</span>
        <h2>{current.label}</h2>
      </div>
      <button type="button" class="copy-btn" onclick={copyCode}>
        {copied ? $t.resources_copied : $t.resources_copy}
      </button>
    </header>

    <div class="block">
      <h3>{$t.resources_install}</h3>
      <pre class="code"><code class="hljs language-bash">{@html installHtml}</code></pre>
    </div>

    <div class="block">
      <h3>{$t.resources_example}</h3>
      <pre class="code example"><code class="hljs language-{current.lang}">{@html codeHtml}</code></pre>
    </div>

    <p class="hint muted">{$t.resources_hint}</p>
  </article>
</section>

<style>
  .resources-page {
    position: relative;
    overflow: hidden;
    padding: 52px 24px 72px;
    background:
      radial-gradient(circle at 78% 12%, rgb(30 100 255 / 12%), transparent 26%),
      linear-gradient(180deg, #020713 0%, #031025 52%, #020713 100%);
  }

  .glow {
    position: absolute;
    inset: 0;
    pointer-events: none;
    opacity: 0.22;
    background-image:
      linear-gradient(rgb(30 90 180 / 12%) 1px, transparent 1px),
      linear-gradient(90deg, rgb(30 90 180 / 12%) 1px, transparent 1px);
    background-size: 56px 56px;
    mask-image: linear-gradient(180deg, #000, transparent 72%);
  }

  .intro {
    position: relative;
    z-index: 1;
    width: min(1080px, 100%);
    margin: 0 auto 28px;
  }

  .intro h1 {
    margin: 0 0 10px;
    font-size: clamp(1.9rem, 3.4vw, 2.5rem);
    letter-spacing: -0.03em;
    font-weight: 750;
  }

  .intro p {
    margin: 0;
    color: #93a6c4;
    font-size: 1.05rem;
    line-height: 1.55;
    max-width: 62ch;
  }

  .section-label {
    position: relative;
    z-index: 1;
    width: min(1080px, 100%);
    margin: 0 auto 14px;
  }

  .section-label.web-label {
    margin-top: 48px;
  }

  .eyebrow {
    margin: 0 0 8px;
    font-size: 11px;
    letter-spacing: 0.16em;
    text-transform: uppercase;
    color: #5f8fd4;
    font-weight: 700;
  }

  .section-h2 {
    margin: 0 0 10px;
    font-size: clamp(1.25rem, 2.2vw, 1.55rem);
    letter-spacing: -0.02em;
  }

  .section-p {
    margin: 0;
    color: #93a6c4;
    font-size: 15px;
    line-height: 1.6;
    max-width: 70ch;
  }

  .feature-row {
    position: relative;
    z-index: 1;
    width: min(1080px, 100%);
    margin: 0 auto 22px;
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 12px;
  }

  .feature-card {
    padding: 16px;
    border-radius: 14px;
    border: 1px solid rgb(70 120 190 / 20%);
    background: linear-gradient(165deg, rgb(10 22 48 / 85%), rgb(5 12 28 / 90%));
  }

  .feature-card h3 {
    margin: 0 0 8px;
    font-size: 14px;
    color: #9ecbff;
  }

  .feature-card p {
    margin: 0;
    font-size: 13px;
    color: #8ea0bd;
    line-height: 1.5;
  }

  .phone-row {
    position: relative;
    z-index: 1;
    width: min(1080px, 100%);
    margin: 0 auto 22px;
    display: grid;
    grid-template-columns: repeat(3, minmax(0, 1fr));
    gap: 16px;
    align-items: start;
  }

  .phone-shot {
    margin: 0;
    text-align: center;
  }

  .phone-frame {
    border-radius: 18px;
    border: none;
    background: transparent;
    padding: 0;
    box-shadow: none;
    overflow: visible;
  }

  .phone-frame img {
    display: block;
    width: 100%;
    height: auto;
    border-radius: 0;
    object-fit: contain;
    object-position: top center;
    max-height: 440px;
    margin: 0 auto;
    filter: drop-shadow(0 16px 28px rgb(0 12 40 / 45%));
  }

  .phone-shot figcaption {
    margin-top: 10px;
    font-size: 12px;
    font-weight: 650;
    color: #8eb6f0;
  }

  .sdk-tabs {
    position: relative;
    z-index: 1;
    width: min(1080px, 100%);
    margin: 0 auto 18px;
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }

  .sdk-tabs button {
    border: 1px solid rgb(70 120 190 / 28%);
    background: rgb(8 16 34 / 70%);
    color: #a8b8d2;
    border-radius: 999px;
    padding: 9px 16px;
    font-size: 13px;
    font-weight: 650;
    cursor: pointer;
  }

  .sdk-tabs button:hover {
    color: #fff;
    border-color: rgb(70 140 255 / 40%);
  }

  .sdk-tabs button.active {
    color: #fff;
    background: linear-gradient(100deg, #1f7bff, #2f6dff);
    border-color: transparent;
    box-shadow: 0 8px 22px rgb(31 123 255 / 28%);
  }

  .sdk-panel {
    position: relative;
    z-index: 1;
    width: min(1080px, 100%);
    margin: 0 auto;
    padding: 22px 22px 24px;
    border-radius: 18px;
    border: 1px solid rgb(70 130 220 / 26%);
    background: linear-gradient(165deg, rgb(10 22 48 / 92%), rgb(5 12 28 / 96%));
    box-shadow: 0 20px 50px rgb(0 12 40 / 30%);
  }

  .sdk-panel.mobile-sdk {
    margin-top: 8px;
    margin-bottom: 0;
  }

  .sdk-sub {
    margin: 6px 0 0;
    font-size: 13px;
    line-height: 1.45;
    max-width: 52ch;
  }

  .sdk-head {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 12px;
    margin-bottom: 18px;
  }

  .pkg {
    display: block;
    font-size: 12px;
    color: #6ea8ff;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
    margin-bottom: 4px;
  }

  .sdk-head h2 {
    margin: 0;
    font-size: 1.25rem;
  }

  .copy-btn {
    border: 1px solid rgb(80 130 200 / 35%);
    background: rgb(12 28 58 / 80%);
    color: #d5e4f8;
    border-radius: 9px;
    padding: 8px 12px;
    font-size: 12px;
    font-weight: 700;
    cursor: pointer;
    white-space: nowrap;
  }

  .copy-btn:hover {
    border-color: rgb(90 160 255 / 50%);
  }

  .block {
    margin-bottom: 16px;
  }

  .block h3 {
    margin: 0 0 8px;
    font-size: 12px;
    letter-spacing: 0.12em;
    text-transform: uppercase;
    color: #6f8fc4;
  }

  .code {
    margin: 0;
    padding: 14px 16px;
    border-radius: 12px;
    border: 1px solid rgb(60 110 190 / 22%);
    background: rgb(3 8 20 / 88%);
    overflow-x: auto;
    color: #d7e4f8;
    font-size: 12.5px;
    line-height: 1.55;
    font-family: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace;
  }

  .code.example {
    max-height: 420px;
    overflow: auto;
  }

  .hint {
    margin: 8px 0 0;
    font-size: 13px;
    line-height: 1.5;
  }

  .muted {
    color: #8ea0bd;
  }

  @media (max-width: 900px) {
    .feature-row,
    .phone-row {
      grid-template-columns: 1fr;
    }

    .phone-frame img {
      max-height: 520px;
      width: auto;
      max-width: 100%;
      margin: 0 auto;
    }
  }

  @media (max-width: 640px) {
    .resources-page {
      padding: 32px 16px 48px;
    }

    .sdk-panel {
      padding: 16px;
    }

    .sdk-head {
      flex-direction: column;
    }
  }
</style>
