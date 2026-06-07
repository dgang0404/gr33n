-- Truncate time-series rows for one farm's sensors (Phase 48 dev-reset --include-readings).
\set ON_ERROR_STOP on

\if :{?farm_id}
\else
\echo 'error: pass -v farm_id=N'
\quit 1
\endif

DELETE FROM gr33ncore.sensor_readings sr
USING gr33ncore.sensors s
WHERE sr.sensor_id = s.id
  AND s.farm_id = :farm_id;

DELETE FROM gr33ncore.actuator_events ae
USING gr33ncore.actuators a
WHERE ae.actuator_id = a.id
  AND a.farm_id = :farm_id;

\echo 'ok  truncated readings/events for farm ' :farm_id
