# Phase 83 — closure (OC-83)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_83_enterprise_agronomy_seed_pack.plan.md`](phase_83_enterprise_agronomy_seed_pack.plan.md)

**Depends on:** Phase 82 crop library + field guides, Phase 84 Postgres crop catalog (`CROP_CATALOG_SOURCE=db` default).

---

## The one job (done)

> A new farm goes from migrate → Guardian-ready in one integrator command — structured crop targets, ingested field guides, operational RAG on a schedule, optional site-specific EC tweaks, and smokes that prove 8B + seed data works.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | Commons `gr33n-cultivator-seed-pack-v1` + migration | `TestPhase83CultivatorSeedPackPublished` |
| **WS2** | `apply-agronomy-overrides.sh` + example YAML | `internal/agronomyoverrides/*_test.go` |
| **WS3** | `guardian-bootstrap-farm.sh` + `make guardian-bootstrap-farm` | Script exit 0 + chunk report |
| **WS4** | `guardian_seed` in site manifest | `site-manifest.example.yaml` + `apply-site-manifest.sh` |
| **WS5** | Operational RAG cron doc + `rag-ingest-farm-operational.sh` | [`scripts/enterprise/README.md`](../../scripts/enterprise/README.md) |
| **WS6** | Settings **Crops & targets** + PUT/DELETE override API | `TestPhase83_CropProfileOverridePutDelete` |
| **WS7** | Readiness checklist + `smoke_phase83_test.go` | This doc + operator tour |
| **WS8** | Documentation sweep | Links below |

---

## Operator quick start

```bash
make migrate
make check-crop-catalog-parity

# Optional commons import audit (farm admin JWT):
./scripts/enterprise/import-agronomy-seed-pack.sh --farm-ids 1

# One-command Guardian bootstrap (needs EMBEDDING_API_KEY for ingest):
make guardian-bootstrap-farm FARM_ID=1

# Optional farm EC tweaks — YAML or Settings UI:
./scripts/enterprise/apply-agronomy-overrides.sh --farm-id 1 --file data/agronomy-override-pack.example.yaml
# Settings → Crops & targets (farm owner/manager)
```

**Site manifest:** set `guardian_seed.enabled: true` — see [`scripts/enterprise/site-manifest.example.yaml`](../../scripts/enterprise/site-manifest.example.yaml).

---

## Guardian readiness (manual)

- [ ] `AI_ENABLED=true`, Ollama probe OK ([`farm-guardian-ollama-setup.md`](../farm-guardian-ollama-setup.md))
- [ ] `EMBEDDING_API_KEY` set (or LAN embedder documented)
- [ ] `make guardian-bootstrap-farm FARM_ID=N` exit 0
- [ ] Field guide chunk count ≥ 12 for farm
- [ ] Ask: *Compare cannabis and eggplant EC targets* — chunks > 0 and/or tool block with **mS/cm**
- [ ] Ask: *How should I feed ramps?* — unsupported / cousin, not invented cannabis schedule
- [ ] Live plants: [`guardian-real-grow-readiness.md`](../guardian-real-grow-readiness.md)

---

## Automated tests

| Test | Path |
|------|------|
| Cultivator seed pack published | `cmd/api/smoke_phase83_test.go` — `TestPhase83CultivatorSeedPackPublished` |
| Crop override PUT/DELETE | `cmd/api/smoke_phase83_test.go` — `TestPhase83_CropProfileOverridePutDelete` |
| Farm override wins in DB | `internal/agronomyoverrides/apply_test.go` |

---

## Documentation index (WS8)

| Doc | Topic |
|-----|--------|
| [`scripts/enterprise/README.md`](../../scripts/enterprise/README.md) | Import, bootstrap, overrides, cron |
| [`farm-guardian-architecture.md` §7.0ae](../farm-guardian-architecture.md#70ae-enterprise-agronomy-bootstrap-phase-83--shipped) | Architecture |
| [`operator-tour.md` §6o](../operator-tour.md#6o-enterprise-agronomy-bootstrap-phase-83--shipped) | Operator walkthrough |
| [`guardian-real-grow-readiness.md`](../guardian-real-grow-readiness.md) | Live-plant checklist |
| [`commons-catalog-operator-playbook.md`](../commons-catalog-operator-playbook.md) | Agronomy pack kind |
| [`hypothetical-enterprise-topology.md`](../hypothetical-enterprise-topology.md) | Warehouse bring-up |
| [`crop-catalog-db-cutover-runbook.md`](../crop-catalog-db-cutover-runbook.md) | DB catalog cutover |

---

## OC-83

Phase 83 is **closed** when migrate + parity check + bootstrap + override path + smokes pass on a seeded farm. Enterprise integrators use [`scripts/enterprise/`](../../scripts/enterprise/README.md); agronomists use **Settings → Crops & targets** for same-key overrides without YAML.
