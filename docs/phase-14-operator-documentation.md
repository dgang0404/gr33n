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
| **WS8 — Smokes** | [`cmd/api/smoke_phase31_ws8_test.go`](../cmd/api/smoke_phase31_ws8_test.go) (live reading path; live GPIO E2E in the opt-in `@hardware` lane — see [INSTALL §6](../INSTALL.md#6-smoke-test)) |

**Multi-site / enterprise (hypothetical):** [`hypothetical-enterprise-topology.md`](hypothetical-enterprise-topology.md) — maps org/farm/zone onto warehouse-scale deployments without new core tables; links Phase 30 PR queue + Phase 31 field proof paths. **Planned:** Phase 33 WS5 [`site-manifest.yaml`](../scripts/enterprise/) bring-up stub.

## Phase 33 & 32 — Guardian next (planned)

| Phase | Focus | Plan |
|-------|--------|------|
| **33 WS1** (optional first) | Read-tool hardening, persona/architecture doc parity | [`phase_33_guardian_polish_and_enterprise_ops.plan.md`](plans/phase_33_guardian_polish_and_enterprise_ops.plan.md) |
| **32** | Grow-setup PR bundles + platform doc RAG | [`phase_32_guardian_grow_setup_prs.plan.md`](plans/phase_32_guardian_grow_setup_prs.plan.md) |
| **33 (shipped)** | context_ref dedup, read-tool audit log, @hardware lane, site manifest | Same Phase 33 plan |

## Phases 34–39 — Guardian polish, grow environment, plant-needs UI, edge fertigation

| Phase | Focus | Plan / operator path |
|-------|--------|----------------------|
| **34** | PR revise loop, operator-supplied blind-spot facts, impact explanations | [`phase_34_guardian_pr_iteration.plan.md`](plans/phase_34_guardian_pr_iteration.plan.md) |
| **35** | Lighting domain — photoperiod programs, presets, timer UX | [`phase_35_lighting_domain.plan.md`](plans/phase_35_lighting_domain.plan.md) · [operator-tour §5a](operator-tour.md) |
| **36** | Greenhouse climate (**shipped**) — shade/vents/fans; Guardian **`summarize_zone_greenhouse_climate`**; WS6 sensor interlocks; OC-36C smokes; **Climate** tab via Phase 38 | [`phase_36_greenhouse_climate.plan.md`](plans/phase_36_greenhouse_climate.plan.md) · [operator-tour §5b](operator-tour.md#5b-greenhouse-shade-vents-and-fans-phase-36) · [architecture §7.0c](farm-guardian-architecture.md) |
| **37** | **Offline field assistant** (**shipped**) — `field_guide` RAG, guided procedures, safety stops, LLM-down degrade, print checklists, background Guardian chat | [`phase_37_guardian_offline_field_assistant.plan.md`](plans/phase_37_guardian_offline_field_assistant.plan.md) · [operator-tour §6d](operator-tour.md#6d-first-field-install-with-guardian-offline-phase-37) · [architecture §7.0e](farm-guardian-architecture.md#70e-offline-field-assistant-phase-37) |
| **38** | **Plant-needs UI** (**shipped**) — Zones → **Water / Light / Climate**; Advanced nav; **`duration_seconds`** pump pulse on `pending_command` | [`phase_38_plant_needs_ui_and_pulse_commands.plan.md`](plans/phase_38_plant_needs_ui_and_pulse_commands.plan.md) · [operator-tour §4a](operator-tour.md#4a-plant-needs-per-zone-phase-38) · [architecture §7.0d](farm-guardian-architecture.md#70d-plant-needs-ui--pulse-phase-38) |
| **39** | **Edge fertigation** (**shipped**) — FIFO queue, `mix_batch`, Pi executor | [`phase_39_edge_fertigation_execution.plan.md`](plans/phase_39_edge_fertigation_execution.plan.md) |
| **39b** | **Plain irrigation** (**shipped**) — RO/well programs without mix | [`phase_39b_plain_irrigation.plan.md`](plans/phase_39b_plain_irrigation.plan.md) |
| **40–47** | **Farmer UX arc** — zone cockpit → farm hub → feeding & water → comfort → operations → setup → sit-in → Guardian LLM PRs | **[`farmer_ux_roadmap_40_plus.plan.md`](plans/farmer_ux_roadmap_40_plus.plan.md)** · [40](plans/phase_40_unified_farmer_ux_zone_cockpit.plan.md) · [41](plans/phase_41_farm_hub_coherence.plan.md) · [47](plans/phase_47_feeding_water_plain_language.plan.md) · [42](plans/phase_42_comfort_targets_automation_plain_language.plan.md) · [43](plans/phase_43_operations_stock_feeding_finance.plan.md) · [44](plans/phase_44_getting_started_edge_wizard.plan.md) · [45](plans/phase_45_farmer_validation_whole_app_polish.plan.md) · [46](plans/phase_46_guardian_llm_tool_proposals.plan.md) · [`farmer-vocabulary.md`](farmer-vocabulary.md) |

**Pre-dev gap index:** [`pre_development_gaps_index.plan.md`](plans/pre_development_gaps_index.plan.md) · **Product backlog:** [`product_backlog_operator_runtime.plan.md`](plans/product_backlog_operator_runtime.plan.md) (**shipped**)

Rollup: [`phase_35_37_operational_closure.plan.md`](plans/phase_35_37_operational_closure.plan.md) (spans **35–47** OC rows). After operator-doc edits, re-ingest Guardian RAG: **`make rag-ingest-platform-docs`** and **`make rag-ingest-field-guides`** (Phase 37 field corpus).

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
| Phase 34 — Guardian PR iteration (plan) | [`plans/phase_34_guardian_pr_iteration.plan.md`](plans/phase_34_guardian_pr_iteration.plan.md) — revise/supersede loop, operator-stated facts |
| Phase 35 — Lighting domain (plan) | [`plans/phase_35_lighting_domain.plan.md`](plans/phase_35_lighting_domain.plan.md) — photoperiod programs, presets, PhotoperiodClockEditor |
| Phase 36 — Greenhouse climate (plan) | [`plans/phase_36_greenhouse_climate.plan.md`](plans/phase_36_greenhouse_climate.plan.md) — shipped (interlocks + smokes) |
| Phase 38 — Plant-needs UI + pulse (plan) | [`plans/phase_38_plant_needs_ui_and_pulse_commands.plan.md`](plans/phase_38_plant_needs_ui_and_pulse_commands.plan.md) — Zones Water/Light/Climate; `duration_seconds` |
| Phase 39 — Edge fertigation execution (plan) | [`plans/phase_39_edge_fertigation_execution.plan.md`](plans/phase_39_edge_fertigation_execution.plan.md) — command queue + automated mix |
| Farmer UX roadmap 40–47 | [`plans/farmer_ux_roadmap_40_plus.plan.md`](plans/farmer_ux_roadmap_40_plus.plan.md) — full site vision + phase order |
| Phase 40–47 plans | 40 cockpit · 41 hub · **47 feed/water** · 42 comfort · 43 ops · 44 setup · 45 sit-in · 46 LLM PRs |
| Farmer vocabulary | [`farmer-vocabulary.md`](farmer-vocabulary.md) — grow-path language contract |
| Pre-development gaps index | [`plans/pre_development_gaps_index.plan.md`](plans/pre_development_gaps_index.plan.md) |
| Product backlog (run now, mobile, …) | [`plans/product_backlog_operator_runtime.plan.md`](plans/product_backlog_operator_runtime.plan.md) — shipped |
| Phase 35–47 operational closure | [`plans/phase_35_37_operational_closure.plan.md`](plans/phase_35_37_operational_closure.plan.md) — OC-40…47 close with each phase WS8 |
| Phase 37 — Guardian offline field assistant (plan) | [`plans/phase_37_guardian_offline_field_assistant.plan.md`](plans/phase_37_guardian_offline_field_assistant.plan.md) — Pi wiring / plumbing walkthroughs, trades corpus, safety gating, offline |
| Pi pending_command + pulse | [`pi-integration-guide.md`](pi-integration-guide.md) §1.1 |
| Workflow — single slot + manual mix | [`workflow-guide.md`](workflow-guide.md) |

## Using this in a new chat

Reference `@docs/phase-14-operator-documentation.md` for Phase 14 deliverables and the **[Phase 31 field validation index](phase-14-operator-documentation.md#phase-31-field-validation-edge)**. For **grow-environment + Guardian doc alignment**, see **[Phases 34–39](#phases-34-39--guardian-polish-grow-environment-plant-needs-ui-edge-fertigation)** (lighting → greenhouse → plant-needs UI → edge mix queue). For **first-time local setup**, use **[`local-operator-bootstrap.md`](local-operator-bootstrap.md)** (`make bootstrap-local`).
