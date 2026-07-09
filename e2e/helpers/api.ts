// Thin wrappers over the public/auth API, used to arrange state quickly
// (e.g. pre-create a colliding tenant) without clicking through the UI.

import type { APIRequestContext, APIResponse } from '@playwright/test';
import type { TenantData } from './data';
import { PLATFORM_EMAIL, PLATFORM_PASSWORD } from './config';

export function registerViaApi(request: APIRequestContext, t: TenantData): Promise<APIResponse> {
  return request.post('/api/public/tenant/register', {
    data: {
      company_name: t.companyName,
      slug: t.slug,
      admin_email: t.email,
      admin_password: t.password,
      admin_display_name: t.displayName
    }
  });
}

export function verifyViaApi(request: APIRequestContext, token: string): Promise<APIResponse> {
  return request.post('/api/public/tenant/verify-email', { data: { token } });
}

export async function platformLogin(request: APIRequestContext): Promise<string> {
  const res = await request.post('/api/auth/login', {
    data: { email: PLATFORM_EMAIL, password: PLATFORM_PASSWORD }
  });
  if (!res.ok()) {
    throw new Error(`platform login failed: ${res.status()} ${await res.text()}`);
  }
  return (await res.json()).access_token as string;
}
