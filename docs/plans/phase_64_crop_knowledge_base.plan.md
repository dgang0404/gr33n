---
name: Phase 64 — Crop knowledge base & target profiles
overview: >
  The foundation Phase 62 grow advisor assumed but never built — a curated,
  versioned, OFFLINE-bundled library of per-crop, per-stage targets (EC, pH, VPD,
  temp, RH, DLI, photoperiod). Guardian grounds every numeric answer on this data
  instead of guessing. Operators clone + override profiles for their own genetics.
todos:
  - id: ws1-schema
    content: "WS1: crop_profiles + crop_profile_stages tables; crop_profile_id FK on plants"
    status: pending
  - id: ws2-bundled-data
    content: "WS2: Curated seed — cannabis, tomato, pepper, lettuce, orchid, basil, strawberry; versioned + sourced"
    status: pending
  - id: ws3-guardian-grounding
    content: "WS3: crop_profile read tool; grow advisor cites real targets; RAG crop guides for narrative"
    status: pending
  - id: ws4-ui
    content: "WS4: Profile picker in start-grow wizard; target-vs-reading chips on zone; profile detail view"
    status: pending
  - id: ws5-override-share
    content: "WS5: Clone-and-override per farm; export/import profile JSON; optional community share"
    status: pending
  - id: ws6-docs-tests
    content: "WS6: farm-guardian-architecture §10; field-guide manifest crop entries; phase-64-closure; OC-64"
    status: pending
isProject: false
---

# Phase 64 — Crop knowledge base & target profiles

## Status

**Planned.** **Prerequisite for [Phase 62](phase_62_guardian_grow_advisor.plan.md)** (grow advisor needs real targets to cite) and pairs with [Phase 56](phase_56_grow_schema_harvest_analytics.plan.md) `plant_id` FK.

**Arc:** [phase_53_59_roadmap.plan.md](phase_53_59_roadmap.plan.md)

---

## The problem this solves

> *"How will Guardian know the EC target of a tomato vs cannabis vs pepper vs orchid?"*

It **won't guess.** Today `crop_cycles.strain_or_variety` is free text and there is **no target data anywhere** — so a grow advisor would either hallucinate numbers (dangerous) or stay generic (useless). This phase gives Guardian a real reference table to read from.

### Design principle

| Layer | Holds | Used for |
|-------|-------|----------|
| **Structured DB profiles** | EC, pH, VPD, temp, RH, DLI, photoperiod per crop per stage | The **math** — Guardian cites exact numbers, never invents them |
| **RAG crop guides (markdown)** | Narrative — why orchids want low EC, cannabis flower ramps, blossom-end rot | The **explanation** — grounded prose, ingested into `field_guide` collection |
| **LLM** | Reasoning over the above | Synthesis only — **forbidden from stating a target not in the profile** |

This mirrors the existing split: Phase 37 RAG field guides for narrative, structured tables for numbers.

---

## WS1 — Schema

```sql
CREATE TABLE gr33ncrops.crop_profiles (
  id            BIGSERIAL PRIMARY KEY,
  farm_id       BIGINT REFERENCES farms(id),  -- NULL = built-in/global
  crop_key      TEXT NOT NULL,                 -- 'cannabis', 'tomato', 'phalaenopsis'
  display_name  TEXT NOT NULL,
  category      TEXT,                           -- 'fruiting', 'leafy', 'epiphyte', 'flower'
  source        TEXT,                           -- citation / provenance
  version       INTEGER NOT NULL DEFAULT 1,
  is_builtin    BOOLEAN NOT NULL DEFAULT FALSE,
  meta          JSONB NOT NULL DEFAULT '{}'
);

CREATE TABLE gr33ncrops.crop_profile_stages (
  id              BIGSERIAL PRIMARY KEY,
  crop_profile_id BIGINT NOT NULL REFERENCES crop_profiles(id) ON DELETE CASCADE,
  stage           gr33nfertigation.growth_stage_enum NOT NULL,
  ec_min          NUMERIC(4,2), ec_target NUMERIC(4,2), ec_max NUMERIC(4,2),  -- mS/cm
  ph_min          NUMERIC(3,1), ph_max    NUMERIC(3,1),
  vpd_min_kpa     NUMERIC(3,2), vpd_max_kpa NUMERIC(3,2),
  temp_min_c      NUMERIC(4,1), temp_max_c  NUMERIC(4,1),
  rh_min_pct      NUMERIC(4,1), rh_max_pct  NUMERIC(4,1),
  dli_target      NUMERIC(4,1),                 -- mol/m²/day
  photoperiod_hrs NUMERIC(3,1),
  notes           TEXT
);

ALTER TABLE gr33ncrops.plants ADD COLUMN crop_profile_id BIGINT
  REFERENCES gr33ncrops.crop_profiles(id) ON DELETE SET NULL;
```

