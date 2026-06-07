import { describe, it, expect } from 'vitest'
import { formatWiringLabel, resolveWiring, wiringIsEmpty } from '../lib/hardwareWiring.js'

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
})
