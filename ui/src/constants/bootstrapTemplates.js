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
  title: 'Included in this starter pack (safe to re-apply; duplicates are skipped)',
  bullets: [
    'Four zones: Seedling Room, Veg Room, Flower Room, Outdoor Beds',
    'Lighting and irrigation schedules (continuous, 18/6 veg, 12/12 flower, daily irrigation hooks)',
    'JADAM input definitions: JMS, JLF General, LAB',
    'Application recipes (JMS drench, JLF drench, combined) with recipe components',
    'Fertigation reservoir, EC targets, and a “Veg Daily JLF” program',
    'One starter task',
  ],
}
