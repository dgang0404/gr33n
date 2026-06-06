/**
 * Phase 44 WS5 — first-run Dashboard checklist (zones → device → targets → schedule).
 */

import { deviceSetupRoute } from './deviceSetupWizard.js'
import { zoneSetupRoute } from './zoneSetupWizard.js'

const DISMISS_PREFIX = 'gr33n_first_run_dismissed_'

/** @typedef {'add_zone'|'connect_device'|'comfort_targets'|'active_schedule'} FirstRunItemId */

/**
 * @param {object} params
 * @param {object[]} [params.zones]
 * @param {object[]} [params.devices]
 * @param {object[]} [params.setpoints]
 * @param {object[]} [params.schedules]
 * @param {number|null} [params.farmId]
 */
export function computeFirstRunChecklist({
  zones = [],
  devices = [],
  setpoints = [],
  schedules = [],
  farmId = null,
} = {}) {
  const hasZones = zones.length > 0
  const hasDevices = devices.length > 0
  const hasComfortTargets = setpoints.some(
    (sp) => sp.min_value != null || sp.max_value != null || sp.ideal_value != null,
  )
  const hasActiveSchedule = schedules.some((s) => s.is_active)

  const zoneTo = farmId ? zoneSetupRoute(farmId) : '/zones'
  const deviceTo = farmId ? deviceSetupRoute(farmId) : '/settings'

  return [
    {
      id: 'add_zone',
      label: 'Add a grow room',
      done: hasZones,
      to: zoneTo,
    },
    {
      id: 'connect_device',
      label: 'Connect edge device',
      done: hasDevices,
      to: deviceTo,
    },
    {
      id: 'comfort_targets',
      label: 'Set comfort targets',
      done: hasComfortTargets,
      to: '/comfort-targets',
    },
    {
      id: 'active_schedule',
      label: 'Turn on one schedule',
      done: hasActiveSchedule,
      to: '/comfort-targets?tab=schedules',
    },
  ]
}

/** @param {Array<{ done: boolean }>} items */
export function isFirstRunComplete(items) {
  return Array.isArray(items) && items.length > 0 && items.every((i) => i.done)
}

export function isFirstRunChecklistDismissed(farmId) {
  if (typeof localStorage === 'undefined' || !farmId) return false
  return localStorage.getItem(`${DISMISS_PREFIX}${farmId}`) === '1'
}

export function dismissFirstRunChecklist(farmId) {
  if (typeof localStorage === 'undefined' || !farmId) return
  localStorage.setItem(`${DISMISS_PREFIX}${farmId}`, '1')
}

export function clearFirstRunChecklistDismiss(farmId) {
  if (typeof localStorage === 'undefined' || !farmId) return
  localStorage.removeItem(`${DISMISS_PREFIX}${farmId}`)
}

/**
 * @param {number|null|undefined} farmId
 * @param {Array<{ done: boolean }>} items
 */
export function shouldShowFirstRunChecklist(farmId, items) {
  if (!farmId) return false
  if (isFirstRunComplete(items)) return false
  if (isFirstRunChecklistDismissed(farmId)) return false
  return true
}
