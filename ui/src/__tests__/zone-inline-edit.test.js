/**
 * Phase 69 WS6 — zone inline edit closure guards.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const uiSrc = join(process.cwd(), 'src')

describe('Phase 69 WS1/WS2 — zone inline edit', () => {
  it('ZoneNeedSection embeds wiring panels and zone lighting editor', () => {
    const src = readFileSync(join(uiSrc, 'components/ZoneNeedSection.vue'), 'utf8')
    expect(src).toContain('HardwareWiringPanel')
    expect(src).toContain('ActuatorWiringPanel')
    expect(src).toContain('ZoneLightingEditor')
    expect(src).toContain('zone-sensor-wiring-toggle')
    expect(src).toContain('zone-actuator-wiring-toggle')
    expect(src).not.toContain("path: '/lighting'")
  })

  it('ZoneAutomationPanel toggles schedules inline', () => {
    const src = readFileSync(join(uiSrc, 'components/ZoneAutomationPanel.vue'), 'utf8')
    expect(src).toContain('zone-schedule-toggle')
    expect(src).toContain('updateScheduleActive')
  })

  it('ZoneDetail overview spine and tab deep links', () => {
    const src = readFileSync(join(uiSrc, 'views/ZoneDetail.vue'), 'utf8')
    expect(src).toContain('zone-overview-spine')
    expect(src).toContain('route.query.tab')
    expect(src).toContain('refreshHardware')
  })
})
