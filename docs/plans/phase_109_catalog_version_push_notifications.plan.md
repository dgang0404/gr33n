---
name: Phase 109 — Catalog version push notifications
overview: >
  Notify farm admins when platform catalog_version bumps after migrate — enterprise
  multi-site ops; optional FCM + in-app banner.
todos:
  - id: ws0-deps
    content: "WS0: Phase 14 WS5 notifications infra; farm admin role"
    status: completed
  - id: ws1-version
    content: "WS1: gr33ncore.platform_catalog_version table or settings key"
    status: completed
  - id: ws2-detect
    content: "WS2: API startup or cron compares catalog_version → emit audit event"
    status: completed
  - id: ws3-notify
    content: "WS3: FCM + in-app notification — 'Knowledge base updated — review new crops'"
    status: completed
  - id: ws4-prefs
    content: "WS4: profiles.preferences.notify.catalog_updates opt-in"
    status: completed
  - id: ws5-runbook
    content: "WS5: Enterprise runbook — migrate HQ → notify sites → re-ingest RAG"
    status: completed
isProject: false
---

# Phase 109 — Catalog version push notifications

## Status

**Shipped** on `main`. Closure: [`phase-109-closure.md`](phase-109-closure.md) (**OC-109**).

Enterprise **multi-site** ops — sites know when HQ publishes new catalog seed.

**Depends on:** [Phase 98](phase_98_enterprise_catalog_promotion.plan.md), [Phase 95](phase_95_catalog_integrator_ops.plan.md), notifications playbook.

**Closure:** **OC-109**

---

## The one job

> After platform **`catalog_version`** bump + migrate, **farm admins** get a push/in-app notice to refresh picker and optionally re-run **`guardian-bootstrap-farm`**.

---

## Trigger (WS2)

| Event | Notify |
|-------|--------|
| `catalog_version` in DB > last seen per farm/org | Yes |
| Farm EC override only | No |
| Field guide re-ingest on one farm | Optional (operational ingest separate) |

---

## Message template

> **gr33n knowledge base updated (v4 → v5).** New crops available in Plants picker. Tap to review or run bootstrap.

Link: Settings or crop-knowledge runbook.

---

## Acceptance

- [ ] Bump seed migration version → notification row for farm admins
- [ ] Opt-out respected in notify prefs
- [ ] Document in enterprise-catalog-promotion-model (Phase 98)

**Prompt loop:** **`phase 109`**.
