/** Monogram fallback when a brand has no logo_url. */

export function brandMonogram(name?: string, slug?: string): string {
  const raw = (name || slug || '?').trim();
  if (!raw) return '?';
  const parts = raw.split(/[\s\-_/]+/).filter(Boolean);
  if (parts.length >= 2) {
    return (parts[0][0] + parts[1][0]).toUpperCase();
  }
  return raw.slice(0, 2).toUpperCase();
}
