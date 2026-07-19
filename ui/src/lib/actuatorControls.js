/**
 * Phase 210 — manual control affordances per actuator_type (mirrors backend
 * internal/handler/actuator.ValidCommands, trimmed to what the zone UI shows).
 */

function norm(t) {
  return String(t || '').toLowerCase().trim().replace(/\s+/g, '_')
}

/** Actuators controlled with Open / Close (not a simple online toggle). */
export function isOpenCloseActuator(actuatorType) {
  const t = norm(actuatorType)
  return t === 'gate' || t === 'water_valve' || t === 'ridge_vent' || t === 'glazing_panel'
}

/** Actuators with a one-shot dispense command (feeder hopper). */
export function isDispenseActuator(actuatorType) {
  return norm(actuatorType) === 'feeder_hopper'
}

/** Labels + emoji for hardware summary chips (Animals, Aquaponics, etc.). */
const CHIP_META = {
  gate: { label: 'Gate', icon: '🚪' },
  feeder_hopper: { label: 'Feeder', icon: '🌾' },
  water_valve: { label: 'Waterer', icon: '💧' },
  pump: { label: 'Pump', icon: '⚙️' },
  air_pump: { label: 'Aeration', icon: '💨' },
  grow_light: { label: 'Light', icon: '💡' },
}

export function actuatorsInZone(actuators, zoneId) {
  const zid = Number(zoneId)
  return (actuators || []).filter((a) => Number(a.zone_id) === zid)
}

export function hardwareChipsForZone(actuators, zoneId) {
  return actuatorsInZone(actuators, zoneId).map((a) => {
    const meta = CHIP_META[norm(a.actuator_type)] || { label: a.name, icon: '⚡' }
    return { ...meta, state: a.current_state_text || 'offline' }
  })
}

/** Animal-primary zones — feeder / waterer / gate only. */
export function animalHardwareChips(actuators, zoneId) {
  const allowed = new Set(['feeder_hopper', 'water_valve', 'gate'])
  return actuatorsInZone(actuators, zoneId)
    .filter((a) => allowed.has(norm(a.actuator_type)))
    .map((a) => {
      const meta = CHIP_META[norm(a.actuator_type)]
      return { ...meta, state: a.current_state_text || 'offline' }
    })
}
