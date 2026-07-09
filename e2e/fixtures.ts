import { test as base, expect } from '@playwright/test';
import { uniqueTenant, type TenantData } from './helpers/data';
import * as db from './helpers/db';

type Fixtures = {
  // Unique registration input; its DB rows are cleaned up after the test.
  tenant: TenantData;
  // True when a Postgres connection is reachable (gates DB-dependent specs).
  dbReady: boolean;
};

export const test = base.extend<Fixtures>({
  dbReady: async ({}, use) => {
    await use(await db.dbAvailable());
  },

  tenant: async ({}, use) => {
    const t = uniqueTenant();
    await use(t);
    try {
      await db.cleanup(t.email, t.slug);
    } catch {
      // best effort
    }
  }
});

export { expect };
