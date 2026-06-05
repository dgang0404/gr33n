/**
 * Phase 47 WS3 — draft + save helpers for inline feeding plan editor.
 */

import { buildDailyFeedCron, defaultDailyFeedTime } from './dailyFeedSchedule.js'
import { resolveEcRange } from './zoneFeedingPlan.js'

/**
 * @param {object} params
 */
export function buildFeedingPlanDraft({
  activeProgram,
  schedule,
  ecTargets = [],
}) {
  const ec = resolveEcRange(activeProgram, ecTargets)
  return {
    volumeLiters: activeProgram?.total_volume_liters ?? 0,
    dailyFeedTime: defaultDailyFeedTime(schedule),
    irrigationOnly: Boolean(activeProgram?.irrigation_only),
    feedingPaused: schedule ? schedule.is_active === false : false,
    ecMin: ec?.min ?? null,
    ecMax: ec?.max ?? null,
    ecFromTarget: Boolean(activeProgram?.ec_target_id),
  }
}

/**
 * @param {object} program
 * @param {object} draft
 */
export function buildProgramPatch(program, draft) {
  return {
    name: program.name,
    description: program.description ?? null,
    reservoir_id: program.reservoir_id ?? null,
    target_zone_id: program.target_zone_id ?? null,
    ec_target_id: program.ec_target_id ?? null,
    application_recipe_id: program.application_recipe_id ?? null,
    total_volume_liters: Number(draft.volumeLiters) || 0,
    is_active: program.is_active !== false,
    irrigation_only: Boolean(draft.irrigationOnly),
  }
}

/**
 * @param {object} schedule
 * @param {object} draft
 * @param {string} cronExpression
 */
export function buildSchedulePatch(schedule, draft, cronExpression) {
  return {
    name: schedule.name,
    description: schedule.description || '',
    schedule_type: schedule.schedule_type || 'cron',
    cron_expression: cronExpression,
    timezone: schedule.timezone || 'UTC',
    is_active: !draft.feedingPaused,
    preconditions: schedule.preconditions || [],
  }
}

/**
 * @param {object} params
 */
export function buildWizardSchedulePayload({
  zoneName,
  farmTimezone = 'UTC',
  dailyFeedTime,
}) {
  return {
    name: `${zoneName} daily feed`,
    description: `Daily feeding for ${zoneName}`,
    schedule_type: 'cron',
    cron_expression: buildDailyFeedCron(dailyFeedTime),
    timezone: farmTimezone,
    is_active: true,
    preconditions: [],
  }
}

/**
 * @param {object} params
 */
export function buildWizardProgramPayload({
  name,
  zoneId,
  scheduleId,
  reservoirId,
  ecTargetId,
  volumeLiters,
  irrigationOnly,
}) {
  return {
    name,
    target_zone_id: zoneId,
    schedule_id: scheduleId,
    reservoir_id: reservoirId ?? null,
    ec_target_id: ecTargetId ?? null,
    application_recipe_id: null,
    total_volume_liters: Number(volumeLiters) || 0,
    ec_trigger_low: 0,
    ph_trigger_low: 5.5,
    ph_trigger_high: 6.5,
    is_active: true,
    irrigation_only: Boolean(irrigationOnly),
  }
}

/**
 * @param {object} params
 */
export function buildWizardEcTargetPayload({
  zoneId,
  ecMin,
  ecMax,
  growthStage = 'vegetative',
}) {
  return {
    growth_stage: growthStage,
    zone_id: zoneId,
    ec_min_mscm: Number(ecMin),
    ec_max_mscm: Number(ecMax),
    ph_min: 5.5,
    ph_max: 6.5,
    notes: 'Created from zone feeding plan wizard',
  }
}

/**
 * Pick a reservoir for a new plan in this room.
 * @param {object[]} reservoirs
 * @param {number} zoneId
 */
export function pickReservoirForZone(reservoirs, zoneId) {
  const forZone = (reservoirs || []).filter((r) => Number(r.zone_id) === Number(zoneId))
  if (forZone.length) return forZone[0]
  return (reservoirs || [])[0] || null
}
