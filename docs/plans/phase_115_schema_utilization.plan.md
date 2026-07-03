---
name: Phase 115 — Schema utilization (surface what the DB already supports)
overview: >
  The July 2026 audit cross-referenced all 69 tables in gr33n-schema-v2-FINAL.sql
  against sqlc queries, routes, and UI. Most are fully wired; this phase productizes
  the ones that aren't: farm module gating, notification templates, system logs,
  the symptom catalog, stage-event timelines, task duration fields, and alert
  delivery status. Kill-or-keep decisions for dead tables are part of the phase.
todos:
  - id: ws1-active-modules
    content: "WS1: farm_active_modules — API CRUD + Settings toggles per domain (animals, aquaponics, naturalfarming…); nav/workspace links gated on is_enabled; default rows seeded on farm create"
    status: pending
  - id: ws2-notification-templates
    content: "WS2: notification_templates — GET/POST/PATCH routes; template picker in RuleForm + fertigation notify actions (replace raw numeric template ID inputs)"
    status: pending
  - id: ws3-system-logs
    content: "WS3: system_logs — wire worker/handler WARN+ERROR events to INSERT (negative stock, currency mismatch, command failures); read-only Diagnostics panel in Settings with severity filter"
    status: pending
  - id: ws4-symptom-catalog
    content: "WS4: agronomy_symptom_entries — GET /commons/agronomy-symptoms route; browsable symptom guide page (filter by crop/category); link from Guardian citations"
    status: pending
  - id: ws5-cycle-timeline
    content: "WS5: crop_cycle_stage_events — visible stage-transition timeline on CropCycleSummary (who/when/auto-vs-manual)"
    status: pending
  - id: ws6-task-fields
    content: "WS6: tasks — expose estimated_duration_minutes + actual start/end in Tasks.vue (est. duration on create; timer or manual times on complete); duration in labor cost rollups"
    status: pending
  - id: ws7-alert-delivery
    content: "WS7: alerts_notifications — show delivery status (scheduled_send_at, delivery_attempts, channel results) on Alerts page for push/email debugging"
    status: pending
  - id: ws8-kill-or-keep
    content: "WS8: validation_rules kill-or-keep — no reader exists; either drop table + schema comment or write one-page design for evaluator; document decision in plan closure"
    status: pending
isProject: false
---

# Phase 115 — Schema utilization (surface what the DB already supports)

## Status

**Planned.** From the July 2026 audit (schema/UI workstream). Sibling phases:
[113](phase_113_security_hardening.plan.md) security,
[114](phase_114_pi_edge_integrity.plan.md) Pi chain,
[116](phase_116_docs_refresh.plan.md) docs, [117](phase_117_test_depth.plan.md) tests.

---

## Audit snapshot

69 tables across 7 schemas; ~66 have sqlc queries; ~60+ have routes. Gaps by level:

| Level | Tables |
|-------|--------|
| **None** (schema only) | `farm_active_modules`, `system_logs`, `validation_rules` |
| **Query-only** (no REST) | `notification_templates`, `agronomy_symptom_entries` (+ internal pipeline tables that are fine as-is: `insert_commons_received_payloads`, `platform_catalog_state`, `session_summaries`, `cost_transaction_idempotency`) |
| **API-only** (no UI) | fertigation `mix-jobs` (UI hook ships in Phase 114 WS4), `commons` catalog browse, `crop_cycle_stage_events` timeline, org usage summary |
| **Partial** | `tasks` duration fields, `alerts_notifications` delivery columns, `farms`/`zones` extended attributes, `weather_data` forecast columns |

Column-level: most `meta_data` JSONB columns are write-rarely/read-never — acceptable
extensibility headroom, no action. `devices.api_key` legacy column deprecation is
covered by Phase 113 WS7.

---

## Design notes

### WS1 — Module gating

Today every install shows Animals/Aquaponics/Natural-Farming nav regardless of use.
`farm_active_modules` was designed exactly for this. Seed one row per known module on
farm create (naturalfarming + crops on, animals/aquaponics off by default); Settings
gets a "Farm modules" card with toggles (owner/manager). SPA hides workspace nav +
routes for disabled modules; API returns 403 with "module disabled for this farm" so
direct calls fail loudly, not silently.

### WS3 — System logs

Insertion points already exist as `slog.Warn/Error` calls in the worker and costing
paths; add a thin `systemlog.Submit(ctx, severity, source, message, contextData)`
mirroring the `auditlog` pattern. Retention: same hypertable policy as sensor data.
Diagnostics panel is read-only, owner/manager, newest-first, severity filter — this is
the "why did my schedule not fire at 6am" answer page for non-IT operators.

### WS8 — validation_rules

No code reads this table. Carrying dead schema costs comprehension every audit.
Decide: (a) drop it via migration with a schema comment pointing at the decision, or
(b) keep with a concrete one-page evaluator design and a target phase. Default
recommendation: **drop**; per-field validation is better handled in handlers/UI.

### Out of scope

- Zone hierarchy (`parent_zone_id`) and GIS boundary editor — real feature, needs its own phase with map tooling
- Weather forecast → automation preconditions — pairs with a future automation phase
- Org-level dashboard — enterprise tier concern

---

## Acceptance

- [ ] Disabling Animals module hides nav + returns 403 on animal routes for that farm; re-enable restores
- [ ] Rule/fertigation notify actions use a template picker; template CRUD via API; no raw ID inputs left
- [ ] Forcing a worker warning (e.g. failed command) produces a visible row in Settings → Diagnostics
- [ ] Symptom guide page lists catalog entries with crop/category filters; deep-linkable from Guardian answers
- [ ] CropCycleSummary shows stage timeline with actor + timestamp per transition
- [ ] Task create accepts estimated duration; completing with actual times feeds labor rollups
- [ ] Alerts page shows per-channel delivery attempts for a push-enabled alert
- [ ] validation_rules decision executed and documented (migration or design doc)
- [ ] openapi.yaml + smoke tests updated for every new route

---

## Files expected to change

| Area | Files |
|------|-------|
| Queries | `db/queries/` (farm_active_modules, notification_templates, system_logs, symptoms) |
| Handlers/routes | `internal/handler/farm/*`, `internal/handler/alert/*`, `cmd/api/routes.go` |
| Worker | `internal/automation/*` (system log submits) |
| UI | Settings modules card, RuleForm picker, Diagnostics panel, symptom guide view, CropCycleSummary, Tasks.vue, Alerts view |
| Schema | migration for seeded module rows; possible `validation_rules` drop |
| Docs/tests | `openapi.yaml`, `cmd/api/smoke_phase115_*.go` |
