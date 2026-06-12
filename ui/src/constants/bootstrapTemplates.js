/**
 * @deprecated Phase 91 — prefer `ui/src/lib/bootstrapCatalog.js`. Re-exports for backward compatibility.
 */
export {
  BOOTSTRAP_TEMPLATE_KEYS,
  FALLBACK_TEMPLATES,
} from '../lib/bootstrapCatalog.fallback.js'

export {
  getBootstrapCatalog,
  loadBootstrapCatalog,
} from '../lib/bootstrapCatalog.js'

import { getBootstrapCatalog } from '../lib/bootstrapCatalog.js'

/** Farm-create / org-default picker options. */
export function bootstrapStarterOptions() {
  return getBootstrapCatalog().starterOptions
}

/** Map template key → { title, bullets } for expandable help on farm create / apply-starter UI. */
export function bootstrapStarterSummaries() {
  return getBootstrapCatalog().summariesByKey
}

/** @deprecated use bootstrapStarterOptions() */
export const BOOTSTRAP_STARTER_OPTIONS = getBootstrapCatalog().starterOptions

/** @deprecated use bootstrapStarterSummaries() */
export const BOOTSTRAP_STARTER_SUMMARIES = getBootstrapCatalog().summariesByKey

const summaries = getBootstrapCatalog().summariesByKey

export const JADAM_INDOOR_PHOTOPERIOD_V1_SUMMARY = summaries.jadam_indoor_photoperiod_v1
export const CHICKEN_COOP_V1_SUMMARY = summaries.chicken_coop_v1
export const GREENHOUSE_CLIMATE_V1_SUMMARY = summaries.greenhouse_climate_v1
export const DRYING_ROOM_V1_SUMMARY = summaries.drying_room_v1
export const SMALL_AQUAPONICS_V1_SUMMARY = summaries.small_aquaponics_v1
