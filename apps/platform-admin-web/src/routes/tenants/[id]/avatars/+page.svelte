<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { page } from '$app/stores';
  import {
    assignTenantAvatar,
    listAvatars,
    listTenantAvatars,
    revokeTenantAvatar,
    type Avatar,
    type TenantAvatarAssignment,
    type TenantAssignmentsResponse
  } from '$lib/api/avatars';
  import { ApiError } from '$lib/api/http';
  import { clearSession, loginPath } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';

  const tenantId = $derived($page.params.id ?? '');

  let data = $state<TenantAssignmentsResponse | null>(null);
  let catalog = $state<Avatar[]>([]);
  let selectedAvatar = $state('');
  let loading = $state(true);
  let assigning = $state(false);
  let promotingId = $state<string | null>(null);
  let revokingId = $state<string | null>(null);
  let revokeTarget = $state<TenantAvatarAssignment | null>(null);

  const capOver = $derived(
    data ? data.cap.active_count >= data.cap.max_ai_employees : false
  );
  const capBlocksAssignment = $derived(capOver && !data?.cap.override_allowed);

  const assignableAvatars = $derived.by(() => {
    if (!data) return catalog;
    const assigned = new Set(data.assignments.map((a) => a.avatar_id));
    return catalog.filter((a) => a.status === 'active' && !assigned.has(a.id));
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

  function showError(err: unknown, fallback: string) {
    const msg = handleError(err, fallback);
    if (msg) feedback.error(msg);
  }

  async function load() {
    loading = true;
    try {
      const [tenantRes, catalogRes] = await Promise.all([
        listTenantAvatars(tenantId),
        listAvatars('active')
      ]);
      data = tenantRes;
      catalog = catalogRes.avatars;
      if (!selectedAvatar && assignableAvatars[0]) selectedAvatar = assignableAvatars[0].id;
    } catch (err) {
      showError(err, 'Failed to load tenant avatars');
    } finally {
      loading = false;
    }
  }

  onMount(load);

  async function assign() {
    if (!selectedAvatar) return;
    assigning = true;
    try {
      await assignTenantAvatar(tenantId, selectedAvatar);
      selectedAvatar = '';
      await load();
      feedback.success('Avatar assigned to tenant');
    } catch (err) {
      showError(err, 'Assign failed');
    } finally {
      assigning = false;
    }
  }

  async function promote(assignment: TenantAvatarAssignment) {
    promotingId = assignment.avatar_id;
    try {
      await assignTenantAvatar(tenantId, assignment.avatar_id);
      await load();
      feedback.success(`${assignment.avatar?.name ?? assignment.avatar_id} promoted`);
    } catch (err) {
      showError(err, 'Promote failed');
    } finally {
      promotingId = null;
    }
  }

  async function revoke() {
    if (!revokeTarget) return;
    revokingId = revokeTarget.avatar_id;
    try {
      await revokeTenantAvatar(tenantId, revokeTarget.avatar_id);
      revokeTarget = null;
      await load();
      feedback.success('Avatar demoted');
    } catch (err) {
      showError(err, 'Demote failed');
    } finally {
      revokingId = null;
    }
  }
</script>

<p><a class="link" href="{base}/avatars">← Avatars</a></p>
<h1 style="margin:0 0 20px;font-size:24px">Tenant avatars — {tenantId}</h1>

{#if loading}
  <p style="color:var(--muted)">Loading…</p>
{:else if data}
  <div
    class="card"
    style="margin-bottom:16px;{capBlocksAssignment ? 'border-color: rgb(255 92 122 / 45%)' : ''}"
  >
    <h2 style="margin:0 0 8px;font-size:16px">Entitlement cap</h2>
    <p style="margin:0;font-size:14px;color:var(--muted)">
      {data.cap.active_count} assigned · {data.cap.max_ai_employees} max
      {#if capBlocksAssignment}
        <span class="error" style="margin-left:8px">
          At or over cap — new assignments return 409 until you demote one.
        </span>
      {:else if capOver && data.cap.override_allowed}
        <span style="margin-left:8px;color:var(--cyan)">
          Demo override — platform admins can still promote or assign avatars.
        </span>
      {/if}
    </p>
  </div>

  <div class="card" style="margin-bottom:16px">
    <h2 style="margin:0 0 16px;font-size:16px">Assigned</h2>
    {#if data.assignments.length === 0}
      <p style="color:var(--muted);margin:0">No avatar assignments for this tenant.</p>
    {:else}
      <table class="table">
        <thead>
          <tr>
            <th>id</th>
            <th>name</th>
            <th>role</th>
            <th>status</th>
            <th>actions</th>
          </tr>
        </thead>
        <tbody>
          {#each data.assignments as assignment (assignment.avatar_id)}
            <tr>
              <td>{assignment.avatar_id}</td>
              <td>{assignment.avatar?.name ?? '—'}</td>
              <td>{assignment.avatar?.role ?? '—'}</td>
              <td>
                <span class="badge" class:success={assignment.status === 'active'}>
                  {assignment.status}
                </span>
              </td>
              <td>
                {#if assignment.status === 'active'}
                  <button
                    class="link"
                    type="button"
                    style="background:none;border:none;padding:0;color:var(--danger)"
                    disabled={revokingId === assignment.avatar_id}
                    onclick={() => (revokeTarget = assignment)}
                  >
                    {revokingId === assignment.avatar_id ? 'Demoting…' : 'Demote'}
                  </button>
                {:else if assignment.avatar?.status === 'active'}
                  <button
                    class="link"
                    type="button"
                    style="background:none;border:none;padding:0"
                    disabled={promotingId === assignment.avatar_id || capBlocksAssignment}
                    onclick={() => promote(assignment)}
                  >
                    {promotingId === assignment.avatar_id ? 'Promoting…' : 'Promote'}
                  </button>
                {/if}
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    {/if}
  </div>

  <div class="card">
    <h2 style="margin:0 0 16px;font-size:16px">Assign avatar</h2>
    {#if assignableAvatars.length === 0}
      <p style="color:var(--muted);margin:0">No active catalog avatars available to assign.</p>
    {:else}
      <div class="field">
        <label for="avatar">Avatar</label>
        <select id="avatar" bind:value={selectedAvatar}>
          {#each assignableAvatars as avatar (avatar.id)}
            <option value={avatar.id}>{avatar.name} ({avatar.id})</option>
          {/each}
        </select>
      </div>
      <button
        class="btn"
        type="button"
        disabled={assigning || !selectedAvatar || capBlocksAssignment}
        onclick={assign}
      >
        {assigning ? 'Assigning…' : 'Assign to tenant'}
      </button>
      {#if capBlocksAssignment}
        <p style="color:var(--muted);font-size:13px;margin:12px 0 0">
          Demote an assignment before adding another when at cap.
        </p>
      {/if}
    {/if}
  </div>
{/if}

{#if revokeTarget}
  <div class="modal-backdrop" role="presentation" onclick={() => (revokeTarget = null)}>
    <div class="card modal" role="dialog" onclick={(e) => e.stopPropagation()}>
      <h3 style="margin:0 0 12px">Demote "{revokeTarget.avatar?.name ?? revokeTarget.avatar_id}"?</h3>
      <p style="color:var(--muted);font-size:14px">This disables the tenant assignment until it is promoted again.</p>
      <div style="display:flex;gap:10px;justify-content:flex-end;margin-top:16px">
        <button class="btn ghost" type="button" onclick={() => (revokeTarget = null)}>Cancel</button>
        <button class="btn danger" type="button" disabled={!!revokingId} onclick={revoke}>
          {revokingId ? 'Demoting…' : 'Demote'}
        </button>
      </div>
    </div>
  </div>
{/if}
