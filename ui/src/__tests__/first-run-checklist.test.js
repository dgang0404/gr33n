import { afterEach, describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import GettingStartedChecklist from '../components/GettingStartedChecklist.vue'
import {
  computeFirstRunChecklist,
  dismissFirstRunChecklist,
  isFirstRunChecklistDismissed,
  isFirstRunComplete,
  shouldShowFirstRunChecklist,
  clearFirstRunChecklistDismiss,
} from '../lib/firstRunChecklist.js'

describe('Phase 44 WS5 — first-run checklist logic', () => {
  afterEach(() => {
    clearFirstRunChecklistDismiss(42)
  })

  it('marks all items incomplete on a blank farm', () => {
    const items = computeFirstRunChecklist({ farmId: 7 })
    expect(items).toHaveLength(4)
    expect(items.every((i) => !i.done)).toBe(true)
    expect(items[0].to).toBe('/farms/7/zones/new')
    expect(items[1].to).toBe('/farms/7/devices/new')
    expect(items[2].to).toBe('/comfort-targets')
    expect(items[3].to).toBe('/comfort-targets?tab=schedules')
  })

  it('marks items done as farm data fills in', () => {
    const items = computeFirstRunChecklist({
      farmId: 3,
      zones: [{ id: 1, name: 'Veg' }],
      devices: [{ id: 10, status: 'offline' }],
      setpoints: [{ id: 1, zone_id: 1, sensor_type: 'humidity', min_value: 40, max_value: 60 }],
      schedules: [{ id: 2, is_active: true }],
    })
    expect(isFirstRunComplete(items)).toBe(true)
    expect(shouldShowFirstRunChecklist(3, items)).toBe(false)
  })

  it('hides when dismissed but reappears after clear', () => {
    const items = computeFirstRunChecklist({ farmId: 42 })
    expect(shouldShowFirstRunChecklist(42, items)).toBe(true)
    dismissFirstRunChecklist(42)
    expect(isFirstRunChecklistDismissed(42)).toBe(true)
    expect(shouldShowFirstRunChecklist(42, items)).toBe(false)
    clearFirstRunChecklistDismiss(42)
    expect(shouldShowFirstRunChecklist(42, items)).toBe(true)
  })
})

describe('Phase 44 WS5 — GettingStartedChecklist component', () => {
  it('renders checklist rows and emits dismiss', async () => {
    setActivePinia(createPinia())
    const items = computeFirstRunChecklist({ farmId: 5 })
    const wrapper = mount(GettingStartedChecklist, {
      props: { items, farmId: 5, starters: [] },
      global: {
        stubs: { GuardianStarterChips: true, RouterLink: { template: '<a><slot /></a>' } },
      },
    })
    expect(wrapper.find('[data-test="first-run-checklist"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="first-run-item-add_zone"]').text()).toContain('Add a zone')
    await wrapper.find('[data-test="first-run-dismiss"]').trigger('click')
    expect(wrapper.emitted('dismiss')).toHaveLength(1)
    expect(isFirstRunChecklistDismissed(5)).toBe(true)
    clearFirstRunChecklistDismiss(5)
  })
})
