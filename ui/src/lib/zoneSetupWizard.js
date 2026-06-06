/**
 * Phase 44 WS2 — add zone wizard helpers.
 */

/** Farmer-facing zone types for the setup wizard (subset of API values). */
export const ZONE_SETUP_TYPES = [
  { value: 'indoor', label: 'Indoor grow room', hint: 'Tent, rack, or warehouse bay' },
  { value: 'greenhouse', label: 'Greenhouse', hint: 'Glazing, shade, vents, and climate profile' },
  { value: 'outdoor', label: 'Outdoor', hint: 'Garden bed, field, or patio grow' },
]

export const GREENHOUSE_COVER_TYPES = [
  { value: 'glass', label: 'Glass' },
  { value: 'polycarbonate', label: 'Polycarbonate' },
  { value: 'film', label: 'Film / poly' },
]

export const GREENHOUSE_AUTOMATION_POLICIES = [
  { value: 'manual', label: 'Manual only', hint: 'You control shade and fans' },
  { value: 'auto', label: 'Auto (sensor rules)', hint: 'Uses lux/temp sensors when wired' },
  { value: 'schedule_only', label: 'Schedule only', hint: 'Time-based, not sensor-driven' },
]

export const ZONE_LIGHTING_PRESETS = [
  { key: '', label: 'Skip for now', onHours: null },
  { key: 'veg_18_6', label: '18/6 Vegetative', onHours: 18 },
  { key: 'flower_12_12', label: '12/12 Flower', onHours: 12 },
  { key: 'seedling_16_8', label: '16/8 Seedling', onHours: 16 },
]

export function zoneSetupRoute(farmId) {
  return `/farms/${farmId}/zones/new`
}

export function isGreenhouseZoneType(zoneType) {
  return String(zoneType || '').toLowerCase() === 'greenhouse'
}

export function supportsLightingPreset(zoneType) {
  const t = String(zoneType || '').toLowerCase()
  return t === 'indoor' || t === 'greenhouse' || t === 'veg' || t === 'flower' || t === 'seedling'
}

/**
 * Build POST /farms/{id}/zones body from wizard form state.
 * @param {object} form
 */
export function buildZoneCreatePayload(form) {
  const name = String(form.name || '').trim()
  if (!name) {
    throw new Error('Room name is required')
  }
  const zoneType = form.zoneType || null
  const payload = {
    name,
    description: form.description?.trim() ? form.description.trim() : null,
    zone_type: zoneType,
  }
  if (isGreenhouseZoneType(zoneType)) {
    const gc = {
      cover_type: form.coverType || undefined,
      automation_policy: form.automationPolicy || 'manual',
      notes: form.greenhouseNotes?.trim() || undefined,
    }
    Object.keys(gc).forEach((k) => {
      if (gc[k] === undefined || gc[k] === '') delete gc[k]
    })
    if (Object.keys(gc).length) {
      payload.meta_data = { greenhouse_climate: gc }
    }
  }
  return payload
}

/**
 * @param {object} params
 */
export function buildLightingPresetRequest({
  farmId,
  zoneId,
  zoneName,
  presetKey,
  actuatorId,
  lightsOnAt = '06:00',
  timezone = 'America/New_York',
}) {
  if (!presetKey || !actuatorId || !zoneId) return null
  const preset = ZONE_LIGHTING_PRESETS.find((p) => p.key === presetKey)
  const label = preset?.label || presetKey
  return {
    url: `/farms/${farmId}/lighting-programs/from-preset`,
    body: {
      preset_key: presetKey,
      name: `${zoneName} — ${label}`,
      zone_id: zoneId,
      actuator_id: actuatorId,
      lights_on_at: lightsOnAt,
      timezone,
    },
  }
}

/**
 * Light/grow actuators eligible for a new zone program.
 * @param {object[]} actuators
 */
export function filterLightActuators(actuators) {
  return (actuators || []).filter(
    (a) => !a.deleted_at && (a.actuator_type === 'light' || a.actuator_type === 'grow_light'),
  )
}

/**
 * Farm devices not yet assigned to a zone (informational — full assign is WS3).
 * @param {object[]} devices
 */
export function listUnassignedDevices(devices) {
  return (devices || []).filter((d) => !d.deleted_at && (d.zone_id == null || d.zone_id === ''))
}
