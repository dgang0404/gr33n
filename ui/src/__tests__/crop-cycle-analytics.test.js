import { describe, it, expect, beforeEach, vi } from 'vitest'
import { mount, flushPromises } from '@vue/test-utils'
import { setActivePinia, createPinia } from 'pinia'

vi.mock('../api', () => ({
  default: {
    get: vi.fn(),
    post: vi.fn(),
    patch: vi.fn(),
    delete: vi.fn(),
    interceptors: { request: { use: vi.fn() }, response: { use: vi.fn() } },
  },
}))

// useRoute() is what CropCycleSummary uses; the $route global doesn't get
// picked up by the composition API. Mocking the whole module is cleaner
// than spinning up a real vue-router in every test.
const routeMock = { params: { id: '42' } }
vi.mock('vue-router', () => ({
  useRoute: () => routeMock,
  useRouter: () => ({ push: vi.fn() }),
}))

import api from '../api'
import { useFarmStore } from '../stores/farm'
import { useFarmContextStore } from '../stores/farmContext'
import CropCycleSummary from '../views/CropCycleSummary.vue'
import CropCycleCompare from '../views/CropCycleCompare.vue'

const sampleSummary = (overrides = {}) => ({
  cycle: {
    id: 42,
    farm_id: 1,
    zone_id: 7,
    name: 'OG Kush — Run 3',
    strain_or_variety: 'OG Kush',
    current_stage: { Gr33nfertigationGrowthStageEnum: 'late_flower', Valid: true },
    is_active: false,
    started_at: '2026-03-01',
    harvested_at: '2026-05-01',
  },
  duration_days: 61,
  fertigation: {
    event_count: 3,
    total_liters: 10,
    avg_ec_mscm: 1.5,
    min_ec_mscm: 1,
    max_ec_mscm: 2,
    avg_ph: 6.05,
  },
  cost: {
    totals: [{ currency: 'USD', total_income: 0, total_expenses: 50, net: -50 }],
    by_category: [{ category: 'fertilizers_soil_amendments', currency: 'USD', income: 0, expense: 50, net: -50, tx_count: 1 }],
  },
  yield: {
    grams: 200,
    grams_per_liter: 20,
    grams_per_day: 3.28,
    cost_per_gram: 0.25,
  },
  stages: [{ stage: 'late_flower', entered_at: '2026-03-01' }],
  stage_history_supported: false,
  ...overrides,
})

describe('farm store — crop cycle analytics (Phase 28 WS2)', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('loadCropCycleSummary GETs /crop-cycles/:id/summary and returns the body', async () => {
    api.get.mockResolvedValue({ data: sampleSummary() })
    const farm = useFarmStore()
    const out = await farm.loadCropCycleSummary(42)
    expect(api.get).toHaveBeenCalledWith('/crop-cycles/42/summary')
    expect(out.cycle.id).toBe(42)
    expect(out.yield.cost_per_gram).toBe(0.25)
  })

  it('loadCropCycleCompare joins ids and passes them as ?ids=', async () => {
    api.get.mockResolvedValue({ data: { cycles: [sampleSummary({ cycle: { ...sampleSummary().cycle, id: 1 } })] } })
    const farm = useFarmStore()
    const out = await farm.loadCropCycleCompare(7, [1, 2, 3])
    expect(api.get).toHaveBeenCalledWith('/farms/7/crop-cycles/compare?ids=1,2,3')
    expect(out.cycles).toHaveLength(1)
  })

  it('loadCropCycleCompare returns {cycles: []} early when no ids supplied (no HTTP call)', async () => {
    const farm = useFarmStore()
    const out = await farm.loadCropCycleCompare(7, [])
    expect(out).toEqual({ cycles: [] })
    expect(api.get).not.toHaveBeenCalled()
  })
})

describe('CropCycleSummary.vue', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    api.get.mockResolvedValue({ data: sampleSummary() })
  })

  it('renders the header strip with cycle name + duration + stage', async () => {
    routeMock.params = { id: '42' }
    const wrapper = mount(CropCycleSummary, {
      global: {
        stubs: {
          'router-link': { template: '<a><slot /></a>' },
          HelpTip: { template: '<span><slot /></span>' },
        },
      },
    })
    await flushPromises()
    const header = wrapper.find('[data-test="summary-header"]')
    expect(header.exists()).toBe(true)
    expect(header.text()).toContain('OG Kush — Run 3')
    expect(header.text()).toContain('61')         // duration
    expect(header.text()).toContain('late_flower') // stage
  })

  it('renders all four metric cards', async () => {
    routeMock.params = { id: '42' }
    const wrapper = mount(CropCycleSummary, {
      global: {
        stubs: {
          'router-link': { template: '<a><slot /></a>' },
          HelpTip: { template: '<span><slot /></span>' },
        },
      },
    })
    await flushPromises()
    expect(wrapper.find('[data-test="card-fertigation"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="card-cost"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="card-yield"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="card-stages"]').exists()).toBe(true)
    const yieldCard = wrapper.find('[data-test="card-yield"]')
    expect(yieldCard.text()).toContain('200')   // grams
    expect(yieldCard.text()).toContain('20')    // g per liter
  })

  it('shows an error message when the store throws', async () => {
    api.get.mockRejectedValue({ response: { data: { error: 'crop cycle not found' } } })
    routeMock.params = { id: '999' }
    const wrapper = mount(CropCycleSummary, {
      global: {
        stubs: {
          'router-link': { template: '<a><slot /></a>' },
          HelpTip: { template: '<span><slot /></span>' },
        },
      },
    })
    await flushPromises()
    expect(wrapper.text()).toContain('crop cycle not found')
  })
})

