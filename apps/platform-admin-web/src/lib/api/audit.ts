import { apiFetch } from '$lib/api/http';

export type AuditEvent = {
  event_id: string;
  occurred_at: string;
  tenant_id: string;
  actor_id: string;
  actor_type: string;
  action: string;
  resource_type: string;
  resource_id?: string;
  request_id: string;
  source: string;
  outcome: string;
  metadata?: Record<string, unknown>;
};

export type AuditPage = {
  events: AuditEvent[];
  next_cursor?: string;
  range: { start_date: string; end_date: string; timezone: string };
};

export type AuditHealth = {
  mode: string;
  queue_depth: number;
  last_successful_transfer: string | null;
  oldest_pending_file_age_seconds: number;
  pending_files: number;
  failed_files: number;
  clickhouse: 'operational' | 'degraded' | 'unavailable' | 'disabled';
};

export function listAuditEvents(params: Record<string, string>): Promise<AuditPage> {
  const query = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value) query.set(key, value);
  }
  const qs = query.toString();
  return apiFetch(`/api/platform/audit-logs${qs ? `?${qs}` : ''}`);
}

export function getAuditHealth(): Promise<AuditHealth> {
  return apiFetch('/api/platform/audit-logs/health');
}
