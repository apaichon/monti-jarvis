import { apiFetch } from './http';

export type TenantSettings = {
  tenant_id: string;
  locale: string;
  timezone: string;
  display_name?: string;
  ai_reply_locale?: string;
  user_tier_label?: string;
  user_group_label?: string;
  created_at?: string;
  updated_at?: string;
};

export type TenantCallLimits = {
  tenant_id: string;
  max_minutes_per_call: number;
  max_call_minutes_per_day: number;
  created_at?: string;
  updated_at?: string;
};

export type PackageLimits = {
  max_ai_employees: number;
  max_monthly_call_minutes: number;
  max_km_documents: number;
  max_concurrent_calls: number;
  voice_enabled: boolean;
  rag_enabled: boolean;
};

export type UsageSnapshot = {
  tenant_id: string;
  status: string;
  period: string;
  package: { id: string; slug: string; name: string } | null;
  limits: PackageLimits | null;
  usage: {
    ai_employees: number;
    monthly_call_minutes: number;
    km_documents: number;
    concurrent_calls: number;
  };
  call_limits?: TenantCallLimits | null;
  daily_usage?: {
    call_minutes: number;
    timezone?: string;
  };
};

export function getSettings() {
  return apiFetch<TenantSettings>('/api/tenant/settings');
}

export function putSettings(body: Partial<TenantSettings>) {
  return apiFetch<TenantSettings>('/api/tenant/settings', {
    method: 'PUT',
    body: JSON.stringify(body)
  });
}

export function getUsage() {
  return apiFetch<UsageSnapshot>('/api/tenant/usage');
}

export function getCallLimits() {
  return apiFetch<TenantCallLimits>('/api/tenant/call-limits');
}

export function putCallLimits(body: {
  max_minutes_per_call: number;
  max_call_minutes_per_day: number;
}) {
  return apiFetch<TenantCallLimits>('/api/tenant/call-limits', {
    method: 'PUT',
    body: JSON.stringify(body)
  });
}
