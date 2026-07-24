import { describe, expect, it, beforeEach, vi } from 'vitest'
import { readFileSync } from 'node:fs'
import { fileURLToPath } from 'node:url'
import { dirname, join } from 'node:path'
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

import api from '../api'
import { useFarmStore } from '../stores/farm'
import CropOpsTimeline from '../components/CropOpsTimeline.vue'
import {
  cropOpsEventHasFormula,
  cropOpsKindLabel,
  formulaSnapshotLines,
} from '../lib/cropOpsTimeline.js'

const root = join(dirname(fileURLToPath(import.meta.url)), '..')

describe('Phase 211.04 crop ops report UI closure', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.clearAllMocks()
  })

  it('farm store loads ops-timeline endpoint with optional range', async () => {
    api.get.mockResolvedValue({ data: { events: [], from: '2026-01-01', to: '2026-03-01' } })
    const farm = useFarmStore()
    await farm.loadCropCycleOpsTimeline(1, 42, { from: '2026-01-01', to: '2026-03-01' })
    expect(api.get).toHaveBeenCalledWith(
      '/farms/1/crop-cycles/42/ops-timeline?from=2026-01-01&to=2026-03-01',
    )
  })

  it('formulaSnapshotLines renders recipe revision snapshot', () => {
    const lines = formulaSnapshotLines({
      application_recipe_revision_id: 9,
      formula_snapshot: {
        recipe_name: 'JMS dilute',
        dilution_ratio: '1:500',
        components: [{ input_name: 'JMS', part_value: 10 }],
      },
    })
    expect(lines.some((l) => l.includes('JMS dilute'))).toBe(true)
    expect(lines.some((l) => l.includes('Revision #9'))).toBe(true)
    expect(cropOpsEventHasFormula({ formula_snapshot: { recipe_name: 'x' } })).toBe(true)
  })

  it('CropOpsTimeline renders mix row with formula-at-time block', async () => {
    api.get.mockResolvedValue({
      data: {
        events: [
          {
            kind: 'mix',
            id: 1,
            occurred_at: '2026-02-15T10:00:00Z',
            summary: 'Flower feed',
            details: {
              formula_snapshot: {
                recipe_name: 'Flower A',
                components: [{ input_name: 'Kelp', part_value: 5 }],
              },
              application_recipe_revision_id: 3,
            },
          },
        ],
        from: '2026-02-01',
        to: '2026-02-28',
      },
    })
    const wrapper = mount(CropOpsTimeline, {
      props: { farmId: 1, cycleId: 42 },
      global: { stubs: { HelpTip: { template: '<span><slot /></span>' } } },
    })
    await flushPromises()
    expect(wrapper.find('[data-test="crop-ops-list"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="crop-ops-row"][data-kind="mix"]').exists()).toBe(true)
    expect(wrapper.find('[data-test="crop-ops-formula"]').text()).toContain('Flower A')
    expect(wrapper.find('[data-test="crop-ops-formula"]').text()).toContain('Revision #3')
    expect(cropOpsKindLabel('program_run')).toBe('Program run')
  })

  it('Money Grows links to summary ops anchor', () => {
    const grows = readFileSync(join(root, 'components/MoneyGrowsSection.vue'), 'utf8')
    expect(grows).toContain('summary#crop-ops-timeline')
    expect(grows).toContain('data-test="money-grow-ops-log"')
  })

  it('CropCycleSummary embeds CropOpsTimeline section', () => {
    const summary = readFileSync(join(root, 'views/CropCycleSummary.vue'), 'utf8')
    expect(summary).toContain('CropOpsTimeline')
    expect(summary).toContain('crop-ops-timeline')
  })

  it('operator tour mentions grow ops log', () => {
    const tour = readFileSync(join(root, '../../docs/operator-tour.md'), 'utf8')
    expect(tour).toContain('Ops log')
  })
})
