import { describe, it, expect } from 'vitest'
import {
  buildZoneCreatePayload,
  buildLightingPresetRequest,
  filterLightActuators,
  listUnassignedDevices,
  isGreenhouseZoneType,
  supportsLightingPreset,
  zoneSetupRoute,
} from '../lib/zoneSetupWizard.js'

describe('Phase 44 WS2 — zone setup wizard helpers', () => {
  it('builds indoor zone payload without meta_data', () => {
    const payload = buildZoneCreatePayload({
      name: 'Veg tent',
      zoneType: 'indoor',
      description: 'Bench A',
    })
    expect(payload.name).toBe('Veg tent')
    expect(payload.zone_type).toBe('indoor')
    expect(payload.meta_data).toBeUndefined()
  })

  it('builds greenhouse payload with climate profile in meta_data', () => {
    const payload = buildZoneCreatePayload({
      name: 'Main GH',
      zoneType: 'greenhouse',
      coverType: 'polycarbonate',
      automationPolicy: 'auto',
      greenhouseNotes: 'South-facing',
    })
    expect(isGreenhouseZoneType(payload.zone_type)).toBe(true)
    expect(payload.meta_data.greenhouse_climate.cover_type).toBe('polycarbonate')
    expect(payload.meta_data.greenhouse_climate.automation_policy).toBe('auto')
  })

  it('requires room name', () => {
    expect(() => buildZoneCreatePayload({ name: '  ', zoneType: 'indoor' })).toThrow(/name/i)
  })

  it('builds lighting preset request when actuator and preset set', () => {
    const req = buildLightingPresetRequest({
      farmId: 1,
      zoneId: 9,
      zoneName: 'Flower',
      presetKey: 'veg_18_6',
      actuatorId: 3,
      timezone: 'America/Chicago',
    })
    expect(req.body.zone_id).toBe(9)
    expect(req.body.preset_key).toBe('veg_18_6')
    expect(req.body.timezone).toBe('America/Chicago')
  })

  it('filters light actuators and unassigned devices', () => {
    const lights = filterLightActuators([
      { id: 1, actuator_type: 'grow_light', deleted_at: null },
      { id: 2, actuator_type: 'pump', deleted_at: null },
    ])
    expect(lights).toHaveLength(1)
    const unassigned = listUnassignedDevices([
      { id: 10, zone_id: null },
      { id: 11, zone_id: 2 },
    ])
    expect(unassigned).toHaveLength(1)
  })

  it('supports lighting preset for indoor and greenhouse', () => {
    expect(supportsLightingPreset('indoor')).toBe(true)
    expect(supportsLightingPreset('greenhouse')).toBe(true)
    expect(supportsLightingPreset('outdoor')).toBe(false)
  })

  it('builds zone wizard route', () => {
    expect(zoneSetupRoute(5)).toBe('/farms/5/zones/new')
  })
})
