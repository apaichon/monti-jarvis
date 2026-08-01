import { apiFetch, handleUnauthorized } from '$lib/api/http';
import { getAccessToken } from '$lib/auth/session';

export type PackageSummary = {
  id: string;
  slug: string;
  name: string;
  description: string;
  price_cents: number;
  currency: string;
  billing_period: string;
  purchase_mode: 'self_serve' | 'quote';
  deployment: 'shared_cloud' | 'dedicated_vm';
  quote_only: boolean;
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
  provider_session_id?: string;
  return_url?: string;
  subscription_id?: string;
  billing_interval?: 'monthly' | 'annual';
};

export type PriceCalculation = {
  package_id: string;
  package_version_id: string;
  package_name: string;
  deployment_mode: 'shared_cloud' | 'dedicated_vm';
  purchase_mode: 'self_serve' | 'quote';
  billing_interval: 'monthly' | 'annual';
  base_price_cents: number;
  addons_cents: number;
  setup_fees_cents: number;
  proration_cents: number;
  subtotal_cents: number;
  discount_cents: number;
  credits_cents: number;
  taxable_amount_cents: number;
  tax_cents: number;
  amount_due_cents: number;
  currency: string;
  tax_rate_bps: number;
  annual_discount_bps: number;
  checkout_eligible: boolean;
  quote_required: boolean;
  calculated_at: string;
};

export type DedicatedQuoteInput = {
  package_id: string;
  company_legal_name: string;
  contact_name: string;
  contact_email: string;
  contact_phone: string;
  tax_registration_id?: string;
  company_size?: '1-10' | '11-50' | '51-200' | '201-500' | '501+' | '';
  expected_concurrency: number;
  preferred_region?: 'th-bangkok' | 'sg-singapore' | 'jp-tokyo' | 'eu-frankfurt' | 'other' | '';
  notes?: string;
  idempotency_key?: string;
};

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
  status:
    | 'submitted'
    | 'under_review'
    | 'capacity_confirmed'
    | 'quoted'
    | 'accepted'
    | 'provisioning'
    | 'active'
    | 'rejected'
    | 'expired'
    | 'withdrawn';
  quoted_amount_cents?: number;
  currency: string;
  quote_expires_at?: string;
  capacity_snapshot: Record<string, unknown>;
  created_at: string;
  updated_at: string;
};

export type CurrentPlanQuotaDimension = {
  dimension: string;
  unit: string;
  period: string;
  unlimited: boolean;
  limit: number | null;
  used: number | null;
  remaining: number | null;
  utilization: number | null;
  source: string;
  freshness: 'current' | 'stale' | 'unavailable' | string;
};

export type CurrentCommercialPlan = {
  tenant_id: string;
  billing_state:
    | 'no_plan'
    | 'no_scheduled_bill'
    | 'quote_pending'
    | 'scheduled'
    | 'past_due'
    | 'grace'
    | 'suspended'
    | string;
  package: {
    id: string;
    slug: string;
    name: string;
    status: string;
    deployment_mode: 'shared_cloud' | 'dedicated_vm';
    purchase_mode: 'self_serve' | 'quote';
    entitlement_id: string;
    valid_from: string;
    valid_until: string | null;
  } | null;
  subscription: {
    id: string;
    status: string;
    billing_interval: 'monthly' | 'annual';
    current_period_start: string;
    current_period_end: string;
    billing_anchor: string;
    grace_until: string | null;
  } | null;
  next_bill: {
    at: string;
    amount_cents: number;
    currency: string;
    state: string;
  } | null;
  quota: CurrentPlanQuotaDimension[];
  compact_utilization: number | null;
  documents: Array<{
    id: string;
    type: string;
    number: string;
    issued_at: string;
    amount_cents: number;
    currency: string;
    href: string;
  }>;
  quote: DedicatedQuote | null;
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
  provider_session_id?: string;
  provider_payment_id?: string;
  provider_status?: string;
  checkout_expires_at?: string | null;
  last_provider_sync_at?: string | null;
  paid_at?: string | null;
  created_at: string;
  documents?: PaymentDocument[];
};

export function getTenantPackages(): Promise<TenantPackagesResponse> {
  return apiFetch('/api/tenant/packages');
}

export function checkoutPackage(
  packageId: string,
  paymentMethod: string,
  billingInterval: 'monthly' | 'annual' = 'monthly'
): Promise<CheckoutResponse> {
  return apiFetch('/api/tenant/checkout', {
    method: 'POST',
    body: JSON.stringify({
      package_id: packageId,
      payment_method: paymentMethod,
      billing_interval: billingInterval
    })
  });
}

export function calculatePackage(
  packageId: string,
  billingInterval: 'monthly' | 'annual'
): Promise<PriceCalculation> {
  return apiFetch('/api/tenant/commercial/calculate', {
    method: 'POST',
    body: JSON.stringify({ package_id: packageId, billing_interval: billingInterval })
  });
}

export function createDedicatedQuote(
  input: DedicatedQuoteInput
): Promise<{ quote: DedicatedQuote; replayed: boolean }> {
  return apiFetch('/api/tenant/commercial/quotes', {
    method: 'POST',
    body: JSON.stringify(input)
  });
}

export function listDedicatedQuotes(): Promise<{ quotes: DedicatedQuote[] }> {
  return apiFetch('/api/tenant/commercial/quotes');
}

export function getCurrentCommercialPlan(): Promise<CurrentCommercialPlan> {
  return apiFetch('/api/tenant/commercial/current-plan');
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
  if (res.status === 401 && !!token) handleUnauthorized(true);
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
