import { test, expect } from '@playwright/test'
import { login, apiLogin, seedPendingProposal } from './helpers.js'

test.describe('Phase 211.07 — Guardian Pending Confirm / Dismiss', () => {
  test('Confirms a seeded pending proposal from the Pending tab', async ({ page, request }) => {
    const title = `E2E confirm ${Date.now()}`
    const token = await apiLogin(request)
    const prop = await seedPendingProposal(request, token, title)

    await login(page, request)
    await page.goto('/chat?tab=pending')
    await expect(page.locator('[data-test="guardian-requests-inbox"]')).toBeVisible({ timeout: 15_000 })
    const card = page.locator(`[data-test="guardian-proposal-card"][data-proposal-id="${prop.proposal_id}"]`)
    await expect(card).toBeVisible({ timeout: 15_000 })
    await expect(card).toContainText(title)
    await card.locator('[data-test="guardian-proposal-confirm"]').click()

    await expect(card).toHaveCount(0, { timeout: 15_000 })
    await expect(page.getByText(title)).toHaveCount(0)
  })

  test('Dismisses a seeded pending proposal from the Pending tab', async ({ page, request }) => {
    const title = `E2E dismiss ${Date.now()}`
    const token = await apiLogin(request)
    const prop = await seedPendingProposal(request, token, title)

    await login(page, request)
    await page.goto('/chat?tab=pending')
    await expect(page.locator('[data-test="guardian-requests-inbox"]')).toBeVisible({ timeout: 15_000 })

    const card = page.locator(`[data-test="guardian-proposal-card"][data-proposal-id="${prop.proposal_id}"]`)
    await expect(card).toBeVisible({ timeout: 15_000 })
    await card.locator('[data-test="guardian-proposal-dismiss"]').click()

    await expect(card).toHaveCount(0, { timeout: 15_000 })
    await expect(page.getByText(title)).toHaveCount(0)
  })
})
