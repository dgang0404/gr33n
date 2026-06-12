-- Phase 91 — Bootstrap template catalog for UI pickers and Guardian validation.

CREATE TABLE IF NOT EXISTS gr33ncore.bootstrap_templates (
    template_key         TEXT PRIMARY KEY,
    label                TEXT NOT NULL,
    short_label          TEXT,
    tagline              TEXT,
    summary_title        TEXT NOT NULL,
    summary_bullets      JSONB NOT NULL DEFAULT '[]'::jsonb,
    module_hints         JSONB NOT NULL DEFAULT '[]'::jsonb,
    icon                 TEXT,
    recommended          BOOLEAN NOT NULL DEFAULT false,
    wizard_primary       BOOLEAN NOT NULL DEFAULT false,
    playbook_section     TEXT,
    related_commons_slug TEXT,
    sort_order           INT NOT NULL DEFAULT 0,
    is_active            BOOLEAN NOT NULL DEFAULT true
);

INSERT INTO gr33ncore.bootstrap_templates (
    template_key, label, short_label, tagline, summary_title, summary_bullets,
    module_hints, icon, recommended, wizard_primary, playbook_section, related_commons_slug, sort_order
) VALUES
(
    'jadam_indoor_photoperiod_v1',
    'Indoor photoperiod starter (v1)',
    'Indoor photoperiod v1',
    'Four zones, feeding programs, inventory, and demo tasks',
    'Included in this starter pack (idempotent — the API skips duplicate rows)',
    '["Four zones: Seedling Room (indoor), Veg Room (indoor), Flower Room (indoor), Outdoor Garden (outdoor)","Lighting schedules (18/6 veg, 12/12 flower) + active irrigation schedules per zone","Inventory: JMS, JLF, FFJ, WCA inputs plus ready-to-use starter batches","Recipes: JMS / JLF / combined drench + FFJ+WCA flowering boost with components","Fertigation: 3 reservoirs, EC targets per zone, 3 programs (veg JLF, flower FFJ+WCA, outdoor JLF drench) each linked to a schedule","Mixing log: 3 mixing events tied to reservoirs, programs, and inventory batches; fertigation events linked to mixes","Crop cycles: active cycle per zone with primary program link","Tasks: reservoir refresh tasks per zone, each linked to its irrigation schedule"]'::jsonb,
    '["zones","fertigation","inventory","lighting"]'::jsonb,
    '🌱', true, true,
    'JADAM indoor photoperiod (`jadam_indoor_photoperiod_v1`)',
    'gr33n-cultivator-seed-pack-v1', 10
),
(
    'greenhouse_climate_v1',
    'Greenhouse climate (v1)',
    'Greenhouse v1',
    'Shade, vents, humidity bands, and Pi placeholder',
    'Greenhouse / tent climate (dew point, VPD, CO2 — pair with Pi derived sensors)',
    '["One zone: Greenhouse + Pi device placeholder","Sensors: air temp, RH, CO2, dew point, VPD (Pa)","Actuators: exhaust fan, humidifier, dehumidifier, shade motor, CO2 injector","Automation rules (inactive): dew/VPD/CO2/temperature thresholds → equipment","Task: weekly CO2 / enrichment checklist"]'::jsonb,
    '["zones","greenhouse","climate"]'::jsonb,
    '🏠', false, true,
    'Greenhouse climate (`greenhouse_climate_v1`)',
    NULL, 20
),
(
    'chicken_coop_v1',
    'Chicken coop (v1)',
    'Chicken coop v1',
    'Coop sensors, feeder, and climate actuators',
    'Chicken coop starter (sensors, actuators, schedules, rules — tune before enabling rules)',
    '["One zone: Chicken Coop + Pi device placeholder","Sensors: water level, feed level, air temperature, humidity","Actuators: feeder hopper, water valve, exhaust fan, heat lamp","Schedules: morning / evening feed reminders (inactive by default)","Automation rules (inactive): low water / low feed → tasks; hot → fan; cold → heat lamp","Task: weekly egg collection reminder"]'::jsonb,
    '["zones","animals","climate"]'::jsonb,
    NULL, false, false,
    'Chicken coop (`chicken_coop_v1`)',
    NULL, 30
),
(
    'drying_room_v1',
    'Drying / cure room (v1)',
    'Drying room v1',
    'Post-harvest environment monitoring',
    'Drying / cure room (defaults skew cannabis; retune for basil, orchids, herbs)',
    '["One zone: Drying Room + Pi device placeholder","Sensors: temperature, humidity, dew point","Actuators: dehumidifier, circulation fan","Automation rules (inactive): dew-point on/off band + high-RH circulation","Task: daily environment log reminder"]'::jsonb,
    '["zones","climate","harvest"]'::jsonb,
    NULL, false, false,
    'Drying room (`drying_room_v1`)',
    NULL, 40
),
(
    'small_aquaponics_v1',
    'Small aquaponics (v1)',
    'Aquaponics v1',
    'Fish tank + grow bed loop starter',
    'Small aquaponics loop (fish tank + grow bed)',
    '["Two zones: Fish Tank, Grow Bed + Pi device placeholder","Tank sensors: water temperature, pH, ammonia, nitrate; bed sensors: pH, EC","Actuators: return pump, air pump","gr33naquaponics.loops row: Main aquaponics loop (meta documents zone names)","Schedule: daily fish-feed reminder (inactive)","Automation rules (inactive): ammonia spike → task; cold tank → task","Task: daily feed fish reminder"]'::jsonb,
    '["zones","aquaponics","water"]'::jsonb,
    NULL, false, false,
    'Small aquaponics (`small_aquaponics_v1`)',
    NULL, 50
)
ON CONFLICT (template_key) DO UPDATE SET
    label = EXCLUDED.label,
    short_label = EXCLUDED.short_label,
    tagline = EXCLUDED.tagline,
    summary_title = EXCLUDED.summary_title,
    summary_bullets = EXCLUDED.summary_bullets,
    module_hints = EXCLUDED.module_hints,
    icon = EXCLUDED.icon,
    recommended = EXCLUDED.recommended,
    wizard_primary = EXCLUDED.wizard_primary,
    playbook_section = EXCLUDED.playbook_section,
    related_commons_slug = EXCLUDED.related_commons_slug,
    sort_order = EXCLUDED.sort_order;
