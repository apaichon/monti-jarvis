// Unique, valid registration input per test. Slug obeys the server rule
// `^[a-z0-9]([a-z0-9-]{1,30}[a-z0-9])?$` and stays under 32 chars.

export type TenantData = {
  companyName: string;
  slug: string;
  email: string;
  password: string;
  displayName: string;
};

let seq = 0;

export function uniqueTenant(prefix = 'e2e'): TenantData {
  seq += 1;
  const stamp = (Date.now().toString(36) + seq.toString(36)).toLowerCase();
  const slug = `${prefix}${stamp}`.replace(/[^a-z0-9]/g, '').slice(0, 30);
  return {
    companyName: `E2E QA ${stamp}`,
    slug,
    email: `${slug}@e2e.test`,
    password: 'e2e-Passw0rd!',
    displayName: 'E2E Admin'
  };
}
