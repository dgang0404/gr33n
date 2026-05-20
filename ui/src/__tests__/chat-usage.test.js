// Phase 28 WS5 — Vitest coverage for the chat-usage Pinia store. The
// Settings.vue card consumes this store; rendering the full Settings
// view in a test pulls in too much surface area, so we focus on the
// store's loader contract + the derived getters that drive the card.

import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import { useChatUsageStore } from '../stores/chatUsage'
import api from '../api'

function happyResponse({ pctUser = 0.32, withFarm = false } = {}) {
  const data = {
    window_hours: 1,
    ai_enabled: true,
    user: {
      used_tokens: Math.round(10000 * pctUser),
      max_tokens: 10000,
      remaining_tokens: 10000 - Math.round(10000 * pctUser),
      pct_used: pctUser,
      warning_threshold_pct: 0.8,
    },
  }
  if (withFarm) {
    data.farm = {
      farm_id: 7,
      used_tokens: 14000,
      max_tokens: 50000,
      remaining_tokens: 36000,
      pct_used: 0.28,
    }
  }
  return { data }
}

describe('chatUsage store (Phase 28 WS5)', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('defaults to a clean, uncapped state before fetch', () => {
    const u = useChatUsageStore()
    expect(u.fetched).toBe(false)
    expect(u.hasAnyCap).toBe(false)
    expect(u.nearLimit).toBe(false)
    expect(u.atLimit).toBe(false)
    expect(u.user.used_tokens).toBe(0)
    expect(u.farm).toBeNull()
  })

  it('load() populates user dimension and flips fetched', async () => {
    api.get.mockResolvedValue(happyResponse({ pctUser: 0.4 }))
    const u = useChatUsageStore()
    await u.load()
    expect(api.get).toHaveBeenCalledWith('/v1/chat/usage', { params: {} })
    expect(u.fetched).toBe(true)
    expect(u.user.max_tokens).toBe(10000)
    expect(u.user.used_tokens).toBe(4000)
    expect(u.hasAnyCap).toBe(true)
    expect(u.nearLimit).toBe(false)
    expect(u.atLimit).toBe(false)
  })

  it('load({ farmId }) sends the query param and populates farm dimension', async () => {
    api.get.mockResolvedValue(happyResponse({ withFarm: true }))
    const u = useChatUsageStore()
    await u.load({ farmId: 7 })
    expect(api.get).toHaveBeenCalledWith('/v1/chat/usage', { params: { farm_id: 7 } })
    expect(u.farm).not.toBeNull()
    expect(u.farm.farm_id).toBe(7)
    expect(u.farm.max_tokens).toBe(50000)
  })

  it('omits farm_id when farmId is 0 or undefined', async () => {
    api.get.mockResolvedValue(happyResponse())
    const u = useChatUsageStore()
    await u.load({ farmId: 0 })
    expect(api.get).toHaveBeenCalledWith('/v1/chat/usage', { params: {} })
  })

  it('omits farm_id for NaN-ish values', async () => {
    api.get.mockResolvedValue(happyResponse())
    const u = useChatUsageStore()
    await u.load({ farmId: 'abc' })
    expect(api.get).toHaveBeenCalledWith('/v1/chat/usage', { params: {} })
  })

  it('flips nearLimit when user crosses 80%', async () => {
    api.get.mockResolvedValue(happyResponse({ pctUser: 0.85 }))
    const u = useChatUsageStore()
    await u.load()
    expect(u.nearLimit).toBe(true)
    expect(u.atLimit).toBe(false)
  })

  it('flips atLimit when user crosses 100%', async () => {
    api.get.mockResolvedValue(happyResponse({ pctUser: 1.05 }))
    const u = useChatUsageStore()
    await u.load()
    expect(u.atLimit).toBe(true)
    expect(u.nearLimit).toBe(true)
  })

  it('treats 503 as AI disabled at the server', async () => {
    const err = new Error('Service Unavailable')
    err.response = { status: 503, data: { error: 'AI is disabled on this deployment' } }
    api.get.mockRejectedValue(err)
    const u = useChatUsageStore()
    await u.load()
    expect(u.fetched).toBe(true)
    expect(u.aiEnabled).toBe(false)
    expect(u.error).toMatch(/disabled/i)
  })

  it('captures non-503 errors but stays "fetched" so the card stops looping', async () => {
    const err = new Error('boom')
    err.response = { status: 500, data: { error: 'internal' } }
    api.get.mockRejectedValue(err)
    const u = useChatUsageStore()
    await u.load()
    expect(u.fetched).toBe(true)
    expect(u.error).toBe('internal')
    expect(u.aiEnabled).toBe(true) // 500 doesn't flip AI off
  })

  it('hasAnyCap is true when only farm dimension has a cap', async () => {
    api.get.mockResolvedValue({
      data: {
        window_hours: 1,
        ai_enabled: true,
        user: { used_tokens: 0, max_tokens: 0, remaining_tokens: 0, pct_used: 0 },
        farm: { farm_id: 7, used_tokens: 100, max_tokens: 5000, remaining_tokens: 4900, pct_used: 0.02 },
      },
    })
    const u = useChatUsageStore()
    await u.load({ farmId: 7 })
    expect(u.hasAnyCap).toBe(true)
  })

  it('hasAnyCap is false when neither dimension has a cap', async () => {
    api.get.mockResolvedValue({
      data: {
        window_hours: 1,
        ai_enabled: true,
        user: { used_tokens: 0, max_tokens: 0, remaining_tokens: 0, pct_used: 0 },
        farm: { farm_id: 7, used_tokens: 0, max_tokens: 0, remaining_tokens: 0, pct_used: 0 },
      },
    })
    const u = useChatUsageStore()
    await u.load({ farmId: 7 })
    expect(u.hasAnyCap).toBe(false)
  })
})
