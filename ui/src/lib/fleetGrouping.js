/**
 * Phase 69 WS3 — fleet tab grouping and pin conflict helpers.
 */

import { findWiringConflict, resolveWiring } from './hardwareWiring.js'

/**
 * @param {Array<{ zone_id?: number | null }>} entities
 * @param {Array<{ id: number, name?: string }>} zones
 * @returns {Array<{ zoneId: number | null, zoneName: string, items: object[] }>}
 */
export function groupEntitiesByZone(entities, zones) {
  const zoneById = new Map(zones.map((z) => [z.id, z]))
  /** @type {Map<number | null, object[]>} */
  const buckets = new Map()

  for (const entity of entities || []) {
    const zoneId = entity.zone_id ?? null
    if (!buckets.has(zoneId)) buckets.set(zoneId, [])
    buckets.get(zoneId).push(entity)
  }

  return [...buckets.entries()]
    .sort(([a], [b]) => {
      if (a == null) return 1
      if (b == null) return -1
      return a - b
    })
    .map(([zoneId, items]) => ({
      zoneId,
      zoneName: zoneId == null ? 'Unassigned' : (zoneById.get(zoneId)?.name || `Zone ${zoneId}`),
      items,
    }))
}

/**
 * @param {object} actuator
 * @param {object[]} sensors
 * @param {object[]} actuators
 * @returns {object | null}
 */
export function actuatorPinConflict(actuator, sensors, actuators) {
  const wiring = resolveWiring(actuator)
  if (!wiring?.device_id) return null
  return findWiringConflict({
    wiring,
    entityType: 'actuator',
    entityId: actuator.id,
    sensors,
    actuators,
  })
}

/**
 * @param {object[]} sensors
 * @param {object[]} actuators
 * @returns {object[]}
 */
export function listFleetPinConflicts(sensors, actuators) {
  const hits = []
  for (const a of actuators || []) {
    const c = actuatorPinConflict(a, sensors, actuators)
    if (c) hits.push({ actuatorId: a.id, actuatorName: a.name, ...c })
  }
  for (const s of sensors || []) {
    const wiring = resolveWiring(s)
    if (!wiring?.device_id) continue
    const c = findWiringConflict({
      wiring,
      entityType: 'sensor',
      entityId: s.id,
      sensors,
      actuators,
    })
    if (c) hits.push({ sensorId: s.id, sensorName: s.name, ...c })
  }
  return hits
}
