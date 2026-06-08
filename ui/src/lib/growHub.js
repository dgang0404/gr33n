/**
 * Phase 53 WS1 — grow cycle helpers (start, harvest, compare).
 */

export const GROWTH_STAGES = [
  'clone',
  'seedling',
  'early_veg',
  'late_veg',
  'transition',
  'early_flower',
  'mid_flower',
  'late_flower',
  'flush',
  'harvest',
  'dry_cure',
]

/**
 * @param {string|null|undefined} stage
 */
export function formatStageLabel(stage) {
  if (!stage) return '—'
  return String(stage).replace(/_/g, ' ')
}

/**
 * @param {string|object} raw
 */
export function parseCycleDate(raw) {
  if (!raw) return null
  let s = ''
  if (typeof raw === 'string') s = raw.slice(0, 10)
  else if (raw.Time) s = String(raw.Time).slice(0, 10)
  else return null
  const parts = s.split('-').map(Number)
  if (parts.length < 3 || !parts[0] || !parts[1] || !parts[2]) return null
  const d = new Date(parts[0], parts[1] - 1, parts[2])
  return Number.isNaN(d.getTime()) ? null : d
}

/**
 * @param {object} cycle
 * @param {Date} [referenceDate]
 */
export function daysSinceStart(cycle, referenceDate = new Date()) {
  const started = parseCycleDate(cycle?.started_at)
  if (!started) return null
  const ms = referenceDate.getTime() - started.getTime()
  return Math.max(0, Math.floor(ms / (24 * 60 * 60 * 1000)))
}

/**
 * @param {object[]} cycles
 * @param {number} zoneId
 */
export function activeCycleForZone(cycles, zoneId) {
  return (cycles || []).find(
    (c) => c.is_active && Number(c.zone_id) === Number(zoneId),
  ) || null
}

/**
 * Most recent harvested (inactive) cycle in the same zone.
 * @param {object[]} cycles
 * @param {number} zoneId
 * @param {number} [excludeId]
 */
export function lastHarvestedCycleInZone(cycles, zoneId, excludeId = null) {
  return (cycles || [])
    .filter((c) => {
      if (!c || c.is_active) return false
      if (Number(c.zone_id) !== Number(zoneId)) return false
      if (excludeId != null && Number(c.id) === Number(excludeId)) return false
      return true
    })
    .sort((a, b) => Number(b.id) - Number(a.id))[0] || null
}

/**
 * @param {string} strain
 * @param {string} [zoneName]
 */
export function defaultCycleName(strain, zoneName = '') {
  const base = (strain || 'Grow').trim() || 'Grow'
  if (!zoneName) return base
  return `${base} — ${zoneName}`
}

/**
 * @param {object} plant
 */
/**
 * @param {object|null} stageRow crop_profile_stages row
 */
export function formatEcTargetChip(stageRow) {
  if (!stageRow) return null
  const fmt = (n) => {
    if (n == null || n === '') return null
    const v = Number(n)
    return Number.isFinite(v) ? (v % 1 === 0 ? String(v) : v.toFixed(2)) : null
  }
  const min = fmt(stageRow.ec_min)
  const max = fmt(stageRow.ec_max)
  const target = fmt(stageRow.ec_target)
  if (min && max) return `EC target ${min}–${max} mS/cm`
  if (target) return `EC target ${target} mS/cm`
  return null
}

export function strainFromPlant(plant) {
  if (!plant) return ''
  const variety = plant.variety_or_cultivar?.trim()
  const name = plant.display_name?.trim() || ''
  if (variety && name) return `${name} (${variety})`
  return name || variety || ''
}

/**
 * @param {object} params
 * @param {number} params.zoneId
 * @param {string} params.strain
 * @param {string} [params.name]
 * @param {string} [params.stage]
 * @param {string} [params.startedAt]
 * @param {number|null} [params.programId]
 * @param {string} [params.notes]
 */
export function buildStartGrowPayload({
  zoneId,
  strain,
  name,
  stage = 'seedling',
  startedAt,
  programId = null,
  plantId = null,
  notes = '',
}) {
  const today = new Date().toISOString().slice(0, 10)
  const payload = {
    zone_id: Number(zoneId),
    name: (name || strain || 'Grow').trim(),
    strain_or_variety: strain?.trim() || undefined,
    current_stage: stage || 'seedling',
    started_at: startedAt || today,
    is_active: true,
  }
  if (programId) payload.primary_program_id = Number(programId)
  if (plantId) payload.plant_id = Number(plantId)
  if (notes?.trim()) payload.cycle_notes = notes.trim()
  return payload
}

/**
 * @param {object} cycle
 * @param {{ yieldGrams?: number|null, yieldNotes?: string, harvestedAt?: string }} params
 */
export function buildHarvestPayload(cycle, { yieldGrams, yieldNotes = '', harvestedAt } = {}) {
  const today = new Date().toISOString().slice(0, 10)
  const payload = {
    name: cycle.name,
    zone_id: Number(cycle.zone_id),
    is_active: false,
    cycle_notes: cycle.cycle_notes || undefined,
    harvested_at: harvestedAt || today,
    primary_program_id: cycle.primary_program_id ?? undefined,
    strain_or_variety: cycle.strain_or_variety || undefined,
  }
  if (yieldGrams != null && yieldGrams !== '' && !Number.isNaN(Number(yieldGrams))) {
    payload.yield_grams = Number(yieldGrams)
  }
  if (yieldNotes?.trim()) payload.yield_notes = yieldNotes.trim()
  return payload
}

/**
 * @param {number} farmId
 * @param {number[]} cycleIds
 */
export function buildCompareRoute(farmId, cycleIds) {
  const ids = (cycleIds || []).filter((id) => id != null).map((id) => Number(id))
  if (!farmId || !ids.length) return { path: '/fertigation' }
  return {
    path: `/farms/${farmId}/crop-cycles/compare`,
    query: { ids: ids.join(',') },
  }
}

/**
 * Compare pair: current harvested cycle + prior harvested in same zone.
 * @param {number} farmId
 * @param {object[]} cycles
 * @param {number} currentCycleId
 * @param {number} zoneId
 */
export function buildPostHarvestCompareRoute(farmId, cycles, currentCycleId, zoneId) {
  const prior = lastHarvestedCycleInZone(cycles, zoneId, currentCycleId)
  const ids = prior ? [currentCycleId, prior.id] : [currentCycleId]
  return buildCompareRoute(farmId, ids)
}
