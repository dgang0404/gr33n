# Enterprise deployment helpers (community extension)

**Status:** Phase 31 WS5 ships a **demo** recipe-pack promotion stub. Larger multi-site suites remain community contributions.

Large multi-site operators (see [`docs/hypothetical-enterprise-topology.md`](../../docs/hypothetical-enterprise-topology.md)) often need **repeatable bring-up**:

- Bulk farm / zone creation from YAML  
- Pi `config.yaml` generation from a device manifest  
- Commons catalog pack import across many `farm_id`s  
- Post-deploy smoke (health, one reading, one actuator round-trip)

## Shipped (Phase 31 WS5)

| Tool | Purpose |
|------|---------|
| [`import-recipe-pack.sh`](import-recipe-pack.sh) | Promote **Recipe Pack v7** demo to comma-separated `farm_id`s via public API |
| [`sample-recipe-pack-v7.body.json`](sample-recipe-pack-v7.body.json) | Opaque `commons_catalog_entries.body` mirror (fertigation program defs + readme) |
| [`db/migrations/20260527_phase31_commons_recipe_pack_v7.sql`](../../db/migrations/20260527_phase31_commons_recipe_pack_v7.sql) | Publishes catalog slug `gr33n-recipe-pack-v7-lettuce-veg` |

### Quick start

```bash
# After migrate + API up (make dev-auth-test)
./scripts/enterprise/import-recipe-pack.sh --dry-run
./scripts/enterprise/import-recipe-pack.sh --farm-ids 1
# Multi-site (when farms exist):
./scripts/enterprise/import-recipe-pack.sh --farm-ids 1,2,3
```

**Idempotency:** catalog import upserts per farm+entry; programs skip when **`name`** already exists. Programs import **`is_active: false`** — enable in UI after review.

**Auth:** farm **admin** JWT (`POST /farms/{id}/commons/catalog-imports`); **Operate** for program create.

## Shipped (Phase 33 WS5) — site manifest

| Tool | Purpose |
|------|---------|
| [`apply-site-manifest.sh`](apply-site-manifest.sh) | Stand up a site from YAML: create farm (optional org), zones, import a recipe pack, print Pi wiring hints |
| [`site-manifest.example.yaml`](site-manifest.example.yaml) | Illustrative schema: `org_slug`, `farm_name`, `zones[]` (name/type), `recipe_pack_slug`, `pi_device_hints` |

```bash
# Plan only (no JWT, no HTTP):
./scripts/enterprise/apply-site-manifest.sh --dry-run \
  --manifest scripts/enterprise/site-manifest.example.yaml

# Real run (API up + farm-admin JWT; needs python3 + PyYAML):
./scripts/enterprise/apply-site-manifest.sh --manifest path/to/site.yaml
```

**Scope:** a starting-point stub (single site), not a 500-site Ansible suite — extend per your fleet. Zones skip when the name already exists; the recipe pack slug must be published in the commons catalog. Pi device hints are **informational** (provision on-site; pairs with the Phase 37 guided wiring procedures).

See [`docs/commons-catalog-operator-playbook.md`](../../docs/commons-catalog-operator-playbook.md) for catalog semantics (import records audit — does not auto-run SQL).

## Shipped (Phase 83) — agronomy seed pack + Guardian bootstrap

**Plan:** [`docs/plans/phase_83_enterprise_agronomy_seed_pack.plan.md`](../../docs/plans/phase_83_enterprise_agronomy_seed_pack.plan.md)

| Tool | Purpose |
|------|---------|
| [`import-agronomy-seed-pack.sh`](import-agronomy-seed-pack.sh) | Promote **`gr33n-cultivator-seed-pack-v1`** + verify Postgres `catalog_version` / row counts |
| [`guardian-bootstrap-farm.sh`](guardian-bootstrap-farm.sh) | Field-guide + platform-doc + operational RAG ingest + readiness report |
| [`sample-cultivator-seed-pack-v1.body.json`](sample-cultivator-seed-pack-v1.body.json) | Commons catalog body mirror (DB contract, smoke prompts) |
| [`db/migrations/20260618_phase83_cultivator_seed_pack_v1.sql`](../../db/migrations/20260618_phase83_cultivator_seed_pack_v1.sql) | Publishes catalog slug `gr33n-cultivator-seed-pack-v1` |

```bash
make migrate
make check-crop-catalog-parity
make add-crop-check              # Phase 95 — pre-migrate YAML + seed drift (no DB)
make dev-auth-test   # optional for import POST

./scripts/enterprise/import-agronomy-seed-pack.sh --dry-run
./scripts/enterprise/import-agronomy-seed-pack.sh --farm-ids 1

./scripts/enterprise/guardian-bootstrap-farm.sh --dry-run --farm-id 1
./scripts/enterprise/guardian-bootstrap-farm.sh --farm-id 1   # needs EMBEDDING_API_KEY for ingest
# or:
make guardian-bootstrap-farm FARM_ID=1
```

