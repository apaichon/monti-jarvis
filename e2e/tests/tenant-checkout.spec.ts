import { test, expect } from '../fixtures';
import { registerViaApi, verifyViaApi, platformLogin } from '../helpers/api';
import * as db from '../helpers/db';

async function tenantToken(request: import('@playwright/test').APIRequestContext, tenant: { email: string; password: string }) {
  const userId = await db.getUserIdByEmail(tenant.email);
  const rawToken = 'e2e-checkout-' + Math.random().toString(36).slice(2);
  await db.insertVerificationToken(userId!, rawToken);
  const verified = await verifyViaApi(request, rawToken);
  expect(verified.status()).toBe(200);
  return (await verified.json()).access_token as string;
}

async function activateTenant(
  request: import('@playwright/test').APIRequestContext,
  tenant: { slug: string; email: string; password: string }
) {
  const token = await tenantToken(request, tenant);
  expect((await request.post('/api/tenant/kyc/submit', { headers: { Authorization: `Bearer ${token}` } })).status()).toBe(200);
  const platform = await platformLogin(request);
  const approve = await request.post(`/api/platform/tenants/${tenant.slug}/kyc/approve`, {
    headers: { Authorization: `Bearer ${platform}` }
  });
  expect(approve.status()).toBe(200);
  return token;
}

test.describe('Tenant checkout (SPRINT-009)', () => {
  test.beforeEach(({ dbReady }) => {
    test.skip(!dbReady, 'requires Postgres (set POSTGRES_URL / run make infra-init)');
  });

  test('mock gateway checkout assigns entitlement', async ({ request, tenant }) => {
    expect((await registerViaApi(request, tenant)).status()).toBe(201);
    const token = await activateTenant(request, tenant);

    const platform = await platformLogin(request);
    const gw = await request.put('/api/platform/payment-gateway', {
      headers: { Authorization: `Bearer ${platform}`, 'Content-Type': 'application/json' },
      data: {
        provider: 'mock',
        mode: 'test',
        merchant_code: 'MOCK',
        base_url: 'http://localhost',
        route_no: 1,
        currency: '764',
        return_url: 'http://localhost:8091/tenant/billing/return'
      }
    });
    expect(gw.status()).toBe(200);

    const catalog = await request.get('/api/tenant/packages', {
      headers: { Authorization: `Bearer ${token}` }
    });
    expect(catalog.status()).toBe(200);
    const catalogBody = await catalog.json();
    expect(catalogBody.packages?.length).toBeGreaterThan(0);

    const catalogMethods = catalogBody.payment_methods as { id: string }[] | undefined;
    expect(catalogMethods?.some((m) => m.id === 'credit_card')).toBeTruthy();
    expect(catalogMethods?.some((m) => m.id === 'qr_promptpay')).toBeTruthy();

    const checkout = await request.post('/api/tenant/checkout', {
      headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
      data: { package_id: 'pkg-pro', payment_method: 'qr_promptpay' }
    });
    expect(checkout.status()).toBe(200);
    const order = await checkout.json();
    expect(order.status).toBe('pending');
    expect(order.payment_method).toBe('qr_promptpay');
    expect(order.payment_url).toContain('mock-pay');

    const mockPay = await request.post(`/api/dev/mock-pay/${order.order_id}`, {
      headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/json' },
      data: { result: 'success' }
    });
    expect(mockPay.status()).toBe(200);
    const paid = await mockPay.json();
    expect(paid.status).toBe('paid');
    expect(paid.documents?.length).toBeGreaterThanOrEqual(2);

    const poll = await request.get(`/api/tenant/orders/${order.order_id}`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    expect(poll.status()).toBe(200);
    const polled = await poll.json();
    expect(polled.status).toBe('paid');
    expect(polled.payment_method).toBe('qr_promptpay');
    expect(polled.documents?.some((d: { doc_type: string }) => d.doc_type === 'receipt')).toBeTruthy();
    expect(polled.documents?.some((d: { doc_type: string }) => d.doc_type === 'tax_invoice')).toBeTruthy();

    const receipt = await request.get(`/api/tenant/orders/${order.order_id}/documents/receipt`, {
      headers: { Authorization: `Bearer ${token}` }
    });
    expect(receipt.status()).toBe(200);
    const receiptBody = await receipt.json();
    expect(receiptBody.doc_number).toContain('RCP-');

    const taxHtml = await request.get(
      `/api/tenant/orders/${order.order_id}/documents/tax_invoice?format=html`,
      { headers: { Authorization: `Bearer ${token}` } }
    );
    expect(taxHtml.status()).toBe(200);
    expect(await taxHtml.text()).toContain('Tax Invoice');

    const ent = await request.get('/api/entitlements/me', {
      headers: { Authorization: `Bearer ${token}` }
    });
    expect(ent.status()).toBe(200);
    const entBody = await ent.json();
    expect(entBody.package?.id).toBe('pkg-pro');
  });
});