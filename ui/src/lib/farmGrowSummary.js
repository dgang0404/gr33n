/**
 * Phase 41 WS1 — farm-wide morning snapshot (client-side aggregate).
 */

import { scheduleRunsLabel } from './cronHumanize.js'
import { isOpenTask, isTaskDueToday, todayDateIso } from './zoneTasks.js'
import { countRoomsWithFeedingPlan } from './farmFeedingHub.js'

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
export function computeFarmMorningSnapshot(params) {
  const {
    tasks = [],
    alerts = [],
    schedules = [],
    devices = [],
    zones = [],
    programs = [],
    queueDepth = 0,
  } = params

  const dueToday = countFarmTasksDueToday(tasks)
  const unread = countFarmUnreadAlerts(alerts)
  const deviceSummary = countFarmDeviceSummary(devices)
  const nextSched = pickNextFarmSchedule(schedules)

  const chips = []

  chips.push({
    id: 'tasks-due',
    icon: '✅',
    label: 'Tasks due today',
    value: dueToday ? String(dueToday) : 'None',
    tone: dueToday ? 'warn' : 'ok',
    to: { path: '/tasks' },
  })

  chips.push({
    id: 'unread-alerts',
    icon: '🔔',
    label: 'Unread alerts',
    value: unread ? String(unread) : 'None',
    tone: unread ? 'warn' : 'ok',
    to: { path: '/alerts' },
  })

  if (nextSched) {
    chips.push({
      id: 'next-schedule',
      icon: '📅',
      label: 'Next schedule',
      value: nextSched.label,
      detail: nextSched.schedule.name,
      to: { path: '/schedules' },
    })
  } else {
    chips.push({
      id: 'next-schedule',
      icon: '📅',
      label: 'Next schedule',
      value: 'Nothing scheduled',
      tone: 'muted',
      to: { path: '/schedules' },
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
    const roomsWithPlan = countRoomsWithFeedingPlan(programs, zones)
    chips.push({
      id: 'feeding',
      icon: '💧',
      label: 'Feed & water',
      value: roomsWithPlan
        ? `${roomsWithPlan} of ${zones.length} rooms planned`
        : `${zones.length} room${zones.length === 1 ? '' : 's'} — no plans yet`,
      tone: roomsWithPlan ? 'ok' : 'muted',
      to: { path: '/feeding' },
    })
  }

  if (queueDepth > 0) {
    chips.push({
      id: 'queue',
      icon: '⏳',
      label: 'Queued commands',
      value: String(queueDepth),
      tone: 'warn',
      to: { path: '/feeding' },
    })
  }

  return { chips, dueToday, unread, queueDepth, deviceSummary }
}
