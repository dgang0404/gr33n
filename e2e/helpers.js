import { expect } from '@playwright/test'

export const devEmail = process.env.E2E_DEV_EMAIL || 'dev@gr33n.local'
export const devPass = process.env.E2E_DEV_PASSWORD || 'devpassword'
export const apiBase = process.env.E2E_API_URL || 'http://127.0.0.1:8080'
export const farmId = Number(process.env.E2E_FARM_ID || '1')

/** One JWT per worker process — avoids auth_test login rate-limit across specs. */
let cachedToken = null

/** JWT for API seed helpers (same demo user as UI login). */
export async function apiLogin(request) {
  if (cachedToken) return cachedToken
  const res = await request.post(`${apiBase}/auth/login`, {
    data: { username: devEmail, password: devPass },
  })
  expect(res.ok(), `login ${res.status()} ${await res.text()}`).toBeTruthy()
  const body = await res.json()
  expect(body.token).toBeTruthy()
  cachedToken = body.token
  return cachedToken
}

/** Authenticate UI without another /auth/login (rate-limit friendly). */
export async function login(page, request) {
  const token = await apiLogin(request)
  await page.goto('/login')
  await page.evaluate(
    ({ token: t, email, id }) => {
      localStorage.setItem('gr33n_token', t)
      localStorage.setItem('gr33n_user', email)
      localStorage.setItem('gr33n_farm_id', String(id))
    },
    { token, email: devEmail, id: farmId },
  )
  await page.goto('/')
  await expect(page).not.toHaveURL(/\/login/, { timeout: 20_000 })
  const farmSelect = page.locator('[data-test="farm-select"]')
  await expect(farmSelect).toBeEnabled({ timeout: 15_000 })
  if ((await farmSelect.inputValue()) !== String(farmId)) {
    await farmSelect.selectOption(String(farmId))
  }
  await expect(farmSelect).toHaveValue(String(farmId))
}

/** Dev/auth_test-only: insert a create_task pending proposal (no LLM). */
export async function seedPendingProposal(request, token, title) {
  const res = await request.post(`${apiBase}/v1/chat/proposals/seed-pending`, {
    headers: { Authorization: `Bearer ${token}` },
    data: { farm_id: farmId, title },
  })
  expect(res.ok(), `seed-pending ${res.status()} ${await res.text()}`).toBeTruthy()
  return res.json()
}
