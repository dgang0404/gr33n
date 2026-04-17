-- ============================================================
-- Phase 20.8 WS5 — Bootstrap upgrade
--
-- Patches the Phase 20.5 bootstrap functions so that:
--   * chicken_coop_v1 seeds a gr33nanimals.animal_groups row
--     ('Layer flock', chicken, count=12, primary_zone_id=<coop>),
--   * small_aquaponics_v1 patches its gr33naquaponics.loops row to
--     carry the fish_tank_zone_id / grow_bed_zone_id typed FKs
--     introduced by Phase 20.95 WS4.
--
-- Idempotent: both patches use NOT EXISTS / UPDATE … WHERE guards
-- so farms that already ran the Phase 20.5 bootstrap can re-run
-- (via gr33ncore.apply_farm_bootstrap_template) with no duplicate
-- rows and no data loss. Farms that haven't adopted the template
-- yet see the new behaviour on first application.
--
-- The outer dispatcher (gr33ncore.apply_farm_bootstrap_template)
-- is unchanged — the `farm_bootstrap_applications` UNIQUE guard
-- still short-circuits re-runs; this migration only rewrites the
-- inner per-template functions.
-- ============================================================

CREATE OR REPLACE FUNCTION gr33ncore._bootstrap_chicken_coop_v1(p_farm_id BIGINT, p_tz TEXT)
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    z_coop   BIGINT;
    dev_id   BIGINT;
    s_water  BIGINT;
    s_feed   BIGINT;
    s_temp   BIGINT;
    a_feed   BIGINT;
    a_fan    BIGINT;
    a_heat   BIGINT;
