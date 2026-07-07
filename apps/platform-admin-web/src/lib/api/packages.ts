import { apiFetch } from './http';

export type RuleFieldSpec = {
  type: string;
  min?: number;
  max?: number;
  required?: boolean;
  default?: boolean | number;
  description?: string;
};

export type RuleSchema = {
  id: string;
  version: number;
  name: string;
  status: string;
  fields: Record<string, RuleFieldSpec>;
};

export type Package = {
  id: string;
  slug: string;
  name: string;
  description?: string;
  status: string;
  price_cents: number;
  currency: string;
  billing_period: string;
  rules_schema_id: string;
  rules: Record<string, boolean | number>;
};

export type Entitlement = {
  tenant_id: string;
  package: { id: string; slug: string; name: string };
  status: string;
  rules_schema_id: string;
  rules: Record<string, boolean | number>;
  valid_from: string;
  valid_until: string | null;
};

export function listRuleSchemas() {
  return apiFetch<{ schemas: RuleSchema[] }>('/api/platform/rule-schemas');
}

export function listPackages(status = '') {
  const q = status ? `?status=${encodeURIComponent(status)}` : '';
  return apiFetch<{ packages: Package[] }>(`/api/platform/packages${q}`);
}

export function getPackage(id: string) {
  return apiFetch<Package>(`/api/platform/packages/${id}`);
}

export function createPackage(body: Partial<Package>) {
  return apiFetch<Package>('/api/platform/packages', {
    method: 'POST',
    body: JSON.stringify(body)
  });
}

export function updatePackage(id: string, body: Partial<Package>) {
  return apiFetch<Package>(`/api/platform/packages/${id}`, {
    method: 'PUT',
    body: JSON.stringify(body)
  });
}

export function archivePackage(id: string) {
  return apiFetch<{ status: string }>(`/api/platform/packages/${id}`, { method: 'DELETE' });
}

export function getTenantEntitlement(tenantId: string) {
  return apiFetch<Entitlement>(`/api/platform/tenants/${tenantId}/entitlement`);
}

export function assignTenantEntitlement(tenantId: string, packageId: string) {
  return apiFetch<Entitlement>(`/api/platform/tenants/${tenantId}/entitlement`, {
    method: 'POST',
    body: JSON.stringify({ package_id: packageId })
  });
}

export function revokeTenantEntitlement(tenantId: string) {
  return apiFetch<{ status: string }>(`/api/platform/tenants/${tenantId}/entitlement`, {
    method: 'DELETE'
  });
}