EC stored as **mS/cm** canonically; UI converts to PPM (500/700 scale) if operator prefers.

---

## WS2 — Bundled starter data (offline)

Ships with the platform as SQL seed + YAML source — **no internet to populate.**

| Crop | Why it's in v1 | Notable contrast |
|------|----------------|------------------|
| **Cannabis** | Primary user base | EC ramps 0.8 → 2.0 veg→flower; photoperiod flip 18→12 |
| **Tomato** | High-EC fruiting | EC up to 3.5 in fruiting; very different from cannabis |
| **Pepper** | Similar-but-lower than tomato | Shows nuance vs tomato |
| **Lettuce / leafy** | Low EC, high turnover | EC ~0.8–1.2; cool temps |
| **Orchid (phalaenopsis)** | Epiphyte extreme | **Very low EC ~0.4–0.8**, high RH, low light — the "opposite" case |
| **Basil / herbs** | Common companion | Moderate EC, warm |
| **Strawberry** | Day-neutral vs June | Photoperiod sensitivity teaching case |

Each profile carries a `source` citation. Versioned so updates don't silently change a farmer's active targets.

---

## WS3 — Guardian grounding

**Read tool** `lookup_crop_targets(crop_profile_id | crop_key, stage)` → returns the stage row.

**Grow advisor (Phase 62) rewrite:** instead of "VPD target for flower," Guardian calls the tool and answers:

> "Your Blue Dream is in early flower. Target EC is **1.6–1.8 mS/cm** (you're at 1.4 — a little light). VPD target **1.0–1.2 kPa**; you're at 1.35, slightly high."

**Hard rule in persona prompt:**
```
NEVER state an EC, pH, VPD, or DLI target unless it comes from lookup_crop_targets.
If no profile is assigned, say so and offer to set one — do not guess from general knowledge.
```

**RAG crop guides:** add `docs/field-guides/crop-*.md` (cannabis-nutrition, tomato-nutrition, orchid-care…) to [field-guide-manifest.yaml](../rag/field-guide-manifest.yaml) for the *why*.

---

## WS4 — UI

- **Start-grow wizard (Phase 53 WS1.2):** crop profile dropdown → sets plant + targets in one step
- **Zone grow strip:** "EC 1.4 / target 1.6–1.8" chip with under/over color
- **Profile detail view:** read-only table of all stages; "Clone to edit" button
- **Comfort targets page:** "Apply crop profile" → pre-fills bands from the profile

---

## WS5 — Override & share

- **Clone-and-override:** built-in profiles are read-only; operator clones to a farm-scoped copy and tweaks for their genetics (their Blue Dream pheno runs hotter)
- **Export/import:** profile → JSON file (share with a friend, back up)
- **Optional community share:** opt-in publish (reuses `insert_commons_opt_in` consent pattern) — deferred toggle, not v1 required

---

## WS6 — Docs, tests, OC-64

- `farm-guardian-architecture.md` §10 crop knowledge grounding
- Crop guide markdown + RAG manifest entries
- `phase-64-closure.test.js` — profile picker; target chip; tool grounding
- Go test: `lookup_crop_targets` returns seeded cannabis flower EC range
- **Persona guard test:** Guardian refuses to state a target with no profile assigned

---

## Definition of done

- [ ] 7 built-in crop profiles seeded offline
- [ ] Start-grow wizard assigns a profile
- [ ] Guardian cites real EC/VPD from the profile, refuses to guess without one
- [ ] Operator can clone + override
- [ ] OC-64 closed

---

## Boundary

- **Not** an agronomy ground truth authority — profiles are starting points with citations, operator owns final targets
- **Not** per-pheno ML tuning (future) — static curated ranges v1
- LLM stays a synthesizer over real data, never the source of numbers
