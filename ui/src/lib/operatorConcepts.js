/**
 * Farmer-facing glossary — each concept maps to real DB tables (not synonyms).
 * @see docs/schema-erd-text.md ops & automation section
 */

/** @typedef {{ label: string, dbTable: string, shortTip: string, detail: string }} OperatorConcept */

/** @type {Record<string, OperatorConcept>} */
export const OPERATOR_CONCEPTS = {
  task: {
    label: 'Task',
    dbTable: 'gr33ncore.tasks',
    shortTip: 'A to-do for a person — not automatic. Created manually, from an alert, or by a rule action.',
    detail: 'Tasks track work you do (inspect, refill, harvest). They never run hardware by themselves. Optional links: zone, schedule, crop cycle.',
  },
  alert: {
    label: 'Alert',
    dbTable: 'gr33ncore.alerts_notifications',
    shortTip: 'A warning when a reading leaves a threshold or something fails — not a to-do until you act.',
    detail: 'Alerts come from sensor thresholds, automation failures, or stock rules. Acknowledge or create a task from one; they do not control pumps or lights.',
  },
  schedule: {
    label: 'Schedule (what runs when)',
    dbTable: 'gr33ncore.schedules',
    shortTip: 'Clock-based timing — e.g. “every day at 8 AM” — that can start feeds, light toggles, or rule checks.',
    detail: 'Schedules use cron/time. They may trigger a fertigation program, call automation rules, or enqueue actuator commands. Pausing a schedule stops timed runs only.',
  },
  rule: {
    label: 'Automation rule',
    dbTable: 'gr33ncore.automation_rules',
    shortTip: 'IF something (time, sensor, schedule) THEN do actions — turn on a fan, create a task, send alert.',
    detail: 'Rules evaluate predicates and run executable_actions. They are separate from comfort bands: a rule might react when humidity leaves the band you set.',
  },
  automation_run: {
    label: 'Rule run',
    dbTable: 'gr33ncore.automation_runs',
    shortTip: 'A log entry each time a schedule or rule actually fired — success, skipped, or failed.',
    detail: 'Rule runs are history, not configuration. Use them to debug “why didn’t my fan turn on?” — they link back to the schedule or rule that fired.',
  },
  comfort_band: {
    label: 'Comfort band',
    dbTable: 'gr33ncore.zone_setpoints',
    shortTip: 'Target min / ideal / max for a sensor type in a zone or grow stage — what “comfortable” means.',
    detail: 'Comfort bands are farmer-friendly setpoints (temp, humidity, etc.). Alerts and rules can reference them; they do not run hardware alone.',
  },
  setpoint: {
    label: 'Raw setpoint',
    dbTable: 'gr33ncore.zone_setpoints',
    shortTip: 'Same table as comfort bands — the technical row view (zone, stage, sensor type, numbers).',
    detail: '“Comfort band” and “raw setpoint” are the same data; Comfort tab is the editor, Raw tab is the full table for power users.',
  },
  sensor_threshold: {
    label: 'Sensor alert threshold',
    dbTable: 'gr33ncore.sensors (low/high on sensor row)',
    shortTip: 'Per-sensor alert limits on the sensor detail page — separate from zone comfort bands.',
    detail: 'Thresholds on a sensor fire alerts for that device. Zone comfort bands are farm/zone targets; both can exist without conflict if ranges align.',
  },
  input_definition: {
    label: 'Input',
    dbTable: 'gr33nnaturalfarming.input_definitions',
    shortTip: 'The type of ferment or supplement — what JMS or JLF is, how to prepare it, storage rules. Not a specific jar on the shelf.',
    detail: 'One row per product type. Many batches can share the same input. Created in Make a batch or imported from the Commons pack.',
  },
  input_batch: {
    label: 'Batch',
    dbTable: 'gr33nnaturalfarming.input_batches',
    shortTip: 'One production run of an input — start date, ferment status, liters left, batch code. What you actually made.',
    detail: 'Status moves planning → fermenting → ready → partially used. Mixing and programs reference a specific batch when dosing.',
  },
  application_recipe: {
    label: 'Apply recipe',
    dbTable: 'gr33nnaturalfarming.application_recipes',
    shortTip: 'How to use inputs on crops — dilution, foliar vs drench, which batches to mix. Links to Feed & water programs; does not start a pump alone.',
    detail: 'Different from making JMS (that is a batch). Apply wires zone programs; the Advanced tab mixing log records what went into the tank.',
  },
  nf_field_guide: {
    label: 'Field guide',
    dbTable: 'recipe canon (read-only)',
    shortTip: 'Read-only reference for how to make inputs and apply them — not your farm inventory. Make a batch when you actually ferment.',
    detail: 'Canon from Phase 208: input prep, apply recipes, bootstrap programs. Your farm rows live under Make a batch, Apply recipes, and Ready batches.',
  },
  feeding_plan: {
    label: 'Daily feed snapshot',
    dbTable: 'gr33nfertigation.programs (active per zone)',
    shortTip: 'One card per zone — next feed time, last run OK/fail, reservoir ready. Change the plan on the zone Water tab or Programs & tanks.',
    detail: 'Read-only rollup for the morning pass. Each card reflects the active program + schedule for that zone, not a separate table.',
  },
  fertigation_program: {
    label: 'Feeding program',
    dbTable: 'gr33nfertigation.programs',
    shortTip: 'Volume, EC targets, reservoir, and schedule for a zone — what runs when a feed fires. Pair with a nutrient tank (reservoir).',
    detail: 'Programs link zones, reservoirs, schedules, and optional apply recipes. Irrigation-only programs skip nutrients (plain water).',
  },
  reservoir: {
    label: 'Nutrient tank',
    dbTable: 'gr33nfertigation.reservoirs',
    shortTip: 'Physical mix tank — capacity, current volume, EC/pH, ready/empty status. Programs draw from a reservoir when feeding.',
    detail: 'Track tank fill level and readiness. Mixing events and fertigation runs reference the reservoir that supplied the feed.',
  },
  mixing_event: {
    label: 'Mixing log',
    dbTable: 'gr33nfertigation.mixing_events',
    shortTip: 'Record what went into the tank — water volume, inputs/batches, final EC/pH. Natural farming batches link here when dosing.',
    detail: 'Mixing is configuration history for the tank, not the feed applied to plants. Applied feeds appear as fertigation events.',
  },
  fertigation_console: {
    label: 'Fertigation console',
    dbTable: 'gr33nfertigation (programs, reservoirs, events…)',
    shortTip: 'Full editor — EC targets, programs, mixing log, crop cycles, and every feed event. Same data as other tabs, table view.',
    detail: 'Power-user surface absorbed from legacy /fertigation. Prefer Daily and Programs & tanks for day-to-day; use Advanced when debugging or bulk editing.',
  },
}

