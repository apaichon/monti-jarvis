import { apiFetch } from '$lib/api/http';

export type LeadKind = 'contact' | 'book_demo' | 'newsletter' | string;
export type LeadStatus =
  | 'new'
  | 'contacted'
  | 'demo_scheduled'
  | 'demo_completed'
  | 'qualified'
  | 'registered'
  | 'kyc_pending'
  | 'package_selected'
  | 'paid'
  | 'converted'
  | 'lost'
  | 'unsubscribed'
  | string;

export const LEAD_STATUSES: LeadStatus[] = [
  'new',
  'contacted',
  'demo_scheduled',
  'demo_completed',
  'qualified',
  'registered',
  'kyc_pending',
  'package_selected',
  'paid',
  'converted',
  'lost',
  'unsubscribed'
];

export type LeadListItem = {
  id: string;
  kind: LeadKind;
  status: LeadStatus;
  email: string;
  full_name?: string;
  company_name?: string;
  phone?: string;
  use_case?: string;
  preferred_channel?: string;
  language?: string;
  utm_source?: string;
  utm_medium?: string;
  utm_campaign?: string;
  utm_content?: string;
  utm_term?: string;
  referral_code?: string;
  landing_path?: string;
  package_interest_id?: string;
  assigned_to?: string;
  converted_tenant_id?: string;
  created_at: string;
  updated_at?: string;
};

export type LeadNote = {
  id: string;
  lead_id: string;
  body: string;
  created_at: string;
  created_by?: string;
};

export type LeadEvent = {
  id: string;
  lead_id: string;
  from_status?: string;
  to_status: string;
  actor?: string;
  created_at: string;
};

export type LeadDetail = LeadListItem & {
  consent_marketing?: boolean;
  consent_contact?: boolean;
  consent_at?: string;
  notes?: LeadNote[];
  history?: LeadEvent[];
  events?: LeadEvent[];
};

export type LeadsListResponse = {
  leads: LeadListItem[];
  total?: number;
  limit?: number;
  offset?: number;
};

export type ListLeadsParams = {
  status?: string;
  kind?: string;
  q?: string;
  limit?: number;
  offset?: number;
};

export function listLeads(params: ListLeadsParams = {}) {
  const qs = new URLSearchParams();
  if (params.status) qs.set('status', params.status);
  if (params.kind) qs.set('kind', params.kind);
  if (params.q) qs.set('q', params.q);
  if (params.limit != null) qs.set('limit', String(params.limit));
  if (params.offset != null) qs.set('offset', String(params.offset));
  const q = qs.toString();
  return apiFetch<LeadsListResponse>(`/api/platform/leads${q ? `?${q}` : ''}`);
}

export function getLead(id: string) {
  return apiFetch<LeadDetail>(`/api/platform/leads/${encodeURIComponent(id)}`);
}

export function patchLead(id: string, body: { status?: string; assigned_to?: string }) {
  return apiFetch<LeadDetail>(`/api/platform/leads/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    body: JSON.stringify(body)
  });
}

export function addLeadNote(id: string, body: string) {
  return apiFetch<LeadNote>(`/api/platform/leads/${encodeURIComponent(id)}/notes`, {
    method: 'POST',
    body: JSON.stringify({ body })
  });
}

export function sourceLabel(lead: Pick<LeadListItem, 'utm_source' | 'utm_campaign' | 'referral_code'>) {
  const parts = [lead.utm_source, lead.utm_campaign].filter(Boolean);
  if (lead.referral_code) parts.push(`ref:${lead.referral_code}`);
  return parts.length ? parts.join('/') : '—';
}
