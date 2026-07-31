import { defineConfig, devices } from '@playwright/test'

// Prefer localhost over 127.0.0.1 — API CORS_ORIGIN defaults to http://localhost:5173.
const baseURL = process.env.E2E_BASE_URL || 'http://localhost:5173'

export default defineConfig({
  testDir: '.',
  timeout: 60_000,
  expect: { timeout: 10_000 },
  fullyParallel: false,
  workers: 1,
  retries: process.env.CI ? 1 : 0,
  reporter: [['list']],
  use: {
    baseURL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
  },
  projects: [
    {
      name: 'chromium',
      use: { ...devices['Desktop Chrome'] },
    },
  ],
})
