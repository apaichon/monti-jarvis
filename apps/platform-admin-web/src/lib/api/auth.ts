import { apiFetch } from './http';
import type { TokenPair, UserProfile } from '$lib/auth/session';

export function login(email: string, password: string) {
  return apiFetch<TokenPair>('/api/auth/login', {
    method: 'POST',
    body: JSON.stringify({ email, password })
  });
}

export function logout(refreshToken: string) {
  return apiFetch<{ status: string }>('/api/auth/logout', {
    method: 'POST',
    body: JSON.stringify({ refresh_token: refreshToken })
  });
}

export function me() {
  return apiFetch<UserProfile>('/api/auth/me');
}