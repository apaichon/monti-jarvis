import { apiFetch } from './http';

export type PaymentGatewayConfig = {
  configured: boolean;
  provider: string;
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