// Phase 216 WS6 — pure helpers for the Fertigation "live dosing sessions" panel.

const DOSING_STATUS_LABEL = {
  idle: 'Waiting to dose',
  dosing: 'Dosing…',
  circulating: 'Circulating…',
  settling: 'Settling…',
  sampling: 'Sampling…',
  evaluating: 'Evaluating…',
  done: 'Done',
  aborted: 'Aborted',
}

/**
 * Operator-facing label for a dosing_session's status.
 * @param {{ status: string }} session
 */
export function dosingStatusLabel(session) {
  return DOSING_STATUS_LABEL[session?.status] || session?.status || ''
}

/**
 * Human-readable elapsed time since a dosing session started.
 * @param {{ started_at: string }} session
 * @param {number} [now] - epoch ms, injectable for tests
 */
export function dosingElapsed(session, now = Date.now()) {
  const started = new Date(session?.started_at).getTime()
  if (!started || Number.isNaN(started)) return ''
  const mins = Math.floor((now - started) / 60000)
  if (mins < 1) return '<1 min'
  if (mins < 60) return `${mins} min`
  return `${Math.floor(mins / 60)}h ${mins % 60}m`
}

/**
 * True once a session is no longer actively running (worker won't touch it again).
 * @param {{ status: string }} session
 */
export function isDosingSessionTerminal(session) {
  return session?.status === 'done' || session?.status === 'aborted'
}
