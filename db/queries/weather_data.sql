-- ============================================================
-- Queries: gr33ncore.weather_data (Phase 66)
-- ============================================================

-- name: InsertWeatherData :one
INSERT INTO gr33ncore.weather_data (
    farm_id, zone_id, recorded_at, data_source, source_sensor_id,
    temperature_celsius, humidity_percent, precipitation_mm,
    wind_speed_ms, wind_direction_degrees, barometric_pressure_hpa,
    solar_radiation_wm2, dew_point_celsius, uv_index, cloud_cover_percent,
    forecast_data, raw_data
) VALUES (
    $1, $2, $3, $4, $5,
    $6, $7, $8,
    $9, $10, $11,
    $12, $13, $14, $15,
    $16, $17
)
RETURNING *;

-- name: GetLatestWeatherForFarm :one
SELECT id, farm_id, zone_id, recorded_at, data_source, source_sensor_id,
       temperature_celsius, humidity_percent, precipitation_mm,
       wind_speed_ms, wind_direction_degrees, barometric_pressure_hpa,
       solar_radiation_wm2, dew_point_celsius, uv_index, cloud_cover_percent,
       forecast_data, raw_data, created_at
FROM gr33ncore.weather_data
WHERE farm_id = $1
ORDER BY recorded_at DESC
LIMIT 1;

-- name: GetLatestAPIWeatherForFarm :one
SELECT id, farm_id, zone_id, recorded_at, data_source, source_sensor_id,
       temperature_celsius, humidity_percent, precipitation_mm,
       wind_speed_ms, wind_direction_degrees, barometric_pressure_hpa,
       solar_radiation_wm2, dew_point_celsius, uv_index, cloud_cover_percent,
       forecast_data, raw_data, created_at
FROM gr33ncore.weather_data
WHERE farm_id = $1
  AND data_source IN ('api_openmeteo', 'api_openweather', 'api_visualcrossing')
ORDER BY recorded_at DESC
LIMIT 1;
