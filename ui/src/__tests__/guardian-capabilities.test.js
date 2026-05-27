/**
 * Phase 30 WS8 — capabilities flags for Guardian / vision UI.
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import { useCapabilitiesStore } from '../stores/capabilities'

vi.mock('../api', () => ({
  default: { get: vi.fn() },
}))

import api from '../api'

describe('Phase 30 WS8 — capabilities store', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('loads vision_chat_enabled from /capabilities', async () => {
    api.get.mockResolvedValue({
      data: { ai_enabled: true, vision_chat_enabled: true },
    })
    const store = useCapabilitiesStore()
    await store.fetch()
    expect(store.aiEnabled).toBe(true)
    expect(store.visionChatEnabled).toBe(true)
    expect(store.loaded).toBe(true)
  })

  it('defaults vision off when field missing (back-compat)', async () => {
    api.get.mockResolvedValue({ data: { ai_enabled: true } })
    const store = useCapabilitiesStore()
    await store.fetch()
    expect(store.visionChatEnabled).toBe(false)
  })
})
