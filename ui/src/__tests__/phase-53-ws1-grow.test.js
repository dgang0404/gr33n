/**
 * Phase 53 WS1 — grow closure (Vitest bundle guard).
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import {
  activeCycleForZone,
  buildCompareRoute,
  buildHarvestPayload,
  buildPostHarvestCompareRoute,
  buildStartGrowPayload,
  daysSinceStart,
  lastHarvestedCycleInZone,
  strainFromPlant,
} from '../lib/growHub.js'

const uiSrc = join(process.cwd(), 'src')

describe('Phase 53 WS1 — grow closure', () => {
  it('picks active cycle and days since start', () => {
    const cycles = [
      { id: 1, zone_id: 3, is_active: false, started_at: '2026-01-01' },
      { id: 2, zone_id: 3, is_active: true, started_at: '2026-06-01', current_stage: 'early_flower' },
    ]
    expect(activeCycleForZone(cycles, 3)?.id).toBe(2)
    const days = daysSinceStart(cycles[1], new Date(2026, 5, 7))
    expect(days).toBe(6)
  })

  it('finds prior harvested cycle in same zone', () => {
    const cycles = [
      { id: 10, zone_id: 2, is_active: false },
      { id: 11, zone_id: 2, is_active: false },
      { id: 12, zone_id: 2, is_active: true },
    ]
    expect(lastHarvestedCycleInZone(cycles, 2, 11)?.id).toBe(10)
  })

  it('builds start and harvest payloads', () => {
    const start = buildStartGrowPayload({
      zoneId: 5,
      strain: 'Basil',
      name: 'Basil — Tent A',
      programId: 9,
    })
    expect(start.zone_id).toBe(5)
    expect(start.is_active).toBe(true)
    expect(start.primary_program_id).toBe(9)

    const cycle = { id: 3, name: 'Run 3', zone_id: 5, strain_or_variety: 'OG' }
    const harvest = buildHarvestPayload(cycle, { yieldGrams: 128, yieldNotes: 'trimmed' })
    expect(harvest.is_active).toBe(false)
    expect(harvest.yield_grams).toBe(128)
    expect(harvest.yield_notes).toBe('trimmed')
  })

  it('builds compare routes with query ids', () => {
    expect(buildCompareRoute(4, [12, 8])).toEqual({
      path: '/farms/4/crop-cycles/compare',
      query: { ids: '12,8' },
    })
    const cycles = [
      { id: 20, zone_id: 1, is_active: false },
      { id: 19, zone_id: 1, is_active: false },
    ]
    expect(buildPostHarvestCompareRoute(4, cycles, 20, 1)).toEqual({
      path: '/farms/4/crop-cycles/compare',
      query: { ids: '20,19' },
    })
  })

  it('formats strain from plant record', () => {
    expect(strainFromPlant({ display_name: 'OG Kush', variety_or_cultivar: 'Pheno A' }))
      .toBe('OG Kush (Pheno A)')
  })

  it('ZoneCurrentGrowStrip and wizards are wired in zone + plants views', () => {
    const zone = readFileSync(join(uiSrc, 'views/ZoneDetail.vue'), 'utf8')
    const plants = readFileSync(join(uiSrc, 'views/Plants.vue'), 'utf8')
    expect(zone).toContain('ZoneCurrentGrowStrip')
    expect(zone).toContain('StartGrowWizard')
    expect(zone).toContain('HarvestWeighIn')
    expect(zone).toContain('PostHarvestScreen')
    expect(plants).toContain('StartGrowWizard')
    expect(plants).toContain('data-test="plant-start-grow"')
    expect(plants).toContain('EmptyStateHint')
  })

  it('zone water story exposes grow connection line', () => {
    const story = readFileSync(join(uiSrc, 'components/ZoneWaterGrowStory.vue'), 'utf8')
    expect(story).toContain('ZoneGrowConnectionLine')
    const line = readFileSync(join(uiSrc, 'components/ZoneGrowConnectionLine.vue'), 'utf8')
    expect(line).toContain('data-test="zone-grow-connection-line"')
  })

  it('phase-53 ws1 test file exists', () => {
    expect(existsSync(join(uiSrc, '__tests__/phase-53-ws1-grow.test.js'))).toBe(true)
  })
})
