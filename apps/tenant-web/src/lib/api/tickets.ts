import { apiFetch } from './http';

export type TicketStatus = 'open' | 'in_progress' | 'waiting_customer' | 'resolved' | 'closed';
export type TicketPriority = 'low' | 'normal' | 'high' | 'urgent';
export type TicketCategory = 'general' | 'billing' | 'technical' | 'other';

export type Ticket = {
  id: string;
  subject: string;
  description?: string;
  source_summary?: Record<string, unknown>;
  category: TicketCategory;
  priority: TicketPriority;
  status: TicketStatus;
  customer_id?: string;
  customer_label?: string;
  avatar_id?: string;
  avatar_name?: string;
  source: string;
  assignee_user_id?: string;
  conversation_record_id?: string;
  call_id?: string;
  contact_email_masked?: string;
  last_activity_at: string;
  resolved_at?: string;
  closed_at?: string;
};

export type TicketEvent = {
  id: string;
  ticket_id: string;
  event_type: string;
  actor_type: string;
  actor_id?: string;
  note?: string;
  payload?: Record<string, unknown>;
  created_at: string;
};

export type TicketDetail = { ticket: Ticket; events: TicketEvent[] };

export type TicketFilters = {
  startDate?: string;
  endDate?: string;
  status?: string;
  priority?: string;
  category?: string;
  avatarId?: string;
  assigneeUserId?: string;
};

export async function listTickets(filters: TicketFilters = {}) {
  const params = new URLSearchParams();
  if (filters.startDate) params.set('start_date', filters.startDate);
  if (filters.endDate) params.set('end_date', filters.endDate);
  if (filters.status) params.set('status', filters.status);
  if (filters.priority) params.set('priority', filters.priority);
  if (filters.category) params.set('category', filters.category);
  if (filters.avatarId) params.set('avatar_id', filters.avatarId);
  if (filters.assigneeUserId) params.set('assignee_user_id', filters.assigneeUserId);
  const query = params.toString() ? `?${params.toString()}` : '';
  return apiFetch<{ tickets: Ticket[]; next_cursor: string | null }>(`/api/tenant/tickets${query}`);
}

export function getTicket(id: string) {
  return apiFetch<TicketDetail>(`/api/tenant/tickets/${encodeURIComponent(id)}`);
}

export function updateTicket(
  id: string,
  body: { status?: TicketStatus; priority?: TicketPriority; assignee_user_id?: string }
) {
  return apiFetch<{ ticket: Ticket }>(`/api/tenant/tickets/${encodeURIComponent(id)}`, {
    method: 'PATCH',
    body: JSON.stringify(body)
  });
}

export function addTicketNote(id: string, note: string) {
  return apiFetch<{ event: TicketEvent }>(`/api/tenant/tickets/${encodeURIComponent(id)}/events`, {
    method: 'POST',
    body: JSON.stringify({ type: 'note', note })
  });
}
