import { test, expect } from '../fixtures';
import { registerViaApi } from '../helpers/api';
import { uniqueTenant, type TenantData } from '../helpers/data';
import type { Page } from '@playwright/test';

// SPRINT-006 · TASK-0026 (register API + validation) · TASK-0029 (tenant-web form)

async function openEmailForm(page: Page) {
  await page.goto('/tenant/register');
  await page.getByRole('button', { name: 'Continue with email' }).click();
  await expect(page.locator('#password')).toBeVisible();
}

async function fillEmailForm(page: Page, t: TenantData) {
  await openEmailForm(page);
  await page.locator('#company').fill(t.companyName);
  await page.locator('#slug').fill(t.slug);
  await page.locator('#email').fill(t.email);
  await page.locator('#display').fill(t.displayName);
  await page.locator('#password').fill(t.password);
  await page.locator('#confirm').fill(t.password);
}

test.describe('Registration form + client validation (TASK-0029)', () => {
  test('renders every required field (AC #2)', async ({ page }) => {
    await openEmailForm(page);
    for (const id of ['#company', '#slug', '#email', '#display', '#password', '#confirm']) {
      await expect(page.locator(id)).toBeVisible();
    }
    await expect(page.getByRole('button', { name: 'Create account with email' })).toBeVisible();
  });

  test('auto-suggests the workspace slug from the company name (AC #2)', async ({ page }) => {
    await openEmailForm(page);
    await page.locator('#company').fill('Acme QA Widgets');
    await expect(page.locator('#slug')).toHaveValue('acme-qa-widgets');
  });

  test('rejects a password shorter than 8 characters (AC #3)', async ({ page, tenant }) => {
    await fillEmailForm(page, { ...tenant, password: 'short12' });
    await page.getByRole('button', { name: 'Create account with email' }).click();
    await expect(page.locator('p.error')).toContainText('at least 8 characters');
  });

  test('rejects mismatched passwords (AC #3)', async ({ page, tenant }) => {
    await fillEmailForm(page, tenant);
    await page.locator('#confirm').fill('different-Passw0rd!');
    await page.getByRole('button', { name: 'Create account with email' }).click();
    await expect(page.locator('p.error')).toContainText('Passwords do not match');
  });

  test('rejects a submit with missing fields (AC #3)', async ({ page }) => {
    await openEmailForm(page);
    await page.getByRole('button', { name: 'Create account with email' }).click();
    await expect(page.locator('p.error')).toContainText('All fields are required');
  });
});

test.describe('Registration server flow (TASK-0026)', () => {
  test.beforeEach(({ dbReady }) => {
    test.skip(!dbReady, 'requires Postgres (set POSTGRES_URL / run make infra-init)');
  });

  test('a valid submission creates the tenant and asks the user to verify email (AC #1, #4)', async ({
    page,
    tenant
  }) => {
    await fillEmailForm(page, tenant);
    await page.getByRole('button', { name: 'Create account with email' }).click();

    await expect(page).toHaveURL(/\/tenant\/register\/check-email/);
    await expect(page.getByRole('heading', { name: 'Check your email' })).toBeVisible();
    await expect(page.getByText(tenant.email)).toBeVisible();
  });

  test('a duplicate slug is rejected with a 409 message (AC #2)', async ({ page, request, tenant }) => {
    // Pre-create the tenant, then attempt the same slug with a fresh email.
    expect((await registerViaApi(request, tenant)).status()).toBe(201);

    await fillEmailForm(page, { ...uniqueTenant(), slug: tenant.slug });
    await page.getByRole('button', { name: 'Create account with email' }).click();
    await expect(page.locator('p.error')).toContainText('slug already taken');
  });

  test('a duplicate email is rejected with a 409 message (AC #3)', async ({ page, request, tenant }) => {
    expect((await registerViaApi(request, tenant)).status()).toBe(201);

    await fillEmailForm(page, { ...uniqueTenant(), email: tenant.email });
    await page.getByRole('button', { name: 'Create account with email' }).click();
    await expect(page.locator('p.error')).toContainText('email already registered');
  });
});
