---
name: Phase 66 — Weather & site context (offline-first)
overview: >
  Bring outdoor reality into Guardian's reasoning — WITHOUT requiring internet.
  The schema (weather_data table) and farm coordinates (location_gis) already exist.
  Tier 1 is a pure-math offline solar engine (sunrise/sunset/daylength/natural DLI
  from lat-long + date, zero network). Tier 2 is a local outdoor sensor. Tier 3 is
  an OPTIONAL online forecast that degrades gracefully when LAN-only.
todos:
  - id: ws1-site-coords
    content: "WS1: Site coordinates UI — confirm/set lat-long + elevation on farm (uses existing location_gis)"
    status: completed
  - id: ws2-solar-engine
    content: "WS2: Pure-Go offline solar engine — sunrise, sunset, daylength, solar noon, clear-sky DLI"
    status: completed
  - id: ws3-ingestion-sources
    content: "WS3: Manual entry UI; outdoor sensor mapping; OPTIONAL online provider behind flag w/ cache"
    status: completed
  - id: ws4-effects
    content: "WS4: Supplemental light DLI top-up; greenhouse vent/shade/heat nudges; frost/heat alerts"
    status: completed
  - id: ws5-guardian
    content: "WS5: weather read tool; starters 'Need supplemental light today?', 'Frost risk tonight?'"
    status: completed
  - id: ws6-docs-tests
    content: "WS6: architecture §11; offline-first note; phase-66-closure; OC-66"
    status: completed
isProject: false
---

# Phase 66 — Weather & site context (offline-first)

## Status

**Shipped.** Offline solar engine, site coords in Settings, `site_weather` read tool, dashboard daylight chip + weather starters, manual weather API. Online forecast provider deferred (Tier 3 flag) → **[Phase 178](phase_178_online_weather_forecast.plan.md)**.

---

## The question this answers

> *"How will weather play in? NASA API needs internet… could it work off GPS?"*

**Great instinct, and the answer is even better than GPS:** you don't need GPS or internet for the most useful weather signal. The farm already stores **`location_gis` (lat/long)** and **`timezone`**, and the **`weather_data` table already exists** with a multi-source enum. Sun position is **pure math** — given coordinates + date, sunrise / sunset / daylength / solar noon / clear-sky solar radiation are computed offline with zero network calls.

### Three tiers (each works without the one above it)

| Tier | Source | Internet? | Gives you |
|------|--------|-----------|-----------|
| **1 — Solar (always on)** | lat/long + date math | **None** | Sunrise, sunset, daylength, solar noon, theoretical clear-sky DLI |
| **2 — Local sensor** | Outdoor BME280 / weather station on the Pi | **None (LAN)** | Real outdoor temp, RH, pressure, actual solar (if pyranometer) |
| **3 — Forecast (optional)** | Open-Meteo (free, no key) / OpenWeather / VisualCrossing | **Yes, optional** | Cloud cover, precip, multi-day forecast, frost outlook |

The platform's offline-first ethos (Phase 51 config cache) carries straight over: Tier 3 caches last-good and degrades to Tier 1+2 when the LAN has no uplink.

---

## WS1 — Site coordinates

- Farm setup: confirm/set **lat/long** (map pin or manual) + **elevation** (improves solar)
- Reuse existing `location_gis GEOMETRY(Point,4326)`; add elevation to `meta_data`
- Plain copy: "Where is your farm? This lets gr33n calculate daylight hours — no internet needed."

---

## WS2 — Offline solar engine (the headline)

Pure-Go implementation (NOAA solar position algorithm — arithmetic, **no network, no dependency**):

```go
type SolarDay struct {
    Sunrise, Sunset, SolarNoon time.Time
    DaylengthHours             float64
    ClearSkyDLI                float64 // mol/m²/day, theoretical max
    MaxSunElevationDeg         float64
}
func SolarForDate(lat, lng float64, tz *time.Location, date time.Time) SolarDay
```

Powers, with no internet:
- **Daylength / photoperiod** vs the crop profile's target (Phase 64) — "natural day is 13.5h, your strain wants 12h to flower → you need blackout"
- **Natural clear-sky DLI** → supplemental lighting top-up math (WS4)
- **Frost-hour awareness** (longest night / lowest sun → cold risk windows)

---

## WS3 — Ingestion sources

| Source | data_source enum | UX |
|--------|------------------|-----|
| Manual | `manual_entry` | "Log today's weather" quick form |
| Outdoor sensor | `iot_sensor_reading` / `farm_weather_station` | Map a sensor as "outdoor"; auto-ingest via Pi |
| Online (optional) | `api_openweather` / `api_visualcrossing` (+ add Open-Meteo) | Settings flag `WEATHER_PROVIDER`; cache last-good; never blocks |

Online tier is **off by default** — opt-in, with a clear "this contacts an external service" notice that fits the LAN-first product promise.

---

## WS4 — Effects (where it gets useful)

| Effect | Logic | Guardian / UI surface |
|--------|-------|----------------------|
| **Supplemental light top-up** | `needed = max(0, crop_dli_target − natural_clear_sky_dli × cloud_factor)` | "Add ~4h of light today to hit 30 DLI" |
| **Greenhouse venting / shade** | Outdoor temp + solar high → vent/shade nudge | Climate tab suggestion |
| **Heating** | Forecast/sensor low < zone min | "Frost tonight — your Veg Tent min is 18°C" |
| **Transpiration / VPD context** | Outdoor RH + temp informs indoor load | Grow advisor cross-reference |
| **Irrigation tweak** | Hot dry day → suggest higher feed frequency | Feeding hint |

Greenhouse vs sealed-indoor: effects scale by `zone` type — sealed rooms get HVAC-load framing, greenhouses get passive-climate framing.

---

## WS5 — Guardian grounding

**Read tool** `site_weather(farm_id, date?)` → solar + latest sensor + cached forecast.

**Starters (weather-aware surfaces):**
- "Do I need supplemental light today?" (uses natural DLI vs crop target)
- "Is there frost risk tonight?" (forecast or seasonal-low heuristic)
- "Should I vent the greenhouse this afternoon?"
- "How long is daylight right now?" (always answerable, offline)

**Persona:** state which tier the answer is from ("based on sun position — no live forecast available offline").

---

## WS6 — Docs, tests, OC-66

- `farm-guardian-architecture.md` §11 weather tiers + offline solar
- **Offline-first note:** solar + sensor tiers work with the uplink unplugged
- Go test: `SolarForDate` matches known sunrise/sunset for a fixed lat/long/date
- `phase-66-closure.test.js` — site coords UI; supplemental-light starter

---

## Definition of done

- [x] Daylight hours + DLI computed with internet disabled
- [x] Supplemental-light recommendation uses natural DLI vs crop target
- [ ] Optional online provider opt-in, caches, degrades gracefully (Tier 3 → [Phase 178](phase_178_online_weather_forecast.plan.md))
- [x] Guardian states which tier it answered from
- [x] OC-66 closed

---

## Boundary

- **Not** a weather forecasting service — Tier 3 just relays a provider
- **Not** required to be online — Tiers 1–2 are the product; Tier 3 is a convenience
- No new heavy dependency for solar — it's arithmetic in stdlib
