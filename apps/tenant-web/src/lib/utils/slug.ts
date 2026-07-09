export function suggestSlug(companyName: string): string {
  let slug = companyName
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .replace(/-{2,}/g, '-');

  if (slug.length < 2) {
    return slug;
  }
  if (slug.length > 32) {
    slug = slug.slice(0, 32).replace(/-+$/g, '');
  }
  return slug;
}