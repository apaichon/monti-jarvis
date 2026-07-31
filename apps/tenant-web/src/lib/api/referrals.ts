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
  return apiFetch<{ tenant_id: string; referrals: Referral[]; bonus: BonusBalance[]; redemptions: ReferralRedemption[] }>(
    '/api/tenant/referrals'
  );
}

export type ReferralRedemption = {
  id?: string;
  redemption_id?: string;
  referral_code?: string;
  status: string;
  applied_at?: string;
  reversed_at?: string;
  bonus?: Array<{ dimension: string; amount: number; expires_at?: string }>;
};

export function validateReferralCode(code: string) {
  return apiFetch<{ eligible: boolean; preview_bonus: Array<{ dimension: string; remaining: number; granted: number; unit?: string }> }>(
    '/api/tenant/referrals/validate',
    { method: 'POST', body: JSON.stringify({ code }) }
  );
}

export function redeemReferralCode(code: string, idempotency_key?: string) {
  return apiFetch<ReferralRedemption>('/api/tenant/referrals/redeem', {
    method: 'POST',
    body: JSON.stringify({ code, idempotency_key })
  });
}
