/**
 * Phase 58 WS1 — task complete + consumption validation.
 */
import { describe, it, expect } from 'vitest'
import { validateConsumptionQty, formatConsumptionLine } from '../lib/taskConsumption.js'

describe('task consumption helpers', () => {
  it('rejects zero or negative quantity', () => {
    expect(validateConsumptionQty(0, { current_quantity_remaining: 10 })).toContain('positive')
    expect(validateConsumptionQty(-1, { current_quantity_remaining: 10 })).toContain('positive')
  })

  it('formats consumption line', () => {
    expect(formatConsumptionLine({ quantity: 2.5, notes: 'mix' })).toContain('2.5')
    expect(formatConsumptionLine({ quantity: 1 })).toBe('1 used')
  })
})
