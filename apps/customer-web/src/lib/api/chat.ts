export type ChatMessage = {
  role: 'user' | 'assistant';
  content: string;
};

export type ChatSource = {
  chunk_id: string;
  document_id: string;
  scope: string;
  excerpt: string;
  score: number;
};

export type ChatResponse = {
  session_id: string;
  reply: string;
  sources?: ChatSource[];
  missing_km?: boolean;
  error?: string;
};

export async function sendChat(
  payload: {
    session_id: string;
    agent_id: string;
    topic: string;
    message: string;
    history: ChatMessage[];
  },
  opts?: { tenantId?: string }
): Promise<ChatResponse> {
  const headers: Record<string, string> = { 'content-type': 'application/json' };
  if (opts?.tenantId) headers['X-Tenant-Id'] = opts.tenantId;
  const res = await fetch('/api/chat', {
    method: 'POST',
    headers,
    body: JSON.stringify(payload)
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || data.message || 'Chat failed');
  return data;
}