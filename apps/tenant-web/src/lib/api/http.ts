import { browser } from '$app/environment';
import { base } from '$app/paths';
import {
  applyTokenPair,
  clearSession,
  getAccessToken,
  getRefreshToken,
  getStoredUser,
  type TokenPair
} from '$lib/auth/session';

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

let authRedirecting = false;
let refreshInFlight: Promise<boolean> | null = null;

/** Clear an expired tenant session and return the user to login once. */
export function handleUnauthorized(hadSession = true) {
  if (!browser || !hadSession || authRedirecting) return;
  authRedirecting = true;
  clearSession();

  const loginPath = `${base}/login`;
  const path = window.location.pathname;
  if (path === loginPath || path.endsWith('/login')) {
    // Allow future redirects after a successful re-login.
    setTimeout(() => {
      authRedirecting = false;
    }, 500);
    return;
  }

  const currentPath = `${window.location.pathname}${window.location.search}`;
  const params = new URLSearchParams({ reason: 'session_expired', next: currentPath });
  window.location.replace(`${loginPath}?${params.toString()}`);
  // Hard navigation — flag resets on new page load.
}

async function tryRefreshAccessToken(): Promise<boolean> {
  if (!browser) return false;
  const refresh = getRefreshToken();
  if (!refresh) return false;
  if (refreshInFlight) return refreshInFlight;

  refreshInFlight = (async () => {
    try {
      const res = await fetch('/api/auth/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refresh })
      });
      if (!res.ok) return false;
      const pair = (await res.json()) as TokenPair;
      if (!pair?.access_token) return false;
      const user = pair.user || getStoredUser() || undefined;
      applyTokenPair({
        access_token: pair.access_token,
        refresh_token: pair.refresh_token || refresh,
        user: user as TokenPair['user']
      });
      return true;
    } catch {
      return false;
    } finally {
      refreshInFlight = null;
    }
  })();

  return refreshInFlight;
}

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  // Let the browser set multipart boundary for FormData.
  if (!headers.has('Content-Type') && init.body && !(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }

  const doFetch = async () => {
    const token = getAccessToken();
    if (token) {
      headers.set('Authorization', `Bearer ${token}`);
    } else {
      headers.delete('Authorization');
    }
    return fetch(path, { ...init, headers });
  };

  let res: Response;
  try {
    res = await doFetch();
  } catch (err) {
    const msg = err instanceof Error ? err.message : 'network error';
    throw new ApiError(0, `Network error: ${msg}`);
  }

  // One refresh + retry on 401 when we had a session.
  if (res.status === 401 && getAccessToken()) {
    const refreshed = await tryRefreshAccessToken();
    if (refreshed) {
      try {
        res = await doFetch();
      } catch (err) {
        const msg = err instanceof Error ? err.message : 'network error';
        throw new ApiError(0, `Network error: ${msg}`);
      }
    }
    if (res.status === 401) {
      handleUnauthorized(true);
    }
  }

  const ct = res.headers.get('content-type') || '';
  const isJSON = ct.includes('application/json');

  if (!res.ok) {
    let message = res.statusText || `HTTP ${res.status}`;
    if (isJSON) {
      try {
        const body = await res.json();
        if (body?.error) message = body.error;
      } catch {
        // ignore
      }
    } else {
      // SPA fallback or proxy returned HTML/text for a missing API route.
      message = `API ${res.status} (not JSON) — is the server up to date? Restart with make restart.`;
    }
    throw new ApiError(res.status, message);
  }

  if (!isJSON) {
    // e.g. old binary: /api/* falls through to customer SPA index.html with 200.
    throw new ApiError(
      res.status,
      'API returned HTML instead of JSON — server needs rebuild/restart (make restart).'
    );
  }

  return (await res.json()) as T;
}
