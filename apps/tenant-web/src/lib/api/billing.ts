import { apiFetch } from '$lib/api/http';
import { getAccessToken } from '$lib/auth/session';

export type PackageSummary = {
  id: string;
  slug: string;
  name: string;
  description: string;
  price_cents: number;
  currency: string;
  billing_period: string;
  rules_summary: Record<string, number | boolean>;
};

export type CurrentEntitlement = {
  package_id: string;
  package_name: string;
  status: string;
} | null;

export type PaymentMethodOption = {
  id: string;
  label: string;
  channel_code: string;
};

export type TenantPackagesResponse = {
  packages: PackageSummary[];
  current_entitlement: CurrentEntitlement;
  payment_methods?: PaymentMethodOption[];
};

export type CheckoutResponse = {
  order_id: string;
  order_no: string;
  package_id: string;
  amount_cents: number;
  currency: string;
  status: string;
  payment_url: string;
  provider: string;
  payment_method?: string;
  return_url?: string;
};

export type PaymentDocument = {
  id: string;
  order_id: string;
  tenant_id?: string;
  doc_type: 'receipt' | 'tax_invoice' | string;
  doc_number: string;
  status?: string;
  buyer_name: string;
  buyer_address: string;
  buyer_tax_id?: string;
  package_name: string;
  amount_cents: number;
  currency: string;
  vat_rate_bps: number;
  net_cents: number;
  vat_cents: number;
  payment_method: string;
  issued_at: string;
};

export type PaymentOrder = {
  id: string;
  order_no: string;
  package_id: string;
  status: string;
  amount_cents: number;
  currency?: string;
  payment_method?: string;
  provider?: string;
  transaction_id?: string;
  paid_at?: string | null;
  created_at: string;
  documents?: PaymentDocument[];
};

export function getTenantPackages(): Promise<TenantPackagesResponse> {
  return apiFetch('/api/tenant/packages');
}

export function checkoutPackage(
  packageId: string,
  paymentMethod: string
): Promise<CheckoutResponse> {
  return apiFetch('/api/tenant/checkout', {
    method: 'POST',
    body: JSON.stringify({ package_id: packageId, payment_method: paymentMethod })
  });
}

export function getPaymentOrder(orderId: string): Promise<PaymentOrder> {
  return apiFetch(`/api/tenant/orders/${orderId}`);
}

export function completeMockPayment(
  orderId: string,
  result: 'paid' | 'failed' = 'paid'
): Promise<PaymentOrder> {
  return apiFetch(`/api/dev/mock-pay/${orderId}`, {
    method: 'POST',
    body: JSON.stringify({ result: result === 'failed' ? 'failed' : 'success' })
  });
}

/** Absolute path for printable document HTML (open with Authorization via new tab needs token cookie — use fetch blob). */
export function documentURL(orderId: string, docType: string): string {
  return `/api/tenant/orders/${orderId}/documents/${docType}?format=html`;
}

export async function openDocumentHTML(orderId: string, docType: string): Promise<void> {
  const headers: Record<string, string> = {};
  const token = getAccessToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  const res = await fetch(documentURL(orderId, docType), { headers });
  if (!res.ok) {
    let message = `Failed to load ${docType}`;
    try {
      const body = await res.json();
      if (body?.error) message = body.error;
    } catch {
      // ignore
    }
    throw new Error(message);
  }
  const html = await res.text();
  const blob = new Blob([html], { type: 'text/html' });
  const url = URL.createObjectURL(blob);
  window.open(url, '_blank', 'noopener,noreferrer');
  setTimeout(() => URL.revokeObjectURL(url), 60_000);
}
