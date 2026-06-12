# Phase 109 — closure (OC-109)

**Status:** **Shipped** on `main`.

**Canonical plan:** [`phase_109_catalog_version_push_notifications.plan.md`](phase_109_catalog_version_push_notifications.plan.md)

**Depends on:** [Phase 98](phase_98_enterprise_catalog_promotion.plan.md) multi-site promotion; Phase 14 notifications infra; [Phase 95](phase_95_catalog_integrator_ops.plan.md) `catalog_version` bump cadence.

**Closes:** Farm admins are notified when platform **`catalog_version`** bumps after migrate — not on per-farm EC overrides or single-farm RAG re-ingest.

---

## The one job (done)

> After HQ ships a catalog seed migration, each site’s API detects the version bump on startup and notifies opted-in **farm owners/managers** to review the picker and optionally re-run bootstrap.

---

## Workstream checklist

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS0** | Notifications infra + farm admin role | Phase 14, `alerts_notifications` |
| **WS1** | Platform catalog state | `gr33ncore.platform_catalog_state`, `farm_catalog_version_seen` |
| **WS2** | Startup detect + emit | `catalognotify.SyncOnStartup` / `Sync()` |
| **WS3** | In-app alert + optional FCM push | `catalog_version_bump` alerts; FCM when configured |
| **WS4** | Opt-out preference | `profiles.preferences.notify.catalog_updates`; Settings UI |
| **WS5** | Enterprise runbook | [`enterprise-catalog-version-notifications.md`](../enterprise-catalog-version-notifications.md) |

---

## Trigger rules

| Event | Notify? |
|-------|---------|
| `catalog_version` in DB > last seen per farm | Yes |
| Farm EC override only | No |
| Single-farm field guide re-ingest | No |

Message links operators to Settings / crop-knowledge runbook; optional `guardian-bootstrap-farm.sh`.

---

## Automated tests

| Test | Path |
|------|------|
| Bump creates alert; opt-out skips | `cmd/api/smoke_phase109_test.go` — `TestPhase109_CatalogVersionNotify` |
| Already-at-version debounced | same — `TestPhase109_CatalogVersionNotifyDebounced` |

Migration: `20260628_phase109_catalog_version_notifications.sql`

---

## OC-109

Phase 109 is **closed** when catalog version bump creates admin alerts, `catalog_updates` opt-out is respected, and the enterprise runbook documents HQ → site rollout.
