import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'

const seedSessions = [
  {
    session_id: '11111111-1111-1111-1111-111111111111',
    title: 'Session A',
    first_user_message: 'how is zone 1?',
    turn_count: 2,
    any_grounded: false,
    last_turn_at: '2026-05-19T17:30:00Z',
    total_prompt_tokens: 10,
    total_completion_tokens: 12,
  },
  {
    session_id: '22222222-2222-2222-2222-222222222222',
    title: 'Session B',
    first_user_message: 'fertigation plan?',
    turn_count: 5,
    any_grounded: true,
    last_turn_at: '2026-05-19T18:00:00Z',
    total_prompt_tokens: 40,
    total_completion_tokens: 50,
  },
  {
    session_id: '33333333-3333-3333-3333-333333333333',
    title: 'Session C',
    first_user_message: 'tasks today?',
    turn_count: 1,
    any_grounded: false,
    last_turn_at: '2026-05-19T18:15:00Z',
    total_prompt_tokens: 5,
    total_completion_tokens: 6,
  },
]

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import FarmGuardianChat from '../views/FarmGuardianChat.vue'
import api from '../api'

const router = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/chat', component: FarmGuardianChat }],
})

describe('FarmGuardianChat — bulk delete (Phase 27 WS6 follow-up)', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    api.get.mockImplementation((url) => {
      if (url === '/capabilities') return Promise.resolve({ data: { ai_enabled: true } })
      if (url === '/v1/chat/sessions') return Promise.resolve({ data: { sessions: seedSessions } })
      return Promise.resolve({ data: {} })
    })
  })

  async function mountReady() {
    await router.push('/chat')
    const wrapper = mount(FarmGuardianChat, { global: { plugins: [router] } })
    await flushPromises()
    await flushPromises()
    return wrapper
  }

  it('Select toggles checkbox UI and hides per-row rename/delete buttons', async () => {
    const wrapper = await mountReady()

    expect(wrapper.find('[data-test="chat-session-rename"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="chat-bulk-toolbar"]').exists()).toBe(false)

    await wrapper.find('[data-test="chat-bulk-select"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-test="chat-bulk-toolbar"]').exists()).toBe(true)
    expect(wrapper.findAll('[data-test="chat-session-checkbox"]')).toHaveLength(seedSessions.length)
    expect(wrapper.find('[data-test="chat-session-rename"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="chat-session-delete"]').exists()).toBe(false)
  })

  it('counts selections and Delete N reflects the count', async () => {
    const wrapper = await mountReady()
    await wrapper.find('[data-test="chat-bulk-select"]').trigger('click')
    await flushPromises()
    const boxes = wrapper.findAll('[data-test="chat-session-checkbox"]')
    await boxes[0].setValue(true)
    await boxes[2].setValue(true)
    await flushPromises()

    expect(wrapper.find('[data-test="chat-bulk-count"]').text()).toContain('2')
    expect(wrapper.find('[data-test="chat-bulk-delete"]').text()).toContain('Delete 2')
  })

  it('confirms then DELETEs each selected session and clears the active session if it was deleted', async () => {
    api.delete.mockResolvedValue({ data: {} })
    const wrapper = await mountReady()

    // Pre-load Session B as the active session via the loadSession path.
    api.get.mockImplementationOnce((url) => {
      if (url === '/v1/chat/sessions/' + seedSessions[1].session_id) {
        return Promise.resolve({ data: { turns: [{ user_message: 'hi', assistant_message: 'hi back', turn_index: 0 }] } })
      }
      return Promise.resolve({ data: {} })
    })
    await wrapper.findAll('[data-test="chat-sessions"] li')[1]
      .find('div.cursor-pointer')
      .trigger('click')
    await flushPromises()

    // Enter select mode and pick A + B (B is the active one).
    await wrapper.find('[data-test="chat-bulk-select"]').trigger('click')
    await flushPromises()
    const boxes = wrapper.findAll('[data-test="chat-session-checkbox"]')
    await boxes[0].setValue(true)
    await boxes[1].setValue(true)
    await flushPromises()

    // Open confirm + submit.
    await wrapper.find('[data-test="chat-bulk-delete"]').trigger('click')
    await flushPromises()
    expect(wrapper.find('[data-test="chat-bulk-confirm"]').exists()).toBe(true)

    await wrapper.find('[data-test="chat-bulk-confirm"] form').trigger('submit.prevent')
    await flushPromises()

    expect(api.delete).toHaveBeenCalledTimes(2)
    expect(api.delete).toHaveBeenCalledWith('/v1/chat/sessions/' + seedSessions[0].session_id)
    expect(api.delete).toHaveBeenCalledWith('/v1/chat/sessions/' + seedSessions[1].session_id)

    // Sidebar now shows only Session C.
    const remaining = wrapper.findAll('[data-test="chat-sessions"] li')
    expect(remaining).toHaveLength(1)
    expect(remaining[0].text()).toContain('Session C')

    // Select mode + modal exited.
    expect(wrapper.find('[data-test="chat-bulk-toolbar"]').exists()).toBe(false)
    expect(wrapper.find('[data-test="chat-bulk-confirm"]').exists()).toBe(false)

    // Active session (Session B) was among the deleted, so transcript is cleared.
    expect(wrapper.find('[data-test="chat-transcript"]').exists()).toBe(false)
  })

  it('keeps failed rows selected and surfaces an error when some DELETEs fail', async () => {
    api.delete.mockImplementation((url) => {
      if (url.endsWith(seedSessions[1].session_id)) {
        return Promise.reject(new Error('boom'))
      }
      return Promise.resolve({ data: {} })
    })
    const wrapper = await mountReady()

    await wrapper.find('[data-test="chat-bulk-select"]').trigger('click')
    await flushPromises()
    const boxes = wrapper.findAll('[data-test="chat-session-checkbox"]')
    await boxes[0].setValue(true)
    await boxes[1].setValue(true)
    await flushPromises()
    await wrapper.find('[data-test="chat-bulk-delete"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-test="chat-bulk-confirm"] form').trigger('submit.prevent')
    await flushPromises()

    // Confirm modal stays open with the error message.
    expect(wrapper.find('[data-test="chat-bulk-confirm"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="chat-bulk-error"]').text()).toContain('Failed to delete 1 of 2')

    // Session A was deleted, Session B remains (failure kept it).
    const remaining = wrapper.findAll('[data-test="chat-sessions"] li')
    expect(remaining.map((li) => li.text())).toEqual(expect.arrayContaining([
      expect.stringContaining('Session B'),
      expect.stringContaining('Session C'),
    ]))
    // Selection is now just the failed id (so the operator can retry it directly).
    expect(wrapper.find('[data-test="chat-bulk-count"]').text()).toContain('1')
  })

  it('Cancel exits select mode without firing any DELETE', async () => {
    const wrapper = await mountReady()
    await wrapper.find('[data-test="chat-bulk-select"]').trigger('click')
    await flushPromises()
    const boxes = wrapper.findAll('[data-test="chat-session-checkbox"]')
    await boxes[0].setValue(true)
    await wrapper.find('[data-test="chat-bulk-cancel"]').trigger('click')
    await flushPromises()

    expect(api.delete).not.toHaveBeenCalled()
    expect(wrapper.find('[data-test="chat-bulk-toolbar"]').exists()).toBe(false)
    // Per-row buttons are back.
    expect(wrapper.find('[data-test="chat-session-rename"]').exists()).toBe(true)
  })

  it('Select all picks every row', async () => {
    const wrapper = await mountReady()
    await wrapper.find('[data-test="chat-bulk-select"]').trigger('click')
    await flushPromises()
    await wrapper.find('[data-test="chat-bulk-select-all"]').trigger('click')
    await flushPromises()

    expect(wrapper.find('[data-test="chat-bulk-count"]').text()).toContain(`${seedSessions.length}`)
    expect(wrapper.find('[data-test="chat-bulk-delete"]').text()).toContain(`Delete ${seedSessions.length}`)
  })
})
