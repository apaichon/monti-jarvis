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

export type TicketOffer = {
  subject: string;
  category: 'general' | 'billing' | 'technical' | 'other';
  reason: string;
};

export type ChatResponse = {
  session_id: string;
  reply: string;
  sources?: ChatSource[];
  missing_km?: boolean;
  ticket_offer?: TicketOffer;
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
  opts?: { tenantId?: string; embedKey?: string; parentOrigin?: string }
): Promise<ChatResponse> {
  const headers: Record<string, string> = customerAuthHeaders({ 'content-type': 'application/json' });
  if (opts?.tenantId) headers['X-Tenant-Id'] = opts.tenantId;
  if (opts?.embedKey) headers['X-Monti-Embed-Key'] = opts.embedKey;
  if (opts?.parentOrigin) headers['X-Embed-Parent-Origin'] = opts.parentOrigin;
  const res = await fetch('/api/chat', {
    method: 'POST',
    headers,
    body: JSON.stringify(payload)
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || data.message || 'Chat failed');
  return data;
}
import { customerAuthHeaders } from './customerAuth';
