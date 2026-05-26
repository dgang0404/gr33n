import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { nextTick } from 'vue'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import api from '../api'
import GuardianActionProposal from '../components/GuardianActionProposal.vue'

const router = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/alerts', component: { template: '<div />' } }],
})

const baseProposal = {
  proposal_id: '550e8400-e29b-41d4-a716-446655440000',
  tool: 'ack_alert',
  args: { alert_id: 4 },
  summary: 'Acknowledge: Humidity high — Flower Room',
  expires_at: new Date(Date.now() + 300_000).toISOString(),
  status: 'pending',
  confirmSummary: '',
  error: '',
}

describe('GuardianActionProposal (Phase 29 WS4)', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  function mountCard(overrides = {}, canOperate = true) {
    return mount(GuardianActionProposal, {
      props: {
        proposal: { ...baseProposal, ...overrides },
        canOperate,
      },
      global: { plugins: [router] },
    })
  }

  it('renders summary and Confirm/Dismiss for pending proposals', () => {
    const wrapper = mountCard()
    expect(wrapper.find('[data-test="guardian-proposal-card"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Humidity high')
    expect(wrapper.find('[data-test="guardian-proposal-confirm"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="guardian-proposal-dismiss"]').exists()).toBe(true)
  })

  it('disables Confirm when canOperate is false (viewer)', () => {
    const wrapper = mountCard({}, false)
    const btn = wrapper.find('[data-test="guardian-proposal-confirm"]')
    expect(btn.attributes('disabled')).toBeDefined()
    expect(btn.attributes('title')).toContain('Operators only')
  })

  it('shows high-risk warning for actuator enqueue', () => {
    const wrapper = mountCard({
      risk_tier: 'high',
      tool: 'enqueue_actuator_command',
      args: { device_id: 1, actuator_id: 2, command: 'on' },
      summary: 'Turn on Veg Room Grow Light',
    })
    expect(wrapper.find('[data-test="guardian-proposal-high-warning"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('device #1')
  })

  it('shows high-risk warning copy and red styling', () => {
    const wrapper = mountCard({
      risk_tier: 'high',
      tool: 'apply_bootstrap_template',
      args: { template: 'jadam_indoor_photoperiod_v1' },
      summary: 'Apply bootstrap: jadam_indoor_photoperiod_v1',
    })
    expect(wrapper.find('[data-test="guardian-proposal-high-warning"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="guardian-proposal-risk-badge"]').text()).toContain('high')
    expect(wrapper.find('[data-test="guardian-proposal-card"]').classes().join(' ')).toMatch(/red/)
  })

  it('shows medium-tier diff summary of frozen args', () => {
    const wrapper = mountCard({
      risk_tier: 'medium',
      tool: 'create_task',
      args: { title: 'Check humidity', zone_id: 2 },
      summary: 'Create task: Check humidity',
    })
    expect(wrapper.find('[data-test="guardian-proposal-diff"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('title: Check humidity')
    expect(wrapper.find('[data-test="guardian-proposal-high-warning"]').exists()).toBe(false)
  })

  it('POST /v1/chat/confirm on Confirm and shows done state', async () => {
    api.post.mockResolvedValueOnce({
      data: { summary: 'Alert acknowledged (#4).', result: { alert_id: 4, is_acknowledged: true } },
    })
    const wrapper = mountCard()
    await wrapper.find('[data-test="guardian-proposal-confirm"]').trigger('click')
    await flushPromises()
    await nextTick()

    expect(api.post).toHaveBeenCalledWith('/v1/chat/confirm', {
      proposal_id: baseProposal.proposal_id,
    })
    expect(wrapper.emitted('confirmed')).toBeTruthy()
    expect(wrapper.find('[data-test="guardian-proposal-done"]').exists()).toBe(true)
  })

  it('emits dismissed without calling API', async () => {
    const wrapper = mountCard()
    await wrapper.find('[data-test="guardian-proposal-dismiss"]').trigger('click')
    await flushPromises()
    await nextTick()
    expect(api.post).not.toHaveBeenCalled()
    expect(wrapper.emitted('dismissed')).toBeTruthy()
    expect(wrapper.text()).toContain('Dismissed')
  })

  it('surfaces confirm errors on the card', async () => {
    api.post.mockRejectedValueOnce({ response: { data: { error: 'proposal expired' } } })
    const wrapper = mountCard()
    await wrapper.find('[data-test="guardian-proposal-confirm"]').trigger('click')
    await flushPromises()
    expect(wrapper.emitted('error')).toBeTruthy()
    expect(wrapper.emitted('error')[0][0].error).toContain('expired')
  })
})
