import { test, expect } from '@playwright/test'
import { login } from './helpers.js'

test.describe('Phase 117 — login to dashboard', () => {
  test('authenticated session lands on Today workspace', async ({ page, request }) => {
    // Token inject (not form) — form login burns auth_test rate limit across the suite.
    await login(page, request)
    await expect(page.getByText(/Today|My zones|Grow/i).first()).toBeVisible()
  })
})
