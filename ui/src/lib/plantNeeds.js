/**
 * Phase 38 / 90 — classify sensors/actuators by plant need via device taxonomy registry.
 */

import { getDeviceTaxonomy } from './deviceTaxonomy.js'

export const PLANT_NEEDS = {
  water: 'water',
  light: 'light',
  air: 'air',
}

/** @typedef {'water'|'light'|'air'} PlantNeed */

export const NEED_META = {
  [PLANT_NEEDS.water]: {
    id: PLANT_NEEDS.water,
    label: 'Feed & water',
    shortLabel: 'Water',
    icon: '💧',
    description: 'Feeding plan, irrigation pumps, EC/pH, soil moisture.',
    manageLinks: [],
  },
  [PLANT_NEEDS.light]: {
    id: PLANT_NEEDS.light,
    label: 'Light',
    shortLabel: 'Light',
    icon: '💡',
    description: 'Photoperiod and grow lights.',
    manageLinks: [
      { to: '/lighting', label: 'Lighting programs' },
    ],
  },
  [PLANT_NEEDS.air]: {
    id: PLANT_NEEDS.air,
    label: 'Air & climate',
    shortLabel: 'Climate',
    icon: '🌬️',
    description: 'Temperature, humidity, fans, vents, shade.',
    manageLinks: [
      { to: '/comfort-targets', label: 'Targets & schedules' },
      { to: '/automation', label: 'Automations (advanced)' },
    ],
  },
}

function normType(t) {
  return String(t || '').toLowerCase().trim().replace(/\s+/g, '_')
}

function lookupSensor(typeKey) {
  return getDeviceTaxonomy().sensorsByKey[normType(typeKey)]
}

function lookupActuator(typeKey) {
  return getDeviceTaxonomy().actuatorsByKey[normType(typeKey)]
}

/**
 * @param {string|undefined|null} sensorType
 * @returns {PlantNeed}
 */
export function sensorPlantNeed(sensorType) {
  const row = lookupSensor(sensorType)
  if (row?.plant_need) return row.plant_need
  const t = normType(sensorType)
  if (!t) return PLANT_NEEDS.air
  if (t.includes('moisture') || t.includes('ec') || t.includes('ph') || t.includes('water')) {
    return PLANT_NEEDS.water
  }
  if (t.includes('lux') || t.includes('par') || t.includes('light')) return PLANT_NEEDS.light
  return PLANT_NEEDS.air
}

/**
 * @param {string|undefined|null} actuatorType
 * @returns {PlantNeed}
 */
export function actuatorPlantNeed(actuatorType) {
  const row = lookupActuator(actuatorType)
  if (row?.plant_need) return row.plant_need
  const t = normType(actuatorType)
  if (!t) return PLANT_NEEDS.air
  if (t.includes('pump') && !t.includes('air')) return PLANT_NEEDS.water
  if (t.includes('valve') || t === 'relay') return PLANT_NEEDS.water
  if (t.includes('light')) return PLANT_NEEDS.light
  return PLANT_NEEDS.air
}

/**
 * @param {string|undefined|null} actuatorType
 */
export function supportsPulseCommand(actuatorType) {
  const row = lookupActuator(actuatorType)
  if (row) return !!row.supports_pulse
  const t = normType(actuatorType)
  return t.includes('pump') || t === 'relay'
}

/**
 * Greenhouse role from registry: shade | vent | fan | ''.
 * @param {string|undefined|null} actuatorType
 */
export function actuatorGHRole(actuatorType) {
  const row = lookupActuator(actuatorType)
  return row?.gh_role || ''
}

export const NEED_TAB_ORDER = [PLANT_NEEDS.water, PLANT_NEEDS.light, PLANT_NEEDS.air]
