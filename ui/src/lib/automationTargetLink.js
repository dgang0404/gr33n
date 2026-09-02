/**
 * Phase 217 WS2 — resolve a schedule/automation_rule to the zone/reservoir it
 * actually touches, entirely client-side from data already in the farm store
 * (no new API calls). Neither gr33ncore.schedules nor gr33ncore.automation_rules
 * carry a direct zone/actuator/reservoir reference — they reach it through
 * executable_actions.target_actuator_id (-> actuators.zone_id) or, for
 * lighting, through lighting_programs (which has zone_id/actuator_id directly).
 *
 * Returns `{ to, label }` for a resolvable target (`to` is a vue-router
 * location, `label` is what kind of resource it is — for the v-nav-hint
 * sidebar path), or `null` when nothing resolves — callers render plain text
 * in that case, never a dead link.
 */
import { feedWaterFertigationRoute } from './workspaceRoutes.js'

/**
 * @param {object} schedule
 * @param {object} ctx
 * @param {object[]} [ctx.programs] - fertigation programs (schedule_id, target_zone_id, reservoir_id)
 * @param {object[]} [ctx.lightingPrograms] - lighting_programs (schedule_on_id, schedule_off_id, zone_id)
 */
export function resolveScheduleTarget(schedule, ctx = {}) {
  const programs = ctx.programs || []
  const lightingPrograms = ctx.lightingPrograms || []
  if (!schedule?.id) return null

  const program = programs.find((p) => Number(p.schedule_id) === Number(schedule.id))
  if (program?.target_zone_id) {
    return { to: { path: `/zones/${program.target_zone_id}` }, navHint: '/zones', label: 'zone' }
  }
  if (program?.reservoir_id) {
    return { to: feedWaterFertigationRoute('reservoirs'), navHint: '/feed-water', label: 'reservoir' }
  }

  const lighting = lightingPrograms.find(
    (p) => Number(p.schedule_on_id) === Number(schedule.id) || Number(p.schedule_off_id) === Number(schedule.id),
  )
  if (lighting?.zone_id) {
    return { to: { path: `/zones/${lighting.zone_id}` }, navHint: '/zones', label: 'zone' }
  }

  return null
}

/**
 * @param {object} rule
 * @param {object} ctx
 * @param {object[]} [ctx.actions] - executable_actions for this rule (target_actuator_id)
 * @param {object[]} [ctx.actuators] - farm actuators (id, zone_id)
 */
export function resolveRuleTarget(rule, ctx = {}) {
  const actions = ctx.actions || []
  const actuators = ctx.actuators || []
  if (!rule?.id) return null

  const action = actions.find((a) => a.action_type === 'control_actuator' && a.target_actuator_id)
  if (!action) return null

  const actuator = actuators.find((a) => Number(a.id) === Number(action.target_actuator_id))
  if (!actuator?.zone_id) return null

  return { to: { path: `/zones/${actuator.zone_id}` }, navHint: '/zones', label: 'zone' }
}
