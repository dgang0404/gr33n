import { describe, it, expect } from 'vitest'
import {
  computeMonthSummary,
  formatSpendCategory,
  buildMoneyActivityRow,
  buildRecentMoneyRows,
  formatMoney,
  isAutologgedTransaction,
  autologPlainLabel,
} from '../lib/moneyHub.js'

describe('Phase 43 WS4 — money hub helpers', () => {
  const ref = new Date('2026-06-15')

  it('computes this-month spend summary from transactions', () => {
    const summary = computeMonthSummary([
      { transaction_date: '2026-06-01', amount: 50, is_income: false },
      { transaction_date: '2026-06-10', amount: 20, is_income: false },
      { transaction_date: '2026-05-30', amount: 99, is_income: false },
      { transaction_date: '2026-06-12', amount: 100, is_income: true },
    ], ref)
    expect(summary.expenses).toBe(70)
    expect(summary.income).toBe(100)
    expect(summary.net).toBe(30)
    expect(summary.count).toBe(3)
    expect(summary.monthLabel).toContain('June')
  })

  it('maps spend category labels from domain enums', () => {
    expect(formatSpendCategory('miscellaneous')).toBe('miscellaneous')
    expect(formatSpendCategory('labor_wages')).toBe('labor wages')
  })

  it('builds activity rows without COA fields', () => {
    const row = buildMoneyActivityRow({
      id: 9,
      transaction_date: '2026-06-02',
      category: 'fertilizers_soil_amendments',
      description: 'OHN restock',
      amount: 42.5,
      currency: 'USD',
      is_income: false,
      receipt_file_id: 3,
    })
    expect(row.label).toBe('OHN restock')
    expect(row.categoryLabel).toBe('fertilizers soil amendments')
    expect(row.hasReceipt).toBe(true)
    expect(row.advancedLink).toEqual({ path: '/costs', query: { highlight: '9' } })
  })

  it('sorts recent rows newest first', () => {
    const rows = buildRecentMoneyRows([
      { id: 1, transaction_date: '2026-06-01', amount: 1, category: 'miscellaneous' },
      { id: 2, transaction_date: '2026-06-10', amount: 2, category: 'miscellaneous' },
    ])
    expect(rows[0].id).toBe(2)
  })

  it('formats money amounts', () => {
    expect(formatMoney(12.5)).toBe('12.50')
    expect(formatMoney(null)).toBe('0.00')
  })

  it('marks autolog rows with plain labels', () => {
    const tx = {
      id: 3,
      transaction_date: '2026-06-04',
      amount: 8,
      category: 'labor_wages',
      related_table_name: 'task_labor_log',
      related_record_id: 2,
    }
    expect(isAutologgedTransaction(tx)).toBe(true)
    expect(autologPlainLabel(tx)).toBe('From task labor time')
    const row = buildMoneyActivityRow(tx)
    expect(row.isAutolog).toBe(true)
    expect(row.autologLink?.path).toBe('/tasks')
  })
})
