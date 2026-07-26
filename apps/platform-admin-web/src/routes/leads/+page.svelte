<script lang="ts">
  import { onMount } from 'svelte';
  import {
    addLeadNote,
    getLead,
    LEAD_STATUSES,
    listLeads,
    patchLead,
    sourceLabel,
    type LeadDetail,
    type LeadListItem
  } from '$lib/api/leads';
  import { ApiError } from '$lib/api/http';
  import { feedback } from '$lib/feedback.svelte';

  let leads = $state<LeadListItem[]>([]);
  let total = $state(0);
  let loading = $state(true);
  let statusFilter = $state('new');
  let kindFilter = $state('');
  let search = $state('');

  let selectedId = $state('');
  let detail = $state<LeadDetail | null>(null);
  let detailLoading = $state(false);
  let editStatus = $state('');
  let editAssigned = $state('');
  let noteBody = $state('');
  let saving = $state(false);
  let noting = $state(false);

  async function load() {
    loading = true;
    try {
      const res = await listLeads({
        status: statusFilter,
        kind: kindFilter,
        q: search.trim(),
        limit: 50,
        offset: 0
      });
      leads = res.leads ?? [];
      total = res.total ?? leads.length;
    } catch (err) {
      leads = [];
      total = 0;
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load leads');
    } finally {
      loading = false;
    }
  }

  async function openLead(id: string) {
    selectedId = id;
    detailLoading = true;
    detail = null;
    try {
      const res = await getLead(id);
      detail = res;
      editStatus = res.status ?? '';
      editAssigned = res.assigned_to ?? '';
      noteBody = '';
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load lead');
      selectedId = '';
    } finally {
      detailLoading = false;
    }
  }

  function closeDrawer() {
    selectedId = '';
    detail = null;
  }

  async function saveDetail() {
    if (!selectedId) return;
    saving = true;
    try {
      const res = await patchLead(selectedId, {
        status: editStatus || undefined,
        assigned_to: editAssigned.trim() || undefined
      });
      detail = res;
      editStatus = res.status ?? editStatus;
      editAssigned = res.assigned_to ?? editAssigned;
      feedback.success('Lead updated');
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to update lead');
    } finally {
      saving = false;
    }
  }

  async function submitNote() {
    if (!selectedId || !noteBody.trim()) return;
    noting = true;
    try {
      await addLeadNote(selectedId, noteBody.trim());
      noteBody = '';
      feedback.success('Note added');
      const res = await getLead(selectedId);
      detail = res;
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to add note');
    } finally {
      noting = false;
    }
  }

  function formatDate(value?: string) {
    if (!value) return '—';
    try {
      return new Date(value).toLocaleString();
    } catch {
      return value;
    }
  }

  function historyRows(d: LeadDetail) {
    return d.history ?? d.events ?? [];
  }

  onMount(load);
</script>

<div class="page-head">
  <div>
    <p class="eyebrow">SALES</p>
    <h1>Leads</h1>
    <p class="lede">Marketing lead lifecycle — no tenant private customer data or payment secrets.</p>
  </div>
  <button class="btn ghost" type="button" onclick={load} disabled={loading}>Refresh</button>
</div>

