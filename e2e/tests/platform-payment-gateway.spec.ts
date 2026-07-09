import { test, expect } from '../fixtures';
import { platformLogin } from '../helpers/api';

test.describe('Platform payment gateway (SPRINT-008)', () => {
  test('configures mock provider and tests connection', async ({ request }) => {
    const platform = await platformLogin(request);

    const put = await request.put('/api/platform/payment-gateway', {
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
    expect(put.status()).toBe(200);
    const saved = await put.json();
    expect(saved.provider).toBe('mock');
    expect(saved.configured).toBe(true);

    const testRes = await request.post('/api/platform/payment-gateway/test', {
      headers: { Authorization: `Bearer ${platform}` }
    });
    expect(testRes.status()).toBe(200);
    const testBody = await testRes.json();
    expect(testBody.ok).toBe(true);

    const infra = await request.get('/api/infra');
    expect(infra.status()).toBe(200);
    const infraBody = await infra.json();
    expect(infraBody.payment_gateway?.provider).toBe('mock');
  });
});