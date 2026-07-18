---
name: Phase 154 — test suite health
overview: >
  Fix `go test ./...` red/build-broken packages at repo root and add a
  `make test-unit` target that passes without live Postgres. cmd/api smoke
  tests remain behind `make test` (full DB harness).
todos:
  - id: ws1-compile-fixes
    content: "Fix handler/sensor mock drift and croplibrary nil-catalog panic"
    status: completed
  - id: ws2-unit-test-fixes
    content: "Fix stale unit tests (device auth, excerpt length, operator-tour needle, crop registry, cost guard env isolation)"
    status: completed
  - id: ws3-seed-drift
    content: "Regenerate db/seed/crop_catalog_from_yaml.sql from YAML"
    status: completed
  - id: ws4-test-unit-target
    content: "Add make test-unit excluding cmd/api DB smokes"
    status: completed
isProject: false
---

# Phase 154 — test suite health

**Status:** shipped

## Problem

Running `go test ./...` from repo root was red out of the box:

- `internal/handler/sensor` did not compile (mock `UpdateSensorConfig` signature drift)
- `internal/cropcycle` panicked when `CROP_CATALOG_SOURCE=db` without a runtime querier
- Several unit tests had stale expectations (auth, env, doc wording, catalog aliases)

`cmd/api` smoke tests require a migrated Postgres (`auth.users` etc.) — expected, but there was no documented fast path for contributors.

## Shipped

| WS | Change |
|----|--------|
| **WS1** | `loadDefaultCatalog()` returns error instead of calling `LoadCatalogFromDB(nil)`; sensor mock uses `db.UpdateSensorConfigParams` |
| **WS2** | Device handler tests use `PiEdgeAuth`; farmguardian `TestMain` sets `CROP_CATALOG_SOURCE=yaml`; cost guard tests isolate `GUARDIAN_COST_GUARD`; excerpt/platform/crop/ingest test expectations updated |
| **WS3** | Regenerated `db/seed/crop_catalog_from_yaml.sql` |
| **WS4** | `make test-unit` — all packages except `cmd/api`, no Postgres required |

## Operator commands

```bash
make test-unit    # fast — no DB smokes
make test         # full — includes cmd/api integration smokes (needs migrate + DATABASE_URL)
```

## Close when

- [x] `make test-unit` exits 0 on a fresh clone without Postgres running
- [x] `internal/handler/sensor` compiles and tests pass
- [x] `internal/cropcycle` no longer panics without DB querier
