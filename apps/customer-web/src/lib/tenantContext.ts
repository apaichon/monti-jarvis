/** Client-only selected tenant for customer portal (Sprint 54). */

export type SelectedTenant = {
  id: string;
  slug: string;
  name: string;
};

export const SELECTED_TENANT_KEY = 'monti_jarvis:selected_tenant';

export function getSelectedTenant(): SelectedTenant | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = window.sessionStorage.getItem(SELECTED_TENANT_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as SelectedTenant;
    if (!parsed?.id || !parsed?.slug) return null;
    return {
      id: String(parsed.id),
      slug: String(parsed.slug),
      name: String(parsed.name || parsed.slug)
    };
  } catch {
    return null;
  }
}

export function setSelectedTenant(tenant: SelectedTenant): void {
  if (typeof window === 'undefined') return;
  window.sessionStorage.setItem(
    SELECTED_TENANT_KEY,
    JSON.stringify({
      id: tenant.id,
      slug: tenant.slug,
      name: tenant.name || tenant.slug
    })
  );
}

export function clearSelectedTenant(): void {
  if (typeof window === 'undefined') return;
  try {
    window.sessionStorage.removeItem(SELECTED_TENANT_KEY);
  } catch {
    /* ignore */
  }
}
