import { apiFetch } from '$lib/api/http';

export type ThemeBranding = {
  brand_name: string;
  subtitle: string;
  logo_url: string;
  logo_alt: string;
};

export type ThemeTokens = Record<string, string>;

export type ContrastPair = { pair: string; ratio: number; pass: boolean };

export type TenantTheme = {
  tenant_id: string;
  preset: string;
  draft_branding: ThemeBranding;
  published_branding: ThemeBranding;
  draft_tokens: ThemeTokens;
  published_tokens: ThemeTokens;
  published_at?: string | null;
  draft_updated_at?: string;
  contrast_report?: { ok: boolean; pairs: ContrastPair[] };
};

export const TOKEN_KEYS = [
  'primary',
  'primary_text',
  'accent',
  'background',
  'surface',
  'surface_elevated',
  'text',
  'muted',
  'line',
  'success',
  'warn',
  'danger'
] as const;

export function getTheme() {
  return apiFetch<TenantTheme>('/api/tenant/theme');
}

export function putTheme(body: { preset: string; branding: ThemeBranding; tokens: ThemeTokens }) {
  return apiFetch<TenantTheme>('/api/tenant/theme', {
    method: 'PUT',
    body: JSON.stringify(body)
  });
}

export function publishTheme(confirmLowContrast = false) {
  return apiFetch<TenantTheme>('/api/tenant/theme/publish', {
    method: 'POST',
    body: JSON.stringify({ confirm_low_contrast: confirmLowContrast })
  });
}

export function resetTheme(preset: string, resetBranding = false) {
  return apiFetch<TenantTheme>('/api/tenant/theme/reset', {
    method: 'POST',
    body: JSON.stringify({ preset, reset_branding: resetBranding })
  });
}

export async function uploadThemeLogo(file: File) {
  const fd = new FormData();
  fd.append('file', file);
  return apiFetch<{ logo_url: string; theme: TenantTheme }>('/api/tenant/theme/logo', {
    method: 'POST',
    body: fd
  });
}
