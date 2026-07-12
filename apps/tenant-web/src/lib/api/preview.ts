import { apiFetch } from './http';
import { getAccessToken, getStoredUser } from '$lib/auth/session';

export type PreviewScenario = {
  id: string;
  topic: string;
  label: string;
  label_th?: string;
  question: string;
};

export type PreviewChatMessage = {
  role: 'user' | 'assistant';
  content: string;
};

export type PreviewChatResult = {
  session_id: string;
  agent_id: string;
  reply: string;
  sources?: { document_id?: string; chunk_id?: string; score?: number; excerpt?: string }[];
  missing_km?: boolean;
  mode: string;
  tenant_id: string;
};

/** Full agent shape for embed-like portrait UI. */
export type PreviewAgent = {
  id: string;
  name: string;
  role?: string;
  trait?: string;
  color?: string;
  image?: string;
  speaking_image?: string;
  expressions?: Record<string, string>;
  greeting?: string;
};

export function listScenarios() {
  return apiFetch<{ scenarios: PreviewScenario[] }>('/api/tenant/preview/scenarios');
}

/** Prefer public workforce (with portraits) scoped by tenant id. */
export async function listPreviewAgents(): Promise<{ agents: PreviewAgent[] }> {
  const headers: Record<string, string> = {};
  const token = getAccessToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  const tid = getStoredUser()?.tenant_id;
  if (tid) headers['X-Tenant-Id'] = tid;

  // Full avatar portraits from workforce catalog.
  try {
    const res = await fetch('/api/workforce', { headers });
    if (res.ok) {
      const data = (await res.json()) as { agents?: PreviewAgent[] };
      if (data.agents?.length) return { agents: data.agents };
    }
  } catch {
    /* fall through */
  }

  // Fallback: KM agent list (ids/names only).
  try {
    const km = await apiFetch<{ agents: PreviewAgent[] }>('/api/tenant/km/agents');
    return { agents: km.agents || [] };
  } catch {
    return {
      agents: [
        { id: 'ava', name: 'Ava', role: 'Reception', color: '#00b7ff', image: '/images/ava.jpg' },
        { id: 'max', name: 'Max', role: 'Billing', color: '#7c5cff', image: '/images/max.jpg' },
        { id: 'luna', name: 'Luna', role: 'Technical', color: '#3dd68c', image: '/images/luna.jpg' },
        { id: 'neo', name: 'Neo', role: 'Triage', color: '#ffb86c', image: '/images/neo.jpg' }
      ]
    };
  }
}

export function previewChat(body: {
  agent_id: string;
  topic: string;
  message: string;
  session_id?: string;
  history?: PreviewChatMessage[];
  lang?: string;
}) {
  return apiFetch<PreviewChatResult>('/api/tenant/preview/chat', {
    method: 'POST',
    body: JSON.stringify(body)
  });
}

/** Build preview voice WebSocket URL with access_token for auth. */
export function previewVoiceURL(agentId: string, topic: string, lang?: string): string {
  const scheme = location.protocol === 'https:' ? 'wss' : 'ws';
  const params = new URLSearchParams({
    agent: agentId,
    topic: topic || 'general'
  });
  if (lang) params.set('lang', lang);
  const token = getAccessToken();
  if (token) params.set('access_token', token);
  const tid = getStoredUser()?.tenant_id;
  if (tid) params.set('tenant_id', tid);
  return `${scheme}://${location.host}/ws/tenant/preview/voice?${params}`;
}
