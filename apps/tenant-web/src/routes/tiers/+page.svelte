<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';
  import { ApiError } from '$lib/api/http';
  import {
    listTiers,
    createTier,
    updateTier,
    deleteTier,
    listGroups,
    createGroup,
    deleteGroup,
    type CustomerTier,
    type CustomerGroup
  } from '$lib/api/tiers';

  let tiers = $state<CustomerTier[]>([]);
  let groups = $state<CustomerGroup[]>([]);
  let loading = $state(true);
  let saving = $state(false);

  let editingId = $state<string | null>(null);
  let name = $state('');
  let slug = $state('');
  let priority = $state(0);
  let description = $state('');
  let defaultAgent = $state('');
  let aiLocale = $state('');
  let maxPerCall = $state(0);
  let maxPerDay = $state(0);
  let active = $state(true);

  let groupName = $state('');
  let groupSlug = $state('');

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/tiers`)}`);
      return;
    }
    await load();
  });

  async function load() {
    loading = true;
    try {
      const [t, g] = await Promise.all([listTiers(), listGroups()]);
      tiers = t.tiers || [];
      groups = g.groups || [];
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load tiers');
    } finally {
      loading = false;
    }
  }

  function resetForm() {
    editingId = null;
    name = '';
    slug = '';
    priority = 0;
    description = '';
    defaultAgent = '';
    aiLocale = '';
    maxPerCall = 0;
    maxPerDay = 0;
    active = true;
  }

  function startEdit(t: CustomerTier) {
    editingId = t.id;
    name = t.name;
    slug = t.slug;
    priority = t.priority;
    description = t.description || '';
    defaultAgent = t.default_agent_id || '';
    aiLocale = t.ai_reply_locale || '';
    maxPerCall = t.max_minutes_per_call;
    maxPerDay = t.max_call_minutes_per_day;
    active = t.active;
  }

  async function saveTier() {
    if (!name.trim()) {
      feedback.error('Name is required');
      return;
    }
    saving = true;
    try {
      const body = {
        name: name.trim(),
        slug: slug.trim() || undefined,
        priority: Number(priority) || 0,
        description: description.trim(),
        default_agent_id: defaultAgent.trim(),
        ai_reply_locale: aiLocale,
        max_minutes_per_call: Number(maxPerCall) || 0,
        max_call_minutes_per_day: Number(maxPerDay) || 0,
        active
      };
      if (editingId) {
        await updateTier(editingId, body);
        feedback.success('Tier updated');
      } else {
        await createTier(body);
        feedback.success('Tier created');
      }
      resetForm();
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Save failed');
    } finally {
      saving = false;
    }
  }

  async function removeTier(t: CustomerTier) {
    if (!confirm(`Delete tier “${t.name}”?`)) return;
    try {
      await deleteTier(t.id);
      feedback.success('Tier deleted');
      if (editingId === t.id) resetForm();
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Delete failed');
    }
  }

  async function addGroup() {
    if (!groupName.trim()) {
      feedback.error('Group name is required');
      return;
    }
    try {
      await createGroup({
        name: groupName.trim(),
        slug: groupSlug.trim() || undefined
      });
      groupName = '';
      groupSlug = '';
      feedback.success('Group created');
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Create group failed');
    }
  }

  async function removeGroup(g: CustomerGroup) {
    if (!confirm(`Delete group “${g.name}”?`)) return;
    try {
      await deleteGroup(g.id);
      feedback.success('Group deleted');
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Delete failed');
    }
  }

  function capLabel(n: number) {
    return n > 0 ? String(n) : 'inherit';
  }
</script>

