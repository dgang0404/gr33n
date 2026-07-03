import { test, expect } from '@playwright/test'

const devEmail = process.env.E2E_DEV_EMAIL || 'dev@gr33n.local'
const devPass = process.env.E2E_DEV_PASSWORD || 'devpassword'

async function login(page) {
  await page.goto('/login')
  await page.getByPlaceholder(/admin|you@example/i).fill(devEmail)
  await page.locator('input[type="password"]').fill(devPass)
  await page.getByRole('button', { name: /sign in|log in/i }).click()
  await expect(page).toHaveURL(/\//, { timeout: 20_000 })
}

test.describe('Phase 117 — Guardian chat shell', () => {
  test('opens Farm Guardian page and shows chat shell', async ({ page }) => {
    await login(page)
    await page.goto('/chat')
    await expect(
      page.getByText(/Farm Guardian|Guardian|Chat/i).first(),
    ).toBeVisible({ timeout: 15_000 })
    await expect(
      page.locator('[data-test="guardian-model-selector"], textarea, [data-test="guardian-chat-input"]').first(),
    ).toBeVisible({ timeout: 15_000 })
  })
})
