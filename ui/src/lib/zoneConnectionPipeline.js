/**
 * Phase 54 WS1 — interactive zone connection pipeline segments.
 */

import { PLANT_NEEDS } from './plantNeeds.js'

/**
 * @param {object[]} [devices]
 */
export function resolvePipelineDeviceHint(devices = []) {
  const list = devices || []
  const hasOffline = list.some((d) => d.status !== 'online')
  return hasOffline ? '/pi-setup' : '/actuators'
}

/**
 * @param {object} [params]
 * @param {string} [params.need]
 * @param {string} [params.deviceHint]
 */
export function buildZoneConnectionSegments({ need = '', deviceHint = '/actuators' } = {}) {
  const automationLabel = need === PLANT_NEEDS.water
    ? 'automation or feed timing'
    : 'automation'

  return [
    { id: 'sensor', label: 'sensor reading', hint: '/sensors' },
    { id: 'target', label: 'target band', hint: '/comfort-targets' },
    { id: 'automation', label: automationLabel, hint: '/automation' },
    { id: 'control', label: 'pump/light/fan', hint: '/actuators' },
    { id: 'device', label: 'device', hint: deviceHint },
  ]
}
