import { test, expect } from '@playwright/test'
import { login } from './helpers.js'

test.describe('Phase 117 — create task journey', () => {
  test('opens tasks surface after login', async ({ page, request }) => {
    await login(page, request)
    await page.goto('/tasks')
    // /tasks may redirect into zones/ops; accept either tasks chrome or a tasks-related heading.
    await expect(
      page.getByText(/task/i).first(),
    ).toBeVisible({ timeout: 15_000 })
  })
})
