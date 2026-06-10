/**
 * Phase 41 WS1 — farm-wide morning snapshot (client-side aggregate).
 */

import { scheduleRunsLabel } from './cronHumanize.js'
import { isOpenTask, isTaskDueToday, todayDateIso } from './zoneTasks.js'
import { countZonesWithFeedingPlan } from './farmFeedingHub.js'
import {
  alertsViewAllRoute,
  comfortRoute,
  feedWaterRoute,
  moneyRoute,
  tasksViewAllRoute,
} from './dashboardWorkspaceLinks.js'

/**
 * @param {object[]} tasks
 */
export function countFarmTasksDueToday(tasks) {
  const today = todayDateIso()
  return (tasks || []).filter((t) => isOpenTask(t) && isTaskDueToday(t, today)).length
}

/**
 * @param {object[]} alerts
 */
export function countFarmUnreadAlerts(alerts) {
  return (alerts || []).filter((a) => !a.is_read && !a.is_acknowledged).length
}

/**
 * @param {object[]} devices
 */
export function countFarmDeviceSummary(devices) {
  const list = devices || []
  const online = list.filter((d) => d.status === 'online').length
  return { total: list.length, online, offline: list.length - online }
}

/**
 * @param {object[]} schedules
 */
export function pickNextFarmSchedule(schedules) {
  const active = (schedules || []).filter((s) => s.is_active !== false)
  if (!active.length) return null
  const ranked = active.map((s) => ({
    schedule: s,
    label: scheduleRunsLabel(s),
  }))
  ranked.sort((a, b) => String(a.schedule.name).localeCompare(String(b.schedule.name)))
  return ranked[0]
}

/**
 * @param {object} params
 * @returns {{ chips: Array<{ id: string, icon: string, label: string, value: string, tone?: string, to?: object|string }> }}
 */
/**
 * @param {number|null|undefined} amount
 */
function formatSpendChip(amount) {
  if (amount == null || Number.isNaN(Number(amount))) return '$0.00'
  return `$${Number(amount).toFixed(2)}`
}

export function computeFarmMorningSnapshot(params) {
  const {
    tasks = [],
    alerts = [],
    schedules = [],
    devices = [],
    zones = [],
    programs = [],
    queueDepth = 0,
    lowStockCount = 0,
    monthExpenses = null,
    sensors = [],
  } = params

  const dueToday = countFarmTasksDueToday(tasks)
  const unread = countFarmUnreadAlerts(alerts)
  const deviceSummary = countFarmDeviceSummary(devices)
  const nextSched = pickNextFarmSchedule(schedules)

  const chips = []

  if (dueToday > 0 || lowStockCount > 0) {
    const parts = []
    if (dueToday > 0) parts.push(`${dueToday} task${dueToday === 1 ? '' : 's'}`)
    if (lowStockCount > 0) parts.push(`${lowStockCount} low stock`)
    const doNextTo = dueToday > 0
      ? tasksViewAllRoute(tasks, zones)
      : moneyRoute('supplies')
    chips.push({
      id: 'do-next',
      icon: '🎯',
      label: 'Do next',
      value: `${dueToday + lowStockCount} item${dueToday + lowStockCount === 1 ? '' : 's'}`,
      detail: parts.join(' · '),
      tone: 'warn',
      to: doNextTo,
    })
  }

  chips.push({
    id: 'tasks-due',
    icon: '✅',
    label: 'Tasks due today',
    value: dueToday ? String(dueToday) : 'None',
    tone: dueToday ? 'warn' : 'ok',
    to: tasksViewAllRoute(tasks, zones),
  })

  chips.push({
    id: 'unread-alerts',
    icon: '🔔',
    label: 'Unread alerts',
    value: unread ? String(unread) : 'None',
    tone: unread ? 'warn' : 'ok',
    to: alertsViewAllRoute(alerts, zones, sensors),
  })

  if (lowStockCount > 0) {
    chips.push({
      id: 'low-stock',
      icon: '🧪',
      label: 'Supplies low',
      value: `${lowStockCount} batch${lowStockCount === 1 ? '' : 'es'}`,
      tone: 'warn',
      to: moneyRoute('supplies'),
    })
  }

  if (monthExpenses != null && Number(monthExpenses) > 0) {
    chips.push({
      id: 'month-spend',
      icon: '💵',
      label: 'Spent this month',
      value: formatSpendChip(monthExpenses),
      tone: 'muted',
      to: moneyRoute('summary'),
    })
  }

  if (nextSched) {
    chips.push({
      id: 'next-schedule',
      icon: '📅',
      label: 'Next schedule',
      value: nextSched.label,
      detail: nextSched.schedule.name,
      to: comfortRoute('schedules'),
    })
  } else {
    chips.push({
      id: 'next-schedule',
      icon: '📅',
      label: 'Next schedule',
      value: 'Nothing scheduled',
      tone: 'muted',
      to: comfortRoute('schedules'),
    })
  }

  const deviceValue = deviceSummary.total
    ? (deviceSummary.offline
      ? `${deviceSummary.online} online · ${deviceSummary.offline} offline`
      : `${deviceSummary.online} online`)
    : 'No devices'
  chips.push({
    id: 'devices',
    icon: '📡',
    label: 'Devices',
    value: deviceValue,
    tone: deviceSummary.offline ? 'warn' : 'ok',
    to: { path: '/zones' },
  })

  if (zones.length) {
    const zonesWithPlan = countZonesWithFeedingPlan(programs, zones)
    chips.push({
      id: 'feeding',
      icon: '💧',
      label: 'Feed & water',
      value: zonesWithPlan
        ? `${zonesWithPlan} of ${zones.length} zones planned`
        : `${zones.length} zone${zones.length === 1 ? '' : 's'} — no plans yet`,
      tone: zonesWithPlan ? 'ok' : 'muted',
      to: feedWaterRoute('daily'),
    })
  }

  if (queueDepth > 0) {
    chips.push({
      id: 'queue',
      icon: '⏳',
      label: 'Queued commands',
      value: String(queueDepth),
      tone: 'warn',
      to: feedWaterRoute('daily'),
    })
  }

  return { chips, dueToday, unread, queueDepth, deviceSummary }
}
