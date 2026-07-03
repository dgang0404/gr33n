/**
 * Phase 90 — bundled device taxonomy fallback (mirrors migration seed).
 */
import { FALLBACK_DRIVER_HOOKUPS } from './driverHookups.js'

/** @param {object[]} entries */
function indexEntries(entries) {
  const sensorsByKey = {}
  const actuatorsByKey = {}
  for (const e of entries) {
    const key = String(e.type_key || '').toLowerCase()
    if (e.device_class === 'sensor') sensorsByKey[key] = e
    if (e.device_class === 'actuator') actuatorsByKey[key] = e
  }
  return { sensorsByKey, actuatorsByKey }
}

const SENSOR_ROWS = [
  ['soil_moisture', 'water', 'Soil moisture'],
  ['moisture', 'water', 'Moisture'],
  ['ec', 'water', 'EC'],
  ['ph', 'water', 'pH'],
  ['water_level', 'water', 'Water level'],
  ['flow_rate', 'water', 'Flow rate'],
  ['water_temp', 'water', 'Water temperature'],
  ['dissolved_oxygen', 'water', 'Dissolved oxygen'],
  ['tds', 'water', 'TDS'],
  ['lux', 'light', 'Light level'],
  ['par', 'light', 'PAR'],
  ['par_umol', 'light', 'PAR'],
  ['ppfd', 'light', 'PPFD'],
  ['light_level', 'light', 'Light level'],
  ['air_temp', 'air', 'Air temperature'],
  ['temperature', 'air', 'Temperature'],
  ['temp', 'air', 'Temperature'],
  ['temp_f', 'air', 'Temperature (°F)'],
  ['humidity', 'air', 'Humidity'],
  ['rh', 'air', 'Humidity'],
  ['co2', 'air', 'CO₂'],
  ['vpd', 'air', 'VPD'],
  ['dew_point', 'air', 'Dew point'],
  ['barometric_pressure', 'air', 'Barometric pressure'],
  ['pressure', 'air', 'Pressure'],
].map(([type_key, plant_need, display_label], i) => ({
  type_key,
  device_class: 'sensor',
  plant_need,
  display_label,
  supports_pulse: false,
  sort_order: i,
}))

const ACTUATOR_ROWS = [
  ['pump', 'water', 'Pump', true, null],
  ['water_valve', 'water', 'Water valve', true, null],
  ['return_pump', 'water', 'Return pump', true, null],
  ['irrigation', 'water', 'Irrigation', false, null],
  ['drip', 'water', 'Drip', false, null],
  ['feeder_hopper', 'water', 'Feeder hopper', true, null],
  ['relay', 'water', 'Relay', true, null],
  ['air_pump', 'water', 'Air pump', true, null],
  ['light', 'light', 'Light', false, null],
  ['grow_light', 'light', 'Grow light', false, null],
  ['exhaust_fan', 'air', 'Exhaust fan', false, 'fan'],
  ['circulation_fan', 'air', 'Circulation fan', false, 'fan'],
  ['fan', 'air', 'Fan', false, 'fan'],
  ['ridge_vent', 'air', 'Ridge vent', false, 'vent'],
  ['glazing_panel', 'air', 'Glazing panel', false, 'vent'],
  ['shade_screen', 'air', 'Shade screen', false, 'shade'],
  ['shade_cloth_motor', 'air', 'Shade cloth motor', false, 'shade'],
  ['shade', 'air', 'Shade', false, 'shade'],
  ['vent', 'air', 'Vent', false, 'vent'],
  ['humidifier', 'air', 'Humidifier', false, null],
  ['dehumidifier', 'air', 'Dehumidifier', false, null],
  ['co2_injector', 'air', 'CO₂ injector', false, null],
  ['heat_lamp', 'air', 'Heat lamp', false, null],
  ['heater', 'air', 'Heater', false, null],
  ['cooler', 'air', 'Cooler', false, null],
].map(([type_key, plant_need, display_label, supports_pulse, gh_role], i) => ({
  type_key,
  device_class: 'actuator',
  plant_need,
  display_label,
  supports_pulse,
  gh_role,
  sort_order: 100 + i,
}))

export const FALLBACK_WIRING_SOURCE_OPTIONS = [
  { value: 'dht22', label: 'DHT22 (temp / humidity)' },
  { value: 'ads1115', label: 'ADS1115 (analog)' },
  { value: 'mhz19', label: 'MH-Z19 (CO₂ serial)' },
  { value: 'bh1750', label: 'BH1750 (light I2C)' },
  { value: 'gpio_digital', label: 'GPIO digital' },
  { value: 'derived', label: 'Derived (computed)' },
]

/** @param {object|null|undefined} payload */
export function indexTaxonomy(payload) {
  const sensors = payload?.sensors || SENSOR_ROWS
  const actuators = payload?.actuators || ACTUATOR_ROWS
  const { sensorsByKey, actuatorsByKey } = indexEntries([...sensors, ...actuators])
  return {
    sensors,
    actuators,
    sensorsByKey,
    actuatorsByKey,
    wiring_source_options: payload?.wiring_source_options || FALLBACK_WIRING_SOURCE_OPTIONS,
    driver_hookups: payload?.driver_hookups || FALLBACK_DRIVER_HOOKUPS,
  }
}

export const FALLBACK_TAXONOMY = indexTaxonomy(null)
