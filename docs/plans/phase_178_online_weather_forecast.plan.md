---
name: Phase 178 — Online weather forecast (Tier 3)
overview: >
  Ship the deferred Phase 66 Tier 3: optional online forecast via Open-Meteo
  (free, no key) with env-gated OpenWeather / VisualCrossing. Cache readings in
  weather_data, extend site-weather with connection status, and surface a clear
  "forecast connected / offline / disabled" indicator on Today and Settings —
  mirroring TopBar "● API online" honesty.
todos:
  - id: ws1-provider-config
    content: "WS1: WEATHER_PROVIDER env + API keys; expose on GET /capabilities"
    status: completed
  - id: ws2-openmeteo-client
    content: "WS2: Open-Meteo HTTP client — current + tonight low + cloud cover; unit tests with httptest"
    status: completed
  - id: ws3-cache-ingest
    content: "WS3: Fetch-on-read with TTL cache → weather_data; add api_openmeteo enum via migration"
    status: completed
  - id: ws4-site-weather-api
    content: "WS4: Extend GET /farms/{id}/site-weather with online_forecast block + status enum"
    status: completed
  - id: ws5-today-ui
    content: "WS5: FarmSiteStrip forecast row + connection badge (connected / cached / offline / disabled)"
    status: completed
  - id: ws6-settings-ui
    content: "WS6: Settings farm-site section — opt-in toggle, provider label, external-service notice"
    status: completed
  - id: ws7-guardian
    content: "WS7: site_weather read tool cites online tier; frost starter uses forecast when connected"
    status: completed
  - id: ws8-docs-tests
    content: "WS8: architecture §11 update; README env vars; phase-178-closure; smoke test"
    status: completed
isProject: false
---

# Phase 178 — Online weather forecast (Tier 3)

**Status:** shipped · **Follows:** [177](phase_177_today_first_impression.plan.md) · **Completes:** [66 Tier 3 deferral](phase_66_weather_site_context.plan.md)

## The gap

Phase 66 shipped Tiers 1–2 (offline solar math + manual/local readings). Tier 3 was
explicitly deferred:

> *Optional online provider opt-in, caches, degrades gracefully*

Today the schema has `api_openweather` and `api_visualcrossing` enum values, but there
is **no fetch job, no HTTP client, and no `WEATHER_PROVIDER` env wiring**. The UI says
"solar math — no internet required" but gives **no signal** when a live forecast is
available or why it isn't.

Operators need the same honesty as TopBar's `● API online` / `● API offline` — for
**weather forecast connectivity**, separate from LAN API health.

## North star

> Sun times always work offline. When the operator **opts in** and the farm has
> coordinates, gr33n shows outdoor temp, cloud cover, and tonight's low — with a
> badge that says exactly which tier answered: **live forecast**, **cached forecast**,
> **offline (solar only)**, or **disabled**.

Unplug the uplink → badge flips to cached-then-offline; sun dial unchanged.

## Three tiers (unchanged contract)

| Tier | Source | Internet? | Phase |
|------|--------|-----------|-------|
| **1 — Solar** | lat/long + date math | None | 66 ✓ |
| **2 — Local** | outdoor sensor / manual log | LAN only | 66 ✓ |
| **3 — Forecast** | Open-Meteo (default) or paid APIs | Optional WAN | **178** |

Tier 3 never blocks Tiers 1–2. Missing coords → forecast disabled with clear copy.

---

## WS1 — Provider config (`internal/weather/config.go`)

**Env vars** (API process only — never exposed to browser as secrets):

| Variable | Values | Default |
|----------|--------|---------|
| `WEATHER_PROVIDER` | `off`, `openmeteo`, `openweather`, `visualcrossing` | `off` |
| `OPENWEATHER_API_KEY` | string | — (required when provider=openweather) |
| `VISUALCROSSING_API_KEY` | string | — (required when provider=visualcrossing) |
| `WEATHER_CACHE_MINUTES` | int | `30` |
| `WEATHER_FETCH_TIMEOUT_SEC` | int | `8` |

**`GET /capabilities`** additions (public, no secrets):

```json
{
  "weather_forecast_available": true,
  "weather_provider": "openmeteo",
  "weather_provider_label": "Open-Meteo (free)"
}
```

`weather_forecast_available` is `true` only when `WEATHER_PROVIDER` is set and
required keys (if any) are present. UI uses this to show/hide Settings opt-in.

**Farm-level opt-in** — store in `farms.meta_data`:

```json
{ "weather_forecast_enabled": true }
```

Default `false`. Even when the API provider is configured, each farm must opt in
(Phase 66: "contacts an external service" notice). PATCH via existing farm site
endpoint or new `PATCH /farms/{id}/site` field.

