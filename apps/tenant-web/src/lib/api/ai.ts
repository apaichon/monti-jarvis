import { apiFetch } from '$lib/api/http';

export type TenantGeminiKey = {
  tenant_id?: string;
  configured: boolean;
  last4?: string;
  key_version?: string;
  updated_at?: string;
};

export type TenantPrompt = {
  agent_id: string;
  enabled: boolean;
  system_prompt: string;
  max_length: number;
  updated_at?: string;
};

export type TenantTool = {
  id: string;
  tenant_id: string;
  tool_key: string;
  display_name: string;
  description: string;
  handler_key: string;
  input_schema: Record<string, unknown>;
  enabled: boolean;
};

export type TenantSkill = {
  id: string;
  tenant_id: string;
  slug: string;
  name: string;
  prompt: string;
  enabled: boolean;
  tool_ids?: string[];
  agent_ids?: string[];
};

export function getTenantGeminiKey() {
  return apiFetch<TenantGeminiKey>('/api/tenant/ai/gemini-key');
}

export function putTenantGeminiKey(api_key: string) {
  return apiFetch<TenantGeminiKey>('/api/tenant/ai/gemini-key', {
    method: 'PUT',
    body: JSON.stringify({ api_key })
  });
}

export function deleteTenantGeminiKey() {
  return apiFetch<{ configured: false }>('/api/tenant/ai/gemini-key', { method: 'DELETE' });
}

export function getTenantPrompt(agentId: string) {
  return apiFetch<TenantPrompt>(`/api/tenant/ai/prompts/${encodeURIComponent(agentId)}`);
}

export function putTenantPrompt(agentId: string, body: { system_prompt: string; enabled: boolean }) {
  return apiFetch<TenantPrompt>(`/api/tenant/ai/prompts/${encodeURIComponent(agentId)}`, {
    method: 'PUT',
    body: JSON.stringify(body)
  });
}

export function listTenantTools() {
  return apiFetch<{ tools: TenantTool[] }>('/api/tenant/ai/tools');
}

export function createTenantTool(body: Omit<TenantTool, 'id' | 'tenant_id'>) {
  return apiFetch<TenantTool>('/api/tenant/ai/tools', {
    method: 'POST',
    body: JSON.stringify(body)
  });
}

export function listTenantSkills() {
  return apiFetch<{ skills: TenantSkill[] }>('/api/tenant/ai/skills');
}

export function createTenantSkill(body: {
  slug: string;
  name: string;
  prompt: string;
  tool_ids: string[];
  agent_ids: string[];
  enabled: boolean;
}) {
  return apiFetch<TenantSkill>('/api/tenant/ai/skills', {
    method: 'POST',
    body: JSON.stringify(body)
  });
}
