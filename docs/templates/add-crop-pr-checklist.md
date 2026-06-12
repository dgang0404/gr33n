# PR checklist — add / update platform crop

Use this in PR descriptions when changing `data/crop_library.yaml`.

**Example crop:** San Pedro cactus (`san_pedro`) — already shipped; copy this template for the next addition.

---

## Summary

- [ ] Crop `crop_key`: `____________`
- [ ] Display name: `____________`
- [ ] Category: `flower` / `fruiting` / `leafy` / `herb` / `epiphyte` / …
- [ ] Field guide: `docs/field-guides/crop-____________-nutrition.md`

---

## Platform team (git)

- [ ] `data/crop_library.yaml` — crop block + stages (EC in **mS/cm** only)
- [ ] Top-level or per-crop `aliases:` for Guardian resolution
- [ ] **`version:` bumped** in YAML header (monotonic)
- [ ] Field guide MD + manifest entry (`docs/rag/field-guide-manifest.yaml`)
- [ ] `./scripts/generate-crop-catalog-seed.sql.sh -o db/seed/crop_catalog_from_yaml.sql`
- [ ] `./scripts/generate-crop-catalog-seed.sql.sh -o db/migrations/YYYYMMDD_catalog_<slug>.sql`
- [ ] `make add-crop-check` passes locally
- [ ] If enterprise pack pins version: update `platform_catalog_version` in agronomy seed pack JSON + migration

---

## Site integrator (after merge)

- [ ] `make migrate` on each environment
- [ ] `make check-catalog-release`
- [ ] `make rag-ingest-field-guides` (if guide body changed)
- [ ] Restart API (`CROP_CATALOG_SOURCE=db`)
- [ ] Smoke: `GET /commons/crop-catalog/{crop_key}` + picker contains new crop
- [ ] Optional: `guardian-bootstrap-farm.sh` on enterprise sites after major bump

---

## Test plan

```bash
make add-crop-check
make migrate && make check-catalog-release
go test -tags dev ./cmd/api/ -run 'TestPhase95|TestPhase107' -count=1
```

---

## Docs

- [Catalog integrator playbook](../catalog-integrator-playbook.md)
