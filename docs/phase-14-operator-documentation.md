# Phase 14 — operator documentation index

This page is the **Phase 14** counterpart to [`phase-13-operator-documentation.md`](phase-13-operator-documentation.md). It links stable operator playbooks and tracked Phase 14 workstreams **through closure** (WS1–WS9); use **[`plans/phase_15_farm_onboarding.plan.md`](plans/phase_15_farm_onboarding.plan.md)** for the current farm-onboarding focus.

**Canonical plan:** [`plans/phase_14_network_and_commons.plan.md`](plans/phase_14_network_and_commons.plan.md)

## Done in Phase 14 (reference)

| Area | Notes |
|------|--------|
| **WS6 — Org governance** | Org- and farm-scoped audit in dashboard Settings; APIs `GET /organizations/{id}/audit-events`, `GET /farms/{id}/audit-events`. Playbook: [`audit-events-operator-playbook.md`](audit-events-operator-playbook.md). |
| **WS9 — Farm bootstrap** | Cross-listed with Phase 15; optional templates and org default. Plan: [`plans/phase_15_farm_onboarding.plan.md`](plans/phase_15_farm_onboarding.plan.md). |
| **WS2 — Insert pipeline** | Canonical ingest validation (strict top-level keys, aggregate shape, boolean `includes_pii`), preview, sync, approval bundles, `package_v1` export. Runbook: [`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md) (**Custom senders** — no extra top-level keys; complete `aggregates`; use preview as golden JSON). Receiver: [`insert-commons-receiver-playbook.md`](insert-commons-receiver-playbook.md). |
| **WS1 — Edge / MQTT** | Broker-neutral pattern; `POST /sensors/readings/batch`; reference bridge [`pi_client/mqtt_telemetry_bridge.py`](../pi_client/mqtt_telemetry_bridge.py); playbook [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md). |
| **WS4 — Federation depth** | Farm API forwards **`Gr33n-Idempotency-Key`** on outbound ingest; pilot receiver stores `source_idempotency_key`, **`GET /v1/stats`**, playbook updates — [`insert-commons-receiver-playbook.md`](insert-commons-receiver-playbook.md); migration `db/migrations/20260425_insert_commons_receiver_idempotency_stats.sql`. |
| **WS3 — Commons catalog** | Published packs + farm import audit — [`commons-catalog-operator-playbook.md`](commons-catalog-operator-playbook.md); migration `db/migrations/20260426_commons_catalog.sql`; OpenAPI tag **commons**. |
| **WS5 — Notify** | Farm alert push via FCM when credentials are set; prefs under `profiles.preferences.notify`; token APIs under `/profile/*` — [`notifications-operator-playbook.md`](notifications-operator-playbook.md); migration `db/migrations/20260427_user_push_tokens.sql`. |
| **WS7 — Domain stubs** | Schemas `gr33ncrops`, `gr33nanimals`, `gr33naquaponics` with minimal placeholder tables; opt-in via `gr33ncore.farm_active_modules` — [`domain-modules-operator-playbook.md`](domain-modules-operator-playbook.md); migration `db/migrations/20260428_phase14_domain_module_stubs.sql`. |

## Phase 31 — field validation & edge

Cross-linked from Phase 14 because enterprise scale-out and MQTT edge patterns start here. **Canonical plan:** [`plans/phase_31_field_validation_and_edge.plan.md`](plans/phase_31_field_validation_and_edge.plan.md).

| WS | Operator path |
|----|----------------|
| **WS1 — Stub loop** | [`local-operator-bootstrap.md`](local-operator-bootstrap.md) · `make edge-smoke-help` · [`scripts/run-edge-stub-client.sh`](../scripts/run-edge-stub-client.sh) |
| **WS2 — Pi field checklist** | [`pi-integration-guide.md`](pi-integration-guide.md) §8 |
| **WS3 — Safe actuator bench** | [`pi-integration-guide.md`](pi-integration-guide.md) §9 · [`operator-troubleshooting.md`](operator-troubleshooting.md) §5 · `make edge-actuator-smoke-help` |
| **WS4 — MQTT room-scale** | [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md) (room-scale section) · [`pi_client/mqtt_bridge_map.room-scale.example.yaml`](../pi_client/mqtt_bridge_map.room-scale.example.yaml) |
| **WS5 — Recipe pack demo** | [`scripts/enterprise/import-recipe-pack.sh`](../scripts/enterprise/import-recipe-pack.sh) · [`commons-catalog-operator-playbook.md`](commons-catalog-operator-playbook.md) |
| **WS6 — Guardian read tools** | Ask Guardian zone/humidity or alert-list questions with a farm selected — [`farm-guardian-architecture.md`](farm-guardian-architecture.md) |
| **WS8 — Smokes** | [`cmd/api/smoke_phase31_ws8_test.go`](../cmd/api/smoke_phase31_ws8_test.go) (live reading path; hardware tests opt-in) |

**Multi-site / enterprise (hypothetical):** [`hypothetical-enterprise-topology.md`](hypothetical-enterprise-topology.md) — maps org/farm/zone onto warehouse-scale deployments without new core tables; links Phase 30 PR queue + Phase 31 field proof paths. **Planned:** Phase 33 WS5 [`site-manifest.yaml`](../scripts/enterprise/) bring-up stub.

## Phase 33 & 32 — Guardian next (planned)

| Phase | Focus | Plan |
|-------|--------|------|
| **33 WS1** (optional first) | Read-tool hardening, persona/architecture doc parity | [`phase_33_guardian_polish_and_enterprise_ops.plan.md`](plans/phase_33_guardian_polish_and_enterprise_ops.plan.md) |
| **32** | Grow-setup PR bundles + platform doc RAG | [`phase_32_guardian_grow_setup_prs.plan.md`](plans/phase_32_guardian_grow_setup_prs.plan.md) |
| **33 WS2–WS5** | context_ref dedup, read audit, hardware CI, site manifest | Same Phase 33 plan |

## Quick links

| Topic | Doc |
|--------|-----|
| OpenAPI | [`openapi.yaml`](../openapi.yaml) |
| Insert Commons farm pipeline | [`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md) |
| Insert Commons receiver | [`insert-commons-receiver-playbook.md`](insert-commons-receiver-playbook.md) |
| Commons catalog (gr33n_inserts) | [`commons-catalog-operator-playbook.md`](commons-catalog-operator-playbook.md) |
| Audit | [`audit-events-operator-playbook.md`](audit-events-operator-playbook.md) |
| Receipt / storage cutover | [`receipt-storage-cutover-runbook.md`](receipt-storage-cutover-runbook.md) |
| Mobile | [`mobile-distribution.md`](mobile-distribution.md) |
| Push / FCM (alerts) | [`notifications-operator-playbook.md`](notifications-operator-playbook.md) |
| Domain module stubs (crops / animals / aquaponics) | [`domain-modules-operator-playbook.md`](domain-modules-operator-playbook.md) |
| Hypothetical multi-site / enterprise topology (sketch) | [`hypothetical-enterprise-topology.md`](hypothetical-enterprise-topology.md) — central HQ vs frontier sites, commons promotion, [`scripts/enterprise/`](../scripts/enterprise/README.md) |
| Phase 31 — field validation & edge (operator index) | [§ Phase 31 below](#phase-31-field-validation-edge) · [`plans/phase_31_field_validation_and_edge.plan.md`](plans/phase_31_field_validation_and_edge.plan.md) |
| Pi integration + field checklist | [`pi-integration-guide.md`](pi-integration-guide.md) §8–§9 |
| Edge actuator safety | [`operator-troubleshooting.md`](operator-troubleshooting.md) §5 |
| Enterprise deployment scripts | [`scripts/enterprise/README.md`](../scripts/enterprise/README.md) |
| Hardware sizing (API, DB, Ollama, RAG, chat) | [`recommended-hardware-and-sizing.md`](recommended-hardware-and-sizing.md) |
| Farm Guardian — architecture & operator expectations | [`farm-guardian-architecture.md`](farm-guardian-architecture.md) (§8), [`operator-tour.md`](operator-tour.md#6-farm-guardian-change-requests-with-your-ok) |
| Farm Guardian — platform persona (WS9 mirror) | [`farm-guardian-persona-platform-context.md`](farm-guardian-persona-platform-context.md) |
| Phase 30 — Guardian PR queue (plan) | [`plans/phase_30_guardian_change_requests.plan.md`](plans/phase_30_guardian_change_requests.plan.md) |
| Phase 31 — field validation & edge (plan) | [`plans/phase_31_field_validation_and_edge.plan.md`](plans/phase_31_field_validation_and_edge.plan.md) · [operator index § Phase 31](#phase-31-field-validation-edge) |
| Phase 32 — Guardian grow setup PRs (plan) | [`plans/phase_32_guardian_grow_setup_prs.plan.md`](plans/phase_32_guardian_grow_setup_prs.plan.md) — setup pack + platform doc RAG |
| Phase 33 — Guardian polish & enterprise ops (plan) | [`plans/phase_33_guardian_polish_and_enterprise_ops.plan.md`](plans/phase_33_guardian_polish_and_enterprise_ops.plan.md) — read-tool hardening, hardware CI, site manifest |

## Using this in a new chat

Reference `@docs/phase-14-operator-documentation.md` for Phase 14 deliverables and the **[Phase 31 field validation index](phase-14-operator-documentation.md#phase-31-field-validation-edge)**. For **next engineering work**, default to **[Phase 33 WS1](plans/phase_33_guardian_polish_and_enterprise_ops.plan.md)** (optional hardening) then **[Phase 32](plans/phase_32_guardian_grow_setup_prs.plan.md)** (grow-setup PR bundles). For **first-time local setup**, use **[`local-operator-bootstrap.md`](local-operator-bootstrap.md)** (`make bootstrap-local`).