---

## WS2 — Open-Meteo client (`internal/weather/openmeteo.go`)

Primary provider — **free, no API key**, fits LAN-first farms that occasionally
have uplink.

**Request** (current + daily min for frost):

```
GET https://api.open-meteo.com/v1/forecast
  ?latitude={lat}&longitude={lng}
  &current=temperature_2m,relative_humidity_2m,cloud_cover,precipitation,wind_speed_10m
  &daily=temperature_2m_min,temperature_2m_max,precipitation_sum
  &timezone=auto
  &forecast_days=2
```

**Mapped fields** → `weather_data` columns + `forecast_data` JSONB (raw daily slice).

**Tests:** `httptest` mock server; assert parse + error handling (timeout, 429, malformed JSON).

**Paid providers (stretch in 178, not blocking):**

- `openweather` → One Call 3.0 or current weather; `data_source = api_openweather`
- `visualcrossing` → Timeline API; `data_source = api_visualcrossing`

Ship Open-Meteo first; paid adapters share a small `Provider` interface.

---

## WS3 — Cache + ingest

**Strategy:** fetch-on-read inside `GET /farms/{id}/site-weather` (no background cron
in v1 — keeps Pi/simple deploys predictable). Same request path the UI already polls.

```
if provider enabled AND farm.weather_forecast_enabled AND coords set:
    latest = GetLatestWeatherForFarm where data_source LIKE 'api_%'
    if latest.recorded_at < now - cacheTTL:
        try fetch provider → InsertWeatherData
    else use cached row
```

**Migration** `db/migrations/20260712_phase178_weather_openmeteo.sql`:

```sql
ALTER TYPE gr33ncore.weather_data_source_enum ADD VALUE IF NOT EXISTS 'api_openmeteo';
```

Update `gr33n-schema-v2-FINAL.sql` enum for greenfield installs.

**Insert** via existing `InsertWeatherData` query; set `forecast_data` to provider
daily payload; `raw_data` to current conditions snapshot.

On fetch failure: return last-good row if any; set status `cached_stale`. No row →
status `offline` (solar still returned).

---

## WS4 — `site-weather` API shape

Extend `buildSiteWeatherResponse` in [`internal/handler/weather/handler.go`](../internal/handler/weather/handler.go).

**New top-level key** `online_forecast`:

```json
{
  "online_forecast": {
    "status": "connected",
    "provider": "openmeteo",
    "provider_label": "Open-Meteo",
    "enabled": true,
    "opted_in": true,
    "fetched_at": "2026-07-12T18:30:00Z",
    "stale": false,
    "message": "Live forecast",
    "current": {
      "temperature_celsius": 24.2,
      "humidity_percent": 58,
      "cloud_cover_percent": 35,
      "wind_speed_ms": 2.1
    },
    "tonight_low_celsius": 11.8,
    "frost_risk": false
  }
}
```

**`status` enum** (stable — UI maps to badge copy):

| status | Meaning | Badge copy |
|--------|---------|------------|
| `disabled` | `WEATHER_PROVIDER=off` or farm not opted in | `● Forecast off` |
| `no_coords` | Farm missing lat/long | `● Set location for forecast` |
| `connected` | Fresh fetch succeeded | `● Forecast live` |
| `cached` | Using row within TTL, no new fetch needed | `● Forecast cached` |
| `cached_stale` | Fetch failed; serving last-good | `● Forecast cached (offline)` |
| `offline` | No cache, fetch failed | `● Forecast offline` |
| `misconfigured` | Provider set but API key missing | `● Forecast misconfigured` |

Always include `tiers` array update: append `online_forecast` when status is
`connected`, `cached`, or `cached_stale`.

**Timeout:** weather fetch uses `WEATHER_FETCH_TIMEOUT_SEC`; must not block solar
response > ~10s total — run fetch in goroutine with deadline or skip if over budget.

---

## WS5 — Today UI (`FarmSiteStrip.vue`)

Add a **Forecast** cell beside the sun dial (same strip, no new row — matches 176
"enrich Site Strip in place" rule).

```
┌ Sun dial ──────┬ Outdoor ──┬ Water ──┬ Forecast ─────────────┐
│ ↑ 6:12 ↓ 20:45 │ 2 sensors │ …       │ 24°C · 35% clouds     │
│ 14.5h daylight │           │         │ ● Forecast live       │
└────────────────┴───────────┴─────────┴───────────────────────┘
```

**Component:** `ui/src/lib/siteWeatherForecast.js`

