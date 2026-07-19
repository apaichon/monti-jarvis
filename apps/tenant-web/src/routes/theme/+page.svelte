<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';
  import { ApiError } from '$lib/api/http';
  import {
    TOKEN_KEYS,
    getTheme,
    putTheme,
    publishTheme,
    resetTheme,
    uploadThemeLogo,
    type ThemeBranding,
    type ThemeTokens,
    type TenantTheme
  } from '$lib/api/theme';

  let loading = $state(true);
  let saving = $state(false);
  let publishing = $state(false);
  let preset = $state('dark');
  let branding = $state<ThemeBranding>({
    brand_name: '',
    subtitle: '',
    logo_url: '',
    logo_alt: ''
  });
  let tokens = $state<ThemeTokens>({});
  let contrastOk = $state(true);
  let contrastPairs = $state<{ pair: string; ratio: number; pass: boolean }[]>([]);
  let publishedAt = $state<string | null>(null);

  const tokenLabels: Record<string, string> = {
    primary: 'Primary',
    primary_text: 'Primary text',
    accent: 'Accent',
    background: 'Background',
    surface: 'Surface',
    surface_elevated: 'Surface elevated',
    text: 'Text',
    muted: 'Muted text',
    line: 'Border / line',
    success: 'Success',
    warn: 'Warning',
    danger: 'Danger'
  };

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/theme`)}`);
      return;
    }
    await load();
  });

  function applyRow(row: TenantTheme) {
    preset = row.preset || 'dark';
    branding = {
      brand_name: row.draft_branding?.brand_name || '',
      subtitle: row.draft_branding?.subtitle || '',
      logo_url: row.draft_branding?.logo_url || '',
      logo_alt: row.draft_branding?.logo_alt || ''
    };
    tokens = { ...(row.draft_tokens || {}) };
    contrastOk = row.contrast_report?.ok !== false;
    contrastPairs = row.contrast_report?.pairs || [];
    publishedAt = row.published_at || null;
  }

  async function load() {
    loading = true;
    try {
      const row = await getTheme();
      applyRow(row);
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load theme');
    } finally {
      loading = false;
    }
  }

  async function saveDraft() {
    saving = true;
    try {
      const row = await putTheme({ preset, branding, tokens });
      applyRow(row);
      feedback.success('Draft saved');
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Save failed');
    } finally {
      saving = false;
    }
  }

  async function doPublish(confirmLow = false) {
    publishing = true;
    try {
      await putTheme({ preset, branding, tokens });
      const row = await publishTheme(confirmLow);
      applyRow(row);
      feedback.success('Theme published — customer & embed will use it');
    } catch (err) {
      if (err instanceof ApiError && err.status === 409) {
        if (confirm('Contrast checks failed. Publish anyway?')) {
          await doPublish(true);
          return;
        }
        feedback.error('Publish cancelled — fix colors or confirm low contrast');
      } else {
        feedback.error(err instanceof ApiError ? err.message : 'Publish failed');
      }
    } finally {
      publishing = false;
    }
  }

  async function doReset() {
    if (!confirm('Reset draft colors to preset defaults?')) return;
    saving = true;
    try {
      const row = await resetTheme(preset, false);
      applyRow(row);
      feedback.success('Draft reset to preset');
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Reset failed');
    } finally {
      saving = false;
    }
  }

  async function onLogo(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    if (!file) return;
    try {
      const res = await uploadThemeLogo(file);
      applyRow(res.theme);
      branding = { ...branding, logo_url: res.logo_url };
      feedback.success('Logo uploaded');
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Logo upload failed');
    } finally {
      input.value = '';
    }
  }

  function previewStyle(): string {
    const p = tokens.primary || '#006dff';
    const pt = tokens.primary_text || '#fff';
    const a = tokens.accent || '#00b7ff';
    const bg = tokens.background || '#020712';
    const s = tokens.surface || '#05101f';
    const se = tokens.surface_elevated || '#08172a';
    const t = tokens.text || '#f7fbff';
    const m = tokens.muted || '#8fa5bf';
    const l = tokens.line || '#3d5a80';
    return [
      `--p:${p}`,
      `--pt:${pt}`,
      `--a:${a}`,
      `--bg:${bg}`,
      `--s:${s}`,
      `--se:${se}`,
      `--t:${t}`,
      `--m:${m}`,
      `--l:${l}`
    ].join(';');
  }

  const previewName = $derived(branding.brand_name?.trim() || 'Your Brand');
  const previewSub = $derived(branding.subtitle?.trim() || 'AI · text & voice');
  const previewLogo = $derived(branding.logo_url?.trim() || `${base}/images/monti-logo.png`);
</script>

