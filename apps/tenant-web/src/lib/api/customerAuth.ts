import { apiFetch } from './http';

export type CustomerAuthSettings = {
  tenant_id: string;
  enabled: boolean;
  auth_mode: 'optional' | 'required';
  allowed_domains: string[];
  otp_ttl_seconds: number;
  session_ttl_seconds: number;
  require_auth_for_workforce: boolean;
  customer_daily_call_seconds: number;
  customer_max_call_seconds: number;
  created_at?: string;
  updated_at?: string;
};

export function getCustomerAuthSettings() {
  return apiFetch<CustomerAuthSettings>('/api/tenant/customer-auth/settings');
}

export function putCustomerAuthSettings(body: Partial<CustomerAuthSettings>) {
  return apiFetch<CustomerAuthSettings>('/api/tenant/customer-auth/settings', {
    method: 'PUT',
    body: JSON.stringify(body)
  });
}
