import { apiFetch } from '$lib/api/http';

export type EmbedConfig = {
  tenant_id: string;
  embed_key: string;
  enabled: boolean;
  auth_required?: boolean;
  allowed_origins: string[];
  default_agent_id?: string;
  snippet: string;
  created_at?: string;
  updated_at?: string;
};

export function getEmbedConfig() {
  return apiFetch<EmbedConfig>('/api/tenant/embed');
}

export function putEmbedConfig(body: {
  enabled?: boolean;
  auth_required?: boolean;
  allowed_origins?: string[];
  default_agent_id?: string;
}) {
  return apiFetch<EmbedConfig>('/api/tenant/embed', {
    method: 'PUT',
    body: JSON.stringify(body)
  });
}

export function rotateEmbedKey() {
  return apiFetch<EmbedConfig>('/api/tenant/embed/rotate-key', {
    method: 'POST',
    body: '{}'
  });
}
