import { apiFetch } from './http';

export type CustomerTier = {
  id: string;
  tenant_id: string;
  name: string;
  slug: string;
  priority: number;
  description?: string;
  default_agent_id?: string;
  ai_reply_locale?: string;
  max_minutes_per_call: number;
  max_call_minutes_per_day: number;
  active: boolean;
  created_at?: string;
  updated_at?: string;
};

export type CustomerGroup = {
  id: string;
  tenant_id: string;
  name: string;
  slug: string;
  description?: string;
  created_at?: string;
  updated_at?: string;
};

export type TierInput = {
  name: string;
  slug?: string;
  priority?: number;
  description?: string;
  default_agent_id?: string;
  ai_reply_locale?: string;
  max_minutes_per_call?: number;
  max_call_minutes_per_day?: number;
  active?: boolean;
};

export function listTiers() {
  return apiFetch<{ tiers: CustomerTier[] }>('/api/tenant/tiers');
}

export function createTier(body: TierInput) {
  return apiFetch<CustomerTier>('/api/tenant/tiers', {
    method: 'POST',
    body: JSON.stringify(body)
  });
}

export function updateTier(id: string, body: Partial<TierInput>) {
  return apiFetch<CustomerTier>(`/api/tenant/tiers/${encodeURIComponent(id)}`, {
    method: 'PUT',
    body: JSON.stringify(body)
  });
}

export function deleteTier(id: string) {
  return apiFetch<{ deleted: boolean }>(`/api/tenant/tiers/${encodeURIComponent(id)}`, {
    method: 'DELETE'
  });
}

export function listGroups() {
  return apiFetch<{ groups: CustomerGroup[] }>('/api/tenant/groups');
}

export function createGroup(body: { name: string; slug?: string; description?: string }) {
  return apiFetch<CustomerGroup>('/api/tenant/groups', {
    method: 'POST',
    body: JSON.stringify(body)
  });
}

export function deleteGroup(id: string) {
  return apiFetch<{ deleted: boolean }>(`/api/tenant/groups/${encodeURIComponent(id)}`, {
    method: 'DELETE'
  });
}
