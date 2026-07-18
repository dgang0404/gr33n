# Phase 105 — closure (OC-105)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_105_catalog_audit_oc84_closure.plan.md`](phase_105_catalog_audit_oc84_closure.plan.md)

**Depends on:** [Phase 83](phase_83_enterprise_agronomy_seed_pack.plan.md) farm overrides; [Phase 84](phase_84_crop_catalog_enterprise_db.plan.md) catalog DB.

**Also closes:** [Phase 84](phase_84_crop_catalog_enterprise_db.plan.md) formal artifact — [`phase-84-closure.md`](phase-84-closure.md) (**OC-84**).

---

## The one job (done)

> **Farm crop override changes are auditable** — PUT/DELETE on `/farms/{id}/crop-profiles/{crop_key}` writes to the farm audit trail with `crop_key` and `catalog_version`; Phase 84 has a formal closure doc like Phase 83.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | OC-84 closure doc | [`phase-84-closure.md`](phase-84-closure.md) |
| **WS2** | Audit on override PUT/DELETE | `internal/handler/cropprofile/override.go` → `user_activity_log` |
| **WS3** | Settings “Last changed” column | `CropTargetsSettings.vue` — `updated_at` on override rows |
| **WS4** | Integrator / compliance runbook | [`audit-events-operator-playbook.md`](../audit-events-operator-playbook.md) § crop override export |

---

## Audit event kinds

| Kind | Trigger | Payload |
|------|---------|---------|
| `crop_profile_override_upsert` | `PUT …/crop-profiles/{crop_key}` | `crop_key`, `catalog_version`, `source`, `stage_count` |
| `crop_profile_override_deleted` | `DELETE …/crop-profiles/{crop_key}` | `crop_key`, `catalog_version` |

Visible via `GET /farms/{id}/audit-events` (owner/manager).

---

## Operator behavior

| Surface | Behavior |
|---------|----------|
| **Settings → Crops & targets** | Override rows show **Last changed** timestamp |
| **Farm audit trail** | Override save/reset appears with catalog version |
| **Compliance export** | SQL filter on `crop_profile_override_%` kinds (playbook) |

---

## Automated tests

| Test | Path |
|------|------|
| Upsert + delete audit rows + feed | `cmd/api/smoke_phase105_test.go` |

---

## OC-105

Phase 105 is **closed** when override changes appear in the audit feed with `catalog_version`, Settings shows last-changed for overrides, and OC-84 closure doc is indexed in phase-14.
