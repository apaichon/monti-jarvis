<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import AvatarImageField from '$lib/components/AvatarImageField.svelte';
  import VoiceProfilesForm from '$lib/components/VoiceProfilesForm.svelte';
  import { feedback } from '$lib/feedback.svelte';
  import {
    createAvatar,
    assignTenantAvatar,
    defaultVoiceRow,
    type AvatarFlags,
    type AvatarVoice
  } from '$lib/api/avatars';
  import { listTenants, type TenantListItem } from '$lib/api/tenants';
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
  let tenants = $state<TenantListItem[]>([]);
  let selectedTenantIds = $state<string[]>([]);
  let tenantsLoading = $state(true);
  let saving = $state(false);

  onMount(async () => {
    try {
      const result = await listTenants('', '', 100, 0);
      tenants = result.tenants;
      selectedTenantIds = tenants.map((tenant) => tenant.id);
    } catch (err) {
      const msg = handleError(err, 'Failed to load tenants');
      if (msg) feedback.error(msg);
    } finally {
      tenantsLoading = false;
    }
  });

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

  function toggleTenant(tenantId: string) {
    selectedTenantIds = selectedTenantIds.includes(tenantId)
      ? selectedTenantIds.filter((id) => id !== tenantId)
      : [...selectedTenantIds, tenantId];
  }

  async function submit(e: Event) {
    e.preventDefault();
    if (voices.length === 0) {
      feedback.error('At least one voice profile is required');
      return;
    }
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
      const assignments = await Promise.allSettled(
        selectedTenantIds.map((tenantId) => assignTenantAvatar(tenantId, created.id))
      );
      const failed = assignments.filter((result) => result.status === 'rejected');
      if (failed.length > 0) {
        feedback.error(`Avatar created, but ${failed.length} tenant assignment${failed.length === 1 ? '' : 's'} failed. Check package caps.`);
      } else if (selectedTenantIds.length > 0) {
        feedback.success(`Avatar created and assigned to ${selectedTenantIds.length} tenant${selectedTenantIds.length === 1 ? '' : 's'}`);
      }
      goto(`${base}/avatars/${created.id}`);
    } catch (err) {
      const msg = handleError(err, 'Create failed');
      if (msg) feedback.error(msg);
    } finally {
      saving = false;
    }
  }
</script>

<p><a class="link" href="{base}/avatars">← Avatars</a></p>
<h1 style="margin:8px 0 20px;font-size:24px">New avatar</h1>

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
    <AvatarImageField avatarId={slug} bind:imageUrl={imageUrl} />
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
    <h2 style="margin:0 0 6px;font-size:16px">Tenant assignments</h2>
    <p style="margin:0 0 14px;color:var(--muted);font-size:13px">
      Optional. Select the tenants that should receive this AI employee immediately after creation.
    </p>
    {#if tenantsLoading}
      <p style="color:var(--muted);font-size:13px">Loading tenants…</p>
    {:else if tenants.length === 0}
      <p style="color:var(--muted);font-size:13px">No tenants available.</p>
    {:else}
      <div style="display:grid;gap:8px;max-height:260px;overflow:auto">
        {#each tenants as tenant (tenant.id)}
          <label style="display:flex;align-items:center;gap:10px;padding:10px;border:1px solid var(--line);border-radius:8px;cursor:pointer">
            <input
              type="checkbox"
              checked={selectedTenantIds.includes(tenant.id)}
              onchange={() => toggleTenant(tenant.id)}
            />
            <span>
              <strong>{tenant.name || tenant.slug}</strong>
              <small style="display:block;color:var(--muted)">{tenant.slug} · {tenant.status}</small>
            </span>
          </label>
        {/each}
      </div>
      <p style="margin:10px 0 0;color:var(--muted);font-size:12px">{selectedTenantIds.length} selected</p>
    {/if}
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
