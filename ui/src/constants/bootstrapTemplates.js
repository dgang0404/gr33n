/** Keys must match `gr33ncore.apply_farm_bootstrap_template` (API / DB). */
export const BOOTSTRAP_TEMPLATE_KEYS = {
  JADAM_INDOOR_PHOTOPERIOD_V1: 'jadam_indoor_photoperiod_v1',
}

/** Farm-create / org-default picker options (value empty = no template / use org default flow in UI). */
export const BOOTSTRAP_STARTER_OPTIONS = [
  {
    value: BOOTSTRAP_TEMPLATE_KEYS.JADAM_INDOOR_PHOTOPERIOD_V1,
    label: 'Indoor photoperiod starter (v1)',
    shortLabel: 'Indoor photoperiod v1',
  },
]

export const JADAM_INDOOR_PHOTOPERIOD_V1_SUMMARY = {
  title: 'Included in this starter pack (idempotent \u2014 the API skips duplicate rows)',
  bullets: [
    'Three zones: Veg Room (indoor), Flower Room (indoor), Outdoor Garden (outdoor)',
    'Lighting schedules (18/6 veg, 12/12 flower) + active irrigation schedules per zone',
    'Inventory: JMS, JLF, FFJ, WCA inputs plus ready-to-use starter batches',
    'Recipes: JMS / JLF / combined drench + FFJ+WCA flowering boost with components',
    'Fertigation: 3 reservoirs, EC targets per zone, 3 programs (veg JLF, flower FFJ+WCA, outdoor JLF drench) each linked to a schedule',
    'Mixing log: 3 mixing events tied to reservoirs, programs, and inventory batches; fertigation events linked to mixes',
    'Crop cycles: active cycle per zone with primary program link',
    'Tasks: reservoir refresh tasks per zone, each linked to its irrigation schedule',
  ],
}
