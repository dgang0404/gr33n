import { describe, it, expect } from 'vitest'
import {
  channelToStackLevel,
  stackLevelToI2cAddress,
  stackLevelToDipBits,
  buildRelayStacks,
} from '../lib/relayStack.js'
import { collectDeviceWiringConflicts, listUnwiredEntities } from '../lib/wiringConflicts.js'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'

describe('Phase 120 — Virtual Pi interactive wiring', () => {
  it('channel 8 is stack 1 with I2C 0x26 and ID0 ON', () => {
    expect(channelToStackLevel(8)).toBe(1)
    expect(stackLevelToI2cAddress(1)).toBe('0x26')
    const dip = stackLevelToDipBits(1)
    expect(dip.id0).toBe(true)
    expect(dip.id1).toBe(false)
    expect(dip.id2).toBe(false)
  })

  it('channel 0 is stack 0 with I2C 0x27 and all DIP OFF', () => {
    expect(stackLevelToI2cAddress(0)).toBe('0x27')
    const dip = stackLevelToDipBits(0)
    expect(dip.id0).toBe(false)
    expect(dip.id1).toBe(false)
    expect(dip.id2).toBe(false)
  })

  it('buildRelayStacks creates second card when ch 8 is in use', () => {
    const stacks = buildRelayStacks([
      { channel: 0, name: 'Pump', id: 1 },
      { channel: 8, name: 'Light', id: 2 },
    ])
    expect(stacks).toHaveLength(2)
    expect(stacks[1].level).toBe(1)
    expect(stacks[1].dip.id0).toBe(true)
    expect(stacks[1].slots.find((s) => s.channel === 8)?.assigned?.name).toBe('Light')
    expect(stacks[1].slots.find((s) => s.channel === 9)?.assigned).toBeNull()
  })

  it('detects duplicate GPIO pin conflicts', () => {
    const sensors = [
      { id: 1, name: 'A', sensor_type: 'temp', wiring: { source: 'gpio_digital', gpio_pin: 17, device_id: 1 } },
      { id: 2, name: 'B', sensor_type: 'temp', wiring: { source: 'gpio_digital', gpio_pin: 17, device_id: 1 } },
    ]
    const { conflicts, conflictPhysicalPins } = collectDeviceWiringConflicts(1, sensors, [])
    expect(conflicts.length).toBeGreaterThan(0)
    expect(conflictPhysicalPins.size).toBeGreaterThan(0)
  })

  it('detects duplicate relay channel conflicts', () => {
    const actuators = [
      { id: 1, name: 'Pump A', device_id: 1, hardware_identifier: '2' },
      { id: 2, name: 'Pump B', device_id: 1, hardware_identifier: 'relay_ch_2' },
    ]
    const { conflicts, conflictChannels } = collectDeviceWiringConflicts(1, [], actuators)
    expect(conflicts.some((c) => c.kind === 'relay')).toBe(true)
    expect(conflictChannels.has(2)).toBe(true)
  })

  it('listUnwiredEntities finds sensors without wiring', () => {
    const { unwiredSensors } = listUnwiredEntities([
      { id: 1, sensor_type: 'humidity' },
      { id: 2, sensor_type: 'ec', wiring: { source: 'ads1115', i2c_channel: 0, device_id: 1 } },
    ], [])
    expect(unwiredSensors).toHaveLength(1)
    expect(unwiredSensors[0].id).toBe(1)
  })

  it('board and drawer components include interactive wiring', () => {
    const uiSrc = join(process.cwd(), 'src')
    const board = readFileSync(join(uiSrc, 'components/VirtualPiBoard.vue'), 'utf8')
    expect(board).toContain('PinWiringDrawer')
    expect(board).toContain('RelayStackView')
    expect(board).toContain('virtual-pi-conflicts')
    const drawer = readFileSync(join(uiSrc, 'components/PinWiringDrawer.vue'), 'utf8')
    expect(drawer).toContain('HardwareWiringPanel')
    expect(drawer).toContain('ActuatorWiringPanel')
  })
})
