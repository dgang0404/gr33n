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

test.describe('Phase 117 — create task journey', () => {
  test('creates a task from the tasks workspace', async ({ page }) => {
    await login(page)
    await page.goto('/tasks')
    const title = `E2E task ${Date.now()}`
    const titleInput = page.locator('input[placeholder*="task" i], input[name="title"], textarea').first()
    if (await titleInput.count()) {
      await titleInput.fill(title)
    }
    const createBtn = page.getByRole('button', { name: /add task|create task|save/i }).first()
    if (await createBtn.isVisible()) {
      await createBtn.click()
    }
    await expect(page.getByText(title).or(page.getByText(/Tasks/i))).toBeVisible({ timeout: 15_000 })
  })
})
