import { apiFetch } from '$lib/api/http';

export type TenantListItem = {
  id: string;
  slug: string;
  name: string;
  status: string;
  registration_id: string;
  admin_email: string;
  kyc_status?: string;
  created_at: string;
};

export type TenantsResponse = {
  tenants: TenantListItem[];
  total: number;
  limit: number;
  offset: number;
};

export function listTenants(status = '', kycStatus = '', limit = 50, offset = 0) {
  const params = new URLSearchParams();
  if (status) params.set('status', status);
  if (kycStatus) params.set('kyc_status', kycStatus);
  params.set('limit', String(limit));
  params.set('offset', String(offset));
  const q = params.toString();
  return apiFetch<TenantsResponse>(`/api/platform/tenants?${q}`);
}