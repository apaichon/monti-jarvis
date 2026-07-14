import { browser } from '$app/environment';
import { base } from '$app/paths';
import { clearSession, getAccessToken } from '$lib/auth/session';

export class ApiError extends Error {
  status: number;

  constructor(status: number, message: string) {
    super(message);
    this.status = status;
  }
}

let authRedirecting = false;

/** Clear an expired tenant session and return the user to login once. */
export function handleUnauthorized(hadSession = true) {
  if (!browser || !hadSession || authRedirecting) return;
  authRedirecting = true;
  clearSession();

  const loginPath = `${base}/login`;
  if (window.location.pathname === loginPath || window.location.pathname.endsWith('/login')) {
    authRedirecting = false;
    return;
  }

  const currentPath = `${window.location.pathname}${window.location.search}`;
  const params = new URLSearchParams({ reason: 'session_expired', next: currentPath });
  window.location.replace(`${loginPath}?${params.toString()}`);
}

export async function apiFetch<T>(path: string, init: RequestInit = {}): Promise<T> {
  const headers = new Headers(init.headers);
  // Let the browser set multipart boundary for FormData.
  if (!headers.has('Content-Type') && init.body && !(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json');
  }
  const token = getAccessToken();
  if (token) {
    headers.set('Authorization', `Bearer ${token}`);
  }

  let res: Response;
  try {
    res = await fetch(path, { ...init, headers });
  } catch (err) {
    const msg = err instanceof Error ? err.message : 'network error';
    throw new ApiError(0, `Network error: ${msg}`);
  }

  const ct = res.headers.get('content-type') || '';
  const isJSON = ct.includes('application/json');

  if (!res.ok) {
    if (res.status === 401 && !!token) {
      handleUnauthorized(true);
    }
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
