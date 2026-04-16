# Domain module stubs (crops, animals, aquaponics)

**Phase 14 WS7** adds three PostgreSQL **schemas** with minimal **placeholder tables** so future product work can extend farm-scoped domains without a big-bang migration. There are **no REST endpoints** for these tables yet.

## Schemas and tables

| Schema | Stub table | Purpose (future) |
|--------|------------|------------------|
| `gr33ncrops` | `plants` | Crop / cultivar tracking per farm |
| `gr33nanimals` | `animal_groups` | Herd, flock, or pen groups |
| `gr33naquaponics` | `loops` | Aquaponics system / loop definitions |

Each row is scoped with `farm_id` → `gr33ncore.farms`. Optional `meta` JSONB holds unstructured extension data. `deleted_at` is reserved for soft-delete patterns used elsewhere in gr33n.

## Enabling a module for a farm

The product convention is to record intent in **`gr33ncore.farm_active_modules`**:

- `module_schema_name` must match the schema exactly: `gr33ncrops`, `gr33nanimals`, or `gr33naquaponics`.
- `is_enabled` — typically `TRUE` when the farm opts in.
- `configuration` — optional JSON for module-specific settings later.

Example (SQL):

```sql
INSERT INTO gr33ncore.farm_active_modules (farm_id, module_schema_name, is_enabled, configuration)
VALUES (1, 'gr33ncrops', TRUE, '{}'::jsonb)
ON CONFLICT (farm_id, module_schema_name) DO UPDATE SET is_enabled = EXCLUDED.is_enabled;
```

Until application code enforces this flag, treat it as **documentation + future RBAC/feature gating**.

## Migration

Apply:

- `db/migrations/20260428_phase14_domain_module_stubs.sql`

Or load the canonical full schema (`db/schema/gr33n-schema-v2-FINAL.sql`), which includes the same definitions.

**Rollback** (development only; destroys all data in these schemas):

```sql
DROP SCHEMA IF EXISTS gr33naquaponics CASCADE;
DROP SCHEMA IF EXISTS gr33nanimals CASCADE;
DROP SCHEMA IF EXISTS gr33ncrops CASCADE;
```

## Related docs

- Phase14 plan: [`plans/phase_14_network_and_commons.plan.md`](plans/phase_14_network_and_commons.plan.md)
- Phase 14 operator index: [`phase-14-operator-documentation.md`](phase-14-operator-documentation.md)
