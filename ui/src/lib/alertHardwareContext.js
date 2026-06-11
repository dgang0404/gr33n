/** Phase 78 — resolve GPIO / relay context for threshold alerts. */

import { formatEntityHardwareLabel } from './hardwareWiring.js'

/**
 * @param {object|null|undefined} alert
 * @param {{ sensors?: object[], actuators?: object[] }} [ctx]
 * @returns {{ entityName: string, hardwareLabel: string, line: string, entityType?: string, entityId?: number } | null}
 */
export function alertHardwareContext(alert, { sensors = [], actuators = [] } = {}) {
  if (!alert) return null
  const sourceType = alert.triggering_event_source_type
  const sourceId = alert.triggering_event_source_id
  if (sourceId == null) return null

  if (sourceType === 'sensor') {
    const sensor = sensors.find((s) => Number(s.id) === Number(sourceId))
    if (!sensor) {
      return {
        entityName: `Sensor #${sourceId}`,
        hardwareLabel: '',
        line: `Sensor #${sourceId}`,
        entityType: 'sensor',
        entityId: Number(sourceId),
      }
    }
    const hardwareLabel = formatEntityHardwareLabel(sensor)
    const entityName = sensor.name || sensor.sensor_type || `Sensor #${sourceId}`
    return {
      entityName,
      hardwareLabel,
      line: hardwareLabel ? `${entityName} · ${hardwareLabel}` : entityName,
      entityType: 'sensor',
      entityId: sensor.id,
    }
  }

  if (sourceType === 'actuator') {
    const actuator = actuators.find((a) => Number(a.id) === Number(sourceId))
    if (!actuator) {
      return {
        entityName: `Actuator #${sourceId}`,
        hardwareLabel: '',
        line: `Actuator #${sourceId}`,
        entityType: 'actuator',
        entityId: Number(sourceId),
      }
    }
    const hardwareLabel = formatEntityHardwareLabel(actuator)
    const entityName = actuator.name || actuator.actuator_type || `Actuator #${sourceId}`
    return {
      entityName,
      hardwareLabel,
      line: hardwareLabel ? `${entityName} · ${hardwareLabel}` : entityName,
      entityType: 'actuator',
      entityId: actuator.id,
    }
  }

  return null
}

/**
 * @param {object|null|undefined} alert
 * @param {{ sensors?: object[], actuators?: object[] }} [ctx]
 * @returns {string}
 */
export function formatAlertHardwareLine(alert, ctx) {
  const hit = alertHardwareContext(alert, ctx)
  return hit?.line || ''
}
