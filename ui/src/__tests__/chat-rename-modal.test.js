import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'

const seedSession = {
  session_id: '11111111-1111-1111-1111-111111111111',
  title: 'My current title',
  first_user_message: 'why did zone 3 alert?',
  turn_count: 3,
  any_grounded: true,
  last_turn_at: '2026-05-19T17:30:00Z',
  total_prompt_tokens: 12,
  total_completion_tokens: 18,
}

vi.mock('../api', () => ({
  default: {
    get: vi.fn((url) => {
      if (url === '/capabilities') return Promise.resolve({ data: { ai_enabled: true } })
      if (url === '/v1/chat/sessions') return Promise.resolve({ data: { sessions: [seedSession] } })
      return Promise.resolve({ data: {} })
    }),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import FarmGuardianChat from '../views/FarmGuardianChat.vue'
import api from '../api'

describe('FarmGuardianChat — inline rename modal (Phase 27 WS6 follow-up)', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    api.get.mockImplementation((url) => {
      if (url === '/capabilities') return Promise.resolve({ data: { ai_enabled: true } })
      if (url === '/v1/chat/sessions') return Promise.resolve({ data: { sessions: [seedSession] } })
      return Promise.resolve({ data: {} })
    })
  })

  async function mountReady() {
    const wrapper = mount(FarmGuardianChat)
    await flushPromises()
    await flushPromises()
    return wrapper
  }

  it('opens the modal pre-filled with the current title (no window.prompt)', async () => {
    const promptSpy = vi.spyOn(window, 'prompt').mockImplementation(() => null)
    const wrapper = await mountReady()

    expect(wrapper.find('[data-test="chat-rename-modal"]').exists()).toBe(false)

    await wrapper.find('[data-test="chat-session-rename"]').trigger('click')
    await flushPromises()

    const modal = wrapper.find('[data-test="chat-rename-modal"]')
    expect(modal.exists()).toBe(true)
    const input = wrapper.find('[data-test="chat-rename-input"]')
    expect(input.element.value).toBe('My current title')
    expect(promptSpy).not.toHaveBeenCalled()
    promptSpy.mockRestore()
  })

  it('PATCHes the new title and closes the modal on save', async () => {
    api.patch.mockResolvedValueOnce({ data: { title: 'Zone 3 troubleshooting' } })
    const wrapper = await mountReady()

    await wrapper.find('[data-test="chat-session-rename"]').trigger('click')
    await flushPromises()
    const input = wrapper.find('[data-test="chat-rename-input"]')
    await input.setValue('Zone 3 troubleshooting')
    // Form submit (also fires when the user presses Enter in the input).
    await wrapper.find('[data-test="chat-rename-modal"] form').trigger('submit.prevent')
    await flushPromises()

    expect(api.patch).toHaveBeenCalledWith(
      '/v1/chat/sessions/' + seedSession.session_id,
      { title: 'Zone 3 troubleshooting' },
    )
    expect(wrapper.find('[data-test="chat-rename-modal"]').exists()).toBe(false)
    // Sidebar reflects the new title.
    expect(wrapper.find('[data-test="chat-sessions"]').text()).toContain('Zone 3 troubleshooting')
  })

  it('closes without saving when cancel is clicked', async () => {
    const wrapper = await mountReady()

    await wrapper.find('[data-test="chat-session-rename"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-test="chat-rename-input"]').setValue('shouldnt persist')
    await wrapper.find('[data-test="chat-rename-cancel"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-test="chat-rename-modal"]').exists()).toBe(false)
    expect(api.patch).not.toHaveBeenCalled()
  })

  it('shows the API error inside the modal and keeps it open', async () => {
    const apiError = Object.assign(new Error('title too long'), {
      response: { data: { error: 'title exceeds 120 characters' } },
    })
    api.patch.mockRejectedValueOnce(apiError)
    const wrapper = await mountReady()

    await wrapper.find('[data-test="chat-session-rename"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-test="chat-rename-input"]').setValue('x'.repeat(150))
    await wrapper.find('[data-test="chat-rename-modal"] form').trigger('submit.prevent')
    await flushPromises()

    expect(wrapper.find('[data-test="chat-rename-modal"]').exists()).toBe(true)
    const err = wrapper.find('[data-test="chat-rename-error"]')
    expect(err.exists()).toBe(true)
    expect(err.text()).toContain('title exceeds 120 characters')
    // Sidebar still shows the original title — no optimistic write.
    expect(wrapper.find('[data-test="chat-sessions"]').text()).toContain('My current title')
  })

  it('sends an empty string to clear the title (backend nulls it)', async () => {
    api.patch.mockResolvedValueOnce({ data: { title: null } })
    const wrapper = await mountReady()

    await wrapper.find('[data-test="chat-session-rename"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-test="chat-rename-input"]').setValue('')
    await wrapper.find('[data-test="chat-rename-modal"] form').trigger('submit.prevent')
    await flushPromises()

    expect(api.patch).toHaveBeenCalledWith(
      '/v1/chat/sessions/' + seedSession.session_id,
      { title: '' },
    )
    // After clearing, sidebar falls back to the first user message.
    expect(wrapper.find('[data-test="chat-sessions"]').text()).toContain('why did zone 3 alert?')
  })
})
