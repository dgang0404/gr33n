import { describe, it, expect } from 'vitest'
import {
  suggestDeviceUid,
  buildDeviceCreatePayload,
  buildActuatorCreatePayload,
  buildPiConfigSnippet,
  isDeviceOnline,
  deviceSetupRoute,
  DEVICE_ACTUATOR_TEMPLATES,
} from '../lib/deviceSetupWizard.js'

describe('Phase 44 WS3 — device setup wizard helpers', () => {
  it('suggests unique device uid', () => {
    const uid = suggestDeviceUid(3)
    expect(uid).toMatch(/^pi-farm3-/)
  })

  it('builds device create payload with zone', () => {
    const payload = buildDeviceCreatePayload({
      name: 'Veg Pi',
      deviceUid: 'pi-veg-01',
      deviceType: 'raspberry_pi_edge',
      zoneId: 5,
    })
    expect(payload.name).toBe('Veg Pi')
    expect(payload.zone_id).toBe(5)
    expect(payload.status).toBe('offline')
  })

  it('requires device name and uid', () => {
    expect(() => buildDeviceCreatePayload({ name: '', deviceUid: 'x' })).toThrow(/name/i)
    expect(() => buildDeviceCreatePayload({ name: 'Pi', deviceUid: '' })).toThrow(/uid/i)
  })

  it('builds actuator payload linked to device', () => {
    const template = DEVICE_ACTUATOR_TEMPLATES[0]
    const { body } = buildActuatorCreatePayload({
      farmId: 1,
      deviceId: 9,
      zoneId: 2,
      template,
      hardwareId: 'BCM17',
    })
    expect(body.device_id).toBe(9)
    expect(body.zone_id).toBe(2)
    expect(body.actuator_type).toBe('grow_light')
  })

  it('builds pi config snippet with farm and device ids', () => {
    const snippet = buildPiConfigSnippet({
      baseUrl: 'http://192.168.1.10:8080',
      farmId: 1,
      deviceId: 42,
      deviceUid: 'pi-veg-01',
    })
    expect(snippet).toContain('farm_id: 1')
    expect(snippet).toContain('device_id: 42')
    expect(snippet).toContain('pi-veg-01')
  })

  it('detects online device status', () => {
    expect(isDeviceOnline({ status: 'online' })).toBe(true)
    expect(isDeviceOnline({ status: 'offline' })).toBe(false)
  })

  it('builds device wizard route with optional zone', () => {
    expect(deviceSetupRoute(2)).toBe('/farms/2/devices/new')
    expect(deviceSetupRoute(2, 7)).toBe('/farms/2/devices/new?zone_id=7')
  })
})
