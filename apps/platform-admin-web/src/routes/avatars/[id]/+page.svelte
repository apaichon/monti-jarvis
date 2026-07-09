<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import AvatarImageField from '$lib/components/AvatarImageField.svelte';
  import VoiceProfilesForm from '$lib/components/VoiceProfilesForm.svelte';
  import { feedback } from '$lib/feedback.svelte';
  import {
    archiveAvatar,
    getAvatar,
    updateAvatar,
    type Avatar,
    type AvatarFlags,
    type AvatarVoice
  } from '$lib/api/avatars';
  import { ApiError } from '$lib/api/http';
  import { clearSession, loginPath } from '$lib/auth/session';

  const id = $derived($page.params.id);

  let avatar = $state<Avatar | null>(null);
  let voices = $state<AvatarVoice[]>([]);
  let popular = $state(false);
  let robot = $state(false);
  let skin = $state('');
  let hair = $state('');
  let saving = $state(false);
  let showArchive = $state(false);
  let archiving = $state(false);

  function handleError(err: unknown, fallback: string) {
    if (err instanceof ApiError) {
      if (err.status === 401) {
        clearSession();
        goto(loginPath($page.url.pathname));
        return '';
      }
      return err.message;
    }
    return fallback;
  }

  function applyFlags(flags: AvatarFlags = {}) {
    popular = !!flags.popular;
    robot = !!flags.robot;
    skin = flags.skin ?? '';
    hair = flags.hair ?? '';
  }

  function buildFlags(): AvatarFlags {
    const flags: AvatarFlags = {};
    if (popular) flags.popular = true;
    if (robot) flags.robot = true;
    if (skin) flags.skin = skin;
    if (hair) flags.hair = hair;
    return flags;
  }

  onMount(async () => {
    try {
      const av = await getAvatar(id);
      avatar = av;
      voices = av.voices.map((v) => ({ ...v }));
      applyFlags(av.flags);
    } catch (err) {
      const msg = handleError(err, 'Failed to load avatar');
      if (msg) feedback.error(msg);
    }
  });

  async function save(e: Event) {
    e.preventDefault();
    if (!avatar) return;
    if (voices.length === 0) {
      feedback.error('At least one voice profile is required');
      return;
    }
    saving = true;
    try {
      avatar = await updateAvatar(avatar.id, {
        slug: avatar.slug,
        name: avatar.name,
        role: avatar.role,
        trait: avatar.trait,
        color: avatar.color,
        image_url: avatar.image_url,
        greeting: avatar.greeting,
        status: avatar.status,
        flags: buildFlags(),
        voices
      });
      voices = avatar.voices.map((v) => ({ ...v }));
      applyFlags(avatar.flags);
      feedback.success(`Saved — ${avatar.name} is now ${avatar.status}.`);
    } catch (err) {
      const msg = handleError(err, 'Save failed');
      if (msg) feedback.error(msg);
    } finally {
      saving = false;
    }
  }

  async function confirmArchive() {
    if (!avatar) return;
    archiving = true;
    try {
      await archiveAvatar(avatar.id);
      goto(`${base}/avatars`);
    } catch (err) {
      const msg = handleError(err, 'Archive failed');
      if (msg) feedback.error(msg);
      showArchive = false;
    } finally {
      archiving = false;
    }
  }
</script>

<p><a class="link" href="{base}/avatars">← Avatars</a></p>

{#if avatar}
  <h1 style="margin:8px 0 20px;font-size:24px">Edit: {avatar.name}</h1>
{:else}
  <h1 style="margin:8px 0 20px;font-size:24px">Edit avatar</h1>
{/if}

{#if avatar}
  <form onsubmit={save}>
    <div class="card" style="margin-bottom:16px">
      <h2 style="margin:0 0 16px;font-size:16px">Metadata</h2>
      <div class="field">
        <label for="slug">Slug</label>
        <input id="slug" bind:value={avatar.slug} required />
      </div>
      <div class="field">
        <label for="name">Name</label>
        <input id="name" bind:value={avatar.name} required />
      </div>
      <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px">
        <div class="field">
          <label for="role">Role</label>
          <input id="role" bind:value={avatar.role} required />
        </div>
        <div class="field">
          <label for="trait">Trait</label>
          <input id="trait" bind:value={avatar.trait} />
        </div>
      </div>
      <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px">
        <div class="field">
          <label for="color">Color</label>
          <input id="color" bind:value={avatar.color} />
        </div>
        <div class="field">
          <label for="status">Status</label>
          <select id="status" bind:value={avatar.status}>
            <option value="draft">draft</option>
            <option value="active">active</option>
            <option value="archived">archived</option>
          </select>
        </div>
      </div>
      <AvatarImageField avatarId={avatar.id} bind:imageUrl={avatar.image_url} />
      <div class="field">
        <label for="greeting">Greeting</label>
        <textarea id="greeting" rows="3" bind:value={avatar.greeting} required></textarea>
      </div>
      <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px">
        <div class="field">
          <label style="display:flex;align-items:center;gap:8px;cursor:pointer">
            <input type="checkbox" bind:checked={popular} />
            Popular
          </label>
        </div>
        <div class="field">
          <label style="display:flex;align-items:center;gap:8px;cursor:pointer">
            <input type="checkbox" bind:checked={robot} />
            Robot
          </label>
        </div>
      </div>
      <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px">
        <div class="field">
          <label for="skin">Skin color</label>
          <input id="skin" bind:value={skin} />
        </div>
        <div class="field">
          <label for="hair">Hair color</label>
          <input id="hair" bind:value={hair} />
        </div>
      </div>
    </div>

    <div class="card" style="margin-bottom:16px">
      <VoiceProfilesForm bind:voices />
    </div>

    <div style="display:flex;gap:10px;justify-content:space-between;flex-wrap:wrap">
      <button
        class="btn danger"
        type="button"
        disabled={avatar.status === 'archived'}
        onclick={() => (showArchive = true)}
      >
        Archive avatar
      </button>
      <button class="btn" type="submit" disabled={saving || voices.length === 0}>
        {saving ? 'Saving…' : 'Save changes'}
      </button>
    </div>
  </form>
{/if}

{#if showArchive}
  <div class="modal-backdrop" role="presentation" onclick={() => (showArchive = false)}>
    <div class="card modal" role="dialog" onclick={(e) => e.stopPropagation()}>
      <h3 style="margin:0 0 12px">Archive "{avatar?.name}"?</h3>
      <p style="color:var(--muted);font-size:14px">Active tenant assignments block archive (409).</p>
      <div style="display:flex;gap:10px;justify-content:flex-end;margin-top:16px">
        <button class="btn ghost" type="button" onclick={() => (showArchive = false)}>Cancel</button>
        <button class="btn danger" type="button" disabled={archiving} onclick={confirmArchive}>
          {archiving ? 'Archiving…' : 'Archive'}
        </button>
      </div>
    </div>
  </div>
{/if}