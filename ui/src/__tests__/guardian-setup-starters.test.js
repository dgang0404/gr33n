import { describe, expect, it } from 'vitest'
import { buildSetupStarters } from '../lib/guardianStarters.js'

describe('Phase 44 WS4 — setup Guardian starters', () => {
  it('zero-zone farm offers first grow room chip (max 4 in drawer)', () => {
    const starters = buildSetupStarters({
      surface: 'setup_mode_chat',
      farmId: 7,
      zoneCount: 0,
      zones: [],
    })
    expect(starters.length).toBeLessThanOrEqual(4)
    expect(starters[0].id).toBe('first-grow-room')
    expect(starters[0].setupMode).toBe(true)
    expect(starters[0].message).toContain('setting up a new farm')
  })

  it('farm setup wizard returns compare-templates chip (max 2)', () => {
    const starters = buildSetupStarters({
      surface: 'farm_setup_wizard',
      farmId: 3,
      zoneCount: 0,
    })
    expect(starters.length).toBeLessThanOrEqual(2)
    expect(starters.some((s) => s.id === 'compare-templates')).toBe(true)
  })

  it('device wizard includes wire Pi procedure starter', () => {
    const starters = buildSetupStarters({
      surface: 'device_wizard',
      farmId: 5,
      zoneCount: 1,
      zones: [{ id: 10, name: 'Veg' }],
      deviceWizardStep: true,
    })
    const wire = starters.find((s) => s.id === 'wire-pi')
    expect(wire).toBeTruthy()
    expect(wire.message).toContain('wire-pi-relay-light')
    expect(wire.contextRef.path).toBe('/farms/5/devices/new')
  })

  it('zone wizard can suggest grow setup pack phrase for new room', () => {
    const starters = buildSetupStarters({
      surface: 'zone_wizard',
      farmId: 2,
      zoneCount: 1,
      zones: [{ id: 11, name: 'Flower Room' }],
      zoneName: 'Flower Room',
      activeCycles: [],
    })
    const grow = starters.find((s) => s.id === 'start-grow')
    expect(grow).toBeTruthy()
    expect(grow.message).toContain('philodendron')
    expect(grow.message).toContain('Flower Room')
  })
})
