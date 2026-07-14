<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { ApiError } from '$lib/api/http';
  import {
    addTicketNote,
    getTicket,
    listTickets,
    updateTicket,
    type Ticket,
    type TicketCategory,
    type TicketDetail,
    type TicketPriority,
    type TicketStatus
  } from '$lib/api/tickets';
  import { feedback } from '$lib/feedback.svelte';

  let tickets = $state<Ticket[]>([]);
  let detail = $state<TicketDetail | null>(null);
  let status = $state<TicketStatus | ''>('open');
  let priority = $state<TicketPriority | ''>('');
  let category = $state<TicketCategory | ''>('');
  let avatarId = $state('');
  let assigneeUserId = $state('');
  let startDate = $state('');
  let endDate = $state('');
  let loading = $state(true);
  let saving = $state(false);
  let savingNote = $state(false);
  let note = $state('');
  let draftStatus = $state<TicketStatus>('open');
  let draftPriority = $state<TicketPriority>('normal');
  let draftAssignee = $state('');

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/tickets`)}`);
      return;
    }
    const today = localISODate();
    startDate = today;
    endDate = today;
    await load();
  });

  function localISODate(date = new Date()) {
    const local = new Date(date.getTime() - date.getTimezoneOffset() * 60_000);
    return local.toISOString().slice(0, 10);
  }

  function errorMessage(err: unknown, fallback: string) {
    return err instanceof ApiError ? err.message : fallback;
  }

  async function load() {
    loading = true;
    try {
      const result = await listTickets({
        startDate,
        endDate,
        status,
        priority,
        category,
        avatarId,
        assigneeUserId
      });
      tickets = result.tickets || [];
      detail = tickets[0] ? await getTicket(tickets[0].id) : null;
      syncDraft();
    } catch (err) {
      feedback.error(errorMessage(err, 'Failed to load tickets'));
    } finally {
      loading = false;
    }
  }

  async function inspect(ticket: Ticket) {
    try {
      detail = await getTicket(ticket.id);
      syncDraft();
    } catch (err) {
      feedback.error(errorMessage(err, 'Failed to load ticket'));
    }
  }

  function syncDraft() {
    if (!detail) return;
    draftStatus = detail.ticket.status;
    draftPriority = detail.ticket.priority;
    draftAssignee = detail.ticket.assignee_user_id || '';
  }

  async function applyFilters(event: Event) {
    event.preventDefault();
    await load();
  }

  function resetToday() {
    const today = localISODate();
    startDate = today;
    endDate = today;
    status = 'open';
    priority = '';
    category = '';
    avatarId = '';
    assigneeUserId = '';
    void load();
  }

  async function saveTicket() {
    if (!detail || saving) return;
    saving = true;
    try {
      const result = await updateTicket(detail.ticket.id, {
        status: draftStatus,
        priority: draftPriority,
        assignee_user_id: draftAssignee
      });
      detail = { ...detail, ticket: result.ticket };
      tickets = tickets.map((item) => item.id === result.ticket.id ? { ...item, ...result.ticket } : item);
      feedback.success('Ticket updated');
    } catch (err) {
      feedback.error(errorMessage(err, 'Ticket update failed'));
    } finally {
      saving = false;
    }
  }

  async function saveNote() {
    if (!detail || !note.trim() || savingNote) return;
    savingNote = true;
    try {
      await addTicketNote(detail.ticket.id, note.trim());
      detail = await getTicket(detail.ticket.id);
      note = '';
      feedback.success('Internal note added');
    } catch (err) {
      feedback.error(errorMessage(err, 'Could not add note'));
    } finally {
      savingNote = false;
    }
  }

  function formatDate(value: string) {
    return new Date(value).toLocaleString();
  }

  function summaryText(key: string) {
    const value = detail?.ticket.source_summary?.[key];
    return typeof value === 'string' ? value : '';
  }

  function summaryNumber(key: string) {
    const value = detail?.ticket.source_summary?.[key];
    return typeof value === 'number' ? value : 0;
  }
</script>

<svelte:head><title>Tickets · Monti Tenant Console</title></svelte:head>

