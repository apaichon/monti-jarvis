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

export async function sendChat(payload: {
  session_id: string;
  agent_id: string;
  topic: string;
  message: string;
  history: ChatMessage[];
}): Promise<ChatResponse> {
  const res = await fetch('/api/chat', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify(payload)
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || 'Chat failed');
  return data;
}