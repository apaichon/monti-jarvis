<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { ApiError } from '$lib/api/http';
  import {
    getConversationRecord,
    getConversationArchiveObjectBlob,
    listConversationRecords,
    retryConversationArchive,
    type ConversationArchiveObject,
    type ConversationRecord,
    type ConversationTranscriptLine
  } from '$lib/api/operations';
  import { feedback } from '$lib/feedback.svelte';

  let records = $state<ConversationRecord[]>([]);
  let selected = $state<ConversationRecord | null>(null);
  let transcript = $state<ConversationTranscriptLine[]>([]);
  let archiveObjects = $state<ConversationArchiveObject[]>([]);
  let status = $state('');
  let startDate = $state('');
  let endDate = $state('');
  let loading = $state(true);
  let retrying = $state('');
  let audioUrls = $state<Record<string, string>>({});

  onDestroy(() => {
    revokeAudioUrls();
  });

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/conversation-records`)}`);
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

  async function load() {
    loading = true;
    try {
      records = await listConversationRecords(status, startDate, endDate);
      selected = records[0] ?? null;
      transcript = [];
      archiveObjects = [];
      revokeAudioUrls();
      if (selected) await inspect(selected.id);
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load records');
    } finally {
      loading = false;
    }
  }

  function applyFilters(event: Event) {
    event.preventDefault();
    void load();
  }

  function clearDateFilters() {
    const today = localISODate();
    startDate = today;
    endDate = today;
    void load();
  }

  async function inspect(id: string) {
    try {
      const detail = await getConversationRecord(id);
      selected = detail.record;
      transcript = detail.transcript ?? [];
      archiveObjects = detail.archive_objects ?? [];
      await loadAudioUrls(selected.id, archiveObjects);
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load record');
    }
  }

  const audioObjects = $derived(archiveObjects.filter((item) => item.object_type === 'audio' && item.status === 'stored'));
  const transcriptObjects = $derived(archiveObjects.filter((item) => item.object_type === 'transcript'));

  function revokeAudioUrls() {
    for (const url of Object.values(audioUrls)) URL.revokeObjectURL(url);
    audioUrls = {};
  }

  async function loadAudioUrls(recordId: string, objects: ConversationArchiveObject[]) {
    revokeAudioUrls();
    const audio = objects.filter((item) => item.object_type === 'audio' && item.status === 'stored');
    const next: Record<string, string> = {};
    await Promise.all(
      audio.map(async (obj) => {
        try {
          const blob = await getConversationArchiveObjectBlob(recordId, obj.id);
          next[obj.id] = URL.createObjectURL(blob);
        } catch {
          // Metadata remains visible even if the object cannot be fetched.
        }
      })
    );
    audioUrls = next;
  }

  async function retry(id: string) {
    retrying = id;
    try {
      await retryConversationArchive(id);
      feedback.success('Archive retry queued');
      await load();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Retry failed');
    } finally {
      retrying = '';
    }
  }
</script>

