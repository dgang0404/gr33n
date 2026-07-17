/**
 * Phase 76 — Today dashboard workspace link helpers.
 */
import { describe, it, expect } from 'vitest'
import { readFileSync } from 'node:fs'
import { join } from 'node:path'
import {
  alertsViewAllRoute,
  firstOpenTaskZoneId,
  newTaskRoute,
  tasksViewAllRoute,
  zoneOpsRoute,
} from '../lib/dashboardWorkspaceLinks.js'

const LEGACY_DASHBOARD_PATHS = [
  '/feeding',
  '/fertigation',
  '/schedules',
  '/automation',
  '/tasks',
  '/alerts',
  '/operations/',
]

describe('Phase 76 — dashboard workspace links', () => {
  it('zone ops routes include tab and ops query', () => {
    expect(zoneOpsRoute(5, 'alerts')).toEqual({
      path: '/zones/5',
      query: { tab: 'ops', ops: 'alerts' },
    })
  })

  it('tasks view-all prefers first open task zone', () => {
    const today = new Date().toISOString().slice(0, 10)
    const route = tasksViewAllRoute([
      { status: 'todo', due_date: today, zone_id: 2 },
      { status: 'todo', due_date: today, zone_id: 7 },
    ], [{ id: 2 }, { id: 7 }])
    expect(route).toEqual({ path: '/zones/2', query: { tab: 'ops', ops: 'tasks' } })
  })

  it('alerts view-all uses farm inbox when unread spans zones or farm-wide alerts', () => {
    expect(alertsViewAllRoute(
      [{ is_read: false, subject_rendered: 'Humidity high — Flower Room' }],
      [{ id: 1, name: 'Flower Room' }],
      [],
    )).toEqual({
      path: '/zones/1',
      query: { tab: 'ops', ops: 'alerts' },
    })
    expect(alertsViewAllRoute(
      [
        { is_read: false, subject_rendered: 'Humidity high — Flower Room' },
        { is_read: false, subject_rendered: 'Device offline: Veg Relay Controller' },
      ],
      [{ id: 1, name: 'Flower Room' }, { id: 2, name: 'Veg Room' }],
      [],
    )).toEqual({ path: '/alerts' })
    expect(alertsViewAllRoute([{ is_read: false }], [{ id: 1 }, { id: 2 }], [])).toEqual({
      path: '/alerts',
    })
  })

  it('new task route opens zone Ops create when a zone exists', () => {
    expect(newTaskRoute([], [{ id: 4 }])).toEqual({
      path: '/zones/4',
      query: { tab: 'ops', ops: 'tasks', create: '1' },
    })
    expect(firstOpenTaskZoneId([{ status: 'todo', due_date: '2099-01-01', zone_id: 9 }])).toBe(9)
  })

  it('Dashboard.vue has no primary links to absorbed legacy routes', () => {
    const dash = readFileSync(join(process.cwd(), 'src/views/Dashboard.vue'), 'utf8')
    for (const legacy of LEGACY_DASHBOARD_PATHS) {
      expect(dash, `legacy path ${legacy}`).not.toMatch(new RegExp(`(?:to=|action-to=|:to=).*${legacy.replace('/', '\\/')}`))
    }
    expect(dash).toContain('dashboardWorkspaceLinks.js')
    expect(dash).toContain('feedWaterDailyLink')
    expect(dash).toContain('tasksViewAllLink')
  })
})