<div style="display:flex;justify-content:space-between;align-items:center;gap:12px;margin-bottom:20px;flex-wrap:wrap">
  <div>
    <h1 style="margin:0;font-size:24px">Theme & branding</h1>
    <p style="margin:6px 0 0;color:var(--muted);font-size:13px">
      Customize caller desk & embed: logo, brand name, subtitle, and colors.
    </p>
  </div>
</div>

{#if loading}
  <p style="color:var(--muted)">Loading…</p>
{:else}
  <div class="grid">
    <div class="col">
      <div class="card" style="margin-bottom:16px">
        <h2 style="margin:0 0 12px;font-size:16px">Brand identity</h2>
        <div class="field">
          <label for="bn">Brand name</label>
          <input id="bn" type="text" bind:value={branding.brand_name} maxlength="80" placeholder="Libra Tech Co.,Ltd" />
        </div>
        <div class="field">
          <label for="st">Subtitle</label>
          <input id="st" type="text" bind:value={branding.subtitle} maxlength="120" placeholder="AI · text & voice" />
        </div>
        <div class="field">
          <label for="logo">Logo</label>
          <div style="display:flex;gap:10px;align-items:center;flex-wrap:wrap">
            {#if branding.logo_url}
              <img src={branding.logo_url} alt="" width="40" height="40" style="border-radius:50%;object-fit:cover;border:1px solid var(--line)" />
            {/if}
            <input id="logo" type="file" accept="image/png,image/jpeg,image/webp,image/gif" onchange={onLogo} />
          </div>
          <p style="margin:8px 0 0;font-size:12px;color:var(--muted)">PNG/JPEG/WebP, max 1MB. Empty uses Monti mark.</p>
        </div>
        <div class="field">
          <label for="alt">Logo alt text</label>
          <input id="alt" type="text" bind:value={branding.logo_alt} maxlength="80" placeholder="Company logo" />
        </div>
      </div>

      <div class="card">
        <h2 style="margin:0 0 12px;font-size:16px">Colors</h2>
        <div class="field">
          <label for="preset">Preset</label>
          <select id="preset" bind:value={preset}>
            <option value="dark">Dark</option>
            <option value="light">Light</option>
            <option value="branded">Branded</option>
          </select>
        </div>
        <div class="token-grid">
          {#each TOKEN_KEYS as key}
            <label class="token">
              <span>{tokenLabels[key] || key}</span>
              <input type="color" value={tokens[key] || '#000000'} oninput={(e) => (tokens[key] = (e.currentTarget as HTMLInputElement).value)} />
              <input
                type="text"
                class="hex"
                value={tokens[key] || ''}
                oninput={(e) => (tokens[key] = (e.currentTarget as HTMLInputElement).value)}
              />
            </label>
          {/each}
        </div>
        <div style="margin-top:14px;font-size:12px;color:var(--muted)">
          Contrast: {#if contrastOk}<span style="color:var(--success,#3dd68c)">OK</span>{:else}<span style="color:var(--danger,#ff5c7a)">warnings</span>{/if}
          {#if contrastPairs.length}
            <ul style="margin:6px 0 0;padding-left:18px">
              {#each contrastPairs as p}
                <li class:fail={!p.pass}>{p.pair}: {p.ratio} {p.pass ? '✓' : '✗'}</li>
              {/each}
            </ul>
          {/if}
        </div>
        {#if publishedAt}
          <p style="margin:12px 0 0;font-size:12px;color:var(--muted)">Last published: {new Date(publishedAt).toLocaleString()}</p>
        {/if}
        <div style="margin-top:16px;display:flex;gap:8px;flex-wrap:wrap">
          <button class="btn" type="button" onclick={saveDraft} disabled={saving}>{saving ? 'Saving…' : 'Save draft'}</button>
          <button class="btn" type="button" onclick={() => doPublish(false)} disabled={publishing}>{publishing ? 'Publishing…' : 'Publish'}</button>
          <button class="btn ghost" type="button" onclick={doReset} disabled={saving}>Reset colors</button>
        </div>
      </div>
    </div>

    <div class="col">
      <div class="card">
        <h2 style="margin:0 0 12px;font-size:16px">Live preview</h2>
        <div class="preview" style={previewStyle()}>
          <div class="pv-hdr">
            <div class="pv-brand">
              <img src={previewLogo} alt="" width="28" height="28" />
              <div>
                <strong>{previewName}</strong>
                <span>{previewSub}</span>
              </div>
            </div>
            <span class="pv-agent">Agent ▾</span>
          </div>
          <div class="pv-orb">
            <div class="pv-halo"></div>
            <div class="pv-name">Agent</div>
            <div class="pv-role">Role · Trait</div>
          </div>
          <div class="pv-voice">
            <span class="pv-timer">00:00:00</span>
            <button class="pv-call" type="button">Start call</button>
          </div>
          <div class="pv-bubble">Hello — this is a preview of your branded embed chrome.</div>
          <div class="pv-composer">
            <span>Type a message…</span>
            <button class="pv-send" type="button">Send</button>
          </div>
        </div>
      </div>
    </div>
  </div>
{/if}

<style>
  .grid {
    display: grid;
    grid-template-columns: minmax(0, 1.1fr) minmax(280px, 0.9fr);
    gap: 16px;
    align-items: start;
  }
  @media (max-width: 960px) {
    .grid {
      grid-template-columns: 1fr;
    }
  }
  .field {
    margin-bottom: 12px;
  }
  .field label {
    display: block;
    font-size: 13px;
    color: var(--muted);
    margin-bottom: 6px;
  }
  .field input[type='text'],
  .field select {
    width: 100%;
    max-width: 420px;
    padding: 10px;
    border-radius: 10px;
    border: 1px solid var(--line);
    background: transparent;
    color: var(--ink);
    font: inherit;
  }
  .token-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(200px, 1fr));
    gap: 10px;
  }
  .token {
    display: grid;
    grid-template-columns: 1fr auto;
    grid-template-rows: auto auto;
    gap: 4px 8px;
    font-size: 12px;
    color: var(--muted);
    align-items: center;
  }
  .token span {
    grid-column: 1 / -1;
  }
  .token input[type='color'] {
    width: 40px;
    height: 32px;
    border: 0;
    background: none;
    padding: 0;
  }
  .token .hex {
    width: 100%;
    padding: 6px 8px;
    border-radius: 8px;
    border: 1px solid var(--line);
    background: transparent;
    color: var(--ink);
    font: inherit;
    font-size: 12px;
  }
  .fail {
    color: var(--danger, #ff5c7a);
  }
  .preview {
    border-radius: 16px;
    overflow: hidden;
    border: 1px solid var(--l);
    background: var(--bg);
    color: var(--t);
    font-size: 12px;
    min-height: 420px;
    display: flex;
    flex-direction: column;
  }
  .pv-hdr {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 14px;
    background: var(--s);
    border-bottom: 1px solid var(--l);
  }
  .pv-brand {
    display: flex;
    gap: 8px;
    align-items: center;
  }
  .pv-brand img {
    border-radius: 50%;
    width: 28px;
    height: 28px;
    object-fit: cover;
  }
  .pv-brand strong {
    display: block;
    font-size: 13px;
  }
  .pv-brand span {
    display: block;
    font-size: 10px;
    color: var(--m);
  }
  .pv-agent {
    border: 1px solid var(--l);
    border-radius: 8px;
    padding: 4px 8px;
    color: var(--m);
  }
  .pv-orb {
    text-align: center;
    padding: 20px 12px 8px;
  }
  .pv-halo {
    width: 88px;
    height: 88px;
    margin: 0 auto 8px;
    border-radius: 50%;
    border: 2px solid var(--a);
    box-shadow: 0 0 24px color-mix(in srgb, var(--a) 50%, transparent);
    background: var(--se);
  }
  .pv-name {
    font-weight: 600;
  }
  .pv-role {
    color: var(--m);
    font-size: 11px;
  }
  .pv-voice {
    display: flex;
    gap: 8px;
    padding: 8px 12px;
  }
  .pv-timer {
    border: 1px solid var(--l);
    border-radius: 10px;
    padding: 8px 10px;
    color: var(--m);
  }
  .pv-call {
    flex: 1;
    border: 0;
    border-radius: 10px;
    padding: 10px;
    font-weight: 600;
    color: var(--pt);
    background: linear-gradient(135deg, var(--p), var(--a));
    cursor: default;
  }
  .pv-bubble {
    margin: 8px 12px;
    padding: 10px 12px;
    border-radius: 14px;
    background: var(--se);
    border: 1px solid var(--l);
    line-height: 1.4;
  }
  .pv-composer {
    margin-top: auto;
    display: flex;
    gap: 8px;
    padding: 12px;
    border-top: 1px solid var(--l);
    background: var(--s);
    align-items: center;
  }
  .pv-composer span {
    flex: 1;
    color: var(--m);
    border: 1px solid var(--l);
    border-radius: 10px;
    padding: 10px;
    background: var(--se);
  }
  .pv-send {
    border: 0;
    border-radius: 10px;
    padding: 10px 14px;
    font-weight: 600;
    color: var(--pt);
    background: linear-gradient(135deg, var(--p), var(--a));
    cursor: default;
  }
</style>
