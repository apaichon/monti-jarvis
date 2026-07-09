import { apiFetch } from '$lib/api/http';
import type { TokenPair } from '$lib/auth/session';

export function login(email: string, password: string) {
  return apiFetch<TokenPair>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password })
  });
}