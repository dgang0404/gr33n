---
name: Phase 211.03 — Farm permissions (API-first, granular caps)
overview: >
  Replace coarse role-only checks and page-layout “access control” with a single
  farm-scoped capability model resolved on every API mutation. Expose caps to the
  UI for disabling actions only. Unblocks removing the redundant Natural farming
  Inputs & batches tab once edit/delete live in context.
todos:
  - id: ws1-scope-catalog
    content: "WS1: Define scope catalog + role templates; document merge rules for farm_memberships.permissions JSONB"
    status: completed
  - id: ws2-caps-resolver
    content: "WS2: Extend farmauthz — ResolveFarmScopes / FarmCapsForUser merges role template + permissions overrides"
    status: completed
  - id: ws3-caps-api
    content: "WS3: GET /farms/{id}/me/caps — current user scopes for selected farm (UI affordances only)"
    status: completed
  - id: ws4-nf-handlers
    content: "WS4: Natural farming handlers — split RequireFarmOperate into nf.* scopes (read vs write vs delete vs pack.apply)"
    status: completed
  - id: ws5-cost-handlers
    content: "WS5: Money/cost paths — align RequireCostRead/Write with money.costs.* scopes; batch unit-cost PATCH"
    status: completed
  - id: ws6-ui-composable
    content: "WS6: useFarmCaps composable; replace duplicated OPERATE_ROLES sets; disable destructive buttons (fail closed on load error)"
    status: completed
  - id: ws7-settings-custom-role
    content: "WS7: Settings — custom_role scope picker for owner/manager (minimal JSON editor or checkbox grid)"
    status: completed
  - id: ws8-nf-tab-removal
    content: "WS8 (after WS1–WS6): Remove Inputs & batches tab; inline edit/delete on Supplies + Apply recipes + Make a batch"
    status: completed
  - id: ws9-tests-closure
    content: "WS9: farmauthz unit tests, smoke caps endpoint, UI closure — viewer 403 on delete, finance blocked on nf.delete"
    status: completed
isProject: false
---

# Phase 211.03 — Farm permissions (API-first, granular caps)

**Status:** Complete (WS1–WS9 shipped) · **Depends on:** [211.01](phase_211_01_nf_studio_declutter.plan.md) · **Before:** [212 dual-install federation](phase_212_dual_farm_federation_test.plan.md)

## The one job

> **Who can do what** is decided by the **API on every write**, not by which
> workspace tab exists. The UI reads caps to disable buttons; hiding a page
> is never security.

## Problem (today)

| Layer | Behavior | Gap |
|-------|----------|-----|
| API | `RequireFarmOperate` / `RequireFarmAdmin` / `RequireCostRead` | **Coarse** — operator can delete inputs same as owner |
| DB | `farm_memberships.permissions` JSONB | **Unused** — `capsForRole()` ignores it |
| `custom_role` enum | Exists | Resolves to **zero caps** — broken default |
| UI | `useFarmOperate` duplicates role list; **fails open** on error | Not aligned with backend; no granular disable |
| NF UX | **Inputs & batches** tab = admin CRUD grid | Redundant with Make a batch / Apply recipes / Money supplies; tempts “security by tab” |

**Symptom:** Product wants many operator levels (finance restocks costs but cannot delete recipes; worker logs mixes but cannot edit input definitions). Adding/removing tabs does not solve this and creates duplicate edit surfaces.

## Principles (non-negotiable)

1. **API is the trust boundary** — every mutating handler calls `RequireFarmScope(...)` (or equivalent). UI hiding is optional polish.
2. **Read stays member-level** — any farm member may GET lists needed to do their job unless a row is cost-sensitive (existing cost read gate).
3. **Scopes are strings, not pages** — `nf.batches.write`, not `natural-farming.manage`.
4. **Role = template, permissions = override** — preset roles stay for onboarding; `custom_role` + JSONB overrides for granular farms.
5. **Default deny for unknown scopes** — new handlers must opt into a scope explicitly.
6. **No WS8 until WS1–WS6 ship** — do not remove **Inputs & batches** or scatter inline delete until caps API and NF handler splits exist.

## Scope catalog (WS1)

Stable scope ids (additive; do not rename once shipped):

