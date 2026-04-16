-- WS9 / Phase 15: optional farm bootstrap templates (idempotent per farm + template key).
-- Apply via SELECT gr33ncore.apply_farm_bootstrap_template(farm_id, template_key);

CREATE TABLE IF NOT EXISTS gr33ncore.farm_bootstrap_applications (
    id                BIGSERIAL PRIMARY KEY,
    farm_id           BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    template_key      TEXT   NOT NULL,
    template_version  INT    NOT NULL DEFAULT 1,
    applied_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_farm_bootstrap_farm_template UNIQUE (farm_id, template_key)
);

CREATE INDEX IF NOT EXISTS idx_farm_bootstrap_farm
    ON gr33ncore.farm_bootstrap_applications (farm_id);

-- Internal: populate starter data for jadam_indoor_photoperiod_v1 (safe to re-run; uses NOT EXISTS / ON CONFLICT).
CREATE OR REPLACE FUNCTION gr33ncore._bootstrap_jadam_indoor_photoperiod_v1(p_farm_id BIGINT, p_tz TEXT)
RETURNS void
LANGUAGE plpgsql
AS $$
BEGIN
    INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
    SELECT p_farm_id, x.n, x.d, x.t
    FROM (
        VALUES
            ('Seedling Room', 'Germination and early seedling stage. High humidity, gentle light.', 'indoor'),
            ('Veg Room', 'Vegetative growth stage. 18/6 light, moderate feeding.', 'indoor'),
            ('Flower Room', 'Flowering and fruiting stage. 12/12 light, flowering feed program.', 'indoor'),
            ('Outdoor Beds', 'Outdoor raised beds. Natural light. JADAM soil program.', 'outdoor')
    ) AS x(n, d, t)
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33ncore.zones z
        WHERE z.farm_id = p_farm_id AND z.name = x.n AND z.deleted_at IS NULL
    );

    -- Light + irrigation schedules (same crons as master_seed; timezone from farm)
    INSERT INTO gr33ncore.schedules (farm_id, name, description, schedule_type, cron_expression, timezone, is_active)
    SELECT p_farm_id, x.n, x.d, x.st, x.cron, p_tz, FALSE
    FROM (
        VALUES
            ('Light ON 24/0 Continuous', 'Lights always on. Seedling / autoflower.', 'lighting', '0 0 * * *'),
            ('Light ON 18/6 Veg', 'Lights on at 06:00 for vegetative growth.', 'lighting', '0 6 * * *'),
            ('Light OFF 18/6 Veg', 'Lights off at midnight (6h dark).', 'lighting', '0 0 * * *'),
            ('Light ON 12/12 Flower', 'Lights on at 06:00 for flowering photoperiod.', 'lighting', '0 6 * * *'),
            ('Light OFF 12/12 Flower', 'Lights off at 18:00 (12h dark).', 'lighting', '0 18 * * *'),
            ('Water Seedling Daily Light', 'Morning light irrigation seedling room.', 'irrigation', '0 8 * * *'),
            ('Water Late Veg Daily', 'Late veg daily irrigation (pairs with veg fertigation program).', 'irrigation', '0 8 * * *'),
            ('Water Early Flower Daily', 'Early flower daily irrigation.', 'irrigation', '0 8 * * *')
    ) AS x(n, d, st, cron)
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33ncore.schedules s
        WHERE s.farm_id = p_farm_id AND s.name = x.n
    );

    -- Core JADAM inputs (minimal set to support veg program recipe)
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
             'Ferment; extract serum.', 'Refrigerate preserved.', 'Use golden layer only.', 'JADAM OF')
    ) AS x(name, cat, descr, ting, prep, store, safe, ref)
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33nnaturalfarming.input_definitions d
        WHERE d.farm_id = p_farm_id AND d.name = x.name AND d.deleted_at IS NULL
    );

    -- Application recipes
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
             ARRAY['All crops']::text[], ARRAY['All stages']::text[])
    ) AS x(n, descr, ttype, dil, inst, freq, cats, stages)
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33nnaturalfarming.application_recipes r
        WHERE r.farm_id = p_farm_id AND r.name = x.n AND r.deleted_at IS NULL
    );

    -- Recipe components (decimal_fraction unit)
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

    -- Main veg reservoir
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

    -- EC / pH targets (subset of growth stages for indoor zones)
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
            ('early_flower', 1.6::numeric, 2.4::numeric, 5.8::numeric, 6.6::numeric),
            ('mid_flower',   1.8::numeric, 2.6::numeric, 5.8::numeric, 6.6::numeric)
    ) AS gs(stage, ec_min, ec_max, ph_min, ph_max) ON TRUE
    WHERE z.farm_id = p_farm_id
      AND z.name IN ('Seedling Room', 'Veg Room', 'Flower Room')
      AND z.deleted_at IS NULL
    ON CONFLICT (farm_id, zone_id, growth_stage) DO NOTHING;

    -- Veg fertigation program
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

    INSERT INTO gr33ncore.tasks (farm_id, zone_id, title, description, task_type, status, priority, due_date)
    SELECT
        p_farm_id,
        z.id,
        'Protocol: refresh veg reservoir mix (18/6)',
        'Template task — mix JLF+JMS batch for Main Nutrient Reservoir; adjust EC to late-veg target.',
        'jadam_mix',
        'todo'::gr33ncore.task_status_enum,
        2,
        CURRENT_DATE + 1
    FROM gr33ncore.zones z
    WHERE z.farm_id = p_farm_id AND z.name = 'Veg Room' AND z.deleted_at IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.tasks t
          WHERE t.farm_id = p_farm_id AND t.deleted_at IS NULL
            AND t.title = 'Protocol: refresh veg reservoir mix (18/6)'
      );
