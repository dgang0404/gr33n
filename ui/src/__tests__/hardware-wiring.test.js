import { describe, it, expect } from 'vitest'
import { formatWiringLabel, findWiringConflict, resolveWiring, wiringIsEmpty, formatEntityHardwareLabel } from '../lib/hardwareWiring.js'

describe('hardwareWiring', () => {
  it('formats BCM GPIO label', () => {
    expect(formatWiringLabel({ source: 'dht22', gpio_pin: 4 })).toBe('DHT22 · BCM GPIO 4')
  })

  it('resolves top-level wiring first', () => {
    const entity = { wiring: { source: 'ads1115', i2c_channel: 1 }, config: { wiring: { source: 'dht22' } } }
    expect(resolveWiring(entity).i2c_channel).toBe(1)
  })

  it('falls back to config.wiring', () => {
    const entity = { config: { wiring: { source: 'gpio_relay', gpio_pin: 17 } } }
    expect(resolveWiring(entity).gpio_pin).toBe(17)
  })

  it('detects empty wiring', () => {
    expect(wiringIsEmpty(null)).toBe(true)
    expect(wiringIsEmpty({})).toBe(true)
    expect(wiringIsEmpty({ source: 'dht22', gpio_pin: 4 })).toBe(false)
  })

  it('formats entity hardware label from wiring or relay channel', () => {
    expect(formatEntityHardwareLabel({ wiring: { source: 'dht22', gpio_pin: 4 } })).toContain('BCM GPIO 4')
    expect(formatEntityHardwareLabel({ hardware_identifier: '3' })).toBe('Relay ch 3')
  })

  it('allows shared dht22 gpio and blocks actuator overlap', () => {
    const sensors = [{ id: 3, name: 'Air Temp', wiring: { source: 'dht22', gpio_pin: 4, device_id: 1 } }]
    expect(findWiringConflict({
      wiring: { source: 'dht22', gpio_pin: 4, device_id: 1 },
      entityType: 'sensor',
      entityId: 5,
      sensors,
    })).toBeNull()

    const actuators = [{ id: 2, name: 'Light', wiring: { gpio_pin: 17, device_id: 1 } }]
    const hit = findWiringConflict({
      wiring: { source: 'gpio_relay', gpio_pin: 17, device_id: 1 },
      entityType: 'sensor',
      entityId: 9,
      actuators,
    })
    expect(hit?.entity_id).toBe(2)
  })
})
