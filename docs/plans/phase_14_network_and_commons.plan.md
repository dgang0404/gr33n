---
name: Phase 14 Field network and commons
overview: Phase 14 builds on Phase 13 — deeper field connectivity (MQTT and microcontrollers), a disciplined data-insert pipeline toward federation-ready sharing, lightweight gr33n_inserts/community commons surfaces, federation and receiver maturity, notification strategy, and governance gaps (org-level audit). Domain module seeds (crops, animals, aquaponics) are optional stretch goals, not a full ERP.
todos:
  - id: edge-mqtt
    content: "WS1-Edge: MQTT broker integration patterns, microcontroller-friendly ingest, and field tasking hooks aligned with existing tasks/devices (no vendor lock-in; document self-hosted Mosquitto vs cloud)."
    status: completed
  - id: insert-pipeline
    content: "WS2-Inserts: Data insert pipeline — validation, scrubbing, optional human approval queue, and export/package formats that match Insert Commons schema evolution (reversible migrations)."
    status: completed
  - id: commons-catalog
    content: "WS3-Commons: gr33n_inserts direction — contribution metadata, licensing hints, minimal browse or import API; stop short of marketplace, payments, or reputation systems."
    status: completed
  - id: federation-depth
    content: "WS4-Federation: Receiver and sender hardening — aggregation or query surfaces for pilots, optional idempotency correlation, retention and operator dashboards; cross-farm benchmarks only where privacy story holds."
    status: completed
  - id: notifications
    content: "WS5-Notify: Push and in-app notification policy — FCM/APNs via Capacitor when needed, web push optional later; tie to alert semantics and operator controls for volume."
    status: completed
  - id: org-governance
    content: "WS6-Governance: Org-scoped audit visibility (today some org rows use farm_id 0), audit farm-to-org linking, and any RBAC gaps surfaced in Phase 13."
    status: completed
  - id: domain-modules
    content: "WS7-Domain (stretch): Starter schemas or stubs for gr33n_crops, gr33n_animals, gr33n_aquaponics behind farm_active_modules — reversible, documented, no full feature modules required."
    status: completed
  - id: phase14-docs
    content: "WS8-Docs: README phase banner, OpenAPI for new routes, operator playbooks for MQTT/edge and insert pipeline; extend phase operator index pattern."
    status: completed
  - id: farm-bootstrap-templates
    content: "WS9-Farm bootstrap (optional; may execute in Phase 15): On farm creation, optional template — blank farm vs versioned starter pack (zones, JADAM inputs/recipes, schedules, fertigation baseline, tasks) for any farm_id; refactor master_seed patterns into idempotent apply (API flag or `POST /farms/{id}/apply-template`). See `docs/plans/phase_15_farm_onboarding.plan.md`."
    status: completed
isProject: false
---

# Phase 14 — Field network & commons growth

## Prerequisites

Phase 13 on **`main`** includes: pilot Insert Commons receiver, farm audit API, orgs + usage summary, cost idempotency and finance metadata, offline cost queue patterns, Capacitor scaffold, and [`docs/phase-13-operator-documentation.md`](../phase-13-operator-documentation.md).

## Themes (pick priority order per release train)

| Theme | Outcome |
|-------|---------|
| **Edge reality** | Field devices and MQTT fit gr33n’s task and sensor model without forcing cloud |
| **Commons discipline** | Inserts are safe to share: validated, optionally reviewed, versioned |
| **Federation maturity** | Receivers and senders trustworthy enough for real multi-farm learning |
| **Operator calm** | Notifications and governance reduce surprise; audit coverage matches org actions |

## Farm bootstrap templates (WS9 / Phase 15)

New farms today only get **full demo defaults** if an operator runs [`db/seeds/master_seed.sql`](../../db/seeds/master_seed.sql) (historically **farm_id = 1**). Product direction:

- **On create (or immediately after):** operator chooses **blank** or a **named template** (e.g. indoor photoperiod starter: 18/6 veg + 12/12 flower, JADAM-style inputs/recipes, irrigation presets, fertigation + mixing examples, protocol tasks).
- **Implementation** belongs primarily in **[`docs/plans/phase_15_farm_onboarding.plan.md`](phase_15_farm_onboarding.plan.md)** — parameterized, idempotent apply per `farm_id`, optional org-level default.
- **WS9** here tracks cross-phase alignment (docs, operator expectations) so Phase 14 work does not silently assume “demo farm only.”

## Explicitly out of scope for Phase 14 (defer)

- Full hardware certification or device marketplace
- Full ERP, payroll, or single-vendor accounting replacement
- Native app rewrite (keep Capacitor as wrapper only)
- Full gr33n_inserts reputation economy, payments, or social graph
- **Full farm-template product** (API + UI + org defaults) — see Phase 15 unless pulled into a Phase 14 patch release intentionally

## Suggested execution order

1. **Org governance / audit gaps** — quick trust wins; unblocks enterprise pilots.
2. **Insert pipeline** — foundation for commons and federation quality. *(WS2: see [`insert-commons-pipeline-runbook.md`](../insert-commons-pipeline-runbook.md).)*
3. **Edge & MQTT** — high visibility; sequence after clear security and auth story for device traffic.
4. **Federation depth** — builds on pipeline and receiver.
5. **Commons catalog** — can parallel federation once insert formats stabilize.
6. **Notifications** — after alert semantics and operator UX are agreed.
7. **Domain modules (stretch)** — only if schema capacity is available without derailing edge/commons.
8. **Docs** continuously; add [`docs/phase-14-operator-documentation.md`](../phase-14-operator-documentation.md) when WS8 starts.
9. **Farm bootstrap (WS9)** — if scheduled in Phase 14, keep thin (e.g. doc + spike); otherwise defer implementation to **[Phase 15 — Farm onboarding & templates](phase_15_farm_onboarding.plan.md)**.

## Remaining work queue (through Phase 14 closure)

The **ordered backlog** for WS1–WS8 (plus WS7 stretch), kept in sync with implementation notes, lives in **[`docs/phase-14-operator-documentation.md`](../phase-14-operator-documentation.md)**. Use that file as the default “next bout of work” until each workstream is done.

## Using this plan in a new chat

Reference `@docs/plans/phase_14_network_and_commons.plan.md` and **`@docs/phase-14-operator-documentation.md`**; pick the next row in the remaining-work queue or adjust todos to match your release train.
