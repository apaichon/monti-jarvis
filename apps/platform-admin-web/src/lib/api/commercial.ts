import { apiFetch } from '$lib/api/http';

export type DedicatedQuote = {
  id: string;
  tenant_id: string;
  package_version_id: string;
  package_id: string;
  package_name: string;
  company_legal_name: string;
  contact_name: string;
  contact_email: string;
  contact_phone: string;
  tax_registration_id: string;
  company_size: string;
  expected_concurrency: number;
  preferred_region: string;
  notes: string;
  status: string;
  quoted_amount_cents?: number;
  currency: string;
  quote_expires_at?: string;
  capacity_snapshot: Record<string, unknown>;
  created_at: string;
  updated_at: string;
  created_by: string;
  updated_by: string;
};

export type QuoteTransition = {
  status: string;
  quoted_amount_cents?: number;
  currency?: string;
  quote_expires_at?: string;
  capacity_snapshot?: Record<string, unknown>;
};

export function listDedicatedQuotes(filters?: {
  tenantId?: string;
  status?: string;
}): Promise<{ quotes: DedicatedQuote[] }> {
  const query = new URLSearchParams();
  if (filters?.tenantId) query.set('tenant_id', filters.tenantId);
  if (filters?.status) query.set('status', filters.status);
  query.set('limit', '200');
  return apiFetch(`/api/platform/commercial/quotes?${query.toString()}`);
}

export function transitionDedicatedQuote(
  id: string,
  input: QuoteTransition
): Promise<DedicatedQuote> {
  return apiFetch(`/api/platform/commercial/quotes/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(input)
  });
}
