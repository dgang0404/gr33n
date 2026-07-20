/**
 * Phase 211 WS4 — Commons natural-farming recipe pack import helpers.
 */

export const NF_COMMONS_TAG = 'natural_farming'
export const NF_RECIPE_PACK_KIND = 'natural_farming_recipe_pack'

/**
 * @param {Record<string, unknown> | null | undefined} entry
 */
export function isNaturalFarmingCatalogEntry(entry) {
  const tags = /** @type {string[]} */ (entry?.tags ?? [])
  return tags.includes(NF_COMMONS_TAG)
}

/**
 * @param {unknown} body
 */
export function parseNaturalFarmingPackBody(body) {
  if (!body) return null
  let b = body
  if (typeof body === 'string') {
    try {
      b = JSON.parse(body)
    } catch {
      return null
    }
  }
  if (typeof b !== 'object' || b === null) return null
  const kind = /** @type {Record<string, unknown>} */ (b).kind
  if (kind !== NF_RECIPE_PACK_KIND) return null
  const pack = /** @type {Record<string, unknown>} */ (b)
  const inputs = /** @type {Array<Record<string, unknown>>} */ (pack.input_definitions ?? [])
  const recipes = /** @type {Array<Record<string, unknown>>} */ (pack.application_recipes ?? [])
  const components = /** @type {unknown[]} */ (pack.recipe_input_components ?? [])
  return {
    kind: NF_RECIPE_PACK_KIND,
    packKey: String(pack.pack_key ?? ''),
    inputCount: inputs.length,
    recipeCount: recipes.length,
    componentCount: components.length,
    inputNames: inputs.map((i) => String(i.name ?? '')).filter(Boolean),
    recipeNames: recipes.map((r) => String(r.name ?? '')).filter(Boolean),
    readme: String(pack.readme_md ?? ''),
  }
}

/**
 * Human-readable import result for Commons recipe pack apply.
 * @param {{ apply?: Record<string, unknown>, error?: string }} out
 */
export function formatCommonsImportMessage(out) {
  if (out?.error) return String(out.error)
  const apply = out?.apply
  if (!apply) return 'Recipe pack imported.'
  const status = String(apply.status ?? '')
  const base = String(apply.message ?? 'Recipe pack imported.')
  if (status === 'noop') {
    const skippedIn = apply.inputs_skipped ?? 0
    const skippedRec = apply.recipes_skipped ?? 0
    return `${base} (${skippedIn} inputs, ${skippedRec} recipes already on farm — safe to click again.)`
  }
  const created = []
  if (apply.inputs_created) created.push(`${apply.inputs_created} inputs`)
  if (apply.recipes_created) created.push(`${apply.recipes_created} recipes`)
  if (apply.components_upserted) created.push(`${apply.components_upserted} components`)
  if (!created.length) return base
  return `${base} Added ${created.join(', ')}.`
}

/**
 * @param {{ inputNames?: string[] } | null | undefined} preview
 */
export function firstBatchQueryForPack(preview) {
  const names = preview?.inputNames ?? []
  if (names.some((n) => /\bJMS\b/i.test(n))) return { tab: 'batch', process: 'jms' }
  if (names.some((n) => /\bJLF\b/i.test(n))) return { tab: 'batch', process: 'jlf' }
  if (names.some((n) => /comfrey/i.test(n))) return { tab: 'batch', process: 'animal_feed' }
  return { tab: 'batch' }
}
