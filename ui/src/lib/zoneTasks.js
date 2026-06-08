/**
 * Phase 40 WS6 — zone-scoped task helpers for Overview due-today strip.
 */

const CLOSED_STATUSES = new Set(['completed', 'cancelled'])

/**
 * @param {object} task
 */
export function isOpenTask(task) {
  return task && !CLOSED_STATUSES.has(String(task.status || '').toLowerCase())
}

/**
 * @param {object} task
 * @param {string} [todayIso] YYYY-MM-DD
 */
export function isTaskDueToday(task, todayIso = todayDateIso()) {
  if (!task?.due_date) return false
  return String(task.due_date).slice(0, 10) <= todayIso
}

/**
 * @param {object} task
 * @param {string} [todayIso]
 */
export function isTaskOverdue(task, todayIso = todayDateIso()) {
  if (!isOpenTask(task) || !task?.due_date) return false
  return String(task.due_date).slice(0, 10) < todayIso
}

/**
 * @param {object[]} tasks
 * @param {number} zoneId
 */
export function countZoneOverdueTasks(tasks, zoneId) {
  return (tasks || []).filter(
    (t) => Number(t.zone_id) === Number(zoneId) && isTaskOverdue(t),
  ).length
}

export function todayDateIso() {
  return new Date().toISOString().slice(0, 10)
}

/**
 * @param {object[]} tasks
 * @param {number} zoneId
 * @param {number} [limit]
 */
export function zoneTasksDueToday(tasks, zoneId, limit = 5) {
  const today = todayDateIso()
  return (tasks || [])
    .filter((t) => Number(t.zone_id) === Number(zoneId) && isOpenTask(t) && isTaskDueToday(t, today))
    .sort((a, b) => {
      const da = String(a.due_date || '').slice(0, 10)
      const db = String(b.due_date || '').slice(0, 10)
      if (da !== db) return da.localeCompare(db)
      return (b.priority ?? 0) - (a.priority ?? 0)
    })
    .slice(0, limit)
}

/**
 * @param {object[]} tasks
 * @param {number} zoneId
 */
export function countZoneOpenTasks(tasks, zoneId) {
  return (tasks || []).filter((t) => Number(t.zone_id) === Number(zoneId) && isOpenTask(t)).length
}

/**
 * @param {string|undefined|null} dueDate
 */
export function snoozeDueDateToTomorrow(dueDate) {
  const base = dueDate
    ? new Date(`${String(dueDate).slice(0, 10)}T12:00:00`)
    : new Date()
  base.setDate(base.getDate() + 1)
  return base.toISOString().slice(0, 10)
}

/**
 * @param {string|undefined|null} dueDate
 */
export function formatTaskDueLabel(dueDate) {
  if (!dueDate) return 'No due date'
  const d = String(dueDate).slice(0, 10)
  const today = todayDateIso()
  if (d < today) return 'Overdue'
  if (d === today) return 'Due today'
  try {
    return `Due ${new Date(`${d}T12:00:00`).toLocaleDateString(undefined, { month: 'short', day: 'numeric' })}`
  } catch {
    return `Due ${d}`
  }
}
