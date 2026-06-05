/**
 * Phase 41 WS6 — shared lighting program display helpers.
 */

/**
 * @param {string|undefined|null} lightsOnAt HH:MM
 * @param {number|undefined|null} onHours
 */
export function computeOffTime(lightsOnAt, onHours) {
  if (!lightsOnAt) return '—'
  const [h, m] = lightsOnAt.split(':').map(Number)
  if (Number.isNaN(h)) return '—'
  const hours = Number(onHours) || 0
  const total = (h * 60 + (m || 0) + hours * 60) % (24 * 60)
  return `${String(Math.floor(total / 60)).padStart(2, '0')}:${String(total % 60).padStart(2, '0')}`
}

/**
 * @param {{ name?: string, lights_on_at?: string, on_hours?: number, off_hours?: number, is_active?: boolean }} prog
 */
export function formatLightingProgramSummary(prog) {
  if (!prog) return ''
  const off = computeOffTime(prog.lights_on_at, prog.on_hours)
  const status = prog.is_active ? 'ON' : 'off'
  return `${prog.name || 'Program'} — ${prog.on_hours || 0}h ON from ${prog.lights_on_at || '—'} (OFF ${off}) · ${status}`
}
