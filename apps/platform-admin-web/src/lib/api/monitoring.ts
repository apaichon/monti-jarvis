import { apiFetch } from '$lib/api/http';

export type MonitoringStatus = 'operational' | 'degraded' | 'unavailable' | 'disabled';
export type AnalyticsStatus = 'current' | 'stale' | 'unavailable' | 'disabled';

export type MonitoringComponent = {
  name: string;
  status: MonitoringStatus;
  latency_ms: number | null;
  checked_at: string;
};

export type MonitoringAnalytics = {
  status: AnalyticsStatus;
  generated_at?: string;
  last_projected_at?: string;
};

export type MonitoringTenant = {
  tenant_id: string;
  slug: string;
  name: string;
  status: Exclude<MonitoringStatus, 'disabled'>;
  analytics: MonitoringAnalytics;
  audit_status: MonitoringStatus;
};

export type PlatformPerformance = {
  overall_status: Exclude<MonitoringStatus, 'stale'>;
  checked_at: string;
  summary: {
    tenants_total: number;
    operational: number;
    degraded: number;
    unavailable: number;
  };
  components: MonitoringComponent[];
  audit: {
    mode: string;
    status: MonitoringStatus;
    pending_files: number;
    failed_files: number;
    oldest_pending_file_age_seconds: number;
    last_successful_transfer: string | null;
  };
  tenants: MonitoringTenant[];
  total: number;
  limit: number;
  offset: number;
};

export function getPlatformSystemPerformance(params: Record<string, string>): Promise<PlatformPerformance> {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value) query.set(key, value);
  }
  const qs = query.toString();
  return apiFetch(`/api/platform/system-performance${qs ? `?${qs}` : ''}`);
}