<div class="page-wrap">
  <div class="head">
    <div>
      <h1>Customer tiers</h1>
      <p class="sub">
        Define VIP / Standard rules for future customer accounts. Caps of 0 inherit tenant settings.
      </p>
    </div>
    <a class="link" href="{base}/preview">Test in Preview →</a>
  </div>

  {#if loading}
    <p class="dim">Loading…</p>
  {:else}
    <div class="grid">
      <section class="card">
        <h2>{editingId ? 'Edit tier' : 'New tier'}</h2>
        <label>
          <span>Name</span>
          <input type="text" bind:value={name} placeholder="VIP" />
        </label>
        <label>
          <span>Slug</span>
          <input type="text" bind:value={slug} placeholder="vip (auto from name)" />
        </label>
        <label>
          <span>Priority (higher = more VIP)</span>
          <input type="number" bind:value={priority} />
        </label>
        <label>
          <span>Description</span>
          <input type="text" bind:value={description} placeholder="Priority support" />
        </label>
        <label>
          <span>Default agent id</span>
          <input type="text" bind:value={defaultAgent} placeholder="ava" />
        </label>
        <label>
          <span>AI reply locale</span>
          <select bind:value={aiLocale}>
            <option value="">Inherit tenant</option>
            <option value="en">English</option>
            <option value="th">ไทย</option>
          </select>
        </label>
        <label>
          <span>Max minutes / call (0 = inherit)</span>
          <input type="number" min="0" bind:value={maxPerCall} />
        </label>
        <label>
          <span>Max minutes / day (0 = inherit)</span>
          <input type="number" min="0" bind:value={maxPerDay} />
        </label>
        <label class="row">
          <input type="checkbox" bind:checked={active} />
          <span>Active</span>
        </label>
        <div class="actions">
          <button class="btn" type="button" disabled={saving} onclick={saveTier}>
            {saving ? '…' : editingId ? 'Update' : 'Create'}
          </button>
          {#if editingId}
            <button class="btn ghost" type="button" onclick={resetForm}>Cancel</button>
          {/if}
        </div>
      </section>

      <section class="card">
        <h2>Tiers ({tiers.length})</h2>
        {#if !tiers.length}
          <p class="dim">No tiers yet — create VIP or Standard.</p>
        {:else}
          <ul class="list">
            {#each tiers as t}
              <li>
                <div>
                  <strong>{t.name}</strong>
                  <code>{t.slug}</code>
                  <span class="meta">
                    prio {t.priority} · locale {t.ai_reply_locale || '—'} · call
                    {capLabel(t.max_minutes_per_call)} / day {capLabel(t.max_call_minutes_per_day)}
                    {#if !t.active}
                      · <em>inactive</em>
                    {/if}
                  </span>
                </div>
                <div class="row-actions">
                  <button class="btn ghost" type="button" onclick={() => startEdit(t)}>Edit</button>
                  <button class="btn ghost danger" type="button" onclick={() => removeTier(t)}
                    >Delete</button
                  >
                </div>
              </li>
            {/each}
          </ul>
        {/if}
      </section>

      <section class="card full">
        <h2>Groups (ops labels)</h2>
        <p class="dim">Assignment to customers comes with S19 identity. Labels only for now.</p>
        <div class="group-form">
          <input type="text" bind:value={groupName} placeholder="Group name" />
          <input type="text" bind:value={groupSlug} placeholder="slug (optional)" />
          <button class="btn" type="button" onclick={addGroup}>Add group</button>
        </div>
        {#if groups.length}
          <ul class="list">
            {#each groups as g}
              <li>
                <div>
                  <strong>{g.name}</strong>
                  <code>{g.slug}</code>
                </div>
                <button class="btn ghost danger" type="button" onclick={() => removeGroup(g)}
                  >Delete</button
                >
              </li>
            {/each}
          </ul>
        {/if}
      </section>
    </div>
  {/if}
</div>

<style>
  .page-wrap {
    max-width: 960px;
    margin: 0 auto;
    padding: 20px;
  }
  .head {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 12px;
    margin-bottom: 16px;
    flex-wrap: wrap;
  }
  h1 {
    margin: 0;
    font-size: 24px;
  }
  .sub {
    margin: 6px 0 0;
    color: var(--muted);
    font-size: 13px;
  }
  .grid {
    display: grid;
    gap: 16px;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
  }
  .card {
    border: 1px solid var(--line);
    border-radius: 12px;
    padding: 16px 18px;
    background: rgb(12 18 32 / 80%);
  }
  .card.full {
    grid-column: 1 / -1;
  }
  .card h2 {
    margin: 0 0 14px;
    font-size: 15px;
  }
  label {
    display: flex;
    flex-direction: column;
    gap: 6px;
    margin-bottom: 10px;
    font-size: 12px;
    color: var(--muted);
  }
  label.row {
    flex-direction: row;
    align-items: center;
    gap: 8px;
  }
  input,
  select {
    padding: 8px 10px;
    border-radius: 8px;
    border: 1px solid var(--line);
    background: rgb(8 12 22);
    color: inherit;
    font-size: 14px;
  }
  .actions {
    display: flex;
    gap: 8px;
    margin-top: 8px;
  }
  .list {
    list-style: none;
    margin: 0;
    padding: 0;
  }
  .list li {
    display: flex;
    justify-content: space-between;
    gap: 12px;
    padding: 10px 0;
    border-bottom: 1px solid var(--line);
    align-items: flex-start;
  }
  .list code {
    font-size: 11px;
    margin-left: 6px;
    color: var(--cyan);
  }
  .meta {
    display: block;
    font-size: 11px;
    color: var(--muted);
    margin-top: 4px;
  }
  .row-actions {
    display: flex;
    gap: 6px;
    flex-shrink: 0;
  }
  .btn.danger {
    color: var(--danger);
  }
  .group-form {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
    margin-bottom: 12px;
  }
  .group-form input {
    flex: 1;
    min-width: 120px;
  }
  .dim {
    color: var(--muted);
    font-size: 13px;
  }
  .link {
    font-size: 13px;
  }
</style>
