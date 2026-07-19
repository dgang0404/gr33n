-- Phase 210 — dedicated animal automation. Adds 'animal_lifecycle_event' as
-- a valid automation_rules.trigger_source so a rule can react to the most
-- recent gr33nanimals.animal_lifecycle_events row for a flock/group (e.g.
-- "flock released to pasture" -> open the pasture gate actuator). The actual
-- condition lives in conditions_jsonb as a new `animal_event` predicate type
-- (internal/automation/predicates.go); this migration only widens the enum
-- so existing rows/validators accept the new value.
ALTER TYPE gr33ncore.automation_trigger_source_enum ADD VALUE IF NOT EXISTS 'animal_lifecycle_event';
