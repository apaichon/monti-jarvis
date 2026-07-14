import { customerAuthHeaders } from './customerAuth';

export type CustomerTicketInput = {
  call_id: string;
  confirm_escalation: true;
  subject: string;
  description: string;
  category: 'general' | 'billing' | 'technical' | 'other';
  contact_name?: string;
  contact_email?: string;
};

export type CustomerTicket = {
  id: string;
  status: string;
  priority: string;
  category: string;
  source: string;
  conversation_record_id?: string;
  created_at: string;
};

export async function createCustomerTicket(
  payload: CustomerTicketInput,
  opts?: { tenantId?: string; idempotencyKey?: string }
) {
  const headers = customerAuthHeaders({ 'content-type': 'application/json' });
  if (opts?.tenantId) headers['X-Tenant-Id'] = opts.tenantId;
  headers['Idempotency-Key'] = opts?.idempotencyKey || crypto.randomUUID();
  const res = await fetch('/api/customer/tickets', {
    method: 'POST',
    headers,
    body: JSON.stringify(payload)
  });
  const data = await res.json();
  if (!res.ok) throw new Error(data.error || data.message || 'Could not create follow-up ticket');
  return data as { ticket: CustomerTicket; idempotent?: boolean };
}
