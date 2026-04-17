-- Phase 20.5 WS2: optional farm bootstrap templates for animal husbandry,
-- greenhouse / tent climate, drying room, and small aquaponics.
-- Idempotent per farm + template_key via gr33ncore.farm_bootstrap_applications.

-- ── chicken_coop_v1 ─────────────────────────────────────────────────────────
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

    -- Rules (start inactive — tune thresholds, then enable).
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
END;
$$;

-- ── greenhouse_climate_v1 ──────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION gr33ncore._bootstrap_greenhouse_climate_v1(p_farm_id BIGINT, p_tz TEXT)
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    z_gh     BIGINT;
    dev_id   BIGINT;
    s_t      BIGINT;
    s_co2    BIGINT;
    s_dp     BIGINT;
    s_vpd    BIGINT;
    a_fan    BIGINT;
    a_hum    BIGINT;
    a_dehu   BIGINT;
    a_shade  BIGINT;
    a_co2    BIGINT;
BEGIN
    INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
    SELECT p_farm_id, x.n, x.d, x.t
    FROM (
        VALUES
            ('Greenhouse', 'Poly tunnel or glasshouse — fans, humidifier, dehumidifier, shade, optional CO2.', 'indoor')
    ) AS x(n, d, t)
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33ncore.zones z
        WHERE z.farm_id = p_farm_id AND z.name = x.n AND z.deleted_at IS NULL
    );

    SELECT z.id INTO z_gh FROM gr33ncore.zones z
    WHERE z.farm_id = p_farm_id AND z.name = 'Greenhouse' AND z.deleted_at IS NULL LIMIT 1;

    INSERT INTO gr33ncore.devices (farm_id, zone_id, name, device_uid, device_type, status)
    SELECT p_farm_id, z_gh, 'Greenhouse climate Pi', 'bootstrap-greenhouse-' || p_farm_id::TEXT, 'raspberry_pi', 'unknown'::gr33ncore.device_status_enum
    WHERE z_gh IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.devices d
          WHERE d.farm_id = p_farm_id AND d.device_uid = 'bootstrap-greenhouse-' || p_farm_id::TEXT
      );

    SELECT d.id INTO dev_id FROM gr33ncore.devices d
    WHERE d.farm_id = p_farm_id AND d.device_uid = 'bootstrap-greenhouse-' || p_farm_id::TEXT LIMIT 1;

    INSERT INTO gr33ncore.sensors (farm_id, zone_id, name, sensor_type, unit_id, reading_interval_seconds, config)
    SELECT p_farm_id, z_gh, x.name, x.stype, u.id, x.ivl, '{}'::jsonb
    FROM (VALUES
        ('GH air temperature', 'temperature', 'celsius', 60),
        ('GH air humidity', 'humidity', 'percent', 60),
        ('GH CO2', 'co2', 'parts_per_million', 120),
        ('GH dew point', 'dew_point', 'celsius', 60),
        ('GH VPD', 'vpd', 'pascal', 60)
    ) AS x(name, stype, uname, ivl)
    JOIN gr33ncore.units u ON u.name = x.uname
    WHERE z_gh IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.sensors s
          WHERE s.farm_id = p_farm_id AND s.name = x.name AND s.deleted_at IS NULL
      );

    INSERT INTO gr33ncore.actuators (farm_id, zone_id, device_id, name, actuator_type)
    SELECT p_farm_id, z_gh, dev_id, x.aname, x.atype
    FROM (VALUES
        ('GH exhaust fan', 'exhaust_fan'),
        ('GH humidifier', 'humidifier'),
        ('GH dehumidifier', 'dehumidifier'),
        ('GH shade motor', 'shade_cloth_motor'),
        ('GH CO2 injector', 'co2_injector')
    ) AS x(aname, atype)
    WHERE z_gh IS NOT NULL AND dev_id IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.actuators a
          WHERE a.farm_id = p_farm_id AND a.name = x.aname AND a.deleted_at IS NULL
      );

    INSERT INTO gr33ncore.tasks (farm_id, zone_id, title, description, task_type, status, priority, due_date)
    SELECT p_farm_id, z_gh,
           'Weekly: CO2 bottle / enrichment check',
           'Template — verify tank level, solenoid, and timer interlocks.',
           'inspection', 'todo'::gr33ncore.task_status_enum, 1, CURRENT_DATE + 7
    WHERE z_gh IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.tasks t
          WHERE t.farm_id = p_farm_id AND t.deleted_at IS NULL
            AND t.title = 'Weekly: CO2 bottle / enrichment check'
      );

    SELECT s.id INTO s_t FROM gr33ncore.sensors s WHERE s.farm_id = p_farm_id AND s.name = 'GH air temperature' AND s.deleted_at IS NULL LIMIT 1;
    SELECT s.id INTO s_co2 FROM gr33ncore.sensors s WHERE s.farm_id = p_farm_id AND s.name = 'GH CO2' AND s.deleted_at IS NULL LIMIT 1;
    SELECT s.id INTO s_dp FROM gr33ncore.sensors s WHERE s.farm_id = p_farm_id AND s.name = 'GH dew point' AND s.deleted_at IS NULL LIMIT 1;
    SELECT s.id INTO s_vpd FROM gr33ncore.sensors s WHERE s.farm_id = p_farm_id AND s.name = 'GH VPD' AND s.deleted_at IS NULL LIMIT 1;

    SELECT a.id INTO a_fan FROM gr33ncore.actuators a WHERE a.farm_id = p_farm_id AND a.name = 'GH exhaust fan' AND a.deleted_at IS NULL LIMIT 1;
    SELECT a.id INTO a_hum FROM gr33ncore.actuators a WHERE a.farm_id = p_farm_id AND a.name = 'GH humidifier' AND a.deleted_at IS NULL LIMIT 1;
    SELECT a.id INTO a_dehu FROM gr33ncore.actuators a WHERE a.farm_id = p_farm_id AND a.name = 'GH dehumidifier' AND a.deleted_at IS NULL LIMIT 1;
    SELECT a.id INTO a_shade FROM gr33ncore.actuators a WHERE a.farm_id = p_farm_id AND a.name = 'GH shade motor' AND a.deleted_at IS NULL LIMIT 1;
    SELECT a.id INTO a_co2 FROM gr33ncore.actuators a WHERE a.farm_id = p_farm_id AND a.name = 'GH CO2 injector' AND a.deleted_at IS NULL LIMIT 1;

    -- Dew point high -> dehumidifier on (tune for crop; default 15°C dew point).
    IF s_dp IS NOT NULL AND a_dehu IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — Dew point high: dehumidify'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id, 'GH — Dew point high: dehumidify', 'Run dehumidifier when dew point exceeds 15°C.', FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_dp, 'op', 'gt', 'value', 15),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_dp, 'op', 'gt', 'value', 15))),
            600
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
        SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, a_dehu, 'on'
        FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — Dew point high: dehumidify';
    END IF;

    -- VPD high (dry air) -> humidifier; threshold 1500 Pa (~1.5 kPa) when Pi posts VPD in pascal.
    IF s_vpd IS NOT NULL AND a_hum IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — VPD high: humidify'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id, 'GH — VPD high: humidify', 'When VPD exceeds ~1.5 kPa (very dry air), run humidifier.', FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_vpd, 'op', 'gt', 'value', 1500),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_vpd, 'op', 'gt', 'value', 1500))),
            600
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
        SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, a_hum, 'on'
        FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — VPD high: humidify';
    END IF;

    IF s_co2 IS NOT NULL AND a_co2 IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — CO2 low: injector on'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id, 'GH — CO2 low: injector on', 'Optional enrichment when CO2 below 800 ppm.', FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_co2, 'op', 'lt', 'value', 800),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_co2, 'op', 'lt', 'value', 800))),
            900
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
        SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, a_co2, 'on'
        FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — CO2 low: injector on';
    END IF;

    IF s_t IS NOT NULL AND a_fan IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — Hot: exhaust fan'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id, 'GH — Hot: exhaust fan', 'Cooling assist when air temperature exceeds 30°C.', FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_t, 'op', 'gt', 'value', 30),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_t, 'op', 'gt', 'value', 30))),
            600
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
        SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, a_fan, 'on'
        FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — Hot: exhaust fan';
    END IF;

    IF s_t IS NOT NULL AND a_shade IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — Bright heat: close shade'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id, 'GH — Bright heat: close shade', 'Close shade cloth when temperature exceeds 28°C (tune with PAR if available).', FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_t, 'op', 'gt', 'value', 28),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_t, 'op', 'gt', 'value', 28))),
            900
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
        SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, a_shade, 'on'
        FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — Bright heat: close shade';
    END IF;
