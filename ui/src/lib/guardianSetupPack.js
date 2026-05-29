/**
 * Phase 32 WS5 — format frozen apply_grow_setup_pack args for Confirm UX.
 */

function num(v) {
  if (v == null || v === '') return null
  const n = Number(v)
  return Number.isFinite(n) ? n : null
}

/** @param {Record<string, unknown>|null|undefined} args */
export function formatSetupPackBundle(args) {
  if (!args || typeof args !== 'object') {
    return { profile: '', zoneName: '', steps: [] }
  }

  const plant = /** @type {Record<string, unknown>} */ (args.plant || {})
  const cycle = /** @type {Record<string, unknown>} */ (args.cycle || {})
  const program = /** @type {Record<string, unknown>} */ (args.program || {})
  const task = /** @type {Record<string, unknown>} */ (args.optional_task || {})

  const steps = []

  const plantName = String(plant.display_name || '').trim()
  if (plantName) {
    let line = `Plant: ${plantName}`
    const variety = String(plant.variety_or_cultivar || '').trim()
    if (variety) line += ` (${variety})`
    const notes = String(plant.notes || '').trim()
    if (notes) line += ` — ${notes}`
    steps.push(line)
  }

  const zoneName = String(args.zone_name || '').trim()
  const zoneId = args.zone_id
  if (zoneName || zoneId != null) {
    steps.push(`Zone: ${zoneName || `#${zoneId}`}${zoneId != null && zoneName ? ` (#${zoneId})` : ''}`)
  }

  const cycleName = String(cycle.name || '').trim()
  const stage = String(cycle.current_stage || '').trim()
  const started = String(cycle.started_at || '').trim()
  if (cycleName || stage) {
    let line = 'Cycle:'
    if (cycleName) line += ` ${cycleName}`
    if (stage) line += ` · stage ${stage}`
    if (started) line += ` · started ${started}`
    steps.push(line)
  }

  const progName = String(program.name || '').trim()
  const vol = num(program.total_volume_liters)
  const ec = num(program.ec_trigger_low)
  const phLo = num(program.ph_trigger_low)
  const phHi = num(program.ph_trigger_high)
  if (progName || vol != null) {
    let line = 'Program:'
    if (progName) line += ` ${progName}`
    const hints = []
    if (vol != null) hints.push(`${vol}L`)
    if (ec != null) hints.push(`EC low ${ec}`)
    if (phLo != null && phHi != null) hints.push(`pH ${phLo}–${phHi}`)
    else if (phLo != null) hints.push(`pH low ${phLo}`)
    if (hints.length) line += ` · ${hints.join(' · ')}`
    steps.push(line)
  }

  const taskTitle = String(task.title || '').trim()
  if (taskTitle) {
    steps.push(`Task: ${taskTitle}`)
  }

  return {
    profile: String(args.profile || '').trim(),
    zoneName,
    steps,
  }
}

export const SETUP_PACK_HIGH_RISK_COPY =
  'High-impact grow setup — Confirm creates a plant, active crop cycle, and fertigation program in one step. Review each row below; nothing is written until you Confirm.'
