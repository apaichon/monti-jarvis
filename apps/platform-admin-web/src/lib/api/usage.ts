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

export type TenantUsage = {
  tenant_id: string;
  package: UsagePackage | null;
  status: string;
  period: string;
  limits: UsageLimits | null;
  usage: UsageCounts;
};

export function getTenantUsage(tenantId: string) {
  return apiFetch<TenantUsage>(`/api/platform/tenants/${encodeURIComponent(tenantId)}/usage`);
}
