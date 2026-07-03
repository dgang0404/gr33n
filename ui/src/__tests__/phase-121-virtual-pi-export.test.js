import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import {
  FALLBACK_DRIVER_HOOKUPS,
  hookupStepsForDriver,
  wiringSourceForEntity,
} from '../lib/driverHookups.js'
import { physicalPinsForHookupRoles } from '../lib/piPinMap.js'
import { wiringDriftStatus, wiringDriftLabel } from '../lib/piConfigDrift.js'

describe('Phase 121 — Virtual Pi hookup + export', () => {
  it('DHT22 hookup lists VCC, DATA, GND', () => {
    const steps = hookupStepsForDriver(FALLBACK_DRIVER_HOOKUPS, 'dht22')
    expect(steps.map((s) => s.wire)).toEqual(['VCC', 'DATA', 'GND'])
    expect(steps.some((s) => s.role === 'gpio')).toBe(true)
  })

  it('physicalPinsForHookupRoles highlights 3v3 and chosen GPIO', () => {
    const roles = ['power3v3', 'gpio', 'gnd']
    const pins = physicalPinsForHookupRoles(roles, 4)
    expect(pins.has(1)).toBe(true)
    expect(pins.has(7)).toBe(true)
    expect(pins.has(6)).toBe(true)
  })

  it('wiringDriftStatus detects stale Pi hash', () => {
    const device = { config: { config_sha256: 'abc' } }
    expect(wiringDriftStatus(device, 'def')).toBe('stale')
    expect(wiringDriftStatus(device, 'abc')).toBe('synced')
    expect(wiringDriftLabel('stale')).toContain('stale')
  })

  it('wiringSourceForEntity resolves sensor driver', () => {
    expect(wiringSourceForEntity('sensor', { wiring: { source: 'dht22' } })).toBe('dht22')
    expect(wiringSourceForEntity('actuator', { config: { wiring: { source: 'relay_hat' } } })).toBe('relay_hat')
  })

  it('VirtualPi view includes download, print, and drift UI', () => {
    const view = readFileSync(join(process.cwd(), 'src/views/VirtualPi.vue'), 'utf8')
    expect(view).toContain('virtual-pi-download-config')
    expect(view).toContain('virtual-pi-print')
    expect(view).toContain('virtual-pi-wiring-stale')
    expect(view).toContain('pi-config')
  })

  it('PinWiringDrawer shows driver hookup steps', () => {
    const drawer = readFileSync(join(process.cwd(), 'src/components/PinWiringDrawer.vue'), 'utf8')
    expect(drawer).toContain('DriverHookupSteps')
    expect(drawer).toContain('hookup-change')
  })

  it('taxonomy payload includes driver_hookups on server', () => {
    const reg = readFileSync(join(process.cwd(), '../internal/platform/devicetaxonomy/registry.go'), 'utf8')
    expect(reg).toContain('DriverHookups')
    expect(reg).toContain('driver_hookups')
  })
})
