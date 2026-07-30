<script lang="ts">
  import { onMount } from 'svelte';
  import {
    AVATAR_IMAGE_HINT,
    activateTenantAvatar,
    archiveTenantAvatar,
    createTenantAvatar,
    deactivateTenantAvatar,
    listGeminiSpeakerVoices,
    listTenantAvatars,
    updateTenantAvatar,
    uploadTenantAvatarImage,
    type AvatarImageVariant,
    type AvatarCap,
    type GeminiSpeakerVoice,
    type TenantAvatar
  } from '$lib/api/avatars';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  let avatars = $state<TenantAvatar[]>([]);
  let speakerVoices = $state<GeminiSpeakerVoice[]>([]);
  let voiceSource = $state('https://aistudio.google.com/generate-speech');
  let cap = $state<AvatarCap>({ active: 0, limit: 0, remaining: 0 });
  let loading = $state(true);
  let showCreate = $state(false);
  let creating = $state(false);
  let busyId = $state<string | null>(null);
  let uploadingId = $state<string | null>(null);
  let voiceBusyId = $state<string | null>(null);

  let name = $state('');
  let role = $state('');
  let trait = $state('');
  let greeting = $state('');
  let color = $state('#38bdf8');
  let voice = $state('Aoede');
  let pendingFile = $state<File | null>(null);
  let pendingPreview = $state('');
  const portraitVariants: AvatarImageVariant[] = ['default', 'dark', 'light'];

  async function load() {
    loading = true;
    try {
      const [res, voicesRes] = await Promise.all([
        listTenantAvatars(),
        listGeminiSpeakerVoices().catch(() => null)
      ]);
      avatars = res.avatars ?? [];
      cap = res.cap ?? { active: 0, limit: 0, remaining: 0 };
      if (voicesRes?.voices?.length) {
        speakerVoices = voicesRes.voices;
        if (voicesRes.source) voiceSource = voicesRes.source;
        if (!voice && speakerVoices[0]) voice = speakerVoices[0].name;
      }
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load avatars');
    } finally {
      loading = false;
    }
  }

  onMount(load);

  function resetCreateForm() {
    name = '';
    role = '';
    trait = '';
    greeting = '';
    color = '#38bdf8';
    voice = speakerVoices[0]?.name ?? 'Aoede';
    clearPendingFile();
  }

  function clearPendingFile() {
    pendingFile = null;
    if (pendingPreview) URL.revokeObjectURL(pendingPreview);
    pendingPreview = '';
  }

  function onPendingFileChange(e: Event) {
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0] ?? null;
    input.value = '';
    if (!file) return;
    if (file.size > AVATAR_IMAGE_HINT.maxBytes) {
      feedback.error(`Image exceeds ${AVATAR_IMAGE_HINT.maxLabel} limit`);
      return;
    }
    const okTypes = AVATAR_IMAGE_HINT.accept.split(',');
    if (file.type && !okTypes.includes(file.type)) {
      feedback.error(`Unsupported type. Use ${AVATAR_IMAGE_HINT.acceptLabel}.`);
      return;
    }
    clearPendingFile();
    pendingFile = file;
    pendingPreview = URL.createObjectURL(file);
  }

  async function onCreate() {
    if (!name.trim()) {
      feedback.error('Name is required');
      return;
    }
    creating = true;
    try {
      const created = await createTenantAvatar({
        name: name.trim(),
        role: role.trim(),
        trait: trait.trim(),
        greeting: greeting.trim(),
        color: color.trim() || '#38bdf8',
        voice: voice || 'Aoede'
      });
      if (pendingFile && created?.id) {
        try {
          await uploadTenantAvatarImage(created.id, pendingFile);
          feedback.success('Avatar created with portrait (library inactive)');
        } catch (upErr) {
          feedback.error(
            upErr instanceof ApiError
              ? `Avatar created, but image upload failed: ${upErr.message}`
              : 'Avatar created, but image upload failed'
          );
        }
      } else {
        feedback.success('Avatar added to library (inactive)');
      }
      resetCreateForm();
      showCreate = false;
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Create failed');
    } finally {
      creating = false;
    }
  }

  async function onRowImageUpload(av: TenantAvatar, e: Event, variant: AvatarImageVariant = 'default') {
    if (!isOwned(av)) {
      feedback.error('Only tenant-owned avatars can upload images here');
      return;
    }
    const input = e.currentTarget as HTMLInputElement;
    const file = input.files?.[0];
    input.value = '';
    if (!file) return;
    if (file.size > AVATAR_IMAGE_HINT.maxBytes) {
      feedback.error(`Image exceeds ${AVATAR_IMAGE_HINT.maxLabel} limit`);
      return;
    }
    uploadingId = av.id;
    try {
      await uploadTenantAvatarImage(av.id, file, variant);
      feedback.success(`${variantLabel(variant)} portrait uploaded`);
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Image upload failed');
    } finally {
      uploadingId = null;
    }
  }

  async function onActivate(av: TenantAvatar) {
    busyId = av.id;
    try {
      await activateTenantAvatar(av.id);
      feedback.success(`${av.name} is now active`);
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Activate failed');
    } finally {
      busyId = null;
    }
  }

  async function onDeactivate(av: TenantAvatar) {
    busyId = av.id;
    try {
      await deactivateTenantAvatar(av.id);
      feedback.success(`${av.name} moved to library`);
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Deactivate failed');
    } finally {
      busyId = null;
    }
  }

  async function onVoiceChange(av: TenantAvatar, e: Event) {
    if (!isOwned(av)) return;
    const select = e.currentTarget as HTMLSelectElement;
    const next = select.value;
    if (!next || next === av.voice) return;
    voiceBusyId = av.id;
    try {
      await updateTenantAvatar(av.id, { voice: next });
      feedback.success(`${av.name} voice → ${next}`);
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Voice update failed');
      select.value = av.voice || 'Aoede';
    } finally {
      voiceBusyId = null;
    }
  }

  function voiceLabel(name: string | undefined) {
    if (!name) return '—';
    const hit = speakerVoices.find((v) => v.name === name);
    return hit?.label ?? name;
  }

  async function onArchive(av: TenantAvatar) {
    if (!av.owner_tenant_id) {
      feedback.error('Platform catalog avatars cannot be archived here');
      return;
    }
    if (!confirm(`Archive ${av.name}? It will leave the workforce.`)) return;
    busyId = av.id;
    try {
      await archiveTenantAvatar(av.id);
      feedback.success(`${av.name} archived`);
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Archive failed');
    } finally {
      busyId = null;
    }
  }

  function barPct() {
    if (!cap.limit) return 0;
    return Math.min(100, Math.round((cap.active / cap.limit) * 100));
  }

  function isActive(av: TenantAvatar) {
    return av.assignment_status === 'active';
  }

  function isOwned(av: TenantAvatar) {
    return Boolean(av.owner_tenant_id);
  }

  function portraitSrc(av: TenantAvatar) {
    const url = (av.image_url || '').trim();
    if (!url || url.includes('default-avatar')) return '';
    return url;
  }

  function themedPortraitSrc(av: TenantAvatar, variant: AvatarImageVariant) {
    const fromTopLevel =
      variant === 'dark'
        ? av.image_dark_url
        : variant === 'light'
          ? av.image_light_url
          : av.image_url;
    const fromFlags =
      variant === 'dark'
        ? av.flags?.image_dark_url
        : variant === 'light'
          ? av.flags?.image_light_url
          : av.image_url;
    const url = String(fromTopLevel || fromFlags || '').trim();
    if (!url || url.includes('default-avatar')) return '';
    return url;
  }

  function variantLabel(variant: AvatarImageVariant) {
    return variant === 'default' ? 'Default' : variant[0].toUpperCase() + variant.slice(1);
  }