describe('CropCycleCompare.vue', () => {
  let farmContext

  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
    farmContext = useFarmContextStore()
    farmContext.farmId = 1
    farmContext.farms = [{ id: 1, name: 'Test Farm' }]
    api.get.mockImplementation((url) => {
      if (url === '/farms/1/crop-cycles') {
        return Promise.resolve({
          data: [
            { id: 1, name: 'Cycle A' },
            { id: 2, name: 'Cycle B' },
            { id: 3, name: 'Cycle C' },
          ],
        })
      }
      if (url.startsWith('/farms/1/crop-cycles/compare?ids=')) {
        return Promise.resolve({
          data: {
            cycles: [
              sampleSummary({ cycle: { ...sampleSummary().cycle, id: 1, name: 'Cycle A' }, yield: { grams: 100, grams_per_liter: 10, grams_per_day: 1, cost_per_gram: 0.5 } }),
              sampleSummary({ cycle: { ...sampleSummary().cycle, id: 2, name: 'Cycle B' }, yield: { grams: 200, grams_per_liter: 20, grams_per_day: 3, cost_per_gram: 0.25 } }),
            ],
          },
        })
      }
      return Promise.resolve({ data: {} })
    })
  })

  it('shows the picker but no table when nothing is selected', async () => {
    const wrapper = mount(CropCycleCompare, {
      global: {
        stubs: {
          'router-link': { template: '<a><slot /></a>' },
          HelpTip: { template: '<span><slot /></span>' },
        },
      },
    })
    await flushPromises()
    expect(wrapper.find('[data-test="picker"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="compare-table"]').exists()).toBe(false)
    expect(wrapper.text()).toContain('Pick two or more crop cycles')
  })

  it('renders the comparison table once cycles are selected', async () => {
    const wrapper = mount(CropCycleCompare, {
      global: {
        stubs: {
          'router-link': { template: '<a><slot /></a>' },
          HelpTip: { template: '<span><slot /></span>' },
        },
      },
    })
    await flushPromises()
    const checkboxes = wrapper.findAll('input[type="checkbox"]')
    await checkboxes[0].trigger('change')
    await checkboxes[1].trigger('change')
    await flushPromises()
    expect(api.get).toHaveBeenCalledWith('/farms/1/crop-cycles/compare?ids=1,2')
    expect(wrapper.find('[data-test="compare-table"]').exists()).toBe(true)
    const tableText = wrapper.find('[data-test="compare-table"]').text()
    expect(tableText).toContain('Cycle A')
    expect(tableText).toContain('Cycle B')
    expect(tableText).toContain('Yield')
  })

  it('disables additional checkboxes once 5 cycles are picked', async () => {
    // Replace the cycles list with 6 items so we can verify the cap.
    api.get.mockImplementation((url) => {
      if (url === '/farms/1/crop-cycles') {
        return Promise.resolve({
          data: Array.from({ length: 6 }, (_, i) => ({ id: i + 1, name: `Cycle ${i + 1}` })),
        })
      }
      return Promise.resolve({ data: { cycles: [] } })
    })
    const wrapper = mount(CropCycleCompare, {
      global: {
        stubs: {
          'router-link': { template: '<a><slot /></a>' },
          HelpTip: { template: '<span><slot /></span>' },
        },
      },
    })
    await flushPromises()
    const boxes = wrapper.findAll('input[type="checkbox"]')
    // Pick the first 5 — the 6th must end up disabled.
    for (let i = 0; i < 5; i++) {
      await boxes[i].trigger('change')
    }
    await flushPromises()
    expect(boxes[5].attributes('disabled')).toBeDefined()
  })

  it('shows the "Select a farm" hint when farmId is unset', async () => {
    farmContext.farmId = null
    const wrapper = mount(CropCycleCompare, {
      global: {
        stubs: {
          'router-link': { template: '<a><slot /></a>' },
          HelpTip: { template: '<span><slot /></span>' },
        },
      },
    })
    await flushPromises()
    expect(wrapper.text()).toContain('Select a farm')
    expect(wrapper.find('[data-test="picker"]').exists()).toBe(false)
  })
})
