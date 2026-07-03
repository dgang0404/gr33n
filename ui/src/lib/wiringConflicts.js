/**
 * Phase 120 — farm-wide wiring conflict scan for Virtual Pi board.
 */

import { findWiringConflict, resolveWiring, wiringIsEmpty } from './hardwareWiring.js'
import { pinByBcm } from './piPinMap.js'

function channelFromActuator(a) {
  const hi = a?.hardware_identifier
  if (hi == null || hi === '') return null
  const m = String(hi).match(/(\d+)$/)
  if (!m) return null
  const n = parseInt(m[1], 10)
  return Number.isFinite(n) && n >= 0 ? n : null
}

function deviceIdForEntity(entity, wiring) {
  if (wiring?.device_id != null) return Number(wiring.device_id)
  if (entity?.device_id != null) return Number(entity.device_id)
  return null
}

/**
 * @param {number} deviceId
 * @param {object[]} sensors
 * @param {object[]} actuators
 */
export function collectDeviceWiringConflicts(deviceId, sensors = [], actuators = []) {
  /** @type {object[]} */
  const conflicts = []
  /** @type {Set<number>} */
  const conflictPhysicalPins = new Set()
  /** @type {Set<number>} */
  const conflictChannels = new Set()

  for (const s of sensors) {
    const w = resolveWiring(s)
    if (deviceIdForEntity(s, w) !== deviceId) continue
    if (w?.gpio_pin == null) continue
    const hit = findWiringConflict({
      wiring: w,
      entityType: 'sensor',
      entityId: s.id,
      sensors,
      actuators,
    })
    if (hit) {
      const physical = pinByBcm(Number(w.gpio_pin))?.physical
      if (physical != null) conflictPhysicalPins.add(physical)
      conflicts.push({
        kind: 'gpio',
        entityType: 'sensor',
        entityId: s.id,
        entityName: s.name || s.sensor_type,
        bcm: w.gpio_pin,
        physical,
        message: hit.message,
        otherType: hit.entity_type,
        otherId: hit.entity_id,
        otherName: hit.entity_name,
      })
    }
  }

  for (const a of actuators) {
    const w = resolveWiring(a)
    const devId = deviceIdForEntity(a, w)
    if (devId !== deviceId && Number(a.device_id) !== deviceId) continue

    if (w?.gpio_pin != null) {
      const hit = findWiringConflict({
        wiring: w,
        entityType: 'actuator',
        entityId: a.id,
        sensors,
        actuators,
      })
      if (hit) {
        const physical = pinByBcm(Number(w.gpio_pin))?.physical
        if (physical != null) conflictPhysicalPins.add(physical)
        conflicts.push({
          kind: 'gpio',
          entityType: 'actuator',
          entityId: a.id,
          entityName: a.name,
          bcm: w.gpio_pin,
          physical,
          message: hit.message,
          otherType: hit.entity_type,
          otherId: hit.entity_id,
          otherName: hit.entity_name,
        })
      }
    }
  }

  /** @type {Map<number, object>} */
  const channelOwners = new Map()
  for (const a of actuators) {
    if (Number(a.device_id) !== deviceId) continue
    const ch = channelFromActuator(a)
    if (ch == null) continue
    const prev = channelOwners.get(ch)
    if (prev) {
      conflictChannels.add(ch)
      conflicts.push({
        kind: 'relay',
        channel: ch,
        message: `Channel ${ch} used by both "${prev.name}" and "${a.name}"`,
        entityType: 'actuator',
        entityId: a.id,
        entityName: a.name,
        otherType: 'actuator',
        otherId: prev.id,
        otherName: prev.name,
      })
    } else {
      channelOwners.set(ch, a)
    }
  }

  return { conflicts, conflictPhysicalPins, conflictChannels }
}

/** Sensors/actuators on this farm with no wiring yet. */
export function listUnwiredEntities(sensors = [], actuators = []) {
  const unwiredSensors = sensors.filter((s) => wiringIsEmpty(resolveWiring(s)))
  const unwiredActuators = actuators.filter((a) => {
    const w = resolveWiring(a)
    const ch = channelFromActuator(a)
    return wiringIsEmpty(w) && ch == null && !a.device_id
  })
  return { unwiredSensors, unwiredActuators }
}
