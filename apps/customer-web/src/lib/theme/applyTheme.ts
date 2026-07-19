/** Apply published Monti theme tokens + branding to the caller/embed shell. */

export type ThemeBranding = {
  brand_name?: string;
  subtitle?: string;
  logo_url?: string;
  logo_alt?: string;
};

export type ThemeTokens = Record<string, string>;

export type PublicTheme = {
  preset?: string;
  source?: string;
  branding?: ThemeBranding;
  tokens?: ThemeTokens;
  css_vars?: Record<string, string>;
};

const DEFAULT_BRANDING: Required<ThemeBranding> = {
  brand_name: 'Monti',
  subtitle: 'AI · text & voice',
  logo_url: '/images/monti-logo.png',
  logo_alt: 'Monti'
};

export function resolveBranding(
  branding?: ThemeBranding | null,
  workspaceName?: string
): Required<ThemeBranding> {
  const name =
    (branding?.brand_name || '').trim() ||
    (workspaceName || '').trim() ||
    DEFAULT_BRANDING.brand_name;
  return {
    brand_name: name,
    subtitle: (branding?.subtitle || '').trim() || DEFAULT_BRANDING.subtitle,
    logo_url: (branding?.logo_url || '').trim() || DEFAULT_BRANDING.logo_url,
    logo_alt: (branding?.logo_alt || '').trim() || name
  };
}

/** Set CSS custom properties on an element (or documentElement). */
export function applyThemeTokens(
  el: HTMLElement | null | undefined,
  theme?: PublicTheme | null
): void {
  if (!el || !theme) return;
  const vars = theme.css_vars || tokenCssVars(theme.tokens);
  for (const [k, v] of Object.entries(vars)) {
    if (k.startsWith('--') && v) {
      el.style.setProperty(k, v);
    }
  }
}

export function tokenCssVars(tokens?: ThemeTokens | null): Record<string, string> {
  if (!tokens) return {};
  const m: Record<string, string> = {};
  for (const [k, v] of Object.entries(tokens)) {
    if (!v) continue;
    m[`--mj-${k.replace(/_/g, '-')}`] = v;
  }
  if (tokens.background) m['--bg'] = tokens.background;
  if (tokens.text) m['--ink'] = tokens.text;
  if (tokens.muted) m['--muted'] = tokens.muted;
  if (tokens.surface) m['--panel'] = tokens.surface;
  if (tokens.surface_elevated) m['--panel-2'] = tokens.surface_elevated;
  if (tokens.line) m['--line'] = tokens.line;
  if (tokens.accent) m['--cyan'] = tokens.accent;
  if (tokens.primary) m['--blue'] = tokens.primary;
  if (tokens.success) m['--green'] = tokens.success;
  if (tokens.danger) m['--red'] = tokens.danger;
  return m;
}

export async function fetchPublicTheme(
  apiBase: string,
  tenantId: string
): Promise<PublicTheme | null> {
  try {
    const base = apiBase.replace(/\/+$/, '');
    const res = await fetch(`${base}/api/public/theme/${encodeURIComponent(tenantId)}`, {
      credentials: 'omit'
    });
    if (!res.ok) return null;
    return (await res.json()) as PublicTheme;
  } catch {
    return null;
  }
}
