import { browser } from '$app/environment';
import { base } from '$app/paths';

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

const ACCESS_KEY = 'monti_admin_access';
const REFRESH_KEY = 'monti_admin_refresh';
const USER_KEY = 'monti_admin_user';

export function getAccessToken(): string | null {
  if (!browser) return null;
  return sessionStorage.getItem(ACCESS_KEY);
}

export function getRefreshToken(): string | null {
  if (!browser) return null;
  return sessionStorage.getItem(REFRESH_KEY);
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

export function saveSession(pair: TokenPair) {
  if (!browser) return;
  sessionStorage.setItem(ACCESS_KEY, pair.access_token);
  sessionStorage.setItem(REFRESH_KEY, pair.refresh_token);
  sessionStorage.setItem(USER_KEY, JSON.stringify(pair.user));
}

export function clearSession() {
  if (!browser) return;
  sessionStorage.removeItem(ACCESS_KEY);
  sessionStorage.removeItem(REFRESH_KEY);
  sessionStorage.removeItem(USER_KEY);
}

export function isPlatformAdmin(): boolean {
  const user = getStoredUser();
  return user?.role === 'platform_admin';
}

export function loginPath(next?: string) {
  const q = next ? `?next=${encodeURIComponent(next)}` : '';
  return `${base}/login${q}`;
}