import { describe, it, expect } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import FarmCanvasZoneTile from '../components/FarmCanvasZoneTile.vue'
import {
  constrainLayout,
  layoutToStyle,
  nudgeLayout,
  sunDialProgress,
  formatSunTimes,
} from '../lib/farmCanvasLayout.js'

describe('Phase 166 WS2 — farm canvas layout', () => {
  it('nudges and constrains normalized layout', () => {
    const moved = nudgeLayout({ x: 0.5, y: 0.5, w: 0.2, h: 0.18 }, 'right')
    expect(moved.x).toBeGreaterThan(0.5)
    const clamped = constrainLayout({ x: 0.95, y: 0.9, w: 0.2, h: 0.18 })
    expect(clamped.x + clamped.w).toBeLessThanOrEqual(1.001)
  })

  it('maps layout to percentage CSS', () => {
    const style = layoutToStyle({ x: 0.1, y: 0.2, w: 0.22, h: 0.18 }, { width: 800, height: 500 })
    expect(style.left).toBe('10%')
    expect(style.width).toBe('22%')
  })

  it('computes sun dial progress between sunrise and sunset', () => {
    const solar = {
      sunrise_at: '2026-07-12T10:00:00Z',
      sunset_at: '2026-07-12T22:00:00Z',
      daylength_hours: 12,
    }
    const mid = sunDialProgress(solar, new Date('2026-07-12T16:00:00Z'))
    expect(mid).toBeGreaterThan(0.4)
    expect(mid).toBeLessThan(0.6)
    const times = formatSunTimes(solar)
    expect(times.sunrise).toBeTruthy()
    expect(times.daylength).toBe('12 h')
  })
})

describe('Phase 166 WS3 — FarmCanvasZoneTile', () => {
  it('renders plants, water, and sensor rows', () => {
    setActivePinia(createPinia())
    const wrapper = mount(FarmCanvasZoneTile, {
      props: {
        zone: { id: 5, name: 'Herb & Greens Room', zone_type: 'indoor' },
        status: {
          health: 'ok',
          plants: { state: 'growing', cropName: 'Basil', stage: 'Veg stage' },
          light: { state: 'none' },
          water: { kind: 'gravity_drip', label: 'Gravity drip · plain water', nextRun: '7:00 AM daily' },
          sensors: { state: 'not_set_up', summary: 'Not set up yet' },
          attention: [],
        },
      },
      global: { stubs: { HelpTip: true } },
    })
    expect(wrapper.find('[data-test="farm-tile-water"]').text()).toContain('Gravity drip')
    expect(wrapper.find('[data-test="farm-tile-sensors"]').text()).toContain('Not set up yet')
  })
})
