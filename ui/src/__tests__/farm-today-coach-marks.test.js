/**
 * Phase 177 — farmTodayCoachMarks unit tests.
 */
import { describe, it, expect } from 'vitest'
import {
  buildTodayCoachSteps,
  isNarrowTodayViewport,
  TODAY_COACH_DONE_KEY,
} from '../lib/farmTodayCoachMarks.js'

describe('farmTodayCoachMarks', () => {
  it('exports session key', () => {
    expect(TODAY_COACH_DONE_KEY).toBe('gr33n_today_coach_done')
  })

  it('builds three desktop steps with attention fallback to pulse', () => {
    const withAttention = buildTodayCoachSteps({ hasAttention: true, narrowViewport: false })
    expect(withAttention).toHaveLength(3)
    expect(withAttention[2].id).toBe('attention')
    expect(withAttention[2].target).toBe('farm-today-attention')

    const withPulse = buildTodayCoachSteps({ hasAttention: false, narrowViewport: false })
    expect(withPulse[2].id).toBe('pulse')
    expect(withPulse[2].target).toBe('farm-site-strip')
  })

  it('narrows to tap-zone only on small viewports', () => {
    const steps = buildTodayCoachSteps({ hasAttention: true, narrowViewport: true })
    expect(steps).toHaveLength(1)
    expect(steps[0].id).toBe('tap_zone')
  })

  it('detects narrow viewport threshold', () => {
    expect(isNarrowTodayViewport(390)).toBe(true)
    expect(isNarrowTodayViewport(1280)).toBe(false)
  })

  it('never includes a Guardian step', () => {
    const steps = buildTodayCoachSteps({ hasAttention: true, narrowViewport: false })
    const text = JSON.stringify(steps)
    expect(text.toLowerCase()).not.toContain('guardian')
    expect(text.toLowerCase()).not.toContain('gr33n')
  })
})
