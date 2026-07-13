export type CustomerProfile = {
  id: string;
  tenant_id: string;
  display_name: string;
  email: string;
  tier_id?: string;
  group_ids: string[];
  locale?: string;
  role: 'customer';
};

export type OTPRequestResponse = {
  challenge_id: string;
  status: 'otp_sent';
  delivery: { channel: 'email'; to: string };
  expires_in: number;
  resend_after: number;
  customer_hint: {
    matched_existing_customer: boolean;
    requires_profile_completion: boolean;
    email_domain_policy: string;
  };
};

export type CustomerAuthResponse = {
  status: 'authenticated';
  access_token: string;
  refresh_token: string;
  token_type: 'Bearer';
  expires_in: number;
  refresh_expires_in: number;
  customer: CustomerProfile;
};

const ACCESS_KEY = 'monti_customer_access_token';
const REFRESH_KEY = 'monti_customer_refresh_token';
const PROFILE_KEY = 'monti_customer_profile';

export function getCustomerAccessToken() {
  if (typeof localStorage === 'undefined') return '';
  return localStorage.getItem(ACCESS_KEY) || '';
}

export function getCustomerRefreshToken() {
  if (typeof localStorage === 'undefined') return '';
  return localStorage.getItem(REFRESH_KEY) || '';
}

export function getStoredCustomer(): CustomerProfile | null {
  if (typeof localStorage === 'undefined') return null;
  const raw = localStorage.getItem(PROFILE_KEY);
  if (!raw) return null;
  try {
    return JSON.parse(raw) as CustomerProfile;
  } catch {
    return null;
  }
}

export function storeCustomerSession(data: CustomerAuthResponse) {
  localStorage.setItem(ACCESS_KEY, data.access_token);
  localStorage.setItem(REFRESH_KEY, data.refresh_token);
  localStorage.setItem(PROFILE_KEY, JSON.stringify(data.customer));
}

export function clearCustomerSession() {
  if (typeof localStorage === 'undefined') return;
  localStorage.removeItem(ACCESS_KEY);
  localStorage.removeItem(REFRESH_KEY);
  localStorage.removeItem(PROFILE_KEY);
}

async function parseJSON<T>(res: Response): Promise<T> {
  const data = await res.json().catch(() => ({}));
  if (!res.ok) {
    throw new Error(data.error || data.message || `HTTP ${res.status}`);
  }
  return data as T;
}

export async function requestCustomerOTP(body: {
  tenant_id?: string;
  email: string;
  display_name?: string;
  locale?: string;
}, opts?: { tenantId?: string }) {
  const headers: Record<string, string> = { 'content-type': 'application/json' };
  if (opts?.tenantId) headers['X-Tenant-Id'] = opts.tenantId;
  const res = await fetch('/api/customer/auth/request-otp', {
    method: 'POST',
    headers,
    body: JSON.stringify(body)
  });
  return parseJSON<OTPRequestResponse>(res);
}

export async function verifyCustomerOTP(body: {
  tenant_id?: string;
  challenge_id: string;
  otp: string;
}, opts?: { tenantId?: string }) {
  const headers: Record<string, string> = { 'content-type': 'application/json' };
  if (opts?.tenantId) headers['X-Tenant-Id'] = opts.tenantId;
  const res = await fetch('/api/customer/auth/verify-otp', {
    method: 'POST',
    headers,
    body: JSON.stringify(body)
  });
  const data = await parseJSON<CustomerAuthResponse>(res);
  storeCustomerSession(data);
  return data;
}

export async function logoutCustomer() {
  const refresh_token = getCustomerRefreshToken();
  await fetch('/api/customer/auth/logout', {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ refresh_token })
  }).catch(() => {});
  clearCustomerSession();
}

export async function loadCustomerMe() {
  const token = getCustomerAccessToken();
  if (!token) return null;
  const res = await fetch('/api/customer/me', {
    headers: { Authorization: `Bearer ${token}` }
  });
  if (res.status === 401 || res.status === 403) {
    clearCustomerSession();
    return null;
  }
  const data = await parseJSON<{ customer: CustomerProfile }>(res);
  localStorage.setItem(PROFILE_KEY, JSON.stringify(data.customer));
  return data.customer;
}

export function customerAuthHeaders(headers: Record<string, string> = {}) {
  const token = getCustomerAccessToken();
  if (token) headers.Authorization = `Bearer ${token}`;
  return headers;
}