- `forecastStatusLabel(status)` → human badge text
- `forecastStatusTone(status)` → `gr33n-400` / `amber` / `zinc` (mirror TopBar)
- `formatForecastCurrent(online_forecast)` → `"24°C · 35% clouds"` or em dash

**`data-test` hooks:**

- `farm-site-forecast`
- `farm-site-forecast-status`
- `farm-site-forecast-reading`

When `status === disabled` and capabilities show provider available → subtle
"Enable in Settings" link.

Refresh: reuse existing `fetchSiteWeather` on Dashboard mount + after coords save.

---

## WS6 — Settings UI (`Settings.vue` farm site section)

Below coordinates form, new **Online forecast** card:

- Toggle: **Use live weather forecast** (`weather_forecast_enabled` in meta_data)
- When provider unavailable at API level → toggle disabled + copy: *"Set `WEATHER_PROVIDER=openmeteo` on the API server to enable."*
- When enabled → amber info box: *"This contacts an external weather service when you open Today. Sun times still work offline."*
- Status line mirroring Today badge after save (optional test fetch via site-weather)

**`data-test`:** `settings-weather-forecast-toggle`, `settings-weather-forecast-status`

---

## WS7 — Guardian

Update [`internal/farmguardian/readtools_weather.go`](../internal/farmguardian/readtools_weather.go):

- When `online_forecast` tier present, append tonight low + frost heuristic to read tool output
- Frost risk: `tonight_low_celsius < 2` (configurable constant) with forecast tier cited
- Persona line: *"Forecast from Open-Meteo at HH:MM; cached if offline"*

No new read tool — extend existing `site_weather` rendering.

Weather starters in Ask drawer (175 demotion) unchanged; answers improve when Tier 3 on.

---

## WS8 — Docs, tests, closure

| Artifact | Content |
|----------|---------|
| `docs/farm-guardian-architecture.md` §11 | Tier 3 row + status enum + env table |
| `README.md` | `WEATHER_PROVIDER`, opt-in, Open-Meteo default |
| `internal/weather/openmeteo_test.go` | Mock HTTP parse |
| `internal/handler/weather/handler_test.go` | Status transitions disabled → connected → cached_stale |
| `ui/src/__tests__/site-weather-forecast.test.js` | Badge labels + FarmSiteStrip render |
| `ui/src/__tests__/phase-178-closure.test.js` | Routes, capabilities keys, Settings toggle |
| `cmd/api/smoke_phase178_test.go` | Route + capabilities smoke |

**OC-178:** close Phase 66 definition-of-done bullet for Tier 3.

---

## Definition of done

- [x] `WEATHER_PROVIDER=openmeteo` + farm opt-in → live temp/cloud on Today with `● Forecast live`
- [x] WAN unplugged → `● Forecast cached (offline)` or `● Forecast offline`; sun dial unchanged
- [x] `WEATHER_PROVIDER=off` or farm toggle off → `● Forecast off`; no outbound HTTP
- [x] Guardian `site_weather` cites forecast tier when present
- [x] No API keys in browser; capabilities exposes only non-secret flags
- [x] phase-178-closure tests green

---

## Boundary

- **Not** a full weather app — no 10-day charts, radar, or push alerts in 178
- **Not** replacing Tier 1 solar — forecast is additive and optional
- **Not** farm-level provider selection in v1 — one provider per API deployment (env)
- Background sync cron deferred to a follow-up if fetch-on-read proves too slow at scale

---

## File map (expected touch)

| Area | Paths |
|------|-------|
| Config | `internal/weather/config.go` |
| Providers | `internal/weather/openmeteo.go`, `provider.go` |
| Handler | `internal/handler/weather/handler.go`, `forecast.go` |
| Migration | `db/migrations/20260712_phase178_weather_openmeteo.sql` |
| Routes / capabilities | `cmd/api/routes.go` |
| Farm opt-in | `internal/handler/farm/site.go`, `db/queries/farms.sql` |
| UI lib | `ui/src/lib/siteWeather.js`, `siteWeatherForecast.js` |
| UI components | `ui/src/components/FarmSiteStrip.vue`, `ui/src/views/Settings.vue` |
| Store | `ui/src/stores/capabilities.js` |
| Guardian | `internal/farmguardian/readtools_weather.go` |

---

## Suggested execution order

1. WS1 + WS3 migration (config + enum)
2. WS2 Open-Meteo client + tests
3. WS4 site-weather response
4. WS5 Today badge (visible win)
5. WS6 Settings opt-in
6. WS7 Guardian + WS8 docs/closure

Estimated size: **medium** — ~1–2 sessions; Open-Meteo only keeps scope tight; paid
providers can land as WS2b without blocking UI.
