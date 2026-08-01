import { apiFetch } from './http';

export type PaymentGatewayConfig = {
  configured: boolean;
  provider: 'mock' | 'chillpay' | 'stripe' | string;
  mode: string;
  status: string;
  merchant_code: string;
  api_key_masked: string;
  md5_key_set: boolean;
  base_url: string;
  route_no: number;
  currency: string;
  callback_url: string;
  return_url: string;
  connection_status: string;
  last_callback_at: string | null;
  last_test_status?: string;
  last_tested_at?: string | null;
  last_test_error?: string;
  last_webhook_status?: string;
  last_webhook_at?: string | null;
  stripe?: {
    publishable_key: string;
    secret_key_set: boolean;
    webhook_secret_set: boolean;
    api_base_url: string;
    success_url: string;
    cancel_url: string;
    callback_url: string;
  };
};

export type PaymentGatewayInput = {
  provider: string;
  mode: string;
  merchant_code: string;
  api_key?: string;
  md5_key?: string;
  base_url: string;
  route_no: number;
  currency: string;
  return_url: string;
  stripe?: {
    publishable_key?: string;
    secret_key?: string;
    webhook_secret?: string;
    api_base_url?: string;
    success_url?: string;
    cancel_url?: string;
  };
};

export function getPaymentGateway(): Promise<PaymentGatewayConfig> {
  return apiFetch<PaymentGatewayConfig>('/api/platform/payment-gateway');
}

export function updatePaymentGateway(body: PaymentGatewayInput): Promise<PaymentGatewayConfig> {
  return apiFetch<PaymentGatewayConfig>('/api/platform/payment-gateway', {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  });
}

export function testPaymentGateway(): Promise<{ ok: boolean; provider: string; message: string }> {
  return apiFetch('/api/platform/payment-gateway/test', { method: 'POST' });
}

export function reconcilePaymentGateway(body: {
  provider?: string;
  since?: string;
  limit?: number;
  dry_run?: boolean;
}): Promise<{ provider: string; dry_run: boolean; count: number; items: Array<Record<string, unknown>> }> {
  return apiFetch('/api/platform/payment-gateway/reconcile', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body)
  });
}
