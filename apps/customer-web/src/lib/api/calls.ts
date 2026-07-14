import { customerAuthHeaders } from './customerAuth';

export type CallSession = {
  id: string;
  tenant_id: string;
  room_name: string;
  status: string;
  started_at: string;
  ended_at?: string;
};

export type CallToken = {
  token: string;
  url: string;
  identity: string;
  room_name: string;
};

export type CallTurn = {
  id: number;
  role: string;
  content: string;
  created_at: string;
};

export type CallAudioArchiveStream = {
  name: string;
  content_type: string;
  data_base64: string;
};

export type CallRating = {
  score: number;
  review?: string;
};

const API = '';

function tenantHeaders(tenantId?: string, json = false): Record<string, string> {
  const headers: Record<string, string> = {};
  if (json) headers['content-type'] = 'application/json';
  if (tenantId) headers['X-Tenant-Id'] = tenantId;
  return customerAuthHeaders(headers);
}

export async function createCall(opts?: { tenantId?: string; agentId?: string }): Promise<CallSession> {
  const res = await fetch(`${API}/api/calls`, {
    method: 'POST',
    headers: tenantHeaders(opts?.tenantId, true),
    body: JSON.stringify({ tenant_id: opts?.tenantId, agent_id: opts?.agentId })
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Failed to create call');
  return data;
}

export async function issueToken(callId: string, opts?: { tenantId?: string }): Promise<CallToken> {
  const res = await fetch(`${API}/api/calls/${callId}/token`, {
    method: 'POST',
    headers: tenantHeaders(opts?.tenantId, true),
    body: JSON.stringify({})
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Failed to get LiveKit token');
  return data;
}

export async function endCall(callId: string, opts?: { tenantId?: string }): Promise<CallSession> {
  const res = await fetch(`${API}/api/calls/${callId}/end`, {
    method: 'POST',
    headers: tenantHeaders(opts?.tenantId)
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Failed to end call');
  return data;
}

export async function archiveCallAudio(
  callId: string,
  streams: CallAudioArchiveStream[],
  opts?: { tenantId?: string }
): Promise<void> {
  if (streams.length === 0) return;
  const res = await fetch(`${API}/api/calls/${callId}/audio`, {
    method: 'POST',
    headers: tenantHeaders(opts?.tenantId, true),
    body: JSON.stringify({ streams })
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Failed to archive call audio');
}

export async function submitCallRating(callId: string, rating: CallRating, opts?: { tenantId?: string }) {
  const res = await fetch(`${API}/api/calls/${callId}/rating`, {
    method: 'POST',
    headers: tenantHeaders(opts?.tenantId, true),
    body: JSON.stringify(rating)
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Failed to save call rating');
}

export async function listTurns(callId: string, opts?: { tenantId?: string }): Promise<CallTurn[]> {
  const res = await fetch(`${API}/api/calls/${callId}/turns`, {
    headers: tenantHeaders(opts?.tenantId)
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Failed to load transcript');
  return data.turns ?? [];
}

export async function addTurn(
  callId: string,
  role: string,
  content: string,
  opts?: { tenantId?: string }
): Promise<void> {
  const res = await fetch(`${API}/api/calls/${callId}/turns`, {
    method: 'POST',
    headers: tenantHeaders(opts?.tenantId, true),
    body: JSON.stringify({ role, content })
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Failed to add transcript turn');
}

export function subscribeTurns(
  callId: string,
  onTurn: (turn: CallTurn) => void,
  opts?: { tenantId?: string }
): () => void {
  const qs = opts?.tenantId ? `?tenant_id=${encodeURIComponent(opts.tenantId)}` : '';
  const source = new EventSource(`${API}/api/calls/${callId}/events${qs}`);
  source.addEventListener('turn', (event) => {
    onTurn(JSON.parse(event.data));
  });
  return () => source.close();
}
