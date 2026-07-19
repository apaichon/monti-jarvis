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
const CHECKOUT_ORDER_ID = 'monti_checkout_order_id';
const CHECKOUT_ORDER_NO = 'monti_checkout_order_no';

/** Prefer localStorage so session survives ChillPay external redirect (sessionStorage is tab-only and flaky across return). */
function storage(): Storage | null {
  if (!browser) return null;
  try {
    return window.localStorage;
  } catch {
    return null;
  }
}

function read(key: string): string | null {
  const s = storage();
  if (!s) return null;
  const v = s.getItem(key);
  if (v) return v;
  // Migrate from older sessionStorage sessions.
  try {
    const legacy = sessionStorage.getItem(key);
    if (legacy) {
      s.setItem(key, legacy);
      sessionStorage.removeItem(key);
      return legacy;
    }
  } catch {
    /* ignore */
  }
  return null;
}

function write(key: string, value: string) {
  const s = storage();
  if (!s) return;
  s.setItem(key, value);
  try {
    sessionStorage.removeItem(key);
  } catch {
    /* ignore */
  }
}

function remove(key: string) {
  const s = storage();
  if (s) s.removeItem(key);
  try {
    sessionStorage.removeItem(key);
  } catch {
    /* ignore */
  }
}

type SessionListener = () => void;
const listeners = new Set<SessionListener>();

/** Subscribe to session save/clear so layout can re-paint nav without full reload (SPRINT-042). */
export function subscribeSession(fn: SessionListener): () => void {
  listeners.add(fn);
  return () => {
    listeners.delete(fn);
  };
}

function notifySession() {
  for (const fn of [...listeners]) {
    try {
      fn();
    } catch {
      /* ignore listener errors */
    }
  }
  if (browser) {
    try {
      window.dispatchEvent(new CustomEvent('monti-tenant-session'));
    } catch {
      /* ignore */
    }
  }
}

export function saveSession(pair: TokenPair, tenantId?: string, registrationId?: string) {
  if (!browser) return;
  write(ACCESS_KEY, pair.access_token);
  write(REFRESH_KEY, pair.refresh_token);
  write(USER_KEY, JSON.stringify(pair.user));
  if (tenantId) write(TENANT_KEY, tenantId);
  if (registrationId) write(REG_KEY, registrationId);
  notifySession();
}

export function getAccessToken(): string | null {
  return read(ACCESS_KEY);
}

/** Write access token only (e.g. bootstrap preview voice on localhost from *.local). */
export function setAccessToken(token: string) {
  if (!browser || !token) return;
  write(ACCESS_KEY, token);
  notifySession();
}

export function applyTokenPair(pair: Pick<TokenPair, 'access_token' | 'refresh_token'> & { user?: UserProfile }) {
  if (!browser) return;
  write(ACCESS_KEY, pair.access_token);
  if (pair.refresh_token) write(REFRESH_KEY, pair.refresh_token);
  if (pair.user) write(USER_KEY, JSON.stringify(pair.user));
  notifySession();
}

export function getRefreshToken(): string | null {
  return read(REFRESH_KEY);
}

export function getStoredTenantId(): string | null {
  return read(TENANT_KEY);
}

export function getStoredUser(): UserProfile | null {
  const raw = read(USER_KEY);
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
  remove(ACCESS_KEY);
  remove(REFRESH_KEY);
  remove(USER_KEY);
  remove(TENANT_KEY);
  remove(REG_KEY);
  notifySession();
}

/**
 * Validate next path for post-login redirect (open-redirect safe).
 * Accepts paths under the app base or /tenant/ prefix.
 */
export function safeTenantNextPath(next: string | null | undefined, appBase: string, fallback?: string): string {
  const fb = fallback || `${appBase}/backoffice`;
  if (!next) return fb;
  const n = next.trim();
  if (!n) return fb;
  // Absolute URLs / protocol-relative
  if (/^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(n) || n.startsWith('//')) return fb;
  if (n.includes('..')) return fb;
  if (appBase && (n === appBase || n.startsWith(`${appBase}/`))) return n;
  if (n.startsWith('/tenant/')) return n;
  // Relative app path without base
  if (n.startsWith('/') && !n.startsWith('//')) {
    // Allow only known app-relative segments when base is empty or /tenant
    if (!appBase || appBase === '') return n;
  }
  return fb;
}

export function saveCheckoutOrder(orderId: string, orderNo?: string) {
  if (orderId) write(CHECKOUT_ORDER_ID, orderId);
  if (orderNo) write(CHECKOUT_ORDER_NO, orderNo);
}

export function getCheckoutOrderId(): string | null {
  return read(CHECKOUT_ORDER_ID);
}

export function getCheckoutOrderNo(): string | null {
  return read(CHECKOUT_ORDER_NO);
}
