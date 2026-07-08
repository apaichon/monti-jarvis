<script lang="ts">
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import AvatarImageField from '$lib/components/AvatarImageField.svelte';
  import StatusMessage from '$lib/components/StatusMessage.svelte';
  import VoiceProfilesForm from '$lib/components/VoiceProfilesForm.svelte';
  import {
    createAvatar,
    defaultVoiceRow,
    type AvatarFlags,
    type AvatarVoice
  } from '$lib/api/avatars';
  import { ApiError } from '$lib/api/http';
  import { clearSession, loginPath } from '$lib/auth/session';

  let slug = $state('');
  let name = $state('');
  let role = $state('');
  let trait = $state('');
  let color = $state('#008cff');
  let imageUrl = $state('');
  let greeting = $state('');
  let status = $state('draft');
  let popular = $state(false);
  let robot = $state(false);
  let skin = $state('');
  let hair = $state('');
  let voices = $state<AvatarVoice[]>([defaultVoiceRow()]);
  let error = $state('');
  let saving = $state(false);

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

  function buildFlags(): AvatarFlags {
    const flags: AvatarFlags = {};
    if (popular) flags.popular = true;
    if (robot) flags.robot = true;
    if (skin) flags.skin = skin;
    if (hair) flags.hair = hair;
    return flags;
  }

  async function submit(e: Event) {
    e.preventDefault();
    if (voices.length === 0) {
      error = 'At least one voice profile is required';
      return;
    }
    error = '';
    saving = true;
    try {
      const created = await createAvatar({
        slug,
        name,
        role,
        trait,
        color,
        image_url: imageUrl || `/images/${slug}.jpg`,
        greeting,
        status,
        flags: buildFlags(),
        voices
      });
      goto(`${base}/avatars/${created.id}`);
    } catch (err) {
      error = handleError(err, 'Create failed');
    } finally {
      saving = false;
    }
  }
</script>

<p><a class="link" href="{base}/avatars">← Avatars</a></p>
<h1 style="margin:8px 0 20px;font-size:24px">New avatar</h1>

<StatusMessage message={error} variant="error" />

<form onsubmit={submit}>
  <div class="card" style="margin-bottom:16px">
    <h2 style="margin:0 0 16px;font-size:16px">Metadata</h2>
    <div class="field">
      <label for="slug">Slug *</label>
      <input id="slug" bind:value={slug} required placeholder="ava" />
    </div>
    <div class="field">
      <label for="name">Name *</label>
      <input id="name" bind:value={name} required />
    </div>
    <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px">
      <div class="field">
        <label for="role">Role *</label>
        <input id="role" bind:value={role} required placeholder="General Support" />
      </div>
      <div class="field">
        <label for="trait">Trait</label>
        <input id="trait" bind:value={trait} placeholder="Warm & Patient" />
      </div>
    </div>
    <div style="display:grid;grid-template-columns:1fr 1fr;gap:12px">
      <div class="field">
        <label for="color">Color</label>
        <input id="color" bind:value={color} placeholder="#008cff" />
      </div>
      <div class="field">
        <label for="status">Status</label>
        <select id="status" bind:value={status}>
          <option value="draft">draft</option>
          <option value="active">active</option>
        </select>
      </div>
    </div>
    <AvatarImageField
      avatarId={slug}
      bind:imageUrl={imageUrl}
      onError={(msg) => {
        error = msg;
      }}
    />
    <div class="field">
      <label for="greeting">Greeting *</label>
      <textarea id="greeting" rows="3" bind:value={greeting} required></textarea>
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
        <input id="skin" bind:value={skin} placeholder="#f0bd9b" />
      </div>
      <div class="field">
        <label for="hair">Hair color</label>
        <input id="hair" bind:value={hair} placeholder="#5a3428" />
      </div>
    </div>
  </div>

  <div class="card" style="margin-bottom:16px">
    <VoiceProfilesForm bind:voices />
  </div>

  <div style="display:flex;gap:10px;justify-content:flex-end">
    <a class="btn ghost" href="{base}/avatars">Cancel</a>
    <button class="btn" type="submit" disabled={saving || voices.length === 0}>
      {saving ? 'Creating…' : 'Create avatar'}
    </button>
  </div>
</form>