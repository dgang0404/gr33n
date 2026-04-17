-- ============================================================
-- gr33n_core schema - patched version
-- Fixes applied:
--   1. auth.users bootstrap for local dev
--   2. actuators device_id NOT NULL + ON DELETE SET NULL conflict fixed
--   3. convert_value IMMUTABLE -> STABLE
--   4. Hypertable PK column order fixed (time col first)
--   5. updated_at auto-trigger added to all relevant tables
--   6. sensor_readings composite index (sensor_id, reading_time DESC)
--   7. Soft-delete unique constraints replaced with partial indexes
-- ============================================================

-- ============================================================
-- BOOTSTRAP: local dev auth schema (Supabase provides this in
-- hosted env; needed for raw PostgreSQL local installs)
-- ============================================================
CREATE SCHEMA IF NOT EXISTS auth;

CREATE TABLE IF NOT EXISTS auth.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email TEXT UNIQUE,
    password_hash BYTEA,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- ============================================================
-- OPTIONAL BUT RECOMMENDED EXTENSIONS
-- Uncomment when PostGIS and TimescaleDB are installed locally
-- ============================================================
CREATE EXTENSION IF NOT EXISTS postgis;
CREATE EXTENSION IF NOT EXISTS timescaledb CASCADE;

-- ============================================================
-- SCHEMA: gr33n_core
-- ============================================================
CREATE SCHEMA IF NOT EXISTS gr33ncore;
COMMENT ON SCHEMA gr33ncore IS
  'Core entities shared across all gr33n farm instances, including users, farms,
  zones, devices, sensors, actuators, tasks, schedules, automation rules,
  notifications, system logging, file attachments, weather data, cost tracking,
  validation rules, and user activity. This schema forms the backbone of the
  gr33n platform.';

-- ============================================================
-- SHARED updated_at TRIGGER FUNCTION
-- Attach to every table that has an updated_at column
-- ============================================================
CREATE OR REPLACE FUNCTION gr33ncore.set_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- ============================================================
-- UNITS REFERENCE SYSTEM
-- ============================================================
CREATE TABLE IF NOT EXISTS gr33ncore.units (
    id                 BIGSERIAL PRIMARY KEY,
    name               VARCHAR(50)      UNIQUE NOT NULL,
    unit_type          VARCHAR(50)      NOT NULL,
    symbol             VARCHAR(20),
    base_unit_name     VARCHAR(50)      NOT NULL,
    conversion_to_base DECIMAL(25,15)   NOT NULL,
    is_base_unit       BOOLEAN          DEFAULT FALSE NOT NULL,
    description        TEXT,
    created_at         TIMESTAMPTZ      DEFAULT NOW() NOT NULL
);

COMMENT ON TABLE gr33ncore.units IS
  'Master reference for all units. Ensures consistency and enables automatic conversions.';

INSERT INTO gr33ncore.units
    (name, unit_type, symbol, base_unit_name, conversion_to_base, is_base_unit, description)
VALUES
-- Mass (base: kilogram)
('kilogram',          'mass',        'kg',      'kilogram',          1.0,        TRUE,  'SI base unit of mass'),
('gram',              'mass',        'g',       'kilogram',          0.001,      FALSE, 'Common small quantity unit'),
('pound',             'mass',        'lb',      'kilogram',          0.453592,   FALSE, 'Imperial mass'),
('ounce',             'mass',        'oz',      'kilogram',          0.0283495,  FALSE, 'Imperial small quantity'),
('ton_metric',        'mass',        't',       'kilogram',          1000.0,     FALSE, 'Metric ton'),
-- Volume (base: liter)
('liter',             'volume',      'L',       'liter',             1.0,        TRUE,  'Base unit of volume'),
('milliliter',        'volume',      'mL',      'liter',             0.001,      FALSE, 'Small volume'),
('gallon_us',         'volume',      'gal',     'liter',             3.78541,    FALSE, 'US gallon'),
('gallon_imperial',   'volume',      'imp gal', 'liter',             4.54609,    FALSE, 'Imperial gallon'),
('fluid_ounce_us',    'volume',      'fl oz',   'liter',             0.0295735,  FALSE, 'US fl oz'),
-- Length (base: meter)
('meter',             'length',      'm',       'meter',             1.0,        TRUE,  'SI length'),
('centimeter',        'length',      'cm',      'meter',             0.01,       FALSE, 'Centimeter'),
('millimeter',        'length',      'mm',      'meter',             0.001,      FALSE, 'Millimeter'),
('foot',              'length',      'ft',      'meter',             0.3048,     FALSE, 'Foot'),
('inch',              'length',      'in',      'meter',             0.0254,     FALSE, 'Inch'),
-- Area (base: square_meter)
('square_meter',      'area',        'm2',      'square_meter',      1.0,        TRUE,  'SI area'),
('hectare',           'area',        'ha',      'square_meter',      10000.0,    FALSE, 'Hectare'),
('acre',              'area',        'ac',      'square_meter',      4046.86,    FALSE, 'Acre'),
('square_foot',       'area',        'ft2',     'square_meter',      0.092903,   FALSE, 'Square foot'),
-- Temperature (base: celsius - special handling)
('celsius',           'temperature', 'C',       'celsius',           1.0,        TRUE,  'Celsius'),
('fahrenheit',        'temperature', 'F',       'celsius',           1.0,        FALSE, 'Fahrenheit (special conversion)'),
('kelvin',            'temperature', 'K',       'celsius',           1.0,        FALSE, 'Kelvin (offset 273.15)'),
-- Percent/dimensionless (base: percent)
('percent',           'percent',     '%',       'percent',           1.0,        TRUE,  'Percentage 0-100'),
('decimal_fraction',  'percent',     '',        'percent',           100.0,      FALSE, 'Decimal fraction 0.0-1.0'),
('parts_per_million', 'percent',     'ppm',     'percent',           0.0001,     FALSE, 'ppm'),
-- Time (base: second)
('second',            'time',        's',       'second',            1.0,        TRUE,  'SI time'),
('minute',            'time',        'min',     'second',            60.0,       FALSE, 'Minute'),
('hour',              'time',        'h',       'second',            3600.0,     FALSE, 'Hour'),
('day',               'time',        'd',       'second',            86400.0,    FALSE, 'Day'),
-- Pressure (base: pascal)
('pascal',            'pressure',    'Pa',      'pascal',            1.0,        TRUE,  'SI pressure'),
('hectopascal',       'pressure',    'hPa',     'pascal',            100.0,      FALSE, 'Atmospheric pressure'),
('bar',               'pressure',    'bar',     'pascal',            100000.0,   FALSE, 'Bar'),
('psi',               'pressure',    'psi',     'pascal',            6894.76,    FALSE, 'Pounds per square inch'),
-- Electrical/power
('volt',              'electrical',  'V',       'volt',              1.0,        TRUE,  'SI voltage'),
('ampere',            'electrical',  'A',       'ampere',            1.0,        TRUE,  'SI current'),
('watt',              'electrical',  'W',       'watt',              1.0,        TRUE,  'SI power'),
('kilowatt',          'electrical',  'kW',      'watt',              1000.0,     FALSE, 'Kilowatt'),
-- Speed (base: meters_per_second)
('meters_per_second', 'speed',       'm/s',     'meters_per_second', 1.0,        TRUE,  'SI speed'),
('kilometers_per_hour','speed',      'km/h',    'meters_per_second', 0.277778,   FALSE, 'km/h'),
('miles_per_hour',    'speed',       'mph',     'meters_per_second', 0.44704,    FALSE, 'mph'),
-- Agricultural extras
('lux',               'light',       'lx',      'lux',               1.0,        TRUE,  'Illuminance'),
('par_umol',          'light',       'umol/m2/s','par_umol',         1.0,        TRUE,  'Photosynthetically active radiation'),
('ms_per_cm',         'conductivity','mS/cm',   'ms_per_cm',         1.0,        TRUE,  'Electrical conductivity (soil/hydro)'),
('ph_unit',           'ph',          'pH',      'ph_unit',           1.0,        TRUE,  'pH dimensionless'),
('cfu_per_ml',        'microbial',   'CFU/mL',  'cfu_per_ml',        1.0,        TRUE,  'Colony forming units per mL')
ON CONFLICT (name) DO NOTHING;

-- ============================================================
-- TEMPERATURE CONVERSION (no DB access - safe as IMMUTABLE)
-- ============================================================
CREATE OR REPLACE FUNCTION gr33ncore.convert_temperature(
    value_in  DECIMAL,
    unit_from VARCHAR(50),
    unit_to   VARCHAR(50)
) RETURNS DECIMAL AS $$
BEGIN
    -- normalise to Celsius first
    IF unit_from = 'fahrenheit' THEN
        value_in := (value_in - 32) * 5.0 / 9.0;
    ELSIF unit_from = 'kelvin' THEN
        value_in := value_in - 273.15;
    END IF;
    -- convert out of Celsius
    IF unit_to = 'fahrenheit' THEN RETURN (value_in * 9.0 / 5.0) + 32;
    ELSIF unit_to = 'kelvin'  THEN RETURN value_in + 273.15;
    ELSE RETURN value_in;
    END IF;
END;
$$ LANGUAGE plpgsql IMMUTABLE;

-- ============================================================
-- GENERAL UNIT CONVERSION
-- FIX #3: STABLE not IMMUTABLE (reads from units table)
-- ============================================================
CREATE OR REPLACE FUNCTION gr33ncore.convert_value(
    value_in     DECIMAL,
    unit_id_from BIGINT,
    unit_id_to   BIGINT
) RETURNS DECIMAL AS $$
DECLARE
    from_unit  RECORD;
    to_unit    RECORD;
    base_value DECIMAL;
BEGIN
    SELECT * INTO from_unit FROM gr33ncore.units WHERE id = unit_id_from;
    SELECT * INTO to_unit   FROM gr33ncore.units WHERE id = unit_id_to;

    IF from_unit.unit_type <> to_unit.unit_type THEN
        RAISE EXCEPTION 'Cannot convert between different unit types: % and %',
            from_unit.unit_type, to_unit.unit_type;
    END IF;

    IF from_unit.unit_type = 'temperature' THEN
        RETURN gr33ncore.convert_temperature(value_in, from_unit.name, to_unit.name);
    END IF;

    base_value := value_in * from_unit.conversion_to_base;
    RETURN base_value / to_unit.conversion_to_base;
END;
$$ LANGUAGE plpgsql STABLE;  -- FIX #3: was IMMUTABLE, now STABLE

