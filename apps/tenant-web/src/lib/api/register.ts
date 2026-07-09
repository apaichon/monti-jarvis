import { apiFetch } from '$lib/api/http';
import type { TokenPair } from '$lib/auth/session';

export type RegisterInput = {
  company_name: string;
  slug: string;
  admin_email: string;
  admin_password: string;
  admin_display_name: string;
};

export type RegisterResponse = Partial<TokenPair> & {
  tenant_id?: string;
  slug?: string;
  registration_id?: string;
  verification_required?: boolean;
  message?: string;
};

export function registerTenant(input: RegisterInput) {
  return apiFetch<RegisterResponse>('/api/public/tenant/register', {
    method: 'POST',
    body: JSON.stringify(input)
  });
}

export function verifyEmail(token: string) {
  return apiFetch<RegisterResponse>('/api/public/tenant/verify-email', {
    method: 'POST',
    body: JSON.stringify({ token })
  });
}

export function fetchOAuthProviders() {
  return apiFetch<{ providers: string[] }>('/api/public/tenant/register/oauth/providers');
}

export function oauthStartURL(
  provider: string,
  params: { company_name?: string; slug?: string; display_name?: string }
) {
  const q = new URLSearchParams();
  if (params.company_name) q.set('company_name', params.company_name);
  if (params.slug) q.set('slug', params.slug);
  if (params.display_name) q.set('display_name', params.display_name);
  const suffix = q.toString() ? `?${q.toString()}` : '';
  return `/api/public/tenant/register/oauth/${provider}${suffix}`;
}

export type OAuthCompleteInput = {
  session_id: string;
  company_name: string;
  slug: string;
};

export function completeOAuthRegistration(input: OAuthCompleteInput) {
  return apiFetch<RegisterResponse & TokenPair>('/api/public/tenant/register/oauth/complete', {
    method: 'POST',
    body: JSON.stringify(input)
  });
}