END;
$$;

-- ── drying_room_v1 ──────────────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION gr33ncore._bootstrap_drying_room_v1(p_farm_id BIGINT, p_tz TEXT)
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    z_dr  BIGINT;
    dev_id BIGINT;
    s_rh  BIGINT;
    s_dp  BIGINT;
    a_dehu BIGINT;
    a_fan BIGINT;
BEGIN
    INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
    SELECT p_farm_id, x.n, x.d, x.t
    FROM (
        VALUES
            ('Drying Room', 'Post-harvest dry and cure space — defaults tuned for cannabis-style 55–62% RH; retune for herbs and flowers.', 'indoor')
    ) AS x(n, d, t)
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33ncore.zones z
        WHERE z.farm_id = p_farm_id AND z.name = x.n AND z.deleted_at IS NULL
    );

    SELECT z.id INTO z_dr FROM gr33ncore.zones z
    WHERE z.farm_id = p_farm_id AND z.name = 'Drying Room' AND z.deleted_at IS NULL LIMIT 1;

    INSERT INTO gr33ncore.devices (farm_id, zone_id, name, device_uid, device_type, status)
    SELECT p_farm_id, z_dr, 'Drying room Pi', 'bootstrap-drying-room-' || p_farm_id::TEXT, 'raspberry_pi', 'unknown'::gr33ncore.device_status_enum
    WHERE z_dr IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.devices d
          WHERE d.farm_id = p_farm_id AND d.device_uid = 'bootstrap-drying-room-' || p_farm_id::TEXT
      );

    SELECT d.id INTO dev_id FROM gr33ncore.devices d
    WHERE d.farm_id = p_farm_id AND d.device_uid = 'bootstrap-drying-room-' || p_farm_id::TEXT LIMIT 1;

    INSERT INTO gr33ncore.sensors (farm_id, zone_id, name, sensor_type, unit_id, reading_interval_seconds, config)
    SELECT p_farm_id, z_dr, x.name, x.stype, u.id, x.ivl, '{}'::jsonb
    FROM (VALUES
        ('Dry room air temperature', 'temperature', 'celsius', 120),
        ('Dry room air humidity', 'humidity', 'percent', 120),
        ('Dry room dew point', 'dew_point', 'celsius', 120)
    ) AS x(name, stype, uname, ivl)
    JOIN gr33ncore.units u ON u.name = x.uname
    WHERE z_dr IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.sensors s
          WHERE s.farm_id = p_farm_id AND s.name = x.name AND s.deleted_at IS NULL
      );

    INSERT INTO gr33ncore.actuators (farm_id, zone_id, device_id, name, actuator_type)
    SELECT p_farm_id, z_dr, dev_id, x.aname, x.atype
    FROM (VALUES
        ('Dry room dehumidifier', 'dehumidifier'),
        ('Dry room circulation fan', 'circulation_fan')
    ) AS x(aname, atype)
    WHERE z_dr IS NOT NULL AND dev_id IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.actuators a
          WHERE a.farm_id = p_farm_id AND a.name = x.aname AND a.deleted_at IS NULL
      );

    SELECT s.id INTO s_rh FROM gr33ncore.sensors s WHERE s.farm_id = p_farm_id AND s.name = 'Dry room air humidity' AND s.deleted_at IS NULL LIMIT 1;
    SELECT s.id INTO s_dp FROM gr33ncore.sensors s WHERE s.farm_id = p_farm_id AND s.name = 'Dry room dew point' AND s.deleted_at IS NULL LIMIT 1;
    SELECT a.id INTO a_dehu FROM gr33ncore.actuators a WHERE a.farm_id = p_farm_id AND a.name = 'Dry room dehumidifier' AND a.deleted_at IS NULL LIMIT 1;
    SELECT a.id INTO a_fan FROM gr33ncore.actuators a WHERE a.farm_id = p_farm_id AND a.name = 'Dry room circulation fan' AND a.deleted_at IS NULL LIMIT 1;

    IF s_dp IS NOT NULL AND a_dehu IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'Dry — Dew point high: dehumidify'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id, 'Dry — Dew point high: dehumidify', 'Cannabis-oriented default: dehumidify when dew point exceeds 12°C.', FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_dp, 'op', 'gt', 'value', 12),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_dp, 'op', 'gt', 'value', 12))),
            600
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
        SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, a_dehu, 'on'
        FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'Dry — Dew point high: dehumidify';
    END IF;

    IF s_dp IS NOT NULL AND a_dehu IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'Dry — Dew point low: dehumidify off'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id, 'Dry — Dew point low: dehumidify off', 'Turn dehumidifier off when dew point falls below 7°C.', FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_dp, 'op', 'lt', 'value', 7),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_dp, 'op', 'lt', 'value', 7))),
            600
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
        SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, a_dehu, 'off'
        FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'Dry — Dew point low: dehumidify off';
    END IF;

    IF s_rh IS NOT NULL AND a_fan IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'Dry — High RH: circulation on'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id, 'Dry — High RH: circulation on', 'Run circulation when RH exceeds 70%.', FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_rh, 'op', 'gt', 'value', 70),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_rh, 'op', 'gt', 'value', 70))),
            600
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
        SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, a_fan, 'on'
        FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'Dry — High RH: circulation on';
    END IF;

    INSERT INTO gr33ncore.tasks (farm_id, zone_id, title, description, task_type, status, priority, due_date)
    SELECT p_farm_id, z_dr,
           'Daily: check dry room environment log',
           'Template — log temp/RH/dew point and adjust targets for your crop.',
           'inspection', 'todo'::gr33ncore.task_status_enum, 1, CURRENT_DATE + 1
    WHERE z_dr IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.tasks t
          WHERE t.farm_id = p_farm_id AND t.deleted_at IS NULL
            AND t.title = 'Daily: check dry room environment log'
      );
