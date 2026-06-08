/**
 * Phase 58 — task consumption helpers (client-side validation + display).
 */

/**
 * @param {number|string} qty
 * @param {object} batch
 * @returns {string} error message or empty string when valid
 */
export function validateConsumptionQty(qty, batch) {
  const q = Number(qty)
  if (!Number.isFinite(q) || q <= 0) return 'Enter a positive quantity.'
  const onHand = Number(batch?.current_quantity_remaining)
  if (Number.isFinite(onHand) && q > onHand) {
    return `Only ${onHand} on hand — you can't log more than that.`
  }
  return ''
}

/**
 * @param {object} consumption
 */
export function formatConsumptionLine(consumption) {
  if (!consumption) return ''
  const qty = consumption.quantity != null ? String(consumption.quantity) : '—'
  const notes = consumption.notes ? ` (${consumption.notes})` : ''
  return `${qty} used${notes}`
}

/**
 * Group consumptions by batch id.
 * @param {object[]} rows
 */
export function consumptionsByBatchId(rows) {
  const map = {}
  for (const row of rows || []) {
    const bid = row.input_batch_id
    if (!bid) continue
    if (!map[bid]) map[bid] = []
    map[bid].push(row)
  }
  return map
}
