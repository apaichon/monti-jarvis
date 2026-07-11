import { apiFetch } from '$lib/api/http';

export type TenantKYCDocument = {
  object_key: string;
  url: string;
};

export type TenantKYCProfile = {
  tenant_id: string;
  contact_name: string;
  contact_phone: string;
  contact_address: string;
  photo_url: string;
  photo_object_key: string;
  documents: TenantKYCDocument[];
  status: string;
  submitted_at?: string;
  reviewed_at?: string;
  reviewed_by?: string;
  rejection_reason?: string;
  updated_at: string;
};

export type TenantKYCReview = {
  tenant: {
    id: string;
    slug: string;
    name: string;
    status: string;
    created_at: string;
  };
  registration: {
    id: string;
    company_name: string;
    admin_email: string;
    status: string;
    created_at: string;
  } | null;
  kyc: TenantKYCProfile;
};

export type KYCDecisionResult = {
  tenant_id: string;
  tenant_status: string;
  registration_status: string;
  kyc_status: string;
  rejection_reason?: string;
  reviewed_at: string;
  reviewed_by: string;
  email_sent?: boolean;
  email_to?: string;
  email_error?: string;
};

export function getTenantKYC(tenantId: string) {
  return apiFetch<TenantKYCReview>(`/api/platform/tenants/${tenantId}/kyc`);
}

export function approveTenantKYC(tenantId: string) {
  return apiFetch<KYCDecisionResult>(`/api/platform/tenants/${tenantId}/kyc/approve`, {
    method: 'POST'
  });
}

export function rejectTenantKYC(tenantId: string, reason: string) {
  return apiFetch<KYCDecisionResult>(`/api/platform/tenants/${tenantId}/kyc/reject`, {
    method: 'POST',
    body: JSON.stringify({ reason })
  });
}