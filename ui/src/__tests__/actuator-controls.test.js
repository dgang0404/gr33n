import { describe, expect, it } from 'vitest'
import {
  actuatorsInZone,
  animalHardwareChips,
  isDispenseActuator,
  isOpenCloseActuator,
} from '../lib/actuatorControls.js'

describe('actuatorControls', () => {
  it('classifies gate and feeder types for manual controls', () => {
    expect(isOpenCloseActuator('gate')).toBe(true)
    expect(isOpenCloseActuator('grow_light')).toBe(false)
    expect(isDispenseActuator('feeder_hopper')).toBe(true)
  })

  it('summarizes animal zone hardware for /animals cards', () => {
    const chips = animalHardwareChips([
      { zone_id: 55, actuator_type: 'feeder_hopper', current_state_text: 'online' },
      { zone_id: 55, actuator_type: 'gate', current_state_text: 'open' },
      { zone_id: 56, actuator_type: 'pump', current_state_text: 'online' },
    ], 55)
    expect(chips.map((c) => c.label)).toEqual(['Feeder', 'Gate'])
    expect(chips[1].state).toBe('open')
  })

  it('filters actuators by zone id', () => {
    const rows = actuatorsInZone([
      { id: 1, zone_id: 55, actuator_type: 'gate' },
      { id: 2, zone_id: 56, actuator_type: 'gate' },
    ], 55)
    expect(rows).toHaveLength(1)
    expect(rows[0].id).toBe(1)
  })
})
