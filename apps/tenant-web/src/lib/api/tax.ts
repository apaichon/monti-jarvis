import { apiFetch, handleUnauthorized } from '$lib/api/http';
import { getAccessToken } from '$lib/auth/session';
import type { PaymentDocument } from '$lib/api/billing';

export type TaxProfile = {
  tenant_id: string;
  company_name: string;
  tax_id: string;
  branch: string;
  address: string;
  invoices_refreshed?: number;
};

export function getTaxProfile(): Promise<TaxProfile> {
  return apiFetch('/api/tenant/tax-profile');
}

export function putTaxProfile(body: {
  company_name: string;
  tax_id: string;
  branch: string;
  address: string;
  refresh_invoices?: boolean;
}): Promise<TaxProfile> {
  return apiFetch('/api/tenant/tax-profile', {
    method: 'PUT',
    body: JSON.stringify(body)
  });
}

export function listBillingDocuments(): Promise<{ documents: PaymentDocument[] }> {
  return apiFetch('/api/tenant/billing/documents');
}

export async function openTenantDocumentHTML(docId: string): Promise<void> {
  const headers: Record<string, string> = {};
  const token = getAccessToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await fetch(`/api/tenant/billing/documents/${docId}?format=html`, { headers });
  if (res.status === 401 && !!token) handleUnauthorized(true);
  if (!res.ok) {
    let message = 'Failed to open document';
    try {
      const body = await res.json();
      if (body?.error) message = body.error;
    } catch {
      // ignore
    }
    throw new Error(message);
  }
  const html = await res.text();
  const url = URL.createObjectURL(new Blob([html], { type: 'text/html' }));
  window.open(url, '_blank', 'noopener,noreferrer');
  setTimeout(() => URL.revokeObjectURL(url), 60_000);
}
