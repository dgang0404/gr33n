---
name: Phase 164 — Demo seed: living farm, no cannabis
overview: >
  The demo farm looks broken and reads like a cannabis grow. Sensors exist in
  the DB but have zero sensor_readings rows, so every tile on Today says
  "NO DATA"; the Veg/Flower rooms run Blue Dream / Gorilla Glue #4 / OG Kush
  cycles. This phase reseeds farm 1 as a living, non-cannabis farm: realistic
  recent readings (with all three health states represented), cannabis crop
  cycles rethemed to chrysanthemum (short-day — the 12/12 schedules stay
  truthful), and one indoor zone set up as a gravity-fed drip demo. This is
  the data foundation for the Phase 166 visual farm canvas.
todos:
  - id: ws1-decannabis-crops
    content: "WS1: Retheme farm-1 crop cycles/plants — cannabis → chrysanthemum (catalog + field guide untouched)"
    status: completed
  - id: ws2-seed-readings
    content: "WS2: Seed recent sensor_readings for wired sensors — healthy baseline values"
    status: completed
  - id: ws3-health-states
    content: "WS3: Curate three hardware states — healthy / needs attention / not set up yet"
    status: completed
  - id: ws4-gravity-drip
    content: "WS4: Gravity-drip demo — Herb & Greens Room drip actuator + irrigation_only program"
    status: completed
  - id: ws5-test-audit
    content: "WS5: Audit + update smoke tests that assume farm-1 cannabis seed rows"
    status: pending
  - id: ws6-closure
    content: "WS6: Seed verify queries + phase-164-closure test + docs note"
    status: pending
isProject: false
---

# Phase 164 — Demo seed: living farm, no cannabis

**Status:** planned · **Feeds:** Phase 165 (layout API) → 166 (visual canvas)

## Why

Two problems undermine the Today redesign before it starts:

1. **"NO DATA" everywhere.** `db/seeds/master_seed.sql` creates 14 sensors for
   farm 1 but never inserts a single `gr33ncore.sensor_readings` row. The
   dashboard sensor grid renders 14 dead tiles. A visual farm canvas built on
   this data would look broken on first open.
2. **Cannabis demo crops.** Crop cycles are Blue Dream / Gorilla Glue #4 /
   OG Kush; the plants catalog row for farm 1 is `cannabis`. The crop catalog
   and field guide keep cannabis (that is enough coverage) — the *demo farm*
   should showcase a different flower.

## WS1 — Retheme demo crops (cannabis → chrysanthemum)

**Crop choice:** chrysanthemum. It is already in the crop catalog
(`db/seed/crop_catalog_from_yaml.sql`, `cousin_of = 'rose'`) and — like
cannabis — is a genuine **short-day plant**: 12 h uninterrupted dark triggers
bloom. Every existing "Light ON/OFF 12/12 Flower" schedule, the "blackout
curtains before the flip" alert, and the veg→flower room flow stay
horticulturally accurate with zero schedule changes. (Alternate if vetoed:
cut-flower rose, which has a field guide but is long-day — light schedules
would need rewording.)

Changes, all in `db/seeds/master_seed.sql` (+ the Phase 124 crop-cycle block
around lines 1396–1491):

| Seed row | Now | Becomes |
|----------|-----|---------|
| `gr33ncrops.plants` farm-1 row | `Cannabis / Mixed photoperiod / cannabis` | `Chrysanthemum / Mixed spray varieties / chrysanthemum` |
| Veg Room cycle | `Veg canopy (18/6)` — Blue Dream | `Veg canopy (18/6)` — batch `'Anastasia Green'` (chrysanthemum, long-day veg under 18/6) |
| Flower Room cycle | `Flower run (12/12)` — Gorilla Glue #4 | `Bloom run (12/12)` — batch `'Zembla White'` |
| Harvested history | `Blue Dream — Run 3 (harvested)` | `Anastasia Green — Run 3 (harvested)` (yield_grams → stem count note in cycle_notes) |
| Propagation Room | `OG Kush — Clone Batch 12` | `Chrysanthemum — Cutting Batch 12` (chrysanthemums propagate from cuttings — copy stays true) |
| Task "Harvest Flower Room A" | "Check trichomes" | "Check bloom openness / stem length" |
| Alert "Humidity high — Flower Room" | "late flower … powdery mildew risk" | keep — powdery mildew is a top chrysanthemum/rose disease; reword "late flower" → "bloom stage" |

**Do NOT touch:** crop catalog entries, field-guide docs (`crop-cannabis-nutrition`
stays), JADAM copy ("JLF General (Weed and Grass)" is fermented lawn weeds, not
cannabis), Recipe Pack v7 catalog rows (Phase 108 catalog data, not farm-1
demo), non-farm-1 test fixtures.

Idempotency: the seed's existing `NOT EXISTS` guards key on names — renamed
rows need `UPDATE`-style migration for already-seeded DBs or a
`make dev-stack-fresh` note. Prefer delete-and-reinsert guarded blocks matching
the existing Phase 124 cleanup pattern (lines 70–82).

## WS2 — Seed sensor readings

Add a new seed section (after the sensor→zone assignment block, ~line 989)
inserting `gr33ncore.sensor_readings` for wired sensors, relative to `NOW()`
so the demo never goes stale:

