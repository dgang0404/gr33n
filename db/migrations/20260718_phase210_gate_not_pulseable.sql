-- Phase 210 — bugfix from phase183_gate_actuator_type.sql: a gate is an
-- open/shut toggle, not a timed-run device like a feeder hopper or water
-- valve. supports_pulse=true made the UI's "Run pulse" control appear for
-- gate actuators, but internal/handler/actuator.PulseDurationAllowed never
-- included "gate" in its switch, so submitting a duration always 400'd.
-- Correct the declarative flag to match the actual backend behaviour.
UPDATE gr33ncore.device_type_registry
SET supports_pulse = false
WHERE type_key = 'gate';
