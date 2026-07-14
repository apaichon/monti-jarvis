import { apiFetch } from './http';

export type CallCenterBucket = {
  id?: string;
  name?: string;
  channel?: string;
  completed: number;
  total_duration_seconds: number;
  average_duration_seconds: number;
};

export type CallCenterStatistics = {
  range: { start_date: string; end_date: string };
  timezone: string;
  freshness?: string;
  total_completed_conversations: number;
  total_duration_seconds: number;
  average_duration_seconds: number;
  by_avatar: CallCenterBucket[];
  by_channel: CallCenterBucket[];
  quota?: {
    status?: string;
    period?: string;
    package?: { name?: string } | null;
    limits?: { max_monthly_call_minutes?: number } | null;
    usage?: { monthly_call_minutes?: number };
  } | null;
  daily_usage: { call_minutes: number; timezone: string };
  call_limits?: { max_minutes_per_call?: number; max_call_minutes_per_day?: number } | null;
};

export async function getCallCenterStatistics(filters: { startDate?: string; endDate?: string } = {}) {
  const params = new URLSearchParams();
  if (filters.startDate) params.set('start_date', filters.startDate);
  if (filters.endDate) params.set('end_date', filters.endDate);
  const query = params.toString() ? `?${params.toString()}` : '';
  return apiFetch<CallCenterStatistics>(`/api/tenant/call-center/statistics${query}`);
}
