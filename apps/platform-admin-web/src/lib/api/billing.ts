import { apiFetch } from '$lib/api/http';
import { getAccessToken } from '$lib/auth/session';

export type BillingOrder = {
  id: string;
  order_no: string;
  tenant_id?: string;
  tenant_name?: string;
  package_id: string;
  package_name?: string;
  status: string;
  amount_cents: number;
  currency?: string;
  payment_method?: string;
  provider?: string;
  transaction_id?: string;
  paid_at?: string | null;
  created_at: string;
  documents?: BillingDocument[];
};

export type BillingDocument = {
  id: string;
  order_id: string;
  tenant_id?: string;
  doc_type: string;
  doc_number: string;
  status: string;
  buyer_name: string;
  buyer_tax_id?: string;
  package_name: string;
  amount_cents: number;
  currency: string;
  issued_at: string;
  void_reason?: string;
  voided_at?: string | null;
};

export type SellerBranding = {
  id: string;
  name: string;
  address: string;
  tax_id: string;
  branch: string;
};

export function listBillingOrders(params: {
  tenant_id?: string;
  status?: string;
}): Promise<{ orders: BillingOrder[] }> {
  const q = new URLSearchParams();
  if (params.tenant_id) q.set('tenant_id', params.tenant_id);
  if (params.status) q.set('status', params.status);
  const qs = q.toString();
  return apiFetch(`/api/platform/billing/orders${qs ? `?${qs}` : ''}`);
}

export function getBillingOrder(id: string): Promise<BillingOrder> {
  return apiFetch(`/api/platform/billing/orders/${id}`);
}

export function listBillingDocuments(params: {
  tenant_id?: string;
  doc_type?: string;
  status?: string;
}): Promise<{ documents: BillingDocument[] }> {
  const q = new URLSearchParams();
  if (params.tenant_id) q.set('tenant_id', params.tenant_id);
  if (params.doc_type) q.set('doc_type', params.doc_type);
  if (params.status) q.set('status', params.status);
  const qs = q.toString();
  return apiFetch(`/api/platform/billing/documents${qs ? `?${qs}` : ''}`);
}

export function voidDocument(id: string, reason: string): Promise<BillingDocument> {
  return apiFetch(`/api/platform/billing/documents/${id}/void`, {
    method: 'POST',
    body: JSON.stringify({ reason })
  });
}

export function reissueDocument(id: string, reason?: string): Promise<BillingDocument> {
  return apiFetch(`/api/platform/billing/documents/${id}/reissue`, {
    method: 'POST',
    body: JSON.stringify({ reason: reason || 'reissued by platform admin' })
  });
}

export function getSellerBranding(): Promise<SellerBranding> {
  return apiFetch('/api/platform/billing/seller-branding');
}

export function putSellerBranding(body: Omit<SellerBranding, 'id'>): Promise<SellerBranding> {
  return apiFetch('/api/platform/billing/seller-branding', {
    method: 'PUT',
    body: JSON.stringify(body)
  });
}

export async function openPlatformDocumentHTML(docId: string): Promise<void> {
  const headers: Record<string, string> = {};
  const token = getAccessToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await fetch(`/api/platform/billing/documents/${docId}?format=html`, { headers });
  if (!res.ok) throw new Error('Failed to open document');
  const html = await res.text();
  const url = URL.createObjectURL(new Blob([html], { type: 'text/html' }));
  window.open(url, '_blank', 'noopener,noreferrer');
  setTimeout(() => URL.revokeObjectURL(url), 60_000);
}
