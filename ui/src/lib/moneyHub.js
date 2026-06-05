/**
 * Phase 43 WS4 — farmer money hub helpers.
 */

/** Simplified spend categories for receipt capture (maps to cost API enums). */
export const FARMER_SPEND_CATEGORIES = [
  { label: 'Supplies & inputs', value: 'fertilizers_soil_amendments' },
  { label: 'Utilities', value: 'utilities_electricity_gas' },
  { label: 'Equipment', value: 'equipment_purchase_rental' },
  { label: 'Labor', value: 'labor_wages' },
  { label: 'Packaging & shipping', value: 'packaging_supplies' },
  { label: 'Other', value: 'miscellaneous' },
]

/**
 * @param {string|object} raw
 */
export function parseTransactionDate(raw) {
  if (!raw) return null
  let s = ''
  if (typeof raw === 'string') s = raw.slice(0, 10)
  else if (raw.Time) s = String(raw.Time).slice(0, 10)
  else return null
  const parts = s.split('-').map(Number)
  if (parts.length < 3 || !parts[0] || !parts[1] || !parts[2]) return null
  const d = new Date(parts[0], parts[1] - 1, parts[2])
  return Number.isNaN(d.getTime()) ? null : d
}

/**
 * @param {object[]} transactions
 * @param {Date} [referenceDate]
 */
export function computeMonthSummary(transactions, referenceDate = new Date()) {
  const year = referenceDate.getFullYear()
  const month = referenceDate.getMonth()
  let expenses = 0
  let income = 0
  let count = 0

  for (const t of transactions || []) {
    const d = parseTransactionDate(t.transaction_date)
    if (!d || d.getFullYear() !== year || d.getMonth() !== month) continue
    count += 1
    const amt = Number(t.amount) || 0
    if (t.is_income) income += amt
    else expenses += amt
  }

  const monthLabel = referenceDate.toLocaleString(undefined, { month: 'long', year: 'numeric' })
  return {
    monthLabel,
    expenses,
    income,
    net: income - expenses,
    count,
  }
}

/**
 * @param {string} category
 */
export function formatSpendCategory(category) {
  const farmer = FARMER_SPEND_CATEGORIES.find((c) => c.value === category)
  if (farmer) return farmer.label
  return category ? String(category).replace(/_/g, ' ') : 'Spend'
}

/**
 * Farmer-facing row for recent activity (no GL / COA fields).
 * @param {object} tx
 */
export function buildMoneyActivityRow(tx) {
  const amt = Number(tx.amount) || 0
  return {
    id: tx.id,
    clientCostId: tx._offline?.clientCostId,
    date: parseTransactionDate(tx.transaction_date),
    dateLabel: isoDateLabel(tx.transaction_date),
    label: tx.description || tx.counterparty || formatSpendCategory(tx.category),
    vendor: tx.counterparty || '',
    categoryLabel: formatSpendCategory(tx.category),
    amount: amt,
    currency: tx.currency || 'USD',
    isIncome: Boolean(tx.is_income),
    hasReceipt: Boolean(tx.receipt_file_id),
    receiptFileId: tx.receipt_file_id,
    queued: Boolean(tx._offline?.queued),
    stale: Boolean(tx._offline?.stale),
    advancedLink: tx.id ? { path: '/costs', query: { highlight: String(tx.id) } } : null,
  }
}

/**
 * @param {string|object} raw
 */
export function isoDateLabel(raw) {
  const d = parseTransactionDate(raw)
  if (!d) return '—'
  return d.toISOString().slice(0, 10)
}

/**
 * @param {number|null|undefined} n
 */
export function formatMoney(n) {
  if (n == null || Number.isNaN(Number(n))) return '0.00'
  return Number(n).toFixed(2)
}

/**
 * @param {object[]} transactions
 * @param {number} [limit]
 */
export function buildRecentMoneyRows(transactions, limit = 12) {
  return (transactions || [])
    .map(buildMoneyActivityRow)
    .sort((a, b) => {
      const ad = a.date ? a.date.getTime() : 0
      const bd = b.date ? b.date.getTime() : 0
      return bd - ad
    })
    .slice(0, limit)
}
