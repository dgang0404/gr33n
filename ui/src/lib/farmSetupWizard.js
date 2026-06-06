/**
 * Phase 44 WS1 — farm setup wizard helpers (template cards, preview, apply result).
 */

import {
  BOOTSTRAP_TEMPLATE_KEYS,
  BOOTSTRAP_STARTER_OPTIONS,
  BOOTSTRAP_STARTER_SUMMARIES,
} from '../constants/bootstrapTemplates.js'

export const FARM_SETUP_BLANK_ID = 'blank'

/** Primary wizard cards (plan: blank + indoor veg + greenhouse). */
export const FARM_SETUP_PRIMARY_CHOICES = [
  {
    id: FARM_SETUP_BLANK_ID,
    label: 'Start blank',
    tagline: 'Empty farm — add rooms and devices yourself',
    icon: '📋',
    recommended: false,
  },
  {
    id: BOOTSTRAP_TEMPLATE_KEYS.JADAM_INDOOR_PHOTOPERIOD_V1,
    label: 'Indoor photoperiod',
    tagline: 'Four grow rooms, feeding programs, inventory, and demo tasks',
    icon: '🌱',
    recommended: true,
  },
  {
    id: BOOTSTRAP_TEMPLATE_KEYS.GREENHOUSE_CLIMATE_V1,
    label: 'Greenhouse climate',
    tagline: 'Shade, vents, humidity bands, and Pi placeholder',
    icon: '🏠',
    recommended: false,
  },
]

/** Additional templates shown under “More starter packs”. */
export function farmSetupMoreChoices() {
  const primary = new Set(FARM_SETUP_PRIMARY_CHOICES.map((c) => c.id))
  return BOOTSTRAP_STARTER_OPTIONS
    .filter((opt) => !primary.has(opt.value))
    .map((opt) => ({
      id: opt.value,
      label: opt.shortLabel || opt.label,
      tagline: opt.label,
    }))
}

/**
 * @param {string} choiceId
 * @returns {{ title: string, bullets: string[], isBlank: boolean }}
 */
export function previewForSetupChoice(choiceId) {
  if (!choiceId || choiceId === FARM_SETUP_BLANK_ID) {
    return {
      isBlank: true,
      title: 'Blank farm — nothing will be created automatically',
      bullets: [
        'No zones, schedules, or inventory until you add them',
        'Use Add zone wizard (Phase 44 WS2) or Zones → create',
        'Connect a Pi when you are ready (device wizard in WS3)',
      ],
    }
  }
  const summary = BOOTSTRAP_STARTER_SUMMARIES[choiceId]
  if (!summary) {
    return {
      isBlank: false,
      title: 'Starter pack preview',
      bullets: ['Zones, schedules, and starter config from the selected template'],
    }
  }
  return {
    isBlank: false,
    title: summary.title,
    bullets: summary.bullets,
  }
}

/**
 * @param {object|null|undefined} bootstrap — API bootstrap result
 */
export function formatBootstrapApplyResult(bootstrap) {
  if (!bootstrap || typeof bootstrap !== 'object') {
    return { ok: true, message: 'Done.' }
  }
  if (bootstrap.skipped) {
    return { ok: true, message: 'No template applied — your farm stays blank.' }
  }
  if (bootstrap.already_applied) {
    return {
      ok: true,
      message: 'This starter pack was already applied to this farm. Existing data was left unchanged.',
    }
  }
  if (bootstrap.error) {
    return { ok: false, message: String(bootstrap.error) }
  }
  if (bootstrap.applied) {
    return {
      ok: true,
      message: 'Starter pack applied. Open My rooms, Feed & water, and Tasks to explore what was created.',
    }
  }
  return { ok: true, message: 'Template step finished.' }
}

export function farmSetupRoute(farmId) {
  return `/farms/${farmId}/setup`
}

/** Same POST path the farm setup wizard uses (`farmContext.applyBootstrapTemplate`). */
export function farmBootstrapApplyPath(farmId) {
  return `/farms/${farmId}/bootstrap-template`
}

const SETUP_DONE_PREFIX = 'gr33n_farm_setup_done_'

export function markFarmSetupComplete(farmId) {
  if (typeof localStorage === 'undefined' || !farmId) return
  localStorage.setItem(`${SETUP_DONE_PREFIX}${farmId}`, '1')
}

export function isFarmSetupMarkedComplete(farmId) {
  if (typeof localStorage === 'undefined' || !farmId) return false
  return localStorage.getItem(`${SETUP_DONE_PREFIX}${farmId}`) === '1'
}
