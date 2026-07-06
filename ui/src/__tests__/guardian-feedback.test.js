import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import GuardianTurnFeedback from '../components/GuardianTurnFeedback.vue'

vi.mock('../api', () => ({
  default: {
    patch: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import api from '../api'

describe('Phase 134 — GuardianTurnFeedback', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('submits thumbs up', async () => {
    api.patch.mockResolvedValue({
      data: { feedback_rating: 'up', feedback_at: '2026-07-06T12:00:00Z' },
    })
    const wrapper = mount(GuardianTurnFeedback, {
      props: { sessionId: 'sess-1', turnIndex: 0, streaming: false },
    })
    await wrapper.get('[data-test="chat-feedback-up"]').trigger('click')
    await flushPromises()
    expect(api.patch).toHaveBeenCalledWith(
      '/v1/chat/sessions/sess-1/turns/0/feedback',
      { rating: 'up' },
    )
    expect(wrapper.text()).toMatch(/Thanks/)
    wrapper.unmount()
  })

  it('shows down form and submits with reason', async () => {
    api.patch.mockResolvedValue({
      data: { feedback_rating: 'down', feedback_reason: 'Missed alert' },
    })
    const wrapper = mount(GuardianTurnFeedback, {
      props: { sessionId: 'abc', turnIndex: 1, streaming: false },
    })
    await wrapper.get('[data-test="chat-feedback-down"]').trigger('click')
    expect(wrapper.find('[data-test="chat-feedback-down-form"]').exists()).toBe(true)
    await wrapper.get('[data-test="chat-feedback-reason"]').setValue('Missed alert')
    await wrapper.get('[data-test="chat-feedback-down-submit"]').trigger('click')
    await flushPromises()
    expect(api.patch).toHaveBeenCalledWith(
      '/v1/chat/sessions/abc/turns/1/feedback',
      { rating: 'down', reason: 'Missed alert' },
    )
    wrapper.unmount()
  })

  it('hides controls while streaming', () => {
    const wrapper = mount(GuardianTurnFeedback, {
      props: { sessionId: 's', turnIndex: 0, streaming: true },
    })
    expect(wrapper.find('[data-test="chat-turn-feedback"]').exists()).toBe(false)
    wrapper.unmount()
  })
})
