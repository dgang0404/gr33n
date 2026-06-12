/**
 * Phase 91 — bundled bootstrap template fallback (mirrors migration seed).
 */

export const BOOTSTRAP_TEMPLATE_KEYS = {
  JADAM_INDOOR_PHOTOPERIOD_V1: 'jadam_indoor_photoperiod_v1',
  CHICKEN_COOP_V1: 'chicken_coop_v1',
  GREENHOUSE_CLIMATE_V1: 'greenhouse_climate_v1',
  DRYING_ROOM_V1: 'drying_room_v1',
  SMALL_AQUAPONICS_V1: 'small_aquaponics_v1',
}

/** @type {import('./bootstrapCatalog.js').BootstrapTemplate[]} */
export const FALLBACK_TEMPLATES = [
  {
    template_key: BOOTSTRAP_TEMPLATE_KEYS.JADAM_INDOOR_PHOTOPERIOD_V1,
    label: 'Indoor photoperiod starter (v1)',
    short_label: 'Indoor photoperiod v1',
    tagline: 'Four zones, feeding programs, inventory, and demo tasks',
    summary_title: 'Included in this starter pack (idempotent — the API skips duplicate rows)',
    summary_bullets: [
      'Four zones: Seedling Room (indoor), Veg Room (indoor), Flower Room (indoor), Outdoor Garden (outdoor)',
      'Lighting schedules (18/6 veg, 12/12 flower) + active irrigation schedules per zone',
      'Inventory: JMS, JLF, FFJ, WCA inputs plus ready-to-use starter batches',
      'Recipes: JMS / JLF / combined drench + FFJ+WCA flowering boost with components',
      'Fertigation: 3 reservoirs, EC targets per zone, 3 programs (veg JLF, flower FFJ+WCA, outdoor JLF drench) each linked to a schedule',
      'Mixing log: 3 mixing events tied to reservoirs, programs, and inventory batches; fertigation events linked to mixes',
      'Crop cycles: active cycle per zone with primary program link',
      'Tasks: reservoir refresh tasks per zone, each linked to its irrigation schedule',
    ],
    module_hints: ['zones', 'fertigation', 'inventory', 'lighting'],
    icon: '🌱',
    recommended: true,
    wizard_primary: true,
    related_commons_slug: 'gr33n-cultivator-seed-pack-v1',
    sort_order: 10,
  },
  {
    template_key: BOOTSTRAP_TEMPLATE_KEYS.GREENHOUSE_CLIMATE_V1,
    label: 'Greenhouse climate (v1)',
    short_label: 'Greenhouse v1',
    tagline: 'Shade, vents, humidity bands, and Pi placeholder',
    summary_title: 'Greenhouse / tent climate (dew point, VPD, CO2 — pair with Pi derived sensors)',
    summary_bullets: [
      'One zone: Greenhouse + Pi device placeholder',
      'Sensors: air temp, RH, CO2, dew point, VPD (Pa)',
      'Actuators: exhaust fan, humidifier, dehumidifier, shade motor, CO2 injector',
      'Automation rules (inactive): dew/VPD/CO2/temperature thresholds → equipment',
      'Task: weekly CO2 / enrichment checklist',
    ],
    module_hints: ['zones', 'greenhouse', 'climate'],
    icon: '🏠',
    recommended: false,
    wizard_primary: true,
    sort_order: 20,
  },
  {
    template_key: BOOTSTRAP_TEMPLATE_KEYS.CHICKEN_COOP_V1,
    label: 'Chicken coop (v1)',
    short_label: 'Chicken coop v1',
    tagline: 'Coop sensors, feeder, and climate actuators',
    summary_title: 'Chicken coop starter (sensors, actuators, schedules, rules — tune before enabling rules)',
    summary_bullets: [
      'One zone: Chicken Coop + Pi device placeholder',
      'Sensors: water level, feed level, air temperature, humidity',
      'Actuators: feeder hopper, water valve, exhaust fan, heat lamp',
      'Schedules: morning / evening feed reminders (inactive by default)',
      'Automation rules (inactive): low water / low feed → tasks; hot → fan; cold → heat lamp',
      'Task: weekly egg collection reminder',
    ],
    module_hints: ['zones', 'animals', 'climate'],
    sort_order: 30,
  },
  {
    template_key: BOOTSTRAP_TEMPLATE_KEYS.DRYING_ROOM_V1,
    label: 'Drying / cure room (v1)',
    short_label: 'Drying room v1',
    tagline: 'Post-harvest environment monitoring',
    summary_title: 'Drying / cure room (defaults skew cannabis; retune for basil, orchids, herbs)',
    summary_bullets: [
      'One zone: Drying Room + Pi device placeholder',
      'Sensors: temperature, humidity, dew point',
      'Actuators: dehumidifier, circulation fan',
      'Automation rules (inactive): dew-point on/off band + high-RH circulation',
      'Task: daily environment log reminder',
    ],
    module_hints: ['zones', 'climate', 'harvest'],
    sort_order: 40,
  },
  {
    template_key: BOOTSTRAP_TEMPLATE_KEYS.SMALL_AQUAPONICS_V1,
    label: 'Small aquaponics (v1)',
    short_label: 'Aquaponics v1',
    tagline: 'Fish tank + grow bed loop starter',
    summary_title: 'Small aquaponics loop (fish tank + grow bed)',
    summary_bullets: [
      'Two zones: Fish Tank, Grow Bed + Pi device placeholder',
      'Tank sensors: water temperature, pH, ammonia, nitrate; bed sensors: pH, EC',
      'Actuators: return pump, air pump',
      'gr33naquaponics.loops row: Main aquaponics loop (meta documents zone names)',
      'Schedule: daily fish-feed reminder (inactive)',
      'Automation rules (inactive): ammonia spike → task; cold tank → task',
      'Task: daily feed fish reminder',
    ],
    module_hints: ['zones', 'aquaponics', 'water'],
    sort_order: 50,
  },
]

/**
 * @param {{ templates?: import('./bootstrapCatalog.js').BootstrapTemplate[] }} payload
 */
export function indexBootstrapCatalog(payload) {
  const templates = [...(payload?.templates?.length ? payload.templates : FALLBACK_TEMPLATES)]
  templates.sort((a, b) => (a.sort_order || 0) - (b.sort_order || 0) || String(a.template_key).localeCompare(String(b.template_key)))

  /** @type {Record<string, import('./bootstrapCatalog.js').BootstrapTemplate>} */
  const byKey = {}
  /** @type {{ value: string, label: string, shortLabel: string }[]} */
  const starterOptions = []
  /** @type {Record<string, { title: string, bullets: string[] }>} */
  const summariesByKey = {}

  for (const t of templates) {
    byKey[t.template_key] = t
    starterOptions.push({
      value: t.template_key,
      label: t.label,
      shortLabel: t.short_label || t.label,
    })
    summariesByKey[t.template_key] = {
      title: t.summary_title,
      bullets: t.summary_bullets || [],
    }
  }

  return { templates, byKey, starterOptions, summariesByKey }
}

export const FALLBACK_BOOTSTRAP_CATALOG = indexBootstrapCatalog({ templates: FALLBACK_TEMPLATES })
