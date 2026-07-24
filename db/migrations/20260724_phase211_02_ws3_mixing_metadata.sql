-- Phase 211.02 WS3 — store formula revision snapshot on mixing events (query via metadata JSONB)

ALTER TABLE gr33nfertigation.mixing_events
    ADD COLUMN IF NOT EXISTS metadata JSONB NOT NULL DEFAULT '{}'::jsonb;