-- ============================================================
-- ENUMERATED TYPES
-- ============================================================
CREATE TYPE gr33ncore.farm_scale_tier_enum       AS ENUM ('small','medium','large','enterprise');
CREATE TYPE gr33ncore.operational_status_enum    AS ENUM ('active','maintenance','planning','archived','decommissioned');
CREATE TYPE gr33ncore.log_level_enum             AS ENUM ('DEBUG','INFO','NOTICE','WARNING','ERROR','CRITICAL','ALERT','EMERGENCY');
CREATE TYPE gr33ncore.user_role_enum             AS ENUM ('user','farm_manager','farm_worker','gr33n_system_admin');
CREATE TYPE gr33ncore.farm_member_role_enum      AS ENUM ('owner','manager','agronomist','worker','viewer','custom_role','operator','finance');
CREATE TYPE gr33ncore.device_status_enum         AS ENUM ('online','offline','error_comms','error_hardware','maintenance_mode','initializing','unknown','decommissioned','pending_activation');
CREATE TYPE gr33ncore.task_status_enum           AS ENUM ('todo','in_progress','on_hold','completed','cancelled','blocked_requires_input','pending_review');
CREATE TYPE gr33ncore.automation_trigger_source_enum AS ENUM ('sensor_reading_threshold','specific_time_cron','actuator_state_changed','manual_api_trigger','task_status_updated','new_system_log_event','external_webhook_received');
CREATE TYPE gr33ncore.executable_action_type_enum AS ENUM ('control_actuator','trigger_another_automation_rule','send_notification','create_task','log_custom_event','http_webhook_call','update_record_in_gr33n');
CREATE TYPE gr33ncore.notification_priority_enum AS ENUM ('low','medium','high','critical');
CREATE TYPE gr33ncore.notification_status_enum   AS ENUM ('pending','queued','sent','delivered','failed_to_send','read_by_user','acknowledged_by_user','archived_by_user','system_cleared');
CREATE TYPE gr33ncore.actuator_event_source_enum AS ENUM ('manual_ui_input','manual_api_call','schedule_trigger','automation_rule_trigger','device_internal_feedback_loop','system_initialization_routine','emergency_stop_signal');
CREATE TYPE gr33ncore.actuator_execution_status_enum AS ENUM ('command_sent_to_device','acknowledged_by_device','execution_started_on_device','execution_completed_success_on_device','execution_completed_with_error_on_device','execution_failed_to_start_on_device','pending_confirmation_from_feedback','timeout_waiting_for_acknowledgement','cancelled_by_user_or_system');
CREATE TYPE gr33ncore.weather_data_source_enum   AS ENUM ('farm_weather_station','api_openweather','api_visualcrossing','manual_entry','iot_sensor_reading');
CREATE TYPE gr33ncore.cost_category_enum         AS ENUM (
    'seeds_plants','fertilizers_soil_amendments','pest_disease_control','water_irrigation',
    'labor_wages','equipment_purchase_rental','equipment_maintenance_fuel','utilities_electricity_gas',
    'land_rent_mortgage','insurance','licenses_permits','feed_livestock','veterinary_services',
    'packaging_supplies','transportation_logistics','marketing_sales','training_consultancy','miscellaneous'
);
CREATE TYPE gr33ncore.validation_rule_type_enum  AS ENUM ('range_check','required_field','format_validation','regex_match','lookup_in_list','cross_field_comparison','custom_function_check');
CREATE TYPE gr33ncore.validation_severity_enum   AS ENUM ('warning','error','critical_stop');
CREATE TYPE gr33ncore.user_action_type_enum      AS ENUM ('login_success','login_failure','logout','create_record','view_record','update_record','delete_record','list_records','execute_action','change_setting','system_event','export_data','import_data');

-- ============================================================
-- TABLES
-- ============================================================

-- Profiles
CREATE TABLE IF NOT EXISTS gr33ncore.profiles (
    user_id    UUID PRIMARY KEY REFERENCES auth.users(id) ON DELETE CASCADE,
    full_name  TEXT,
    email      TEXT UNIQUE NOT NULL,
    avatar_url TEXT,
    role       gr33ncore.user_role_enum DEFAULT 'user' NOT NULL,
    preferences JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);
