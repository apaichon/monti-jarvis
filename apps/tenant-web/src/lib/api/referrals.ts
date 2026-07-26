import { apiFetch } from './http';

export type ReferralCode = {
  id: string;
  tenant_id: string;
  code: string;
  status: string;
};

export type Referral = {
  id: string;
  referred_tenant_id: string;
  code: string;
  status: string;
  source: string;
  qualification_reason?: string;
  attributed_at?: string;
  qualified_at?: string;
};

export type BonusBalance = {
  dimension: string;
  unit: string;
  granted: number;
  used: number;
  expired: number;
  reversed: number;
  remaining: number;
  expires_at?: string;
};

export function getReferralCode() {
  return apiFetch<ReferralCode>('/api/tenant/referral');
}

export function getReferrals() {
  return apiFetch<{ tenant_id: string; referrals: Referral[]; bonus: BonusBalance[] }>('/api/tenant/referrals');
}