| Scope | Typical use | Replaces / narrows |
|-------|-------------|-------------------|
| `farm.member` | Implicit — any membership | `RequireFarmMember` |
| `farm.admin` | Membership, modules, bootstrap, pack import admin | `RequireFarmAdmin` |
| `money.costs.read` | Ledger, unit costs, supplies economics | `RequireCostRead` |
| `money.costs.write` | Receipts, unit cost PATCH, restock cost fields | `RequireCostWrite` |
| `nf.read` | List inputs, batches, recipes | member GET (no extra gate) |
| `nf.inputs.write` | Create/update input_definitions | part of `Operate` today |
| `nf.inputs.delete` | DELETE input_definitions | part of `Operate` today |
| `nf.batches.write` | Create/update batches (status, qty, threshold) | part of `Operate` today |
| `nf.batches.delete` | DELETE batches | part of `Operate` today |
| `nf.recipes.write` | Create/update application_recipes + components | part of `Operate` today |
| `nf.recipes.delete` | DELETE application_recipes | part of `Operate` today |
| `nf.pack.apply` | Commons / switchover pack apply | `RequireFarmAdmin` on apply_pack today |
| `farm.operate` | Zones, tasks, sensors, actuators, fertigation, mixing, programs | `RequireFarmOperate` on non-NF handlers |

**Role templates** (defaults when `permissions` empty or `{}`):

| Role | Scopes (summary) |
|------|------------------|
| `owner`, `manager` | All scopes |
| `operator`, `worker`, `agronomist` | `farm.operate`, all `nf.*` **except** `nf.inputs.delete`, `nf.batches.delete`, `nf.pack.apply` |
| `finance` | `money.costs.read`, `money.costs.write`, `nf.read`, `nf.batches.write` (restock qty only — no recipe delete) |
| `viewer` | `nf.read` only (+ member reads elsewhere as today) |
| `custom_role` | **`permissions.scopes` only** — no implicit grants |

### `permissions` JSONB shape

```json
{
  "scopes": ["nf.batches.write", "money.costs.read"],
  "deny": ["nf.recipes.delete"]
}
```

Merge order: **role template → add `scopes` → subtract `deny`**. Owner always full caps (ignore overrides except audit log).

## WS2 — Resolver (`internal/farmauthz`)

Extend `FarmCaps` or add parallel `FarmScopes map[string]bool` — prefer **explicit scope set** over growing bool struct:

```go
func ResolveFarmScopes(ctx, q, userID, farmID) (map[string]bool, error)
func RequireFarmScope(w, r, q, farmID, scope string) bool
```

- Keep `RequireFarmOperate` as alias → `farm.operate` during migration (deprecate in comments).
- `FarmCapsForUser` returns `{ scopes: [...], legacy: FarmCaps }` for one release if needed.
- Wire `custom_role` through `permissions.scopes`; fix current “empty caps” bug.

## WS3 — Caps API

```
GET /farms/{farm_id}/me/caps
→ { "role_in_farm": "operator", "scopes": ["farm.operate", "nf.read", ...] }
```

- Member-only; 403 if not on farm.
- Cached in UI per farm session (invalidate on role change / farm switch).
- **Not** a substitute for enforcement — handlers still check scopes.

## WS4 — Natural farming handlers

| Route / action | Scope |
|----------------|-------|
| GET inputs/batches/recipes | member (+ cost fields gated separately if ever needed) |
| POST/PATCH input_definitions | `nf.inputs.write` |
| DELETE input_definitions | `nf.inputs.delete` |
| POST/PATCH input_batches | `nf.batches.write` |
| DELETE input_batches | `nf.batches.delete` |
| POST/PATCH recipes / components | `nf.recipes.write` |
| DELETE recipes | `nf.recipes.delete` |
| POST apply-pack | `nf.pack.apply` |

Files: `internal/handler/naturalfarming/handler.go`, `recipe.go`, `apply_pack.go`.

## WS5 — Money / costs

- Supplies restock (`current_quantity_remaining`) → `nf.batches.write` (operator path).
- Unit cost on input definition → `money.costs.write` (finance path).
- Ledger CRUD → existing cost scopes (rename internally to `money.costs.*`).

Ensures **finance can restock and edit costs without recipe delete**.

## WS6 — UI composable