<div class="page-wrap">
  <div class="head">
    <div>
      <h1>Conversation records</h1>
      <p>Tenant-scoped chat and voice archive metadata.</p>
    </div>
    <form class="filters" onsubmit={applyFilters}>
      <label>
        Status
        <select bind:value={status}>
          <option value="">All</option>
          <option value="recording">Recording</option>
          <option value="archived">Archived</option>
          <option value="archive_failed">Archive failed</option>
        </select>
      </label>
      <label>
        Start date
        <input type="date" bind:value={startDate} />
      </label>
      <label>
        End date
        <input type="date" bind:value={endDate} />
      </label>
      <button class="btn" type="submit">Filter</button>
      <button class="btn secondary" type="button" onclick={clearDateFilters}>Clear</button>
    </form>
  </div>

  {#if loading}
    <p class="muted">Loading…</p>
  {:else}
    <div class="grid">
      <section class="card">
        <table>
          <thead><tr><th>Started</th><th>Channel</th><th>Avatar</th><th>Status</th><th>Gaps</th></tr></thead>
          <tbody>
            {#each records as row (row.id)}
              <tr class:active={selected?.id === row.id} onclick={() => inspect(row.id)}>
                <td>{new Date(row.started_at).toLocaleString()}</td>
                <td>{row.channel}</td>
                <td>{row.avatar_name || row.avatar_id || '—'}</td>
                <td>{row.status}</td>
                <td>{row.knowledge_gap_count ?? 0}</td>
              </tr>
            {/each}
          </tbody>
        </table>
      </section>

      <section class="card">
        {#if selected}
          <h2>{selected.id}</h2>
          <p class="muted">Call {selected.call_id || '—'} · Customer {selected.customer_id || 'anonymous'}</p>
          <p>Status: <strong>{selected.status}</strong> · Objects: {selected.archive_object_count ?? 0}</p>
          <section class="subcard">
            <h3>Conversation log</h3>
            {#if transcript.length > 0}
              <div class="turns">
                {#each transcript as line (line.id)}
                  <div class="turn" class:user={line.role === 'caller' || line.role === 'user'}>
                    <span>{line.role}</span>
                    <p>{line.content}</p>
                  </div>
                {/each}
              </div>
            {:else if transcriptObjects.length > 0}
              <p class="muted">Transcript archive exists. Re-open after server reload if turn preview is not populated yet.</p>
            {:else}
              <p class="muted">No transcript turns found for this record yet.</p>
            {/if}
          </section>
          <section class="subcard">
            <h3>Voice/audio</h3>
            {#if audioObjects.length > 0}
              {#each audioObjects as obj (obj.id)}
                <div class="audio-row">
                  <p class="muted">Call recording: {obj.object_key} · {obj.content_type} · {obj.size_bytes} bytes</p>
                  {#if audioUrls[obj.id]}
                    <audio controls src={audioUrls[obj.id]}></audio>
                  {:else}
                    <p class="muted">Loading audio…</p>
                  {/if}
                </div>
              {/each}
            {:else}
              <p class="muted">Audio playback is not available for this record. Current local voice flow stores transcript/archive metadata; audio recording will appear here after an audio object is archived.</p>
            {/if}
          </section>
          <section class="subcard">
            <h3>Archive objects</h3>
            {#if archiveObjects.length > 0}
              {#each archiveObjects as obj (obj.id)}
                <p class="muted">{obj.object_type} · {obj.status} · {obj.protection_mode} · {obj.object_key || obj.error_code}</p>
              {/each}
            {:else}
              <p class="muted">No archive object metadata.</p>
            {/if}
          </section>
          <pre>{JSON.stringify(selected.summary ?? {}, null, 2)}</pre>
          {#if selected.status === 'archive_failed'}
            <button class="btn" type="button" disabled={retrying === selected.id} onclick={() => retry(selected!.id)}>
              {retrying === selected.id ? 'Retrying…' : 'Retry archive'}
            </button>
          {/if}
        {:else}
          <p class="muted">No record selected.</p>
        {/if}
      </section>
    </div>
  {/if}
</div>

<style>
  .page-wrap { max-width: 1100px; margin: 0 auto; padding: 20px; }
  .head { display: flex; justify-content: space-between; gap: 16px; align-items: end; margin-bottom: 16px; }
  .filters { display: flex; flex-wrap: wrap; gap: 10px; align-items: end; justify-content: end; }
  .filters label { display: grid; gap: 4px; color: var(--muted); font-size: 12px; }
  h1 { margin: 0; font-size: 24px; }
  p { margin: 6px 0; }
  .muted { color: var(--muted); font-size: 13px; }
  .grid { display: grid; grid-template-columns: minmax(0, 1.4fr) minmax(320px, 0.8fr); gap: 16px; }
  .card { border: 1px solid var(--line); border-radius: 12px; background: rgb(12 18 32 / 80%); padding: 16px; overflow: auto; }
  table { width: 100%; border-collapse: collapse; font-size: 13px; }
  th, td { padding: 10px; border-bottom: 1px solid var(--line); text-align: left; }
  tr { cursor: pointer; }
  tr.active { background: rgb(0 140 255 / 12%); }
  select, input { padding: 8px 10px; border-radius: 8px; border: 1px solid var(--line); background: rgb(8 12 22); color: inherit; }
  .btn { padding: 8px 12px; border: 1px solid var(--cyan); border-radius: 8px; background: rgb(0 140 255 / 15%); color: inherit; cursor: pointer; }
  .btn.secondary { border-color: var(--line); background: transparent; }
  pre { white-space: pre-wrap; word-break: break-word; background: #071120; padding: 12px; border-radius: 10px; }
  .subcard { border: 1px solid var(--line); border-radius: 10px; padding: 12px; margin: 12px 0; background: rgb(7 17 32 / 68%); }
  .subcard h3 { margin: 0 0 8px; font-size: 14px; }
  .turns { display: grid; gap: 8px; }
  .turn { border-left: 3px solid var(--cyan); padding: 8px 10px; background: rgb(0 140 255 / 8%); border-radius: 8px; }
  .turn.user { border-left-color: #7c3aed; background: rgb(124 58 237 / 10%); }
  .turn span { display: block; color: var(--muted); font-size: 11px; text-transform: uppercase; margin-bottom: 4px; }
  .turn p { margin: 0; }
  .audio-row { display: grid; gap: 8px; margin: 10px 0; }
  audio { width: 100%; height: 36px; }
  @media (max-width: 850px) { .grid { grid-template-columns: 1fr; } .head { align-items: start; flex-direction: column; } }
</style>
