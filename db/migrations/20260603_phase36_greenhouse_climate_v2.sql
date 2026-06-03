-- Phase 36 WS2+WS5: upgrade greenhouse_climate_v1 bootstrap to typed actuators
-- and add the greenhouse_climate zone profile (meta_data.greenhouse_climate).
-- Also adds the Phase 36 WS3 greenhouse rule-template function.
--
-- IMPORTANT: this migration replaces the _bootstrap_greenhouse_climate_v1 function
-- body so the next apply_farm_bootstrap_template call generates typed actuators
-- for NEW farms.  Existing farms that already applied v1 keep their rows (no
-- destructive migration of live data).

-- ── Updated bootstrap function ────────────────────────────────────────────────
CREATE OR REPLACE FUNCTION gr33ncore._bootstrap_greenhouse_climate_v1(p_farm_id BIGINT, p_tz TEXT)
RETURNS void
LANGUAGE plpgsql
AS $$
DECLARE
    z_gh     BIGINT;
    dev_id   BIGINT;
    s_t      BIGINT;
    s_lux    BIGINT;
    s_co2    BIGINT;
    s_dp     BIGINT;
    s_vpd    BIGINT;
    a_fan    BIGINT;
    a_circ   BIGINT;
    a_hum    BIGINT;
    a_dehu   BIGINT;
    a_shade  BIGINT;
    a_vent   BIGINT;
    a_co2    BIGINT;
    v_profile JSONB;
