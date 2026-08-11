-- Phase 214 — device IP history for multi-device / enterprise topologies.
-- devices.ip_address only ever held the value set at CreateDevice time and
-- was never refreshed on heartbeat, so it silently went stale the moment a
-- device's network changed (the exact class of bug that made the Pi LED
-- rig unreachable — config pointed at a dead subnet with no DB trail to
-- explain why). This adds an append-only history of observed IP changes,
-- keyed by time so it scales the same way sensor_readings/actuator_events
-- already do when the DB lives on its own server.

CREATE TABLE IF NOT EXISTS gr33ncore.device_ip_events (
    device_id    BIGINT NOT NULL REFERENCES gr33ncore.devices(id) ON DELETE CASCADE,
    farm_id      BIGINT NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    old_ip       INET,
    new_ip       INET NOT NULL,
    observed_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    PRIMARY KEY (observed_at, device_id)
);

CREATE INDEX IF NOT EXISTS idx_device_ip_events_device
    ON gr33ncore.device_ip_events (device_id, observed_at DESC);

CREATE INDEX IF NOT EXISTS idx_device_ip_events_farm
    ON gr33ncore.device_ip_events (farm_id, observed_at DESC);

-- Hypertable conversion mirrors sensor_readings/actuator_events (Phase 25 WS4).
-- No-op locally when TimescaleDB isn't installed (matches the guarded block
-- in db/schema/gr33n-schema-v2-FINAL.sql).
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'timescaledb') THEN
    PERFORM create_hypertable('gr33ncore.device_ip_events', 'observed_at',
      if_not_exists => TRUE, chunk_time_interval => INTERVAL '30 days');
  END IF;
END $$;
