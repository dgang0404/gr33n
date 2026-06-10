/**
 * Phase 45 WS3 — Vocabulary v2 (zones not rooms) closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import { buildNavGroups, mobileBottomNav } from '../lib/navGroups.js'
import { GROW_PATH_ZONE_LABELS } from '../lib/farmerVocabulary.js'
import { buildSetupStarters } from '../lib/guardianStarters.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const uiSrc = join(process.cwd(), 'src')

describe('Phase 45 WS3 — vocabulary v2 closure', () => {
  it('farmer-vocabulary.md marks v2 shipped', () => {
    const vocab = readFileSync(join(repoDocs, 'farmer-vocabulary.md'), 'utf8')
    expect(vocab).toContain('Vocabulary v2 (zones not rooms)** — **shipped**')
    expect(vocab).toContain('phase-45-ws3-closure.test.js')
  })

  it('nav uses My zones and mobile Zones', () => {
    const groups = buildNavGroups()
    const grow = groups.find((g) => g.label === 'Grow & operate')
    const zonesItem = grow.items.find((i) => i.to === '/zones')
    expect(zonesItem.label).toBe(GROW_PATH_ZONE_LABELS.navMyZones)
    expect(zonesItem.navTitle).toMatch(/room|zone/i)
    expect(mobileBottomNav.find((i) => i.to === '/zones')?.label).toBe('Zones')
  })

  it('Guardian setup starters use zone fallback not this room', () => {
    const starters = buildSetupStarters({
      surface: 'empty_zone_grow',
      farmId: 1,
      zoneCount: 1,
      zones: [{ id: 2, name: 'Bench A' }],
      zoneName: '',
      activeCycles: [],
    })
    const grow = starters.find((s) => s.id === 'start-grow')
    expect(grow.label).toContain('Bench A')
    const js = readFileSync(join(uiSrc, 'lib/guardianStarters.js'), 'utf8')
    expect(js).not.toMatch(/\|\|\s*'this room'/)
  })

  it('operator-tour uses zone product language in §7b and nav', () => {
    const tour = readFileSync(join(repoDocs, 'operator-tour.md'), 'utf8')
    expect(tour).toContain('## 7b. Feeding & water for this zone (Phase 47)')
    expect(tour).toContain('**My zones**')
    expect(tour).toContain('one card per zone')
    expect(tour).not.toContain('## 7b. Feeding & water for this room')
  })

  it('phase 45 parent plan marks WS3 complete', () => {
    const plan = readFileSync(
      join(repoDocs, 'plans/phase_45_farmer_validation_whole_app_polish.plan.md'),
      'utf8',
    )
    expect(plan).toMatch(/ws3-copy-pass-v2[\s\S]*status: completed/)
  })

  it('closure test and extended grow-path scan exist', () => {
    expect(existsSync(join(uiSrc, '__tests__/phase-45-ws3-closure.test.js'))).toBe(true)
    const growTest = readFileSync(join(uiSrc, '__tests__/farmer-vocabulary-grow-path.test.js'), 'utf8')
    const vocabLib = readFileSync(join(uiSrc, 'lib/farmerVocabulary.js'), 'utf8')
    expect(growTest).toContain('guardianStarters.js')
    expect(vocabLib).toContain('GROW_PATH_GENERIC_ROOM_BANS')
  })
})
