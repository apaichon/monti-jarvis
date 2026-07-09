import { test, expect } from '../fixtures';
import { registerViaApi, verifyViaApi, platformLogin } from '../helpers/api';
import * as db from '../helpers/db';
import { PLATFORM_EMAIL, PLATFORM_PASSWORD } from '../helpers/config';

// SPRINT-006 · TASK-0028 (platform admin tenant list + status filter + guards)

test.describe('Platform tenant list (TASK-0028)', () => {
  test.beforeEach(({ dbReady }) => {
    test.skip(!dbReady, 'requires Postgres (set POSTGRES_URL / run make infra-init)');
  });

  test('rejects an unauthenticated request with 401 (AC #3)', async ({ request }) => {
    const res = await request.get('/api/platform/tenants');
    expect(res.status()).toBe(401);
  });

  test('forbids a tenant_admin with 403 (AC #3)', async ({ request, tenant }) => {
    expect((await registerViaApi(request, tenant)).status()).toBe(201);
    const userId = await db.getUserIdByEmail(tenant.email);
    const rawToken = 'e2e-role-' + Math.random().toString(36).slice(2);
    await db.insertVerificationToken(userId!, rawToken);
    const verified = await verifyViaApi(request, rawToken);
    expect(verified.status()).toBe(200);
    const tenantToken = (await verified.json()).access_token as string;

    const res = await request.get('/api/platform/tenants', {
      headers: { Authorization: `Bearer ${tenantToken}` }
    });
    expect(res.status()).toBe(403);
  });

  test('lists the new pending tenant for a platform admin (AC #1, #2)', async ({ request, tenant }) => {
    expect((await registerViaApi(request, tenant)).status()).toBe(201);

    const token = await platformLogin(request);
    const res = await request.get('/api/platform/tenants?status=pending_kyc&limit=100', {
      headers: { Authorization: `Bearer ${token}` }
    });
    expect(res.status()).toBe(200);

    const body = await res.json();
    const found = (body.tenants ?? []).find((x: { slug: string }) => x.slug === tenant.slug);
    expect(found, `tenant "${tenant.slug}" present in pending_kyc list`).toBeTruthy();
    expect(found.status).toBe('pending_kyc');
    expect(found.admin_email).toBe(tenant.email);
  });

  test('shows the new pending tenant in the admin UI after login (AC #4)', async ({ page, request, tenant }) => {
    expect((await registerViaApi(request, tenant)).status()).toBe(201);

    // Fresh browser → the tenants route redirects to the admin login.
    await page.goto('/admin/tenants');
    await expect(page).toHaveURL(/\/admin\/login/);

    await page.locator('#email').fill(PLATFORM_EMAIL);
    await page.locator('#password').fill(PLATFORM_PASSWORD);
    await page.getByRole('button', { name: 'Sign in', exact: true }).click();

    await expect(page).toHaveURL(/\/admin\/tenants/);
    await expect(page.getByRole('heading', { name: 'Tenants' })).toBeVisible();
    // Default filter is pending_kyc and the list is newest-first.
    await expect(page.locator('table.table tbody')).toContainText(tenant.slug, { timeout: 10000 });
  });
});
