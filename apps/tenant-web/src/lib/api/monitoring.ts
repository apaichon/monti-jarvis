import { apiFetch } from './http';

export type MonitoringComponentStatus = 'operational' | 'degraded' | 'unavailable' | 'disabled';
export type AnalyticsStatus = 'current' | 'stale' | 'unavailable' | 'disabled';

export type MonitoringComponent = {
  name: string;
  status: MonitoringComponentStatus;
  latency_ms: number | null;
  checked_at: string;
};

export type AnalyticsHealth = {
  status: AnalyticsStatus;
  generated_at?: string;
  last_projected_at?: string;
};

export type SystemPerformanceSnapshot = {
  overall_status: 'operational' | 'degraded' | 'unavailable';
  checked_at: string;
  components: MonitoringComponent[];
  analytics: AnalyticsHealth;
};

export function getSystemPerformance() {
  return apiFetch<SystemPerformanceSnapshot>('/api/tenant/system-performance');
}
