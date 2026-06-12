/**
 * Phase 88 — platform domain enums from GET /platform/domain-enums.
 */
import { FALLBACK_DOMAIN_ENUMS } from './domainEnums.fallback.js'

/** @typedef {{ value: string, label: string }} EnumOption */

let cached = null
/** @type {Promise<object>|null} */
let loadPromise = null

/**
 * @param {object|null|undefined} payload
 */
export function normalizeDomainEnums(payload) {
  const src = payload || FALLBACK_DOMAIN_ENUMS
  return {
    growth_stages: src.growth_stages ?? FALLBACK_DOMAIN_ENUMS.growth_stages,
    reservoir_statuses: src.reservoir_statuses ?? FALLBACK_DOMAIN_ENUMS.reservoir_statuses,
    cost_categories: src.cost_categories ?? FALLBACK_DOMAIN_ENUMS.cost_categories,
    application_targets: src.application_targets ?? FALLBACK_DOMAIN_ENUMS.application_targets,
    input_definition_categories: src.input_definition_categories ?? FALLBACK_DOMAIN_ENUMS.input_definition_categories,
    batch_statuses: src.batch_statuses ?? FALLBACK_DOMAIN_ENUMS.batch_statuses,
    zone_types: src.zone_types ?? FALLBACK_DOMAIN_ENUMS.zone_types,
    greenhouse_cover_types: src.greenhouse_cover_types ?? FALLBACK_DOMAIN_ENUMS.greenhouse_cover_types,
    greenhouse_automation_policies: src.greenhouse_automation_policies ?? FALLBACK_DOMAIN_ENUMS.greenhouse_automation_policies,
  }
}

/**
 * @param {{ get: (url: string) => Promise<{ data: object }> }} api
 */
export async function loadDomainEnums(api) {
  if (cached) return cached
  if (loadPromise) return loadPromise
  loadPromise = api
    .get('/platform/domain-enums')
    .then(({ data }) => {
      cached = normalizeDomainEnums(data)
      return cached
    })
    .catch(() => {
      cached = normalizeDomainEnums(null)
      return cached
    })
    .finally(() => {
      loadPromise = null
    })
  return loadPromise
}

/** @returns {object} */
export function getDomainEnums() {
  return cached || FALLBACK_DOMAIN_ENUMS
}

/** @param {object|null|undefined} enums @param {string} key */
export function enumValues(enums, key) {
  const rows = (enums || getDomainEnums())[key] || []
  return rows.map((r) => r.value)
}

/** @param {object|null|undefined} enums */
export function growthStageValues(enums) {
  return enumValues(enums, 'growth_stages')
}

/** @param {string} listKey @param {string} value @param {object|null|undefined} enums */
export function enumLabel(listKey, value, enums) {
  if (!value) return ''
  const rows = (enums || getDomainEnums())[listKey] || []
  const hit = rows.find((r) => r.value === value)
  return hit?.label || String(value).replace(/_/g, ' ')
}

/** @param {object|null|undefined} enums */
export function adminZoneTypes(enums) {
  return (enums || getDomainEnums()).zone_types || []
}

/** @param {object|null|undefined} enums */
export function wizardZoneTypes(enums) {
  return adminZoneTypes(enums)
    .filter((r) => r.wizard_visible)
    .map((r) => ({ value: r.value, label: r.label, hint: r.hint || '' }))
}

/** @param {object|null|undefined} enums */
export function greenhouseCoverTypes(enums) {
  return (enums || getDomainEnums()).greenhouse_cover_types || []
}

/** @param {object|null|undefined} enums */
export function greenhouseAutomationPolicies(enums) {
  return (enums || getDomainEnums()).greenhouse_automation_policies || []
}

/** @param {string} value @param {object|null|undefined} enums */
export function zoneTypeLabel(value, enums) {
  return enumLabel('zone_types', value, enums)
}

/** Test helper */
export function resetDomainEnumsCache() {
  cached = null
  loadPromise = null
}
