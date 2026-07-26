/** Allowlisted attribution / conversion query keys (DES-0043 §8). */
export const ALLOWLISTED_KEYS = [
  'utm_source',
  'utm_medium',
  'utm_campaign',
  'utm_content',
  'utm_term',
  'ref',
  'package_id',
  'lang',
  'lead_id'
] as const;

export type AttributionKey = (typeof ALLOWLISTED_KEYS)[number];
export type AttributionMap = Partial<Record<AttributionKey, string>>;

const STORAGE_KEY = 'monti_product_attribution';
const SESSION_KEY = 'monti_product_session_key';

const SAFE_EXACT = new Set(['/', '/tenant/register', '/tenant/login', '/tenant/billing']);

function isBrowser(): boolean {
  return typeof window !== 'undefined' && typeof sessionStorage !== 'undefined';
}

function isSafeRelativePath(path: string): boolean {
  if (!path || typeof path !== 'string') return false;
  const trimmed = path.trim();
  if (!trimmed.startsWith('/')) return false;
  if (trimmed.startsWith('//')) return false;
  if (/^[a-zA-Z][a-zA-Z0-9+.-]*:/.test(trimmed)) return false;
  if (trimmed.includes('\\')) return false;
  if (SAFE_EXACT.has(trimmed)) return true;
  if (trimmed === '/product' || trimmed.startsWith('/product/')) return true;
  // Allow product-web internal paths without host prefix (handled by callers with base)
  return false;
}

/** Capture allowlisted keys from the current URL into sessionStorage (merge, last wins). */
export function captureAttributionFromSearch(search: string | URLSearchParams): AttributionMap {
  const params = typeof search === 'string' ? new URLSearchParams(search) : search;
  const next: AttributionMap = { ...getAttribution() };
  for (const key of ALLOWLISTED_KEYS) {
    const value = params.get(key);
    if (value != null && value.trim() !== '') {
      next[key] = value.trim().slice(0, 256);
    }
  }
  if (isBrowser()) {
    try {
      sessionStorage.setItem(STORAGE_KEY, JSON.stringify(next));
    } catch {
      // ignore quota / private mode
    }
  }
  return next;
}

export function getAttribution(): AttributionMap {
  if (!isBrowser()) return {};
  try {
    const raw = sessionStorage.getItem(STORAGE_KEY);
    if (!raw) return {};
    const parsed = JSON.parse(raw) as Record<string, unknown>;
    const out: AttributionMap = {};
    for (const key of ALLOWLISTED_KEYS) {
      const v = parsed[key];
      if (typeof v === 'string' && v.trim()) out[key] = v.trim().slice(0, 256);
    }
    return out;
  } catch {
    return {};
  }
}

/** Opaque client session key for funnel beacons (not PII). */
export function getOrCreateSessionKey(): string {
  if (!isBrowser()) return 'ssr';
  try {
    let key = sessionStorage.getItem(SESSION_KEY);
    if (!key) {
      key =
        typeof crypto !== 'undefined' && 'randomUUID' in crypto
          ? crypto.randomUUID()
          : `s_${Date.now().toString(36)}_${Math.random().toString(36).slice(2, 10)}`;
      sessionStorage.setItem(SESSION_KEY, key);
    }
    return key;
  } catch {
    return `s_${Date.now().toString(36)}`;
  }
}

/**
 * Build a safe href for Monti-relative paths only.
 * Rejects external hosts, protocol-relative URLs, and javascript:/data: schemes.
 * Merges stored attribution with optional extra params (extra wins).
 */
export function buildSafeHref(path: string, extra: Record<string, string> = {}): string {
  if (!isSafeRelativePath(path)) {
    return '/product';
  }
  const attrs = { ...getAttribution(), ...extra };
  const qs = new URLSearchParams();
  for (const key of ALLOWLISTED_KEYS) {
    const value = attrs[key];
    if (value != null && String(value).trim() !== '') {
      qs.set(key, String(value).trim().slice(0, 256));
    }
  }
  // package_id may also come from extra under that name
  if (extra.package_id && !qs.has('package_id')) {
    qs.set('package_id', extra.package_id.slice(0, 256));
  }
  const query = qs.toString();
  return query ? `${path}?${query}` : path;
}

/** Convenience: demo (customer portal root). */
export function demoHref(extra: Record<string, string> = {}): string {
  return buildSafeHref('/', extra);
}

/** Convenience: tenant register. */
export function registerHref(extra: Record<string, string> = {}): string {
  return buildSafeHref('/tenant/register', extra);
}
