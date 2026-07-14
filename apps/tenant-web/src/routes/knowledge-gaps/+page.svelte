<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { ApiError } from '$lib/api/http';
  import { listKnowledgeGaps, patchKnowledgeGap, type KnowledgeGap } from '$lib/api/operations';
  import { feedback } from '$lib/feedback.svelte';

  let gaps = $state<KnowledgeGap[]>([]);
  let status = $state<KnowledgeGap['status'] | ''>('open');
  let note = $state('');
  let loading = $state(true);
  let saving = $state('');

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/knowledge-gaps`)}`);
      return;
    }
    await load();
  });

  async function load() {
    loading = true;
    try {
      gaps = await listKnowledgeGaps(status);
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load knowledge gaps');
    } finally {
      loading = false;
    }
  }

  async function updateGap(id: string, nextStatus: KnowledgeGap['status']) {
    saving = id;
    try {
      await patchKnowledgeGap(id, { status: nextStatus, reviewer_note: note });
      feedback.success('Knowledge gap updated');
      note = '';
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Update failed');
    } finally {
      saving = '';
    }
  }
</script>

<div class="page-wrap">
  <div class="head">
    <div>
      <h1>Knowledge gaps</h1>
      <p>Review RAG misses and low-confidence answers from tenant conversations.</p>
    </div>
    <label>
      Status
      <select bind:value={status} onchange={load}>
        <option value="">All</option>
        <option value="open">Open</option>
        <option value="snoozed">Snoozed</option>
        <option value="resolved">Resolved</option>
        <option value="ignored">Ignored</option>
      </select>
    </label>
  </div>

  {#if loading}
    <p class="muted">Loading…</p>
  {:else if gaps.length === 0}
    <section class="card"><p class="muted">No knowledge gaps for this filter.</p></section>
  {:else}
    <div class="gap-list">
      {#each gaps as gap (gap.id)}
        <section class="card">
          <div class="gap-head">
            <div>
              <strong>{gap.question}</strong>
              <p class="muted">{gap.gap_reason} · confidence {gap.confidence}</p>
            </div>
            <span class="badge">{gap.status}</span>
          </div>
          {#if gap.answer_excerpt}
            <p>{gap.answer_excerpt}</p>
          {/if}
          <p class="muted">Conversation {gap.conversation_record_id || '—'} · Avatar {gap.avatar_id || '—'}</p>
          <textarea bind:value={note} placeholder="Reviewer note for next action"></textarea>
          <div class="actions">
            <button class="btn" disabled={saving === gap.id} onclick={() => updateGap(gap.id, 'resolved')}>Resolve</button>
            <button class="btn ghost" disabled={saving === gap.id} onclick={() => updateGap(gap.id, 'snoozed')}>Snooze</button>
            <button class="btn ghost" disabled={saving === gap.id} onclick={() => updateGap(gap.id, 'ignored')}>Ignore</button>
          </div>
        </section>
      {/each}
    </div>
  {/if}
</div>

<style>
  .page-wrap { max-width: 960px; margin: 0 auto; padding: 20px; }
  .head { display: flex; justify-content: space-between; gap: 16px; align-items: end; margin-bottom: 16px; }
  h1 { margin: 0; font-size: 24px; }
  p { margin: 6px 0; }
  .muted { color: var(--muted); font-size: 13px; }
  .gap-list { display: grid; gap: 14px; }
  .card { border: 1px solid var(--line); border-radius: 12px; background: rgb(12 18 32 / 80%); padding: 16px; }
  .gap-head { display: flex; justify-content: space-between; gap: 12px; }
  .badge { border: 1px solid var(--line); border-radius: 999px; padding: 4px 8px; color: var(--muted); font-size: 12px; }
  textarea, select { width: 100%; box-sizing: border-box; padding: 8px 10px; border-radius: 8px; border: 1px solid var(--line); background: rgb(8 12 22); color: inherit; }
  textarea { min-height: 72px; margin-top: 10px; }
  .actions { display: flex; gap: 8px; margin-top: 10px; flex-wrap: wrap; }
  @media (max-width: 720px) { .head { align-items: start; flex-direction: column; } }
</style>