END;
$$;

CREATE OR REPLACE FUNCTION gr33ncore.apply_farm_bootstrap_template(p_farm_id BIGINT, p_template TEXT)
RETURNS JSONB
LANGUAGE plpgsql
AS $$
DECLARE
    v_tz   TEXT;
    v_norm TEXT;
BEGIN
    IF p_template IS NULL OR btrim(p_template) = '' OR lower(btrim(p_template)) IN ('none', 'blank') THEN
        RETURN jsonb_build_object('skipped', TRUE);
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM gr33ncore.farms f
        WHERE f.id = p_farm_id AND f.deleted_at IS NULL
    ) THEN
        RETURN jsonb_build_object('error', 'farm_not_found', 'farm_id', p_farm_id);
    END IF;

    SELECT timezone INTO v_tz FROM gr33ncore.farms WHERE id = p_farm_id LIMIT 1;
    IF v_tz IS NULL OR btrim(v_tz) = '' THEN
        v_tz := 'UTC';
    END IF;

    v_norm := lower(btrim(p_template));

    IF v_norm = 'jadam_indoor_photoperiod_v1' THEN
        IF EXISTS (
            SELECT 1 FROM gr33ncore.farm_bootstrap_applications a
            WHERE a.farm_id = p_farm_id AND a.template_key = 'jadam_indoor_photoperiod_v1'
        ) THEN
            RETURN jsonb_build_object(
                'applied', FALSE,
                'already_applied', TRUE,
                'template', 'jadam_indoor_photoperiod_v1',
                'version', 1
            );
        END IF;

        PERFORM gr33ncore._bootstrap_jadam_indoor_photoperiod_v1(p_farm_id, v_tz);

        BEGIN
            INSERT INTO gr33ncore.farm_bootstrap_applications (farm_id, template_key, template_version)
            VALUES (p_farm_id, 'jadam_indoor_photoperiod_v1', 1);
        EXCEPTION
            WHEN unique_violation THEN
                RETURN jsonb_build_object(
                    'applied', FALSE,
                    'already_applied', TRUE,
                    'template', 'jadam_indoor_photoperiod_v1',
                    'version', 1
                );
        END;

        RETURN jsonb_build_object(
            'applied', TRUE,
            'template', 'jadam_indoor_photoperiod_v1',
            'version', 1
        );
    END IF;

    RETURN jsonb_build_object('error', 'unknown_template', 'template', p_template);
END;
$$;