Replace ad-hoc role sets with:

```js
// useFarmCaps(farmId) → { scopes, has(scope), loading, refresh }
```

- **Fail closed** on caps load error (inverse of today’s `useFarmOperate`).
- Guardian Confirm: `has('farm.operate')` (unchanged behavior for operators).
- Supplies: disable Delete / Advanced edit without `nf.inputs.write` / `nf.batches.delete`.
- Apply recipes: disable delete without `nf.recipes.delete`.
- **Do not** remove sidebar tabs based on caps in this phase (optional later).

## WS7 — Settings (minimal)

- Owner/manager editing a `custom_role` member: checkbox grid of scopes (or raw JSON advanced).
- Preset roles remain dropdown — no per-scope editing for presets in v1 (ponytail: templates only).

## WS8 — NF tab removal (only after WS1–WS6)

Remove **Inputs & batches** (`manage` tab) from Natural farming workspace:

| Action | New home |
|--------|----------|
| Create input + batch (guided) | **Make a batch** (unchanged) |
| Apply recipe CRUD | **Apply recipes** (unchanged) |
| Restock qty, unit cost, quick new batch | **Money → Supplies on hand** (unchanged) |
| Edit batch metadata, delete batch/input | Inline on Supplies / row overflow — **disabled** without scope |
| Full row editor footer link | Remove or → contextual edit on Supplies |

Redirects: `/natural-farming?tab=manage&…` → Supplies or Apply recipes by `inv` query (same as today’s money inventory redirects).

**Delete `FarmRowsPanel.vue`** once inline paths exist.

## WS9 — Acceptance

- Viewer: GET NF lists OK; PATCH/DELETE → **403** with stable error body.
- Finance: restock + unit cost OK; DELETE recipe → **403**.
- Operator: create batch + recipe OK; DELETE input → **403** (unless template changed).
- `custom_role` with `{ "scopes": ["nf.recipes.write"] }` can PATCH recipe, cannot DELETE.
- UI disables matching buttons; bypassing UI still 403.
- `go test ./internal/farmauthz/...` + NF handler tests + `npm test` caps composable closure.
- **No regression:** owner/manager behavior matches today’s full access.

## Out of scope (211.03)

- Per-zone or per-resource ACL (farm-wide only).
- Org-level roles across multiple farms.
- Audit log UI for permission changes.
- Hiding entire workspaces from sidebar by role (future UX; not security).
- Crop ops report UI — see [211.04](phase_211_04_crop_ops_report_ui.plan.md) (renamed from 211.02 footnote).

## File touch list

| Layer | Files |
|-------|--------|
| Auth | `internal/farmauthz/capabilities.go`, `scopes.go` (new), tests |
| API | `internal/handler/farm/caps.go` (new), route in `cmd/api/routes.go` |
| NF | `internal/handler/naturalfarming/*.go` |
| Cost | `internal/handler/cost/*`, NF batch cost PATCH paths |
| UI | `ui/src/composables/useFarmCaps.js` (new), retire/ wrap `useFarmOperate.js`, `SuppliesHub.vue`, `RecipesApplyPanel.vue`, `MakeBatchPanel.vue`, `Settings.vue` |
| NF declutter | `ui/src/lib/workspaces.js`, `NaturalFarmingWorkspace.vue`, delete `FarmRowsPanel.vue`, redirect helpers in `workspaceRoutes.js` |
| Docs | This plan; update `docs/farm-guardian-architecture.md` § Confirm caps |

## Sequencing

```
211.03 WS1–WS7  (API + caps endpoint + UI disable)
       ↓
211.03 WS8      (remove Inputs & batches tab — safe now)
       ↓
211.02 / 211.04 / 212 as planned
```

**Do not start WS8** until a viewer integration test proves DELETE is 403 and Supplies buttons respect caps.

## Related

- [211.01 NF declutter](phase_211_01_nf_studio_declutter.plan.md) — vocabulary, Money supplies home, nav-hint fixes.
- [211.02 recipe revisions](phase_211_02_recipe_formula_history.plan.md) — revision writes should record `created_by_user_id`; same auth model.
- [212 dual-install federation](phase_212_dual_farm_federation_test.plan.md) — run auth matrix on Install A after 211.03 WS9.
