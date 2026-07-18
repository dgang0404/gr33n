# Enterprise catalog promotion model (Phase 98)

**Audience:** Multi-site integrators, HQ platform team, org admins.

**Closes blind spot #9:** What promotes **platform-wide** vs stays **farm-local** when one org runs many sites.

**Closure:** [Phase 98](plans/archive/phase_98_enterprise_catalog_promotion.plan.md) · [`phase-98-closure.md`](plans/archive/phase-98-closure.md)

---

## Mental model

```
HQ platform team                Each site integrator
─────────────────               ────────────────────
crop_library.yaml               make migrate
       │                               │
       ▼                               ▼
catalog seed migration ────────► same Postgres rows on every site
(platform catalog)              (crop_catalog_entries, builtins)

Commons agronomy pack ──optional──► import-agronomy-seed-pack.sh (audit only)

Farm A Settings EC ─────────────► farm_id scoped — never copies to Farm B

YAML override pack ─────────────► apply-agronomy-overrides.sh per farm_id
```

**Rule:** Copying Farm A's override YAML to Farm B is a **local** operation — it does **not** update the platform catalog.

---

## Promote vs local matrix

| Artifact | Scope | How it moves | Auto-promotes to other farms? |
|----------|-------|--------------|-------------------------------|
| `crop_catalog_entries` + builtin `crop_profiles` | **Platform** | SQL migration on every site | ✅ same rows after migrate |
| `agronomy_field_guides` (Postgres) | **Platform** | Catalog seed migration | ✅ same after migrate |
| Commons **agronomy seed pack** | **Org optional** | `import-agronomy-seed-pack.sh` per farm | ❌ audit record only |
| Commons **recipe pack** | **Org optional** | `import-recipe-pack.sh` per farm | ❌ programs by name idempotency |
| Farm EC override (`PUT crop-profiles/{crop_key}`) | **Single farm** | Settings UI or override YAML | ❌ never |
| Genetics EC profile (Phase 94) | **Single farm** | Genetics API / Plants link | ❌ never |
| `plants.crop_key` slots | **Single farm** | Plants / Start grow | ❌ never |
| Field guide **RAG chunks** | **Per farm** | `guardian-bootstrap-farm.sh` | ❌ per `farm_id` ingest |
| Fertigation programs | **Single farm** | UI or recipe pack import | ❌ unless script copies |

---

## HQ catalog release (all sites)

1. Platform team ships YAML + seed migration ([catalog integrator playbook](catalog-integrator-playbook.md)).
2. **Every site** runs `make migrate` — bumps `crop_catalog_entries.catalog_version`.
3. Phase 109 notifies farm admins (`catalog_version_bump` alerts) when API restarts.
4. Each site optionally runs `make rag-ingest-field-guides` if guide bodies changed ([Phase 97](plans/archive/phase_97_rag_structured_truth_governance.plan.md)).
5. Farm EC overrides **remain** — structured profiles still win over stale RAG.

**Pin in site manifest:** `platform.catalog_version_min` must match post-migrate DB (see [`site-manifest.example.yaml`](../scripts/enterprise/site-manifest.example.yaml)).

---

## Org commons import (optional per farm)

| Pack | What import does | What import does **not** do |
|------|------------------|----------------------------|
| `gr33n-cultivator-seed-pack-v1` | Records audit + verifies `platform_catalog_version` | Run migrations, RAG ingest, EC overrides |
| `gr33n-recipe-pack-v7-lettuce-veg` | Creates inactive fertigation programs | Enable programs or copy to other farms |

Follow import with per-farm bootstrap:

```bash
./scripts/enterprise/import-agronomy-seed-pack.sh --farm-ids 1
./scripts/enterprise/guardian-bootstrap-farm.sh --farm-id 1
```

---

## Farm-local overrides (never platform)

| Method | Tool |
|--------|------|
| Settings → Crops & targets | UI `PUT /farms/{id}/crop-profiles/{crop_key}` |
| Batch YAML | [`apply-agronomy-overrides.sh`](../scripts/enterprise/apply-agronomy-overrides.sh) |
| Per-variety EC | Phase 94 genetics API |

**Anti-pattern:** Exporting Farm A's override JSON and expecting Farm B + all sites to pick it up without explicit per-farm apply.

---

## Site manifest fields (Phase 98)

```yaml
platform:
  catalog_version_min: 4      # MIN crop_catalog_entries.catalog_version after migrate
  catalog_source: db          # CROP_CATALOG_SOURCE default

farm_local:
  agronomy_override_pack: null   # path to YAML — applied only to this farm
```

See [`apply-site-manifest.sh`](../scripts/enterprise/apply-site-manifest.sh) — manifest documents expectations; integrator runs migrate + optional override apply separately.

---

## Verification

```bash
make migrate
make check-catalog-release
go test -tags dev ./cmd/api/ -run TestPhase98 -count=1
```

Manual two-farm check: override cannabis on Farm A → Farm B effective profile unchanged (builtin EC).

---

## Related

- [Commons catalog playbook](commons-catalog-operator-playbook.md)
- [Catalog integrator playbook](catalog-integrator-playbook.md)
- [Hypothetical enterprise topology](hypothetical-enterprise-topology.md)
- [Enterprise README](../scripts/enterprise/README.md)
- [Crop knowledge operator runbook](crop-knowledge-operator-runbook.md)
