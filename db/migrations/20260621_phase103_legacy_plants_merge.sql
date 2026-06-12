-- Phase 103 — Legacy plant dedupe & catalog backfill for farms created before Phase 85.

DROP FUNCTION IF EXISTS gr33ncrops.merge_legacy_plants();

CREATE OR REPLACE FUNCTION gr33ncrops.merge_legacy_plants()
RETURNS TABLE(
    backfilled INT,
    cycles_relinked INT,
    plants_merged INT,
    unresolved INT
)
LANGUAGE plpgsql
AS $$
DECLARE
    v_backfilled INT := 0;
    v_cycles INT := 0;
    v_merged INT := 0;
    v_unresolved INT := 0;
BEGIN
    CREATE TEMP TABLE _p103_targets (
        plant_id BIGINT PRIMARY KEY,
        farm_id  BIGINT NOT NULL,
        crop_key TEXT NOT NULL
    ) ON COMMIT DROP;

    -- Profile link.
    INSERT INTO _p103_targets (plant_id, farm_id, crop_key)
    SELECT pl.id, pl.farm_id, cp.crop_key
    FROM gr33ncrops.plants pl
    JOIN gr33ncrops.crop_profiles cp ON pl.crop_profile_id = cp.id
    WHERE pl.deleted_at IS NULL
      AND pl.crop_key IS NULL
      AND cp.crop_key IS NOT NULL
    ON CONFLICT (plant_id) DO NOTHING;

    -- Exact catalog display_name.
    INSERT INTO _p103_targets (plant_id, farm_id, crop_key)
    SELECT pl.id, pl.farm_id, ce.crop_key
    FROM gr33ncrops.plants pl
    JOIN gr33ncrops.crop_catalog_entries ce
      ON ce.supported = TRUE
     AND lower(btrim(pl.display_name)) = lower(btrim(ce.display_name))
    WHERE pl.deleted_at IS NULL
      AND pl.crop_key IS NULL
    ON CONFLICT (plant_id) DO NOTHING;

    -- Catalog alias.
    INSERT INTO _p103_targets (plant_id, farm_id, crop_key)
    SELECT pl.id, pl.farm_id, ce.crop_key
    FROM gr33ncrops.plants pl
    JOIN gr33ncrops.crop_catalog_aliases a
      ON lower(btrim(pl.display_name)) = lower(btrim(a.alias))
    JOIN gr33ncrops.crop_catalog_entries ce ON ce.crop_key = a.crop_key AND ce.supported = TRUE
    WHERE pl.deleted_at IS NULL
      AND pl.crop_key IS NULL
    ON CONFLICT (plant_id) DO NOTHING;

    -- Slug-normalized label.
    INSERT INTO _p103_targets (plant_id, farm_id, crop_key)
    SELECT pl.id, pl.farm_id, m.crop_key
    FROM gr33ncrops.plants pl
    JOIN LATERAL (
        SELECT COALESCE(ce.crop_key, ca.crop_key) AS crop_key
        FROM (SELECT lower(regexp_replace(btrim(pl.display_name), '[^a-z0-9]+', '_', 'g')) AS slug) s
        LEFT JOIN gr33ncrops.crop_catalog_entries ce
            ON ce.supported = TRUE AND ce.crop_key = s.slug
        LEFT JOIN gr33ncrops.crop_catalog_aliases ca ON ca.alias = s.slug
    ) m ON m.crop_key IS NOT NULL
    WHERE pl.deleted_at IS NULL
      AND pl.crop_key IS NULL
    ON CONFLICT (plant_id) DO NOTHING;

    CREATE TEMP TABLE _p103_keepers (
        farm_id  BIGINT NOT NULL,
        crop_key TEXT NOT NULL,
        keep_id  BIGINT NOT NULL,
        PRIMARY KEY (farm_id, crop_key)
    ) ON COMMIT DROP;

    INSERT INTO _p103_keepers (farm_id, crop_key, keep_id)
    SELECT farm_id, crop_key, MIN(id)
    FROM (
        SELECT id, farm_id, crop_key
        FROM gr33ncrops.plants
        WHERE deleted_at IS NULL AND crop_key IS NOT NULL
        UNION ALL
        SELECT plant_id, farm_id, crop_key
        FROM _p103_targets
    ) u
    GROUP BY farm_id, crop_key;

    -- Relink cycles off plants that will be merged away.
    WITH doomed AS (
        SELECT p.id AS old_id, k.keep_id
        FROM gr33ncrops.plants p
        JOIN _p103_keepers k ON p.farm_id = k.farm_id
        JOIN _p103_targets t ON t.plant_id = p.id AND t.crop_key = k.crop_key
        WHERE p.deleted_at IS NULL
          AND p.id <> k.keep_id
        UNION
        SELECT p.id, k.keep_id
        FROM gr33ncrops.plants p
        JOIN _p103_keepers k ON p.farm_id = k.farm_id AND p.crop_key = k.crop_key
        WHERE p.deleted_at IS NULL
          AND p.id <> k.keep_id
    )
    UPDATE gr33nfertigation.crop_cycles c
    SET plant_id = d.keep_id
    FROM doomed d
    WHERE c.plant_id = d.old_id;
    GET DIAGNOSTICS v_cycles = ROW_COUNT;

    -- Soft-delete non-keeper rows slotted for the same crop_key.
    WITH doomed AS (
        SELECT p.id
        FROM gr33ncrops.plants p
        JOIN _p103_keepers k ON p.farm_id = k.farm_id
        JOIN _p103_targets t ON t.plant_id = p.id AND t.crop_key = k.crop_key
        WHERE p.deleted_at IS NULL AND p.id <> k.keep_id
        UNION
        SELECT p.id
        FROM gr33ncrops.plants p
        JOIN _p103_keepers k ON p.farm_id = k.farm_id AND p.crop_key = k.crop_key
        WHERE p.deleted_at IS NULL AND p.id <> k.keep_id
    )
    UPDATE gr33ncrops.plants p
    SET deleted_at = NOW()
    FROM doomed d
    WHERE p.id = d.id;
    GET DIAGNOSTICS v_merged = ROW_COUNT;

    -- Bind crop_key on surviving keeper rows (one row per farm+crop_key — safe for unique index).
    UPDATE gr33ncrops.plants pl
    SET crop_key = t.crop_key,
        display_name = ce.display_name,
        crop_profile_id = COALESCE(
            pl.crop_profile_id,
            (SELECT p.id
             FROM gr33ncrops.crop_profiles p
             WHERE p.farm_id IS NULL
               AND p.is_builtin = TRUE
               AND p.crop_key = t.crop_key
             LIMIT 1)
        )
    FROM _p103_targets t
    JOIN _p103_keepers k ON k.farm_id = t.farm_id AND k.crop_key = t.crop_key
    JOIN gr33ncrops.crop_catalog_entries ce ON ce.crop_key = t.crop_key
    WHERE pl.id = k.keep_id
      AND pl.id = t.plant_id
      AND pl.deleted_at IS NULL
      AND pl.crop_key IS NULL;
    GET DIAGNOSTICS v_backfilled = ROW_COUNT;

    -- Also backfill keepers that already existed with crop_key but need catalog display sync.
    UPDATE gr33ncrops.plants pl
    SET display_name = ce.display_name
    FROM gr33ncrops.crop_catalog_entries ce
    WHERE pl.deleted_at IS NULL
      AND pl.crop_key = ce.crop_key
      AND pl.display_name IS DISTINCT FROM ce.display_name;

    SELECT count(*)::INT
    INTO v_unresolved
    FROM gr33ncrops.plants pl
    WHERE pl.deleted_at IS NULL
      AND pl.crop_key IS NULL;

    backfilled := v_backfilled;
    cycles_relinked := v_cycles;
    plants_merged := v_merged;
    unresolved := v_unresolved;
    RETURN NEXT;
END;
$$;

COMMENT ON FUNCTION gr33ncrops.merge_legacy_plants() IS
    'Phase 103 — backfill plants.crop_key from profiles/catalog/aliases, dedupe per farm, relink crop_cycles.plant_id. Idempotent.';

SELECT * FROM gr33ncrops.merge_legacy_plants();