- **Backfill:** 24 h of readings at each sensor's `reading_interval_seconds`
  is overkill for seed; insert a sparse series (e.g. every 30 min for the last
  6 h + one "just now" row) per sensor. Values inside `value_min/max_expected`
  and within alert thresholds, with mild sinusoidal-ish variation via
  `generate_series` + offsets.
- Set `sensors.last_reading_time = NOW()` for seeded sensors (the reservoir
  block at line 789 already models this pattern).
- Respect the normalize trigger (`normalize_sensor_reading`) — insert raw
  `value` + `unit_id`, let the trigger fill normalized columns.
- Baseline values: Air Temp Indoor 24.2 °C, Root Zone Temp 21.5 °C, Media
  Moisture 46 %, EC 1.6 mS/cm, pH 6.1, CO2 950 ppm, Lux 28 000, PAR 620,
  Soil Moisture Outdoor 41 %.

## WS3 — Three health states, on purpose

| State | Zone / sensor | How seeded |
|-------|---------------|------------|
| **Healthy** | Veg Room (7 sensors) + Outdoor Garden soil moisture | WS2 in-range readings |
| **Needs attention** | Flower Room `Air Humidity Indoor` | Latest reading **72.4 % RH** — matches the existing seeded alert "Humidity high — Flower Room" exactly, so alert + reading tell one coherent story. PAR Sensor Indoor stays healthy. |
| **Not set up yet** | Phase 124 bed sensors (Propagation Dome Temp, Herb Room Air Temp, Pepper Bed + Berry Patch Soil Moisture) | **No readings inserted** — these stay "unwired until an operator assigns hardware," per the existing seed comment. Phase 166 renders this state as "Not set up yet," not "NO DATA." |

Document the intended state per sensor in a seed comment table so future
phases don't "fix" the intentionally-unwired ones.

## WS4 — Gravity-drip demo zone

Herb & Greens Room becomes the try-it-at-home gravity drip story:

- New actuator: `Herb Room Gravity Drip Valve`, `actuator_type = 'drip'`, on
  the existing demo device pattern (`config: {"channel": N, "simulation": true}`).
- New program: `Herb Room Gravity Drip` — `irrigation_only = TRUE`, no
  reservoir mix, short `run_duration_seconds` (e.g. 180), linked to a new
  daily schedule `Water Herbs Gravity Drip Daily` (morning). Description spells
  out the real-world setup: elevated bucket/tank, drip line, valve timed by
  the platform — no pump required.
- One seeded `fertigation_events` row (recent) so "Recent feeds" / zone Water
  tab shows it ran.

This is exactly the plain-irrigation path Phase 39b shipped; nothing new in
the backend — seed + naming only.

## WS5 — Test audit

Grep-verified blast radius (tests referencing farm-1 *seed* rows, not the
catalog):

- `cmd/api/smoke_phase86_test.go` — comments/logic assume "Phase 124 demo seed
  already has a permanent cannabis plants row" on farm 1; it also inserts its
  own cannabis plant. Verify it still passes when the seed row is
  chrysanthemum; adjust the duplicate-row expectations if needed. The
  catalog-level `crop-profiles` early_flower lookup is untouched (catalog keeps
  cannabis).
- `cmd/api/smoke_phase94/96/97/98_test.go` — operate on the **catalog**
  cannabis profile, not seed cycles; expected to pass unchanged. Run to confirm.
- `cmd/api/smoke_phase108_test.go` — Recipe Pack v7 cannabis program is
  catalog/migration data, untouched.
- Guardian QA fixtures / `make guardian-qa-*` prompts that name Blue Dream or
  Gorilla Glue — grep `internal/farmguardian` + `scripts/` and update prompt
  fixtures to the new batch labels.
- UI tests asserting seed names (grep `ui/src/__tests__` for `Blue Dream`,
  `Gorilla`, `OG Kush`).

## WS6 — Closure

- Extend the seed VERIFY section: reading counts per sensor, zero farm-1 rows
  where `batch_label ~* 'blue dream|gorilla|og kush'`, drip program exists.
- `ui/src/__tests__/phase-164-closure.test.js` — guards: master_seed.sql
  contains a `sensor_readings` insert section; contains no cannabis batch
  labels; contains the gravity-drip program name.
- Note in `docs/current-state.md` (demo farm description).

## Acceptance criteria

1. `make dev-stack-fresh` → Today shows live values on all Veg Room tiles, a
   red/amber Flower Room humidity, and clearly-unwired bed sensors.
2. `psql`: farm 1 has zero crop cycles or plants with cannabis crop_key or
   cannabis strain batch labels; crop catalog still has `cannabis`.
3. Herb & Greens Room shows a drip actuator + active gravity-drip program with
   one logged event.
4. `go test ./cmd/api/... -run 'Phase86|Phase94|Phase96|Phase97|Phase98|Phase108'`
   green; `cd ui && npm test -- --run` green.

## Verification

```bash
make dev-stack-fresh
psql "$DATABASE_URL" -c "SELECT s.name, count(r.*) FROM gr33ncore.sensors s LEFT JOIN gr33ncore.sensor_readings r ON r.sensor_id = s.id WHERE s.farm_id=1 GROUP BY 1 ORDER BY 1;"
psql "$DATABASE_URL" -c "SELECT batch_label FROM gr33nfertigation.crop_cycles WHERE farm_id=1;"
go test ./cmd/api/... -count=1
cd ui && npm test -- --run src/__tests__/phase-164-closure.test.js
```
