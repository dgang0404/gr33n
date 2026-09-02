-- ============================================================
-- Queries: gr33nfertigation.dosing_sessions (Phase 216 WS1/WS2)
-- ============================================================

-- name: CreateDosingSession :one
INSERT INTO gr33nfertigation.dosing_sessions (
    mixing_event_id, farm_id, reservoir_id, circulation_actuator_id,
    dose_increment_ml, max_total_dose_ml, circulate_seconds, settle_seconds,
    samples_per_cycle, max_no_improvement_cycles, next_action_at
) VALUES (
    $1, $2, $3, $4,
    $5, $6, $7, $8,
    $9, $10, $11
)
RETURNING *;

-- name: GetDosingSessionByID :one
SELECT * FROM gr33nfertigation.dosing_sessions
WHERE id = $1;

-- name: GetDosingSessionByMixingEventID :one
SELECT * FROM gr33nfertigation.dosing_sessions
WHERE mixing_event_id = $1;

-- name: ListDueDosingSessions :many
-- Worker tick: non-terminal sessions whose next_action_at has arrived.
SELECT * FROM gr33nfertigation.dosing_sessions
WHERE status NOT IN ('done', 'aborted')
  AND next_action_at IS NOT NULL
  AND next_action_at <= NOW()
ORDER BY next_action_at ASC;

-- name: ListActiveDosingSessionsByFarm :many
SELECT * FROM gr33nfertigation.dosing_sessions
WHERE farm_id = $1 AND status NOT IN ('done', 'aborted')
ORDER BY started_at DESC;

-- name: AdvanceDosingSessionState :one
-- Generic state transition: new status + next wake time, no other fields touched.
UPDATE gr33nfertigation.dosing_sessions
SET status = $2, next_action_at = $3
WHERE id = $1
RETURNING *;

-- name: RecordDosingSessionSample :one
-- Sampling state: bump the collected-this-cycle counter, stash the raw
-- reading, and add it to the running sum the evaluate step averages from.
UPDATE gr33nfertigation.dosing_sessions
SET samples_collected_this_cycle = samples_collected_this_cycle + 1,
    samples_sum_ec_mscm = samples_sum_ec_mscm + $2,
    last_sample_ec_mscm = $2
WHERE id = $1
RETURNING *;

-- name: EvaluateDosingSessionCycle :one
-- Evaluating state: record this cycle's average, reset the per-cycle sample
-- counter/sum, and update the fouled-sensor counter (no_improvement_cycles is
-- computed by the caller since it depends on the previous average).
UPDATE gr33nfertigation.dosing_sessions
SET last_sample_avg_ec_mscm = $2,
    samples_collected_this_cycle = 0,
    samples_sum_ec_mscm = 0,
    no_improvement_cycles = $3
WHERE id = $1
RETURNING *;

-- name: RecordDosingSessionDose :one
-- Dosing state: add to the running total after a pump-fire is queued.
UPDATE gr33nfertigation.dosing_sessions
SET total_dosed_ml = total_dosed_ml + $2,
    status = $3,
    next_action_at = $4
WHERE id = $1
RETURNING *;

-- name: CompleteDosingSession :one
UPDATE gr33nfertigation.dosing_sessions
SET status = 'done', next_action_at = NULL, completed_at = NOW()
WHERE id = $1
RETURNING *;

-- name: AbortDosingSession :one
UPDATE gr33nfertigation.dosing_sessions
SET status = 'aborted', next_action_at = NULL, completed_at = NOW(), aborted_reason = $2
WHERE id = $1
RETURNING *;