<div class="ticket-page">
  <div class="page-head">
    <div>
      <p class="eyebrow">Operations</p>
      <h1>Tickets</h1>
      <p class="muted">Customer-confirmed follow-up requests for your tenant team.</p>
    </div>
    <span class="scope-badge">Tenant scoped</span>
  </div>

  <form class="filters card" onsubmit={applyFilters}>
    <label>Start date<input type="date" bind:value={startDate} /></label>
    <label>End date<input type="date" bind:value={endDate} /></label>
    <label>Status
      <select bind:value={status}>
        <option value="">All statuses</option>
        <option value="open">Open</option>
        <option value="in_progress">In progress</option>
        <option value="waiting_customer">Waiting customer</option>
        <option value="resolved">Resolved</option>
        <option value="closed">Closed</option>
      </select>
    </label>
    <label>Priority
      <select bind:value={priority}>
        <option value="">All priorities</option>
        <option value="low">Low</option>
        <option value="normal">Normal</option>
        <option value="high">High</option>
        <option value="urgent">Urgent</option>
      </select>
    </label>
    <label>Category
      <select bind:value={category}>
        <option value="">All categories</option>
        <option value="general">General</option>
        <option value="billing">Billing</option>
        <option value="technical">Technical</option>
        <option value="other">Other</option>
      </select>
    </label>
    <label>Avatar ID<input bind:value={avatarId} placeholder="Any avatar" /></label>
    <label>Assignee ID<input bind:value={assigneeUserId} placeholder="Any assignee" /></label>
    <div class="filter-actions">
      <button class="btn" type="submit">Filter</button>
      <button class="btn ghost" type="button" onclick={resetToday}>Today</button>
    </div>
  </form>

  {#if loading}
    <p class="muted loading">Loading tickets…</p>
  {:else if tickets.length === 0}
    <section class="card empty"><strong>No tickets match this filter.</strong><span>Confirmed customer requests will appear here.</span></section>
  {:else}
    <div class="ticket-grid">
      <section class="card queue" aria-label="Ticket queue">
        <div class="section-head"><h2>Queue</h2><span>{tickets.length}</span></div>
        <div class="queue-list">
          {#each tickets as ticket (ticket.id)}
            <button class="queue-row" class:selected={detail?.ticket.id === ticket.id} type="button" onclick={() => inspect(ticket)}>
              <span class="ticket-id">{ticket.id}</span>
              <strong>{ticket.subject}</strong>
              <span class="queue-meta"><i class:urgent={ticket.priority === 'urgent'}>{ticket.priority}</i><em>{ticket.status.replace('_', ' ')}</em></span>
              <small>{ticket.customer_label || 'Anonymous'} · {ticket.avatar_name || ticket.avatar_id || 'AI employee'} · {ticket.call_id || 'No call ref'}</small>
            </button>
          {/each}
        </div>
      </section>

      <section class="card detail" aria-label="Ticket detail">
        {#if detail}
          <div class="detail-head">
            <div>
              <p class="eyebrow">{detail.ticket.id}</p>
              <h2>{detail.ticket.subject}</h2>
              <p class="muted">{detail.ticket.customer_label || 'Anonymous'} · {detail.ticket.avatar_name || detail.ticket.avatar_id || 'AI employee'} · {detail.ticket.source.replace('_', ' ')}</p>
            </div>
            <span class:high={detail.ticket.priority === 'high' || detail.ticket.priority === 'urgent'} class="priority">{detail.ticket.priority}</span>
          </div>
          <p class="description">{detail.ticket.description || 'No description provided.'}</p>
          <div class="source-line">
            <span>Contact {detail.ticket.contact_email_masked || 'not provided'}</span>
            {#if detail.ticket.call_id || detail.ticket.conversation_record_id}
              <a
                href={`${base}/conversation-records?${detail.ticket.conversation_record_id ? `record_id=${encodeURIComponent(detail.ticket.conversation_record_id)}` : `call_id=${encodeURIComponent(detail.ticket.call_id || '')}`}`}
                class="link"
              >Open call record · {detail.ticket.call_id || detail.ticket.conversation_record_id}</a>
            {/if}
          </div>

          {#if detail.ticket.source_summary && Object.keys(detail.ticket.source_summary).length > 0}
            <section class="context-panel" aria-label="Source context">
              <div class="section-head"><h3>Customer context</h3><span>Source summary</span></div>
              {#if summaryText('customer_context')}<p class="context-request">{summaryText('customer_context')}</p>{/if}
              <div class="context-grid">
                {#if summaryText('topic')}<span><small>Topic</small><strong>{summaryText('topic')}</strong></span>{/if}
                {#if summaryText('channel')}<span><small>Channel</small><strong>{summaryText('channel')}</strong></span>{/if}
                {#if summaryText('avatar_name')}<span><small>Avatar</small><strong>{summaryText('avatar_name')}</strong></span>{/if}
                {#if summaryNumber('duration_seconds') > 0}<span><small>Duration</small><strong>{summaryNumber('duration_seconds')} sec</strong></span>{/if}
              </div>
            </section>
          {/if}

          <div class="edit-grid">
            <label>Status
              <select bind:value={draftStatus}>
                <option value="open">Open</option>
                <option value="in_progress">In progress</option>
                <option value="waiting_customer">Waiting customer</option>
                <option value="resolved">Resolved</option>
                <option value="closed">Closed</option>
              </select>
            </label>
            <label>Priority
              <select bind:value={draftPriority}>
                <option value="low">Low</option>
                <option value="normal">Normal</option>
                <option value="high">High</option>
                <option value="urgent">Urgent</option>
              </select>
            </label>
            <label>Assignee
              <input placeholder="Tenant user ID" bind:value={draftAssignee} />
            </label>
            <button class="btn save" type="button" disabled={saving} onclick={saveTicket}>{saving ? 'Saving…' : 'Save changes'}</button>
          </div>

          <div class="timeline">
            <div class="section-head"><h3>Timeline</h3><span>{detail.events.length} events</span></div>
            {#each detail.events as event (event.id)}
              <div class="event"><span class="event-dot"></span><div><strong>{event.event_type.replace('_', ' ')}</strong><small>{event.actor_type} · {formatDate(event.created_at)}</small>{#if event.note}<p>{event.note}</p>{/if}</div></div>
            {/each}
          </div>
          <div class="note-box">
            <label for="ticket-note">Internal note</label>
            <textarea id="ticket-note" bind:value={note} placeholder="Add context for the tenant team"></textarea>
            <button class="btn ghost" type="button" disabled={savingNote || !note.trim()} onclick={saveNote}>{savingNote ? 'Adding…' : 'Add note'}</button>
          </div>
        {/if}
      </section>
    </div>
  {/if}
</div>

<style>
  .ticket-page { max-width: 1240px; margin: 0 auto; }
  .page-head, .section-head, .detail-head, .source-line { display: flex; align-items: center; justify-content: space-between; gap: 16px; }
  .page-head { margin-bottom: 20px; }
  .eyebrow { margin: 0 0 6px; color: var(--cyan); font-size: 11px; letter-spacing: .12em; text-transform: uppercase; }
  h1, h2, h3 { margin: 0; }
  h1 { font-size: clamp(28px, 4vw, 40px); }
  h2 { font-size: 18px; }
  h3 { font-size: 14px; }
  .muted { color: var(--muted); }
  .scope-badge, .priority, .section-head > span { border: 1px solid var(--line); border-radius: 999px; padding: 6px 10px; color: var(--muted); font-size: 11px; white-space: nowrap; }
  .priority.high { color: var(--warn); border-color: rgb(240 184 63 / 45%); }
  .filters { display: grid; grid-template-columns: repeat(5, minmax(120px, 1fr)) auto; gap: 12px; align-items: end; margin-bottom: 18px; padding: 14px; }
  label { display: grid; gap: 6px; color: var(--muted); font-size: 11px; }
  input, select, textarea { width: 100%; border: 1px solid var(--line); border-radius: 8px; padding: 9px 10px; color: var(--ink); background: rgb(4 9 20 / 74%); }
  .filter-actions { display: flex; gap: 8px; }
  .btn { border: 1px solid rgb(74 135 255 / 46%); border-radius: 9px; padding: 9px 13px; background: linear-gradient(100deg, var(--blue), var(--violet)); color: var(--ink); font-weight: 650; white-space: nowrap; }
  .btn.ghost { background: rgb(13 23 42 / 62%); }
  .btn:disabled { opacity: .55; cursor: not-allowed; }
  .loading { padding: 24px 0; }
  .empty { display: grid; gap: 8px; padding: 32px; color: var(--muted); }
  .ticket-grid { display: grid; grid-template-columns: minmax(290px, .9fr) minmax(0, 1.5fr); gap: 18px; align-items: start; }
  .queue, .detail { min-width: 0; }
  .section-head { margin-bottom: 14px; }
  .section-head > span { padding: 4px 8px; }
  .queue-list { display: grid; gap: 7px; }
  .queue-row { display: grid; gap: 5px; width: 100%; border: 1px solid transparent; border-radius: 9px; padding: 12px; color: var(--ink); background: rgb(6 14 29 / 70%); text-align: left; }
  .queue-row:hover, .queue-row.selected { border-color: rgb(22 199 255 / 48%); background: rgb(12 35 65 / 78%); }
  .ticket-id, .queue-row small, .queue-meta { color: var(--muted); font-size: 11px; }
  .queue-meta { display: flex; gap: 8px; }
  .queue-meta i, .queue-meta em { font-style: normal; }
  .queue-meta i.urgent { color: var(--danger); }
  .detail { display: grid; gap: 16px; }
  .detail-head { align-items: start; }
  .detail-head h2 { font-size: 24px; }
  .description { margin: 0; line-height: 1.55; }
  .source-line { justify-content: flex-start; flex-wrap: wrap; color: var(--muted); font-size: 12px; }
  .context-panel { display: grid; gap: 10px; border: 1px solid rgb(22 199 255 / 22%); border-radius: 10px; padding: 12px; background: rgb(8 30 54 / 42%); }
  .context-panel .section-head { margin-bottom: 0; }
  .context-panel .section-head > span { text-transform: uppercase; letter-spacing: .08em; }
  .context-request { margin: 0; line-height: 1.5; }
  .context-grid { display: flex; flex-wrap: wrap; gap: 8px 18px; }
  .context-grid span { display: grid; gap: 3px; }
  .context-grid small { color: var(--muted); font-size: 10px; text-transform: uppercase; letter-spacing: .08em; }
  .context-grid strong { font-size: 12px; }
  .edit-grid { display: grid; grid-template-columns: repeat(3, minmax(120px, 1fr)) auto; gap: 10px; align-items: end; padding: 14px 0; border-top: 1px solid var(--line); border-bottom: 1px solid var(--line); }
  .save { height: 37px; }
  .timeline { display: grid; gap: 10px; }
  .event { display: grid; grid-template-columns: 14px 1fr; gap: 9px; align-items: start; }
  .event-dot { width: 8px; height: 8px; margin-top: 5px; border-radius: 50%; background: var(--cyan); box-shadow: 0 0 10px rgb(22 199 255 / 55%); }
  .event strong, .event small { display: block; }
  .event small { margin-top: 2px; color: var(--muted); font-size: 11px; }
  .event p { margin: 5px 0 0; color: var(--muted); font-size: 12px; }
  .note-box { display: grid; gap: 8px; padding-top: 4px; }
  textarea { min-height: 78px; resize: vertical; }
  @media (max-width: 980px) { .filters { grid-template-columns: repeat(3, minmax(120px, 1fr)); } .filter-actions { grid-column: 1 / -1; } }
  @media (max-width: 780px) { .ticket-grid { grid-template-columns: 1fr; } .edit-grid { grid-template-columns: repeat(2, minmax(120px, 1fr)); } .save { width: 100%; } }
  @media (max-width: 560px) { .page-head { align-items: start; flex-direction: column; } .filters { grid-template-columns: 1fr 1fr; } .edit-grid { grid-template-columns: 1fr; } .detail-head { align-items: start; flex-direction: column; } }
</style>
