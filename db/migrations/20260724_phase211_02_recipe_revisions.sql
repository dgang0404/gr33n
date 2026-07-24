-- Phase 211.02 WS1 — Immutable application recipe revision snapshots
-- Phase 211.02 WS2 — programs pin revision at link time (nullable for legacy rows)

CREATE TABLE IF NOT EXISTS gr33nnaturalfarming.application_recipe_revisions (
    id                      BIGSERIAL PRIMARY KEY,
    application_recipe_id   BIGINT NOT NULL
        REFERENCES gr33nnaturalfarming.application_recipes(id) ON DELETE CASCADE,
    revision_number         INTEGER NOT NULL CHECK (revision_number > 0),
    snapshot                JSONB NOT NULL,
    change_summary          TEXT,
    created_at              TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_by_user_id      UUID REFERENCES gr33ncore.profiles(user_id) ON DELETE SET NULL,
    UNIQUE (application_recipe_id, revision_number)
);

CREATE INDEX IF NOT EXISTS idx_application_recipe_revisions_recipe
    ON gr33nnaturalfarming.application_recipe_revisions (application_recipe_id, revision_number DESC);

ALTER TABLE gr33nfertigation.programs
    ADD COLUMN IF NOT EXISTS application_recipe_revision_id BIGINT
        REFERENCES gr33nnaturalfarming.application_recipe_revisions(id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_programs_recipe_revision
    ON gr33nfertigation.programs (application_recipe_revision_id)
    WHERE application_recipe_revision_id IS NOT NULL AND deleted_at IS NULL;

-- Bootstrap revision 1 for every active recipe (audit truth at migration time)
INSERT INTO gr33nnaturalfarming.application_recipe_revisions (
    application_recipe_id,
    revision_number,
    snapshot,
    change_summary,
    created_at,
    created_by_user_id
)
SELECT
    r.id,
    1,
    jsonb_build_object(
        'recipe', jsonb_build_object(
            'id', r.id,
            'farm_id', r.farm_id,
            'name', r.name,
            'input_definition_id', r.input_definition_id,
            'description', r.description,
            'target_application_type', r.target_application_type::text,
            'dilution_ratio', r.dilution_ratio,
            'instructions', r.instructions,
            'frequency_guidelines', r.frequency_guidelines,
            'notes', r.notes
        ),
        'components', COALESCE((
            SELECT jsonb_agg(
                jsonb_build_object(
                    'input_definition_id', c.input_definition_id,
                    'input_name', d.name,
                    'part_value', c.part_value,
                    'part_unit_id', c.part_unit_id,
                    'notes', c.notes
                )
                ORDER BY d.name
            )
            FROM gr33nnaturalfarming.recipe_input_components c
            JOIN gr33nnaturalfarming.input_definitions d ON d.id = c.input_definition_id
            WHERE c.application_recipe_id = r.id
        ), '[]'::jsonb)
    ),
    'bootstrap backfill (phase 211.02)',
    r.created_at,
    r.updated_by_user_id
FROM gr33nnaturalfarming.application_recipes r
WHERE r.deleted_at IS NULL
  AND NOT EXISTS (
      SELECT 1
      FROM gr33nnaturalfarming.application_recipe_revisions rev
      WHERE rev.application_recipe_id = r.id
  );

-- Pin existing programs to bootstrap revision 1 where recipe is set
UPDATE gr33nfertigation.programs p
SET application_recipe_revision_id = rev.id
FROM gr33nnaturalfarming.application_recipe_revisions rev
WHERE p.deleted_at IS NULL
  AND p.application_recipe_id IS NOT NULL
  AND p.application_recipe_revision_id IS NULL
  AND rev.application_recipe_id = p.application_recipe_id
  AND rev.revision_number = 1;
