export type ChatMessage = {
  role: 'user' | 'assistant';
  content: string;
};

export type ChatResponse = {
  session_id: string;
  reply: string;
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