/**
 * Phase 34 WS4 — "if you Confirm, this will…" impact explanation for any proposal.
 *
 * The backend computes an ordered `impact_summary` per tool/risk tier; this module
 * prefers that and falls back to a client-side derivation (generalized from the
 * Phase 32 setup-pack formatter) so cards still explain consequences offline.
 */
import { formatSetupPackBundle } from './guardianSetupPack.js'

function num(v) {
  if (v == null || v === '') return null
  const n = Number(v)
  return Number.isFinite(n) ? n : null
}

function fmtLiters(v) {
  const n = num(v)
  if (n == null) return null
  return Number.isInteger(n) ? `${n}L` : `${n}L`
}

function programHints(program) {
  const hints = []
  const vol = fmtLiters(program.total_volume_liters)
  if (vol) hints.push(vol)
  const ec = num(program.ec_trigger_low)
  if (ec != null) hints.push(`EC ${ec}`)
  const phLo = num(program.ph_trigger_low)
  const phHi = num(program.ph_trigger_high)
  if (phLo != null && phHi != null) hints.push(`pH ${phLo}–${phHi}`)
  else if (phLo != null) hints.push(`pH low ${phLo}`)
  return hints.length ? ` (${hints.join(' / ')})` : ''
}

/** Client-side fallback impact lines when the backend did not supply impact_summary. */
function deriveImpactLines(tool, args) {
  const a = args && typeof args === 'object' ? args : {}
  switch (tool) {
    case 'apply_grow_setup_pack': {
      const { steps } = formatSetupPackBundle(a)
      return steps
    }
    case 'patch_fertigation_program': {
      const parts = []
      const vol = fmtLiters(a.total_volume_liters)
      if (vol) parts.push(`volume → ${vol}`)
      const ec = num(a.ec_trigger_low)
      if (ec != null) parts.push(`EC target → ${ec}`)
      if (typeof a.is_active === 'boolean') parts.push(a.is_active ? 'set active' : 'set inactive')
      return parts.length
        ? [`Update fertigation program: ${parts.join(', ')} (no run triggered now)`]
        : ['Update the fertigation program (no run triggered now)']
    }
    case 'create_fertigation_program':
      return [`Create fertigation program ${String(a.name || '')}${programHints(a)} — no run triggered now`.trim()]
    case 'create_lighting_program': {
      const preset = String(a.preset_key || '')
      const z = a.zone_id != null ? ` for zone ${a.zone_id}` : ''
      return [`Create lighting program${preset ? ` from preset ${preset}` : ''}${z} — ON/OFF schedules will be created`.trim()]
    }
    case 'create_plant':
      return [`Create plant ${String(a.display_name || '')} (editable later)`.trim()]
    case 'create_crop_cycle':
      return [`Start crop cycle ${String(a.name || '')} (no harvest data yet)`.trim()]
    case 'create_task':
    case 'create_task_from_alert':
      return [`Create task ${String(a.title || '')} (reversible)`.trim()]
    case 'update_cycle_stage':
      return [`Update crop cycle stage to ${String(a.current_stage || '')} (reversible)`.trim()]
    case 'ack_alert':
      return ['Acknowledge the alert (reversible)']
    case 'mark_alert_read':
      return ['Mark the alert as read (reversible)']
    case 'enqueue_actuator_command':
      return ['Queue a hardware command — the Pi fires the relay on its next poll']
    case 'patch_rule':
      return a.is_active === false
        ? ['Disable this automation rule — it stops firing until re-enabled']
        : ['Update the automation rule (reversible)']
    default:
      return []
  }
}

/**
 * impactForProposal returns the ordered consequence lines and operator-stated
 * facts for a proposal, preferring backend-provided fields.
 * @param {Record<string, any>} proposal
 * @returns {{ lines: string[], facts: Array<Record<string, any>> }}
 */
export function impactForProposal(proposal) {
  if (!proposal || typeof proposal !== 'object') return { lines: [], facts: [] }
  const facts = Array.isArray(proposal.operator_provided) ? proposal.operator_provided : []
  let lines = Array.isArray(proposal.impact_summary) ? proposal.impact_summary.slice() : null
  if (!lines || !lines.length) {
    lines = deriveImpactLines(proposal.tool, proposal.args)
    for (const f of facts) {
      lines.push(`Assumes ${f.label || `${f.field} ${f.value} (operator-stated, not measured)`}`)
    }
  }
  return { lines, facts }
}

/** revisionLabel returns "Revision N" when a proposal has been refined. */
export function revisionLabel(proposal) {
  const r = Number(proposal?.revision)
  if (Number.isFinite(r) && r > 1) return `Revision ${r}`
  return ''
}
