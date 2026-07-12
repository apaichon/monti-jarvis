import { apiFetch } from '$lib/api/http';

export type KMScope = { id: string; label: string };

export type KMAgent = {
  id: string;
  name: string;
  role: string;
  doc_count: number;
  chunk_count: number;
  by_scope: Record<string, number>;
  default_scopes: string[];
  assigned: boolean;
};

export type KMDocument = {
  id: string;
  tenant_id: string;
  agent_id: string;
  filename: string;
  mime: string;
  status: string;
  km_scope: string;
  km_version: number;
  chunk_count: number;
  created_at?: string;
  updated_at?: string;
};

export type KMGap = {
  id: string;
  tenant_id: string;
  agent_id: string;
  topic: string;
  question: string;
  source: string;
  status: string;
  occurrence_count: number;
  last_seen_at: string;
  notes?: string;
  resolved_document_id?: string;
};

export function listScopes() {
  return apiFetch<{ scopes: KMScope[] }>('/api/tenant/km/scopes');
}

export function listAgents() {
  return apiFetch<{ agents: KMAgent[] }>('/api/tenant/km/agents');
}

export function listDocuments(agentId: string) {
  return apiFetch<{ agent_id: string; documents: KMDocument[] }>(
    `/api/tenant/km/agents/${encodeURIComponent(agentId)}/documents`
  );
}

export function uploadDocument(agentId: string, file: File, scope: string) {
  const body = new FormData();
  body.append('file', file);
  body.append('scope', scope);
  return apiFetch<KMDocument>(`/api/tenant/km/agents/${encodeURIComponent(agentId)}/documents`, {
    method: 'POST',
    body
  });
}

export function patchDocumentScope(id: string, km_scope: string) {
  return apiFetch<KMDocument>(`/api/tenant/km/documents/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    body: JSON.stringify({ km_scope })
  });
}

export function deleteDocument(id: string) {
  return apiFetch<{ deleted: boolean; id: string }>(
    `/api/tenant/km/documents/${encodeURIComponent(id)}`,
    { method: 'DELETE' }
  );
}

export function resetAgent(agentId: string) {
  return apiFetch<{ agent_id: string; status: string }>(
    `/api/tenant/km/agents/${encodeURIComponent(agentId)}/reset`,
    { method: 'POST', body: '{}' }
  );
}

export function listGaps(params?: { status?: string; agent_id?: string }) {
  const q = new URLSearchParams();
  if (params?.status) q.set('status', params.status);
  if (params?.agent_id) q.set('agent_id', params.agent_id);
  const qs = q.toString();
  return apiFetch<{ gaps: KMGap[] }>(`/api/tenant/km/gaps${qs ? `?${qs}` : ''}`);
}

export function patchGap(id: string, body: { status: string; notes?: string; resolved_document_id?: string }) {
  return apiFetch<KMGap>(`/api/tenant/km/gaps/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    body: JSON.stringify(body)
  });
}

export function statusLabel(status: string): string {
  switch (status) {
    case 'indexed':
      return 'Ready';
    case 'uploaded':
    case 'indexing':
      return 'Processing';
    case 'failed':
      return 'Failed';
    default:
      return status || '—';
  }
}
