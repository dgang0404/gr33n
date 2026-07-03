import { test, expect } from '@playwright/test'

const devEmail = process.env.E2E_DEV_EMAIL || 'dev@gr33n.local'
const devPass = process.env.E2E_DEV_PASSWORD || 'devpassword'

test.describe('Phase 117 — login to dashboard', () => {
  test('signs in and lands on Today workspace', async ({ page }) => {
    await page.goto('/login')
    await page.getByPlaceholder(/admin|you@example/i).fill(devEmail)
    await page.locator('input[type="password"]').fill(devPass)
    await page.getByRole('button', { name: /sign in|log in/i }).click()
    await expect(page).toHaveURL(/\//, { timeout: 20_000 })
    await expect(page.getByText(/Today|My zones|Grow/i).first()).toBeVisible()
  })
})
