import { test, expect } from '../fixtures';
import { registerViaApi, verifyViaApi } from '../helpers/api';
import * as db from '../helpers/db';

// SPRINT-006 · TASK-0027 (auth integration: JWT session, pending-tenant login rules)
//
// Registration issues an email-verification token; delivery (email) is stubbed
// by inserting the token row directly — see helpers/db.ts.

test.describe('Email verification + tenant login (TASK-0027)', () => {
  test.beforeEach(({ dbReady }) => {
    test.skip(!dbReady, 'requires Postgres (set POSTGRES_URL / run make infra-init)');
  });

  test('an unverified admin cannot sign in — 403 "email not verified" (AC #3)', async ({
    page,
    request,
    tenant
  }) => {
    expect((await registerViaApi(request, tenant)).status()).toBe(201);

    await page.goto('/tenant/login');
    await page.locator('#email').fill(tenant.email);
    await page.locator('#password').fill(tenant.password);
    await page.getByRole('button', { name: 'Sign in', exact: true }).click();

    await expect(page.locator('p.error')).toContainText('email not verified');
  });

  test('the verification link issues a session and lands on the success page (AC #1, #2)', async ({
    page,
    request,
    tenant
  }) => {
    expect((await registerViaApi(request, tenant)).status()).toBe(201);

    const userId = await db.getUserIdByEmail(tenant.email);
    expect(userId, 'registered user exists in DB').toBeTruthy();
    const rawToken = 'e2e-verify-' + Math.random().toString(36).slice(2);
    await db.insertVerificationToken(userId!, rawToken);

    await page.goto(`/tenant/register/verify?token=${encodeURIComponent(rawToken)}`);

    // The verify page swaps the token for a session, then redirects to success.
    await expect(page.getByRole('heading', { name: 'Registration complete' })).toBeVisible({
      timeout: 15000
    });
    const access = await page.evaluate(() => sessionStorage.getItem('monti_tenant_access'));
    expect(access, 'tenant access token stored after verification').toBeTruthy();
  });

  test('a verified admin can sign in and reach the backoffice (AC #1, #2)', async ({
    page,
    request,
    tenant
  }) => {
    expect((await registerViaApi(request, tenant)).status()).toBe(201);

    const userId = await db.getUserIdByEmail(tenant.email);
    const rawToken = 'e2e-login-' + Math.random().toString(36).slice(2);
    await db.insertVerificationToken(userId!, rawToken);
    // Complete verification via the API so this test exercises login on its own.
    expect((await verifyViaApi(request, rawToken)).status()).toBe(200);

    await page.goto('/tenant/login');
    await page.locator('#email').fill(tenant.email);
    await page.locator('#password').fill(tenant.password);
    await page.getByRole('button', { name: 'Sign in', exact: true }).click();

    await expect(page).toHaveURL(/\/tenant\/backoffice/);
    await expect(page.getByRole('heading', { name: 'Tenant backoffice' })).toBeVisible();
  });
});
