/** Phase 211.05 WS5 — recipe track record formatting and ops attribution helpers. */

export const RECIPE_OUTCOME_MIN_SAMPLE = 2
export const RECIPE_ATTRIBUTION_THRESHOLD = 0.6

/**
 * @param {Array<{ kind: string, details?: Record<string, unknown> }>} events
 */
export function attributionHitsFromOpsEvents(events) {
  if (!Array.isArray(events)) return []
  return events
    .filter((e) => e?.kind === 'mix' || e?.kind === 'program_run')
    .map((e) => ({
      application_recipe_id: Number(e.details?.application_recipe_id),
      application_recipe_revision_id: numOrNull(e.details?.application_recipe_revision_id),
    }))
    .filter((h) => Number.isFinite(h.application_recipe_id) && h.application_recipe_id > 0)
}

/**
 * @param {Array<{ application_recipe_id: number, application_recipe_revision_id?: number|null }>} hits
 */
export function attributeRecipeFromHits(hits) {
  if (!hits?.length) {
    return { empty: true, mixed: false, application_recipe_id: null, application_recipe_revision_id: null }
  }
  const counts = new Map()
  for (const h of hits) {
    const rev = h.application_recipe_revision_id ?? 0
    const key = `${h.application_recipe_id}|${rev}`
    counts.set(key, (counts.get(key) || 0) + 1)
  }
  let bestKey = ''
  let bestCount = 0
  for (const [k, c] of counts) {
    if (c > bestCount) {
      bestCount = c
      bestKey = k
    }
  }
  const total = hits.length
  const mixed = bestCount / total < RECIPE_ATTRIBUTION_THRESHOLD
  const [rid, revRaw] = bestKey.split('|')
  const rev = Number(revRaw)
  return {
    empty: false,
    mixed,
    application_recipe_id: Number(rid),
    application_recipe_revision_id: rev > 0 ? rev : null,
  }
}

/**
 * @param {Array<Record<string, unknown>>} outcomes
 * @param {number} recipeId
 * @param {number|null|undefined} revisionId
 */
export function findRecipeOutcome(outcomes, recipeId, revisionId) {
  if (!Array.isArray(outcomes) || !recipeId) return null
  const exact = outcomes.find(
    (o) =>
      Number(o.application_recipe_id) === Number(recipeId)
      && (revisionId == null
        ? o.application_recipe_revision_id == null
        : Number(o.application_recipe_revision_id) === Number(revisionId)),
  )
  if (exact) return exact
  return outcomes.find((o) => Number(o.application_recipe_id) === Number(recipeId)) || null
}

/**
 * @param {Array<Record<string, unknown>>} outcomes
 * @param {number} recipeId
 */
export function bestOutcomeForRecipe(outcomes, recipeId) {
  if (!Array.isArray(outcomes) || !recipeId) return null
  const rows = outcomes.filter((o) => Number(o.application_recipe_id) === Number(recipeId))
  if (!rows.length) return null
  return rows.reduce((a, b) => ((a.cycle_count || 0) >= (b.cycle_count || 0) ? a : b))
}

/**
 * @param {Record<string, unknown>|null|undefined} outcome
 * @param {{ showCosts?: boolean }} [opts]
 */
export function formatRecipeTrackRecord(outcome, { showCosts = true } = {}) {
  if (!outcome || Number(outcome.cycle_count) < RECIPE_OUTCOME_MIN_SAMPLE) return ''
  const parts = [`Used in ${outcome.cycle_count} harvested cycles`]
  if (outcome.avg_yield_grams != null) {
    parts.push(`avg ${Math.round(Number(outcome.avg_yield_grams))}g`)
  }
  if (showCosts && outcome.avg_cost_per_gram != null && outcome.cost_currency) {
    parts.push(`${Number(outcome.avg_cost_per_gram).toFixed(2)} ${outcome.cost_currency}/g`)
  }
  return parts.join(' · ')
}

/**
 * @param {Record<string, unknown>|null|undefined} outcome
 * @param {{ showCosts?: boolean, recipeLabel?: string, revisionId?: number|null }} [opts]
 */
export function formatCycleRecipeTrackRecord(outcome, { showCosts = true, recipeLabel = '', revisionId = null } = {}) {
  if (!outcome || Number(outcome.cycle_count) < RECIPE_OUTCOME_MIN_SAMPLE) return ''
  const name = recipeLabel || outcome.recipe_name || `recipe #${outcome.application_recipe_id}`
  const rev = revisionId ?? outcome.application_recipe_revision_id
  const revBit = rev != null ? ` rev #${rev}` : ''
  let line = `This grow's recipe (${name}${revBit}) averaged ${Math.round(Number(outcome.avg_yield_grams))}g across ${outcome.cycle_count} harvested cycles`
  if (showCosts && outcome.avg_cost_per_gram != null && outcome.cost_currency) {
    line += ` · avg ${Number(outcome.avg_cost_per_gram).toFixed(2)} ${outcome.cost_currency}/g`
  }
  line += ' — historical average, not a forecast.'
  return line
}

function numOrNull(v) {
  if (v == null || v === '') return null
  const n = Number(v)
  return Number.isFinite(n) && n > 0 ? n : null
}