BEGIN
    INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
    SELECT p_farm_id, x.n, x.d, x.t
    FROM (
        VALUES
            ('Chicken Coop', 'Layer or broiler enclosure — feeders, drinkers, ventilation, and optional heat.', 'indoor')
    ) AS x(n, d, t)
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33ncore.zones z
        WHERE z.farm_id = p_farm_id AND z.name = x.n AND z.deleted_at IS NULL
    );

    SELECT z.id INTO z_coop FROM gr33ncore.zones z
    WHERE z.farm_id = p_farm_id AND z.name = 'Chicken Coop' AND z.deleted_at IS NULL
    LIMIT 1;

    INSERT INTO gr33ncore.devices (farm_id, zone_id, name, device_uid, device_type, status)
    SELECT p_farm_id, z_coop, 'Chicken Coop Pi', 'bootstrap-chicken-coop-' || p_farm_id::TEXT, 'raspberry_pi', 'unknown'::gr33ncore.device_status_enum
    WHERE z_coop IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.devices d
          WHERE d.farm_id = p_farm_id AND d.device_uid = 'bootstrap-chicken-coop-' || p_farm_id::TEXT
      );

    SELECT d.id INTO dev_id FROM gr33ncore.devices d
    WHERE d.farm_id = p_farm_id AND d.device_uid = 'bootstrap-chicken-coop-' || p_farm_id::TEXT
    LIMIT 1;

    INSERT INTO gr33ncore.sensors (farm_id, zone_id, name, sensor_type, unit_id, reading_interval_seconds, config)
    SELECT p_farm_id, z_coop, x.name, x.stype, u.id, x.ivl, '{}'::jsonb
    FROM (VALUES
        ('Coop water level', 'water_level', 'percent', 300),
        ('Coop feed level', 'feed_level', 'percent', 300),
        ('Coop air temperature', 'temperature', 'celsius', 120),
        ('Coop air humidity', 'humidity', 'percent', 120)
    ) AS x(name, stype, uname, ivl)
    JOIN gr33ncore.units u ON u.name = x.uname
    WHERE z_coop IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.sensors s
          WHERE s.farm_id = p_farm_id AND s.name = x.name AND s.deleted_at IS NULL
      );

    INSERT INTO gr33ncore.actuators (farm_id, zone_id, device_id, name, actuator_type)
    SELECT p_farm_id, z_coop, dev_id, x.aname, x.atype
    FROM (VALUES
        ('Coop feeder hopper', 'feeder_hopper'),
        ('Coop water valve', 'water_valve'),
        ('Coop exhaust fan', 'exhaust_fan'),
        ('Coop heat lamp', 'heat_lamp')
    ) AS x(aname, atype)
    WHERE z_coop IS NOT NULL AND dev_id IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.actuators a
          WHERE a.farm_id = p_farm_id AND a.name = x.aname AND a.deleted_at IS NULL
      );

    INSERT INTO gr33ncore.schedules (farm_id, name, description, schedule_type, cron_expression, timezone, is_active)
    SELECT p_farm_id, x.n, x.d, x.st, x.cron, p_tz, FALSE
    FROM (
        VALUES
            ('Coop morning feed', 'Reminder to trigger hopper / hand feed at dawn.', 'feeding', '0 6 * * *'),
            ('Coop evening feed', 'Reminder for second daily feed.', 'feeding', '0 18 * * *')
    ) AS x(n, d, st, cron)
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33ncore.schedules s WHERE s.farm_id = p_farm_id AND s.name = x.n
    );

    INSERT INTO gr33ncore.tasks (farm_id, zone_id, title, description, task_type, status, priority, due_date)
    SELECT p_farm_id, z_coop,
           'Weekly: egg collection and nest check',
           'Template task — collect eggs, refresh bedding near nests, note any sick birds.',
           'inspection', 'todo'::gr33ncore.task_status_enum, 1, CURRENT_DATE + 7
    WHERE z_coop IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.tasks t
          WHERE t.farm_id = p_farm_id AND t.deleted_at IS NULL
            AND t.title = 'Weekly: egg collection and nest check'
      );

    SELECT s.id INTO s_water FROM gr33ncore.sensors s
    WHERE s.farm_id = p_farm_id AND s.name = 'Coop water level' AND s.deleted_at IS NULL LIMIT 1;
    SELECT s.id INTO s_feed FROM gr33ncore.sensors s
    WHERE s.farm_id = p_farm_id AND s.name = 'Coop feed level' AND s.deleted_at IS NULL LIMIT 1;
    SELECT s.id INTO s_temp FROM gr33ncore.sensors s
    WHERE s.farm_id = p_farm_id AND s.name = 'Coop air temperature' AND s.deleted_at IS NULL LIMIT 1;
    SELECT a.id INTO a_feed FROM gr33ncore.actuators a
    WHERE a.farm_id = p_farm_id AND a.name = 'Coop feeder hopper' AND a.deleted_at IS NULL LIMIT 1;
    SELECT a.id INTO a_fan FROM gr33ncore.actuators a
    WHERE a.farm_id = p_farm_id AND a.name = 'Coop exhaust fan' AND a.deleted_at IS NULL LIMIT 1;
    SELECT a.id INTO a_heat FROM gr33ncore.actuators a
    WHERE a.farm_id = p_farm_id AND a.name = 'Coop heat lamp' AND a.deleted_at IS NULL LIMIT 1;

    IF s_water IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'Coop — Low water (task)'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id, 'Coop — Low water (task)', 'Create a refill task when water level drops below 20%.', FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_water, 'op', 'lt', 'value', 20),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_water, 'op', 'lt', 'value', 20))),
            300
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, action_parameters)
        SELECT r.id, 0, 'create_task'::gr33ncore.executable_action_type_enum,
            jsonb_build_object(
                'title', 'Refill coop waterer',
                'description', 'Water level dropped below 20% — top up drinkers.',
                'task_type', 'inspection', 'priority', 2, 'due_in_days', 0, 'zone_id', z_coop
            )
        FROM gr33ncore.automation_rules r
        WHERE r.farm_id = p_farm_id AND r.name = 'Coop — Low water (task)';
    END IF;

    IF s_feed IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'Coop — Low feed (task)'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id, 'Coop — Low feed (task)', 'Create a task when feed hopper is below 15%.', FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_feed, 'op', 'lt', 'value', 15),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_feed, 'op', 'lt', 'value', 15))),
            300
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, action_parameters)
        SELECT r.id, 0, 'create_task'::gr33ncore.executable_action_type_enum,
            jsonb_build_object(
                'title', 'Refill feed hopper',
                'description', 'Feed level below 15% — refill or adjust sensor.',
                'task_type', 'feeding', 'priority', 2, 'due_in_days', 0, 'zone_id', z_coop
            )
        FROM gr33ncore.automation_rules r
        WHERE r.farm_id = p_farm_id AND r.name = 'Coop — Low feed (task)';
    END IF;

    IF s_temp IS NOT NULL AND a_fan IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'Coop — Hot: exhaust on'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id, 'Coop — Hot: exhaust on', 'Turn exhaust fan on when coop air exceeds 32°C.', FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_temp, 'op', 'gt', 'value', 32),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_temp, 'op', 'gt', 'value', 32))),
            600
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
        SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, a_fan, 'on'
        FROM gr33ncore.automation_rules r
        WHERE r.farm_id = p_farm_id AND r.name = 'Coop — Hot: exhaust on';
    END IF;

    IF s_temp IS NOT NULL AND a_heat IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'Coop — Cold: heat lamp on'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id, 'Coop — Cold: heat lamp on', 'Turn heat lamp on when coop air drops below 5°C.', FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_temp, 'op', 'lt', 'value', 5),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_temp, 'op', 'lt', 'value', 5))),
            600
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
        SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, a_heat, 'on'
        FROM gr33ncore.automation_rules r
        WHERE r.farm_id = p_farm_id AND r.name = 'Coop — Cold: heat lamp on';
    END IF;

    -- Phase 20.8 WS5 addition — seed a head-count anchor so the
    -- Animals page has something to show immediately after bootstrap.
    -- Idempotent on (farm_id, label) for active rows.
    INSERT INTO gr33nanimals.animal_groups (farm_id, label, species, count, primary_zone_id, meta)
    SELECT p_farm_id, 'Layer flock', 'chicken', 12, z_coop,
           jsonb_build_object('template_key', 'chicken_coop_v1')
    WHERE z_coop IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33nanimals.animal_groups g
          WHERE g.farm_id = p_farm_id
            AND g.label = 'Layer flock'
            AND g.deleted_at IS NULL
      );