END;
$$;

-- ── small_aquaponics_v1 ─────────────────────────────────────────────────────
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

    INSERT INTO gr33naquaponics.loops (farm_id, label, meta)
    SELECT p_farm_id, 'Main aquaponics loop',
           jsonb_build_object('template_key', 'small_aquaponics_v1', 'fish_zone', 'Fish Tank', 'grow_bed_zone', 'Grow Bed')
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33naquaponics.loops l
        WHERE l.farm_id = p_farm_id AND l.label = 'Main aquaponics loop' AND l.deleted_at IS NULL
    );

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

-- Dispatcher: extend supported template keys (replaces function body from prior migrations).
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

    IF v_norm = 'chicken_coop_v1' THEN
        IF EXISTS (
            SELECT 1 FROM gr33ncore.farm_bootstrap_applications a
            WHERE a.farm_id = p_farm_id AND a.template_key = 'chicken_coop_v1'
        ) THEN
            RETURN jsonb_build_object(
                'applied', FALSE, 'already_applied', TRUE,
                'template', 'chicken_coop_v1', 'version', 1
            );
        END IF;
        PERFORM gr33ncore._bootstrap_chicken_coop_v1(p_farm_id, v_tz);
        BEGIN
            INSERT INTO gr33ncore.farm_bootstrap_applications (farm_id, template_key, template_version)
            VALUES (p_farm_id, 'chicken_coop_v1', 1);
        EXCEPTION WHEN unique_violation THEN
            RETURN jsonb_build_object(
                'applied', FALSE, 'already_applied', TRUE,
                'template', 'chicken_coop_v1', 'version', 1
            );
        END;
        RETURN jsonb_build_object('applied', TRUE, 'template', 'chicken_coop_v1', 'version', 1);
    END IF;

    IF v_norm = 'greenhouse_climate_v1' THEN
        IF EXISTS (
            SELECT 1 FROM gr33ncore.farm_bootstrap_applications a
            WHERE a.farm_id = p_farm_id AND a.template_key = 'greenhouse_climate_v1'
        ) THEN
            RETURN jsonb_build_object(
                'applied', FALSE, 'already_applied', TRUE,
                'template', 'greenhouse_climate_v1', 'version', 1
            );
        END IF;
        PERFORM gr33ncore._bootstrap_greenhouse_climate_v1(p_farm_id, v_tz);
        BEGIN
            INSERT INTO gr33ncore.farm_bootstrap_applications (farm_id, template_key, template_version)
            VALUES (p_farm_id, 'greenhouse_climate_v1', 1);
        EXCEPTION WHEN unique_violation THEN
            RETURN jsonb_build_object(
                'applied', FALSE, 'already_applied', TRUE,
                'template', 'greenhouse_climate_v1', 'version', 1
            );
        END;
        RETURN jsonb_build_object('applied', TRUE, 'template', 'greenhouse_climate_v1', 'version', 1);
    END IF;

    IF v_norm = 'drying_room_v1' THEN
        IF EXISTS (
            SELECT 1 FROM gr33ncore.farm_bootstrap_applications a
            WHERE a.farm_id = p_farm_id AND a.template_key = 'drying_room_v1'
        ) THEN
            RETURN jsonb_build_object(
                'applied', FALSE, 'already_applied', TRUE,
                'template', 'drying_room_v1', 'version', 1
            );
        END IF;
        PERFORM gr33ncore._bootstrap_drying_room_v1(p_farm_id, v_tz);
        BEGIN
            INSERT INTO gr33ncore.farm_bootstrap_applications (farm_id, template_key, template_version)
            VALUES (p_farm_id, 'drying_room_v1', 1);
        EXCEPTION WHEN unique_violation THEN
            RETURN jsonb_build_object(
                'applied', FALSE, 'already_applied', TRUE,
                'template', 'drying_room_v1', 'version', 1
            );
        END;
        RETURN jsonb_build_object('applied', TRUE, 'template', 'drying_room_v1', 'version', 1);
    END IF;

    IF v_norm = 'small_aquaponics_v1' THEN
        IF EXISTS (
            SELECT 1 FROM gr33ncore.farm_bootstrap_applications a
            WHERE a.farm_id = p_farm_id AND a.template_key = 'small_aquaponics_v1'
        ) THEN
            RETURN jsonb_build_object(
                'applied', FALSE, 'already_applied', TRUE,
                'template', 'small_aquaponics_v1', 'version', 1
            );
        END IF;
        PERFORM gr33ncore._bootstrap_small_aquaponics_v1(p_farm_id, v_tz);
        BEGIN
            INSERT INTO gr33ncore.farm_bootstrap_applications (farm_id, template_key, template_version)
            VALUES (p_farm_id, 'small_aquaponics_v1', 1);
        EXCEPTION WHEN unique_violation THEN
            RETURN jsonb_build_object(
                'applied', FALSE, 'already_applied', TRUE,
                'template', 'small_aquaponics_v1', 'version', 1
            );
        END;
        RETURN jsonb_build_object('applied', TRUE, 'template', 'small_aquaponics_v1', 'version', 1);
    END IF;

    RETURN jsonb_build_object('error', 'unknown_template', 'template', p_template);
END;
$$;
