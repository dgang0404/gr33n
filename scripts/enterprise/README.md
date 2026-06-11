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

## Planned (Phase 83 — agronomy seed pack)

**Plan:** [`docs/plans/phase_83_enterprise_agronomy_seed_pack.plan.md`](../../docs/plans/phase_83_enterprise_agronomy_seed_pack.plan.md)

| Tool (planned) | Purpose |
|----------------|---------|
| `import-agronomy-seed-pack.sh` | Promote **`gr33n-cultivator-seed-pack-v1`** to farm(s) via commons catalog |
| `guardian-bootstrap-farm.sh` | Field-guide + platform-doc + operational RAG ingest + readiness report |
| `apply-agronomy-overrides.sh` | Farm-specific EC/VPD/DLI overrides (YAML) on top of Phase 82 builtins |
| `site-manifest.yaml` → `guardian_seed:` | Extend site bring-up to call bootstrap after zones + recipe pack |

**Depends on Phase 82** (`crop_library.yaml`, field guides, zero-chunk guardrail). Run readiness smokes on **8B + full seed** before upgrading to 70B — same pipelines, better synthesis.

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
- [`docs/plans/phase_33_guardian_polish_and_enterprise_ops.plan.md`](../../docs/plans/phase_33_guardian_polish_and_enterprise_ops.plan.md) — site manifest WS5 (planned)  
- [`docs/plans/phase_83_enterprise_agronomy_seed_pack.plan.md`](../../docs/plans/phase_83_enterprise_agronomy_seed_pack.plan.md) — Guardian agronomy seed pack + bootstrap (planned)  
- [`docs/plans/phase_30_guardian_change_requests.plan.md`](../../docs/plans/phase_30_guardian_change_requests.plan.md)  
- [`docs/plans/phase_31_field_validation_and_edge.plan.md`](../../docs/plans/phase_31_field_validation_and_edge.plan.md)