BEGIN
    -- Zone: set zone_type='greenhouse' (Phase 36 WS1/WS5)
    INSERT INTO gr33ncore.zones (farm_id, name, description, zone_type)
    SELECT p_farm_id, x.n, x.d, x.t
    FROM (
        VALUES
            ('Greenhouse', 'Poly tunnel or glasshouse — shade screen, ridge vent, exhaust/circulation fans, optional CO2.', 'greenhouse')
    ) AS x(n, d, t)
    WHERE NOT EXISTS (
        SELECT 1 FROM gr33ncore.zones z
        WHERE z.farm_id = p_farm_id AND z.name = x.n AND z.deleted_at IS NULL
    );

    -- Upgrade zone_type to 'greenhouse' if the prior bootstrap created it as 'indoor'
    UPDATE gr33ncore.zones
    SET zone_type = 'greenhouse'
    WHERE farm_id = p_farm_id AND name = 'Greenhouse' AND zone_type = 'indoor' AND deleted_at IS NULL;

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

    -- Sensors: add lux sensor in addition to prior climate sensors
    INSERT INTO gr33ncore.sensors (farm_id, zone_id, name, sensor_type, unit_id, reading_interval_seconds, config)
    SELECT p_farm_id, z_gh, x.name, x.stype, u.id, x.ivl, '{}'::jsonb
    FROM (VALUES
        ('GH air temperature', 'temperature',  'celsius',            60),
        ('GH air humidity',    'humidity',      'percent',            60),
        ('GH CO2',             'co2',           'parts_per_million', 120),
        ('GH dew point',       'dew_point',     'celsius',            60),
        ('GH VPD',             'vpd',           'pascal',             60),
        ('GH lux',             'lux',           'lux',                30)
    ) AS x(name, stype, uname, ivl)
    JOIN gr33ncore.units u ON u.name = x.uname
    WHERE z_gh IS NOT NULL
      AND NOT EXISTS (
          SELECT 1 FROM gr33ncore.sensors s
          WHERE s.farm_id = p_farm_id AND s.name = x.name AND s.deleted_at IS NULL
      );

    -- Actuators: typed Phase 36 types
    -- shade_screen replaces shade_cloth_motor for new farms
    -- ridge_vent is new; exhaust_fan + circulation_fan replace generic 'exhaust_fan'
    INSERT INTO gr33ncore.actuators (farm_id, zone_id, device_id, name, actuator_type)
    SELECT p_farm_id, z_gh, dev_id, x.aname, x.atype
    FROM (VALUES
        ('GH exhaust fan',      'exhaust_fan'),
        ('GH circulation fan',  'circulation_fan'),
        ('GH humidifier',       'humidifier'),
        ('GH dehumidifier',     'dehumidifier'),
        ('GH shade screen',     'shade_screen'),
        ('GH ridge vent',       'ridge_vent'),
        ('GH CO2 injector',     'co2_injector')
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

    -- Resolve sensor/actuator IDs
    SELECT s.id INTO s_t    FROM gr33ncore.sensors s WHERE s.farm_id = p_farm_id AND s.name = 'GH air temperature' AND s.deleted_at IS NULL LIMIT 1;
    SELECT s.id INTO s_lux  FROM gr33ncore.sensors s WHERE s.farm_id = p_farm_id AND s.name = 'GH lux'             AND s.deleted_at IS NULL LIMIT 1;
    SELECT s.id INTO s_co2  FROM gr33ncore.sensors s WHERE s.farm_id = p_farm_id AND s.name = 'GH CO2'             AND s.deleted_at IS NULL LIMIT 1;
    SELECT s.id INTO s_dp   FROM gr33ncore.sensors s WHERE s.farm_id = p_farm_id AND s.name = 'GH dew point'       AND s.deleted_at IS NULL LIMIT 1;
    SELECT s.id INTO s_vpd  FROM gr33ncore.sensors s WHERE s.farm_id = p_farm_id AND s.name = 'GH VPD'             AND s.deleted_at IS NULL LIMIT 1;

    SELECT a.id INTO a_fan   FROM gr33ncore.actuators a WHERE a.farm_id = p_farm_id AND a.name = 'GH exhaust fan'     AND a.deleted_at IS NULL LIMIT 1;
    SELECT a.id INTO a_circ  FROM gr33ncore.actuators a WHERE a.farm_id = p_farm_id AND a.name = 'GH circulation fan' AND a.deleted_at IS NULL LIMIT 1;
    SELECT a.id INTO a_hum   FROM gr33ncore.actuators a WHERE a.farm_id = p_farm_id AND a.name = 'GH humidifier'      AND a.deleted_at IS NULL LIMIT 1;
    SELECT a.id INTO a_dehu  FROM gr33ncore.actuators a WHERE a.farm_id = p_farm_id AND a.name = 'GH dehumidifier'    AND a.deleted_at IS NULL LIMIT 1;
    SELECT a.id INTO a_shade FROM gr33ncore.actuators a WHERE a.farm_id = p_farm_id AND a.name IN ('GH shade screen', 'GH shade motor') AND a.deleted_at IS NULL LIMIT 1;
    SELECT a.id INTO a_vent  FROM gr33ncore.actuators a WHERE a.farm_id = p_farm_id AND a.name = 'GH ridge vent'      AND a.deleted_at IS NULL LIMIT 1;
    SELECT a.id INTO a_co2   FROM gr33ncore.actuators a WHERE a.farm_id = p_farm_id AND a.name = 'GH CO2 injector'    AND a.deleted_at IS NULL LIMIT 1;

    -- Write greenhouse_climate profile to zone meta_data (idempotent jsonb merge).
    -- Only sets fan_actuator_ids and optional shade/vent refs when actuators exist.
    v_profile := jsonb_build_object(
        'cover_type',          'polycarbonate',
        'automation_policy',   'auto',
        'notes',               'Default greenhouse climate bootstrap — tune thresholds and enable rules before activating.'
    );
    IF a_shade IS NOT NULL THEN
        v_profile := v_profile || jsonb_build_object('shade_actuator_id', a_shade);
    END IF;
    IF a_vent IS NOT NULL THEN
        v_profile := v_profile || jsonb_build_object('vent_actuator_id', a_vent);
    END IF;
    IF a_fan IS NOT NULL OR a_circ IS NOT NULL THEN
        DECLARE fan_ids JSONB := '[]'::jsonb; BEGIN
            IF a_fan IS NOT NULL THEN
                fan_ids := fan_ids || jsonb_build_array(a_fan);
            END IF;
            IF a_circ IS NOT NULL THEN
                fan_ids := fan_ids || jsonb_build_array(a_circ);
            END IF;
            v_profile := v_profile || jsonb_build_object('fan_actuator_ids', fan_ids);
        END;
    END IF;

    IF z_gh IS NOT NULL THEN
        UPDATE gr33ncore.zones
        SET meta_data = meta_data || jsonb_build_object('greenhouse_climate', v_profile),
            updated_at = NOW()
        WHERE id = z_gh
          AND (meta_data -> 'greenhouse_climate') IS NULL;
    END IF;

    -- ── Automation rules (all inactive by default) ──────────────────────────

    -- Dew point high -> dehumidifier on
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

    -- VPD high -> humidifier on
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

    -- CO2 low -> injector on
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

    -- High temp -> exhaust fan on (Phase 36 WS3: heat template)
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

    -- High lux -> deploy shade screen (Phase 36 WS3: UV/high-sun template)
    -- Threshold 80 000 lux ≈ strong midday sun; operators can tune this.
    IF s_lux IS NOT NULL AND a_shade IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — High lux: deploy shade'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id,
            'GH — High lux: deploy shade',
            'Deploy shade screen when lux exceeds 80 000 (strong midday sun). Requires lux sensor. Cooldown 30 min to prevent flutter.',
            FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_lux, 'op', 'gt', 'value', 80000),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_lux, 'op', 'gt', 'value', 80000))),
            1800
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
        SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, a_shade, 'deploy'
        FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — High lux: deploy shade';
    END IF;

    -- Night retract (temp-based proxy when no lux sensor)
    -- Retract shade when temp drops to 18°C (nightfall proxy). Operators
    -- should swap this for a time-based schedule once cron TZ is configured.
    IF s_t IS NOT NULL AND a_shade IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — Night temp: retract shade'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id,
            'GH — Night temp: retract shade',
            'Retract shade screen when temperature drops below 18°C (nightfall proxy). Replace with a cron schedule when timezone is set.',
            FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_t, 'op', 'lt', 'value', 18),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_t, 'op', 'lt', 'value', 18))),
            1800
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
        SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, a_shade, 'retract'
        FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — Night temp: retract shade';
    END IF;

    -- High temp -> open ridge vent (Phase 36 WS3: heat template)
    IF s_t IS NOT NULL AND a_vent IS NOT NULL AND NOT EXISTS (
        SELECT 1 FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — Hot: open ridge vent'
    ) THEN
        INSERT INTO gr33ncore.automation_rules (
            farm_id, name, description, is_active, trigger_source, trigger_configuration,
            condition_logic, conditions_jsonb, cooldown_period_seconds
        ) VALUES (
            p_farm_id, 'GH — Hot: open ridge vent', 'Open ridge vent when air temperature exceeds 28°C.', FALSE,
            'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
            jsonb_build_object('sensor_id', s_t, 'op', 'gt', 'value', 28),
            'ALL',
            jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                jsonb_build_object('sensor_id', s_t, 'op', 'gt', 'value', 28))),
            600
        );
        INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
        SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, a_vent, 'open'
        FROM gr33ncore.automation_rules r WHERE r.farm_id = p_farm_id AND r.name = 'GH — Hot: open ridge vent';
    END IF;
