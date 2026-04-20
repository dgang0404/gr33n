-- ============================================================
-- Queries: schedules, executable actions, automation runs
-- ============================================================

-- name: ListSchedulesByFarm :many
SELECT * FROM gr33ncore.schedules
WHERE farm_id = $1
ORDER BY name ASC;

-- name: GetScheduleByID :one
SELECT * FROM gr33ncore.schedules
WHERE id = $1;

-- name: UpdateScheduleActive :one
UPDATE gr33ncore.schedules
SET is_active = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListActiveSchedules :many
SELECT * FROM gr33ncore.schedules
WHERE is_active = TRUE;

-- name: MarkScheduleTriggered :one
UPDATE gr33ncore.schedules
SET last_triggered_time = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: ListExecutableActionsBySchedule :many
SELECT * FROM gr33ncore.executable_actions
WHERE schedule_id = $1
ORDER BY execution_order ASC, id ASC;

-- name: CreateAutomationRun :one
INSERT INTO gr33ncore.automation_runs (
    farm_id, schedule_id, rule_id, program_id,
    status, message, details, executed_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: ListAutomationRunsByFarm :many
SELECT * FROM gr33ncore.automation_runs
WHERE farm_id = $1
ORDER BY executed_at DESC
LIMIT $2;

-- name: ListAutomationRunsByFarmAfterID :many
SELECT * FROM gr33ncore.automation_runs
WHERE farm_id = $1 AND id > $2
ORDER BY id ASC
LIMIT $3;

-- name: GetLastSuccessfulRunBySchedule :one
SELECT * FROM gr33ncore.automation_runs
WHERE schedule_id = $1 AND status = 'success'
ORDER BY executed_at DESC
LIMIT 1;

-- name: GetAutomationRunByDetails :one
SELECT * FROM gr33ncore.automation_runs
WHERE schedule_id = $1 AND details @> $2::jsonb
LIMIT 1;

-- name: GetAutomationRunByProgramAndDetails :one
SELECT * FROM gr33ncore.automation_runs
WHERE program_id = $1 AND details @> $2::jsonb
LIMIT 1;

-- name: CreateSchedule :one
INSERT INTO gr33ncore.schedules (
    farm_id, name, description, schedule_type, cron_expression,
    timezone, is_active, meta_data, preconditions
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateSchedule :one
UPDATE gr33ncore.schedules
SET name = $2, description = $3, schedule_type = $4,
    cron_expression = $5, timezone = $6, is_active = $7,
    meta_data = $8, preconditions = $9, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteSchedule :exec
DELETE FROM gr33ncore.schedules WHERE id = $1;

-- ============================================================
-- Queries: automation_rules (Phase 20 WS1)
-- ============================================================

-- name: ListAutomationRulesByFarm :many
SELECT * FROM gr33ncore.automation_rules
WHERE farm_id = $1
ORDER BY name ASC;

-- name: GetAutomationRuleByID :one
SELECT * FROM gr33ncore.automation_rules
WHERE id = $1;

-- name: ListActiveAutomationRules :many
SELECT * FROM gr33ncore.automation_rules
WHERE is_active = TRUE;

-- name: CreateAutomationRule :one
INSERT INTO gr33ncore.automation_rules (
    farm_id, name, description, is_active,
    trigger_source, trigger_configuration,
    condition_logic, conditions_jsonb,
    cooldown_period_seconds
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateAutomationRule :one
UPDATE gr33ncore.automation_rules
SET name = $2, description = $3, is_active = $4,
    trigger_source = $5, trigger_configuration = $6,
    condition_logic = $7, conditions_jsonb = $8,
    cooldown_period_seconds = $9, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdateAutomationRuleActive :one
UPDATE gr33ncore.automation_rules
SET is_active = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkAutomationRuleEvaluated :one
UPDATE gr33ncore.automation_rules
SET last_evaluated_time = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: MarkAutomationRuleTriggered :one
UPDATE gr33ncore.automation_rules
SET last_triggered_time = $2, last_evaluated_time = $2, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteAutomationRule :exec
DELETE FROM gr33ncore.automation_rules WHERE id = $1;

-- ============================================================
-- Queries: executable_actions bound to rules (Phase 20 WS1)
-- ============================================================

-- name: ListExecutableActionsByRule :many
SELECT * FROM gr33ncore.executable_actions
WHERE rule_id = $1
ORDER BY execution_order ASC, id ASC;

-- name: GetExecutableActionByID :one
SELECT * FROM gr33ncore.executable_actions
WHERE id = $1;

-- name: CreateExecutableActionForRule :one
INSERT INTO gr33ncore.executable_actions (
    rule_id, execution_order, action_type,
    target_actuator_id, target_automation_rule_id, target_notification_template_id,
    action_command, action_parameters, delay_before_execution_seconds
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: UpdateExecutableAction :one
UPDATE gr33ncore.executable_actions
SET execution_order = $2, action_type = $3,
    target_actuator_id = $4, target_automation_rule_id = $5,
    target_notification_template_id = $6,
    action_command = $7, action_parameters = $8,
    delay_before_execution_seconds = $9
WHERE id = $1
RETURNING *;

-- name: DeleteExecutableAction :exec
DELETE FROM gr33ncore.executable_actions WHERE id = $1;

-- ============================================================
-- Queries: executable_actions bound to fertigation programs (Phase 20.95 WS3)
-- Phase 20.7 WS3 will wire these into the program editor UI; for now we expose
-- the CRUD surface so the DB round-trip is covered.
-- ============================================================

-- name: ListExecutableActionsByProgram :many
SELECT * FROM gr33ncore.executable_actions
WHERE program_id = $1
ORDER BY execution_order ASC, id ASC;

-- name: CreateExecutableActionForProgram :one
INSERT INTO gr33ncore.executable_actions (
    program_id, execution_order, action_type,
    target_actuator_id, target_automation_rule_id, target_notification_template_id,
    action_command, action_parameters, delay_before_execution_seconds
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;
