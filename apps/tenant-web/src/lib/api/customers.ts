import { apiFetch } from './http';

export type Customer = {
  id: string;
  email?: string;
  phone?: string;
  display_name: string;
  locale?: string;
  tier_id?: string;
  group_ids: string[];
  source: string;
  external_id?: string;
  status: 'active' | 'inactive';
  metadata: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

export type CustomerInput = {
  email?: string;
  phone?: string;
  display_name: string;
  locale?: string;
  tier_id?: string;
  group_ids?: string[];
  source?: string;
  external_id?: string;
  status?: 'active' | 'inactive';
  metadata?: Record<string, unknown>;
};

export type ImportRowError = {
  row: number;
  field: string;
  code: string;
  message: string;
};

export type CustomerImport = {
  id: string;
  filename: string;
  mode: 'dry_run' | 'commit';
  status: 'validating' | 'validated' | 'completed' | 'failed';
  total_rows: number;
  accepted_rows: number;
  created_rows: number;
  updated_rows: number;
  rejected_rows: number;
  errors: ImportRowError[];
  created_at: string;
  updated_at: string;
};

export type DomainRule = {
  id: string;
  domain: string;
  policy: 'allow' | 'deny';
  default_tier_id?: string;
  default_group_id?: string;
  active: boolean;
  created_at: string;
  updated_at: string;
};

export type DomainRuleInput = {
  domain: string;
  policy: 'allow' | 'deny';
  default_tier_id?: string;
  default_group_id?: string;
  active?: boolean;
};

export function listCustomers(filters: { q?: string; status?: string; tier_id?: string } = {}) {
  const params = new URLSearchParams();
  if (filters.q) params.set('q', filters.q);
  if (filters.status) params.set('status', filters.status);
  if (filters.tier_id) params.set('tier_id', filters.tier_id);
  const query = params.toString();
  return apiFetch<{ customers: Customer[]; next_cursor: string }>(
    `/api/tenant/customers${query ? `?${query}` : ''}`
  );
}

export function createCustomer(body: CustomerInput) {
  return apiFetch<{ customer: Customer; outcome: 'created' | 'updated' }>(
    '/api/tenant/customers',
    { method: 'POST', body: JSON.stringify(body) }
  );
}

export function updateCustomer(id: string, body: CustomerInput) {
  return apiFetch<Customer>(`/api/tenant/customers/${encodeURIComponent(id)}`, {
    method: 'PUT',
    body: JSON.stringify(body)
  });
}

export function deactivateCustomer(id: string) {
  return apiFetch<{ id: string; status: 'inactive' }>(
    `/api/tenant/customers/${encodeURIComponent(id)}`,
    { method: 'DELETE' }
  );
}

export function importCustomers(file: File, dryRun: boolean) {
  const body = new FormData();
  body.set('file', file);
  body.set('dry_run', String(dryRun));
  body.set('source', 'csv');
  return apiFetch<CustomerImport>('/api/tenant/customer-imports', { method: 'POST', body });
}

export function listDomainRules() {
  return apiFetch<{ rules: DomainRule[] }>('/api/tenant/customer-domain-rules');
}

export function createDomainRule(body: DomainRuleInput) {
  return apiFetch<DomainRule>('/api/tenant/customer-domain-rules', {
    method: 'POST', body: JSON.stringify(body)
  });
}

export function updateDomainRule(id: string, body: DomainRuleInput) {
  return apiFetch<DomainRule>(`/api/tenant/customer-domain-rules/${encodeURIComponent(id)}`, {
    method: 'PUT', body: JSON.stringify(body)
  });
}

export function deleteDomainRule(id: string) {
  return apiFetch<{ deleted: boolean }>(
    `/api/tenant/customer-domain-rules/${encodeURIComponent(id)}`,
    { method: 'DELETE' }
  );
}
