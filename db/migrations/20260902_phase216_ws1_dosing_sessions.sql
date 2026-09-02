-- Phase 216 WS1 — dosing session state machine (1:1 with a mixing_event).
-- Tracks the in-progress "dose → circulate → settle → sample → evaluate" loop;
-- the parent mixing_event stays the append-once finished record.

CREATE TYPE gr33nfertigation.dosing_session_status_enum AS ENUM (
    'idle', 'circulating', 'settling', 'sampling', 'evaluating', 'dosing', 'done', 'aborted'
);

CREATE TABLE IF NOT EXISTS gr33nfertigation.dosing_sessions (
    id                            BIGSERIAL PRIMARY KEY,
    mixing_event_id               BIGINT NOT NULL UNIQUE REFERENCES gr33nfertigation.mixing_events(id) ON DELETE CASCADE,
    farm_id                       BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    reservoir_id                  BIGINT NOT NULL REFERENCES gr33nfertigation.reservoirs(id) ON DELETE CASCADE,
    circulation_actuator_id       BIGINT REFERENCES gr33ncore.actuators(id) ON DELETE SET NULL,
    status                        gr33nfertigation.dosing_session_status_enum NOT NULL DEFAULT 'idle',
    next_action_at                TIMESTAMPTZ,
    dose_increment_ml             NUMERIC(8,3) NOT NULL CHECK (dose_increment_ml > 0),
    max_total_dose_ml             NUMERIC(10,3) NOT NULL CHECK (max_total_dose_ml > 0),
    total_dosed_ml                NUMERIC(10,3) NOT NULL DEFAULT 0 CHECK (total_dosed_ml >= 0),
    circulate_seconds             INTEGER NOT NULL DEFAULT 30 CHECK (circulate_seconds >= 0),
    settle_seconds                INTEGER NOT NULL DEFAULT 30 CHECK (settle_seconds >= 0),
    samples_per_cycle             INTEGER NOT NULL DEFAULT 3 CHECK (samples_per_cycle >= 1),
    samples_collected_this_cycle  INTEGER NOT NULL DEFAULT 0 CHECK (samples_collected_this_cycle >= 0),
    last_sample_ec_mscm           NUMERIC(5,3) CHECK (last_sample_ec_mscm >= 0),
    last_sample_avg_ec_mscm       NUMERIC(5,3) CHECK (last_sample_avg_ec_mscm >= 0),
    no_improvement_cycles         INTEGER NOT NULL DEFAULT 0 CHECK (no_improvement_cycles >= 0),
    max_no_improvement_cycles     INTEGER NOT NULL DEFAULT 3 CHECK (max_no_improvement_cycles >= 1),
    aborted_reason                TEXT,
    started_at                    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at                  TIMESTAMPTZ,
    metadata                      JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at                    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at                    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE TRIGGER trg_dosing_sessions_updated_at
    BEFORE UPDATE ON gr33nfertigation.dosing_sessions
    FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
-- Worker tick scans for due, non-terminal sessions.
CREATE INDEX IF NOT EXISTS idx_dosing_sessions_due
    ON gr33nfertigation.dosing_sessions (next_action_at)
    WHERE status NOT IN ('done', 'aborted');
CREATE INDEX IF NOT EXISTS idx_dosing_sessions_reservoir
    ON gr33nfertigation.dosing_sessions (reservoir_id);
