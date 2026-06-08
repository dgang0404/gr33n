/**
 * Phase 62 WS5 / OC-62 — Guardian grow advisor closure.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync, existsSync } from 'node:fs'
import { join } from 'node:path'
import {
  buildZoneGrowStripStarters,
  buildHarvestFlowStarters,
} from '../lib/guardianStarters.js'

const repoDocs = join(process.cwd(), '..', 'docs')
const repoRoot = join(process.cwd(), '..')

describe('Phase 62 WS5 / OC-62 — grow advisor closure', () => {
  const zone = { id: 4, name: 'Flower Zone' }
  const activeCycle = {
    id: 22,
    zone_id: 4,
    name: 'OG Spring',
    strain_or_variety: 'OG Kush',
    current_stage: 'early_veg',
    is_active: true,
    started_at: '2026-01-01',
  }

  it('documents grow_advisor read tool and architecture section', () => {
    const arch = readFileSync(join(repoDocs, 'farm-guardian-architecture.md'), 'utf8')
    const plan = readFileSync(join(repoDocs, 'plans/phase_62_guardian_grow_advisor.plan.md'), 'utf8')
    expect(existsSync(join(repoRoot, 'internal/farmguardian/readtools_grow.go'))).toBe(true)
    expect(arch).toContain('Phase 62')
    expect(arch).toContain('grow_advisor')
    expect(plan).toContain('**Shipped.**')
  })

  it('zone grow strip offers VPD starter when active cycle present', () => {
    const starters = buildZoneGrowStripStarters({
      zone,
      activeCycle,
      farmId: 7,
      dayCount: 21,
    })
    const vpd = starters.find((s) => s.id === 'vpd-on-target')
    expect(vpd).toBeTruthy()
    expect(vpd.label).toBe('Is my VPD on target?')
    expect(vpd.message).toContain('grow_advisor')
    expect(vpd.message).toContain('Flower Zone')
    const flip = starters.find((s) => s.id === 'days-to-flip')
    expect(flip).toBeTruthy()
    const summarize = starters.find((s) => s.id === 'summarize-grow')
    expect(summarize).toBeTruthy()
  })

  it('late flower stage surfaces harvest readiness starter', () => {
    const starters = buildZoneGrowStripStarters({
      zone,
      activeCycle: { ...activeCycle, current_stage: 'late_flower' },
      farmId: 7,
    })
    expect(starters.find((s) => s.id === 'ready-to-harvest')).toBeTruthy()
    expect(starters.find((s) => s.id === 'days-to-flip')).toBeFalsy()
  })

  it('post-harvest starter asks what to change next run', () => {
    const starters = buildHarvestFlowStarters({
      zone,
      activeCycle,
      priorHarvestedCycle: { id: 18, name: 'OG Winter' },
      farmId: 7,
      surface: 'post_harvest',
    })
    const nextRun = starters.find((s) => s.id === 'next-run-diff')
    expect(nextRun).toBeTruthy()
    expect(nextRun.message).toContain('differently next run')
    expect(nextRun.message).toContain('grow_advisor')
  })

  it('grow strip wires dayCount into starters', () => {
    const strip = readFileSync(join(process.cwd(), 'src/components/ZoneCurrentGrowStrip.vue'), 'utf8')
    expect(strip).toContain('dayCount: dayCount.value')
  })
})
