---
domain: natural_farming
title: Indoor photoperiod JADAM programs (bootstrap)
safety_tier: safe
tradition: jadam
reference_source: "gr33ncore._bootstrap_jadam_indoor_photoperiod_v1"
source_tier: cho_named
---

# Indoor photoperiod JADAM programs (bootstrap)

## What it is (1 paragraph)

Maps the **`jadam_indoor_photoperiod_v1`** bootstrap template — veg (18/6), flower (12/12), and outdoor JLF drench programs wired to audited application recipes on demo and new farms.

## When to use

After applying bootstrap template or when mirroring demo farm fertigation layout. Commercial EC programs stay on **Feed & water** — these are parallel natural paths.

## Ingredients (list with amounts)

Program-linked batches (template IDs **TPL-JLF-GEN-001**, **TPL-JMS-001**, **TPL-FFJ-001**, **TPL-WCA-001**) — see input guides.

## Step-by-step preparation

1. Apply bootstrap `jadam_indoor_photoperiod_v1` or use demo farm seed.
2. Confirm batches exist for JLF, JMS, FFJ, WCA.
3. Refresh reservoir mix tasks before scheduled irrigations.

## Ferment / wait timeline

Maintain rolling JMS at peak foam; JLF/FFJ batches per input guide storage windows.

## Ready signs (smell, foam, color)

Batch status **ready_for_use** in inventory before linking to a mix event.

## Storage

Veg **Main Nutrient Reservoir** and flower **Flower Nutrient Reservoir** — mix same day as irrigation schedule.

## Safety & water (non-chlorinated, PPE)

Non-chlorinated make-up water in reservoirs. JMS at **1:10** with JLF **1:20** in combined veg tank (not legacy 1:500 JMS).

## How to apply (link to application recipe name)

| Program | Zone | Recipe | Schedule |
|---------|------|--------|----------|
| **Veg Daily JLF Program** | Veg Room 18/6 | **JLF and JMS Combined Drench** | Water Late Veg Daily |
| **Flower Daily FFJ+WCA Program** | Flower Room 12/12 | **FFJ and WCA Flowering Boost** | Water Early Flower Daily |
| **Outdoor JLF Soil Drench** | Outdoor Garden | **JLF General Soil Drench** | Water Outdoor Garden Daily |

See [application recipes](natural-farming-application-recipes.md) for dilutions.

## Dilution table (start conservative → stronger)

Combined veg tank: JLF **1:20** + JMS **1:10**. Flower tank: FFJ 1:500 + WCA 1:1000. Outdoor: JLF start **1:100** if unsure.

## Common mistakes

- Expecting EC 1.6–1.8 mS/cm Mericle targets without converting mindset — tune volume/cron, not bottle A/B
- Legacy JMS 1:500 in mix notes — pre-audit; refresh to 1:10
