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

## Phase 33 & 32 — Guardian grow setup (shipped)

| Phase | Focus | Plan |
|-------|--------|------|
| **32** ✅ | Grow-setup PR bundles + platform doc RAG | [`phase_32_guardian_grow_setup_prs.plan.md`](plans/phase_32_guardian_grow_setup_prs.plan.md) |
| **33** ✅ | context_ref dedup, read-tool audit log, @hardware lane, site manifest | [`phase_33_guardian_polish_and_enterprise_ops.plan.md`](plans/phase_33_guardian_polish_and_enterprise_ops.plan.md) |

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
| **40** | **Zone cockpit** (**shipped**) — Today strip, comfort targets, zone alerts, grow story, Ask gr33n starters | [`phase_40_unified_farmer_ux_zone_cockpit.plan.md`](plans/phase_40_unified_farmer_ux_zone_cockpit.plan.md) · [operator-tour §4b](operator-tour.md#4b-zone-cockpit-walkthrough-phase-40) · [architecture §7.0f](farm-guardian-architecture.md#70f-zone-cockpit-phase-40) |
| **41–47** | **Farmer UX arc** (**shipped**) — farm hub → feeding & water → comfort → operations → setup → sit-in → Guardian LLM PRs | **[`farmer_ux_roadmap_40_plus.plan.md`](plans/farmer_ux_roadmap_40_plus.plan.md)** · [41](plans/phase_41_farm_hub_coherence.plan.md) · [47](plans/phase_47_feeding_water_plain_language.plan.md) · [42](plans/phase_42_comfort_targets_automation_plain_language.plan.md) · [43](plans/phase_43_operations_stock_feeding_finance.plan.md) · [44](plans/phase_44_getting_started_edge_wizard.plan.md) · [45](plans/phase_45_farmer_validation_whole_app_polish.plan.md) · [46](plans/phase_46_guardian_llm_tool_proposals.plan.md) · [`farmer-vocabulary.md`](farmer-vocabulary.md) |

## Phases 48–67 — Farmer closure + Guardian arcs (shipped)

Hub: [`phase_53_59_roadmap.plan.md`](plans/phase_53_59_roadmap.plan.md). Includes nav polish (49), Pi wiring + config sync (50–51), Guardian UI context (52), grow/stock/money closure (53–59), morning walkthrough + nudges + grow advisor + session memory (60–63), crop knowledge base (64), Pi diagnostics (65), weather/site (66), hands-free field assistant (67).

**Historical closure rollup (OC-35 … OC-67):** [`phase_35_37_operational_closure.plan.md`](plans/phase_35_37_operational_closure.plan.md) — **archived, do not extend.** Shipped phases' "close when" conditions are deprecated per [plan lifecycle rules](plans/phase_68_73_spa_workspace_roadmap.plan.md#plan-lifecycle-rules-for-all-phase-plans).

## Phases 68–81 — SPA workspace refactor

Hub: [`phase_68_73_spa_workspace_roadmap.plan.md`](plans/phase_68_73_spa_workspace_roadmap.plan.md) (covers **68–77** formally; **78–81** are post-arc UI polish on `main`). **UI arc shipped** through Phase 81; Phases **70–73** remain for Pi GPIO depth and Guardian PR discoverability. Closure for 68+ lives in each phase plan + arc hub OC table + Vitest `phase-*-closure.test.js` files.

| Phase | Focus | Plan / tests |
|-------|--------|------|
| **68** ✅ | Workspace shell | [`phase_68_workspace_shell_spa_nav.plan.md`](plans/phase_68_workspace_shell_spa_nav.plan.md) |
| **69** ✅ | Zone inline hub | [`phase_69_zone_workspace_hub.plan.md`](plans/phase_69_zone_workspace_hub.plan.md) |
| **70** | Hardware / Pi GPIO | [`phase_70_hardware_pi_control_spa.plan.md`](plans/phase_70_hardware_pi_control_spa.plan.md) |
| **71** | Feed & Water SPA | [`phase_71_feed_water_unification.plan.md`](plans/phase_71_feed_water_unification.plan.md) |
| **72** | Money SPA | [`phase_72_money_unification.plan.md`](plans/phase_72_money_unification.plan.md) |
| **73** | Guardian discoverability | [`phase_73_guardian_pr_discoverability.plan.md`](plans/phase_73_guardian_pr_discoverability.plan.md) |
| **74** ✅ | Zone ops (Tasks, Alerts, Plants) | [`phase_74_zone_ops_inbox.plan.md`](plans/phase_74_zone_ops_inbox.plan.md) |
| **75** ✅ | Comfort & automation workspace | [`phase_75_automation_comfort_workspace.plan.md`](plans/phase_75_automation_comfort_workspace.plan.md) |
| **76** ✅ | Today + mobile nav alignment | [`phase_76_today_dashboard_nav_alignment.plan.md`](plans/phase_76_today_dashboard_nav_alignment.plan.md) |
| **77** ✅ | Post-arc polish | [`phase_77_post_arc_ui_polish.plan.md`](plans/phase_77_post_arc_ui_polish.plan.md) |
| **78** ✅ | Zone-first hardware & GPIO on alerts | `ui/src/__tests__/phase-78-closure.test.js` |
| **79** ✅ | Tasks fix, Money inventory tab, operator glossary | `ui/src/__tests__/phase-79-closure.test.js` |
| **80** ✅ | Routing hashes, zones tab labels, workspace routes | `ui/src/__tests__/phase-80-closure.test.js` |
| **81** ✅ | `/pi-setup` restore, Help Pi tab, zone hardware on Overview only | `ui/src/__tests__/phase-81-closure.test.js` |
| **82** | Guardian plant intelligence — ≥25 crop library, plant context bundle, substrate watering, deficiency guides, zero-chunk guardrail | [`phase_82_guardian_crop_grounding_hardening.plan.md`](plans/phase_82_guardian_crop_grounding_hardening.plan.md) |
| **83** ✅ | Enterprise agronomy seed pack — commons cultivator pack, `guardian-bootstrap-farm`, site-manifest hook, farm crop overrides, scheduled RAG ingest, readiness smokes | [`phase_83_enterprise_agronomy_seed_pack.plan.md`](plans/phase_83_enterprise_agronomy_seed_pack.plan.md) · [`phase-83-closure.md`](plans/phase-83-closure.md) |
| **84** ✅ | Enterprise crop catalog in Postgres — picker API, EC/stage targets, field guides DB, commons API | [`phase_84_crop_catalog_enterprise_db.plan.md`](plans/phase_84_crop_catalog_enterprise_db.plan.md) · [`crop-catalog-db-cutover-runbook.md`](crop-catalog-db-cutover-runbook.md) |
| **85** | **Plants UX** — catalog dropdown, `plants.crop_key`, one slot per crop; Settings EC link | [`phase_85_catalog_bound_plants.plan.md`](plans/phase_85_catalog_bound_plants.plan.md) |
| **86** | Grow ops + **Guardian crop chain** — strip, Water/Light, `lookup_crop_targets` same as UI | [`phase_86_grow_ops_catalog_chain.plan.md`](plans/phase_86_grow_ops_catalog_chain.plan.md) |
| **87** | Guardian crop API smokes + operator runbook + OC-87 | [`phase_87_crop_knowledge_operator_closure.plan.md`](plans/phase_87_crop_knowledge_operator_closure.plan.md) |
| **88** | Domain enums API — growth stages, reservoir, cost categories (UI drops duplicates) | [`phase_88_domain_enums_api.plan.md`](plans/phase_88_domain_enums_api.plan.md) |
| **89** | Lighting presets — wire `GET /lighting-programs/presets` (API exists, UI unused) | [`phase_89_lighting_presets_api_wiring.plan.md`](plans/phase_89_lighting_presets_api_wiring.plan.md) |
| **90** | Device taxonomy registry — sensor/actuator → water/light/climate + Guardian | [`phase_90_device_taxonomy_registry.plan.md`](plans/phase_90_device_taxonomy_registry.plan.md) |
| **91** | Bootstrap template catalog — replace `bootstrapTemplates.js` | [`phase_91_bootstrap_template_catalog.plan.md`](plans/phase_91_bootstrap_template_catalog.plan.md) |
| **92** | Zone types + greenhouse enums from API | [`phase_92_zone_greenhouse_vocabulary.plan.md`](plans/phase_92_zone_greenhouse_vocabulary.plan.md) |

## Master roadmap — Phases 84–100

**Locked order + blind spot map:** [`plans/phase_84_100_master_roadmap.plan.md`](plans/phase_84_100_master_roadmap.plan.md)

**Execute:** 85 → 86 → 87 → **93** → then 89 → 88 → 90–92 → 95–100 as needed.

| Phase | Focus |
|-------|--------|
| **93** | Identity cleanup — `batch_label`, no typed display_name |
| **94** | Genetics / batch EC profiles |
| **95** | Catalog integrator ops cadence |
| **96** | Feeding program vs stage validation |
| **97** | RAG vs structured truth governance |
| **98** | Enterprise catalog promotion model |
| **99** | CI `check-ui-domain-parity` |
| **100** | Offline catalog cache (LAN/mobile) |

## Phases 84–87 — Plants & crop knowledge base

**Roadmap hub:** [`plans/phase_84_87_crop_identity_roadmap.plan.md`](plans/phase_84_87_crop_identity_roadmap.plan.md) · **Sit-in:** [`workstreams/sit-in-crop-catalog-enterprise-db.md`](workstreams/sit-in-crop-catalog-enterprise-db.md)

**Plants are a first-class gr33n surface.** Zone → **Plants**, the **Plants** workspace, and **Start grow** all use a **Postgres-backed catalog dropdown** (every `crop_library.yaml` crop + EC/light/watering preview). **Settings → Crops & targets** tunes EC per farm. **Farm Guardian** must use the **same crop APIs / DB profiles** as the UI — `lookup_crop_targets` — never invented EC.

| Phase | Status | One job |
|-------|--------|---------|
| **84** ✅ | Shipped | Full catalog + picker API + field guides in Postgres |
| **85** | Planned | Catalog-bound **plants** — dropdown only; no “strain” / free-text flooding |
| **86** | Planned | Grow strip + Water/Light + Guardian resolve `plants.crop_key` on active cycle |
| **87** | Planned | Guardian crop API smokes + operator runbook + OC-87 |

**Picker 404?** `make migrate` · restart API · `CROP_CATALOG_SOURCE=db`.

**Prompt loop:** `phase 85 ws1`, … or `phase 85` for full phase.

## Phases 88–92 — UI static data → DB/API

**Hub:** [`plans/phase_88_92_platform_data_gaps_roadmap.plan.md`](plans/phase_88_92_platform_data_gaps_roadmap.plan.md)

Hardcoded UI constants (growth stages, lighting presets, sensor taxonomy, bootstrap templates, zone types) that should be **fetched from API/DB** so operators and Guardian stay aligned. Suggested order: **89** (quick) → **88** → **90** → **91** → **92**.

## Phase 83 — Enterprise agronomy seed pack (shipped)

Cross-linked from Phase 82/84 crop catalog work. **Canonical plan:** [`plans/phase_83_enterprise_agronomy_seed_pack.plan.md`](plans/phase_83_enterprise_agronomy_seed_pack.plan.md) · **Closure:** [`plans/phase-83-closure.md`](plans/phase-83-closure.md).

| WS | Operator path |
|----|----------------|
| **WS1** | [`import-agronomy-seed-pack.sh`](../scripts/enterprise/import-agronomy-seed-pack.sh) · commons slug `gr33n-cultivator-seed-pack-v1` |
| **WS2** | [`apply-agronomy-overrides.sh`](../scripts/enterprise/apply-agronomy-overrides.sh) · [`data/agronomy-override-pack.example.yaml`](../data/agronomy-override-pack.example.yaml) |
| **WS3** | `make guardian-bootstrap-farm FARM_ID=N` · [`guardian-bootstrap-farm.sh`](../scripts/enterprise/guardian-bootstrap-farm.sh) |
| **WS4** | [`site-manifest.example.yaml`](../scripts/enterprise/site-manifest.example.yaml) — `guardian_seed` block |
| **WS5** | [`rag-ingest-farm-operational.sh`](../scripts/rag-ingest-farm-operational.sh) · cron example in enterprise README |
| **WS6** | **Settings → Crops & targets** · `PUT/DELETE /farms/{id}/crop-profiles/{crop_key}` |
| **WS7** | [`guardian-real-grow-readiness.md`](guardian-real-grow-readiness.md) · [`cmd/api/smoke_phase83_test.go`](../cmd/api/smoke_phase83_test.go) |
| **WS8** | This index + [`farm-guardian-architecture.md` §7.0ae](farm-guardian-architecture.md#70ae-enterprise-agronomy-bootstrap-phase-83--shipped) · [operator tour §6o](operator-tour.md#6o-enterprise-agronomy-bootstrap-phase-83--shipped) |

**Pre-flight:** `make migrate` · `make check-crop-catalog-parity` · [`crop-catalog-db-cutover-runbook.md`](crop-catalog-db-cutover-runbook.md).

**Pre-dev gap index (archived):** [`pre_development_gaps_index.plan.md`](plans/pre_development_gaps_index.plan.md) · **Product backlog:** [`product_backlog_operator_runtime.plan.md`](plans/product_backlog_operator_runtime.plan.md) (**shipped**)

After operator-doc edits, re-ingest Guardian RAG: **`make rag-ingest-platform-docs`** and **`make rag-ingest-field-guides`** (Phase 37 field corpus).

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
| Phase 83 — Enterprise agronomy seed pack (**shipped**) | [`plans/phase_83_enterprise_agronomy_seed_pack.plan.md`](plans/phase_83_enterprise_agronomy_seed_pack.plan.md) · [`phase-83-closure.md`](plans/phase-83-closure.md) — Guardian bootstrap, commons pack, crop overrides |
| Phases 84–87 — Crop identity & knowledge base | [`plans/phase_84_87_crop_identity_roadmap.plan.md`](plans/phase_84_87_crop_identity_roadmap.plan.md) · [84](plans/phase_84_crop_catalog_enterprise_db.plan.md) · [85](plans/phase_85_catalog_bound_plants.plan.md) · [86](plans/phase_86_grow_ops_catalog_chain.plan.md) · [87](plans/phase_87_crop_knowledge_operator_closure.plan.md) |
| Phases 88–92 — UI static data → DB/API | [`plans/phase_88_92_platform_data_gaps_roadmap.plan.md`](plans/phase_88_92_platform_data_gaps_roadmap.plan.md) |
| Phases 84–100 — Master roadmap (order + blind spots) | [`plans/phase_84_100_master_roadmap.plan.md`](plans/phase_84_100_master_roadmap.plan.md) |
| Phases 93–100 — Blind spot closure | [93](plans/phase_93_plant_identity_vocabulary_cleanup.plan.md) · [94](plans/phase_94_genetics_batch_ec_profiles.plan.md) · [95](plans/phase_95_catalog_integrator_ops.plan.md) · [96](plans/phase_96_grow_feeding_program_validation.plan.md) · [97](plans/phase_97_rag_structured_truth_governance.plan.md) · [98](plans/phase_98_enterprise_catalog_promotion.plan.md) · [99](plans/phase_99_ci_domain_parity_guards.plan.md) · [100](plans/phase_100_offline_catalog_cache.plan.md) |
| Crop catalog DB cutover | [`crop-catalog-db-cutover-runbook.md`](crop-catalog-db-cutover-runbook.md) |
| Hardware sizing (API, DB, Ollama, RAG, chat) | [`recommended-hardware-and-sizing.md`](recommended-hardware-and-sizing.md) |
| Guardian real-grow readiness (live plants) | [`guardian-real-grow-readiness.md`](guardian-real-grow-readiness.md) |
| First session after clone | [`first-session-after-clone.md`](first-session-after-clone.md) |
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
| Pre-development gaps index (archived) | [`plans/pre_development_gaps_index.plan.md`](plans/pre_development_gaps_index.plan.md) |
| Product backlog (run now, mobile, …) | [`plans/product_backlog_operator_runtime.plan.md`](plans/product_backlog_operator_runtime.plan.md) — shipped |
| Operational closure (archived OC-35 … OC-67) | [`plans/phase_35_37_operational_closure.plan.md`](plans/phase_35_37_operational_closure.plan.md) — frozen; do not extend |
| SPA workspace arc 68–81 (UI shipped; 70–73 planned) | [`plans/phase_68_73_spa_workspace_roadmap.plan.md`](plans/phase_68_73_spa_workspace_roadmap.plan.md) · [operator tour §7e–§7j](operator-tour.md#7e-workspaces--sidebar-jobs-tabs-inside-phase-68) |
| Phase 37 — Guardian offline field assistant (plan) | [`plans/phase_37_guardian_offline_field_assistant.plan.md`](plans/phase_37_guardian_offline_field_assistant.plan.md) — Pi wiring / plumbing walkthroughs, trades corpus, safety gating, offline |
| Pi pending_command + pulse | [`pi-integration-guide.md`](pi-integration-guide.md) §1.1 |
| Workflow — single slot + manual mix | [`workflow-guide.md`](workflow-guide.md) |

## Using this in a new chat

Reference `@docs/phase-14-operator-documentation.md` for Phase 14 deliverables and the **[Phase 31 field validation index](phase-14-operator-documentation.md#phase-31-field-validation-edge)**. For **grow-environment + Guardian doc alignment**, see **[Phases 34–39](#phases-34-39--guardian-polish-grow-environment-plant-needs-ui-edge-fertigation)**. For **current UI nav**, see **[Phases 68–81](#phases-6881--spa-workspace-refactor)** and **[operator tour §7e–§7j](operator-tour.md#7e-workspaces--sidebar-jobs-tabs-inside-phase-68)**. For **first-time local setup**, use **[`local-operator-bootstrap.md`](local-operator-bootstrap.md)** (`make bootstrap-local`).
