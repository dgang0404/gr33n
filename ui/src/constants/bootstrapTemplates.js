/** Keys must match `gr33ncore.apply_farm_bootstrap_template` (API / DB). */
export const BOOTSTRAP_TEMPLATE_KEYS = {
  JADAM_INDOOR_PHOTOPERIOD_V1: 'jadam_indoor_photoperiod_v1',
  CHICKEN_COOP_V1: 'chicken_coop_v1',
  GREENHOUSE_CLIMATE_V1: 'greenhouse_climate_v1',
  DRYING_ROOM_V1: 'drying_room_v1',
  SMALL_AQUAPONICS_V1: 'small_aquaponics_v1',
}

/** Farm-create / org-default picker options (value empty = no template / use org default flow in UI). */
export const BOOTSTRAP_STARTER_OPTIONS = [
  {
    value: BOOTSTRAP_TEMPLATE_KEYS.JADAM_INDOOR_PHOTOPERIOD_V1,
    label: 'Indoor photoperiod starter (v1)',
    shortLabel: 'Indoor photoperiod v1',
  },
  {
    value: BOOTSTRAP_TEMPLATE_KEYS.CHICKEN_COOP_V1,
    label: 'Chicken coop (v1)',
    shortLabel: 'Chicken coop v1',
  },
  {
    value: BOOTSTRAP_TEMPLATE_KEYS.GREENHOUSE_CLIMATE_V1,
    label: 'Greenhouse climate (v1)',
    shortLabel: 'Greenhouse v1',
  },
  {
    value: BOOTSTRAP_TEMPLATE_KEYS.DRYING_ROOM_V1,
    label: 'Drying / cure room (v1)',
    shortLabel: 'Drying room v1',
  },
  {
    value: BOOTSTRAP_TEMPLATE_KEYS.SMALL_AQUAPONICS_V1,
    label: 'Small aquaponics (v1)',
    shortLabel: 'Aquaponics v1',
  },
]

export const JADAM_INDOOR_PHOTOPERIOD_V1_SUMMARY = {
  title: 'Included in this starter pack (idempotent \u2014 the API skips duplicate rows)',
  bullets: [
    'Four zones: Seedling Room (indoor), Veg Room (indoor), Flower Room (indoor), Outdoor Garden (outdoor)',
    'Lighting schedules (18/6 veg, 12/12 flower) + active irrigation schedules per zone',
    'Inventory: JMS, JLF, FFJ, WCA inputs plus ready-to-use starter batches',
    'Recipes: JMS / JLF / combined drench + FFJ+WCA flowering boost with components',
    'Fertigation: 3 reservoirs, EC targets per zone, 3 programs (veg JLF, flower FFJ+WCA, outdoor JLF drench) each linked to a schedule',
    'Mixing log: 3 mixing events tied to reservoirs, programs, and inventory batches; fertigation events linked to mixes',
    'Crop cycles: active cycle per zone with primary program link',
    'Tasks: reservoir refresh tasks per zone, each linked to its irrigation schedule',
  ],
}

export const CHICKEN_COOP_V1_SUMMARY = {
  title: 'Chicken coop starter (sensors, actuators, schedules, rules — tune before enabling rules)',
  bullets: [
    'One zone: Chicken Coop + Pi device placeholder',
    'Sensors: water level, feed level, air temperature, humidity',
    'Actuators: feeder hopper, water valve, exhaust fan, heat lamp',
    'Schedules: morning / evening feed reminders (inactive by default)',
    'Automation rules (inactive): low water / low feed → tasks; hot → fan; cold → heat lamp',
    'Task: weekly egg collection reminder',
  ],
}

export const GREENHOUSE_CLIMATE_V1_SUMMARY = {
  title: 'Greenhouse / tent climate (dew point, VPD, CO2 — pair with Pi derived sensors)',
  bullets: [
    'One zone: Greenhouse + Pi device placeholder',
    'Sensors: air temp, RH, CO2, dew point, VPD (Pa)',
    'Actuators: exhaust fan, humidifier, dehumidifier, shade motor, CO2 injector',
    'Automation rules (inactive): dew/VPD/CO2/temperature thresholds → equipment',
    'Task: weekly CO2 / enrichment checklist',
  ],
}

export const DRYING_ROOM_V1_SUMMARY = {
  title: 'Drying / cure room (defaults skew cannabis; retune for basil, orchids, herbs)',
  bullets: [
    'One zone: Drying Room + Pi device placeholder',
    'Sensors: temperature, humidity, dew point',
    'Actuators: dehumidifier, circulation fan',
    'Automation rules (inactive): dew-point on/off band + high-RH circulation',
    'Task: daily environment log reminder',
  ],
}

export const SMALL_AQUAPONICS_V1_SUMMARY = {
  title: 'Small aquaponics loop (fish tank + grow bed)',
  bullets: [
    'Two zones: Fish Tank, Grow Bed + Pi device placeholder',
    'Tank sensors: water temperature, pH, ammonia, nitrate; bed sensors: pH, EC',
    'Actuators: return pump, air pump',
    'gr33naquaponics.loops row: Main aquaponics loop (meta documents zone names)',
    'Schedule: daily fish-feed reminder (inactive)',
    'Automation rules (inactive): ammonia spike → task; cold tank → task',
    'Task: daily feed fish reminder',
  ],
}

/** Map template key → { title, bullets } for expandable help on farm create / apply-starter UI. */
export const BOOTSTRAP_STARTER_SUMMARIES = {
  [BOOTSTRAP_TEMPLATE_KEYS.JADAM_INDOOR_PHOTOPERIOD_V1]: JADAM_INDOOR_PHOTOPERIOD_V1_SUMMARY,
  [BOOTSTRAP_TEMPLATE_KEYS.CHICKEN_COOP_V1]: CHICKEN_COOP_V1_SUMMARY,
  [BOOTSTRAP_TEMPLATE_KEYS.GREENHOUSE_CLIMATE_V1]: GREENHOUSE_CLIMATE_V1_SUMMARY,
  [BOOTSTRAP_TEMPLATE_KEYS.DRYING_ROOM_V1]: DRYING_ROOM_V1_SUMMARY,
  [BOOTSTRAP_TEMPLATE_KEYS.SMALL_AQUAPONICS_V1]: SMALL_AQUAPONICS_V1_SUMMARY,
}
