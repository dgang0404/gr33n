/**
 * Phase 54 WS1 — zone connection pipeline helpers.
 */
import { describe, it, expect } from 'vitest'
import {
  buildZoneConnectionSegments,
  resolvePipelineDeviceHint,
} from '../lib/zoneConnectionPipeline.js'
import { PLANT_NEEDS } from '../lib/plantNeeds.js'

describe('Phase 54 WS1 — zone connection pipeline', () => {
  it('builds five segments with water-specific automation label', () => {
    const segs = buildZoneConnectionSegments({ need: PLANT_NEEDS.water })
    expect(segs).toHaveLength(5)
    expect(segs[0].hint).toBe('/sensors')
    expect(segs[1].hint).toBe('/comfort-targets')
    expect(segs[2].label).toContain('feed timing')
    expect(segs[4].label).toBe('device')
  })

  it('points device segment at pi-setup when any device is offline', () => {
    expect(resolvePipelineDeviceHint([{ status: 'online' }, { status: 'offline' }])).toBe('/pi-setup')
    expect(resolvePipelineDeviceHint([{ status: 'online' }])).toBe('/actuators')
    expect(resolvePipelineDeviceHint([])).toBe('/actuators')
  })
})
