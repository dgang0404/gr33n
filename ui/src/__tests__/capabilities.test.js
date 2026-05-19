import { describe, it, expect, beforeEach, vi } from 'vitest'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import { useCapabilitiesStore } from '../stores/capabilities'
import api from '../api'

describe('capabilities store (Phase 27 WS6)', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('defaults to AI on before fetch completes', () => {
    const caps = useCapabilitiesStore()
    expect(caps.aiEnabled).toBe(true)
    expect(caps.loaded).toBe(false)
    expect(caps.isLite).toBe(false)
  })

  it('reads ai_enabled from /capabilities', async () => {
    api.get.mockResolvedValue({ data: { ai_enabled: false } })
    const caps = useCapabilitiesStore()
    await caps.fetch()
    expect(api.get).toHaveBeenCalledWith('/capabilities')
    expect(caps.aiEnabled).toBe(false)
    expect(caps.loaded).toBe(true)
    expect(caps.isLite).toBe(true)
  })

  it('treats missing flag as AI on', async () => {
    api.get.mockResolvedValue({ data: {} })
    const caps = useCapabilitiesStore()
    await caps.fetch()
    expect(caps.aiEnabled).toBe(true)
    expect(caps.isLite).toBe(false)
  })

  it('falls back to AI on when endpoint errors (older API)', async () => {
    api.get.mockRejectedValue(new Error('404 not found'))
    const caps = useCapabilitiesStore()
    await caps.fetch()
    expect(caps.aiEnabled).toBe(true)
    expect(caps.loaded).toBe(true)
    expect(caps.fetchError).toContain('404')
    expect(caps.isLite).toBe(false)
  })
})
