-- Phase 111 — farm-default Guardian LLM model + audit enum value

ALTER TABLE gr33ncore.farms
  ADD COLUMN IF NOT EXISTS guardian_preferred_model TEXT NULL;

COMMENT ON COLUMN gr33ncore.farms.guardian_preferred_model IS
  'Farm-default Ollama model for Guardian chat. NULL = use server LLM_MODEL env.';

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_enum e
    JOIN pg_type t ON e.enumtypid = t.oid
    JOIN pg_namespace n ON n.oid = t.typnamespace
    WHERE n.nspname = 'gr33ncore'
      AND t.typname = 'user_action_type_enum'
      AND e.enumlabel = 'guardian_model_changed'
  ) THEN
    ALTER TYPE gr33ncore.user_action_type_enum ADD VALUE 'guardian_model_changed';
  END IF;
END $$;
