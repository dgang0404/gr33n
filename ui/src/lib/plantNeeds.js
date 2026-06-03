/**
 * Phase 38 — classify sensors/actuators by what the plant needs.
 * Single source of truth for zone hub, dashboard, and device list filters.
 */

export const PLANT_NEEDS = {
  water: 'water',
  light: 'light',
  air: 'air',
}

/** @typedef {'water'|'light'|'air'} PlantNeed */

export const NEED_META = {
  [PLANT_NEEDS.water]: {
    id: PLANT_NEEDS.water,
    label: 'Water & feeding',
    shortLabel: 'Water',
    icon: '💧',
    description: 'Fertigation, irrigation pumps, EC/pH, soil moisture.',
    manageLinks: [
      { to: '/fertigation', label: 'Fertigation programs' },
      { to: '/setpoints', label: 'Targets (setpoints)' },
    ],
  },
  [PLANT_NEEDS.light]: {
    id: PLANT_NEEDS.light,
    label: 'Light',
    shortLabel: 'Light',
    icon: '💡',
    description: 'Photoperiod and grow lights.',
    manageLinks: [
      { to: '/lighting', label: 'Lighting programs' },
      { to: '/schedules', label: 'Schedules' },
    ],
  },
  [PLANT_NEEDS.air]: {
    id: PLANT_NEEDS.air,
    label: 'Air & climate',
    shortLabel: 'Climate',
    icon: '🌬️',
    description: 'Temperature, humidity, fans, vents, shade.',
    manageLinks: [
      { to: '/automation', label: 'Automation rules' },
      { to: '/setpoints', label: 'Targets (setpoints)' },
    ],
  },
}

const WATER_SENSOR = new Set([
  'soil_moisture', 'moisture', 'ec', 'ph', 'water_level', 'flow_rate', 'water_temp',
  'dissolved_oxygen', 'tds',
])

const LIGHT_SENSOR = new Set(['lux', 'par', 'par_umol', 'ppfd', 'light_level'])

const AIR_SENSOR = new Set([
  'air_temp', 'temperature', 'temp', 'humidity', 'rh', 'co2', 'vpd', 'dew_point',
  'barometric_pressure', 'pressure',
])

const WATER_ACTUATOR = new Set([
  'pump', 'water_valve', 'return_pump', 'irrigation', 'drip', 'feeder_hopper',
])

const LIGHT_ACTUATOR = new Set(['light', 'grow_light'])

const AIR_ACTUATOR = new Set([
  'exhaust_fan', 'circulation_fan', 'ridge_vent', 'shade_screen', 'shade_cloth_motor',
  'glazing_panel', 'humidifier', 'dehumidifier', 'co2_injector', 'heat_lamp',
  'heater', 'cooler', 'fan', 'vent', 'shade',
])

function normType(t) {
  return String(t || '').toLowerCase().trim().replace(/\s+/g, '_')
}

/**
 * @param {string|undefined|null} sensorType
 * @returns {PlantNeed}
 */
export function sensorPlantNeed(sensorType) {
  const t = normType(sensorType)
  if (!t) return PLANT_NEEDS.air
  if (WATER_SENSOR.has(t)) return PLANT_NEEDS.water
  if (LIGHT_SENSOR.has(t)) return PLANT_NEEDS.light
  if (AIR_SENSOR.has(t)) return PLANT_NEEDS.air
  if (t.includes('moisture') || t.includes('ec') || t.includes('ph') || t.includes('water')) {
    return PLANT_NEEDS.water
  }
  if (t.includes('lux') || t.includes('par') || t.includes('light')) return PLANT_NEEDS.light
  if (t.includes('temp') || t.includes('humid') || t.includes('co2') || t.includes('fan') || t.includes('vent')) {
    return PLANT_NEEDS.air
  }
  return PLANT_NEEDS.air
}

/**
 * @param {string|undefined|null} actuatorType
 * @returns {PlantNeed}
 */
export function actuatorPlantNeed(actuatorType) {
  const t = normType(actuatorType)
  if (!t) return PLANT_NEEDS.air
  if (WATER_ACTUATOR.has(t)) return PLANT_NEEDS.water
  if (LIGHT_ACTUATOR.has(t)) return PLANT_NEEDS.light
  if (AIR_ACTUATOR.has(t)) return PLANT_NEEDS.air
  if (t.includes('pump') && !t.includes('air')) return PLANT_NEEDS.water
  if (t.includes('valve')) return PLANT_NEEDS.water
  if (t.includes('light')) return PLANT_NEEDS.light
  if (t.includes('fan') || t.includes('vent') || t.includes('shade') || t.includes('humid')) {
    return PLANT_NEEDS.air
  }
  if (t === 'relay') return PLANT_NEEDS.water
  return PLANT_NEEDS.air
}

/** Actuator types that support timed pulse (on for N seconds then off). */
const PULSE_ACTUATOR = new Set([
  'pump', 'relay', 'water_valve', 'return_pump', 'air_pump', 'feeder_hopper',
])

/**
 * @param {string|undefined|null} actuatorType
 */
export function supportsPulseCommand(actuatorType) {
  const t = normType(actuatorType)
  if (PULSE_ACTUATOR.has(t)) return true
  return t.includes('pump') || t === 'relay'
}

export const NEED_TAB_ORDER = [PLANT_NEEDS.water, PLANT_NEEDS.light, PLANT_NEEDS.air]
