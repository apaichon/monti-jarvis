import { defineConfig, devices } from '@playwright/test';
import { BASE_URL } from './helpers/config';

// E2E runs against the Go server serving the *built* SvelteKit apps, so the
// UI and API are same-origin — the production topology. `ensure-server.sh`
// builds + starts that server on E2E_PORT (default 8099) unless one is already
// listening. Set E2E_BASE_URL to point at a server you manage yourself.

export default defineConfig({
  testDir: './tests',
  globalSetup: './global-setup.ts',
  timeout: 30_000,
  expect: { timeout: 7_000 },
  // Shared DB + a rate-limited endpoint: run serially for deterministic state.
  fullyParallel: false,
  workers: 1,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  reporter: [['list'], ['html', { open: 'never' }]],
  use: {
    baseURL: BASE_URL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure'
  },
  projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
  webServer: {
    command: 'bash scripts/ensure-server.sh',
    url: `${BASE_URL}/healthz`,
    reuseExistingServer: true,
    timeout: 240_000,
    stdout: 'pipe',
    stderr: 'pipe'
  }
});
