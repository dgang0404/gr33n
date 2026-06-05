/**
 * Phase 42 WS4 — one-sentence automation rule summaries for farmer views.
 */

const OP_LABEL = { lt: 'below', lte: 'at or below', eq: 'at', gte: 'at or above', gt: 'above', ne: 'not' }

/**
 * @param {object} rule
 * @param {object} ctx
 * @param {object[]} [ctx.sensors]
 * @param {object[]} [ctx.actuators]
 * @param {object[]} [ctx.actions]
 */
export function ruleSummary(rule, ctx = {}) {
  const sensors = ctx.sensors || []
  const actuators = ctx.actuators || []
  const actions = ctx.actions || []

  const when = conditionPhrase(rule, sensors)
  const then = actionPhrase(actions, actuators)

  if (when && then) return `When ${when}, ${then}.`
  if (when) return `When ${when}, run automation.`
  if (then) return then.charAt(0).toUpperCase() + then.slice(1) + '.'
  return rule?.name || 'Automation rule'
}

function conditionPhrase(rule, sensors) {
  const conds = parsePredicates(rule)
  if (!conds.length) {
    const cfg = rule?.trigger_configuration || {}
    if (rule?.trigger_source === 'sensor_reading_threshold' && cfg.sensor_id) {
      return `${sensorName(sensors, cfg.sensor_id)} changes`
    }
    return ''
  }
  const joiner = rule?.condition_logic === 'ANY' ? ' or ' : ' and '
  return conds
    .map((p) => `${sensorName(sensors, p.sensor_id)} is ${OP_LABEL[p.op] || p.op} ${p.value}`)
    .join(joiner)
}

function actionPhrase(actions, actuators) {
  if (!actions.length) return 'run configured actions'
  const parts = actions.map((a) => {
    if (a.action_type === 'control_actuator' && a.target_actuator_id) {
      const name = actuatorName(actuators, a.target_actuator_id)
      const cmd = a.action_command ? ` ${a.action_command}` : ''
      return `${cmd.trim() ? cmd.trim() : 'run'} ${name}`
    }
    if (a.action_type === 'send_notification') return 'send an alert'
    if (a.action_type === 'create_task' && a.action_parameters?.title) {
      return `create task "${a.action_parameters.title}"`
    }
    return a.action_type?.replace(/_/g, ' ') || 'run action'
  })
  return parts.join(', then ')
}

function parsePredicates(rule) {
  const raw = rule?.conditions_jsonb
  const parsed = typeof raw === 'string'
    ? (() => { try { return JSON.parse(raw) } catch { return {} } })()
    : (raw || {})
  return Array.isArray(parsed.predicates) ? parsed.predicates : []
}

function sensorName(sensors, id) {
  return sensors.find((s) => Number(s.id) === Number(id))?.name
    || sensors.find((s) => Number(s.id) === Number(id))?.sensor_type
    || 'sensor'
}

function actuatorName(actuators, id) {
  return actuators.find((a) => Number(a.id) === Number(id))?.name || 'device'
}