</script>

<div class="page-head">
  <div>
    <h1 style="margin:0;font-size:24px">AI avatars</h1>
    <p style="margin:6px 0 0;color:var(--muted);font-size:14px">
      Create unlimited library agents with portraits. Only active agents count toward your package limit.
    </p>
  </div>
  <button class="btn" type="button" onclick={() => (showCreate = true)}>+ Create avatar</button>
</div>

{#if loading}
  <p style="color:var(--muted)">Loading…</p>
{:else}
  <div class="card" style="margin-bottom:16px">
    <div style="display:flex;justify-content:space-between;align-items:baseline;gap:12px;flex-wrap:wrap">
      <div>
        <strong>Active workforce</strong>
        <span style="color:var(--muted);margin-left:8px">{cap.active} / {cap.limit}</span>
      </div>
      <span style="font-size:13px;color:var(--muted)"
        >{cap.remaining} slot{cap.remaining === 1 ? '' : 's'} remaining</span
      >
    </div>
    <div class="bar" style="margin-top:10px">
      <div class="fill" style="width:{barPct()}%"></div>
    </div>
    <p style="margin:10px 0 0;font-size:12px;color:var(--muted)">
      Package limit applies to <strong>active</strong> agents only — you can create as many drafts as you need.
    </p>
  </div>

  <div class="card">
    <h2 style="margin:0 0 12px;font-size:16px">Library</h2>
    {#if avatars.length === 0}
      <p style="color:var(--muted);margin:0">
        No avatars yet. Create one or ask platform to assign catalog agents.
      </p>
    {:else}
      <div style="overflow-x:auto">
        <table class="data-table">
          <thead>
            <tr>
              <th>Avatar</th>
              <th>Role</th>
              <th>Voice</th>
              <th>Status</th>
              <th>Owner</th>
              <th>Portrait</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {#each avatars as av (av.id)}
              <tr>
                <td>
                  <div style="display:flex;align-items:center;gap:10px">
                    {#if portraitSrc(av)}
                      <img class="row-portrait" src={portraitSrc(av)} alt="" />
                    {:else}
                      <span class="swatch" style="background:{av.color || '#334155'}"></span>
                    {/if}
                    <div>
                      <strong>{av.name}</strong>
                      <div style="font-size:12px;color:var(--muted)">{av.slug}</div>
                    </div>
                  </div>
                </td>
                <td>{av.role || '—'}</td>
                <td>
                  {#if isOwned(av) && speakerVoices.length}
                    <select
                      class="voice-select"
                      value={av.voice || 'Aoede'}
                      disabled={voiceBusyId === av.id}
                      onchange={(e) => onVoiceChange(av, e)}
                      aria-label="Speaker voice for {av.name}"
                    >
                      {#each speakerVoices as v (v.name)}
                        <option value={v.name}>{v.label}</option>
                      {/each}
                    </select>
                  {:else}
                    <span style="font-size:13px;color:var(--muted)">{voiceLabel(av.voice)}</span>
                  {/if}
                </td>
                <td>
                  {#if isActive(av)}
                    <span class="badge success">Active</span>
                  {:else}
                    <span class="badge">Library</span>
                  {/if}
                </td>
                <td style="font-size:13px;color:var(--muted)">
                  {isOwned(av) ? 'Tenant' : 'Platform'}
                </td>
                <td>
                  {#if isOwned(av)}
                    <div class="portrait-upload-set">
                      {#each portraitVariants as variant (variant)}
                        <label class="upload-inline">
                          <input
                            type="file"
                            accept={AVATAR_IMAGE_HINT.accept}
                            disabled={uploadingId === av.id}
                            onchange={(e) => onRowImageUpload(av, e, variant)}
                          />
                          <span class="btn ghost small">
                            {uploadingId === av.id ? 'Uploading…' : variantLabel(variant)}
                          </span>
                        </label>
                      {/each}
                      <div class="portrait-variant-status" aria-label="Portrait variants">
                        <span class:ready={Boolean(portraitSrc(av))}>Default</span>
                        <span class:ready={Boolean(themedPortraitSrc(av, 'dark'))}>Dark</span>
                        <span class:ready={Boolean(themedPortraitSrc(av, 'light'))}>Light</span>
                      </div>
                    </div>
                  {:else}
                    <span style="font-size:12px;color:var(--muted)">—</span>
                  {/if}
                </td>
                <td>
                  <div style="display:flex;gap:8px;flex-wrap:wrap">
                    {#if isActive(av)}
                      <button
                        class="btn ghost"
                        type="button"
                        disabled={busyId === av.id}
                        onclick={() => onDeactivate(av)}
                      >
                        Deactivate
                      </button>
                    {:else}
                      <button
                        class="btn"
                        type="button"
                        disabled={busyId === av.id || cap.remaining <= 0}
                        title={cap.remaining <= 0 ? 'Active package limit reached' : 'Activate'}
                        onclick={() => onActivate(av)}
                      >
                        Activate
                      </button>
                    {/if}
                    {#if isOwned(av)}
                      <button
                        class="btn ghost"
                        type="button"
                        disabled={busyId === av.id}
                        onclick={() => onArchive(av)}
                      >
                        Archive
                      </button>
                    {/if}
                  </div>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
      <p class="image-spec" style="margin-top:12px">
        <strong>Portrait specs:</strong>
        {AVATAR_IMAGE_HINT.acceptLabel} · max {AVATAR_IMAGE_HINT.maxLabel} · recommended
        {AVATAR_IMAGE_HINT.recommendSize} ({AVATAR_IMAGE_HINT.aspect}). {AVATAR_IMAGE_HINT.recommendNote}
      </p>
    {/if}
  </div>
{/if}

{#if showCreate}
  <div class="modal-backdrop" role="presentation">
    <button
      class="modal-scrim"
      type="button"
      aria-label="Close create avatar"
      onclick={() => {
        showCreate = false;
        resetCreateForm();
      }}
    ></button>
    <div class="card modal" role="dialog" tabindex="-1">
      <h3 style="margin:0 0 8px">Create avatar</h3>
      <p style="margin:0 0 16px;font-size:13px;color:var(--muted)">
        Adds to your library as inactive. Activate later within your package limit.
      </p>
      <div class="field">
        <label for="av-name">Name *</label>
        <input id="av-name" type="text" bind:value={name} placeholder="Maya" />
      </div>
      <div class="field">
        <label for="av-role">Role</label>
        <input id="av-role" type="text" bind:value={role} placeholder="Sales assistant" />
      </div>
      <div class="field">
        <label for="av-trait">Trait</label>
        <input id="av-trait" type="text" bind:value={trait} placeholder="Friendly" />
      </div>
      <div class="field">
        <label for="av-greeting">Greeting</label>
        <input id="av-greeting" type="text" bind:value={greeting} placeholder="Hi, how can I help?" />
      </div>
      <div class="field">
        <label for="av-color">Color</label>
        <input id="av-color" type="color" bind:value={color} />
      </div>

      <div class="field">
        <label for="av-voice">Speaker voice</label>
        <select id="av-voice" bind:value={voice}>
          {#each speakerVoices as v (v.name)}
            <option value={v.name}>{v.label}</option>
          {/each}
        </select>
        <p class="image-spec" style="margin-top:6px">
          Gemini AI Studio speaker settings from
          <a href={voiceSource} target="_blank" rel="noopener noreferrer">generate-speech</a>
          (name + style, e.g. <em>Aoede — Breezy</em>).
        </p>
      </div>

      <div class="field">
        <label for="av-portrait">Portrait image</label>
        <div class="portrait-row">
          {#if pendingPreview}
            <div class="portrait-preview">
              <img src={pendingPreview} alt="Portrait preview" />
            </div>
          {:else}
            <div class="portrait-preview placeholder" aria-hidden="true">No image</div>
          {/if}
          <div class="portrait-controls">
            <input
              id="av-portrait"
              type="file"
              accept={AVATAR_IMAGE_HINT.accept}
              onchange={onPendingFileChange}
            />
            {#if pendingFile}
              <button class="btn ghost small" type="button" onclick={clearPendingFile}>Clear</button>
            {/if}
            <p class="image-spec">
              <strong>File types:</strong> {AVATAR_IMAGE_HINT.acceptLabel}<br />
              <strong>Max size:</strong> {AVATAR_IMAGE_HINT.maxLabel}<br />
              <strong>Recommended size:</strong> {AVATAR_IMAGE_HINT.recommendSize} square ({AVATAR_IMAGE_HINT.aspect})<br />
              <span>{AVATAR_IMAGE_HINT.recommendNote}</span>
            </p>
          </div>
        </div>
      </div>

      <div style="display:flex;gap:10px;justify-content:flex-end;margin-top:16px">
        <button
          class="btn ghost"
          type="button"
          onclick={() => {
            showCreate = false;
            resetCreateForm();
          }}>Cancel</button
        >
        <button class="btn" type="button" disabled={creating || !name.trim()} onclick={onCreate}>
          {creating ? 'Creating…' : 'Create — library inactive'}
        </button>
      </div>
    </div>
  </div>
{/if}

<style>
  .page-head {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 16px;
    margin-bottom: 20px;
    flex-wrap: wrap;
  }
  .bar {
    height: 8px;
    border-radius: 999px;
    background: var(--border, #1e293b);
    overflow: hidden;
  }
  .fill {
    height: 100%;
    background: linear-gradient(90deg, #38bdf8, #818cf8);
    border-radius: 999px;
  }
  .data-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 14px;
  }
  .data-table th {
    text-align: left;
    color: var(--muted);
    font-weight: 500;
    padding: 8px;
    border-bottom: 1px solid var(--border, #1e293b);
  }
  .data-table td {
    padding: 10px 8px;
    border-bottom: 1px solid var(--border, #1e293b);
    vertical-align: middle;
  }
  .swatch {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    display: inline-block;
    flex-shrink: 0;
  }
  .row-portrait {
    width: 40px;
    height: 40px;
    border-radius: 50%;
    object-fit: cover;
    border: 1px solid rgb(70 132 190 / 30%);
    flex-shrink: 0;
  }
  .badge {
    display: inline-block;
    padding: 2px 8px;
    border-radius: 999px;
    font-size: 12px;
    background: var(--surface-2, #0f172a);
    color: var(--muted);
  }
  .badge.success {
    background: rgba(34, 197, 94, 0.15);
    color: #4ade80;
  }
  .modal-backdrop {
    position: fixed;
    inset: 0;
    background: rgba(0, 0, 0, 0.55);
    display: grid;
    place-items: center;
    z-index: 50;
    padding: 16px;
  }
  .modal-scrim {
    position: absolute;
    inset: 0;
    border: 0;
    background: transparent;
    cursor: default;
  }
  .modal {
    position: relative;
    z-index: 1;
    width: min(480px, 100%);
    max-height: 90vh;
    overflow: auto;
  }
  .portrait-row {
    display: flex;
    gap: 14px;
    align-items: flex-start;
    flex-wrap: wrap;
  }
  .portrait-preview {
    width: 96px;
    height: 96px;
    border-radius: 50%;
    overflow: hidden;
    border: 2px solid rgb(0 183 255 / 35%);
    flex-shrink: 0;
    background: rgb(3 11 23 / 88%);
  }
  .portrait-preview img {
    width: 100%;
    height: 100%;
    object-fit: cover;
    display: block;
  }
  .portrait-preview.placeholder {
    display: grid;
    place-items: center;
    font-size: 11px;
    color: var(--muted);
    text-align: center;
    padding: 8px;
  }
  .portrait-controls {
    flex: 1;
    min-width: 200px;
    display: grid;
    gap: 8px;
  }
  .portrait-controls input[type='file'] {
    font-size: 13px;
    color: var(--muted);
  }
  .image-spec {
    margin: 0;
    font-size: 12px;
    color: var(--muted);
    line-height: 1.45;
  }
  .upload-inline {
    display: inline-flex;
    cursor: pointer;
  }
  .upload-inline input[type='file'] {
    position: absolute;
    width: 1px;
    height: 1px;
    opacity: 0;
    overflow: hidden;
  }
  .btn.small {
    font-size: 12px;
    padding: 6px 10px;
  }
  .portrait-upload-set {
    display: grid;
    gap: 6px;
    min-width: 160px;
  }
  .portrait-upload-set .upload-inline {
    margin-right: 4px;
  }
  .portrait-variant-status {
    display: flex;
    gap: 4px;
    flex-wrap: wrap;
  }
  .portrait-variant-status span {
    border: 1px solid rgb(70 132 190 / 24%);
    border-radius: 999px;
    color: var(--muted);
    font-size: 11px;
    line-height: 1;
    padding: 4px 6px;
  }
  .portrait-variant-status span.ready {
    border-color: rgb(34 197 94 / 35%);
    background: rgb(34 197 94 / 12%);
    color: #4ade80;
  }
  .voice-select,
  select#av-voice {
    width: 100%;
    max-width: 220px;
    border: 1px solid rgb(70 132 190 / 24%);
    border-radius: 10px;
    background: rgb(3 11 23 / 88%);
    color: var(--ink);
    padding: 8px 10px;
    font-size: 13px;
  }
  select#av-voice {
    max-width: none;
  }
</style>
