/**
 * Phase 44 WS2 — add zone wizard helpers.
 */

import {
  getDomainEnums,
  wizardZoneTypes,
  greenhouseCoverTypes,
  greenhouseAutomationPolicies,
} from './domainEnums.js'

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

/** Wizard zone type cards from GET /platform/domain-enums. */
export function zoneSetupTypeOptions(enums) {
  return wizardZoneTypes(enums)
}

/** Greenhouse cover options from domain enums. */
export function zoneSetupCoverTypes(enums) {
  return greenhouseCoverTypes(enums)
}

/** Greenhouse automation policy options from domain enums. */
export function zoneSetupAutomationPolicies(enums) {
  return greenhouseAutomationPolicies(enums)
}

/** @deprecated use zoneSetupTypeOptions() */
export const ZONE_SETUP_TYPES = wizardZoneTypes(getDomainEnums())

/** @deprecated use zoneSetupCoverTypes() */
export const GREENHOUSE_COVER_TYPES = greenhouseCoverTypes(getDomainEnums())

/** @deprecated use zoneSetupAutomationPolicies() */
export const GREENHOUSE_AUTOMATION_POLICIES = greenhouseAutomationPolicies(getDomainEnums())

/**
 * Build POST /farms/{id}/zones body from wizard form state.
 * @param {object} form
 */
export function buildZoneCreatePayload(form) {
  const name = String(form.name || '').trim()
  if (!name) {
    throw new Error('Zone name is required')
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
  presets = [],
}) {
  if (!presetKey || !actuatorId || !zoneId) return null
  const preset = presets.find((p) => p.key === presetKey)
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
