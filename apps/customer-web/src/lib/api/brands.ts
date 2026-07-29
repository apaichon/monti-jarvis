/** Public brand / tenant directory for customer portal entry (Sprint 54). */

export type PublicBrand = {
  id: string;
  slug: string;
  name: string;
  blurb?: string;
  logo_url?: string;
  category?: string;
  languages?: string[];
  listed?: boolean;
  status?: string;
};

export type PublicBrandList = {
  items: PublicBrand[];
  total: number;
  limit: number;
  offset: number;
};

export async function listPublicBrands(opts?: {
  q?: string;
  limit?: number;
  offset?: number;
}): Promise<PublicBrandList> {
  const params = new URLSearchParams();
  if (opts?.q) params.set('q', opts.q);
  if (opts?.limit != null) params.set('limit', String(opts.limit));
  if (opts?.offset != null) params.set('offset', String(opts.offset));
  const qs = params.toString();
  const res = await fetch(`/api/public/brands${qs ? `?${qs}` : ''}`);
  const data = await res.json().catch(() => ({}));
  if (!res.ok) throw new Error(data.error || 'Failed to load brands');
  return {
    items: Array.isArray(data.items) ? data.items : [],
    total: Number(data.total ?? 0),
    limit: Number(data.limit ?? opts?.limit ?? 50),
    offset: Number(data.offset ?? opts?.offset ?? 0)
  };
}

/** Resolve listable brand by slug or tenant id. */
export async function getPublicBrand(slugOrId: string): Promise<PublicBrand> {
  const key = encodeURIComponent(slugOrId.trim());
  const res = await fetch(`/api/public/brands/${key}`);
  const data = await res.json().catch(() => ({}));
  if (res.status === 404) throw new Error(data.error || 'brand not found');
  if (!res.ok) throw new Error(data.error || 'Failed to load brand');
  const item = data.item as PublicBrand | undefined;
  if (!item?.id || !item?.slug) throw new Error('Invalid brand response');
  return item;
}
