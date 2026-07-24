/** Labels and formatting for GET /farms/:id/crop-cycles/:cid/ops-timeline events. */

export const CROP_OPS_KIND_LABELS = {
  stage: 'Stage',
  apply: 'Apply',
  mix: 'Mix',
  program_run: 'Program run',
  light: 'Light',
}

export function cropOpsKindLabel(kind) {
  return CROP_OPS_KIND_LABELS[kind] || kind || 'Event'
}

export function cropOpsKindClass(kind) {
  switch (kind) {
    case 'stage':
      return 'bg-amber-900/40 text-amber-300 border-amber-800/60'
    case 'apply':
      return 'bg-sky-900/40 text-sky-300 border-sky-800/60'
    case 'mix':
      return 'bg-emerald-900/40 text-emerald-300 border-emerald-800/60'
    case 'program_run':
      return 'bg-violet-900/40 text-violet-300 border-violet-800/60'
    case 'light':
      return 'bg-yellow-900/40 text-yellow-200 border-yellow-800/60'
    default:
      return 'bg-zinc-800 text-zinc-300 border-zinc-700'
  }
}

export function formatCropOpsWhen(iso) {
  if (!iso) return '—'
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return iso
  return d.toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}

export function formatCropOpsDateInput(iso) {
  if (!iso) return ''
  const d = new Date(iso)
  if (Number.isNaN(d.getTime())) return ''
  return d.toISOString().slice(0, 10)
}

export function formulaSnapshotLines(details) {
  if (!details || typeof details !== 'object') return []
  const snap = details.formula_snapshot
  const lines = []
  if (snap && typeof snap === 'object') {
    if (snap.recipe_name) lines.push(String(snap.recipe_name))
    if (snap.dilution_ratio != null) lines.push(`Dilution ${snap.dilution_ratio}`)
    const comps = Array.isArray(snap.components) ? snap.components : []
    for (const c of comps) {
      const name = c.input_name || `input #${c.input_definition_id}`
      const part = c.part_value != null ? `${c.part_value}` : ''
      lines.push(part ? `${name}: ${part}` : name)
    }
  }
  if (details.application_recipe_revision_id != null) {
    lines.push(`Revision #${details.application_recipe_revision_id}`)
  }
  if (details.formula_revision_unpinned) {
    lines.push('Latest revision (program not pinned)')
  }
  return lines
}

export function cropOpsEventHasFormula(details) {
  return formulaSnapshotLines(details).length > 0
}

export function cropOpsEventSubtitle(event) {
  const d = event?.details || {}
  const bits = []
  if (event?.kind === 'apply') {
    if (d.volume_liters != null) bits.push(`${d.volume_liters} L`)
    if (d.ec_after_mscm != null) bits.push(`EC ${d.ec_after_mscm}`)
    if (d.ph_after != null) bits.push(`pH ${d.ph_after}`)
  }
  if (event?.kind === 'mix') {
    if (d.water_volume_liters != null) bits.push(`${d.water_volume_liters} L water`)
    const n = Array.isArray(d.components) ? d.components.length : 0
    if (n) bits.push(`${n} component${n === 1 ? '' : 's'}`)
  }
  if (event?.kind === 'program_run' && d.status) bits.push(String(d.status))
  if (event?.kind === 'light') {
    if (d.on_hours != null && d.off_hours != null) bits.push(`${d.on_hours}h on / ${d.off_hours}h off`)
    else if (d.status) bits.push(String(d.status))
  }
  if (event?.kind === 'stage' && d.growth_stage) bits.push(String(d.growth_stage))
  return bits.join(' · ')
}
