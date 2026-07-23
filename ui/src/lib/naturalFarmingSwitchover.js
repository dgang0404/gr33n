/**
 * Phase 209 WS2 — switchover wizard logic (commercial EC → natural recipes).
 */
import { findApplicationRecipe, findCanonInput } from './naturalFarmingCanon.js'
import { BOOTSTRAP_TEMPLATE_KEYS } from './bootstrapCatalog.fallback.js'

export const SWITCHOVER_STEPS = ['context', 'pattern', 'mapping', 'first-batch', 'actions']

export const CONTEXT_OPTIONS = [
  {
    id: 'indoor_soil',
    label: 'Indoor soil beds',
    hint: 'Containers or raised beds indoors — bottle nutrients, dry salts, or organic liquids on soil',
  },
  {
    id: 'outdoor',
    label: 'Outdoor beds',
    hint: 'Garden beds, orchard edge, or field rows',
  },
  {
    id: 'indoor',
    label: 'Indoor hydro',
    hint: 'Tent, rack, or recirculating reservoir — EC-tuned bottles or dry salts',
  },
  {
    id: 'greenhouse',
    label: 'Greenhouse',
    hint: 'Glazed or poly house — commercial EC, soil, or organic bottles',
  },
  {
    id: 'livestock',
    label: 'Livestock supplement',
    hint: 'Comfrey, sprouted grain, or pasture edge — not a full ration calculator',
  },
]

/** Maps wizard choice → commercial_to_natural.commercial key in recipe-canonical.yaml */
export const COMMERCIAL_PATTERNS = [
  {
    id: 'single_part_ec',
    label: 'Single-part EC feed',
    hint: 'One bottle tuned to 1.6–1.8 mS/cm veg',
    commercialKey: 'Daily EC veg feed 1.6–1.8 mS/cm',
  },
  {
    id: 'ab_two_part',
    label: 'A+B flower boost',
    hint: 'Separate grow and bloom bottles',
    commercialKey: 'Flower boost A+B',
  },
  {
    id: 'dry_salts',
    label: 'Dry salts / cloner feed',
    hint: 'Light EC for seedlings or clones',
    commercialKey: 'Seedling/cloner light feed',
  },
  {
    id: 'organic_bottled',
    label: 'Organic bottled nutrients',
    hint: 'Liquid organic lines — start with soil biology + gentle drenches',
    commercialKey: 'Daily EC veg feed 1.6–1.8 mS/cm',
  },
]

export const FIRST_BATCH_SEED_NAMES = [
  'JMS (JADAM Microbial Solution)',
  'JLF General (Weed and Grass)',
]

/** Phase 211 WS2 — switchover pack keys from data/natural-farming-packs/switchover-packs.yaml */
export const SWITCHOVER_PACK_KEYS = {
  MERICLE_VEG_TO_JLF_V1: 'mericle_veg_to_jlf_v1',
  MERICLE_FLOWER_TO_FFJ_V1: 'mericle_flower_to_ffj_v1',
  LIVESTOCK_COMFREY_FEED_V1: 'livestock_comfrey_feed_v1',
}

/**
 * Maps wizard commercial pattern → apply-pack key (null = no dedicated pack).
 * @param {string} patternId
 */
export function switchoverPackKeyForPattern(patternId) {
  const pattern = COMMERCIAL_PATTERNS.find((p) => p.id === patternId)
  if (!pattern) return null
  switch (pattern.commercialKey) {
    case 'Daily EC veg feed 1.6–1.8 mS/cm':
      return SWITCHOVER_PACK_KEYS.MERICLE_VEG_TO_JLF_V1
    case 'Flower boost A+B':
      return SWITCHOVER_PACK_KEYS.MERICLE_FLOWER_TO_FFJ_V1
    default:
      return null
  }
}

/**
 * @param {string} guideFile e.g. natural-farming-jms.md
 */
export function fieldGuideDocPath(guideFile) {
  const base = String(guideFile || '').trim()
  if (!base) return ''
  return base.endsWith('.md') ? `field-guides/${base}` : `field-guides/${base}.md`
}

/**
 * @param {string} guideFile
 */
export function fieldGuideLearnRoute(guideFile) {
  const cited = fieldGuideDocPath(guideFile)
  if (!cited) return { path: '/operator-guide', query: { tab: 'knowledge' } }
  return { path: '/operator-guide', query: { tab: 'knowledge', cited_doc: cited } }
}

/**
 * @param {string} contextId
 */
export function bootstrapTemplateForContext(contextId) {
  if (contextId === 'livestock') return BOOTSTRAP_TEMPLATE_KEYS.CHICKEN_COOP_V1
  if (contextId === 'greenhouse') return BOOTSTRAP_TEMPLATE_KEYS.GREENHOUSE_CLIMATE_V1
  return BOOTSTRAP_TEMPLATE_KEYS.JADAM_INDOOR_PHOTOPERIOD_V1
}

/**
 * @param {Record<string, unknown>} canon
 */
export function firstBatchSuggestions(canon) {
  return FIRST_BATCH_SEED_NAMES.map((name) => findCanonInput(name, canon)).filter(Boolean)
}

/**
 * @param {string} contextId
 * @param {string} patternId
 * @param {Record<string, unknown>} canon
 */
