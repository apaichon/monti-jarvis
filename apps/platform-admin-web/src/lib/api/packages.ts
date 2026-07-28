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
  valid_from?: string;
  valid_until?: string | null;
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

export type PromotionTaxInvoice = {
  id: string;
  doc_type: string;
  doc_number: string;
  status: string;
  amount_cents: number;
  currency: string;
  issued_at: string;
};

export type PromotionGrant = {
  id: string;
  tenant_id: string;
  package_id: string;
  order_id: string;
  reason: string;
  amount_cents: number;
  status: string;
  created_at: string;
  created_by?: string;
  valid_until?: string | null;
  idempotency_key?: string;
  entitlement?: Entitlement;
  tax_invoice?: PromotionTaxInvoice;
  tax_invoice_id?: string;
  tax_invoice_number?: string;
  replayed?: boolean;
  currency?: string;
};

export type CreatePromotionGrantBody = {
  package_id: string;
  reason: string;
  valid_until?: string;
  amount_cents?: number;
  idempotency_key?: string;
};

export function createPromotionGrant(tenantId: string, body: CreatePromotionGrantBody) {
  return apiFetch<PromotionGrant>(`/api/platform/tenants/${encodeURIComponent(tenantId)}/promotion-grants`, {
    method: 'POST',
    body: JSON.stringify(body)
  });
}

export function listPromotionGrants(tenantId: string, limit = 50) {
  const q = limit ? `?limit=${limit}` : '';
  return apiFetch<{ grants: PromotionGrant[] }>(
    `/api/platform/tenants/${encodeURIComponent(tenantId)}/promotion-grants${q}`
  );
}