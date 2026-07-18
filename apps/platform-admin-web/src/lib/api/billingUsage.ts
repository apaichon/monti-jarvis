import { apiFetch } from '$lib/api/http';

export type BillingUsageResponse = {
  range: { start_date: string; end_date: string; timezone: string };
  freshness: { status: string; generated_at: string; activity_last_projected_at?: string; ai_usage_last_projected_at?: string };
  billing: { paid_orders: number; paid_amount_minor: number; currency: string; status: string };
  quota: { reporting_minutes: number; enforcement: { status: string; monthly_used: number; monthly_limit: number; daily_used: number; daily_limit: number; daily_status?: string } };
  ai_cost: { rate_version: string; pricing_as_of?: string; currency: string; observed_cost_microunits: number; estimated_cost_microunits: number; observed_events: number; estimated_events: number; unavailable_events: number; coverage_percent: number; status: string };
  reconciliation: { activity_quota: string; orders_entitlements: string; ai_coverage: string };
  tenants: BillingUsageTenant[];
  pagination: { total: number; limit: number; offset: number };
};

export type BillingUsageTenant = {
  tenant_id: string;
  slug: string;
  name: string;
  package: { name: string; status: string };
  paid_orders: number;
  paid_amount_minor: number;
  currency: string;
  reporting_minutes: number;
  quota: { status: string; monthly_used: number; monthly_limit: number; daily_used: number; daily_limit: number; daily_status?: string };
  ai_observed_cost_microunits: number;
  ai_estimated_cost_microunits: number;
  ai_observed_events: number;
  ai_estimated_events: number;
  ai_unavailable_events: number;
  ai_coverage_percent: number;
  status: string;
};

export function getBillingUsage(params: { start_date: string; end_date: string; tenant_id?: string; limit?: number; offset?: number }): Promise<BillingUsageResponse> {
  const query = new URLSearchParams({ start_date: params.start_date, end_date: params.end_date });
  if (params.tenant_id) query.set('tenant_id', params.tenant_id);
  if (params.limit !== undefined) query.set('limit', String(params.limit));
  if (params.offset !== undefined) query.set('offset', String(params.offset));
  return apiFetch(`/api/platform/billing/usage?${query.toString()}`);
}
