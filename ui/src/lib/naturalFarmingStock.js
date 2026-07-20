/**
 * Phase 209 WS5 — on-hand batch bridge (ready batches + low stock).
 */
import {
  buildSupplyRows,
  filterLowStockAlerts,
  listLowStockBatches,
} from './suppliesHub.js'

export const READY_BATCH_STATUSES = new Set(['ready_for_use', 'partially_used'])

/**
 * @param {object[]} batches
 */
export function filterReadyBatches(batches) {
  return (batches || []).filter((b) => READY_BATCH_STATUSES.has(String(b.status || '')))
}

/**
 * @param {object[]} batches
 * @param {object[]} inputs
 */
export function stockRows(batches, inputs) {
  return buildSupplyRows(filterReadyBatches(batches), inputs)
}

/**
 * Low-stock rows limited to ready / partially-used batches.
 * @param {object[]} batches
 * @param {object[]} inputs
 */
export function lowStockFromReady(batches, inputs) {
  return listLowStockBatches(filterReadyBatches(batches), inputs)
}

export { filterLowStockAlerts }

/**
 * @param {number|string|null|undefined} value
 */
export function formatStockQty(value) {
  if (value == null || value === '') return '—'
  const n = Number(value)
  return Number.isFinite(n) ? String(n) : String(value)
}
