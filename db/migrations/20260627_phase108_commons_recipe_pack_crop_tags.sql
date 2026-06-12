-- Phase 108 — tag commons recipe pack v7 programs with Phase 102 crop_key metadata.
-- Source mirror: scripts/enterprise/sample-recipe-pack-v7.body.json

UPDATE gr33ncore.commons_catalog_entries
SET body = $body$
{
  "catalog_version": "gr33n.commons_catalog.v1",
  "kind": "fertigation_recipe_pack",
  "pack_version": "7",
  "pack_id": "gr33n-recipe-pack-v7-lettuce-veg",
  "readme_md": "# Recipe Pack v7 — Lettuce veg / flower (demo)\n\nHypothetical **HQ → site** promotion pack for multi-farm integrators. This JSON is **opaque payload** stored in `commons_catalog_entries.body` — the API does not auto-apply it.\n\n## What integrators do\n\n1. Curator publishes this row in the commons catalog.\n2. Per site, farm admin runs `scripts/enterprise/import-recipe-pack.sh` to record catalog import audit + create programs via public API.\n\nPrograms import with **`is_active: false`** — review before enabling automation.\n",
  "promotion_note": "Phase 31 WS5 demo — not a live agronomic prescription.",
  "programs": [
    {
      "pack_program_key": "lettuce-veg-standard",
      "name": "Recipe Pack v7 — Lettuce Veg Standard",
      "description": "Conservative veg-stage EC/pH triggers for indoor lettuce (demo pack).",
      "total_volume_liters": 2.0,
      "ec_trigger_low": 1.0,
      "ph_trigger_low": 5.8,
      "ph_trigger_high": 6.4,
      "is_active": false,
      "target_zone_name_hint": "Veg Room",
      "recommended_crop_keys": ["lettuce"],
      "recommended_stages": ["early_veg", "late_veg"],
      "profile_ec_source": { "crop_key": "lettuce", "stage": "early_veg" },
      "ec_band_mscm": { "min": 0.8, "max": 1.3 }
    },
    {
      "pack_program_key": "lettuce-flower-boost",
      "name": "Recipe Pack v7 — Lettuce Flower Boost",
      "description": "Late-veg / pre-harvest feed profile (demo pack). Enable only after operator review.",
      "total_volume_liters": 1.5,
      "ec_trigger_low": 1.4,
      "ph_trigger_low": 5.9,
      "ph_trigger_high": 6.3,
      "is_active": false,
      "target_zone_name_hint": "Flower Room",
      "recommended_crop_keys": ["lettuce"],
      "recommended_stages": ["late_veg"],
      "profile_ec_source": { "crop_key": "lettuce", "stage": "late_veg" },
      "ec_band_mscm": { "min": 0.9, "max": 1.3 }
    },
    {
      "pack_program_key": "cannabis-flower-standard",
      "name": "Recipe Pack v7 — Cannabis Flower Standard",
      "description": "Conservative flower-stage EC/pH triggers for indoor cannabis (demo pack).",
      "total_volume_liters": 2.0,
      "ec_trigger_low": 1.6,
      "ph_trigger_low": 5.8,
      "ph_trigger_high": 6.2,
      "is_active": false,
      "target_zone_name_hint": "Flower Room",
      "recommended_crop_keys": ["cannabis"],
      "recommended_stages": ["early_flower", "mid_flower", "late_flower"],
      "profile_ec_source": { "crop_key": "cannabis", "stage": "early_flower" },
      "ec_band_mscm": { "min": 1.6, "max": 2.4 }
    }
  ]
}
$body$::jsonb,
    updated_at = NOW()
WHERE slug = 'gr33n-recipe-pack-v7-lettuce-veg';
