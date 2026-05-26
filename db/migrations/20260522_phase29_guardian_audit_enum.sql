-- Phase 29 WS5 — audit action type for confirmed Guardian tool executions.

DO $$ BEGIN
    ALTER TYPE gr33ncore.user_action_type_enum ADD VALUE 'guardian_tool_executed';
EXCEPTION
    WHEN duplicate_object THEN NULL;
END $$;
