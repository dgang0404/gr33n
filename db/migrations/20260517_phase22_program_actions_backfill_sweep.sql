-- ============================================================
-- Phase 22 WS2 — Program actions backfill sweep + per-program log
--
-- The 20260515 migration installed gr33ncore._backfill_program_actions
-- and ran it across every existing program once. Between that deploy
-- and this one, the API could accept new programs carrying
-- metadata.steps (via legacy clients that haven't adopted the
-- /fertigation/programs/{id}/actions endpoints added in Phase 20.9).
--
-- This migration is the final sweep:
--
--   1. Re-run the backfill across every non-deleted program. The
--      function's internal "already has executable_actions rows?"
--      guard makes this idempotent — already-migrated programs
--      return 0 inserts and incur a single COUNT.
--   2. Log a per-program NOTICE whenever the backfill inserted rows,
--      so the deploy log captures exactly which programs still needed
--      backfilling (and how many steps they contributed).
--   3. Emit a summary NOTICE at the end so ops can eyeball the numbers
--      without parsing individual rows.
--
-- After this migration, the worker's ResolveProgramActions fallback
-- warning should only fire for programs created *after* deploy by a
-- stubbornly legacy client. Those surface in worker logs and should be
-- chased down manually (the warning spells out the remediation URL).
-- ============================================================

DO $$
DECLARE
    p_id           BIGINT;
    p_name         TEXT;
    inserted_cnt   INTEGER;
    scanned_total  INTEGER := 0;
    migrated_total INTEGER := 0;
    programs_hit   INTEGER := 0;
BEGIN
    FOR p_id, p_name IN
        SELECT id, name
          FROM gr33nfertigation.programs
         WHERE deleted_at IS NULL
         ORDER BY id
    LOOP
        scanned_total := scanned_total + 1;
        inserted_cnt := gr33ncore._backfill_program_actions(p_id);
        IF inserted_cnt > 0 THEN
            programs_hit := programs_hit + 1;
            migrated_total := migrated_total + inserted_cnt;
            RAISE NOTICE 'Phase 22 WS2 backfill: program id=% name=% inserted % executable_actions rows',
                p_id, p_name, inserted_cnt;
        END IF;
    END LOOP;

    RAISE NOTICE 'Phase 22 WS2 backfill sweep complete: scanned % programs, migrated % programs, inserted % rows total.',
        scanned_total, programs_hit, migrated_total;
END $$;
