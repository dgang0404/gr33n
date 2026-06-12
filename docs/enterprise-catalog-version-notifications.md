# Enterprise catalog version notifications (Phase 109)

When HQ publishes a new platform **`crop_catalog_entries.catalog_version`** via migrate, each site’s API detects the bump on startup and notifies **farm owners and managers** who have not opted out.

## Flow

1. **HQ** ships migration that bumps `catalog_version` on seed rows (e.g. Phase 84 seed v4 → v5).
2. **Each site** runs `make migrate` — Postgres max `catalog_version` increases.
3. **API startup** runs `catalognotify.SyncOnStartup`:
   - Reads `gr33ncore.platform_catalog_state` (last notified version).
   - Compares to live max from `gr33ncrops.crop_catalog_entries`.
   - For each farm where `farm_catalog_version_seen < current`, creates **`catalog_version_bump`** rows in `gr33ncore.alerts_notifications` for opted-in admins.
   - Updates platform state and per-farm seen version (idempotent per version).

## Operator actions after notice

| Step | Command / UI |
|------|----------------|
| Review new crops | Settings → **Crops & targets**, Plants picker |
| Refresh Guardian RAG | [`scripts/enterprise/guardian-bootstrap-farm.sh`](../../scripts/enterprise/guardian-bootstrap-farm.sh) |
| Optional commons pack | [`import-agronomy-seed-pack.sh`](../../scripts/enterprise/import-agronomy-seed-pack.sh) if org uses seed pack provenance |

Farm **EC overrides** and single-farm field guide re-ingest do **not** trigger this notice — only platform catalog version bumps.

## Notification preferences

Stored under `profiles.preferences.notify`:

| Field | Default | Meaning |
|-------|---------|---------|
| `catalog_updates` | `true` | In-app alert + push (when FCM enabled) on catalog bump |
| `push_enabled` | `false` | FCM delivery (Phase 14 WS5) |

Patch via `PATCH /profile/notification-preferences` with `{ "catalog_updates": false }` to opt out.

Settings → **Push Notifications** exposes **Knowledge base update notices**.

## Push (FCM)

When `FCM_SERVICE_ACCOUNT_JSON` or `GOOGLE_APPLICATION_CREDENTIALS` is set, opted-in admins receive a push with `data.kind=catalog_update`. In-app alerts always use `triggering_event_source_type=catalog_version_bump`.

See [`notifications-operator-playbook.md`](notifications-operator-playbook.md).

## Multi-site rollout (HQ → sites)

```text
HQ: merge migration → tag release (catalog v5)
Site A/B/C: pull → make migrate → restart API → admins see alert
Optional: guardian-bootstrap-farm.sh per farm
```

Pin expected version in site manifest — see [Phase 98 promotion model](plans/phase_98_enterprise_catalog_promotion.plan.md).

## Related

- [`crop-catalog-db-cutover-runbook.md`](crop-catalog-db-cutover-runbook.md)
- [`commons-catalog-operator-playbook.md`](commons-catalog-operator-playbook.md)
- Phase 109 plan: [`plans/phase_109_catalog_version_push_notifications.plan.md`](plans/phase_109_catalog_version_push_notifications.plan.md)
