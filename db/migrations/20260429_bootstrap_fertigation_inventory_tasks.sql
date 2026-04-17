-- Phase 15+: richer JADAM starter — inventory batches, flower fertigation, mixing history,
-- fertigation demo events, tasks linked to irrigation schedules, tasks.schedule_id column.

ALTER TABLE gr33ncore.tasks
    ADD COLUMN IF NOT EXISTS schedule_id BIGINT REFERENCES gr33ncore.schedules(id) ON DELETE SET NULL;

CREATE OR REPLACE FUNCTION gr33ncore._bootstrap_jadam_indoor_photoperiod_v1(p_farm_id BIGINT, p_tz TEXT)
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    r_veg      BIGINT;
    r_flower   BIGINT;
    r_outdoor  BIGINT;
    p_veg      BIGINT;
    p_flower   BIGINT;
    p_outdoor  BIGINT;
    mix_veg    BIGINT;
    mix_fl     BIGINT;
    mix_out    BIGINT;
    b_jlf      BIGINT;
    b_jms      BIGINT;
    b_ffj      BIGINT;
    b_wca      BIGINT;
    i_jlf      BIGINT;
    i_jms      BIGINT;
    i_ffj      BIGINT;
    i_wca      BIGINT;
BEGIN
    INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
    SELECT p_farm_id, x.n, x.d, x.t
    FROM (
        VALUES
            ('Veg Room', 'Vegetative growth stage. 18/6 light, JLF+JMS feeding.', 'indoor'),
            ('Flower Room', 'Flowering and fruiting stage. 12/12 light, FFJ+WCA program.', 'indoor'),
            ('Outdoor Garden', 'Outdoor raised beds and garden rows. Natural light. JADAM soil program.', 'outdoor')
    ) AS x(n, d, t)
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33ncore.zones z
        WHERE z.farm_id = p_farm_id AND z.name = x.n AND z.deleted_at IS NULL
    );

    INSERT INTO gr33ncore.schedules (farm_id, name, description, schedule_type, cron_expression, timezone, is_active)
    SELECT p_farm_id, x.n, x.d, x.st, x.cron, p_tz, x.active
    FROM (
        VALUES
            ('Light ON 18/6 Veg', 'Lights on at 06:00 for vegetative growth.', 'lighting', '0 6 * * *', FALSE),
            ('Light OFF 18/6 Veg', 'Lights off at midnight (6h dark).', 'lighting', '0 0 * * *', FALSE),
            ('Light ON 12/12 Flower', 'Lights on at 06:00 for flowering photoperiod.', 'lighting', '0 6 * * *', FALSE),
            ('Light OFF 12/12 Flower', 'Lights off at 18:00 (12h dark).', 'lighting', '0 18 * * *', FALSE),
            ('Water Late Veg Daily', 'Late veg daily irrigation (pairs with veg fertigation program).', 'irrigation', '0 8 * * *', TRUE),
            ('Water Early Flower Daily', 'Early flower daily irrigation.', 'irrigation', '0 8 * * *', TRUE),
            ('Water Outdoor Garden Daily', 'Morning irrigation for outdoor garden beds.', 'irrigation', '0 7 * * *', TRUE)
    ) AS x(n, d, st, cron, active)
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33ncore.schedules s
        WHERE s.farm_id = p_farm_id AND s.name = x.n
    );

    INSERT INTO gr33nnaturalfarming.input_definitions
        (farm_id, name, category, description, typical_ingredients, preparation_summary,
         storage_guidelines, safety_precautions, reference_source)
    SELECT p_farm_id, x.name, x.cat::gr33nnaturalfarming.input_category_enum, x.descr, x.ting, x.prep, x.store, x.safe, x.ref
    FROM (
        VALUES
            ('JMS (JADAM Microbial Solution)', 'microbial_inoculant',
             'JADAM microbial inoculant from leaf mold and potato water.',
             'Leaf mold, potato water, pinch salt',
             'Ferment 3–7 days.', 'Use within ~1 week active.', 'Non-chlorinated water.', 'JADAM OF'),
            ('JLF General (Weed and Grass)', 'other_ferment',
             'JADAM liquid fertilizer from local weeds — main veg fertility input.',
             'Weeds, leaf mold, water',
             'Ferment 7–14 days; strain.', 'Strained: use within ~30 days.', 'No herbicide material.', 'JADAM OF'),
            ('LAB (Lactic Acid Bacteria Serum)', 'microbial_inoculant',
             'LAB serum for soil conditioning.',
             'Rice wash, milk',
             'Ferment; extract serum.', 'Refrigerate preserved.', 'Use golden layer only.', 'JADAM OF'),
            ('FFJ (Fermented Fruit Juice)', 'fermented_plant_juice',
             'Sweet ferment for flowering support.',
             'Fruit, sugar, water',
             'Ferment; strain.', 'Refrigerate.', 'Avoid over-concentration.', 'JADAM OF'),
            ('WCA (Water-Soluble Calcium)', 'water_soluble_nutrient',
             'Calcium from eggshell vinegar extract.',
             'Eggshells, vinegar',
             'Extract; dilute.', 'Store labeled.', 'Acid — eye protection.', 'JADAM OF')
    ) AS x(name, cat, descr, ting, prep, store, safe, ref)
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33nnaturalfarming.input_definitions d
        WHERE d.farm_id = p_farm_id AND d.name = x.name AND d.deleted_at IS NULL
    );

    INSERT INTO gr33nnaturalfarming.application_recipes
        (farm_id, name, description, target_application_type, dilution_ratio, instructions, frequency_guidelines,
         target_crop_categories, target_growth_stages)
    SELECT p_farm_id, x.n, x.descr, x.ttype::gr33nnaturalfarming.application_target_enum, x.dil, x.inst, x.freq, x.cats, x.stages
    FROM (
        VALUES
            ('JMS Soil Drench', 'Base soil microbe inoculant.', 'soil_drench', '1:500 (JMS:water)',
             'Dilute; drench root zone.', 'Every 2–4 weeks.',
             ARRAY['All crops']::text[], ARRAY['All stages']::text[]),
            ('JLF General Soil Drench', 'Primary JLF soil fertility.', 'soil_drench', '1:20 (JLF:water)',
             'Strain; dilute; drench.', 'Every 1–2 weeks active growth.',
             ARRAY['All crops']::text[], ARRAY['All stages']::text[]),
            ('JLF and JMS Combined Drench', 'Nutrients + microbes in one pass.', 'soil_drench', 'JLF 1:20 + JMS 1:500',
             'Fill tank; add JLF then JMS dilutions.', 'Weekly peak season.',
             ARRAY['All crops']::text[], ARRAY['All stages']::text[]),
            ('FFJ and WCA Flowering Boost', 'Flower-phase tank and foliar oriented mix.', 'foliar_spray', 'FFJ 1:500 + WCA 1:1000',
             'Mix per label; apply morning.', 'Weekly in early flower.',
             ARRAY['All crops']::text[], ARRAY['Flowering']::text[])
    ) AS x(n, descr, ttype, dil, inst, freq, cats, stages)
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33nnaturalfarming.application_recipes r
        WHERE r.farm_id = p_farm_id AND r.name = x.n AND r.deleted_at IS NULL
    );

    INSERT INTO gr33nnaturalfarming.recipe_input_components
        (application_recipe_id, input_definition_id, part_value, part_unit_id, notes)
    SELECT r.id, d.id, 1.0, u.id, 'Template component ratio — tune in production.'
    FROM gr33nnaturalfarming.application_recipes r
    CROSS JOIN gr33nnaturalfarming.input_definitions d
    CROSS JOIN gr33ncore.units u
    WHERE r.farm_id = p_farm_id AND d.farm_id = p_farm_id AND u.name = 'decimal_fraction'
      AND r.name = 'JMS Soil Drench' AND d.name LIKE 'JMS%'
      AND NOT EXISTS (
          SELECT 1 FROM gr33nnaturalfarming.recipe_input_components c
          WHERE c.application_recipe_id = r.id AND c.input_definition_id = d.id
      );

    INSERT INTO gr33nnaturalfarming.recipe_input_components
        (application_recipe_id, input_definition_id, part_value, part_unit_id, notes)
    SELECT r.id, d.id, 1.0, u.id, 'JLF general at 1:20 basis.'
    FROM gr33nnaturalfarming.application_recipes r
    CROSS JOIN gr33nnaturalfarming.input_definitions d
    CROSS JOIN gr33ncore.units u
    WHERE r.farm_id = p_farm_id AND d.farm_id = p_farm_id AND u.name = 'decimal_fraction'
      AND r.name = 'JLF General Soil Drench' AND d.name LIKE 'JLF General%'
      AND NOT EXISTS (
          SELECT 1 FROM gr33nnaturalfarming.recipe_input_components c
          WHERE c.application_recipe_id = r.id AND c.input_definition_id = d.id
      );

    INSERT INTO gr33nnaturalfarming.recipe_input_components
        (application_recipe_id, input_definition_id, part_value, part_unit_id, notes)
    SELECT r.id, d.id, 1.0, u.id, 'JLF at 1:20'
    FROM gr33nnaturalfarming.application_recipes r
    CROSS JOIN gr33nnaturalfarming.input_definitions d
    CROSS JOIN gr33ncore.units u
    WHERE r.farm_id = p_farm_id AND d.farm_id = p_farm_id AND u.name = 'decimal_fraction'
      AND r.name = 'JLF and JMS Combined Drench' AND d.name LIKE 'JLF General%'
      AND NOT EXISTS (
          SELECT 1 FROM gr33nnaturalfarming.recipe_input_components c
          WHERE c.application_recipe_id = r.id AND c.input_definition_id = d.id
      );

    INSERT INTO gr33nnaturalfarming.recipe_input_components
        (application_recipe_id, input_definition_id, part_value, part_unit_id, notes)
    SELECT r.id, d.id, 0.025, u.id, 'JMS at 1:500 relative to JLF base'
    FROM gr33nnaturalfarming.application_recipes r
    CROSS JOIN gr33nnaturalfarming.input_definitions d
    CROSS JOIN gr33ncore.units u
    WHERE r.farm_id = p_farm_id AND d.farm_id = p_farm_id AND u.name = 'decimal_fraction'
      AND r.name = 'JLF and JMS Combined Drench' AND d.name LIKE 'JMS%'
      AND NOT EXISTS (
          SELECT 1 FROM gr33nnaturalfarming.recipe_input_components c
          WHERE c.application_recipe_id = r.id AND c.input_definition_id = d.id
      );

    INSERT INTO gr33nnaturalfarming.recipe_input_components
        (application_recipe_id, input_definition_id, part_value, part_unit_id, notes)
    SELECT r.id, d.id, 1.0, u.id, 'FFJ contribution'
    FROM gr33nnaturalfarming.application_recipes r
    CROSS JOIN gr33nnaturalfarming.input_definitions d
    CROSS JOIN gr33ncore.units u
    WHERE r.farm_id = p_farm_id AND d.farm_id = p_farm_id AND u.name = 'decimal_fraction'
      AND r.name = 'FFJ and WCA Flowering Boost' AND d.name LIKE 'FFJ%'
      AND NOT EXISTS (
          SELECT 1 FROM gr33nnaturalfarming.recipe_input_components c
          WHERE c.application_recipe_id = r.id AND c.input_definition_id = d.id
      );

    INSERT INTO gr33nnaturalfarming.recipe_input_components
        (application_recipe_id, input_definition_id, part_value, part_unit_id, notes)
    SELECT r.id, d.id, 0.5, u.id, 'WCA contribution'
    FROM gr33nnaturalfarming.application_recipes r
    CROSS JOIN gr33nnaturalfarming.input_definitions d
    CROSS JOIN gr33ncore.units u
    WHERE r.farm_id = p_farm_id AND d.farm_id = p_farm_id AND u.name = 'decimal_fraction'
      AND r.name = 'FFJ and WCA Flowering Boost' AND d.name LIKE 'WCA%'
      AND NOT EXISTS (
          SELECT 1 FROM gr33nnaturalfarming.recipe_input_components c
          WHERE c.application_recipe_id = r.id AND c.input_definition_id = d.id
      );

    INSERT INTO gr33nfertigation.reservoirs
        (farm_id, zone_id, name, description, capacity_liters, current_volume_liters, status)
    SELECT
        p_farm_id,
        z.id,
        'Main Nutrient Reservoir',
        'Primary fertigation reservoir (18/6 veg JLF+JMS program).',
        500.00,
        320.00,
        'ready'::gr33nfertigation.reservoir_status_enum
    FROM gr33ncore.zones z
    WHERE z.farm_id = p_farm_id AND z.name = 'Veg Room' AND z.deleted_at IS NULL
    ON CONFLICT (farm_id, name) DO NOTHING;

    INSERT INTO gr33nfertigation.reservoirs
        (farm_id, zone_id, name, description, capacity_liters, current_volume_liters, status)
    SELECT
        p_farm_id,
        z.id,
        'Flower Nutrient Reservoir',
        'Dedicated tank for 12/12 flower feeding (FFJ+WCA-style program).',
        400.00,
        220.00,
        'ready'::gr33nfertigation.reservoir_status_enum
    FROM gr33ncore.zones z
    WHERE z.farm_id = p_farm_id AND z.name = 'Flower Room' AND z.deleted_at IS NULL
    ON CONFLICT (farm_id, name) DO NOTHING;

    INSERT INTO gr33nfertigation.reservoirs
        (farm_id, zone_id, name, description, capacity_liters, current_volume_liters, status)
    SELECT
        p_farm_id,
        z.id,
        'Outdoor Drench Tank',
        'JLF soil drench tank for outdoor garden beds. Fill-and-apply, no recirculation.',
        200.00,
        150.00,
        'ready'::gr33nfertigation.reservoir_status_enum
    FROM gr33ncore.zones z
    WHERE z.farm_id = p_farm_id AND z.name = 'Outdoor Garden' AND z.deleted_at IS NULL
    ON CONFLICT (farm_id, name) DO NOTHING;

    INSERT INTO gr33nfertigation.ec_targets
        (farm_id, zone_id, growth_stage, ec_min_mscm, ec_max_mscm, ph_min, ph_max, rationale)
    SELECT
        p_farm_id,
        z.id,
        gs.stage::gr33nfertigation.growth_stage_enum,
        gs.ec_min,
        gs.ec_max,
        gs.ph_min,
        gs.ph_max,
        'Bootstrap template baseline'
    FROM gr33ncore.zones z
    JOIN (
        VALUES
            ('seedling',     0.5::numeric, 1.2::numeric, 5.8::numeric, 6.6::numeric),
            ('early_veg',    1.0::numeric, 1.8::numeric, 5.8::numeric, 6.6::numeric),
            ('late_veg',     1.4::numeric, 2.2::numeric, 5.8::numeric, 6.6::numeric),
            ('transition',   1.6::numeric, 2.4::numeric, 5.8::numeric, 6.6::numeric),
            ('early_flower', 1.6::numeric, 2.4::numeric, 5.8::numeric, 6.6::numeric),
            ('mid_flower',   1.8::numeric, 2.6::numeric, 5.8::numeric, 6.6::numeric),
            ('late_flower',  1.6::numeric, 2.4::numeric, 5.8::numeric, 6.6::numeric),
            ('flush',        0.0::numeric, 0.5::numeric, 5.8::numeric, 6.8::numeric)
    ) AS gs(stage, ec_min, ec_max, ph_min, ph_max) ON TRUE
    WHERE z.farm_id = p_farm_id
      AND z.name IN ('Veg Room', 'Flower Room', 'Outdoor Garden')
      AND z.deleted_at IS NULL
    ON CONFLICT (farm_id, zone_id, growth_stage) DO NOTHING;

    INSERT INTO gr33nfertigation.programs
        (farm_id, name, description, application_recipe_id, reservoir_id, target_zone_id, schedule_id,
         ec_target_id, total_volume_liters, run_duration_seconds, ec_trigger_low, ph_trigger_low, ph_trigger_high, is_active)
    SELECT
        p_farm_id,
        'Veg Daily JLF Program',
        'Template: daily veg fertigation using JLF + JMS combined drench recipe.',
        r.id,
        rv.id,
        z.id,
        s.id,
        et.id,
        120.000,
        900,
        1.200,
        5.8,
        6.8,
        TRUE
    FROM gr33ncore.zones z
    JOIN gr33nnaturalfarming.application_recipes r
        ON r.farm_id = p_farm_id AND r.name = 'JLF and JMS Combined Drench' AND r.deleted_at IS NULL
    JOIN gr33ncore.schedules s
        ON s.farm_id = p_farm_id AND s.name = 'Water Late Veg Daily'
    JOIN gr33nfertigation.reservoirs rv
        ON rv.farm_id = p_farm_id AND rv.name = 'Main Nutrient Reservoir'
    JOIN gr33nfertigation.ec_targets et
        ON et.farm_id = p_farm_id AND et.zone_id = z.id
       AND et.growth_stage = 'late_veg'::gr33nfertigation.growth_stage_enum
    WHERE z.farm_id = p_farm_id AND z.name = 'Veg Room' AND z.deleted_at IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33nfertigation.programs p
          WHERE p.farm_id = p_farm_id AND p.name = 'Veg Daily JLF Program' AND p.deleted_at IS NULL
      );

    INSERT INTO gr33nfertigation.programs
        (farm_id, name, description, application_recipe_id, reservoir_id, target_zone_id, schedule_id,
         ec_target_id, total_volume_liters, run_duration_seconds, ec_trigger_low, ph_trigger_low, ph_trigger_high, is_active)
    SELECT
        p_farm_id,
        'Flower Daily FFJ+WCA Program',
        '12/12 flower room: scheduled irrigations with FFJ+WCA flowering recipe.',
        r.id,
        rv.id,
        z.id,
        s.id,
        et.id,
        95.000,
        840,
        1.400,
        5.8,
        6.8,
        TRUE
    FROM gr33ncore.zones z
    JOIN gr33nnaturalfarming.application_recipes r
        ON r.farm_id = p_farm_id AND r.name = 'FFJ and WCA Flowering Boost' AND r.deleted_at IS NULL
    JOIN gr33ncore.schedules s
        ON s.farm_id = p_farm_id AND s.name = 'Water Early Flower Daily'
    JOIN gr33nfertigation.reservoirs rv
        ON rv.farm_id = p_farm_id AND rv.name = 'Flower Nutrient Reservoir'
    JOIN gr33nfertigation.ec_targets et
        ON et.farm_id = p_farm_id AND et.zone_id = z.id
       AND et.growth_stage = 'early_flower'::gr33nfertigation.growth_stage_enum
    WHERE z.farm_id = p_farm_id AND z.name = 'Flower Room' AND z.deleted_at IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33nfertigation.programs p
          WHERE p.farm_id = p_farm_id AND p.name = 'Flower Daily FFJ+WCA Program' AND p.deleted_at IS NULL
      );

    INSERT INTO gr33nfertigation.programs
        (farm_id, name, description, application_recipe_id, reservoir_id, target_zone_id, schedule_id,
         total_volume_liters, run_duration_seconds, ec_trigger_low, ph_trigger_low, ph_trigger_high, is_active)
    SELECT
        p_farm_id,
        'Outdoor JLF Soil Drench',
        'Daily outdoor drench: JLF General 1:20 via drench tank.',
        r.id,
        rv.id,
        z.id,
        s.id,
        60.000,
        600,
        0.800,
        5.8,
        7.0,
        TRUE
    FROM gr33ncore.zones z
    JOIN gr33nnaturalfarming.application_recipes r
        ON r.farm_id = p_farm_id AND r.name = 'JLF General Soil Drench' AND r.deleted_at IS NULL
    JOIN gr33ncore.schedules s
        ON s.farm_id = p_farm_id AND s.name = 'Water Outdoor Garden Daily'
    JOIN gr33nfertigation.reservoirs rv
        ON rv.farm_id = p_farm_id AND rv.name = 'Outdoor Drench Tank'
    WHERE z.farm_id = p_farm_id AND z.name = 'Outdoor Garden' AND z.deleted_at IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33nfertigation.programs p
          WHERE p.farm_id = p_farm_id AND p.name = 'Outdoor JLF Soil Drench' AND p.deleted_at IS NULL
      );

    INSERT INTO gr33nnaturalfarming.input_batches (
        farm_id, input_definition_id, batch_identifier, creation_start_date, actual_ready_date,
        quantity_produced, quantity_unit_id, current_quantity_remaining, status,
        storage_location, observations_notes
    )
    SELECT
        p_farm_id,
        d.id,
        v.batch_identifier,
        v.started::date,
        v.ready::date,
        v.qty_l,
        u.id,
        v.remaining_l,
        'ready_for_use'::gr33nnaturalfarming.input_batch_status_enum,
        v.location,
        v.notes
    FROM (VALUES
        ('TPL-JLF-GEN-001',  DATE '2026-01-10', DATE '2026-01-24', 45.0::numeric, 38.0::numeric,
         'Veg Room — concentrate shelf',
         'Starter JLF lot for reservoir mixes.'),
        ('TPL-JMS-001',      DATE '2026-02-01', DATE '2026-02-05', 25.0::numeric, 22.0::numeric,
         'Veg Room — fridge',
         'Starter JMS concentrate.'),
        ('TPL-FFJ-001',      DATE '2026-02-15', DATE '2026-03-01', 8.0::numeric, 6.5::numeric,
         'Flower Room — fridge',
         'Starter FFJ for flower tank.'),
        ('TPL-WCA-001',      DATE '2026-01-20', DATE '2026-02-10', 12.0::numeric, 10.0::numeric,
         'Flower Room — bench',
         'Starter WCA extract.')
    ) AS v(batch_identifier, started, ready, qty_l, remaining_l, location, notes)
    JOIN gr33ncore.units u ON u.name = 'liter'
    JOIN gr33nnaturalfarming.input_definitions d ON d.farm_id = p_farm_id AND d.deleted_at IS NULL
     AND (
       (v.batch_identifier = 'TPL-JLF-GEN-001' AND d.name LIKE 'JLF General%')
       OR (v.batch_identifier = 'TPL-JMS-001' AND d.name LIKE 'JMS%')
       OR (v.batch_identifier = 'TPL-FFJ-001' AND d.name LIKE 'FFJ%')
       OR (v.batch_identifier = 'TPL-WCA-001' AND d.name LIKE 'WCA%')
     )
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33nnaturalfarming.input_batches b
        WHERE b.farm_id = p_farm_id AND b.batch_identifier = v.batch_identifier AND b.deleted_at IS NULL
    );

    INSERT INTO gr33ncore.tasks (farm_id, zone_id, schedule_id, title, description, task_type, status, priority, due_date)
    SELECT
        p_farm_id,
        z.id,
        s.id,
        'Refresh veg reservoir mix (18/6)',
        'Main Nutrient Reservoir: JLF+JMS batch. Hit EC 1.4–2.2 late veg.',
        'jadam_mix',
        'todo'::gr33ncore.task_status_enum,
        2,
        CURRENT_DATE + 1
    FROM gr33ncore.zones z
    JOIN gr33ncore.schedules s ON s.farm_id = p_farm_id AND s.name = 'Water Late Veg Daily'
    WHERE z.farm_id = p_farm_id AND z.name = 'Veg Room' AND z.deleted_at IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.tasks t
          WHERE t.farm_id = p_farm_id AND t.deleted_at IS NULL
            AND t.title = 'Refresh veg reservoir mix (18/6)'
      );

    INSERT INTO gr33ncore.tasks (farm_id, zone_id, schedule_id, title, description, task_type, status, priority, due_date)
    SELECT
        p_farm_id,
        z.id,
        s.id,
        'Refresh flower reservoir mix (12/12)',
        'Flower Nutrient Reservoir: FFJ+WCA batch for early flower.',
        'jadam_mix',
        'todo'::gr33ncore.task_status_enum,
        2,
        CURRENT_DATE + 1
    FROM gr33ncore.zones z
    JOIN gr33ncore.schedules s ON s.farm_id = p_farm_id AND s.name = 'Water Early Flower Daily'
    WHERE z.farm_id = p_farm_id AND z.name = 'Flower Room' AND z.deleted_at IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.tasks t
          WHERE t.farm_id = p_farm_id AND t.deleted_at IS NULL
            AND t.title = 'Refresh flower reservoir mix (12/12)'
      );

    INSERT INTO gr33ncore.tasks (farm_id, zone_id, schedule_id, title, description, task_type, status, priority, due_date)
    SELECT
        p_farm_id,
        z.id,
        s.id,
        'Refresh outdoor drench tank',
        'Outdoor Drench Tank: top up JLF 1:20 mix before morning schedule.',
        'jadam_mix',
        'todo'::gr33ncore.task_status_enum,
        1,
        CURRENT_DATE + 1
    FROM gr33ncore.zones z
    JOIN gr33ncore.schedules s ON s.farm_id = p_farm_id AND s.name = 'Water Outdoor Garden Daily'
    WHERE z.farm_id = p_farm_id AND z.name = 'Outdoor Garden' AND z.deleted_at IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.tasks t
          WHERE t.farm_id = p_farm_id AND t.deleted_at IS NULL
            AND t.title = 'Refresh outdoor drench tank'
      );

    INSERT INTO gr33nfertigation.fertigation_events
        (farm_id, program_id, reservoir_id, zone_id, applied_at, growth_stage,
         volume_applied_liters, run_duration_seconds, ec_before_mscm, ec_after_mscm,
         ph_before, ph_after, trigger_source, notes)
    SELECT
        p_farm_id,
        p.id,
        rv.id,
        z.id,
        TIMESTAMPTZ '2026-03-01 08:00:00+00',
        'late_veg'::gr33nfertigation.growth_stage_enum,
        112.500,
        860,
        1.150,
        1.720,
        6.05,
        6.22,
        'schedule_cron'::gr33nfertigation.program_trigger_enum,
        'Template veg fertigation event (links program + reservoir).'
    FROM gr33ncore.zones z
    JOIN gr33nfertigation.programs p
        ON p.farm_id = p_farm_id AND p.name = 'Veg Daily JLF Program' AND p.deleted_at IS NULL
    JOIN gr33nfertigation.reservoirs rv
        ON rv.farm_id = p_farm_id AND rv.name = 'Main Nutrient Reservoir'
    WHERE z.farm_id = p_farm_id AND z.name = 'Veg Room'
      AND NOT EXISTS (
          SELECT 1 FROM gr33nfertigation.fertigation_events fe
          WHERE fe.farm_id = p_farm_id
            AND fe.applied_at = TIMESTAMPTZ '2026-03-01 08:00:00+00'
            AND fe.zone_id = z.id
      );

    INSERT INTO gr33nfertigation.fertigation_events
        (farm_id, program_id, reservoir_id, zone_id, applied_at, growth_stage,
         volume_applied_liters, run_duration_seconds, ec_before_mscm, ec_after_mscm,
         ph_before, ph_after, trigger_source, notes)
    SELECT
        p_farm_id,
        p.id,
        rv.id,
        z.id,
        TIMESTAMPTZ '2026-03-05 08:00:00+00',
        'early_flower'::gr33nfertigation.growth_stage_enum,
        88.000,
        800,
        1.550,
        1.920,
        6.00,
        6.18,
        'schedule_cron'::gr33nfertigation.program_trigger_enum,
        'Template flower fertigation event (links program + reservoir).'
    FROM gr33ncore.zones z
    JOIN gr33nfertigation.programs p
        ON p.farm_id = p_farm_id AND p.name = 'Flower Daily FFJ+WCA Program' AND p.deleted_at IS NULL
    JOIN gr33nfertigation.reservoirs rv
        ON rv.farm_id = p_farm_id AND rv.name = 'Flower Nutrient Reservoir'
    WHERE z.farm_id = p_farm_id AND z.name = 'Flower Room'
      AND NOT EXISTS (
          SELECT 1 FROM gr33nfertigation.fertigation_events fe
          WHERE fe.farm_id = p_farm_id
            AND fe.applied_at = TIMESTAMPTZ '2026-03-05 08:00:00+00'
            AND fe.zone_id = z.id
      );

    -- Mixing history: reservoir ↔ program ↔ inventory batches
        SELECT id INTO r_veg FROM gr33nfertigation.reservoirs WHERE farm_id = p_farm_id AND name = 'Main Nutrient Reservoir' LIMIT 1;
        SELECT id INTO r_flower FROM gr33nfertigation.reservoirs WHERE farm_id = p_farm_id AND name = 'Flower Nutrient Reservoir' LIMIT 1;
        SELECT id INTO p_veg FROM gr33nfertigation.programs WHERE farm_id = p_farm_id AND name = 'Veg Daily JLF Program' AND deleted_at IS NULL LIMIT 1;
        SELECT id INTO p_flower FROM gr33nfertigation.programs WHERE farm_id = p_farm_id AND name = 'Flower Daily FFJ+WCA Program' AND deleted_at IS NULL LIMIT 1;

        SELECT id INTO i_jlf FROM gr33nnaturalfarming.input_definitions WHERE farm_id = p_farm_id AND name LIKE 'JLF General%' AND deleted_at IS NULL LIMIT 1;
        SELECT id INTO i_jms FROM gr33nnaturalfarming.input_definitions WHERE farm_id = p_farm_id AND name LIKE 'JMS%' AND deleted_at IS NULL LIMIT 1;
        SELECT id INTO i_ffj FROM gr33nnaturalfarming.input_definitions WHERE farm_id = p_farm_id AND name LIKE 'FFJ%' AND deleted_at IS NULL LIMIT 1;
        SELECT id INTO i_wca FROM gr33nnaturalfarming.input_definitions WHERE farm_id = p_farm_id AND name LIKE 'WCA%' AND deleted_at IS NULL LIMIT 1;

        SELECT id INTO b_jlf FROM gr33nnaturalfarming.input_batches WHERE farm_id = p_farm_id AND batch_identifier = 'TPL-JLF-GEN-001' AND deleted_at IS NULL LIMIT 1;
        SELECT id INTO b_jms FROM gr33nnaturalfarming.input_batches WHERE farm_id = p_farm_id AND batch_identifier = 'TPL-JMS-001' AND deleted_at IS NULL LIMIT 1;
        SELECT id INTO b_ffj FROM gr33nnaturalfarming.input_batches WHERE farm_id = p_farm_id AND batch_identifier = 'TPL-FFJ-001' AND deleted_at IS NULL LIMIT 1;
        SELECT id INTO b_wca FROM gr33nnaturalfarming.input_batches WHERE farm_id = p_farm_id AND batch_identifier = 'TPL-WCA-001' AND deleted_at IS NULL LIMIT 1;

        SELECT id INTO r_outdoor FROM gr33nfertigation.reservoirs WHERE farm_id = p_farm_id AND name = 'Outdoor Drench Tank' LIMIT 1;
        SELECT id INTO p_outdoor FROM gr33nfertigation.programs WHERE farm_id = p_farm_id AND name = 'Outdoor JLF Soil Drench' AND deleted_at IS NULL LIMIT 1;

        IF r_veg IS NOT NULL AND NOT EXISTS (
            SELECT 1 FROM gr33nfertigation.mixing_events me
            WHERE me.farm_id = p_farm_id AND me.reservoir_id = r_veg AND me.notes LIKE '%[tpl:veg-mix]%'
        ) THEN
            INSERT INTO gr33nfertigation.mixing_events (
                farm_id, reservoir_id, program_id, mixed_at,
                water_volume_liters, water_source, water_ec_mscm, water_ph,
                final_ec_mscm, final_ph, ec_target_met,
                notes
            ) VALUES (
                p_farm_id, r_veg, p_veg, TIMESTAMPTZ '2026-03-01 07:15:00+00',
                300.0, 'RO + rain blend', 0.05, 6.50,
                1.65, 6.12, TRUE,
                'Template veg mix — JLF+JMS style before irrigations. [tpl:veg-mix]'
            ) RETURNING id INTO mix_veg;

            IF i_jlf IS NOT NULL THEN
                INSERT INTO gr33nfertigation.mixing_event_components
                    (mixing_event_id, input_definition_id, input_batch_id, volume_added_ml, dilution_ratio, notes)
                VALUES (mix_veg, i_jlf, b_jlf, 15000.000, 'concentrate to ~1:20 tank', 'Draw from Inventory batch TPL-JLF-GEN-001');
            END IF;
            IF i_jms IS NOT NULL THEN
                INSERT INTO gr33nfertigation.mixing_event_components
                    (mixing_event_id, input_definition_id, input_batch_id, volume_added_ml, dilution_ratio, notes)
                VALUES (mix_veg, i_jms, b_jms, 600.000, '1:500 in tank', 'Draw from Inventory batch TPL-JMS-001');
            END IF;
        END IF;

        IF r_flower IS NOT NULL AND NOT EXISTS (
            SELECT 1 FROM gr33nfertigation.mixing_events me
            WHERE me.farm_id = p_farm_id AND me.reservoir_id = r_flower AND me.notes LIKE '%[tpl:flower-mix]%'
        ) THEN
            INSERT INTO gr33nfertigation.mixing_events (
                farm_id, reservoir_id, program_id, mixed_at,
                water_volume_liters, water_source, water_ec_mscm, water_ph,
                final_ec_mscm, final_ph, ec_target_met,
                notes
            ) VALUES (
                p_farm_id, r_flower, p_flower, TIMESTAMPTZ '2026-03-05 07:20:00+00',
                220.0, 'RO', 0.04, 6.45,
                1.78, 6.05, TRUE,
                'Template flower mix — FFJ+WCA oriented batch. [tpl:flower-mix]'
            ) RETURNING id INTO mix_fl;

            IF i_ffj IS NOT NULL THEN
                INSERT INTO gr33nfertigation.mixing_event_components
                    (mixing_event_id, input_definition_id, input_batch_id, volume_added_ml, dilution_ratio, notes)
                VALUES (mix_fl, i_ffj, b_ffj, 2200.000, 'light feed', 'Inventory TPL-FFJ-001');
            END IF;
            IF i_wca IS NOT NULL THEN
                INSERT INTO gr33nfertigation.mixing_event_components
                    (mixing_event_id, input_definition_id, input_batch_id, volume_added_ml, dilution_ratio, notes)
                VALUES (mix_fl, i_wca, b_wca, 800.000, '1:1000 relative', 'Inventory TPL-WCA-001');
            END IF;
        END IF;

    UPDATE gr33nfertigation.fertigation_events fe SET mixing_event_id = me.id
    FROM gr33nfertigation.mixing_events me
    JOIN gr33nfertigation.reservoirs rv ON me.reservoir_id = rv.id AND rv.farm_id = p_farm_id
    WHERE fe.farm_id = p_farm_id
      AND fe.mixing_event_id IS NULL
      AND fe.applied_at = TIMESTAMPTZ '2026-03-01 08:00:00+00'
      AND rv.name = 'Main Nutrient Reservoir'
      AND me.farm_id = p_farm_id
      AND me.notes LIKE '%[tpl:veg-mix]%';

    UPDATE gr33nfertigation.fertigation_events fe
    SET mixing_event_id = me.id
    FROM gr33nfertigation.mixing_events me
    JOIN gr33nfertigation.reservoirs rv ON me.reservoir_id = rv.id AND rv.farm_id = p_farm_id
    WHERE fe.farm_id = p_farm_id
      AND fe.mixing_event_id IS NULL
      AND fe.applied_at = TIMESTAMPTZ '2026-03-05 08:00:00+00'
      AND rv.name = 'Flower Nutrient Reservoir'
      AND me.farm_id = p_farm_id
      AND me.notes LIKE '%[tpl:flower-mix]%';

    -- Outdoor fertigation event
    INSERT INTO gr33nfertigation.fertigation_events
        (farm_id, program_id, reservoir_id, zone_id, applied_at, growth_stage,
         volume_applied_liters, run_duration_seconds, ec_before_mscm, ec_after_mscm,
         ph_before, ph_after, trigger_source, notes)
    SELECT
        p_farm_id, p_outdoor, r_outdoor, z.id,
        TIMESTAMPTZ '2026-03-08 07:15:00+00',
        'early_veg'::gr33nfertigation.growth_stage_enum,
        55.000, 580, 0.350, 0.950, 6.40, 6.55,
        'schedule_cron'::gr33nfertigation.program_trigger_enum,
        'Template outdoor JLF soil drench event.'
    FROM gr33ncore.zones z
    WHERE z.farm_id = p_farm_id AND z.name = 'Outdoor Garden'
      AND r_outdoor IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33nfertigation.fertigation_events fe
          WHERE fe.farm_id = p_farm_id AND fe.applied_at = TIMESTAMPTZ '2026-03-08 07:15:00+00' AND fe.zone_id = z.id
      );

    -- Outdoor mixing event
    IF r_outdoor IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33nfertigation.mixing_events me
        WHERE me.farm_id = p_farm_id AND me.reservoir_id = r_outdoor AND me.notes LIKE '%[tpl:outdoor-mix]%'
    ) THEN
        INSERT INTO gr33nfertigation.mixing_events (
            farm_id, reservoir_id, program_id, mixed_at,
            water_volume_liters, water_source, water_ec_mscm, water_ph,
            final_ec_mscm, final_ph, ec_target_met, notes
        ) VALUES (
            p_farm_id, r_outdoor, p_outdoor, TIMESTAMPTZ '2026-03-08 06:45:00+00',
            150.0, 'Rain barrel', 0.03, 6.80,
            0.92, 6.55, TRUE,
            'Template outdoor JLF drench mix. [tpl:outdoor-mix]'
        ) RETURNING id INTO mix_out;

        IF i_jlf IS NOT NULL THEN
            INSERT INTO gr33nfertigation.mixing_event_components
                (mixing_event_id, input_definition_id, input_batch_id, volume_added_ml, dilution_ratio, notes)
            VALUES (mix_out, i_jlf, b_jlf, 7500.000, '1:20 in drench tank', 'Inventory TPL-JLF-GEN-001');
        END IF;
    END IF;

    UPDATE gr33nfertigation.fertigation_events fe
    SET mixing_event_id = me.id
    FROM gr33nfertigation.mixing_events me
    JOIN gr33nfertigation.reservoirs rv ON me.reservoir_id = rv.id AND rv.farm_id = p_farm_id
    WHERE fe.farm_id = p_farm_id
      AND fe.mixing_event_id IS NULL
      AND fe.applied_at = TIMESTAMPTZ '2026-03-08 07:15:00+00'
      AND rv.name = 'Outdoor Drench Tank'
      AND me.farm_id = p_farm_id
      AND me.notes LIKE '%[tpl:outdoor-mix]%';

    -- Crop cycles
    INSERT INTO gr33nfertigation.crop_cycles
        (farm_id, zone_id, name, strain_or_variety, current_stage, is_active, started_at, cycle_notes)
    SELECT p_farm_id, z.id, 'Veg canopy (18/6)', 'Generic photoperiod',
        'late_veg'::gr33nfertigation.growth_stage_enum, TRUE, CURRENT_DATE - 35,
        'Template: veg fertigation + 18/6 light.'
    FROM gr33ncore.zones z
    WHERE z.farm_id = p_farm_id AND z.name = 'Veg Room' AND z.deleted_at IS NULL
      AND NOT EXISTS (SELECT 1 FROM gr33nfertigation.crop_cycles cc WHERE cc.zone_id = z.id AND cc.is_active = TRUE);

    INSERT INTO gr33nfertigation.crop_cycles
        (farm_id, zone_id, name, strain_or_variety, current_stage, is_active, started_at, cycle_notes)
    SELECT p_farm_id, z.id, 'Flower run (12/12)', 'Generic photoperiod',
        'early_flower'::gr33nfertigation.growth_stage_enum, TRUE, CURRENT_DATE - 14,
        'Template: flower fertigation + 12/12 light.'
    FROM gr33ncore.zones z
    WHERE z.farm_id = p_farm_id AND z.name = 'Flower Room' AND z.deleted_at IS NULL
      AND NOT EXISTS (SELECT 1 FROM gr33nfertigation.crop_cycles cc WHERE cc.zone_id = z.id AND cc.is_active = TRUE);

    INSERT INTO gr33nfertigation.crop_cycles
        (farm_id, zone_id, name, strain_or_variety, current_stage, is_active, started_at, cycle_notes)
    SELECT p_farm_id, z.id, 'Outdoor raised beds — spring', 'Mixed greens / herbs',
        'early_veg'::gr33nfertigation.growth_stage_enum, TRUE, CURRENT_DATE - 21,
        'Template: outdoor JADAM soil drench + natural light.'
    FROM gr33ncore.zones z
    WHERE z.farm_id = p_farm_id AND z.name = 'Outdoor Garden' AND z.deleted_at IS NULL
      AND NOT EXISTS (SELECT 1 FROM gr33nfertigation.crop_cycles cc WHERE cc.zone_id = z.id AND cc.is_active = TRUE);

    -- Set primary_program_id on crop cycles
    UPDATE gr33nfertigation.crop_cycles cc SET primary_program_id = p.id
    FROM gr33nfertigation.programs p
    WHERE cc.farm_id = p_farm_id AND cc.primary_program_id IS NULL AND cc.is_active = TRUE
      AND p.farm_id = p_farm_id AND p.deleted_at IS NULL
      AND (
        (cc.name = 'Veg canopy (18/6)'              AND p.name = 'Veg Daily JLF Program')
        OR (cc.name = 'Flower run (12/12)'           AND p.name = 'Flower Daily FFJ+WCA Program')
        OR (cc.name = 'Outdoor raised beds — spring' AND p.name = 'Outdoor JLF Soil Drench')
      );

    -- Link fertigation events to crop cycles by zone
    UPDATE gr33nfertigation.fertigation_events fe SET crop_cycle_id = cc.id
    FROM gr33nfertigation.crop_cycles cc
    WHERE fe.farm_id = p_farm_id AND fe.crop_cycle_id IS NULL
      AND cc.farm_id = p_farm_id AND cc.is_active = TRUE AND fe.zone_id = cc.zone_id;
END;
$$;
