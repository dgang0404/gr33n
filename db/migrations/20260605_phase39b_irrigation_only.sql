-- Phase 39b — plain irrigation programs (RO/well): pulse only, no mix_batch.
ALTER TABLE gr33nfertigation.programs
    ADD COLUMN IF NOT EXISTS irrigation_only BOOLEAN NOT NULL DEFAULT FALSE;

COMMENT ON COLUMN gr33nfertigation.programs.irrigation_only IS
    'When true, program tick enqueues pulse irrigation only — no recipe, mix calculator, or mix_batch.';
