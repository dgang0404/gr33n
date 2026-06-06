import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'

vi.mock('../api', () => ({
  default: {
    get: vi.fn((url) => {
      if (url === '/capabilities') return Promise.resolve({ data: { ai_enabled: true } })
      return Promise.reject(new Error(`unexpected ${url}`))
    }),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

import GuardianActionProposal from '../components/GuardianActionProposal.vue'
import GuardianStarterChips from '../components/GuardianStarterChips.vue'
import { useCapabilitiesStore } from '../stores/capabilities'
import {
  guardianProposalAriaLabel,
  runFeedNowAriaLabel,
  runPulseAriaLabel,
} from '../lib/farmerA11y.js'

const router = createRouter({
  history: createMemoryHistory(),
  routes: [{ path: '/', component: { template: '<div />' } }],
})

describe('Phase 45 WS6 — farmer a11y helpers', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('builds descriptive aria labels for Guardian actions', () => {
    expect(guardianProposalAriaLabel('confirm', 'Ack humidity')).toContain('Confirm proposed action')
    expect(guardianProposalAriaLabel('dismiss', 'Setup pack')).toContain('without changing farm data')
    expect(guardianProposalAriaLabel('refine', 'Patch feed')).toContain('Refine')
  })

  it('builds run-now labels for zone feeding', () => {
    expect(runFeedNowAriaLabel('Flower Room', 'Daily feed')).toContain('Flower Room')
    expect(runFeedNowAriaLabel('Flower Room', 'Daily feed')).toContain('Daily feed')
    expect(runPulseAriaLabel('Irrigation pump', 30)).toContain('30 seconds')
  })

  it('Guardian proposal buttons expose aria-label and touch-friendly classes', () => {
    const wrapper = mount(GuardianActionProposal, {
      props: {
        proposal: {
          proposal_id: 'p1',
          tool: 'ack_alert',
          args: { alert_id: 2 },
          summary: 'Acknowledge: Humidity high',
          status: 'pending',
        },
        canOperate: true,
      },
      global: { plugins: [router] },
    })
    const confirm = wrapper.find('[data-test="guardian-proposal-confirm"]')
    const dismiss = wrapper.find('[data-test="guardian-proposal-dismiss"]')
    expect(confirm.attributes('aria-label')).toContain('Confirm proposed action')
    expect(dismiss.attributes('aria-label')).toContain('without changing farm data')
    expect(confirm.classes().join(' ')).toMatch(/min-h-\[44px\]/)
    expect(dismiss.classes().join(' ')).toMatch(/border-zinc-600/)
  })

  it('Guardian starter chips have group label and per-chip aria-label', async () => {
    await useCapabilitiesStore().fetch()
    const wrapper = mount(GuardianStarterChips, {
      props: {
        starters: [{ id: 'next-feed', label: 'Next feed', message: 'When is the next feed?' }],
      },
      global: { plugins: [router] },
    })
    await flushPromises()
    expect(wrapper.attributes('aria-label')).toBe('Guardian conversation starters')
    const chip = wrapper.find('[data-test="guardian-starter-next-feed"]')
    expect(chip.attributes('aria-label')).toBe('Ask Guardian: Next feed')
    expect(chip.classes().join(' ')).toMatch(/min-h-\[44px\]/)
  })
})
