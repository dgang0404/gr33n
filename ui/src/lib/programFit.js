/**
 * Phase 102 — fertigation program crop/stage fit helpers.
 */

/**
 * @param {object|null|undefined} metadata
 */
export function parseProgramMeta(metadata) {
  const m = metadata && typeof metadata === 'object' ? metadata : {}
  return {
    recommended_crop_keys: m.recommended_crop_keys || [],
    recommended_stages: m.recommended_stages || [],
    ec_band_mscm: m.ec_band_mscm || null,
  }
}

function includesFold(list, value) {
  const want = String(value || '').toLowerCase()
  if (!want) return false
  return (list || []).some((v) => String(v).toLowerCase() === want)
}

/**
 * @param {object|null|undefined} program
 * @param {{ cropKey?: string, stage?: string }} ctx
 */
export function programFitResult(program, ctx = {}) {
  const meta = parseProgramMeta(program?.metadata)
  if (!meta.recommended_crop_keys.length && !meta.recommended_stages.length) {
    return { ok: true, warnings: [] }
  }
  const warnings = []
  const cropKey = ctx.cropKey || ''
  const stage = ctx.stage || ''
  if (meta.recommended_crop_keys.length && cropKey && !includesFold(meta.recommended_crop_keys, cropKey)) {
    warnings.push(`program is tagged for ${meta.recommended_crop_keys.join(', ')} but grow uses ${cropKey}`)
  }
  if (meta.recommended_stages.length && stage && !includesFold(meta.recommended_stages, stage)) {
    warnings.push(`program is tagged for ${meta.recommended_stages.join(', ')} but grow is in ${stage}`)
  }
  return { ok: warnings.length === 0, warnings }
}

/**
 * @param {object|null|undefined} program
 * @param {{ cropKey?: string, stage?: string }} ctx
 * @returns {'fit'|'mismatch'|'unknown'}
 */
export function programFitBadge(program, ctx) {
  const meta = parseProgramMeta(program?.metadata)
  if (!meta.recommended_crop_keys.length && !meta.recommended_stages.length) return 'unknown'
  return programFitResult(program, ctx).ok ? 'fit' : 'mismatch'
}

/**
 * @param {object[]} programs
 * @param {{ cropKey?: string, stage?: string }} ctx
 */
export function sortProgramsByFit(programs, ctx) {
  return [...(programs || [])].sort((a, b) => {
    const rank = (p) => {
      const badge = programFitBadge(p, ctx)
      if (badge === 'fit') return 0
      if (badge === 'unknown') return 1
      return 2
    }
    const d = rank(a) - rank(b)
    if (d !== 0) return d
    return String(a.name || '').localeCompare(String(b.name || ''))
  })
}

/**
 * @param {object|null|undefined} program
 * @param {{ cropKey?: string, stage?: string }} ctx
 */
export function programOptionSuffix(program, ctx) {
  const badge = programFitBadge(program, ctx)
  if (badge === 'fit') return ' ✓'
  if (badge === 'mismatch') return ' ⚠'
  return ''
}

/** Operator-facing mismatch line for Water tab / Start grow (Phase 96). */
export function programMismatchSummary(program, ctx) {
  const fit = programFitResult(program, ctx)
  if (!fit.warnings.length) return ''
  return `${fit.warnings[0]} EC on the zone strip comes from the crop profile; the pump recipe may differ.`
}
