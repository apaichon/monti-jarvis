<script lang="ts">
  import { onMount } from 'svelte';
  import { base } from '$app/paths';
  import { goto } from '$app/navigation';
  import { hasRegistrationSession } from '$lib/auth/session';
  import { feedback } from '$lib/feedback.svelte';
  import { ApiError } from '$lib/api/http';
  import {
    listScopes,
    listAgents,
    listDocuments,
    uploadDocument,
    patchDocumentScope,
    deleteDocument,
    resetAgent,
    listGaps,
    patchGap,
    statusLabel,
    type KMAgent,
    type KMDocument,
    type KMGap,
    type KMScope
  } from '$lib/api/km';

  let agents = $state<KMAgent[]>([]);
  let scopes = $state<KMScope[]>([]);
  let selectedId = $state('');
  let documents = $state<KMDocument[]>([]);
  let gaps = $state<KMGap[]>([]);
  let loading = $state(true);
  let busy = $state(false);
  let uploadScope = $state('general');
  let filterScope = $state('all');
  let fileInput: HTMLInputElement | undefined = $state();

  const selected = $derived(agents.find((a) => a.id === selectedId) || null);
  const visibleDocuments = $derived(
    filterScope === 'all' ? documents : documents.filter((d) => d.km_scope === filterScope)
  );

  // Keep upload scope aligned with the selected agent's primary retrieval scope.
  $effect(() => {
    if (!selected) return;
    const def = selected.default_scopes?.[0];
    if (def && scopes.some((s) => s.id === def)) {
      uploadScope = def;
    }
  });

  onMount(async () => {
    if (!hasRegistrationSession()) {
      goto(`${base}/login?next=${encodeURIComponent(`${base}/km`)}`);
      return;
    }
    await bootstrap();
  });

  async function bootstrap() {
    loading = true;
    try {
      const [sc, ag] = await Promise.all([listScopes(), listAgents()]);
      scopes = sc.scopes || [];
      agents = ag.agents || [];
      if (!selectedId && agents.length) selectedId = agents[0].id;
      if (scopes.length && !scopes.find((s) => s.id === uploadScope)) {
        uploadScope = scopes[0].id;
      }
      await refreshDocsAndGaps();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load knowledge');
    } finally {
      loading = false;
    }
  }

  async function refreshDocsAndGaps() {
    if (!selectedId) {
      documents = [];
      return;
    }
    const [docs, g] = await Promise.all([
      listDocuments(selectedId),
      listGaps({ status: 'open' })
    ]);
    documents = docs.documents || [];
    gaps = g.gaps || [];
    // refresh agent counts from listAgents
    try {
      const ag = await listAgents();
      agents = ag.agents || agents;
    } catch {
      /* ignore */
    }
  }

  async function selectAgent(id: string) {
    selectedId = id;
    busy = true;
    try {
      await refreshDocsAndGaps();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Failed to load documents');
    } finally {
      busy = false;
    }
  }

  async function onUpload() {
    const file = fileInput?.files?.[0];
    if (!file || !selectedId) {
      feedback.error('Choose a file and agent first');
      return;
    }
    busy = true;
    try {
      await uploadDocument(selectedId, file, uploadScope);
      if (fileInput) fileInput.value = '';
      feedback.success('Document uploaded and indexed');
      await refreshDocsAndGaps();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Upload failed');
    } finally {
      busy = false;
    }
  }

  async function onScopeChange(doc: KMDocument, scope: string) {
    if (scope === doc.km_scope) return;
    busy = true;
    try {
      await patchDocumentScope(doc.id, scope);
      feedback.success('Scope updated');
      await refreshDocsAndGaps();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Scope update failed');
    } finally {
      busy = false;
    }
  }

  async function onDelete(doc: KMDocument) {
    if (!confirm(`Delete ${doc.filename}? Embeddings will be removed.`)) return;
    busy = true;
    try {
      await deleteDocument(doc.id);
      feedback.success('Document deleted');
      await refreshDocsAndGaps();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Delete failed');
    } finally {
      busy = false;
    }
  }

  async function onReset() {
    if (!selected) return;
    if (!confirm(`Delete all knowledge for ${selected.name}? This cannot be undone.`)) return;
    busy = true;
    try {
      await resetAgent(selected.id);
      feedback.success(`${selected.name} knowledge cleared`);
      await refreshDocsAndGaps();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Reset failed');
    } finally {
      busy = false;
    }
  }

  async function onDismissGap(g: KMGap) {
    busy = true;
    try {
      await patchGap(g.id, { status: 'dismissed' });
      feedback.success('Gap dismissed');
      await refreshDocsAndGaps();
    } catch (err) {
      feedback.error(err instanceof ApiError ? err.message : 'Update failed');
    } finally {
      busy = false;
    }
  }

  function formatWhen(iso: string) {
    try {
      return new Date(iso).toLocaleString();
    } catch {
      return iso;
    }
  }