**Site manifest:** set `guardian_seed.enabled: true` in YAML — `apply-site-manifest.sh` calls import + bootstrap after farm create. See [`site-manifest.example.yaml`](site-manifest.example.yaml).

**Depends on Phase 82 + Phase 84** (crop profiles + `crop_catalog_*` in Postgres). Cutover: [`docs/crop-catalog-db-cutover-runbook.md`](../../docs/crop-catalog-db-cutover-runbook.md). **Add crops:** [`docs/catalog-integrator-playbook.md`](../../docs/catalog-integrator-playbook.md) · PR checklist: [`docs/templates/add-crop-pr-checklist.md`](../../docs/templates/add-crop-pr-checklist.md). **Multi-site promote vs local (Phase 98):** [`docs/enterprise-catalog-promotion-model.md`](../../docs/enterprise-catalog-promotion-model.md).

### WS6 — farm crop override UI (Settings)

Farm owners/managers edit site-specific EC/VPD/DLI in **Settings → Crops & targets** without YAML. Overrides use the **same `crop_key`** as builtins (e.g. `cannabis`); Guardian `lookup_crop_targets` reads the farm row on the next chat turn.

| API | Role |
|-----|------|
| `GET /farms/{id}/crop-profiles/{crop_key}` | Effective profile + stages (member) |
| `PUT /farms/{id}/crop-profiles/{crop_key}` | Upsert override (farm admin) |
| `DELETE /farms/{id}/crop-profiles/{crop_key}` | Reset to builtin (farm admin) |

YAML alternative: [`apply-agronomy-overrides.sh`](apply-agronomy-overrides.sh) (WS2).

### WS5 — scheduled operational ingest

Example cron on the API host (`/etc/cron.d/gr33n-rag-ingest`):

```cron
0 */6 * * * gr33n cd /opt/gr33n-platform && set -a && . ./.env && set +a && \
  ./scripts/rag-ingest-farm-operational.sh --farm-id 1 --watermark-file /var/lib/gr33n/rag-watermark-farm-1
```

Or: `make rag-ingest-farm-operational FARM_ID=1`

## Shipped (Phase 83 WS2) — farm agronomy overrides

| Tool | Purpose |
|------|---------|
| [`apply-agronomy-overrides.sh`](apply-agronomy-overrides.sh) | Apply EC/VPD/DLI deltas from YAML onto farm crop profiles |
| [`data/agronomy-override-pack.example.yaml`](../../data/agronomy-override-pack.example.yaml) | Example override pack |

```bash
./scripts/enterprise/apply-agronomy-overrides.sh --dry-run --farm-id 1
./scripts/enterprise/apply-agronomy-overrides.sh --farm-id 1 --file data/agronomy-override-pack.example.yaml
# or with import:
./scripts/enterprise/import-agronomy-seed-pack.sh --farm-ids 1 --apply-overrides data/agronomy-override-pack.example.yaml
```

Unsupported catalog keys are rejected. Farm overrides win over builtins in `GetCropProfileByKey`.

## Contributing

If you build deployment pipeline tooling against the **public HTTP API**:

1. Prefer **config + scripts that call the API** over forking `cmd/api` unless you must patch core behavior.  
2. Open a **pull request** to this directory with a short README per tool (inputs, outputs, idempotency story).  
3. Do not commit secrets, `.env` files, or customer-specific hostnames.

## License note

gr33n platform code is **[AGPL v3](../../LICENSE)**. Ops scripts in this folder are intended as **operator tooling**; if your organization modifies the gr33n **application** itself and exposes it to users over a network, AGPL obligations apply to that software — consult counsel. Upstreaming deployment helpers here benefits everyone and avoids fork drift.

## Related

- [`docs/hypothetical-enterprise-topology.md`](../../docs/hypothetical-enterprise-topology.md)  
- [`docs/phase-14-operator-documentation.md`](../../docs/phase-14-operator-documentation.md#phase-31-field-validation-edge) — Phase 31 operator index  
- [`docs/plans/phase_33_guardian_polish_and_enterprise_ops.plan.md`](../../docs/plans/phase_33_guardian_polish_and_enterprise_ops.plan.md) — site manifest WS5  
- [`docs/plans/phase_83_enterprise_agronomy_seed_pack.plan.md`](../../docs/plans/phase_83_enterprise_agronomy_seed_pack.plan.md) — **shipped** — Guardian agronomy seed pack + bootstrap · [`phase-83-closure.md`](../../docs/plans/phase-83-closure.md)  
- [`docs/plans/phase_30_guardian_change_requests.plan.md`](../../docs/plans/phase_30_guardian_change_requests.plan.md)  
- [`docs/plans/phase_31_field_validation_and_edge.plan.md`](../../docs/plans/phase_31_field_validation_and_edge.plan.md)
