import { apiFetch } from '$lib/api/http';

export type UsagePackage = {
  id: string;
  slug: string;
  name: string;
};

export type UsageLimits = {
  max_ai_employees: number;
  max_monthly_call_minutes: number;
  max_km_documents: number;
  max_concurrent_calls: number;
  voice_enabled: boolean;
  rag_enabled: boolean;
};

export type UsageCounts = {
  ai_employees: number;
  monthly_call_minutes: number;
  km_documents: number;
  concurrent_calls: number;
};

export type UsageDimension = {
  dimension: string;
  base_limit: number;
  total_limit: number;
  bonus_granted: number;
  bonus_used: number;
  bonus_remaining: number;
  consumed?: number | null;
};

export type TenantUsage = {
  tenant_id: string;
  package: UsagePackage | null;
  status: string;
  period: string;
  limits: UsageLimits | null;
  usage: UsageCounts;
  current_dimensions?: UsageDimension[];
};

export function getTenantUsage(tenantId: string) {
  return apiFetch<TenantUsage>(`/api/platform/tenants/${encodeURIComponent(tenantId)}/usage`);
}
