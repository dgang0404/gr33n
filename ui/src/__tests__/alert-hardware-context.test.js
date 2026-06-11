import { describe, it, expect } from 'vitest'
import {
  alertHardwareContext,
  formatAlertHardwareLine,
} from '../lib/alertHardwareContext.js'

describe('alertHardwareContext', () => {
  it('includes BCM GPIO for sensor-triggered alerts', () => {
    const alert = {
      triggering_event_source_type: 'sensor',
      triggering_event_source_id: 10,
    }
    const sensors = [{
      id: 10,
      name: 'Soil moisture A',
      sensor_type: 'soil_moisture',
      wiring: { source: 'ads1115', gpio_pin: 18, i2c_channel: 0 },
    }]

    expect(formatAlertHardwareLine(alert, { sensors })).toBe(
      'Soil moisture A · ADS1115 · BCM GPIO 18 · I2C ch 0',
    )
  })

  it('includes relay channel for actuator-triggered alerts', () => {
    const alert = {
      triggering_event_source_type: 'actuator',
      triggering_event_source_id: 3,
    }
    const actuators = [{
      id: 3,
      name: 'Irrigation pump',
      actuator_type: 'pump',
      hardware_identifier: '2',
    }]

    const ctx = alertHardwareContext(alert, { actuators })
    expect(ctx?.hardwareLabel).toBe('Relay ch 2')
    expect(ctx?.line).toContain('Irrigation pump')
  })

  it('falls back when source entity is missing from zone list', () => {
    const line = formatAlertHardwareLine({
      triggering_event_source_type: 'sensor',
      triggering_event_source_id: 99,
    }, { sensors: [] })
    expect(line).toBe('Sensor #99')
  })
})
