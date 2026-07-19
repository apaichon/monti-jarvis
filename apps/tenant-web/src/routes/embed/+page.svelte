<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';
  import { ApiError } from '$lib/api/http';
  import {
    getEmbedConfig,
    putEmbedConfig,
    rotateEmbedKey,
    type EmbedConfig
  } from '$lib/api/embed';

  let cfg = $state<EmbedConfig | null>(null);
  let enabled = $state(false);
  let originsText = $state('');
  let defaultAgent = $state('');
  let loading = $state(true);
  let saving = $state(false);
  let rotating = $state(false);
  let tab = $state<'snippet' | 'framework'>('snippet');
  let framework = $state<'vue' | 'react' | 'svelte' | 'wc'>('vue');

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/embed`)}`);
      return;
    }
    await load();
  });

  async function load() {
    loading = true;
    try {
      cfg = await getEmbedConfig();
      enabled = cfg.enabled;
      originsText = (cfg.allowed_origins || []).join('\n');
      defaultAgent = cfg.default_agent_id || '';
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load embed config');
    } finally {
      loading = false;
    }
  }

  function parseOrigins(): string[] {
    return originsText
      .split(/\n|,/)
      .map((s) => s.trim())
      .filter(Boolean);
  }

  async function save() {
    saving = true;
    try {
      cfg = await putEmbedConfig({
        enabled,
        allowed_origins: parseOrigins(),
        default_agent_id: defaultAgent.trim()
      });
      originsText = (cfg.allowed_origins || []).join('\n');
      defaultAgent = cfg.default_agent_id || '';
      enabled = cfg.enabled;
      feedback.success('Embed settings saved');
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Save failed');
    } finally {
      saving = false;
    }
  }

  async function rotate() {
    if (!confirm('Rotate embed key? Existing snippets will stop working until updated.')) return;
    rotating = true;
    try {
      cfg = await rotateEmbedKey();
      feedback.success('Embed key rotated — update your website snippet');
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Rotate failed');
    } finally {
      rotating = false;
    }
  }

  async function copyText(text: string, label: string) {
    try {
      await navigator.clipboard.writeText(text);
      feedback.success(`${label} copied`);
    } catch {
      feedback.error('Copy failed — select and copy manually');
    }
  }

  function apiBaseFromConfig(): string {
    const snippet = cfg?.snippet || '';
    const m = snippet.match(/src=["'](https?:\/\/[^"']+)\/embed\/monti-embed\.js/i);
    if (m?.[1]) return m[1];
    if (typeof window !== 'undefined') return window.location.origin;
    return 'http://localhost:8091';
  }

  function frameworkSnippet(kind: 'vue' | 'react' | 'svelte' | 'wc'): string {
    const key = cfg?.embed_key || 'emb_YOUR_KEY';
    const api = apiBaseFromConfig();
    // Svelte compiler closes on the literal sequence less-than slash script greater-than.
    const close = ['<', '/', 'script', '>'].join('');
    const open = ['<', 'script'].join('');
    switch (kind) {
      case 'vue':
        return [
          '// npm install @monti/embed-vue @monti/embed-core',
          open + ' setup>',
          "import { MontiEmbedVue } from '@monti/embed-vue'",
          close,
          '<template>',
          '  <MontiEmbedVue',
          `    embed-key="${key}"`,
          `    api-base="${api}"`,
          '    position="bottom-right"',
          '  />',
          '</template>'
        ].join('\n');
      case 'react':
        return [
          '// npm install @monti/embed-react @monti/embed-core',
          "import { MontiEmbedReact } from '@monti/embed-react'",
          '',
          'export function MontiWidget() {',
          '  return (',
          '    <MontiEmbedReact',
          `      embedKey="${key}"`,
          `      apiBase="${api}"`,
          '      position="bottom-right"',
          '    />',
          '  )',
          '}'
        ].join('\n');
      case 'svelte':
        return [
          '<!-- npm install @monti/embed-svelte @monti/embed-core -->',
          open + '>',
          "  import MontiEmbed from '@monti/embed-svelte/MontiEmbed.svelte'",
          close,
          '<MontiEmbed',
          `  embedKey="${key}"`,
          `  apiBase="${api}"`,
          '  position="bottom-right"',
          '/>'
        ].join('\n');
      case 'wc':
        return [
          '<!-- npm install @monti/embed-web-component @monti/embed-core -->',
          open + ' type="module">',
          "  import '@monti/embed-web-component'",
          close,
          '<monti-embed',
          `  embed-key="${key}"`,
          `  api-base="${api}"`,
          '  position="bottom-right"',
          '></monti-embed>'
        ].join('\n');
    }
  }
</script>

<div style="display:flex;justify-content:space-between;align-items:center;gap:12px;margin-bottom:20px;flex-wrap:wrap">
  <div>
    <h1 style="margin:0;font-size:24px">Embed to Web</h1>
    <p style="margin:6px 0 0;color:var(--muted);font-size:13px">
      Add Monti chat to your website with a copy-paste snippet or framework SDK.
    </p>
  </div>
</div>

{#if loading}
  <p style="color:var(--muted)">Loading…</p>
{:else if cfg}
  <div class="card" style="margin-bottom:16px">
    <h2 style="margin:0 0 16px;font-size:16px">Settings</h2>

    <label class="field-row">
      <input type="checkbox" bind:checked={enabled} />
      <span>Enabled</span>
    </label>

    <div class="field" style="margin-top:16px">
      <label for="embed-key">Embed key</label>
      <div style="display:flex;gap:8px;flex-wrap:wrap;align-items:center">
        <code id="embed-key" style="flex:1;word-break:break-all">{cfg.embed_key}</code>
        <button class="btn ghost" type="button" onclick={() => copyText(cfg!.embed_key, 'Key')}>
          Copy key
        </button>
        <button class="btn ghost" type="button" onclick={rotate} disabled={rotating}>
          {rotating ? 'Rotating…' : 'Rotate key'}
        </button>
      </div>
    </div>

    <div class="field" style="margin-top:16px">
      <label for="origins">Allowed origins (one per line)</label>
      <textarea
        id="origins"
        rows="4"
        bind:value={originsText}
        placeholder="https://shop.example&#10;http://localhost:5500"
        style="width:100%;font:inherit;background:transparent;color:var(--ink);border:1px solid var(--line);border-radius:10px;padding:10px"
      ></textarea>
      <p style="margin:8px 0 0;font-size:12px;color:var(--muted)">
        Use full origins like <code>https://example.com</code> (scheme + host + optional port).
        <strong>Empty list</strong> allows any origin in development — set an allowlist before
        production.
      </p>
    </div>

    <div class="field" style="margin-top:16px">
      <label for="agent">Default agent id (optional)</label>
      <input
        id="agent"
        type="text"
        bind:value={defaultAgent}
        placeholder="ava"
        style="width:100%;max-width:280px;padding:10px;border-radius:10px;border:1px solid var(--line);background:transparent;color:var(--ink)"
      />
    </div>

    <div style="margin-top:20px">
      <button class="btn" type="button" onclick={save} disabled={saving}>
        {saving ? 'Saving…' : 'Save'}
      </button>
    </div>
  </div>

  <div class="card">
    <div class="tabs" role="tablist" aria-label="Embed integration">
      <button
        type="button"
        class="tab"
        class:active={tab === 'snippet'}
        role="tab"
        aria-selected={tab === 'snippet'}
        onclick={() => (tab = 'snippet')}
      >
        Script snippet
      </button>
      <button
        type="button"
        class="tab"
        class:active={tab === 'framework'}
        role="tab"
        aria-selected={tab === 'framework'}
        onclick={() => (tab = 'framework')}
      >
        Framework SDKs
      </button>
    </div>

    {#if tab === 'snippet'}
      <h2 style="margin:16px 0 12px;font-size:16px">Snippet</h2>
      <pre class="snippet">{cfg.snippet}</pre>
      <button class="btn" type="button" onclick={() => copyText(cfg!.snippet, 'Snippet')}>
        Copy snippet
      </button>
      <p style="margin:16px 0 0;font-size:12px;color:var(--muted)">
        Paste before <code>&lt;/body&gt;</code> on your site. Enable embed and save first. Zero
        npm dependency path.
      </p>
    {:else}
      <h2 style="margin:16px 0 12px;font-size:16px">Framework packages</h2>
      <p style="margin:0 0 12px;font-size:13px;color:var(--muted)">
        First-class SDKs share the same public resolve + iframe surface as the script tag. See repo
        guide <code>docs/EMBED_WEB_INTEGRATION.md</code> § Framework SDKs · packages
        <code>packages/embed-*</code>.
      </p>
      <div class="fw-tabs" role="tablist" aria-label="Framework">
        {#each [
          ['vue', 'Vue 3'],
          ['react', 'React'],
          ['svelte', 'Svelte'],
          ['wc', 'Web Component']
        ] as [id, label]}
          <button
            type="button"
            class="tab sm"
            class:active={framework === id}
            role="tab"
            aria-selected={framework === id}
            onclick={() => (framework = id as typeof framework)}
          >
            {label}
          </button>
        {/each}
      </div>
      <pre class="snippet">{frameworkSnippet(framework)}</pre>
      <button
        class="btn"
        type="button"
        onclick={() => copyText(frameworkSnippet(framework), `${framework} snippet`)}
      >
        Copy {framework === 'wc' ? 'Web Component' : framework} snippet
      </button>
      <p style="margin:16px 0 0;font-size:12px;color:var(--muted)">
        Add this site's origin (and your shop origin) to <strong>Allowed origins</strong>. Packages:
        <code>@monti/embed-vue</code>, <code>@monti/embed-react</code>,
        <code>@monti/embed-svelte</code>, <code>@monti/embed-web-component</code> (+
        <code>@monti/embed-core</code>).
      </p>
    {/if}
  </div>
{:else}
  <p style="color:var(--muted)">Could not load embed configuration.</p>
{/if}

<style>
  .field-row {
    display: flex;
    align-items: center;
    gap: 10px;
    font-size: 14px;
  }
  .field label {
    display: block;
    font-size: 13px;
    color: var(--muted);
    margin-bottom: 6px;
  }
  .snippet {
    background: rgb(0 0 0 / 35%);
    border: 1px solid var(--line);
    border-radius: 10px;
    padding: 12px;
    overflow: auto;
    font-size: 12px;
    line-height: 1.45;
    white-space: pre-wrap;
    word-break: break-all;
  }
  .tabs,
  .fw-tabs {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .tab {
    appearance: none;
    border: 1px solid var(--line);
    background: transparent;
    color: var(--ink);
    border-radius: 999px;
    padding: 8px 14px;
    font: inherit;
    font-size: 13px;
    cursor: pointer;
  }
  .tab.sm {
    padding: 6px 12px;
    font-size: 12px;
  }
  .tab.active {
    border-color: rgba(0, 183, 255, 0.55);
    background: rgba(0, 132, 255, 0.15);
  }
  .fw-tabs {
    margin-bottom: 12px;
  }
</style>
