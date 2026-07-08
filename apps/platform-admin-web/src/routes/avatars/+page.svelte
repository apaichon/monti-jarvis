<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import { archiveAvatar, listAvatars, type Avatar } from '$lib/api/avatars';
  import { ApiError } from '$lib/api/http';
  import { clearSession, loginPath } from '$lib/auth/session';

  let avatars = $state<Avatar[]>([]);
  let statusFilter = $state('active');
  let error = $state('');
  let loading = $state(true);
  let archiveTarget = $state<Avatar | null>(null);
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

  async function load() {
    loading = true;
    error = '';
    try {
      const res = await listAvatars(statusFilter);
      avatars = res.avatars;
    } catch (err) {
      error = handleError(err, 'Failed to load avatars');
    } finally {
      loading = false;
    }
  }

  onMount(load);

  async function confirmArchive() {
    if (!archiveTarget) return;
    archiving = true;
    try {
      await archiveAvatar(archiveTarget.id);
      archiveTarget = null;
      await load();
    } catch (err) {
      error = handleError(err, 'Archive failed');
    } finally {
      archiving = false;
    }
  }
</script>

<div style="display:flex;justify-content:space-between;align-items:center;gap:12px;margin-bottom:20px;flex-wrap:wrap">
  <h1 style="margin:0;font-size:24px">Avatars</h1>
  <div style="display:flex;gap:10px;align-items:center">
    <label style="font-size:13px;color:var(--muted)">
      Status
      <select bind:value={statusFilter} onchange={load} style="margin-left:8px">
        <option value="">all</option>
        <option value="active">active</option>
        <option value="draft">draft</option>
        <option value="archived">archived</option>
      </select>
    </label>
    <a class="btn" href="{base}/avatars/new">+ New</a>
  </div>
</div>

{#if error}
  <p class="error" style="margin-bottom:12px">{error}</p>
{/if}

<div class="card">
  {#if loading}
    <p style="color:var(--muted)">Loading…</p>
  {:else if avatars.length === 0}
    <p style="color:var(--muted)">No avatars found. <a class="link" href="{base}/avatars/new">Create one</a></p>
  {:else}
    <table class="table">
      <thead>
        <tr>
          <th>slug</th>
          <th>name</th>
          <th>role</th>
          <th>status</th>
          <th>actions</th>
        </tr>
      </thead>
      <tbody>
        {#each avatars as avatar (avatar.id)}
          <tr>
            <td>{avatar.slug}</td>
            <td>{avatar.name}</td>
            <td>{avatar.role}</td>
            <td><span class="badge">{avatar.status}</span></td>
            <td>
              <div class="row-actions">
                <a class="link" href="{base}/avatars/{avatar.id}">Edit</a>
                <a class="link" href="{base}/tenants/demo/avatars">Assign demo</a>
                {#if avatar.status !== 'archived'}
                  <button
                    class="link"
                    type="button"
                    style="background:none;border:none;padding:0;color:var(--danger)"
                    onclick={() => (archiveTarget = avatar)}
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
  {/if}
</div>

{#if archiveTarget}
  <div class="modal-backdrop" role="presentation" onclick={() => (archiveTarget = null)}>
    <div class="card modal" role="dialog" onclick={(e) => e.stopPropagation()}>
      <h3 style="margin:0 0 12px">Archive "{archiveTarget.name}"?</h3>
      <p style="color:var(--muted);font-size:14px">
        Active tenant assignments block archive (409).
      </p>
      <div style="display:flex;gap:10px;justify-content:flex-end;margin-top:16px">
        <button class="btn ghost" type="button" onclick={() => (archiveTarget = null)}>Cancel</button>
        <button class="btn danger" type="button" disabled={archiving} onclick={confirmArchive}>
          {archiving ? 'Archiving…' : 'Archive'}
        </button>
      </div>
    </div>
  </div>
{/if}