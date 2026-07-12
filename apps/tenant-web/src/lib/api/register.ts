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

/** Shared Google/GitHub OAuth for login + register (one API surface). */
export function fetchOAuthProviders() {
  return apiFetch<{ providers: string[] }>('/api/public/tenant/oauth/providers');
}

/**
 * Start OAuth. Same URL for login and register.
 * - Login: oauthStartURL('google')
 * - Register: oauthStartURL('google', { company_name, slug })
 * Callback auto-login if account exists, else register / complete workspace.
 */
export function oauthStartURL(
  provider: string,
  params: { company_name?: string; slug?: string; display_name?: string } = {}
) {
  const q = new URLSearchParams();
  if (params.company_name) q.set('company_name', params.company_name);
  if (params.slug) q.set('slug', params.slug);
  if (params.display_name) q.set('display_name', params.display_name);
  const suffix = q.toString() ? `?${q.toString()}` : '';
  return `/api/public/tenant/oauth/${encodeURIComponent(provider)}${suffix}`;
}

export type OAuthCompleteInput = {
  session_id: string;
  company_name: string;
  slug: string;
};

export function completeOAuthRegistration(input: OAuthCompleteInput) {
  return apiFetch<RegisterResponse & TokenPair>('/api/public/tenant/oauth/complete', {
    method: 'POST',
    body: JSON.stringify(input)
  });
}