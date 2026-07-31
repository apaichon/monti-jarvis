import { apiFetch } from '$lib/api/http';

export type ReferralRedemption = {
  id: string;
  redeemer_tenant_id: string;
  referrer_tenant_id: string;
  referral_code: string;
  status: 'applied' | 'reversed' | string;
  applied_at: string;
  reversed_at?: string;
};

export function listReferralRedemptions(filters?: {
  tenantId?: string;
  code?: string;
  status?: string;
}): Promise<{ redemptions: ReferralRedemption[] }> {
  const query = new URLSearchParams();
  if (filters?.tenantId) query.set('tenant_id', filters.tenantId);
  if (filters?.code) query.set('code', filters.code);
  if (filters?.status) query.set('status', filters.status);
  const suffix = query.size ? `?${query.toString()}` : '';
  return apiFetch(`/api/platform/referrals/redemptions${suffix}`);
}

export function reverseReferralRedemption(id: string, reason: string): Promise<ReferralRedemption> {
  return apiFetch(`/api/platform/referrals/redemptions/${encodeURIComponent(id)}/reverse`, {
    method: 'POST',
    body: JSON.stringify({ reason })
  });
}
