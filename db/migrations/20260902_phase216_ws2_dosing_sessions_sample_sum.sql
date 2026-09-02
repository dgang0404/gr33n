-- Phase 216 WS2 — accumulate the sampling-cycle EC sum so the app layer can
-- compute a rolling average without depending on an untyped SQL AVG() scan
-- (Postgres NUMERIC through pgx comes back as interface{} without a
-- registered type — a plain running-sum column keeps this fully typed).
ALTER TABLE gr33nfertigation.dosing_sessions
    ADD COLUMN IF NOT EXISTS samples_sum_ec_mscm NUMERIC(12,3) NOT NULL DEFAULT 0;
