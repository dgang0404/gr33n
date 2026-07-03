import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import {
  PI_HEADER_PINS,
  pinByBcm,
  pinByPhysical,
  assignmentsForDevice,
  devicesWithWiring,
  headerGridRows,
} from '../lib/piPinMap.js'
import { buildNavGroups } from '../lib/navGroups.js'
import router from '../router/index.js'

describe('Phase 119 — Virtual Pi board view', () => {
  it('pin map has 40 unique physical positions', () => {
    expect(PI_HEADER_PINS).toHaveLength(40)
    const physicals = PI_HEADER_PINS.map((p) => p.physical)
    expect(new Set(physicals).size).toBe(40)
  })

  it('BCM GPIO 4 maps to physical pin 7', () => {
    expect(pinByBcm(4)?.physical).toBe(7)
    expect(pinByPhysical(7)?.bcm).toBe(4)
  })

  it('I2C bus uses pins 3 and 5', () => {
    expect(pinByPhysical(3)?.buses).toContain('i2c1')
    expect(pinByPhysical(5)?.buses).toContain('i2c1')
  })

  it('no duplicate BCM numbers among GPIO pins', () => {
    const bcms = PI_HEADER_PINS.filter((p) => p.bcm != null).map((p) => p.bcm)
    expect(new Set(bcms).size).toBe(bcms.length)
  })

  it('header grid has 20 rows (2 columns)', () => {
    const rows = headerGridRows()
    expect(rows).toHaveLength(20)
    expect(rows[0].left?.physical).toBe(1)
    expect(rows[0].right?.physical).toBe(2)
  })

  it('assignmentsForDevice maps DHT22 on BCM 4 to physical 7', () => {
    const sensors = [{
      id: 8,
      name: 'Air Humidity Indoor',
      sensor_type: 'humidity',
      zone_id: 2,
      wiring: { source: 'dht22', gpio_pin: 4, device_id: 1 },
    }]
    const { byPhysical } = assignmentsForDevice(1, sensors, [])
    expect(byPhysical.get(7)).toHaveLength(1)
    expect(byPhysical.get(7)[0].name).toBe('Air Humidity Indoor')
  })

  it('relay actuators appear on I2C attachment list, not GPIO pins', () => {
    const actuators = [{
      id: 3,
      name: 'Flower pump',
      actuator_type: 'pump',
      zone_id: 2,
      device_id: 1,
      hardware_identifier: 'relay_ch_2',
    }]
    const { relayChannels, byPhysical } = assignmentsForDevice(1, [], actuators)
    expect(relayChannels).toHaveLength(1)
    expect(relayChannels[0].channel).toBe(2)
    expect(byPhysical.size).toBe(0)
  })

  it('devicesWithWiring returns devices with assignments', () => {
    const devices = [{ id: 1, name: 'Veg Pi' }]
    const sensors = [{
      id: 1,
      zone_id: 1,
      wiring: { gpio_pin: 4, device_id: 1, source: 'dht22' },
    }]
    expect(devicesWithWiring(devices, sensors, [])).toHaveLength(1)
  })

  it('registers /virtual-pi route', () => {
    expect(router.resolve('/virtual-pi').name).toBe('virtual-pi')
  })

  it('sidebar includes Wiring under Grow & operate', () => {
    const grow = buildNavGroups().find((g) => g.label === 'Grow & operate')
    const wiring = grow.items.find((i) => i.to === '/virtual-pi')
    expect(wiring?.label).toBe('Wiring')
    expect(wiring?.navTitle).toMatch(/Virtual Pi/i)
  })

  it('VirtualPi view and board components exist', () => {
    const uiSrc = join(process.cwd(), 'src')
    expect(readFileSync(join(uiSrc, 'views/VirtualPi.vue'), 'utf8')).toContain('VirtualPiBoard')
    expect(readFileSync(join(uiSrc, 'components/VirtualPiBoard.vue'), 'utf8')).toContain('virtual-pi-board')
  })
})
