import { test, expect } from '@playwright/test'
import { login } from './helpers.js'

test.describe('Phase 117 / 211.07 — Guardian chat shell', () => {
  test('opens Farm Guardian page and shows chat shell', async ({ page, request }) => {
    await login(page, request)
    await page.goto('/chat')
    await expect(
      page.getByText(/Farm Guardian|Guardian|Chat/i).first(),
    ).toBeVisible({ timeout: 15_000 })
    await expect(
      page.locator('[data-test="guardian-model-selector"], textarea, [data-test="guardian-chat-input"], [data-test="chat-message-input"]').first(),
    ).toBeVisible({ timeout: 15_000 })
  })

  test('Pending tab shows inbox chrome without a live LLM', async ({ page, request }) => {
    await login(page, request)
    await page.goto('/chat')
    await expect(page.locator('[data-test="guardian-tab-nav"]')).toBeVisible({ timeout: 15_000 })
    await page.locator('[data-test="guardian-tab-pending"]').click()
    await expect(page).toHaveURL(/tab=pending/)
    await expect(page.locator('[data-test="guardian-requests-inbox"]')).toBeVisible({ timeout: 15_000 })
    // Empty or populated — either is fine; prove shell + empty-state or list hooks exist.
    const empty = page.locator('[data-test="guardian-inbox-empty"]')
    const list = page.locator('[data-test="guardian-inbox-list"]')
    await expect(empty.or(list)).toBeVisible({ timeout: 15_000 })
  })
})