<section class="filters card">
  <label>
    Status
    <select bind:value={statusFilter} onchange={load}>
      <option value="">all</option>
      {#each LEAD_STATUSES as s}
        <option value={s}>{s}</option>
      {/each}
    </select>
  </label>
  <label>
    Kind
    <select bind:value={kindFilter} onchange={load}>
      <option value="">all</option>
      <option value="book_demo">book_demo</option>
      <option value="contact">contact</option>
      <option value="newsletter">newsletter</option>
    </select>
  </label>
  <label class="search">
    Search
    <input
      bind:value={search}
      placeholder="email / company"
      onkeydown={(e) => e.key === 'Enter' && load()}
    />
  </label>
  <div class="filter-actions">
    <button class="btn" type="button" onclick={load} disabled={loading}>Apply</button>
  </div>
</section>

<section class="card table-card">
  <div class="section-head">
    <div>
      <h2>Inbox</h2>
      <p>{loading ? 'Loading…' : `${leads.length} shown · ${total} total`}</p>
    </div>
  </div>

  {#if loading}
    <p class="muted">Loading leads…</p>
  {:else if !leads.length}
    <div class="empty">
      <strong>No leads found.</strong>
      <span>Try another status, kind, or search term.</span>
    </div>
  {:else}
    <div class="table-wrap">
      <table>
        <thead>
          <tr>
            <th>ID</th>
            <th>Email</th>
            <th>Company</th>
            <th>Kind</th>
            <th>Status</th>
            <th>Source</th>
            <th>Created</th>
            <th></th>
          </tr>
        </thead>
        <tbody>
          {#each leads as lead (lead.id)}
            <tr class:selected={selectedId === lead.id}>
              <td><code>{lead.id}</code></td>
              <td>
                <strong>{lead.email}</strong>
                {#if lead.full_name}<small>{lead.full_name}</small>{/if}
              </td>
              <td>{lead.company_name || '—'}</td>
              <td><span class="badge">{lead.kind}</span></td>
              <td><span class="badge status">{lead.status}</span></td>
              <td class="muted source">{sourceLabel(lead)}</td>
              <td class="time">{formatDate(lead.created_at)}</td>
              <td>
                <button class="btn sm ghost" type="button" onclick={() => openLead(lead.id)}>Open</button>
              </td>
            </tr>
          {/each}
        </tbody>
      </table>
    </div>
  {/if}
</section>

{#if selectedId}
  <div class="drawer-backdrop" role="presentation" onclick={closeDrawer}></div>
  <aside class="drawer" role="dialog" aria-modal="true" aria-label="Lead detail">
    <header class="drawer-head">
      <div>
        <p class="eyebrow">LEAD DETAIL</p>
        <h2><code>{selectedId}</code></h2>
      </div>
      <button class="btn sm ghost" type="button" onclick={closeDrawer}>Close</button>
    </header>

    {#if detailLoading || !detail}
      <p class="muted pad">Loading detail…</p>
    {:else}
      <div class="drawer-body">
        <div class="meta">
          <div><span class="muted">Email</span><strong>{detail.email}</strong></div>
          <div><span class="muted">Name</span><strong>{detail.full_name || '—'}</strong></div>
          <div><span class="muted">Company</span><strong>{detail.company_name || '—'}</strong></div>
          <div><span class="muted">Phone</span><strong>{detail.phone || '—'}</strong></div>
          <div><span class="muted">Kind</span><strong>{detail.kind}</strong></div>
          <div><span class="muted">Channel</span><strong>{detail.preferred_channel || '—'}</strong></div>
        </div>

        {#if detail.use_case}
          <div class="block">
            <span class="muted">Use case</span>
            <p>{detail.use_case}</p>
          </div>
        {/if}

        <div class="block">
          <span class="muted">Attribution</span>
          <ul class="attr">
            <li>utm_source: {detail.utm_source || '—'}</li>
            <li>utm_medium: {detail.utm_medium || '—'}</li>
            <li>utm_campaign: {detail.utm_campaign || '—'}</li>
            <li>utm_content: {detail.utm_content || '—'}</li>
            <li>utm_term: {detail.utm_term || '—'}</li>
            <li>ref: {detail.referral_code || '—'}</li>
            <li>landing: {detail.landing_path || '—'}</li>
            <li>package interest: {detail.package_interest_id || '—'}</li>
          </ul>
        </div>

        <div class="edit-row">
          <label>
            Status
            <select bind:value={editStatus}>
              {#each LEAD_STATUSES as s}
                <option value={s}>{s}</option>
              {/each}
            </select>
          </label>
          <label>
            Assigned to
            <input bind:value={editAssigned} placeholder="user id (optional)" />
          </label>
          <button class="btn" type="button" onclick={saveDetail} disabled={saving}>
            {saving ? 'Saving…' : 'Save'}
          </button>
        </div>

        <div class="block">
          <span class="muted">Notes</span>
          <div class="note-form">
            <textarea bind:value={noteBody} rows="3" maxlength="4000" placeholder="Follow-up note…"
            ></textarea>
            <button class="btn sm" type="button" onclick={submitNote} disabled={noting || !noteBody.trim()}>
              {noting ? 'Adding…' : 'Add note'}
            </button>
          </div>
          {#if detail.notes?.length}
            <ul class="notes">
              {#each detail.notes as note (note.id)}
                <li>
                  <p>{note.body}</p>
                  <small class="muted"
                    >{formatDate(note.created_at)} · {note.created_by || 'system'}</small
                  >
                </li>
              {/each}
            </ul>
          {:else}
            <p class="muted empty-notes">No notes yet.</p>
          {/if}
        </div>

        <div class="block">
          <span class="muted">History</span>
          {#if historyRows(detail).length}
            <ul class="history">
              {#each historyRows(detail) as ev (ev.id)}
                <li>
                  <code>{ev.from_status || '∅'}</code> → <code>{ev.to_status}</code>
                  <small class="muted"
                    >{ev.actor || 'system'} · {formatDate(ev.created_at)}</small
                  >
                </li>
              {/each}
            </ul>
          {:else}
            <p class="muted empty-notes">No status history.</p>
          {/if}
        </div>
      </div>
    {/if}
  </aside>
{/if}

<style>
  .page-head {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 16px;
    margin-bottom: 22px;
  }
  .eyebrow {
    margin: 0 0 8px;
    color: var(--cyan);
    font-size: 11px;
    letter-spacing: 0.16em;
    font-weight: 700;
  }
  h1 {
    margin: 0;
    font-size: 32px;
  }
  .lede {
    margin: 7px 0 0;
    color: var(--muted);
  }
  .filters {
    display: grid;
    grid-template-columns: repeat(4, minmax(120px, 1fr));
    gap: 14px;
    margin-bottom: 14px;
  }
  label {
    display: grid;
    gap: 6px;
    color: var(--muted);
    font-size: 12px;
  }
  input,
  select,
  textarea {
    width: 100%;
    min-height: 38px;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: rgb(4 10 22 / 74%);
    color: var(--ink);
    padding: 8px 10px;
  }
  textarea {
    min-height: 80px;
    resize: vertical;
  }
  .filter-actions {
    display: flex;
    align-items: end;
  }
  .table-card {
    overflow: hidden;
  }
  .section-head {
    display: flex;
    justify-content: space-between;
    margin-bottom: 14px;
  }
  h2 {
    margin: 0;
    font-size: 18px;
  }
  .section-head p,
  .muted {
    margin: 5px 0 0;
    color: var(--muted);
    font-size: 12px;
  }
  .table-wrap {
    overflow-x: auto;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    min-width: 920px;
    font-size: 12px;
  }
  th,
  td {
    padding: 11px 9px;
    border-bottom: 1px solid var(--line);
    text-align: left;
    vertical-align: top;
  }
  th {
    color: var(--muted);
    font-size: 11px;
    font-weight: 600;
  }
  td small {
    display: block;
    color: var(--muted);
    margin-top: 3px;
  }
  tr.selected {
    background: rgb(24 93 220 / 12%);
  }
  .time {
    color: var(--muted);
    white-space: nowrap;
  }
  code {
    color: #a9c9ff;
    font-size: 11px;
  }
  .badge {
    display: inline-flex;
    border: 1px solid var(--line);
    border-radius: 999px;
    padding: 2px 8px;
    font-size: 11px;
  }
  .badge.status {
    border-color: rgb(22 199 255 / 28%);
    color: var(--cyan);
  }
  .source {
    max-width: 160px;
    word-break: break-word;
  }
  .btn.sm {
    padding: 6px 10px;
    font-size: 12px;
  }
  .empty {
    display: grid;
    gap: 7px;
    padding: 35px 0;
    color: var(--muted);
  }
  .empty strong {
    color: var(--ink);
  }

  .drawer-backdrop {
    position: fixed;
    inset: 0;
    background: rgb(0 0 0 / 48%);
    z-index: 40;
  }
  .drawer {
    position: fixed;
    top: 0;
    right: 0;
    width: min(440px, 100vw);
    height: 100vh;
    z-index: 50;
    border-left: 1px solid var(--line);
    background: linear-gradient(180deg, #0a1222, #060b16);
    box-shadow: -20px 0 60px rgb(0 0 0 / 45%);
    display: flex;
    flex-direction: column;
    overflow: hidden;
  }
  .drawer-head {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    gap: 12px;
    padding: 18px 18px 12px;
    border-bottom: 1px solid var(--line);
  }
  .drawer-head h2 {
    margin: 0;
    font-size: 14px;
    word-break: break-all;
  }
  .drawer-body {
    padding: 16px 18px 32px;
    overflow: auto;
    display: grid;
    gap: 16px;
  }
  .pad {
    padding: 18px;
  }
  .meta {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 12px;
  }
  .meta div {
    display: grid;
    gap: 3px;
  }
  .meta strong {
    font-size: 13px;
    word-break: break-word;
  }
  .block {
    display: grid;
    gap: 8px;
  }
  .block p {
    margin: 0;
    font-size: 13px;
    line-height: 1.5;
  }
  .attr {
    margin: 0;
    padding-left: 16px;
    color: #b8c7e2;
    font-size: 12px;
    line-height: 1.55;
  }
  .edit-row {
    display: grid;
    gap: 10px;
    padding: 12px;
    border: 1px solid var(--line);
    border-radius: 10px;
    background: rgb(8 14 28 / 70%);
  }
  .note-form {
    display: grid;
    gap: 8px;
  }
  .notes,
  .history {
    list-style: none;
    margin: 0;
    padding: 0;
    display: grid;
    gap: 10px;
  }
  .notes li,
  .history li {
    border: 1px solid var(--line);
    border-radius: 8px;
    padding: 10px;
    background: rgb(6 12 24 / 55%);
  }
  .notes p {
    margin: 0 0 6px;
    font-size: 13px;
  }
  .history li {
    font-size: 12px;
  }
  .history small,
  .notes small {
    display: block;
    margin-top: 4px;
  }
  .empty-notes {
    margin: 0;
  }

  @media (max-width: 820px) {
    .filters {
      grid-template-columns: 1fr 1fr;
    }
  }
  @media (max-width: 520px) {
    .filters {
      grid-template-columns: 1fr;
    }
    .meta {
      grid-template-columns: 1fr;
    }
  }
</style>
