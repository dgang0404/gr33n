/**
 * Phase 44 / 91 — farm setup wizard helpers (template cards, preview, apply result).
 */

import { getBootstrapCatalog } from './bootstrapCatalog.js'
import { BOOTSTRAP_TEMPLATE_KEYS } from './bootstrapCatalog.fallback.js'

export const FARM_SETUP_BLANK_ID = 'blank'

/** Primary wizard cards (blank + wizard_primary templates from catalog). */
export function farmSetupPrimaryChoices() {
  const { templates } = getBootstrapCatalog()
  const blank = {
    id: FARM_SETUP_BLANK_ID,
    label: 'Start blank',
    tagline: 'Empty farm — add zones and devices yourself',
    icon: '📋',
    recommended: false,
  }
  const cards = templates
    .filter((t) => t.wizard_primary)
    .map((t) => ({
      id: t.template_key,
      label: t.short_label || t.label,
      tagline: t.tagline || t.label,
      icon: t.icon || '📦',
      recommended: !!t.recommended,
    }))
  return [blank, ...cards]
}

/** @deprecated prefer farmSetupPrimaryChoices() after loadBootstrapCatalog */
export const FARM_SETUP_PRIMARY_CHOICES = farmSetupPrimaryChoices()

/** Additional templates shown under “More starter packs”. */
export function farmSetupMoreChoices() {
  const primary = new Set(farmSetupPrimaryChoices().map((c) => c.id))
  return getBootstrapCatalog()
    .starterOptions.filter((opt) => !primary.has(opt.value))
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
  const summary = getBootstrapCatalog().summariesByKey[choiceId]
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
      message: 'Starter pack applied. Open My zones, Feed & water, and Tasks to explore what was created.',
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

export { BOOTSTRAP_TEMPLATE_KEYS }