export function resolveSwitchoverMapping(contextId, patternId, canon) {
  if (contextId === 'livestock') {
    return {
      commercialLabel: 'Bag feed / premix supplement',
      naturalEquivalent: [
        {
          recipe: 'Plant-based livestock supplement',
          frequency: 'Comfrey slurry or sprouted grain — see livestock primer',
          dilution: null,
          guide: 'natural-farming-livestock-plant-feed.md',
        },
      ],
      summaryGuide: 'natural-farming-livestock-plant-feed.md',
      bootstrapTemplate: bootstrapTemplateForContext(contextId),
      switchoverPackKey: SWITCHOVER_PACK_KEYS.LIVESTOCK_COMFREY_FEED_V1,
    }
  }

  const pattern = COMMERCIAL_PATTERNS.find((p) => p.id === patternId)
  const mappings = /** @type {Array<Record<string, unknown>>} */ (canon?.commercial_to_natural ?? [])
  const hit = mappings.find((m) => m.commercial === pattern?.commercialKey)

  const naturalEquivalent = (hit?.natural_equivalent ?? []).map((row) => {
    const recipeName = row.recipe
    const app = findApplicationRecipe(recipeName, canon)
    return {
      recipe: recipeName,
      frequency: row.frequency ?? '',
      dilution: app?.dilution ?? null,
      guide: app?.guide ?? 'natural-farming-application-recipes.md',
    }
  })

  return {
    commercialLabel: hit?.commercial ?? pattern?.commercialKey ?? 'Your commercial program',
    naturalEquivalent,
    summaryGuide: 'natural-farming-application-recipes.md',
    bootstrapTemplate: bootstrapTemplateForContext(contextId),
    switchoverPackKey: switchoverPackKeyForPattern(patternId),
  }
}

/**
 * Field guide for a canon input row (first-batch step).
 * @param {Record<string, unknown> | null | undefined} input
 */
export function learnGuideForInput(input) {
  const pt = String(input?.process_type ?? '').toLowerCase()
  if (pt === 'jlf') return 'natural-farming-jlf-general.md'
  if (pt === 'jms') return 'natural-farming-jms.md'
  if (pt === 'ffj') return 'natural-farming-ffj.md'
  return 'natural-farming-application-recipes.md'
}

/**
 * Field guide for commercial bottle pattern (wizard step 2).
 * @param {string} patternId
 */
export function learnGuideForPattern(patternId) {
  switch (patternId) {
    case 'ab_two_part':
      return 'natural-farming-ffj.md'
    case 'organic_bottled':
    case 'single_part_ec':
      return 'natural-farming-jlf-general.md'
    case 'dry_salts':
      return 'natural-farming-jms.md'
    default:
      return 'natural-farming-application-recipes.md'
  }
}

/**
 * Step 5 intro — clarify DB bootstrap vs tank dosing.
 * @param {string} contextId
 */
export function actionsStepIntro(contextId) {
  if (contextId === 'livestock') {
    return 'Ferment or import feed inputs, or apply the livestock starter pack — it wires animal groups and supplement recipes on this farm (not a physical dose).'
  }
  if (contextId === 'indoor_soil' || contextId === 'outdoor') {
    return 'Ferment your first concentrates, or apply a starter pack that seeds zones, soil-drench programs, and audited recipes on this farm. Starter apply does not pump a tank — use Feed & water → Programs (and your Pi mix plan) for ml dosing into a reservoir.'
  }
  if (contextId === 'greenhouse') {
    return 'Ferment concentrates or apply the greenhouse climate starter pack — zones, programs, and recipes on this farm. Dosing runs through Feed & water after batches are ready.'
  }
  return 'Ferment concentrates or apply the indoor photoperiod starter pack — zones, programs, and recipes on this farm. Dosing runs through Feed & water after batches are ready.'
}

/**
 * @param {string} contextId
 */
export function bootstrapApplyButtonLabel(contextId) {
  if (contextId === 'livestock') return 'Apply livestock starter to farm'
  if (contextId === 'greenhouse') return 'Apply greenhouse starter to farm'
  if (contextId === 'indoor_soil' || contextId === 'outdoor') return 'Apply JADAM starter to farm'
  return 'Apply indoor photoperiod starter to farm'
}

/**
 * @param {string} stepId
 * @param {string} contextId
 * @param {string} [patternId]
 */
export function learnGuideForStep(stepId, contextId, patternId = '') {
  switch (stepId) {
    case 'context':
      if (contextId === 'livestock') return 'natural-farming-livestock-plant-feed.md'
      if (contextId === 'outdoor' || contextId === 'indoor_soil') return 'natural-farming-jlf-general.md'
      if (contextId === 'indoor' || contextId === 'greenhouse') {
        return 'natural-farming-indoor-photoperiod-program.md'
      }
      return 'natural-farming-application-recipes.md'
    case 'pattern':
      return patternId ? learnGuideForPattern(patternId) : 'natural-farming-application-recipes.md'
    case 'mapping':
      return 'natural-farming-application-recipes.md'
    case 'first-batch':
      return 'natural-farming-jms.md'
    case 'actions':
      if (contextId === 'outdoor' || contextId === 'indoor_soil') return 'natural-farming-jlf-general.md'
      return 'natural-farming-indoor-photoperiod-program.md'
    default:
      return 'natural-farming-application-recipes.md'
  }
}

/**
 * @param {Record<string, unknown>} input
 */
export function batchTabQueryForInput(input) {
  const process = String(input?.process_type ?? '').trim()
  return process ? { tab: 'batch', process } : { tab: 'batch' }
}