</script>

<div class="page-head">
  <div>
    <h1>Knowledge base</h1>
    <p class="muted">
      Upload FAQs for each AI agent. Scope tags match caller topics: general · billing · technical.
      <strong>Embed tip:</strong> product FAQs work best on <strong>Ava + General</strong> (default embed
      agent). If you upload under Luna/Technical, pick <strong>Luna</strong> in the chat widget or
      re-upload to Ava.
    </p>
  </div>
</div>

{#if loading}
  <p class="muted">Loading…</p>
{:else}
  <section class="card agents">
    <div class="label">Agents</div>
    <div class="chips">
      {#each agents as a}
        <button
          type="button"
          class="chip"
          class:active={a.id === selectedId}
          disabled={busy}
          onclick={() => selectAgent(a.id)}
        >
          {a.name}
          <span class="count">{a.doc_count}</span>
        </button>
      {/each}
    </div>
    {#if selected}
      <div class="overview">
        <strong>{selected.name}</strong>
        <span class="muted">
          default retrieval: {(selected.default_scopes || []).join(', ') || 'general'}
        </span>
        <div class="scope-counts">
          {#each scopes as s}
            <span>{s.id}: {selected.by_scope?.[s.id] ?? 0}</span>
          {/each}
        </div>
      </div>
    {/if}
  </section>

  <section class="card">
    <div class="label">Upload</div>
    <div class="upload-row">
      <input type="file" accept=".md,.txt,text/plain,text/markdown" bind:this={fileInput} disabled={busy} />
      <select bind:value={uploadScope} disabled={busy}>
        {#each scopes as s}
          <option value={s.id}>{s.label}</option>
        {/each}
      </select>
      <button type="button" class="btn primary" disabled={busy || !selectedId} onclick={onUpload}>
        Upload
      </button>
    </div>
  </section>

  <section class="card">
    <div class="docs-head">
      <div class="label" style="margin:0">Documents {selected ? `· ${selected.name}` : ''}</div>
      <label class="filter-scope">
        <span>Filter scope</span>
        <select bind:value={filterScope} disabled={busy}>
          <option value="all">All scopes</option>
          {#each scopes as s}
            <option value={s.id}>{s.label}</option>
          {/each}
        </select>
      </label>
    </div>
    {#if !documents.length}
      <p class="empty">
        No documents for {selected?.name || 'this agent'} yet. Upload a Markdown FAQ to ground answers.
      </p>
    {:else if !visibleDocuments.length}
      <p class="empty">
        No documents in scope <strong>{filterScope}</strong>. Change filter or upload with that scope.
      </p>
    {:else}
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Filename</th>
              <th>Scope</th>
              <th>Status</th>
              <th>Chunks</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {#each visibleDocuments as doc}
              <tr>
                <td>{doc.filename}</td>
                <td>
                  <select
                    value={doc.km_scope}
                    disabled={busy || doc.status === 'indexing' || doc.status === 'uploaded'}
                    onchange={(e) => onScopeChange(doc, (e.currentTarget as HTMLSelectElement).value)}
                  >
                    {#each scopes as s}
                      <option value={s.id}>{s.label}</option>
                    {/each}
                  </select>
                </td>
                <td>{statusLabel(doc.status)}</td>
                <td>{doc.chunk_count}</td>
                <td>
                  <button type="button" class="btn ghost danger" disabled={busy} onclick={() => onDelete(doc)}>
                    Delete
                  </button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
    {#if selected}
      <div class="footer-actions">
        <button type="button" class="btn ghost" disabled={busy} onclick={onReset}>
          Reset {selected.name} knowledge…
        </button>
      </div>
    {/if}
  </section>

  <section class="card">
    <div class="label">Knowledge gaps (unanswered)</div>
    <p class="muted small">
      Questions where AI found no KM match. Add FAQs above, then dismiss or leave open.
    </p>
    {#if !gaps.length}
      <p class="empty">No open gaps yet.</p>
    {:else}
      <div class="table-wrap">
        <table>
          <thead>
            <tr>
              <th>Question</th>
              <th>Agent</th>
              <th>Count</th>
              <th>Last seen</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {#each gaps as g}
              <tr>
                <td class="q">{g.question}</td>
                <td>{g.agent_id}</td>
                <td>{g.occurrence_count}</td>
                <td class="muted small">{formatWhen(g.last_seen_at)}</td>
                <td>
                  <button type="button" class="btn ghost" disabled={busy} onclick={() => onDismissGap(g)}>
                    Dismiss
                  </button>
                </td>
              </tr>
            {/each}
          </tbody>
        </table>
      </div>
    {/if}
  </section>
{/if}

<style>
  .page-head {
    margin-bottom: 20px;
  }
  h1 {
    margin: 0;
    font-size: 24px;
  }
  .muted {
    color: var(--muted);
    font-size: 13px;
    margin: 6px 0 0;
  }
  .muted.small,
  .small {
    font-size: 12px;
  }
  .card {
    background: rgb(8 14 28 / 80%);
    border: 1px solid var(--line);
    border-radius: 12px;
    padding: 16px;
    margin-bottom: 16px;
  }
  .label {
    font-size: 12px;
    font-weight: 600;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--muted);
    margin-bottom: 10px;
  }
  .docs-head {
    display: flex;
    flex-wrap: wrap;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    margin-bottom: 10px;
  }
  .filter-scope {
    display: flex;
    align-items: center;
    gap: 8px;
    font-size: 12px;
    color: var(--muted);
  }
  .filter-scope select {
    padding: 6px 10px;
    border-radius: 8px;
    border: 1px solid var(--line);
    background: transparent;
    color: var(--ink);
    font: inherit;
  }
  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 8px;
  }
  .chip {
    border: 1px solid var(--line);
    background: #0c1a2e;
    color: #f7fbff;
    border-radius: 999px;
    padding: 8px 12px;
    cursor: pointer;
    font-size: 13px;
  }
  .chip.active {
    border-color: #00b7ff;
    box-shadow: 0 0 0 1px rgb(0 183 255 / 40%);
  }
  .chip .count {
    margin-left: 6px;
    opacity: 0.7;
    font-variant-numeric: tabular-nums;
  }
  .overview {
    margin-top: 12px;
    display: flex;
    flex-direction: column;
    gap: 4px;
    font-size: 13px;
  }
  .scope-counts {
    display: flex;
    gap: 12px;
    flex-wrap: wrap;
    color: var(--muted);
    font-size: 12px;
  }
  .upload-row {
    display: flex;
    flex-wrap: wrap;
    gap: 10px;
    align-items: center;
  }
  select,
  input[type='file'] {
    background: #0a1628;
    color: #f7fbff;
    border: 1px solid var(--line);
    border-radius: 8px;
    padding: 8px 10px;
    font-size: 13px;
  }
  .btn {
    border: 0;
    border-radius: 8px;
    padding: 8px 14px;
    font-weight: 600;
    font-size: 13px;
    cursor: pointer;
  }
  .btn.primary {
    background: linear-gradient(135deg, #0084ff, #00b7ff);
    color: #fff;
  }
  .btn.ghost {
    background: transparent;
    border: 1px solid var(--line);
    color: #cfe6ff;
  }
  .btn.danger {
    color: #ff8a9a;
  }
  .btn:disabled {
    opacity: 0.5;
    cursor: not-allowed;
  }
  .table-wrap {
    overflow: auto;
  }
  table {
    width: 100%;
    border-collapse: collapse;
    font-size: 13px;
  }
  th,
  td {
    text-align: left;
    padding: 8px 6px;
    border-bottom: 1px solid rgb(0 183 255 / 12%);
  }
  th {
    color: var(--muted);
    font-weight: 600;
    font-size: 11px;
    text-transform: uppercase;
  }
  td.q {
    max-width: 280px;
    word-break: break-word;
  }
  .empty {
    color: var(--muted);
    font-size: 13px;
    margin: 0;
  }
  .footer-actions {
    margin-top: 12px;
  }
</style>
