/**
 * Phase 69 WS3 — fleet tab grouping tests.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import { groupEntitiesByZone, actuatorPinConflict, listFleetPinConflicts } from '../lib/fleetGrouping.js'

const uiSrc = join(process.cwd(), 'src')

describe('Phase 69 WS3 — fleet tab', () => {
  it('groups entities by zone', () => {
    const groups = groupEntitiesByZone(
      [
        { id: 1, zone_id: 2, name: 'A' },
        { id: 2, zone_id: 2, name: 'B' },
        { id: 3, zone_id: null, name: 'C' },
      ],
      [{ id: 2, name: 'Flower' }],
    )
    expect(groups).toHaveLength(2)
    expect(groups[0].zoneName).toBe('Flower')
    expect(groups[0].items).toHaveLength(2)
    expect(groups[1].zoneName).toBe('Unassigned')
  })

  it('flags actuator pin conflicts', () => {
    const sensors = []
    const actuators = [
      { id: 1, name: 'Pump A', wiring: { device_id: 1, gpio_pin: 17, source: 'gpio_relay' } },
      { id: 2, name: 'Pump B', wiring: { device_id: 1, gpio_pin: 17, source: 'gpio_relay' } },
    ]
    expect(actuatorPinConflict(actuators[1], sensors, actuators)?.entity_id).toBe(1)
    expect(listFleetPinConflicts(sensors, actuators).length).toBeGreaterThanOrEqual(1)
  })

  it('ZonesWorkspace fleet views group by zone', () => {
    const ws = readFileSync(join(uiSrc, 'views/workspaces/ZonesWorkspace.vue'), 'utf8')
    expect(ws).toContain('group-by-zone')
    expect(ws).toContain('embedded')
    const sensors = readFileSync(join(uiSrc, 'views/Sensors.vue'), 'utf8')
    expect(sensors).toContain('fleet-zone-group')
    expect(sensors).toContain('groupByZone')
  })
})
