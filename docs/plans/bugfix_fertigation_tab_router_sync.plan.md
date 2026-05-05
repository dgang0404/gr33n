# Bugfix — Fertigation tab bar vs router query

**Status: closed (2026-04-21)** — behaviour verified locally; append new symptoms under *Follow-ups* if anything regresses.

## Symptom

Clicking a tab (especially **Events**) sometimes appeared to do nothing: the visible panel did not match the highlighted tab, or deep links like `/fertigation?tab=events` did not stay in sync when switching tabs manually. In some cases the Events table threw at render time and the page looked stuck on **Loading…** or blank.

## Cause

1. **Router vs local state** — Tab state lived in local `activeTab` while many entry points use **`?tab=`** on the URL (`Dashboard`, `Inventory`, internal links). Without keeping URL and `activeTab` aligned, navigation and UI could diverge.
2. **`route.query.tab` shape** — Vue Router may expose `tab` as a string or string array; a watcher that only handled plain strings missed updates.
3. **Global loading hid all tabs** — While `loading` was true, tab panels were chained with `v-else-if` after the loading row, so only **Loading…** showed even when the real failure was elsewhere.
4. **Events row render crash** — `trigger_source` from the API is often a **nullable enum object** in JSON (`{ valid, gr33nfertigation_program_trigger_enum }`), not a plain string. Calling `.replace` on that value threw **Uncaught (in promise) TypeError**, which broke the Events list render.

## Fix (implemented)

- **`selectTab(id)`** — Sets `activeTab` immediately for responsive UI, then **`router.replace({ name: 'fertigation', query: { ...route.query, tab: id } })`** so the URL matches the panel; duplicate navigations are ignored in `.catch`.
- **`watch(() => [route.name, route.fullPath], …, { immediate: true })`** — Normalizes tab from **`tabQueryParam(route.query)`** (supports string | string[]); on `fertigation`, applies a valid `tab` to `activeTab`, otherwise defaults to **reservoirs**.
- **Tab panels** — Use **`v-if="activeTab === '…'"`** per panel instead of chaining off **`v-if="loading"`**, so tab content can render alongside the loading line when needed.
- **`refresh()`** — Guards missing **`farmId`**, watches **`farmContext.farmId`** for bootstrap timing, uses refresh generation + timeout so **`loading`** cannot wedge indefinitely.
- **`formatTriggerSource(raw)`** — Supports string or nullable enum object from the API before applying **`replace(/_/g, ' ')`**.

## Files

- `ui/src/views/Fertigation.vue`

## Follow-ups (append here)

_Add new bullets if a related regression appears (e.g. tab query lost on a specific navigation path)._

## Not deprecated

Fertigation Phase 20.x work is active; this is a **routing/state sync** and **Events rendering** fix, not a replacement for a future phase.
