import { test, expect } from '../fixtures';
import { registerViaApi, verifyViaApi, platformLogin } from '../helpers/api';
import * as db from '../helpers/db';

// SPRINT-007 · platform KYC approve flow

async function tenantToken(request: import('@playwright/test').APIRequestContext, tenant: { email: string; password: string }) {
  const userId = await db.getUserIdByEmail(tenant.email);
  const rawToken = 'e2e-kyc-' + Math.random().toString(36).slice(2);
  await db.insertVerificationToken(userId!, rawToken);
  const verified = await verifyViaApi(request, rawToken);
  expect(verified.status()).toBe(200);
  return (await verified.json()).access_token as string;
}

test.describe('Platform KYC review (SPRINT-007)', () => {
  test.beforeEach(({ dbReady }) => {
    test.skip(!dbReady, 'requires Postgres (set POSTGRES_URL / run make infra-init)');
  });

  test('approves submitted KYC and activates tenant (AC #3, #5)', async ({ request, tenant }) => {
    expect((await registerViaApi(request, tenant)).status()).toBe(201);
    const token = await tenantToken(request, tenant);

    const submit = await request.post('/api/tenant/kyc/submit', {
      headers: { Authorization: `Bearer ${token}` }
    });
    expect(submit.status()).toBe(200);

    const platform = await platformLogin(request);
    const pkg = await request.get(`/api/platform/tenants/${tenant.slug}/kyc`, {
      headers: { Authorization: `Bearer ${platform}` }
    });
    expect(pkg.status()).toBe(200);
    const body = await pkg.json();
    expect(body.kyc.status).toBe('submitted');

    const approve = await request.post(`/api/platform/tenants/${tenant.slug}/kyc/approve`, {
      headers: { Authorization: `Bearer ${platform}` }
    });
    expect(approve.status()).toBe(200);
    const decision = await approve.json();
    expect(decision.tenant_status).toBe('active');
    expect(decision.kyc_status).toBe('approved');

    const me = await request.get('/api/auth/me', {
      headers: { Authorization: `Bearer ${token}` }
    });
    expect(me.status()).toBe(200);
  });

  test('rejects KYC without reason with 400', async ({ request, tenant }) => {
    expect((await registerViaApi(request, tenant)).status()).toBe(201);
    const token = await tenantToken(request, tenant);
    expect((await request.post('/api/tenant/kyc/submit', { headers: { Authorization: `Bearer ${token}` } })).status()).toBe(200);

    const platform = await platformLogin(request);
    const res = await request.post(`/api/platform/tenants/${tenant.slug}/kyc/reject`, {
      headers: { Authorization: `Bearer ${platform}` },
      data: {}
    });
    expect(res.status()).toBe(400);
  });
});