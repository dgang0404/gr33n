/**
 * Phase 41 WS4 — reusable why-empty copy for farm + zone surfaces.
 */

export const EMPTY_HINT_REASONS = Object.freeze({
  no_data: 'no_data',
  no_telemetry: 'no_telemetry',
  no_setpoint: 'no_setpoint',
  automation_off: 'automation_off',
  wrong_farm: 'wrong_farm',
})

/** @typedef {'no_data'|'no_telemetry'|'no_setpoint'|'automation_off'|'wrong_farm'} EmptyHintReason */

const DEFAULTS = {
  no_data: {
    message: 'Nothing recorded yet for this farm.',
    actionLabel: 'Set up zones',
    actionTo: '/zones',
  },
  no_telemetry: {
    message: 'No recent readings — check that your edge device is online and posting to the API.',
    actionLabel: 'Pi integration guide',
    actionTo: '/operator-guide',
  },
  no_setpoint: {
    message: 'No target band for this sensor type yet.',
    actionLabel: 'Comfort targets',
    actionTo: null,
  },
  automation_off: {
    message: 'Rules or schedules exist but nothing is active right now.',
    actionLabel: 'Schedules',
    actionTo: '/schedules',
  },
  wrong_farm: {
    message: 'This list is empty for the farm selected in the header — try another farm.',
    actionLabel: null,
    actionTo: null,
  },
}

/**
 * @param {EmptyHintReason|string} reason
 * @param {{ message?: string, actionLabel?: string|null, actionTo?: string|null }} [overrides]
 */
export function emptyHintConfig(reason, overrides = {}) {
  const base = DEFAULTS[reason] || DEFAULTS.no_data
  return {
    reason,
    message: overrides.message ?? base.message,
    actionLabel: overrides.actionLabel !== undefined ? overrides.actionLabel : base.actionLabel,
    actionTo: overrides.actionTo !== undefined ? overrides.actionTo : base.actionTo,
  }
}
