-- ============================================================
-- Phase 20.9 WS3 — Program → executable_actions backfill
--
-- gr33ncore.executable_actions already has program_id (added in
-- Phase 20.95 WS3) and the chk_executable_source CHECK enforces
-- exactly-one of (schedule_id, rule_id, program_id). This
-- migration layers an idempotent backfill on top: for every row
-- in gr33nfertigation.programs whose metadata->'steps' is a
-- non-empty JSON array AND which has zero existing
-- executable_actions rows, emit one row per step.
--
-- Step schema (loose; rows that fail the CHECK are silently
-- skipped via the DO-block's EXCEPTION clause so a malformed step
-- never blocks the migration):
--
--   {
--     "action_type":     "control_actuator" | "create_task" |
--                        "send_notification" | "log_custom_event" |
--                        "http_webhook_call" | "update_record_in_gr33n" |
--                        "trigger_another_automation_rule",
--     "execution_order":           <int, default 0>,
--     "target_actuator_id":        <bigint or null>,
--     "target_automation_rule_id": <bigint or null>,
--     "target_notification_template_id": <bigint or null>,
--     "action_command":            <text or null>,
--     "action_parameters":         <jsonb or null>,
--     "delay_before_execution_seconds": <int, default 0>
--   }
--
-- Idempotency: the function short-circuits on programs that
-- already have at least one executable_actions row with
-- program_id = p_program_id. Run it more than once without fear.
-- ============================================================

CREATE OR REPLACE FUNCTION gr33ncore._backfill_program_actions(p_program_id BIGINT)
RETURNS INTEGER
LANGUAGE plpgsql
AS $$
DECLARE
    step          JSONB;
    inserted      INTEGER := 0;
    existing_cnt  BIGINT;
    steps         JSONB;
BEGIN
    SELECT COUNT(*) INTO existing_cnt
      FROM gr33ncore.executable_actions
     WHERE program_id = p_program_id;
    IF existing_cnt > 0 THEN
        RETURN 0;
    END IF;

    SELECT metadata -> 'steps' INTO steps
      FROM gr33nfertigation.programs
     WHERE id = p_program_id;

    IF steps IS NULL OR jsonb_typeof(steps) <> 'array' THEN
        RETURN 0;
    END IF;

    FOR step IN SELECT * FROM jsonb_array_elements(steps) LOOP
        -- Wrap each INSERT in its own EXCEPTION block so a malformed
        -- step doesn't abort the whole program's backfill.
        BEGIN
            INSERT INTO gr33ncore.executable_actions (
                program_id,
                execution_order,
                action_type,
                target_actuator_id,
                target_automation_rule_id,
                target_notification_template_id,
                action_command,
                action_parameters,
                delay_before_execution_seconds
            ) VALUES (
                p_program_id,
                COALESCE((step ->> 'execution_order')::INTEGER, 0),
                (step ->> 'action_type')::gr33ncore.executable_action_type_enum,
                NULLIF(step ->> 'target_actuator_id', '')::BIGINT,
                NULLIF(step ->> 'target_automation_rule_id', '')::BIGINT,
                NULLIF(step ->> 'target_notification_template_id', '')::BIGINT,
                NULLIF(step ->> 'action_command', ''),
                CASE WHEN step ? 'action_parameters' THEN step -> 'action_parameters' ELSE NULL END,
                COALESCE((step ->> 'delay_before_execution_seconds')::INTEGER, 0)
            );
            inserted := inserted + 1;
        EXCEPTION
            WHEN check_violation OR invalid_text_representation OR not_null_violation THEN
                RAISE NOTICE 'Skipped malformed step in program %: %', p_program_id, step;
        END;
    END LOOP;

    RETURN inserted;
END;
$$;

COMMENT ON FUNCTION gr33ncore._backfill_program_actions(BIGINT) IS
    'Phase 20.9 WS3 — idempotent backfill of programs.metadata->''steps'' into '
    'executable_actions rows. Returns the number of rows inserted. '
    'Re-running against an already-backfilled program returns 0.';

-- Run it across every existing program. Programs with no steps,
-- already-backfilled programs, or soft-deleted programs are
-- short-circuited inside the function.
DO $$
DECLARE
    p_id    BIGINT;
    total   INTEGER := 0;
    scanned INTEGER := 0;
BEGIN
    FOR p_id IN
        SELECT id FROM gr33nfertigation.programs WHERE deleted_at IS NULL
    LOOP
        scanned := scanned + 1;
        total := total + gr33ncore._backfill_program_actions(p_id);
    END LOOP;
    RAISE NOTICE 'Phase 20.9 WS3 backfill: scanned % programs, inserted % executable_actions rows.',
        scanned, total;
END $$;
