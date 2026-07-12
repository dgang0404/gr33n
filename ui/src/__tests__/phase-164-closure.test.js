/**
 * Phase 164 — Demo seed: living farm, no cannabis.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

const repoRoot = join(process.cwd(), '..')

describe('Phase 164 WS1 — decannabis demo seed', () => {
  const seed = readFileSync(join(repoRoot, 'db/seeds/master_seed.sql'), 'utf8')

  it('rethemes farm-1 plants to chrysanthemum', () => {
    expect(seed).toContain("'Chrysanthemum', 'Mixed spray varieties', 'chrysanthemum'")
    expect(seed).toContain("DELETE FROM gr33ncrops.plants\nWHERE farm_id = 1 AND crop_key = 'cannabis'")
    expect(seed).not.toMatch(
      /INSERT INTO gr33ncrops\.plants \(farm_id, display_name, variety_or_cultivar, crop_key\)\s*VALUES[\s\S]*?'cannabis'[\s\S]*?ON CONFLICT DO NOTHING/,
    )
  })

  it('uses chrysanthemum batch labels instead of cannabis strains', () => {
    expect(seed).toContain("'Anastasia Green'")
    expect(seed).toContain("'Zembla White'")
    expect(seed).toContain("'Bloom run (12/12)'")
    expect(seed).toContain('Chrysanthemum — Cutting Batch 12')
    // Fresh INSERT rows — legacy names only appear in idempotent migration UPDATEs.
    expect(seed).not.toMatch(/INSERT INTO gr33nfertigation\.crop_cycles[\s\S]*'Gorilla Glue/)
    expect(seed).not.toMatch(/INSERT INTO gr33nfertigation\.crop_cycles[\s\S]*'Blue Dream'/)
    expect(seed).not.toMatch(/\) AS v\(zone_name[\s\S]*'OG Kush'/)
  })

  it('updates grower-facing copy for bloom stage', () => {
    expect(seed).toContain('bloom openness and stem length')
    expect(seed).toContain('bloom stage')
    expect(seed).not.toContain('Check trichomes')
  })
})

describe('Phase 164 WS2+WS3 — seeded sensor readings', () => {
  const seed = readFileSync(join(repoRoot, 'db/seeds/master_seed.sql'), 'utf8')

  it('inserts phase164_demo sensor_readings for wired sensors', () => {
    expect(seed).toContain('INSERT INTO gr33ncore.sensor_readings')
    expect(seed).toContain('"seed":"phase164_demo"')
    expect(seed).toContain("'Air Humidity Indoor',     72.4")
    expect(seed).toContain("'PAR Sensor Indoor',      620.0")
  })

  it('documents intentionally unwired bed sensors', () => {
    expect(seed).toContain('not set up')
    expect(seed).toContain('Propagation Dome Temp')
    expect(seed).toContain('Herb Room Air Temp')
    expect(seed).not.toMatch(
      /INSERT INTO gr33ncore\.sensor_readings[\s\S]*'Propagation Dome Temp'/,
    )
  })
})

describe('Phase 164 WS4 — gravity drip demo zone', () => {
  const seed = readFileSync(join(repoRoot, 'db/seeds/master_seed.sql'), 'utf8')

  it('seeds herb room gravity drip hardware and program', () => {
    expect(seed).toContain('demo-herb-relay-01')
    expect(seed).toContain('Herb Room Gravity Drip Valve')
    expect(seed).toContain("'drip'")
    expect(seed).toContain('Herb Room Gravity Header')
    expect(seed).toContain('Herb Room Gravity Drip')
    expect(seed).toContain('irrigation_only')
    expect(seed).toContain('Water Herbs Gravity Drip Daily')
    expect(seed).toContain('[seed:herb-gravity-drip-demo]')
  })
})

describe('Phase 164 WS5 — smoke test audit', () => {
  it('ships farm-1 seed assertion smokes', () => {
    const smoke = readFileSync(join(repoRoot, 'cmd/api/smoke_phase164_test.go'), 'utf8')
    expect(smoke).toContain('TestPhase164_Farm1NoCannabisPlantRow')
    expect(smoke).toContain('TestPhase164_Farm1ChrysanthemumDemoCycles')
    expect(smoke).toContain('TestPhase164_Farm1WiredSensorsHaveReadings')
    expect(smoke).toContain('TestPhase164_Farm1GravityDripProgram')
  })
})

describe('Phase 164 WS6 — closure', () => {
  const seed = readFileSync(join(repoRoot, 'db/seeds/master_seed.sql'), 'utf8')
  const currentState = readFileSync(join(repoRoot, 'docs/current-state.md'), 'utf8')

  it('ships phase164 verify queries in master seed', () => {
    expect(seed).toContain('Phase 164 WS6 — VERIFY')
    expect(seed).toContain('phase164_cannabis_plants_farm1')
    expect(seed).toContain('phase164_demo_sensor_readings')
    expect(seed).toContain('phase164_gravity_drip_program')
  })

  it('documents living demo farm in current-state', () => {
    expect(currentState).toContain('Demo farm seed')
    expect(currentState).toContain('Anastasia Green')
    expect(currentState).toContain('gravity-fed drip')
    expect(currentState).toContain('Phase164')
  })
})
