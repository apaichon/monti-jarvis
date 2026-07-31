/**
 * Normalize the platform lead-list response without renaming its API fields.
 *
 * @template T
 * @param {{items?: T[] | null, total?: number | null, limit?: number, offset?: number}} response
 * @returns {{items: T[], total: number, limit?: number, offset?: number}}
 */
export function normalizeLeadListResponse(response) {
  const items = Array.isArray(response.items) ? response.items : [];
  return {
    items,
    total: Number.isFinite(response.total) ? Number(response.total) : items.length,
    limit: response.limit,
    offset: response.offset
  };
}
