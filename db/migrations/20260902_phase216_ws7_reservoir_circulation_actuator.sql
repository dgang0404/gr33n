-- Phase 216 WS7 — reservoirs.circulation_actuator_id: this reservoir's
-- circulation_pump actuator, used both as the WS2 dosing-session default and
-- the WS7 idle-maintenance schedule's target. Separate from
-- delivery_actuator_id (the dosing pump).
ALTER TABLE gr33nfertigation.reservoirs
    ADD COLUMN IF NOT EXISTS circulation_actuator_id BIGINT
    REFERENCES gr33ncore.actuators(id) ON DELETE SET NULL;