CREATE TRIGGER trg_profiles_updated_at
    BEFORE UPDATE ON gr33ncore.profiles
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Organizations (multi-farm tenant grouping; farms.organization_id optional)
CREATE TABLE IF NOT EXISTS gr33ncore.organizations (
    id BIGSERIAL PRIMARY KEY,
    name            TEXT        NOT NULL,
    plan_tier       TEXT        NOT NULL DEFAULT 'pilot',
    billing_status  TEXT        NOT NULL DEFAULT 'none',
    default_bootstrap_template TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TRIGGER trg_organizations_updated_at
    BEFORE UPDATE ON gr33ncore.organizations
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

CREATE TABLE IF NOT EXISTS gr33ncore.organization_memberships (
    organization_id BIGINT NOT NULL REFERENCES gr33ncore.organizations(id) ON DELETE CASCADE,
    user_id         UUID NOT NULL REFERENCES gr33ncore.profiles(user_id) ON DELETE CASCADE,
    role_in_org     TEXT   NOT NULL CHECK (role_in_org IN ('owner', 'admin', 'member')),
    joined_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (organization_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_org_memberships_user
    ON gr33ncore.organization_memberships (user_id);

-- Farms
CREATE TABLE IF NOT EXISTS gr33ncore.farms (
    id                 BIGSERIAL PRIMARY KEY,
    name               TEXT        NOT NULL,
    description        TEXT,
    location_text      TEXT,
    location_gis       GEOMETRY(Point,4326),
    size_hectares      NUMERIC(10,2),
    farm_type          TEXT,
    scale_tier         gr33ncore.farm_scale_tier_enum    DEFAULT 'small'  NOT NULL,
    owner_user_id      UUID        NOT NULL REFERENCES gr33ncore.profiles(user_id),
    timezone           TEXT        DEFAULT 'UTC' NOT NULL,
    currency           CHAR(3)     DEFAULT 'USD' NOT NULL
                           CHECK (currency ~ '^[A-Z]{3}$'),
    operational_status gr33ncore.operational_status_enum DEFAULT 'active' NOT NULL,
    created_at         TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at         TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_by_user_id UUID        REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    deleted_at         TIMESTAMPTZ DEFAULT NULL,
    organization_id    BIGINT REFERENCES gr33ncore.organizations(id) ON DELETE SET NULL,
    insert_commons_opt_in BOOLEAN NOT NULL DEFAULT FALSE,
    insert_commons_last_sync_at TIMESTAMPTZ,
    insert_commons_last_attempt_at TIMESTAMPTZ,
    insert_commons_last_delivery_status TEXT,
    insert_commons_last_error TEXT,
    insert_commons_backoff_until TIMESTAMPTZ,
    insert_commons_consecutive_failures INT NOT NULL DEFAULT 0,
    insert_commons_require_approval BOOLEAN NOT NULL DEFAULT FALSE
);
CREATE TRIGGER trg_farms_updated_at
    BEFORE UPDATE ON gr33ncore.farms
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

CREATE INDEX IF NOT EXISTS idx_farms_organization_id
    ON gr33ncore.farms (organization_id)
    WHERE deleted_at IS NULL AND organization_id IS NOT NULL;

-- Insert Commons outbound bundles (approval queue + export)
CREATE TABLE IF NOT EXISTS gr33ncore.insert_commons_bundles (
    id                   BIGSERIAL PRIMARY KEY,
    farm_id              BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    idempotency_key      TEXT,
    payload_hash         TEXT        NOT NULL,
    payload              JSONB       NOT NULL,
    status               TEXT        NOT NULL CHECK (status IN (
        'pending_approval', 'approved', 'rejected', 'delivered', 'delivery_failed'
    )),
    reviewer_user_id     UUID REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    reviewed_at          TIMESTAMPTZ,
    review_note TEXT,
    delivery_http_status INT,
    delivery_error       TEXT,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_insert_commons_bundles_farm_hash
    ON gr33ncore.insert_commons_bundles (farm_id, payload_hash);
CREATE INDEX IF NOT EXISTS idx_insert_commons_bundles_farm_status_created
    ON gr33ncore.insert_commons_bundles (farm_id, status, created_at DESC);

-- Insert Commons sync audit (farm-side sender)
CREATE TABLE IF NOT EXISTS gr33ncore.insert_commons_sync_events (
    id               BIGSERIAL PRIMARY KEY,
    farm_id          BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    idempotency_key  TEXT,
    status           TEXT NOT NULL,
    http_status      INT,
    error            TEXT,
    payload          JSONB NOT NULL,
    bundle_id        BIGINT REFERENCES gr33ncore.insert_commons_bundles(id) ON DELETE SET NULL,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_insert_commons_sync_farm_idem
    ON gr33ncore.insert_commons_sync_events (farm_id, idempotency_key)
    WHERE idempotency_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_insert_commons_sync_farm_created
    ON gr33ncore.insert_commons_sync_events (farm_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_insert_commons_sync_bundle
    ON gr33ncore.insert_commons_sync_events (bundle_id)
    WHERE bundle_id IS NOT NULL;

-- Insert Commons receiver (pilot ingest store; optional separate process)
CREATE TABLE IF NOT EXISTS gr33ncore.insert_commons_received_payloads (
    id BIGSERIAL PRIMARY KEY,
    received_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    payload_hash             TEXT        NOT NULL,
    farm_pseudonym           TEXT        NOT NULL,
    schema_version           TEXT        NOT NULL,
    generated_at             TIMESTAMPTZ NOT NULL,
    payload                  JSONB       NOT NULL,
    source_idempotency_key   TEXT        NULL,
    CONSTRAINT uq_insert_commons_received_payload_hash UNIQUE (payload_hash)
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_insert_commons_received_farm_idem
    ON gr33ncore.insert_commons_received_payloads (farm_pseudonym, source_idempotency_key)
    WHERE source_idempotency_key IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_insert_commons_received_received_at
    ON gr33ncore.insert_commons_received_payloads (received_at DESC);

-- Commons catalog (gr33n_inserts — published packs + farm import audit)
CREATE TABLE IF NOT EXISTS gr33ncore.commons_catalog_entries (
    id                   BIGSERIAL PRIMARY KEY,
    slug                 TEXT        NOT NULL UNIQUE,
    title                TEXT        NOT NULL,
    summary              TEXT        NOT NULL DEFAULT '',
    body                 JSONB       NOT NULL DEFAULT '{}'::jsonb,
    contributor_display  TEXT        NOT NULL DEFAULT '',
    contributor_uri      TEXT,
    license_spdx         TEXT        NOT NULL DEFAULT 'CC-BY-4.0',
    license_notes        TEXT,
    tags                 TEXT[]      NOT NULL DEFAULT ARRAY[]::TEXT[],
    published            BOOLEAN     NOT NULL DEFAULT FALSE,
    sort_order           INT         NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_commons_catalog_published_sort
    ON gr33ncore.commons_catalog_entries (published, sort_order, title)
    WHERE published = TRUE;
CREATE TABLE IF NOT EXISTS gr33ncore.farm_commons_catalog_imports (
    id                 BIGSERIAL PRIMARY KEY,
    farm_id            BIGINT      NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    catalog_entry_id   BIGINT      NOT NULL REFERENCES gr33ncore.commons_catalog_entries(id) ON DELETE CASCADE,
    imported_by        UUID        NOT NULL REFERENCES gr33ncore.profiles(user_id) ON DELETE CASCADE,
    imported_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    note               TEXT,
    UNIQUE (farm_id, catalog_entry_id)
);
CREATE INDEX IF NOT EXISTS idx_farm_commons_imports_farm
    ON gr33ncore.farm_commons_catalog_imports (farm_id, imported_at DESC);
CREATE TRIGGER trg_commons_catalog_entries_updated_at
    BEFORE UPDATE ON gr33ncore.commons_catalog_entries
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Farm memberships
CREATE TABLE IF NOT EXISTS gr33ncore.farm_memberships (
    farm_id      BIGINT NOT NULL REFERENCES gr33ncore.farms(id)    ON DELETE CASCADE,
    user_id      UUID   NOT NULL REFERENCES gr33ncore.profiles(user_id) ON DELETE CASCADE,
    role_in_farm gr33ncore.farm_member_role_enum NOT NULL,
    permissions  JSONB  DEFAULT '{}'::jsonb,
    joined_at    TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    PRIMARY KEY (farm_id, user_id)
);

-- Active modules
CREATE TABLE IF NOT EXISTS gr33ncore.farm_active_modules (
    farm_id            BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    module_schema_name TEXT   NOT NULL,
    is_enabled         BOOLEAN DEFAULT TRUE NOT NULL,
    configuration      JSONB   DEFAULT '{}'::jsonb,
    activated_at       TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    PRIMARY KEY (farm_id, module_schema_name)
);

-- Zones
CREATE TABLE IF NOT EXISTS gr33ncore.zones (
    id             BIGSERIAL PRIMARY KEY,
    farm_id        BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    parent_zone_id BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
    name           TEXT   NOT NULL,
    description    TEXT,
    zone_type      TEXT,
    area_sqm       NUMERIC(12,2),
    boundary_gis   GEOMETRY(Polygon,4326),
    meta_data      JSONB  DEFAULT '{}'::jsonb,
    created_at     TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at     TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_by_user_id UUID REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    deleted_at     TIMESTAMPTZ DEFAULT NULL
);
CREATE TRIGGER trg_zones_updated_at
    BEFORE UPDATE ON gr33ncore.zones
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Devices
CREATE TABLE IF NOT EXISTS gr33ncore.devices (
    id                 BIGSERIAL PRIMARY KEY,
    farm_id            BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    zone_id            BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
    name               TEXT   NOT NULL,
    device_uid         TEXT   UNIQUE,
    device_type        TEXT   NOT NULL,
    ip_address         INET,
    firmware_version   TEXT,
    status             gr33ncore.device_status_enum DEFAULT 'unknown' NOT NULL,
    last_heartbeat     TIMESTAMPTZ,
    api_key            TEXT   UNIQUE,
    config             JSONB  DEFAULT '{}'::jsonb,
    meta_data          JSONB  DEFAULT '{}'::jsonb,
    created_at         TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at         TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_by_user_id UUID   REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    deleted_at         TIMESTAMPTZ DEFAULT NULL
);
CREATE TRIGGER trg_devices_updated_at
    BEFORE UPDATE ON gr33ncore.devices
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Sensors
CREATE TABLE IF NOT EXISTS gr33ncore.sensors (
    id                       BIGSERIAL PRIMARY KEY,
    device_id                BIGINT REFERENCES gr33ncore.devices(id) ON DELETE SET NULL,
    farm_id                  BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    zone_id                  BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
    name                     TEXT   NOT NULL,
    sensor_type              TEXT   NOT NULL,
    unit_id                  BIGINT NOT NULL REFERENCES gr33ncore.units(id) ON DELETE RESTRICT,
    hardware_identifier      TEXT,
    value_min_expected       NUMERIC,
    value_max_expected       NUMERIC,
    alert_threshold_low      NUMERIC,
    alert_threshold_high     NUMERIC,
    alert_duration_seconds   INTEGER     NOT NULL DEFAULT 0,
    alert_cooldown_seconds   INTEGER     NOT NULL DEFAULT 300,
    alert_breach_started_at  TIMESTAMPTZ NULL,
    reading_interval_seconds INTEGER,
    is_calibrated            BOOLEAN DEFAULT FALSE,
    last_calibration_date    DATE,
    calibration_data         JSONB,
    config                   JSONB   DEFAULT '{}'::jsonb,
    meta_data                JSONB   DEFAULT '{}'::jsonb,
    created_at               TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at               TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_by_user_id       UUID    REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    deleted_at               TIMESTAMPTZ DEFAULT NULL,
    CONSTRAINT chk_sensor_farm_context CHECK (device_id IS NOT NULL OR farm_id IS NOT NULL)
);
CREATE TRIGGER trg_sensors_updated_at
    BEFORE UPDATE ON gr33ncore.sensors
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Sensor readings
-- FIX #4: PK order is (reading_time, sensor_id) for TimescaleDB hypertable compatibility
CREATE TABLE IF NOT EXISTS gr33ncore.sensor_readings (
    reading_time          TIMESTAMPTZ NOT NULL,
    sensor_id             BIGINT      NOT NULL REFERENCES gr33ncore.sensors(id) ON DELETE CASCADE,
    value_raw             NUMERIC     NOT NULL,
    value_normalized      NUMERIC,
    normalized_unit_id    BIGINT      REFERENCES gr33ncore.units(id),
    value_text            TEXT,
    value_json            JSONB,
    battery_level_percent NUMERIC(5,2)
        CHECK (battery_level_percent IS NULL OR (battery_level_percent >= 0 AND battery_level_percent <= 100)),
    signal_strength_dbm   INTEGER,
    is_valid              BOOLEAN     DEFAULT TRUE,
    meta_data             JSONB       DEFAULT '{}'::jsonb,
    PRIMARY KEY (reading_time, sensor_id)  -- FIX #4: time col first
);

-- Normalization trigger function
CREATE OR REPLACE FUNCTION gr33ncore.normalize_sensor_reading()
RETURNS TRIGGER AS $$
DECLARE
    sensor_unit_record RECORD;
    base_unit_record   RECORD;
BEGIN
    SELECT u.* INTO sensor_unit_record
    FROM gr33ncore.units u
    JOIN gr33ncore.sensors s ON s.unit_id = u.id
    WHERE s.id = NEW.sensor_id;

    SELECT * INTO base_unit_record
    FROM gr33ncore.units
    WHERE unit_type = sensor_unit_record.unit_type AND is_base_unit = TRUE
    LIMIT 1;

    NEW.normalized_unit_id := base_unit_record.id;

    IF sensor_unit_record.unit_type = 'temperature' THEN
        NEW.value_normalized := gr33ncore.convert_temperature(
            NEW.value_raw, sensor_unit_record.name, base_unit_record.name);
    ELSE
        NEW.value_normalized := NEW.value_raw * sensor_unit_record.conversion_to_base;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trigger_normalize_sensor_reading ON gr33ncore.sensor_readings;
CREATE TRIGGER trigger_normalize_sensor_reading
    BEFORE INSERT OR UPDATE ON gr33ncore.sensor_readings
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.normalize_sensor_reading();

-- Actuators
-- FIX #2: device_id is nullable (was NOT NULL with ON DELETE SET NULL - contradiction)
CREATE TABLE IF NOT EXISTS gr33ncore.actuators (
    id                     BIGSERIAL PRIMARY KEY,
    device_id              BIGINT     REFERENCES gr33ncore.devices(id) ON DELETE SET NULL,  -- FIX #2: removed NOT NULL
    farm_id                BIGINT     NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    zone_id                BIGINT     REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
    name                   TEXT       NOT NULL,
    actuator_type          TEXT       NOT NULL,
    hardware_identifier    TEXT,
    current_state_numeric  NUMERIC,
    current_state_text     TEXT,
    last_known_state_time  TIMESTAMPTZ,
    last_command_sent_time TIMESTAMPTZ,
    feedback_sensor_id     BIGINT     REFERENCES gr33ncore.sensors(id) ON DELETE SET NULL,
    config                 JSONB      DEFAULT '{}'::jsonb,
    meta_data              JSONB      DEFAULT '{}'::jsonb,
    created_at             TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at             TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_by_user_id     UUID       REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    deleted_at             TIMESTAMPTZ DEFAULT NULL,
    -- Phase 20.95 WS2: wattage for the nightly electricity rollup (0 = unknown/unmetered).
    watts                  NUMERIC(10,2) DEFAULT 0 NOT NULL
);
CREATE TRIGGER trg_actuators_updated_at
    BEFORE UPDATE ON gr33ncore.actuators
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Actuator events
-- FIX #4: PK order is (event_time, actuator_id) for TimescaleDB compatibility
CREATE TABLE IF NOT EXISTS gr33ncore.actuator_events (
    event_time                     TIMESTAMPTZ NOT NULL,
    actuator_id                    BIGINT      NOT NULL REFERENCES gr33ncore.actuators(id) ON DELETE CASCADE,
    command_sent                   TEXT,
    parameters_sent                JSONB,
    triggered_by_user_id           UUID        REFERENCES gr33ncore.profiles(user_id),
    triggered_by_schedule_id       BIGINT,
    triggered_by_rule_id           BIGINT,
    source                         gr33ncore.actuator_event_source_enum NOT NULL,
    response_received_from_device  TEXT,
    execution_status               gr33ncore.actuator_execution_status_enum,
    resulting_state_numeric_actual NUMERIC,
    resulting_state_text_actual    TEXT,
    meta_data                      JSONB DEFAULT '{}'::jsonb,
    PRIMARY KEY (event_time, actuator_id)  -- FIX #4: time col first
);

-- Schedules
CREATE TABLE IF NOT EXISTS gr33ncore.schedules (
    id                         BIGSERIAL PRIMARY KEY,
    farm_id                    BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    name                       TEXT   NOT NULL,
    description                TEXT,
    schedule_type              TEXT   NOT NULL,
    cron_expression            TEXT   NOT NULL,
    timezone                   TEXT   DEFAULT 'UTC' NOT NULL,
    is_active                  BOOLEAN DEFAULT TRUE NOT NULL,
    last_triggered_time        TIMESTAMPTZ,
    next_expected_trigger_time TIMESTAMPTZ,
    meta_data                  JSONB  DEFAULT '{}'::jsonb,
    preconditions              JSONB  NOT NULL DEFAULT '[]'::jsonb
        CHECK (jsonb_typeof(preconditions) = 'array'),
    created_at                 TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at                 TIMESTAMPTZ DEFAULT NOW() NOT NULL
);
CREATE TRIGGER trg_schedules_updated_at
    BEFORE UPDATE ON gr33ncore.schedules
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Tasks (after schedules — schedule_id FK)
CREATE TABLE IF NOT EXISTS gr33ncore.tasks (
    id                         BIGSERIAL PRIMARY KEY,
    farm_id                    BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    zone_id                    BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
    schedule_id                BIGINT REFERENCES gr33ncore.schedules(id) ON DELETE SET NULL,
    title                      TEXT   NOT NULL,
    description                TEXT,
    task_type                  TEXT,
    status                     gr33ncore.task_status_enum DEFAULT 'todo' NOT NULL,
    priority                   INTEGER DEFAULT 1 CHECK (priority BETWEEN 0 AND 3),
    assigned_to_user_id        UUID   REFERENCES gr33ncore.profiles(user_id),
    due_date                   DATE,
    estimated_duration_minutes INTEGER,
    actual_start_time          TIMESTAMPTZ,
    actual_end_time            TIMESTAMPTZ,
    related_module_schema      TEXT,
    related_table_name         TEXT,
    related_record_id          BIGINT,
    source_alert_id            BIGINT REFERENCES gr33ncore.alerts_notifications(id) ON DELETE SET NULL,
    source_rule_id             BIGINT REFERENCES gr33ncore.automation_rules(id) ON DELETE SET NULL,
    created_by_user_id         UUID   REFERENCES gr33ncore.profiles(user_id),
    created_at                 TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at                 TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_by_user_id         UUID   REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    deleted_at                 TIMESTAMPTZ DEFAULT NULL,
    -- Phase 20.95 WS1: denormalised SUM(task_labor_log.minutes) maintained by handler.
    time_spent_minutes         INTEGER
);
CREATE INDEX IF NOT EXISTS idx_tasks_source_alert_id
    ON gr33ncore.tasks (source_alert_id)
    WHERE source_alert_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_source_rule_id
    ON gr33ncore.tasks (source_rule_id)
    WHERE source_rule_id IS NOT NULL;
CREATE TRIGGER trg_tasks_updated_at
    BEFORE UPDATE ON gr33ncore.tasks
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Phase 20.95 WS1 — task labor log (minutes + optional hourly-rate snapshot).
-- tasks.time_spent_minutes is a running SUM over surviving log rows,
-- written by the task labor handler on every insert/delete.
CREATE TABLE IF NOT EXISTS gr33ncore.task_labor_log (
    id                    BIGSERIAL PRIMARY KEY,
    farm_id               BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    task_id               BIGINT NOT NULL REFERENCES gr33ncore.tasks(id) ON DELETE CASCADE,
    user_id               UUID   REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    started_at            TIMESTAMPTZ NOT NULL,
    ended_at              TIMESTAMPTZ,
    minutes               INTEGER NOT NULL CHECK (minutes >= 0),
    hourly_rate_snapshot  NUMERIC(10,2),
    currency              CHAR(3) CHECK (currency IS NULL OR currency ~ '^[A-Z]{3}$'),
    notes                 TEXT,
    created_at            TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at            TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_task_labor_log_task
    ON gr33ncore.task_labor_log (task_id);
CREATE INDEX IF NOT EXISTS idx_task_labor_log_farm
    ON gr33ncore.task_labor_log (farm_id);
CREATE TRIGGER trg_task_labor_log_updated_at
    BEFORE UPDATE ON gr33ncore.task_labor_log
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Automation rules
CREATE TABLE IF NOT EXISTS gr33ncore.automation_rules (
    id                   BIGSERIAL PRIMARY KEY,
    farm_id              BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    name                 TEXT   NOT NULL,
    description          TEXT,
    is_active            BOOLEAN DEFAULT TRUE NOT NULL,
    trigger_source       gr33ncore.automation_trigger_source_enum NOT NULL,
    trigger_configuration JSONB NOT NULL,
    condition_logic      TEXT  DEFAULT 'ALL' CHECK (condition_logic IN ('ALL','ANY')),
    conditions_jsonb     JSONB DEFAULT '[]'::jsonb,
    last_evaluated_time  TIMESTAMPTZ,
    last_triggered_time  TIMESTAMPTZ,
    cooldown_period_seconds INTEGER DEFAULT 0,
    created_at           TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at           TIMESTAMPTZ DEFAULT NOW() NOT NULL
);
CREATE TRIGGER trg_automation_rules_updated_at
    BEFORE UPDATE ON gr33ncore.automation_rules
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Executable actions
CREATE TABLE IF NOT EXISTS gr33ncore.executable_actions (
    id                          BIGSERIAL PRIMARY KEY,
    schedule_id                 BIGINT REFERENCES gr33ncore.schedules(id) ON DELETE CASCADE,
    rule_id                     BIGINT REFERENCES gr33ncore.automation_rules(id) ON DELETE CASCADE,
    program_id                  BIGINT, -- Phase 20.95 WS3; FK added after gr33nfertigation.programs is created below.
    execution_order             INTEGER DEFAULT 0 NOT NULL,
    action_type                 gr33ncore.executable_action_type_enum NOT NULL,
    target_actuator_id          BIGINT REFERENCES gr33ncore.actuators(id) ON DELETE SET NULL,
    target_automation_rule_id   BIGINT REFERENCES gr33ncore.automation_rules(id) ON DELETE CASCADE,
    target_notification_template_id BIGINT,
    action_command              TEXT,
    action_parameters           JSONB,
    delay_before_execution_seconds INTEGER DEFAULT 0,
    CONSTRAINT chk_executable_source CHECK (num_nonnulls(schedule_id, rule_id, program_id) = 1),
    CONSTRAINT chk_executable_action_details CHECK (
        (action_type = 'control_actuator'              AND target_actuator_id IS NOT NULL AND action_command IS NOT NULL) OR
        (action_type = 'trigger_another_automation_rule' AND target_automation_rule_id IS NOT NULL) OR
        (action_type = 'send_notification'             AND target_notification_template_id IS NOT NULL) OR
        (action_type = 'create_task'                   AND action_parameters IS NOT NULL) OR
        (action_type = 'log_custom_event'              AND action_parameters IS NOT NULL) OR
        (action_type = 'http_webhook_call'             AND action_parameters->>'url' IS NOT NULL) OR
        (action_type = 'update_record_in_gr33n'        AND action_parameters->>'target_module_schema' IS NOT NULL
                                                       AND action_parameters->>'target_table_name' IS NOT NULL
                                                       AND action_parameters->'fields_to_update' IS NOT NULL)
    )
);

-- Automation run log (observability for scheduler executions)
CREATE TABLE IF NOT EXISTS gr33ncore.automation_runs (
    id              BIGSERIAL PRIMARY KEY,
    farm_id         BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    schedule_id     BIGINT REFERENCES gr33ncore.schedules(id) ON DELETE SET NULL,
    rule_id         BIGINT REFERENCES gr33ncore.automation_rules(id) ON DELETE SET NULL,
    status          TEXT NOT NULL CHECK (status IN ('success', 'partial_success', 'failed', 'skipped')),
    message         TEXT,
    details         JSONB DEFAULT '{}'::jsonb,
    executed_at     TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- FK back-references for actuator_events
ALTER TABLE gr33ncore.actuator_events
    ADD CONSTRAINT fk_actuator_event_schedule
    FOREIGN KEY (triggered_by_schedule_id) REFERENCES gr33ncore.schedules(id) ON DELETE SET NULL;
ALTER TABLE gr33ncore.actuator_events
    ADD CONSTRAINT fk_actuator_event_rule
    FOREIGN KEY (triggered_by_rule_id) REFERENCES gr33ncore.automation_rules(id) ON DELETE SET NULL;

-- Notification templates
CREATE TABLE IF NOT EXISTS gr33ncore.notification_templates (
    id                       BIGSERIAL PRIMARY KEY,
    farm_id                  BIGINT REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    template_key             TEXT   NOT NULL,
    description              TEXT,
    subject_template         TEXT,
    body_template_text       TEXT,
    body_template_html       TEXT,
    default_delivery_channels TEXT[] DEFAULT ARRAY['in_app','email']::TEXT[],
    default_priority         gr33ncore.notification_priority_enum DEFAULT 'medium',
    is_system_template       BOOLEAN DEFAULT FALSE,
    created_at               TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at               TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT uq_notification_template_key UNIQUE (farm_id, template_key)
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_system_notification_template_key
    ON gr33ncore.notification_templates (template_key) WHERE farm_id IS NULL;
CREATE TRIGGER trg_notification_templates_updated_at
    BEFORE UPDATE ON gr33ncore.notification_templates
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

ALTER TABLE gr33ncore.executable_actions
    ADD CONSTRAINT fk_action_notification_template
    FOREIGN KEY (target_notification_template_id)
    REFERENCES gr33ncore.notification_templates(id) ON DELETE SET NULL;

-- Alerts & notifications
CREATE TABLE IF NOT EXISTS gr33ncore.alerts_notifications (
    id                       BIGSERIAL PRIMARY KEY,
    farm_id                  BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    recipient_user_id        UUID   REFERENCES gr33ncore.profiles(user_id),
    notification_template_id BIGINT REFERENCES gr33ncore.notification_templates(id) ON DELETE SET NULL,
    triggering_event_source_type TEXT,
    triggering_event_source_id   BIGINT,
    severity                 gr33ncore.notification_priority_enum DEFAULT 'medium',
    subject_rendered         TEXT,
    message_text_rendered    TEXT,
    message_html_rendered    TEXT,
    delivery_attempts        JSONB  DEFAULT '{}'::jsonb,
    status                   gr33ncore.notification_status_enum DEFAULT 'pending',
    is_read                  BOOLEAN DEFAULT FALSE,
    read_at                  TIMESTAMPTZ,
    is_acknowledged          BOOLEAN DEFAULT FALSE,
    acknowledged_at          TIMESTAMPTZ,
    acknowledged_by_user_id  UUID   REFERENCES gr33ncore.profiles(user_id),
    created_at               TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    scheduled_send_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS gr33ncore.user_push_tokens (
    id         BIGSERIAL PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES gr33ncore.profiles(user_id) ON DELETE CASCADE,
    platform   TEXT NOT NULL CHECK (platform IN ('android', 'ios', 'web')),
    fcm_token  TEXT NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_push_tokens_user_id
    ON gr33ncore.user_push_tokens (user_id);

CREATE TRIGGER trg_user_push_tokens_updated_at
    BEFORE UPDATE ON gr33ncore.user_push_tokens
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- System logs
CREATE TABLE IF NOT EXISTS gr33ncore.system_logs (
    id               BIGSERIAL NOT NULL,
    farm_id          BIGINT REFERENCES gr33ncore.farms(id) ON DELETE SET NULL,
    user_id          UUID   REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    log_time         TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    log_level        gr33ncore.log_level_enum NOT NULL,
    event_type       TEXT,
    message          TEXT   NOT NULL,
    source_component TEXT,
    context_data     JSONB  DEFAULT '{}'::jsonb,
    created_at       TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    PRIMARY KEY (log_time, id)
);

-- File attachments
CREATE TABLE IF NOT EXISTS gr33ncore.file_attachments (
    id                  BIGSERIAL PRIMARY KEY,
    farm_id             BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    related_module_schema TEXT  NOT NULL,
    related_table_name  TEXT   NOT NULL,
    related_record_id   TEXT   NOT NULL,
    file_name           TEXT   NOT NULL,
    file_type           TEXT   NOT NULL,
    file_size_bytes     BIGINT,
    storage_path        TEXT   NOT NULL,
    mime_type           TEXT,
    description         TEXT,
    uploaded_by_user_id UUID   REFERENCES gr33ncore.profiles(user_id),
    created_at          TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at          TIMESTAMPTZ DEFAULT NOW() NOT NULL
);
CREATE TRIGGER trg_file_attachments_updated_at
    BEFORE UPDATE ON gr33ncore.file_attachments
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Weather data
CREATE TABLE IF NOT EXISTS gr33ncore.weather_data (
    id                     BIGSERIAL NOT NULL,
    farm_id                BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    zone_id                BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
    recorded_at            TIMESTAMPTZ NOT NULL,
    data_source            gr33ncore.weather_data_source_enum NOT NULL,
    source_sensor_id       BIGINT REFERENCES gr33ncore.sensors(id) ON DELETE SET NULL,
    temperature_celsius    NUMERIC(5,2)
        CHECK (temperature_celsius IS NULL OR (temperature_celsius > -60 AND temperature_celsius < 70)),
    humidity_percent       NUMERIC(5,2)
        CHECK (humidity_percent IS NULL OR (humidity_percent >= 0 AND humidity_percent <= 100)),
    precipitation_mm       NUMERIC(6,2)
        CHECK (precipitation_mm IS NULL OR precipitation_mm >= 0),
    wind_speed_ms          NUMERIC(5,2)
        CHECK (wind_speed_ms IS NULL OR wind_speed_ms >= 0),
    wind_direction_degrees INTEGER
        CHECK (wind_direction_degrees IS NULL OR (wind_direction_degrees >= 0 AND wind_direction_degrees <= 360)),
    barometric_pressure_hpa NUMERIC(7,2)
        CHECK (barometric_pressure_hpa IS NULL OR (barometric_pressure_hpa > 500 AND barometric_pressure_hpa < 1200)),
    solar_radiation_wm2    NUMERIC(8,2)
        CHECK (solar_radiation_wm2 IS NULL OR solar_radiation_wm2 >= 0),
    dew_point_celsius      NUMERIC(5,2),
    uv_index               NUMERIC(4,1)
        CHECK (uv_index IS NULL OR uv_index >= 0),
    cloud_cover_percent    NUMERIC(5,2)
        CHECK (cloud_cover_percent IS NULL OR (cloud_cover_percent >= 0 AND cloud_cover_percent <= 100)),
    forecast_data          JSONB,
    raw_data               JSONB,
    created_at             TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    PRIMARY KEY (recorded_at, id)
    -- NOTE: unique constraint removed due to NULL pitfall with zone_id/source_sensor_id.
    -- Use partial unique index below instead.
);

-- Cost transactions
CREATE TABLE IF NOT EXISTS gr33ncore.cost_transactions (
    id               BIGSERIAL PRIMARY KEY,
    farm_id          BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    transaction_date DATE   NOT NULL,
    category         gr33ncore.cost_category_enum NOT NULL,
    subcategory      TEXT,
    amount           NUMERIC(12,2) NOT NULL,
    currency         CHAR(3)       NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    description      TEXT,
    related_module_schema TEXT,
    related_table_name    TEXT,
    related_record_id     BIGINT,
    receipt_file_id  BIGINT REFERENCES gr33ncore.file_attachments(id) ON DELETE SET NULL,
    is_income        BOOLEAN DEFAULT FALSE NOT NULL,
    document_type      TEXT,
    document_reference TEXT,
    counterparty       TEXT,
    created_by_user_id UUID REFERENCES gr33ncore.profiles(user_id),
    created_at       TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at       TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    -- Phase 20.95 WS2: optional link so Phase 21 can report "$ per cycle".
    crop_cycle_id    BIGINT REFERENCES gr33nfertigation.crop_cycles(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_cost_tx_crop_cycle
    ON gr33ncore.cost_transactions (crop_cycle_id)
    WHERE crop_cycle_id IS NOT NULL;
CREATE TRIGGER trg_cost_transactions_updated_at
    BEFORE UPDATE ON gr33ncore.cost_transactions
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Phase 20.95 WS2 — per-farm $/kWh pricing history. Phase 20.7 WS4
-- will multiply actuators.watts * runtime_hours * price_per_kwh.
CREATE TABLE IF NOT EXISTS gr33ncore.farm_energy_prices (
    id               BIGSERIAL PRIMARY KEY,
    farm_id          BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    effective_from   DATE   NOT NULL,
    effective_to     DATE,
    price_per_kwh    NUMERIC(10,4) NOT NULL CHECK (price_per_kwh >= 0),
    currency         CHAR(3)       NOT NULL CHECK (currency ~ '^[A-Z]{3}$'),
    notes            TEXT,
    created_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
    updated_at       TIMESTAMPTZ   NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_farm_energy_prices_active
    ON gr33ncore.farm_energy_prices (farm_id, effective_from DESC);
CREATE TRIGGER trg_farm_energy_prices_updated_at
    BEFORE UPDATE ON gr33ncore.farm_energy_prices
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Cost transaction idempotency (offline / safe retries) — must follow cost_transactions
CREATE TABLE IF NOT EXISTS gr33ncore.cost_transaction_idempotency (
    farm_id              BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    idempotency_key      TEXT   NOT NULL,
    cost_transaction_id  BIGINT NOT NULL REFERENCES gr33ncore.cost_transactions(id) ON DELETE CASCADE,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (farm_id, idempotency_key)
);
CREATE INDEX IF NOT EXISTS idx_cost_idem_transaction
    ON gr33ncore.cost_transaction_idempotency (cost_transaction_id);

-- Farm finance COA mapping overrides (used for GL exports)
CREATE TABLE IF NOT EXISTS gr33ncore.farm_finance_account_mappings (
    id            BIGSERIAL PRIMARY KEY,
    farm_id       BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    cost_category gr33ncore.cost_category_enum NOT NULL,
    account_code  TEXT   NOT NULL,
    account_name  TEXT   NOT NULL,
    is_active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at    TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at    TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    UNIQUE (farm_id, cost_category)
);
CREATE TRIGGER trg_farm_finance_account_mappings_updated_at
    BEFORE UPDATE ON gr33ncore.farm_finance_account_mappings
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Validation rules
CREATE TABLE IF NOT EXISTS gr33ncore.validation_rules (
    id                   BIGSERIAL PRIMARY KEY,
    farm_id              BIGINT REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    rule_name            TEXT   NOT NULL,
    description          TEXT,
    target_module_schema TEXT   NOT NULL,
    target_table_name    TEXT   NOT NULL,
    target_column_name   TEXT   NOT NULL,
    rule_type            gr33ncore.validation_rule_type_enum NOT NULL,
    rule_config          JSONB  NOT NULL,
    error_message_template TEXT,
    is_active            BOOLEAN DEFAULT TRUE NOT NULL,
    severity             gr33ncore.validation_severity_enum DEFAULT 'error' NOT NULL,
    evaluation_trigger   TEXT    DEFAULT 'on_save' NOT NULL
        CHECK (evaluation_trigger IN ('on_save','on_change','manual_batch')),
    created_at           TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at           TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT uq_validation_rule_name_farm UNIQUE (farm_id, rule_name)
);
CREATE UNIQUE INDEX IF NOT EXISTS uq_validation_rule_name_global
    ON gr33ncore.validation_rules (rule_name) WHERE farm_id IS NULL;
CREATE TRIGGER trg_validation_rules_updated_at
    BEFORE UPDATE ON gr33ncore.validation_rules
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- User activity log
CREATE TABLE IF NOT EXISTS gr33ncore.user_activity_log (
    id                       BIGSERIAL NOT NULL,
    user_id                  UUID   REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    farm_id                  BIGINT REFERENCES gr33ncore.farms(id) ON DELETE SET NULL,
    activity_time            TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    action_type              gr33ncore.user_action_type_enum NOT NULL,
    target_module_schema     TEXT,
    target_table_name        TEXT,
    target_record_id         TEXT,
    target_record_description TEXT,
    ip_address               INET,
    user_agent               TEXT,
    session_id               TEXT,
    status                   TEXT CHECK (status IN ('success','failure','pending')),
    failure_reason           TEXT,
    details                  JSONB DEFAULT '{}'::jsonb,
    created_at               TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    PRIMARY KEY (activity_time, id)
);

-- ============================================================
-- INDEXES
-- ============================================================

-- Units
CREATE INDEX IF NOT EXISTS idx_units_type ON gr33ncore.units(unit_type);
CREATE INDEX IF NOT EXISTS idx_units_base ON gr33ncore.units(is_base_unit) WHERE is_base_unit = TRUE;

-- Soft-delete partial indexes
CREATE INDEX IF NOT EXISTS idx_farms_active    ON gr33ncore.farms(id)    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_zones_active    ON gr33ncore.zones(id)    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_devices_active  ON gr33ncore.devices(id)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_sensors_active  ON gr33ncore.sensors(id)  WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_tasks_active    ON gr33ncore.tasks(id)    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_actuators_active ON gr33ncore.actuators(id) WHERE deleted_at IS NULL;

-- PostGIS spatial
CREATE INDEX IF NOT EXISTS idx_farms_location_gis  ON gr33ncore.farms USING GIST (location_gis);
CREATE INDEX IF NOT EXISTS idx_zones_boundary_gis  ON gr33ncore.zones USING GIST (boundary_gis);

-- Sensor readings
-- FIX #6: (sensor_id, reading_time DESC) is the primary query pattern
CREATE INDEX IF NOT EXISTS idx_sensor_readings_sensor_time
    ON gr33ncore.sensor_readings(sensor_id, reading_time DESC);
CREATE INDEX IF NOT EXISTS idx_sensor_readings_time_desc
    ON gr33ncore.sensor_readings(reading_time DESC);
CREATE INDEX IF NOT EXISTS idx_sensor_readings_normalized_unit
    ON gr33ncore.sensor_readings(normalized_unit_id);

-- Farms
CREATE INDEX IF NOT EXISTS idx_farms_owner   ON gr33ncore.farms(owner_user_id);
CREATE INDEX IF NOT EXISTS idx_farms_name    ON gr33ncore.farms(name text_pattern_ops);

-- Zones
CREATE INDEX IF NOT EXISTS idx_zones_farm    ON gr33ncore.zones(farm_id);
CREATE INDEX IF NOT EXISTS idx_zones_parent  ON gr33ncore.zones(parent_zone_id);

-- Devices
CREATE INDEX IF NOT EXISTS idx_devices_farm_zone ON gr33ncore.devices(farm_id, zone_id);
CREATE INDEX IF NOT EXISTS idx_devices_status    ON gr33ncore.devices(farm_id, status);
CREATE INDEX IF NOT EXISTS idx_devices_type      ON gr33ncore.devices(farm_id, device_type);

-- Sensors
CREATE INDEX IF NOT EXISTS idx_sensors_farm_type ON gr33ncore.sensors(farm_id, sensor_type);
CREATE INDEX IF NOT EXISTS idx_sensors_farm_zone ON gr33ncore.sensors(farm_id, zone_id);
CREATE INDEX IF NOT EXISTS idx_sensors_device    ON gr33ncore.sensors(device_id);
CREATE INDEX IF NOT EXISTS idx_sensors_unit      ON gr33ncore.sensors(unit_id);

-- Actuator events
CREATE INDEX IF NOT EXISTS idx_actuator_events_time_desc
    ON gr33ncore.actuator_events(event_time DESC);
CREATE INDEX IF NOT EXISTS idx_actuator_events_actuator_time
    ON gr33ncore.actuator_events(actuator_id, event_time DESC);

-- Tasks
CREATE INDEX IF NOT EXISTS idx_tasks_assignment
    ON gr33ncore.tasks(farm_id, assigned_to_user_id, status);
CREATE INDEX IF NOT EXISTS idx_tasks_status_due
    ON gr33ncore.tasks(farm_id, status, due_date);
CREATE INDEX IF NOT EXISTS idx_tasks_related
    ON gr33ncore.tasks(related_module_schema, related_table_name, related_record_id);

-- Alerts
CREATE INDEX IF NOT EXISTS idx_alerts_user_status
    ON gr33ncore.alerts_notifications(recipient_user_id, status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_farm_status
    ON gr33ncore.alerts_notifications(farm_id, status, created_at DESC);

-- System logs
CREATE INDEX IF NOT EXISTS idx_system_logs_farm_time
    ON gr33ncore.system_logs(farm_id, log_time DESC);
CREATE INDEX IF NOT EXISTS idx_system_logs_level_time
    ON gr33ncore.system_logs(log_level, log_time DESC);

-- Weather data
CREATE INDEX IF NOT EXISTS idx_weather_farm_zone_time
    ON gr33ncore.weather_data(farm_id, zone_id, recorded_at DESC);
-- FIX #9: partial unique index replaces nullable UNIQUE constraint
CREATE UNIQUE INDEX IF NOT EXISTS uq_weather_data_active
    ON gr33ncore.weather_data(farm_id, recorded_at, data_source)
    WHERE zone_id IS NULL AND source_sensor_id IS NULL;

-- Cost transactions
CREATE INDEX IF NOT EXISTS idx_cost_farm_date
    ON gr33ncore.cost_transactions(farm_id, transaction_date DESC);
CREATE INDEX IF NOT EXISTS idx_cost_category
    ON gr33ncore.cost_transactions(farm_id, category, subcategory);

-- Validation rules
CREATE INDEX IF NOT EXISTS idx_validation_target
    ON gr33ncore.validation_rules(target_module_schema, target_table_name, target_column_name);

-- User activity
CREATE INDEX IF NOT EXISTS idx_activity_user_time
    ON gr33ncore.user_activity_log(user_id, activity_time DESC);
CREATE INDEX IF NOT EXISTS idx_activity_farm_time
    ON gr33ncore.user_activity_log(farm_id, activity_time DESC);
CREATE INDEX IF NOT EXISTS idx_activity_action_type
    ON gr33ncore.user_activity_log(action_type, activity_time DESC);
CREATE INDEX IF NOT EXISTS idx_automation_runs_farm_time
    ON gr33ncore.automation_runs(farm_id, executed_at DESC);
CREATE INDEX IF NOT EXISTS idx_automation_runs_schedule_time
    ON gr33ncore.automation_runs(schedule_id, executed_at DESC);

-- ============================================================
-- TIMESCALEDB HYPERTABLE CONVERSIONS
-- Uncomment AFTER enabling the extension.
-- PK column order is now correct (time col first) for all tables.
-- ============================================================
SELECT create_hypertable('gr33ncore.sensor_readings',  'reading_time', if_not_exists => TRUE, chunk_time_interval => INTERVAL '1 day');
SELECT create_hypertable('gr33ncore.actuator_events',  'event_time',   if_not_exists => TRUE, chunk_time_interval => INTERVAL '1 day');
SELECT create_hypertable('gr33ncore.weather_data',     'recorded_at',  if_not_exists => TRUE, chunk_time_interval => INTERVAL '7 days');
SELECT create_hypertable('gr33ncore.user_activity_log','activity_time',if_not_exists => TRUE, chunk_time_interval => INTERVAL '7 days');
SELECT create_hypertable('gr33ncore.system_logs',      'log_time',     if_not_exists => TRUE, chunk_time_interval => INTERVAL '7 days');

-- ============================================================
-- SCHEMA: gr33n_natural_farming
-- ============================================================
CREATE SCHEMA IF NOT EXISTS gr33nnaturalfarming;

CREATE TYPE gr33nnaturalfarming.input_category_enum AS ENUM (
    'microbial_inoculant','fermented_plant_juice','water_soluble_nutrient','oriental_herbal_nutrient',
    'fish_amino_acid','insect_attractant_repellent','soil_conditioner','compost_tea_extract',
    'biochar_preparation','other_ferment','other_extract',
    -- Phase 20.95 WS2 — livestock categories so animal feed / bedding / vet supply cost correctly.
    'animal_feed','bedding','veterinary_supply'
);
CREATE TYPE gr33nnaturalfarming.input_batch_status_enum AS ENUM (
    'planning','ingredients_gathered','mixing_in_progress','fermenting_brewing','maturing_aging',
    'ready_for_use','partially_used','fully_used','expired_discarded','failed_production'
);
CREATE TYPE gr33nnaturalfarming.application_target_enum AS ENUM (
    'soil_drench','foliar_spray','seed_treatment','compost_pile_inoculant',
    'livestock_water_supplement','other'
);

-- Input definitions
CREATE TABLE IF NOT EXISTS gr33nnaturalfarming.input_definitions (
    id                   BIGSERIAL PRIMARY KEY,
    farm_id              BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    name                 TEXT   NOT NULL,
    category             gr33nnaturalfarming.input_category_enum NOT NULL,
    description          TEXT,
    typical_ingredients  TEXT,
    preparation_summary  TEXT,
    storage_guidelines   TEXT,
    safety_precautions   TEXT,
    reference_source     TEXT,
    file_attachment_id   BIGINT REFERENCES gr33ncore.file_attachments(id) ON DELETE SET NULL,
    created_at           TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at           TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_by_user_id   UUID   REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    deleted_at           TIMESTAMPTZ DEFAULT NULL,
    -- Phase 20.95 WS2 — optional per-input unit-cost metadata for auto-cost rollups.
    unit_cost            NUMERIC(12,4),
    unit_cost_currency   CHAR(3) CHECK (unit_cost_currency IS NULL OR unit_cost_currency ~ '^[A-Z]{3}$'),
    unit_cost_unit_id    BIGINT REFERENCES gr33ncore.units(id) ON DELETE SET NULL
);
-- FIX #7: partial unique index instead of UNIQUE with deleted_at
CREATE UNIQUE INDEX IF NOT EXISTS uq_input_definition_farm_name_active
    ON gr33nnaturalfarming.input_definitions(farm_id, name)
    WHERE deleted_at IS NULL;

CREATE TRIGGER trg_input_definitions_updated_at
    BEFORE UPDATE ON gr33nnaturalfarming.input_definitions
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Input batches
CREATE TABLE IF NOT EXISTS gr33nnaturalfarming.input_batches (
    id                        BIGSERIAL PRIMARY KEY,
    farm_id                   BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    input_definition_id       BIGINT NOT NULL REFERENCES gr33nnaturalfarming.input_definitions(id) ON DELETE RESTRICT,
    batch_identifier          TEXT,
    creation_start_date       DATE   NOT NULL,
    creation_end_date         DATE,
    expected_ready_date       DATE,
    actual_ready_date         DATE,
    quantity_produced         NUMERIC(10,2),
    quantity_unit_id          BIGINT REFERENCES gr33ncore.units(id) ON DELETE RESTRICT,
    current_quantity_remaining NUMERIC(10,2),
    status                    gr33nnaturalfarming.input_batch_status_enum DEFAULT 'planning' NOT NULL,
    storage_location          TEXT,
    shelf_life_days           INTEGER,
    ph_value                  NUMERIC(4,2) CHECK (ph_value IS NULL OR (0 <= ph_value AND ph_value <= 14)),
    ec_value_ms_cm            NUMERIC(6,2) CHECK (ec_value_ms_cm IS NULL OR ec_value_ms_cm >= 0),
    temperature_during_making TEXT,
    ingredients_used          TEXT,
    procedure_followed        TEXT,
    observations_notes        TEXT,
    made_by_user_id           UUID   REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    related_task_id           BIGINT REFERENCES gr33ncore.tasks(id) ON DELETE SET NULL,
    file_attachment_id        BIGINT REFERENCES gr33ncore.file_attachments(id) ON DELETE SET NULL,
    created_at                TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at                TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_by_user_id        UUID   REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    deleted_at                TIMESTAMPTZ DEFAULT NULL,
    -- Phase 20.95 WS2 — low-stock trigger marker for the future alerting rollup.
    low_stock_threshold       NUMERIC(12,4),
    CONSTRAINT chk_quantity_consistency CHECK (
        (quantity_produced IS NULL) OR
        (quantity_unit_id IS NOT NULL AND current_quantity_remaining <= quantity_produced)
    )
);
-- FIX #7: partial unique index for active batches
CREATE UNIQUE INDEX IF NOT EXISTS uq_input_batch_farm_identifier_active
    ON gr33nnaturalfarming.input_batches(farm_id, batch_identifier)
    WHERE deleted_at IS NULL AND batch_identifier IS NOT NULL;

CREATE TRIGGER trg_input_batches_updated_at
    BEFORE UPDATE ON gr33nnaturalfarming.input_batches
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Application recipes
CREATE TABLE IF NOT EXISTS gr33nnaturalfarming.application_recipes (
    id                      BIGSERIAL PRIMARY KEY,
    farm_id                 BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    name                    TEXT   NOT NULL,
    input_definition_id     BIGINT REFERENCES gr33nnaturalfarming.input_definitions(id) ON DELETE SET NULL,
    description             TEXT,
    target_application_type gr33nnaturalfarming.application_target_enum NOT NULL,
    dilution_ratio          TEXT,
    components              JSONB  DEFAULT '{}'::jsonb,
    instructions            TEXT,
    frequency_guidelines    TEXT,
    target_crop_categories  TEXT[],
    target_growth_stages    TEXT[],
    notes                   TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at              TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_by_user_id      UUID   REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    deleted_at              TIMESTAMPTZ DEFAULT NULL
);
-- FIX #7: partial unique index for active recipes
CREATE UNIQUE INDEX IF NOT EXISTS uq_application_recipe_farm_name_active
    ON gr33nnaturalfarming.application_recipes(farm_id, name)
    WHERE deleted_at IS NULL;

CREATE TRIGGER trg_application_recipes_updated_at
    BEFORE UPDATE ON gr33nnaturalfarming.application_recipes
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- Recipe components
CREATE TABLE IF NOT EXISTS gr33nnaturalfarming.recipe_input_components (
    application_recipe_id BIGINT NOT NULL
        REFERENCES gr33nnaturalfarming.application_recipes(id) ON DELETE CASCADE,
    input_definition_id   BIGINT NOT NULL
        REFERENCES gr33nnaturalfarming.input_definitions(id) ON DELETE CASCADE,
    part_value            NUMERIC(10,3) NOT NULL,
    part_unit_id          BIGINT REFERENCES gr33ncore.units(id) ON DELETE RESTRICT,
    notes                 TEXT,
    PRIMARY KEY (application_recipe_id, input_definition_id)
);

-- Natural farming indexes
CREATE INDEX IF NOT EXISTS idx_input_batches_unit
    ON gr33nnaturalfarming.input_batches(quantity_unit_id);
CREATE INDEX IF NOT EXISTS idx_input_batches_farm
    ON gr33nnaturalfarming.input_batches(farm_id);
CREATE INDEX IF NOT EXISTS idx_recipe_components_unit
    ON gr33nnaturalfarming.recipe_input_components(part_unit_id);
CREATE INDEX IF NOT EXISTS idx_input_definitions_farm
    ON gr33nnaturalfarming.input_definitions(farm_id) WHERE deleted_at IS NULL;

-- ============================================================
-- SCHEMA: gr33n_fertigation
-- ============================================================
CREATE SCHEMA IF NOT EXISTS gr33nfertigation;

CREATE TYPE gr33nfertigation.growth_stage_enum AS ENUM (
    'clone', 'seedling', 'early_veg', 'late_veg',
    'transition', 'early_flower', 'mid_flower', 'late_flower',
    'flush', 'harvest', 'dry_cure'
);
CREATE TYPE gr33nfertigation.program_trigger_enum AS ENUM (
    'manual',
    'schedule_cron',
    'ec_threshold_low',
    'ph_out_of_range',
    'automation_rule',
    'pi_client_local'
);
CREATE TYPE gr33nfertigation.reservoir_status_enum AS ENUM (
    'ready', 'mixing', 'needs_top_up', 'needs_flush',
    'flushing', 'offline', 'empty'
);

CREATE TABLE IF NOT EXISTS gr33nfertigation.reservoirs (
    id                      BIGSERIAL PRIMARY KEY,
    farm_id                 BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    zone_id                 BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
    name                    TEXT NOT NULL,
    description             TEXT,
    capacity_liters         NUMERIC(10,2) NOT NULL CHECK (capacity_liters > 0),
    current_volume_liters   NUMERIC(10,2) CHECK (current_volume_liters >= 0 AND current_volume_liters <= capacity_liters),
    status                  gr33nfertigation.reservoir_status_enum DEFAULT 'ready' NOT NULL,
    ec_sensor_id            BIGINT REFERENCES gr33ncore.sensors(id) ON DELETE SET NULL,
    ph_sensor_id            BIGINT REFERENCES gr33ncore.sensors(id) ON DELETE SET NULL,
    temp_sensor_id          BIGINT REFERENCES gr33ncore.sensors(id) ON DELETE SET NULL,
    water_level_sensor_id   BIGINT REFERENCES gr33ncore.sensors(id) ON DELETE SET NULL,
    delivery_actuator_id    BIGINT REFERENCES gr33ncore.actuators(id) ON DELETE SET NULL,
    last_ec_mscm            NUMERIC(6,3) CHECK (last_ec_mscm >= 0),
    last_ph                 NUMERIC(4,2) CHECK (last_ph >= 0 AND last_ph <= 14),
    last_reading_time       TIMESTAMPTZ,
    metadata                JSONB DEFAULT '{}',
    created_at              TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at              TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at              TIMESTAMPTZ DEFAULT NULL,
    CONSTRAINT uq_reservoir_farm_name UNIQUE (farm_id, name)
);
CREATE TRIGGER trg_reservoirs_updated_at
    BEFORE UPDATE ON gr33nfertigation.reservoirs
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

CREATE TABLE IF NOT EXISTS gr33nfertigation.ec_targets (
    id                  BIGSERIAL PRIMARY KEY,
    farm_id             BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    zone_id             BIGINT REFERENCES gr33ncore.zones(id) ON DELETE CASCADE,
    growth_stage        gr33nfertigation.growth_stage_enum NOT NULL,
    ec_min_mscm         NUMERIC(5,3) NOT NULL CHECK (ec_min_mscm >= 0),
    ec_max_mscm         NUMERIC(5,3) NOT NULL CHECK (ec_max_mscm >= 0),
    ph_min              NUMERIC(4,2) DEFAULT 5.8 CHECK (ph_min >= 0 AND ph_min <= 14),
    ph_max              NUMERIC(4,2) DEFAULT 6.8 CHECK (ph_max >= 0 AND ph_max <= 14),
    notes               TEXT,
    rationale           TEXT,
    created_at          TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at          TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT uq_ec_target_zone_stage UNIQUE (farm_id, zone_id, growth_stage),
    CONSTRAINT chk_ec_range CHECK (ec_min_mscm < ec_max_mscm),
    CONSTRAINT chk_ph_range CHECK (ph_min < ph_max)
);
CREATE TRIGGER trg_ec_targets_updated_at
    BEFORE UPDATE ON gr33nfertigation.ec_targets
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

CREATE TABLE IF NOT EXISTS gr33nfertigation.crop_cycles (
    id                          BIGSERIAL PRIMARY KEY,
    farm_id                     BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    zone_id                     BIGINT NOT NULL REFERENCES gr33ncore.zones(id) ON DELETE RESTRICT,
    name                        TEXT NOT NULL,
    strain_or_variety           TEXT,
    current_stage               gr33nfertigation.growth_stage_enum DEFAULT 'seedling',
    is_active                   BOOLEAN DEFAULT TRUE NOT NULL,
    started_at                  DATE NOT NULL,
    harvested_at                DATE,
    yield_grams                 NUMERIC(10,2) CHECK (yield_grams >= 0),
    yield_notes                 TEXT,
    cycle_notes                 TEXT,
    created_at                  TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at                  TIMESTAMPTZ DEFAULT NOW() NOT NULL
);
CREATE UNIQUE INDEX uq_active_crop_cycle
    ON gr33nfertigation.crop_cycles(zone_id) WHERE is_active = TRUE;
CREATE TRIGGER trg_crop_cycles_updated_at
    BEFORE UPDATE ON gr33nfertigation.crop_cycles
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

CREATE TABLE IF NOT EXISTS gr33nfertigation.programs (
    id                          BIGSERIAL PRIMARY KEY,
    farm_id                     BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    name                        TEXT NOT NULL,
    description                 TEXT,
    application_recipe_id       BIGINT REFERENCES gr33nnaturalfarming.application_recipes(id) ON DELETE SET NULL,
    reservoir_id                BIGINT REFERENCES gr33nfertigation.reservoirs(id) ON DELETE SET NULL,
    target_zone_id              BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
    schedule_id                 BIGINT REFERENCES gr33ncore.schedules(id) ON DELETE SET NULL,
    ec_target_id                BIGINT REFERENCES gr33nfertigation.ec_targets(id) ON DELETE SET NULL,
    volume_liters_per_sqm       NUMERIC(8,3) CHECK (volume_liters_per_sqm >= 0),
    total_volume_liters         NUMERIC(10,3) CHECK (total_volume_liters >= 0),
    dilution_ratio              TEXT,
    run_duration_seconds        INTEGER CHECK (run_duration_seconds >= 0),
    ec_trigger_low              NUMERIC(5,3) CHECK (ec_trigger_low >= 0),
    ph_trigger_low              NUMERIC(4,2) CHECK (ph_trigger_low >= 0 AND ph_trigger_low <= 14),
    ph_trigger_high             NUMERIC(4,2) CHECK (ph_trigger_high >= 0 AND ph_trigger_high <= 14),
    is_active                   BOOLEAN DEFAULT TRUE NOT NULL,
    metadata                    JSONB DEFAULT '{}',
    created_at                  TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at                  TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    deleted_at                  TIMESTAMPTZ DEFAULT NULL
);
CREATE TRIGGER trg_programs_updated_at
    BEFORE UPDATE ON gr33nfertigation.programs
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
ALTER TABLE gr33nfertigation.crop_cycles
    ADD COLUMN primary_program_id BIGINT REFERENCES gr33nfertigation.programs(id) ON DELETE SET NULL;

-- Phase 20.95 WS3 — FK on executable_actions.program_id (column created above;
-- FK deferred here because gr33nfertigation.programs is defined below core).
ALTER TABLE gr33ncore.executable_actions
    ADD CONSTRAINT fk_executable_actions_program
    FOREIGN KEY (program_id) REFERENCES gr33nfertigation.programs(id) ON DELETE CASCADE;
CREATE INDEX IF NOT EXISTS idx_executable_actions_program
    ON gr33ncore.executable_actions (program_id);

CREATE TABLE IF NOT EXISTS gr33nfertigation.mixing_events (
    id                      BIGSERIAL PRIMARY KEY,
    farm_id                 BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    reservoir_id            BIGINT NOT NULL REFERENCES gr33nfertigation.reservoirs(id) ON DELETE RESTRICT,
    program_id              BIGINT REFERENCES gr33nfertigation.programs(id) ON DELETE SET NULL,
    mixed_by_user_id        UUID REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    mixed_at                TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    water_volume_liters     NUMERIC(10,3) NOT NULL CHECK (water_volume_liters >= 0),
    water_source            TEXT,
    water_ec_mscm           NUMERIC(5,3) CHECK (water_ec_mscm >= 0),
    water_ph                NUMERIC(4,2) CHECK (water_ph >= 0 AND water_ph <= 14),
    final_ec_mscm           NUMERIC(5,3) CHECK (final_ec_mscm >= 0),
    final_ph                NUMERIC(4,2) CHECK (final_ph >= 0 AND final_ph <= 14),
    final_temp_celsius      NUMERIC(5,2),
    ec_target_id            BIGINT REFERENCES gr33nfertigation.ec_targets(id) ON DELETE SET NULL,
    ec_target_met           BOOLEAN,
    notes                   TEXT,
    observations            TEXT,
    created_at              TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE TABLE IF NOT EXISTS gr33nfertigation.mixing_event_components (
    id                      BIGSERIAL PRIMARY KEY,
    mixing_event_id         BIGINT NOT NULL REFERENCES gr33nfertigation.mixing_events(id) ON DELETE CASCADE,
    input_definition_id     BIGINT NOT NULL REFERENCES gr33nnaturalfarming.input_definitions(id) ON DELETE RESTRICT,
    input_batch_id          BIGINT REFERENCES gr33nnaturalfarming.input_batches(id) ON DELETE SET NULL,
    volume_added_ml         NUMERIC(10,3) NOT NULL CHECK (volume_added_ml > 0),
    dilution_ratio          TEXT,
    notes                   TEXT
);

CREATE TABLE IF NOT EXISTS gr33nfertigation.fertigation_events (
    id                          BIGSERIAL PRIMARY KEY,
    farm_id                     BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    program_id                  BIGINT REFERENCES gr33nfertigation.programs(id) ON DELETE SET NULL,
    reservoir_id                BIGINT REFERENCES gr33nfertigation.reservoirs(id) ON DELETE SET NULL,
    zone_id                     BIGINT NOT NULL REFERENCES gr33ncore.zones(id) ON DELETE RESTRICT,
    actuator_id                 BIGINT REFERENCES gr33ncore.actuators(id) ON DELETE SET NULL,
    mixing_event_id             BIGINT REFERENCES gr33nfertigation.mixing_events(id) ON DELETE SET NULL,
    crop_cycle_id               BIGINT REFERENCES gr33nfertigation.crop_cycles(id) ON DELETE SET NULL,
    applied_at                  TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    growth_stage                gr33nfertigation.growth_stage_enum,
    volume_applied_liters       NUMERIC(10,3) CHECK (volume_applied_liters >= 0),
    run_duration_seconds        INTEGER CHECK (run_duration_seconds >= 0),
    ec_before_mscm              NUMERIC(5,3) CHECK (ec_before_mscm >= 0),
    ec_after_mscm               NUMERIC(5,3) CHECK (ec_after_mscm >= 0),
    ph_before                   NUMERIC(4,2) CHECK (ph_before >= 0 AND ph_before <= 14),
    ph_after                    NUMERIC(4,2) CHECK (ph_after >= 0 AND ph_after <= 14),
    runoff_ec_mscm              NUMERIC(5,3) CHECK (runoff_ec_mscm >= 0),
    runoff_ph                   NUMERIC(4,2) CHECK (runoff_ph >= 0 AND runoff_ph <= 14),
    trigger_source              gr33nfertigation.program_trigger_enum DEFAULT 'manual',
    triggered_by_rule_id        BIGINT REFERENCES gr33ncore.automation_rules(id) ON DELETE SET NULL,
    triggered_by_schedule_id    BIGINT REFERENCES gr33ncore.schedules(id) ON DELETE SET NULL,
    triggered_by_user_id        UUID REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    plant_response              TEXT,
    notes                       TEXT,
    metadata                    JSONB DEFAULT '{}',
    created_at                  TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_reservoirs_farm
    ON gr33nfertigation.reservoirs(farm_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_ec_targets_zone_stage
    ON gr33nfertigation.ec_targets(zone_id, growth_stage);
CREATE INDEX IF NOT EXISTS idx_programs_farm_active
    ON gr33nfertigation.programs(farm_id, is_active) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_mixing_events_farm
    ON gr33nfertigation.mixing_events(farm_id, mixed_at DESC);
CREATE INDEX IF NOT EXISTS idx_mixing_events_res
    ON gr33nfertigation.mixing_events(reservoir_id, mixed_at DESC);
CREATE INDEX IF NOT EXISTS idx_fert_events_farm_time
    ON gr33nfertigation.fertigation_events(farm_id, applied_at DESC);
CREATE INDEX IF NOT EXISTS idx_fert_events_zone_time
    ON gr33nfertigation.fertigation_events(zone_id, applied_at DESC);
CREATE INDEX IF NOT EXISTS idx_fert_events_cycle
    ON gr33nfertigation.fertigation_events(crop_cycle_id, applied_at DESC);
CREATE INDEX IF NOT EXISTS idx_crop_cycles_zone
    ON gr33nfertigation.crop_cycles(zone_id, started_at DESC);

-- ============================================================
-- Phase 20.6 WS1 — stage-scoped setpoints (gr33ncore.zone_setpoints)
-- ============================================================
-- Placed here (not with the rest of gr33ncore early in the file) because
-- it references gr33nfertigation.crop_cycles, which isn't declared until
-- the fertigation block above. Strictly additive; no existing tables
-- change. `stage` is TEXT (not growth_stage_enum) so non-crop zones —
-- drying rooms, propagation areas, aquaponics loops — can carry
-- setpoints too. Resolution precedence at eval time:
-- cycle+stage > cycle-any-stage > zone+stage > zone-any-stage > nothing.
CREATE TABLE IF NOT EXISTS gr33ncore.zone_setpoints (
    id              BIGSERIAL PRIMARY KEY,
    farm_id         BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    zone_id         BIGINT REFERENCES gr33ncore.zones(id) ON DELETE CASCADE,
    crop_cycle_id   BIGINT REFERENCES gr33nfertigation.crop_cycles(id) ON DELETE CASCADE,
    stage           TEXT,
    sensor_type     TEXT NOT NULL,
    min_value       NUMERIC,
    max_value       NUMERIC,
    ideal_value     NUMERIC,
    meta            JSONB NOT NULL DEFAULT '{}',
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_setpoint_scope CHECK (zone_id IS NOT NULL OR crop_cycle_id IS NOT NULL),
    CONSTRAINT chk_setpoint_numeric_coherent CHECK (
        (min_value IS NULL OR max_value IS NULL OR min_value <= max_value) AND
        (ideal_value IS NULL OR min_value IS NULL OR ideal_value >= min_value) AND
        (ideal_value IS NULL OR max_value IS NULL OR ideal_value <= max_value)
    )
);
CREATE INDEX IF NOT EXISTS idx_zone_setpoints_zone_stage
    ON gr33ncore.zone_setpoints (zone_id, stage, sensor_type);
CREATE INDEX IF NOT EXISTS idx_zone_setpoints_cycle_stage
    ON gr33ncore.zone_setpoints (crop_cycle_id, stage, sensor_type);
CREATE INDEX IF NOT EXISTS idx_zone_setpoints_farm
    ON gr33ncore.zone_setpoints (farm_id);
DROP TRIGGER IF EXISTS trg_zone_setpoints_updated_at ON gr33ncore.zone_setpoints;
CREATE TRIGGER trg_zone_setpoints_updated_at
    BEFORE UPDATE ON gr33ncore.zone_setpoints
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- ============================================================
-- SCHEMAS: gr33ncrops, gr33nanimals, gr33naquaponics (Phase 14 WS7 stubs)
-- ============================================================
-- Enable per farm with gr33ncore.farm_active_modules.module_schema_name
-- matching the schema name. See docs/domain-modules-operator-playbook.md.

CREATE SCHEMA IF NOT EXISTS gr33ncrops;

CREATE TABLE IF NOT EXISTS gr33ncrops.plants (
    id                   BIGSERIAL PRIMARY KEY,
    farm_id              BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    display_name         TEXT NOT NULL,
    variety_or_cultivar  TEXT,
    meta                 JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at           TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_gr33ncrops_plants_farm
    ON gr33ncrops.plants (farm_id)
    WHERE deleted_at IS NULL;

CREATE TRIGGER trg_gr33ncrops_plants_updated_at
    BEFORE UPDATE ON gr33ncrops.plants
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

CREATE SCHEMA IF NOT EXISTS gr33nanimals;

CREATE TABLE IF NOT EXISTS gr33nanimals.animal_groups (
    id              BIGSERIAL PRIMARY KEY,
    farm_id         BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    label           TEXT NOT NULL,
    species         TEXT,
    meta            JSONB NOT NULL DEFAULT '{}'::jsonb,
    -- Phase 20.95 WS4 scope columns (nullable / safe defaults).
    count           INTEGER,
    primary_zone_id BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
    active          BOOLEAN NOT NULL DEFAULT TRUE,
    archived_at     TIMESTAMPTZ,
    archived_reason TEXT,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_gr33nanimals_groups_farm
    ON gr33nanimals.animal_groups (farm_id)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_gr33nanimals_groups_primary_zone
    ON gr33nanimals.animal_groups (primary_zone_id)
    WHERE deleted_at IS NULL;

CREATE TRIGGER trg_gr33nanimals_animal_groups_updated_at
    BEFORE UPDATE ON gr33nanimals.animal_groups
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

CREATE SCHEMA IF NOT EXISTS gr33naquaponics;

CREATE TABLE IF NOT EXISTS gr33naquaponics.loops (
    id                BIGSERIAL PRIMARY KEY,
    farm_id           BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    label             TEXT NOT NULL,
    meta              JSONB NOT NULL DEFAULT '{}'::jsonb,
    -- Phase 20.95 WS4 topology columns (nullable).
    fish_tank_zone_id BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
    grow_bed_zone_id  BIGINT REFERENCES gr33ncore.zones(id) ON DELETE SET NULL,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    deleted_at        TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_gr33naquaponics_loops_farm
    ON gr33naquaponics.loops (farm_id)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_gr33naquaponics_loops_fish_tank_zone
    ON gr33naquaponics.loops (fish_tank_zone_id)
    WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_gr33naquaponics_loops_grow_bed_zone
    ON gr33naquaponics.loops (grow_bed_zone_id)
    WHERE deleted_at IS NULL;

CREATE TRIGGER trg_gr33naquaponics_loops_updated_at
    BEFORE UPDATE ON gr33naquaponics.loops
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();

-- ============================================================
-- MIGRATION NOTES (read before running)
-- ============================================================
-- 1. auth schema bootstrap at the top is for LOCAL DEV ONLY.
--    Remove or skip it when deploying to Supabase hosted.
-- 2. PostGIS/TimescaleDB extensions: uncomment the CREATE
--    EXTENSION lines at the top once installed locally.
-- 3. TimescaleDB hypertables: uncomment the create_hypertable
--    calls at the bottom AFTER the base tables are created and
--    extensions are enabled.
-- 4. Supabase production: replace auth bootstrap with your
--    actual Supabase auth.users reference and enable RLS
--    policies before going live.
-- ============================================================
