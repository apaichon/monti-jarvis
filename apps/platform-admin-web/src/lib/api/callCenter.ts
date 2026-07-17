import { apiFetch } from '$lib/api/http';

export type PlatformCallCenterStatistics = {
  range: { start_date: string; end_date: string; timezone: string };
  freshness: {
    source: string;
    status: 'current' | 'stale' | 'empty' | 'unavailable';
    generated_at: string;
    last_projected_at?: string;
  };
  totals: {
    completed_conversations: number;
    total_duration_seconds: number;
    average_duration_seconds: number;
    chat_conversations: number;
    voice_conversations: number;
    range_call_minutes: number;
    reviewed_conversations: number;
    average_satisfaction: number;
    review_completion_rate: number;
    satisfaction_distribution: Record<string, number>;
  };
  by_channel: CallCenterDimension[];
  by_avatar: CallCenterDimension[];
  package_usage: {
    active_package_tenants: number;
    range_call_minutes: number;
    enforcement_counters: string;
  };
  enrichment: { satisfaction: 'available' | 'unavailable'; packages: 'available' | 'unavailable' };
  tenants: PlatformCallCenterTenant[];
  pagination: { total: number; limit: number; offset: number };
};

export type CallCenterDimension = {
  id?: string;
  name?: string;
  channel?: string;
  completed: number;
  total_duration_seconds: number;
  average_duration_seconds: number;
};

export type PlatformCallCenterTenant = {
  tenant_id: string;
  slug: string;
  name: string;
  logo_url?: string;
  package: { name: string; status: string };
  analytics_status: 'current' | 'stale' | 'empty' | 'unavailable';
  completed_conversations: number;
  total_duration_seconds: number;
  average_duration_seconds: number;
  range_call_minutes: number;
  reviewed_conversations: number;
  average_satisfaction: number;
  review_completion_rate: number;
};

export function getPlatformCallCenterStatistics(params: Record<string, string>) {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value) query.set(key, value);
  }
  const qs = query.toString();
  return apiFetch<PlatformCallCenterStatistics>(`/api/platform/call-center/statistics${qs ? `?${qs}` : ''}`);
}
