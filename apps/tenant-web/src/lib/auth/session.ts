import { browser } from '$app/environment';

export type UserProfile = {
  id: string;
  email: string;
  display_name: string;
  role: string;
  tenant_id?: string;
};

export type TokenPair = {
  access_token: string;
  refresh_token: string;
  expires_in: number;
  token_type: string;
  user: UserProfile;
};

const ACCESS_KEY = 'monti_tenant_access';
const REFRESH_KEY = 'monti_tenant_refresh';
const USER_KEY = 'monti_tenant_user';
const TENANT_KEY = 'monti_tenant_slug';
const REG_KEY = 'monti_tenant_registration';

export function saveSession(pair: TokenPair, tenantId?: string, registrationId?: string) {
  if (!browser) return;
  sessionStorage.setItem(ACCESS_KEY, pair.access_token);
  sessionStorage.setItem(REFRESH_KEY, pair.refresh_token);
  sessionStorage.setItem(USER_KEY, JSON.stringify(pair.user));
  if (tenantId) sessionStorage.setItem(TENANT_KEY, tenantId);
  if (registrationId) sessionStorage.setItem(REG_KEY, registrationId);
}

export function getAccessToken(): string | null {
  if (!browser) return null;
  return sessionStorage.getItem(ACCESS_KEY);
}

export function getStoredTenantId(): string | null {
  if (!browser) return null;
  return sessionStorage.getItem(TENANT_KEY);
}

export function getStoredUser(): UserProfile | null {
  if (!browser) return null;
  const raw = sessionStorage.getItem(USER_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as UserProfile;
  } catch {
    return null;
  }
}

export function hasRegistrationSession(): boolean {
  return !!getAccessToken();
}

export function clearSession() {
  if (!browser) return;
  sessionStorage.removeItem(ACCESS_KEY);
  sessionStorage.removeItem(REFRESH_KEY);
  sessionStorage.removeItem(USER_KEY);
  sessionStorage.removeItem(TENANT_KEY);
  sessionStorage.removeItem(REG_KEY);
}