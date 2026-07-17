---
name: Phase 203 — Handler package consolidation (catalog + natural farming)
overview: >
  Redundancy audit found adjacent HTTP packages with duplicated helpers and
  split domains: recipe/naturalfarming, commonscatalog/commonscropcatalog,
  fieldguides vs agronomy-field-guides. Packages work; clutter is navigation and
  copy-paste (numericFromFloat64, path ID parsers, limit/offset). This phase
  consolidates helpers first, then merges packages only where routes and OpenAPI
  stay stable.
todos:
  - id: ws1-shared-helpers
    content: "WS1: httputil — ParsePathInt, ParseLimitOffset; single numericFromFloat64 in httputil or internal/conv; delete local copies in cost/cropcycle/fertigation/recipe + farmguardian/tools/args.go"
    status: pending
  - id: ws2-path-id-style
    content: "WS2: migrate inline strconv.ParseInt(r.PathValue) to httputil.PathID where path shape matches; delete dead pathSegment/idSegment in devicecmd/handler.go"
    status: pending
  - id: ws3-dead-exports
    content: "WS3: remove unused commontypes enums (ValidationRule*, UserActionTypeEnum), plantcatalog.ResolveCropKeyFromProfile if still unreferenced, farmbootstrap.IsKnownTemplate if internal-only"
    status: pending
  - id: ws4-naturalfarming-merge
    content: "WS4: merge handler/recipe into handler/naturalfarming (or rename package naturalfarming → recipes); single routes.go block; no URL changes"
    status: pending
  - id: ws5-commons-catalog-merge
    content: "WS5: merge handler/commonscropcatalog + handler/fieldguides agronomy overlap into handler/commonscatalog OR document sub-routers in one package; unify /commons/* registration in routes.go"
    status: pending
  - id: ws6-tests-openapi
    content: "WS6: cmd/api smoke tests pass; openapi.yaml tag descriptions updated; phase-203-closure.test.js (helper single-source grep)"
    status: pending
isProject: false
---

# Phase 203 — Handler package consolidation

**Status:** planned · **Depends on:** none (backend janitorial)

## The problem

### Duplicate helpers (high confidence)

| Pattern | Locations |
|---------|-----------|
| `numericFromFloat64` | `handler/cost`, `cropcycle`, `fertigation`, `recipe`, `farmguardian/tools/args.go` |
| `parseDate` (pgtype.Date) | `cost/handler.go`, `cropcycle/handler.go` |
| limit/offset query clamp | `alert`, `audit`, `commonscatalog`, `cost`, `chat/proposals`, `farm/insert_commons` |
| `farmIDFromPath` / `resourceIDFromPath` | `fertigation`, `lighting`, `naturalfarming`, `guardian/parse.go` |
| Dead `pathSegment` / `idSegment` | `devicecmd/handler.go` (zero callers) |

~20 packages use `httputil.PathID`; ~22 roll their own `ParseInt(r.PathValue("id"))` — style drift, not always wrong.

### Adjacent packages (merge candidates)

| Pair | Overlap |
|------|---------|
| `handler/recipe` + `handler/naturalfarming` | Same `/farms/{id}/naturalfarming/...` tree in routes.go |
| `handler/commonscatalog` + `handler/commonscropcatalog` | Both `/commons/*` |
| `handler/fieldguides` + crop catalog agronomy | Static `/v1/field-guides` vs DB `/commons/agronomy-field-guides` — keep two data sources, maybe one package |
| `handler/chat` + `handler/guardian` | guardian/ is nudge + reingest only (~3 files) — candidate to fold into chat/ |

### Dead exports

- `commontypes.ValidationRuleTypeEnum`, `ValidationSeverityEnum`, `UserActionTypeEnum` — table dropped phase 115
- `db.Gr33ncoreValidationRule` types — no queries
- `plantcatalog.ResolveCropKeyFromProfile` — only self-reference

## What to ship

### WS1 — Shared helpers (do first, lowest risk)

Add to `internal/httputil/`:

```go
func ParseLimitOffset(r *http.Request, defaultLimit, maxLimit int) (limit, offset int)
func NumericFromFloat64(v float64) (pgtype.Numeric, error) // or existing pg helper
```

Replace copies; run `go test ./internal/handler/...`.

### WS2 — Path ID consistency

Use `httputil.PathID` for `{id}`-style paths documented in OpenAPI.
Keep bespoke parsers only when path has multiple segments (document why).

Delete dead code in devicecmd.

### WS3 — Dead export cleanup

Remove unused enum types after grep confirms zero references (including sqlc).

### WS4 — naturalfarming + recipe

**Preferred:** single package `naturalfarming` with `recipe.go`, `inputs.go` files.
**Routes:** unchanged paths — internal move only.
**Tests:** update import paths.

### WS5 — commons catalog surface

**Option A:** one `commonscatalog` package, subfiles `crop_catalog.go`, `field_guides.go`.
**Option B:** keep packages, single `registerCommonsRoutes(mux)` in routes.go with comment block.

Do **not** merge field_guide RAG chunks with agronomy DB in this phase — HTTP only.

### WS6 — Verification

- `go test ./...`
- `grep -r numericFromFloat64 internal/` → one definition
- OpenAPI tag **commons** still accurate

## Acceptance criteria

- [ ] Single `NumericFromFloat64` (or equivalent) in codebase
- [ ] `ParseLimitOffset` used by ≥4 handlers that duplicated logic
- [ ] recipe package merged or explicitly documented as subfolder of naturalfarming
- [ ] commons HTTP registration readable in one routes.go section
- [ ] phase-203-closure.test.js
- [ ] No API URL changes (breaking)

## Out of scope

- Merging chat + rag packages
- Rewriting migrations (bootstrap function REPLACE noise is historical)
- Frontend changes

## Ponytail note

**WS1–WS3 before WS4–WS5.** Helper dedup is the biggest win per line changed; package merges only when you're already editing those files.
