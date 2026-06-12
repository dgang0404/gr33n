/**
 * Phase 72 — autolog deep-links resolve within Money / Feed & Water workspaces.
 */
import { describe, it, expect } from 'vitest'
import { autologContextLink, buildMoneyActivityRow } from '../lib/moneyHub.js'

describe('Phase 72 — money autolog links', () => {
  it('mixing autolog jumps to Feed & Water nutrients tab', () => {
    const link = autologContextLink({
      related_table_name: 'mixing_log',
      related_record_id: 1,
    })
    expect(link).toEqual({ path: '/feed-water', query: { tab: 'nutrients' } })
  })

  it('energy autolog jumps to Money ledger tab', () => {
    const link = autologContextLink({
      related_table_name: 'energy_readings',
      related_record_id: 2,
      category: 'utilities_electricity_gas',
    })
    expect(link).toEqual({ path: '/money', query: { tab: 'ledger' } })
  })

  it('manual receipt advanced link opens ledger with highlight', () => {
    const row = buildMoneyActivityRow({
      id: 42,
      transaction_date: '2026-06-02',
      amount: 10,
      category: 'miscellaneous',
      description: 'Receipt',
    })
    expect(row.advancedLink).toEqual({
      path: '/money',
      query: { tab: 'ledger', highlight: '42' },
    })
  })
})