END;
$$;

-- ── Greenhouse rule-template function (Phase 36 WS3) ─────────────────────────
-- Clone greenhouse climate rules for any zone on the farm.
-- Usage: SELECT gr33ncore.apply_greenhouse_rule_templates(farm_id, zone_id, shade_actuator_id, fan_actuator_id, lux_sensor_id, temp_sensor_id);
-- All actuator/sensor args are optional — pass NULL to skip that rule family.
CREATE OR REPLACE FUNCTION gr33ncore.apply_greenhouse_rule_templates(
    p_farm_id         BIGINT,
    p_zone_id         BIGINT,
    p_shade_id        BIGINT DEFAULT NULL,
    p_fan_id          BIGINT DEFAULT NULL,
    p_lux_sensor_id   BIGINT DEFAULT NULL,
    p_temp_sensor_id  BIGINT DEFAULT NULL
)
RETURNS JSONB
LANGUAGE plpgsql
AS $$
DECLARE
    v_created INT := 0;
BEGIN
    IF NOT EXISTS (SELECT 1 FROM gr33ncore.farms WHERE id = p_farm_id AND deleted_at IS NULL) THEN
        RETURN jsonb_build_object('error', 'farm_not_found');
    END IF;

    -- High lux → deploy shade
    IF p_lux_sensor_id IS NOT NULL AND p_shade_id IS NOT NULL THEN
        IF NOT EXISTS (
            SELECT 1 FROM gr33ncore.automation_rules r
            WHERE r.farm_id = p_farm_id
              AND r.name = 'GH — High lux: deploy shade (zone ' || p_zone_id::TEXT || ')'
        ) THEN
            INSERT INTO gr33ncore.automation_rules (
                farm_id, name, description, is_active, trigger_source, trigger_configuration,
                condition_logic, conditions_jsonb, cooldown_period_seconds
            ) VALUES (
                p_farm_id,
                'GH — High lux: deploy shade (zone ' || p_zone_id::TEXT || ')',
                'Deploy shade screen when lux exceeds 80 000 in zone ' || p_zone_id::TEXT || '.',
                FALSE,
                'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
                jsonb_build_object('sensor_id', p_lux_sensor_id, 'op', 'gt', 'value', 80000, 'zone_id', p_zone_id),
                'ALL',
                jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                    jsonb_build_object('sensor_id', p_lux_sensor_id, 'op', 'gt', 'value', 80000))),
                1800
            );
            INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
            SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, p_shade_id, 'deploy'
            FROM gr33ncore.automation_rules r
            WHERE r.farm_id = p_farm_id AND r.name = 'GH — High lux: deploy shade (zone ' || p_zone_id::TEXT || ')';
            v_created := v_created + 1;
        END IF;
    END IF;

    -- High temp → exhaust fan on
    IF p_temp_sensor_id IS NOT NULL AND p_fan_id IS NOT NULL THEN
        IF NOT EXISTS (
            SELECT 1 FROM gr33ncore.automation_rules r
            WHERE r.farm_id = p_farm_id
              AND r.name = 'GH — High temp: fan (zone ' || p_zone_id::TEXT || ')'
        ) THEN
            INSERT INTO gr33ncore.automation_rules (
                farm_id, name, description, is_active, trigger_source, trigger_configuration,
                condition_logic, conditions_jsonb, cooldown_period_seconds
            ) VALUES (
                p_farm_id,
                'GH — High temp: fan (zone ' || p_zone_id::TEXT || ')',
                'Run exhaust fan when air temperature exceeds 30°C in zone ' || p_zone_id::TEXT || '.',
                FALSE,
                'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
                jsonb_build_object('sensor_id', p_temp_sensor_id, 'op', 'gt', 'value', 30, 'zone_id', p_zone_id),
                'ALL',
                jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                    jsonb_build_object('sensor_id', p_temp_sensor_id, 'op', 'gt', 'value', 30))),
                600
            );
            INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
            SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, p_fan_id, 'on'
            FROM gr33ncore.automation_rules r
            WHERE r.farm_id = p_farm_id AND r.name = 'GH — High temp: fan (zone ' || p_zone_id::TEXT || ')';
            v_created := v_created + 1;
        END IF;
    END IF;

    -- Night retract (temp proxy)
    IF p_temp_sensor_id IS NOT NULL AND p_shade_id IS NOT NULL THEN
        IF NOT EXISTS (
            SELECT 1 FROM gr33ncore.automation_rules r
            WHERE r.farm_id = p_farm_id
              AND r.name = 'GH — Night retract: shade (zone ' || p_zone_id::TEXT || ')'
        ) THEN
            INSERT INTO gr33ncore.automation_rules (
                farm_id, name, description, is_active, trigger_source, trigger_configuration,
                condition_logic, conditions_jsonb, cooldown_period_seconds
            ) VALUES (
                p_farm_id,
                'GH — Night retract: shade (zone ' || p_zone_id::TEXT || ')',
                'Retract shade when temp drops below 18°C (nightfall proxy) in zone ' || p_zone_id::TEXT || '.',
                FALSE,
                'sensor_reading_threshold'::gr33ncore.automation_trigger_source_enum,
                jsonb_build_object('sensor_id', p_temp_sensor_id, 'op', 'lt', 'value', 18, 'zone_id', p_zone_id),
                'ALL',
                jsonb_build_object('logic', 'ALL', 'predicates', jsonb_build_array(
                    jsonb_build_object('sensor_id', p_temp_sensor_id, 'op', 'lt', 'value', 18))),
                1800
            );
            INSERT INTO gr33ncore.executable_actions (rule_id, execution_order, action_type, target_actuator_id, action_command)
            SELECT r.id, 0, 'control_actuator'::gr33ncore.executable_action_type_enum, p_shade_id, 'retract'
            FROM gr33ncore.automation_rules r
            WHERE r.farm_id = p_farm_id AND r.name = 'GH — Night retract: shade (zone ' || p_zone_id::TEXT || ')';
            v_created := v_created + 1;
        END IF;
    END IF;

    RETURN jsonb_build_object('rules_created', v_created, 'zone_id', p_zone_id);
END;
$$;
