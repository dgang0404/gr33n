/**
 * Phase 44 WS6 — wizard step navigation (Vitest).
 */
import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import FarmSetupWizard from '../views/FarmSetupWizard.vue'
import { useFarmContextStore } from '../stores/farmContext.js'
import { useFarmStore } from '../stores/farm.js'
import { FARM_SETUP_BLANK_ID } from '../lib/farmSetupWizard.js'

vi.mock('vue-router', () => ({
  useRoute: () => ({ params: { id: '7' } }),
}))

describe('Phase 44 WS6 — wizard navigation', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    const farmContext = useFarmContextStore()
    farmContext.farms = [{ id: 7, name: 'Test Farm' }]
    farmContext.farmId = 7
    const farmStore = useFarmStore()
    farmStore.farm = { id: 7, name: 'Test Farm' }
    farmStore.zones = []
    farmStore.fetchFarms = vi.fn()
    farmContext.fetchFarms = vi.fn().mockResolvedValue(undefined)
    farmContext.selectFarm = vi.fn().mockResolvedValue(undefined)
    farmContext.applyBootstrapTemplate = vi.fn()
  })

  it('farm setup wizard navigates choose → preview', async () => {
    const wrapper = mount(FarmSetupWizard, {
      global: {
        stubs: {
          GuardianStarterChips: true,
          RouterLink: { template: '<a><slot /></a>' },
        },
      },
    })
    await wrapper.vm.$nextTick()
    expect(wrapper.find('[data-test="farm-setup-wizard"]').exists()).toBe(true)
    expect(wrapper.text()).toContain('Start blank')
    await wrapper.find('[data-test="setup-continue-preview"]').trigger('click')
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toMatch(/blank farm|What this template creates/i)
    expect(wrapper.find('[data-test="setup-finish-blank"]').exists()).toBe(true)
  })

  it('farm setup preview shows blank finish when blank card selected', async () => {
    const wrapper = mount(FarmSetupWizard, {
      global: {
        stubs: {
          GuardianStarterChips: true,
          RouterLink: { template: '<a><slot /></a>' },
        },
      },
    })
    await wrapper.find(`[data-test="setup-card-${FARM_SETUP_BLANK_ID}"]`).trigger('click')
    await wrapper.find('[data-test="setup-continue-preview"]').trigger('click')
    await wrapper.vm.$nextTick()
    expect(wrapper.text()).toMatch(/blank farm/i)
    expect(wrapper.find('[data-test="setup-finish-blank"]').exists()).toBe(true)
  })
})
