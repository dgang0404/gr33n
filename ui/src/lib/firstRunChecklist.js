/**
 * Phase 44 WS5 — first-run Dashboard checklist (zones → device → targets → schedule).
 */

import { deviceSetupRoute } from './deviceSetupWizard.js'
import { zoneSetupRoute } from './zoneSetupWizard.js'
import { isAutologgedTransaction } from './moneyHub.js'

const DISMISS_PREFIX = 'gr33n_first_run_dismissed_'

/** @typedef {'add_zone'|'connect_device'|'comfort_targets'|'active_schedule'|'start_grow'|'restock_input'|'log_receipt'} FirstRunItemId */

/**
 * @param {object[]} [batches]
 */
function hasRestockedInput(batches = []) {
  return batches.some((b) => {
    const remaining = Number(b.current_quantity_remaining)
    const initial = Number(b.initial_quantity ?? b.quantity_received ?? 0)
    if (!Number.isFinite(remaining)) return false
    if (remaining > 0 && initial > 0 && remaining > initial) return true
    return remaining > 0
  })
}

/**
 * @param {object[]} [transactions]
 */
function hasManualReceipt(transactions = []) {
  return transactions.some((t) => !isAutologgedTransaction(t))
}

/**
 * @param {object} params
 * @param {object[]} [params.zones]
 * @param {object[]} [params.devices]
 * @param {object[]} [params.setpoints]
 * @param {object[]} [params.schedules]
 * @param {object[]} [params.cropCycles]
 * @param {object[]} [params.nfBatches]
 * @param {object[]} [params.costTransactions]
 * @param {boolean} [params.includeGrowClosure]
 * @param {number|null} [params.farmId]
 */
export function computeFirstRunChecklist({
  zones = [],
  devices = [],
  setpoints = [],
  schedules = [],
  cropCycles = [],
  nfBatches = [],
  costTransactions = [],
  includeGrowClosure = false,
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

  const firstZoneId = zones[0]?.id

  const base = [
    {
      id: 'add_zone',
      label: 'Add a zone',
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

  if (!includeGrowClosure) return base

  const growTo = firstZoneId
    ? { path: `/zones/${firstZoneId}`, query: { start_grow: '1' } }
    : '/plants?start_grow=1'

  return [
    ...base,
    {
      id: 'start_grow',
      label: 'Start a grow',
      done: cropCycles.length > 0,
      to: growTo,
      optional: true,
    },
    {
      id: 'restock_input',
      label: 'Restock one input',
      done: hasRestockedInput(nfBatches),
      to: '/operations/supplies',
      optional: true,
    },
    {
      id: 'log_receipt',
      label: 'Log first receipt',
      done: hasManualReceipt(costTransactions),
      to: '/money',
      optional: true,
    },
  ]
}

/** @param {Array<{ done: boolean, optional?: boolean }>} items */
export function isFirstRunComplete(items) {
  const required = (items || []).filter((i) => !i.optional)
  return required.length > 0 && required.every((i) => i.done)
}

/** True when any checklist step remains (ignores dismiss — for setup-mode persona). */
export function isFirstRunIncomplete(items) {
  return Array.isArray(items) && items.length > 0 && !isFirstRunComplete(items)
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
  if (Array.isArray(items) && items.length > 0 && items.every((i) => i.done)) return false
  if (isFirstRunChecklistDismissed(farmId)) return false
  return true
}