END;
$$;

CREATE OR REPLACE FUNCTION gr33ncore._bootstrap_small_aquaponics_v1(p_farm_id BIGINT, p_tz TEXT)
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    z_fish BIGINT;
    z_bed  BIGINT;
    dev_id BIGINT;
    s_tw   BIGINT;
    s_nh3  BIGINT;
BEGIN
    INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
    SELECT p_farm_id, x.n, x.d, x.t
    FROM (
        VALUES
            ('Fish Tank', 'Aquaculture side — water quality and life support.', 'indoor'),
            ('Grow Bed', 'Hydroponic / media bed linked to fish loop.', 'indoor')
    ) AS x(n, d, t)
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33ncore.zones z
        WHERE z.farm_id = p_farm_id AND z.name = x.n AND z.deleted_at IS NULL
    );

    SELECT z.id INTO z_fish FROM gr33ncore.zones z
    WHERE z.farm_id = p_farm_id AND z.name = 'Fish Tank' AND z.deleted_at IS NULL LIMIT 1;
    SELECT z.id INTO z_bed FROM gr33ncore.zones z
    WHERE z.farm_id = p_farm_id AND z.name = 'Grow Bed' AND z.deleted_at IS NULL LIMIT 1;

    INSERT INTO gr33ncore.devices (farm_id, zone_id, name, device_uid, device_type, status)
    SELECT p_farm_id, z_fish, 'Aquaponics Pi', 'bootstrap-aquaponics-' || p_farm_id::TEXT, 'raspberry_pi', 'unknown'::gr33ncore.device_status_enum
    WHERE z_fish IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.devices d
          WHERE d.farm_id = p_farm_id AND d.device_uid = 'bootstrap-aquaponics-' || p_farm_id::TEXT
      );

    SELECT d.id INTO dev_id FROM gr33ncore.devices d
    WHERE d.farm_id = p_farm_id AND d.device_uid = 'bootstrap-aquaponics-' || p_farm_id::TEXT LIMIT 1;

    INSERT INTO gr33ncore.sensors (farm_id, zone_id, name, sensor_type, unit_id, reading_interval_seconds, config)
    SELECT p_farm_id, z_fish, x.name, x.stype, u.id, x.ivl, '{}'::jsonb
    FROM (VALUES
        ('Tank water temperature', 'water_temperature', 'celsius', 120),
        ('Tank pH', 'ph', 'ph_unit', 300),
        ('Tank ammonia', 'ammonia_ppm', 'parts_per_million', 300),
        ('Tank nitrate', 'nitrate_ppm', 'parts_per_million', 300)
    ) AS x(name, stype, uname, ivl)
    JOIN gr33ncore.units u ON u.name = x.uname
    WHERE z_fish IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.sensors s
          WHERE s.farm_id = p_farm_id AND s.name = x.name AND s.deleted_at IS NULL
      );

    INSERT INTO gr33ncore.sensors (farm_id, zone_id, name, sensor_type, unit_id, reading_interval_seconds, config)
    SELECT p_farm_id, z_bed, x.name, x.stype, u.id, x.ivl, '{}'::jsonb
    FROM (VALUES
        ('Bed pH', 'ph', 'ph_unit', 300),
        ('Bed EC', 'ec', 'ms_per_cm', 300)
    ) AS x(name, stype, uname, ivl)
    JOIN gr33ncore.units u ON u.name = x.uname
    WHERE z_bed IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.sensors s
          WHERE s.farm_id = p_farm_id AND s.name = x.name AND s.deleted_at IS NULL
      );

    INSERT INTO gr33ncore.actuators (farm_id, zone_id, device_id, name, actuator_type)
    SELECT p_farm_id, z_fish, dev_id, x.aname, x.atype
    FROM (VALUES
        ('Return pump', 'return_pump'),
        ('Air pump', 'air_pump')
    ) AS x(aname, atype)
    WHERE z_fish IS NOT NULL AND dev_id IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.actuators a
          WHERE a.farm_id = p_farm_id AND a.name = x.aname AND a.deleted_at IS NULL
      );

    -- Phase 20.8 WS5 — seed the loop row with typed FKs on first apply
    -- and patch the Phase-20.5 row (which was created without them) if
    -- the farm already bootstrapped. idempotent both ways.
    INSERT INTO gr33naquaponics.loops (farm_id, label, fish_tank_zone_id, grow_bed_zone_id, meta)
    SELECT p_farm_id, 'Main aquaponics loop', z_fish, z_bed,
           jsonb_build_object('template_key', 'small_aquaponics_v1', 'fish_zone', 'Fish Tank', 'grow_bed_zone', 'Grow Bed')
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33naquaponics.loops l
        WHERE l.farm_id = p_farm_id AND l.label = 'Main aquaponics loop' AND l.deleted_at IS NULL
    );

    UPDATE gr33naquaponics.loops l
       SET fish_tank_zone_id = COALESCE(l.fish_tank_zone_id, z_fish),
           grow_bed_zone_id  = COALESCE(l.grow_bed_zone_id,  z_bed)
     WHERE l.farm_id = p_farm_id
       AND l.label = 'Main aquaponics loop'
       AND l.deleted_at IS NULL
       AND (l.fish_tank_zone_id IS NULL OR l.grow_bed_zone_id IS NULL);

    INSERT INTO gr33ncore.schedules (farm_id, name, description, schedule_type, cron_expression, timezone, is_active)
    SELECT p_farm_id, 'Aquaponics daily fish feed', 'Reminder to feed fish and observe feeding response.', 'feeding', '0 8 * * *', p_tz, FALSE
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33ncore.schedules s WHERE s.farm_id = p_farm_id AND s.name = 'Aquaponics daily fish feed'
    );

    SELECT s.id INTO s_tw FROM gr33ncore.sensors s WHERE s.farm_id = p_farm_id AND s.name = 'Tank water temperature' AND s.deleted_at IS NULL LIMIT 1;
    SELECT s.id INTO s_nh3 FROM gr33ncore.sensors s WHERE s.farm_id = p_farm_id AND s.name = 'Tank ammonia' AND s.deleted_at IS NULL LIMIT 1;

    IF s_nh3 IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'Aqua — Ammonia spike (task)'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id, 'Aqua — Ammonia spike (task)', 'Create task when total ammonia exceeds 0.5 ppm.', FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_nh3, 'op', 'gt', 'value', 0.5),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_nh3, 'op', 'gt', 'value', 0.5))),
            600
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, action_parameters)
        SELECT r.id, 0, 'create_task'::gr33ncore.executable_action_type_enum,
            jsonb_build_object(
                'title', 'Ammonia spike — check fish and biofilter',
                'description', 'Ammonia above 0.5 ppm. Check stocking, feeding, and biofilter.',
                'task_type', 'inspection', 'priority', 3, 'due_in_days', 0, 'zone_id', z_fish
            )
        FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'Aqua — Ammonia spike (task)';
    END IF;

    IF s_tw IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'Aqua — Cold tank (task)'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id, 'Aqua — Cold tank (task)', 'Cold water stresses fish; investigate heater.', FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_tw, 'op', 'lt', 'value', 18),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_tw, 'op', 'lt', 'value', 18))),
            900
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, action_parameters)
        SELECT r.id, 0, 'create_task'::gr33ncore.executable_action_type_enum,
            jsonb_build_object(
                'title', 'Tank water below 18°C — heater check',
                'description', 'Template rule fired on low tank temperature.',
                'task_type', 'inspection', 'priority', 2, 'due_in_days', 0, 'zone_id', z_fish
            )
        FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'Aqua — Cold tank (task)';
    END IF;

    INSERT INTO gr33ncore.tasks (farm_id, zone_id, title, description, task_type, status, priority, due_date)
    SELECT p_farm_id, z_fish,
           'Daily: feed fish and observe appetite',
           'Template — match feed rate to temperature and fish size.',
           'feeding', 'todo'::gr33ncore.task_status_enum, 1, CURRENT_DATE + 1
    WHERE z_fish IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.tasks t
          WHERE t.farm_id = p_farm_id AND t.deleted_at IS NULL
            AND t.title = 'Daily: feed fish and observe appetite'
      );
END;
$$;