/** Concepts shown on Comfort & automation workspace. */
export const COMFORT_WORKSPACE_CONCEPTS = [
  'comfort_band',
  'schedule',
  'rule',
  'automation_run',
  'setpoint',
  'alert',
  'task',
]

/** How concepts relate — shown in the glossary banner. */
export const OPERATOR_CONCEPT_RELATIONSHIPS = [
  'Comfort bands set targets; rules can react when readings leave those bands.',
  'Schedules fire on a clock; rules decide what happens when they fire.',
  'Alerts notify you; tasks are work for you (or created by a rule) to follow up.',
]

/** Concepts for Natural farming workspace. */
export const NATURAL_FARMING_WORKSPACE_CONCEPTS = [
  'input_definition',
  'input_batch',
  'application_recipe',
]

export const NATURAL_FARMING_CONCEPT_RELATIONSHIPS = [
  'Input = what it is. Batch = what you made. Apply recipe = how you use it on plants.',
  'Field guide is read-only. Make a batch → Apply recipes links programs. Edit rows under Inputs & batches; costs under Money → Supplies.',
]

/** Concepts for Feed & water workspace tabs. */
export const FEED_WATER_WORKSPACE_CONCEPTS = [
  'feeding_plan',
  'fertigation_program',
  'reservoir',
  'mixing_event',
  'fertigation_console',
  'application_recipe',
  'schedule',
]

export const FEED_WATER_CONCEPT_RELATIONSHIPS = [
  'Program = what to run. Reservoir = tank it draws from. Mixing log = what you put in the tank. Events = what reached the plants.',
  'Apply recipes (Natural farming) wire into programs; Daily cards show the active plan per zone.',
]

/**
 * @param {string} id
 * @returns {OperatorConcept | null}
 */
export function operatorConcept(id) {
  return OPERATOR_CONCEPTS[id] ?? null
}
