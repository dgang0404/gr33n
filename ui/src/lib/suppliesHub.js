/**
 * Phase 43 WS2 — farm-wide supplies hub helpers (client-side aggregates).
 */

/**
 * @param {object} batch
 */
export function isBatchLowStock(batch) {
  if (batch?.low_stock_threshold == null || batch?.low_stock_threshold === '') return false
  const remaining = Number(batch.current_quantity_remaining)
  const threshold = Number(batch.low_stock_threshold)
  if (!Number.isFinite(remaining) || !Number.isFinite(threshold)) return false
  return remaining < threshold
}

/**
 * @param {object[]} batches
 * @param {object[]} [inputs]
 */
export function listLowStockBatches(batches, inputs = []) {
  const inputById = new Map((inputs || []).map((i) => [i.id, i]))
  return (batches || [])
    .filter(isBatchLowStock)
    .map((b) => ({
      batch: b,
      inputName: inputById.get(b.input_definition_id)?.name || `Input #${b.input_definition_id}`,
      remaining: Number(b.current_quantity_remaining),
      threshold: Number(b.low_stock_threshold),
    }))
    .sort((a, b) => a.inputName.localeCompare(b.inputName))
}

/**
 * @param {object[]} batches
 * @param {object[]} [inputs]
 */
export function buildSupplyRows(batches, inputs = []) {
  const inputById = new Map((inputs || []).map((i) => [i.id, i]))
  return (batches || [])
    .map((b) => {
      const input = inputById.get(b.input_definition_id)
      return {
        id: b.id,
        batchLabel: b.batch_identifier || `#${b.id}`,
        inputName: input?.name || `Input #${b.input_definition_id}`,
        category: input?.category || '',
        remaining: b.current_quantity_remaining,
        threshold: b.low_stock_threshold,
        status: b.status,
        storageLocation: b.storage_location,
        scope: 'farm',
        lowStock: isBatchLowStock(b),
      }
    })
    .sort((a, b) => {
      if (a.lowStock !== b.lowStock) return a.lowStock ? -1 : 1
      return a.inputName.localeCompare(b.inputName)
    })
}

/**
 * Unread low-stock alerts from the worker (`inventory_low_stock`).
 * @param {object[]} alerts
 */
export function filterLowStockAlerts(alerts) {
  return (alerts || []).filter(
    (a) => !a.is_read && !a.is_acknowledged
      && a.triggering_event_source_type === 'inventory_low_stock',
  )
}
