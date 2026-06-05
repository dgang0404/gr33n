import { describe, it, expect } from 'vitest'
import { computeFarmMorningSnapshot, countFarmTasksDueToday } from '../lib/farmGrowSummary.js'

describe('Phase 41 WS1 — farm morning snapshot', () => {
  it('counts open tasks due today', () => {
    const today = new Date().toISOString().slice(0, 10)
    const n = countFarmTasksDueToday([
      { status: 'todo', due_date: today },
      { status: 'completed', due_date: today },
      { status: 'todo', due_date: '2099-01-01' },
    ])
    expect(n).toBe(1)
  })

  it('builds morning chips with links', () => {
    const today = new Date().toISOString().slice(0, 10)
    const snap = computeFarmMorningSnapshot({
      tasks: [{ status: 'todo', due_date: today, title: 'Check EC' }],
      alerts: [{ is_read: false, is_acknowledged: false }],
      schedules: [{ id: 1, name: 'Water Daily', is_active: true, cron_expression: '0 6 * * *' }],
      devices: [{ status: 'online' }, { status: 'offline' }],
      zones: [{ id: 1, name: 'Veg' }, { id: 2, name: 'Flower' }],
      programs: [{ id: 10, target_zone_id: 1, is_active: true }],
      queueDepth: 2,
    })
    expect(snap.dueToday).toBe(1)
    expect(snap.unread).toBe(1)
    expect(snap.queueDepth).toBe(2)
    const ids = snap.chips.map((c) => c.id)
    expect(ids).toContain('tasks-due')
    expect(ids).toContain('feeding')
    expect(ids).toContain('queue')
    expect(snap.chips.find((c) => c.id === 'tasks-due').to).toEqual({ path: '/tasks' })
    expect(snap.chips.find((c) => c.id === 'feeding').to).toEqual({ path: '/feeding' })
    expect(snap.chips.find((c) => c.id === 'queue').to).toEqual({ path: '/feeding' })
  })
})
