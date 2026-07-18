-- Phase 180 — resync every serial/identity sequence to max(id) across the
-- platform schemas. db/seeds/master_seed.sql inserts several rows (farm 1,
-- its zones, sensors, etc.) with explicit ids, so their sequences never
-- advanced via nextval(). The first real INSERT through the API after
-- seeding (e.g. POST /farms) then calls nextval() -> 1, collides with the
-- seeded row, and fails with a 500. This migration is the one-time fix for
-- databases seeded before the master_seed.sql fix landed; safe to re-run.
DO $$
DECLARE
  r RECORD;
  max_id BIGINT;
BEGIN
  FOR r IN
    SELECT n.nspname AS schema_name, t.relname AS table_name,
           a.attname AS col_name, s.relname AS seq_name
    FROM pg_class s
    JOIN pg_depend d ON d.objid = s.oid AND d.deptype = 'a'
    JOIN pg_class t ON d.refobjid = t.oid
    JOIN pg_namespace n ON n.oid = t.relnamespace
    JOIN pg_attribute a ON a.attrelid = t.oid AND a.attnum = d.refobjsubid
    WHERE s.relkind = 'S'
      AND n.nspname IN (
        'gr33ncore', 'gr33ncrops', 'gr33nfertigation',
        'gr33nnaturalfarming', 'gr33naquaponics', 'auth'
      )
  LOOP
    EXECUTE format('SELECT COALESCE(MAX(%I), 0) FROM %I.%I', r.col_name, r.schema_name, r.table_name)
      INTO max_id;
    PERFORM setval(format('%I.%I', r.schema_name, r.seq_name), GREATEST(max_id, 1), max_id > 0);
  END LOOP;
END $$;
