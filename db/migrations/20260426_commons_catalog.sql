-- Phase 14 WS3: gr33n_inserts / commons catalog — browse + farm import audit (no marketplace)

CREATE TABLE IF NOT EXISTS gr33ncore.commons_catalog_entries (
    id                   BIGSERIAL PRIMARY KEY,
    slug                 TEXT        NOT NULL UNIQUE,
    title                TEXT        NOT NULL,
    summary              TEXT        NOT NULL DEFAULT '',
    body                 JSONB       NOT NULL DEFAULT '{}'::jsonb,
    contributor_display  TEXT        NOT NULL DEFAULT '',
    contributor_uri      TEXT,
    license_spdx         TEXT        NOT NULL DEFAULT 'CC-BY-4.0',
    license_notes        TEXT,
    tags                 TEXT[]      NOT NULL DEFAULT ARRAY[]::TEXT[],
    published            BOOLEAN     NOT NULL DEFAULT FALSE,
    sort_order           INT         NOT NULL DEFAULT 0,
    created_at           TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at           TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_commons_catalog_published_sort
    ON gr33ncore.commons_catalog_entries (published, sort_order, title)
    WHERE published = TRUE;

CREATE TABLE IF NOT EXISTS gr33ncore.farm_commons_catalog_imports (
    id                 BIGSERIAL PRIMARY KEY,
    farm_id            BIGINT      NOT NULL REFERENCES gr33ncore.farms(id) ON DELETE CASCADE,
    catalog_entry_id   BIGINT      NOT NULL REFERENCES gr33ncore.commons_catalog_entries(id) ON DELETE CASCADE,
    imported_by        UUID        NOT NULL REFERENCES gr33ncore.profiles(user_id) ON DELETE CASCADE,
    imported_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    note               TEXT,
    UNIQUE (farm_id, catalog_entry_id)
);

CREATE INDEX IF NOT EXISTS idx_farm_commons_imports_farm
    ON gr33ncore.farm_commons_catalog_imports (farm_id, imported_at DESC);

DO $$ BEGIN
  IF NOT EXISTS (
    SELECT 1 FROM pg_trigger WHERE tgname = 'trg_commons_catalog_entries_updated_at'
  ) THEN
    CREATE TRIGGER trg_commons_catalog_entries_updated_at
      BEFORE UPDATE ON gr33ncore.commons_catalog_entries
      FOR EACH ROW EXECUTE FUNCTION gr33ncore.set_updated_at();
  END IF;
END $$;

-- Demo published pack (documentation / metadata only — no executable SQL in v1)
INSERT INTO gr33ncore.commons_catalog_entries (
    slug, title, summary, body, contributor_display, contributor_uri,
    license_spdx, license_notes, tags, published, sort_order
) VALUES (
    'gr33n-insert-commons-v1-readme',
    'Insert Commons v1 — operator reference',
    'Links and context for coarse aggregate sharing (Insert Commons); validation, opt-in, receiver contract.',
    '{"catalog_version":"gr33n.commons_catalog.v1","kind":"documentation_pack","readme_md":"# Insert Commons v1\n\nFarm operators share **pseudonymous coarse aggregates** when opted in. See the pipeline runbook for JSON shape, preview, approval bundles, and strict validation.\n\nThis catalog entry carries **no executable SQL** — it is a signpost for humans and tools.\n","related_urls":[{"label":"Pipeline runbook (farm API)","path":"docs/insert-commons-pipeline-runbook.md"},{"label":"Receiver playbook","path":"docs/insert-commons-receiver-playbook.md"}]}'::jsonb,
    'gr33n platform',
    NULL,
    'CC-BY-4.0',
    'Share and adapt with attribution; not legal advice.',
    ARRAY['insert-commons', 'federation', 'documentation', 'gr33n_inserts'],
    TRUE,
    0
) ON CONFLICT (slug) DO NOTHING;
