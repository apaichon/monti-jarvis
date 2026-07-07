<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { me } from '$lib/api/auth';
  import type { UserProfile } from '$lib/auth/session';
  import { clearSession } from '$lib/auth/session';

  let profile = $state<UserProfile | null>(null);
  let error = $state('');

  onMount(async () => {
    try {
      profile = await me();
    } catch {
      error = 'Failed to load profile';
      clearSession();
      goto(`${base}/login`);
    }
  });
</script>

<h1 style="margin:0 0 20px;font-size:24px">Profile</h1>

{#if error}
  <p class="error">{error}</p>
{:else if profile}
  <div class="card">
    <h2 style="margin:0 0 16px;font-size:16px">Account</h2>
    <div class="field">
      <label>Email</label>
      <div>{profile.email}</div>
    </div>
    <div class="field">
      <label>Display name</label>
      <div>{profile.display_name}</div>
    </div>
    <div class="field">
      <label>Role</label>
      <span class="badge">{profile.role}</span>
    </div>
    <div class="field">
      <label>Tenant</label>
      <div>{profile.tenant_id || '—'}</div>
    </div>
    <div class="field">
      <label>User ID</label>
      <div style="font-family:ui-monospace,monospace">{profile.id}</div>
    </div>
    <p style="color:var(--muted);font-size:13px;margin:0">
      Password change and MFA are not in Sprint 4 scope.
    </p>
  </div>
{/if}