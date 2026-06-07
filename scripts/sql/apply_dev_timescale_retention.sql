-- Phase 48 WS5 — optional dev/staging Timescale retention.
-- Run only via scripts/apply-dev-retention.sh (requires TIMESCALE_RETENTION_DAYS).

\set ON_ERROR_STOP on

\if :{?retention_days}
\else
\echo 'error: pass -v retention_days=N'
\quit 1
\endif

SELECT add_retention_policy(
    'gr33ncore.sensor_readings',
    (:retention_days || ' days')::interval,
    if_not_exists => TRUE
);

SELECT add_retention_policy(
    'gr33ncore.actuator_events',
    (:retention_days || ' days')::interval,
    if_not_exists => TRUE
);

\echo 'ok  retention policies ensured for sensor_readings and actuator_events'
