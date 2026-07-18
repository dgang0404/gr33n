# Phase 14 — operator documentation index

This page is the **Phase 14** counterpart to [`phase-13-operator-documentation.md`](phase-13-operator-documentation.md). It links stable operator playbooks and tracked Phase 14 workstreams **through closure** (WS1–WS9); use **[`plans/phase_15_farm_onboarding.plan.md`](plans/phase_15_farm_onboarding.plan.md)** for the current farm-onboarding focus.

**Canonical plan:** [`plans/phase_14_network_and_commons.plan.md`](plans/phase_14_network_and_commons.plan.md)

## Start here (Phase 157)

| Doc | Use when |
|-----|----------|
| **[`current-state.md`](current-state.md)** | **What gr33n looks like today** — routes, schemas, Guardian, smoke commands |
| [`operator-tour.md`](operator-tour.md) | Walkthrough of the live UI |
| [`first-session-after-clone.md`](first-session-after-clone.md) | Bootstrap after `git clone` |
| [`farm-guardian-architecture.md`](farm-guardian-architecture.md) | Guardian design & operator expectations |
| [`plans/archive/`](plans/archive/) | Closed phase plans (history preserved) |

**Regenerate snapshot:** `make docs-current-state-hint` then edit `current-state.md` prose.

## Active & recent phases

| Phase | Status | Plan |
|-------|--------|------|
| **115** | Shipped | [`phase_115_schema_utilization.plan.md`](plans/phase_115_schema_utilization.plan.md) |
| **157** | Shipped | [`phase_157_docs_consolidation.plan.md`](plans/phase_157_docs_consolidation.plan.md) — `current-state.md`, plans archive |
| **158** | Shipped | [`phase_158_accessibility_pass.plan.md`](plans/phase_158_accessibility_pass.plan.md) — skip link, Guardian a11y, zone tabs |
| **159** | Shipped | [`phase_159_guardian_citation_completeness.plan.md`](plans/phase_159_guardian_citation_completeness.plan.md) — citation WS2b + accuracy_note persist |
| **160** | Shipped | [`phase_160_a11y_residuals.plan.md`](plans/phase_160_a11y_residuals.plan.md) — lighting modal, mobile drawer trap |
| **161** | Shipped | [`phase_161_guardian_ecph_smoke_closure.plan.md`](plans/phase_161_guardian_ecph_smoke_closure.plan.md) — ec-ph tail trim + crop drift |
| **162** | Shipped | [`phase_162_guardian_confirm_db_smoke.plan.md`](plans/phase_162_guardian_confirm_db_smoke.plan.md) — change-request Confirm→DB smoke |
| **202** | Shipped | [`phase_202_closure_test_consolidation.plan.md`](plans/phase_202_closure_test_consolidation.plan.md) — UI closure-test consolidation |
| **203** | Shipped | [`phase_203_handler_package_consolidation.plan.md`](plans/phase_203_handler_package_consolidation.plan.md) — backend helper/package consolidation |
| **204** | Shipped | [`phase_204_docs_navigation_cleanup.plan.md`](plans/phase_204_docs_navigation_cleanup.plan.md) — README rewrite + [roadmap](roadmap/README.md) |
| **205** | Shipped | [`phase_205_pre_existing_test_debt.plan.md`](plans/phase_205_pre_existing_test_debt.plan.md) — 24 UI test fixes + `make check-ui-test-baseline` |

Full narrative history (all eras, plain language): **[`roadmap/README.md`](roadmap/README.md)**. Phases 163–201 (Today cockpit, weather, sit-in arc, answer-quality audit, post-audit follow-through) are summarized there; see individual `plans/phase_N_*.plan.md` for implementation detail.

## Shipped arcs (hub links)

| Arc | Closure / hub |
|-----|----------------|
| Farmer UX 40–67 | [`farmer_ux_roadmap_40_plus.plan.md`](plans/farmer_ux_roadmap_40_plus.plan.md) |
| SPA 68–81 | [`phase_68_73_spa_workspace_roadmap.plan.md`](plans/phase_68_73_spa_workspace_roadmap.plan.md) |
| Crop intelligence 82–110 | [`phase-84-110-closure.md`](plans/phase-84-110-closure.md) |
| Guardian 111–153 | [`phase-129-139-closure.md`](plans/phase-129-139-closure.md) · [`guardian-qa-smoke-report-20260707.md`](guardian-qa-smoke-report-20260707.md) |
| Infra & trust 154–158 | [`phase_154_158_infra_trust_gaps_backlog.plan.md`](plans/phase_154_158_infra_trust_gaps_backlog.plan.md) — 154–158 shipped |
| Post-158 follow-through 159–160 | [`phase_159_160_post_158_gaps_backlog.plan.md`](plans/phase_159_160_post_158_gaps_backlog.plan.md) — Guardian citations + a11y residuals |
| Guardian ec-ph smoke 161 | [`phase_161_guardian_ecph_smoke_closure.plan.md`](plans/phase_161_guardian_ecph_smoke_closure.plan.md) — tail trim + crop drift |
| Guardian confirm→DB 162 | [`phase_162_guardian_confirm_db_smoke.plan.md`](plans/phase_162_guardian_confirm_db_smoke.plan.md) — change-request Confirm smoke |

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

Hub: [`phase_68_73_spa_workspace_roadmap.plan.md`](plans/phase_68_73_spa_workspace_roadmap.plan.md) (covers **68–77** formally; **78–81** are post-arc UI polish on `main`). **SPA arc shipped** through Phase 81 (68–73 + 74–81). Closure for 68+ lives in each phase plan + arc hub OC table + Vitest `phase-*-closure.test.js` files.

| Phase | Focus | Plan / tests |
|-------|--------|------|
| **68** ✅ | Workspace shell | [`phase_68_workspace_shell_spa_nav.plan.md`](plans/phase_68_workspace_shell_spa_nav.plan.md) |
| **69** ✅ | Zone inline hub | [`phase_69_zone_workspace_hub.plan.md`](plans/phase_69_zone_workspace_hub.plan.md) |
| **70** ✅ | Hardware / Pi GPIO | [`phase_70_hardware_pi_control_spa.plan.md`](plans/phase_70_hardware_pi_control_spa.plan.md) · [`phase-70-closure.md`](plans/phase-70-closure.md) |
| **71** ✅ | Feed & Water SPA | [`phase_71_feed_water_unification.plan.md`](plans/phase_71_feed_water_unification.plan.md) · [`phase-71-closure.md`](plans/phase-71-closure.md) |
| **72** ✅ | Money SPA | [`phase_72_money_unification.plan.md`](plans/phase_72_money_unification.plan.md) · [`phase-72-closure.md`](plans/phase-72-closure.md) |
| **73** ✅ | Guardian discoverability | [`phase_73_guardian_pr_discoverability.plan.md`](plans/phase_73_guardian_pr_discoverability.plan.md) · [`phase-73-closure.md`](plans/phase-73-closure.md) |
| **74** ✅ | Zone ops (Tasks, Alerts, Plants) | [`phase_74_zone_ops_inbox.plan.md`](plans/phase_74_zone_ops_inbox.plan.md) |
| **75** ✅ | Comfort & automation workspace | [`phase_75_automation_comfort_workspace.plan.md`](plans/phase_75_automation_comfort_workspace.plan.md) |
| **76** ✅ | Today + mobile nav alignment | [`phase_76_today_dashboard_nav_alignment.plan.md`](plans/phase_76_today_dashboard_nav_alignment.plan.md) |
| **77** ✅ | Post-arc polish | [`phase_77_post_arc_ui_polish.plan.md`](plans/phase_77_post_arc_ui_polish.plan.md) |
| **78** ✅ | Zone-first hardware & GPIO on alerts | `ui/src/__tests__/phase-78-closure.test.js` |
| **79** ✅ | Tasks fix, Money inventory tab, operator glossary | `ui/src/__tests__/phase-79-closure.test.js` |
| **80** ✅ | Routing hashes, zones tab labels, workspace routes | `ui/src/__tests__/phase-80-closure.test.js` |
| **81** ✅ | `/pi-setup` restore, Help Pi tab, zone hardware on Overview only | `ui/src/__tests__/phase-81-closure.test.js` |
| **82** ✅ | Guardian plant intelligence — crop library, multi-crop lookup, zero-chunk guardrail (WS7/WS11 deferred) | [`phase_82_guardian_crop_grounding_hardening.plan.md`](plans/phase_82_guardian_crop_grounding_hardening.plan.md) · [`phase-82-closure.md`](plans/phase-82-closure.md) |
| **83** ✅ | Enterprise agronomy seed pack — commons cultivator pack, `guardian-bootstrap-farm`, site-manifest hook, farm crop overrides, scheduled RAG ingest, readiness smokes | [`phase_83_enterprise_agronomy_seed_pack.plan.md`](plans/phase_83_enterprise_agronomy_seed_pack.plan.md) · [`phase-83-closure.md`](plans/phase-83-closure.md) |
| **84** ✅ | Enterprise crop catalog in Postgres — picker API, EC/stage targets, field guides DB, commons API | [`phase_84_crop_catalog_enterprise_db.plan.md`](plans/phase_84_crop_catalog_enterprise_db.plan.md) · [`phase-84-closure.md`](plans/phase-84-closure.md) · [`crop-catalog-db-cutover-runbook.md`](crop-catalog-db-cutover-runbook.md) |
| **85** ✅ | **Plants UX** — catalog dropdown, `plants.crop_key`, one slot per crop; Settings EC link | [`phase_85_catalog_bound_plants.plan.md`](plans/phase_85_catalog_bound_plants.plan.md) |
| **86** ✅ | Grow ops + **Guardian crop chain** — strip, Water/Light, `lookup_crop_targets` same as UI | [`phase_86_grow_ops_catalog_chain.plan.md`](plans/phase_86_grow_ops_catalog_chain.plan.md) |
| **87** ✅ | Guardian crop API smokes + operator runbook + OC-87 | [`phase_87_crop_knowledge_operator_closure.plan.md`](plans/phase_87_crop_knowledge_operator_closure.plan.md) · [`phase-87-closure.md`](plans/phase-87-closure.md) · [`crop-knowledge-operator-runbook.md`](crop-knowledge-operator-runbook.md) |
| **88** ✅ | Domain enums API — growth stages, reservoir, cost categories (UI drops duplicates) | [`phase_88_domain_enums_api.plan.md`](plans/phase_88_domain_enums_api.plan.md) · [`phase-88-closure.md`](plans/phase-88-closure.md) |
| **89** ✅ | Lighting presets — wire `GET /lighting-programs/presets` (API exists, UI unused) | [`phase_89_lighting_presets_api_wiring.plan.md`](plans/phase_89_lighting_presets_api_wiring.plan.md) · [`phase-89-closure.md`](plans/phase-89-closure.md) |
| **90** ✅ | Device taxonomy registry — sensor/actuator → water/light/climate + Guardian | [`phase_90_device_taxonomy_registry.plan.md`](plans/phase_90_device_taxonomy_registry.plan.md) · [`phase-90-closure.md`](plans/phase-90-closure.md) |
| **91** ✅ | Bootstrap template catalog — replace `bootstrapTemplates.js` | [`phase_91_bootstrap_template_catalog.plan.md`](plans/phase_91_bootstrap_template_catalog.plan.md) · [`phase-91-closure.md`](plans/phase-91-closure.md) |
| **92** ✅ | Zone types + greenhouse enums from API | [`phase_92_zone_greenhouse_vocabulary.plan.md`](plans/phase_92_zone_greenhouse_vocabulary.plan.md) · [`phase-92-closure.md`](plans/phase-92-closure.md) |
| **99** ✅ | CI domain parity guards — `make check-ui-domain-parity` | [`phase_99_ci_domain_parity_guards.plan.md`](plans/phase_99_ci_domain_parity_guards.plan.md) · [`phase-99-closure.md`](plans/phase-99-closure.md) |
| **96** ✅ | Grow feeding program validation — warn on stage/program mismatch | [`phase_96_grow_feeding_program_validation.plan.md`](plans/phase_96_grow_feeding_program_validation.plan.md) · [`phase-96-closure.md`](plans/phase-96-closure.md) |
| **101** ✅ | Guardian write tools — `crop_key` on create_plant | [`phase_101_guardian_write_tools_crop_key.plan.md`](plans/phase_101_guardian_write_tools_crop_key.plan.md) · [`phase-101-closure.md`](plans/phase-101-closure.md) |
| **102** ✅ | Fertigation + recipe ↔ crop_key / profile EC linkage | [`phase_102_fertigation_program_catalog_metadata.plan.md`](plans/phase_102_fertigation_program_catalog_metadata.plan.md) · [`phase-102-closure.md`](plans/phase-102-closure.md) |
| **103** ✅ | Legacy plant dedupe & backfill | [`phase_103_legacy_plant_dedupe_backfill.plan.md`](plans/phase_103_legacy_plant_dedupe_backfill.plan.md) · [`phase-103-closure.md`](plans/phase-103-closure.md) |
| **104** ✅ | Harvest analytics by crop_key | [`phase_104_harvest_analytics_by_crop_key.plan.md`](plans/phase_104_harvest_analytics_by_crop_key.plan.md) · [`phase-104-closure.md`](plans/phase-104-closure.md) |
| **105** ✅ | Catalog override audit + OC-84 | [`phase_105_catalog_audit_oc84_closure.plan.md`](plans/phase_105_catalog_audit_oc84_closure.plan.md) · [`phase-84-closure.md`](plans/phase-84-closure.md) |
| **106** ✅ | Deficiency / pest symptom catalog + `lookup_crop_symptoms` | [`phase_106_deficiency_pest_symptom_catalog.plan.md`](plans/phase_106_deficiency_pest_symptom_catalog.plan.md) · [`phase-106-closure.md`](plans/phase-106-closure.md) |
| **107** ✅ | Crop catalog photos — picker thumbnails + commons `image_url` | [`phase_107_crop_catalog_photos.plan.md`](plans/phase_107_crop_catalog_photos.plan.md) · [`phase-107-closure.md`](plans/phase-107-closure.md) |
| **108** ✅ | Commons recipe packs `crop_key` tags | [`phase_108_commons_recipe_packs_crop_key.plan.md`](plans/phase_108_commons_recipe_packs_crop_key.plan.md) · [`phase-108-closure.md`](plans/phase-108-closure.md) |
| **109** ✅ | Catalog version push notifications | [`phase_109_catalog_version_push_notifications.plan.md`](plans/phase_109_catalog_version_push_notifications.plan.md) · [`phase-109-closure.md`](plans/phase-109-closure.md) · [`enterprise-catalog-version-notifications.md`](enterprise-catalog-version-notifications.md) |
| **110** ✅ | Phase 82 formal closure audit (OC-82) | [`phase_110_phase_82_formal_closure.plan.md`](plans/phase_110_phase_82_formal_closure.plan.md) · [`phase-110-closure.md`](plans/phase-110-closure.md) · [`phase-82-closure.md`](plans/phase-82-closure.md) |

## Phases 164–171 — Visual Today farm cockpit

Hub: operator tour [§7k](operator-tour.md#7k-visual-farm-cockpit-phases-164168--shipped). **Shipped** on `main` — transforms **Today** (`/`) into the grower cockpit: spatial farm map, quick actions, attention strip, one-tap Farm counsel.

| Phase | Focus | Plan / tests |
|-------|--------|--------------|
| **164** ✅ | Living demo seed — chrysanthemum, sensor readings, gravity drip, health states | [`phase_164_demo_seed_living_farm.plan.md`](plans/phase_164_demo_seed_living_farm.plan.md) · `phase-164-closure.test.js` · `go test -run Phase164` |
| **165** ✅ | Farm layout API + background image | [`phase_165_farm_layout_api.plan.md`](plans/phase_165_farm_layout_api.plan.md) · `phase-165-closure.test.js` |
| **166** ✅ | Today visual farm canvas (desktop) | [`phase_166_today_visual_farm_canvas.plan.md`](plans/phase_166_today_visual_farm_canvas.plan.md) · `farm-canvas.test.js` |
| **167** ✅ | Mobile zone stack + quick actions | [`phase_167_mobile_stack_quick_actions.plan.md`](plans/phase_167_mobile_stack_quick_actions.plan.md) · `zone-quick-actions.test.js` |
| **168** ✅ | Checklist removal + farmer copy polish | [`phase_168_today_cleanup_polish.plan.md`](plans/phase_168_today_cleanup_polish.plan.md) · `phase-168-closure.test.js` |
| **169** ✅ | Attention strip + Guardian attention starters | [`phase_169_today_attention_cockpit.plan.md`](plans/phase_169_today_attention_cockpit.plan.md) |
| **170** ✅ | One-tap Farm counsel from Today starters | [`phase_170_today_guardian_one_tap.plan.md`](plans/phase_170_today_guardian_one_tap.plan.md) |
| **171** ✅ | Demo zone layouts in seed | [`phase_171_demo_zone_layouts_seed.plan.md`](plans/phase_171_demo_zone_layouts_seed.plan.md) |
| **173** ✅ | Large-farm Today filters + paging | [`phase_173_today_large_farm_navigation.plan.md`](plans/phase_173_today_large_farm_navigation.plan.md) · `phase-173-closure.test.js` |
| **174** ✅ | Today visual hierarchy — health header, naming, tile polish | [`phase_174_today_visual_hierarchy.plan.md`](plans/phase_174_today_visual_hierarchy.plan.md) · `phase-174-closure.test.js` |
| **175** ✅ | Farm-first actions; Guardian demoted | [`phase_175_today_farm_first_actions.plan.md`](plans/phase_175_today_farm_first_actions.plan.md) · `phase-175-closure.test.js` |
| **176** ✅ | Farm pulse in Site Strip | [`phase_176_today_farm_pulse.plan.md`](plans/phase_176_today_farm_pulse.plan.md) · `phase-176-closure.test.js` |
| **177** ✅ | First impression + arc closure | [`phase_177_today_first_impression.plan.md`](plans/phase_177_today_first_impression.plan.md) · `phase-177-closure.test.js` · `today-excellence-arc.test.js` |

**Roadmap:** [`phase_173_177_today_excellence_roadmap.plan.md`](plans/phase_173_177_today_excellence_roadmap.plan.md) — **arc complete** after Phase 177.

## Phase 172 — Field guides + documentation

| Phase | Focus | Artifacts |
|-------|--------|-----------|
| **172** ✅ | Expand demo-farm crop field guides; add **marigold** + **geranium** to catalog; phase-14 + current-state sync | `docs/field-guides/crop-*.md` · `data/crop_library.yaml` · `db/seed/crop_catalog_from_yaml.sql` |

## Master roadmap — Phases 84–110

**Locked order:** [`plans/phase_84_100_master_roadmap.plan.md`](plans/phase_84_100_master_roadmap.plan.md)

**Execute:** 85 → 86 → 87 → **93** → **101** + **103** (with 85) → 89 → 88 → 90–92 → 95–100 → **102** → **104** → **105** → **106–109** as deps allow · **110** (audit) anytime.

| Arc | Phases | Focus |
|-----|--------|--------|
| A | 84–87, 93 | Plants & crop knowledge |
| B | 88–92, 99 | UI enums & CI parity |
| C | 93–100 | Blind spots |
| D | 101–105 | Guardian writes, recipes, analytics |
| E | 106–110 | Symptoms, photos, commons, push, OC-82 |

## Phases 84–87 — Plants & crop knowledge base

**Roadmap hub:** [`plans/phase_84_87_crop_identity_roadmap.plan.md`](plans/phase_84_87_crop_identity_roadmap.plan.md) · **Sit-in:** [`workstreams/sit-in-crop-catalog-enterprise-db.md`](workstreams/sit-in-crop-catalog-enterprise-db.md)

**Plants are a first-class gr33n surface.** Zone → **Plants**, the **Plants** workspace, and **Start grow** all use a **Postgres-backed catalog dropdown** (every `crop_library.yaml` crop + EC/light/watering preview). **Settings → Crops & targets** tunes EC per farm. **Farm Guardian** must use the **same crop APIs / DB profiles** as the UI — `lookup_crop_targets` — never invented EC.

| Phase | Status | One job |
|-------|--------|---------|
| **84** ✅ | Shipped | Full catalog + picker API + field guides in Postgres |
| **85** ✅ | Shipped | Catalog-bound **plants** — dropdown only; no “strain” / free-text flooding |
| **86** ✅ | Shipped | Grow strip + Water/Light + Guardian resolve `plants.crop_key` on active cycle |
| **87** ✅ | Shipped | Guardian crop API smokes + operator runbook + OC-87 · [`phase-87-closure.md`](plans/phase-87-closure.md) |

**Picker 404?** `make migrate` · restart API · `CROP_CATALOG_SOURCE=db`.

**Prompt loop:** `phase 85 ws1`, … or `phase 85` for full phase.

## Phases 88–92 — UI static data → DB/API (archived)

**Shipped** — plans moved to [`plans/archive/`](plans/archive/) (stubs at old paths). Hub: [`archive/phase_88_92_platform_data_gaps_roadmap.plan.md`](plans/archive/phase_88_92_platform_data_gaps_roadmap.plan.md) · closure: [`phase-84-110-closure.md`](plans/phase-84-110-closure.md).

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

**Pre-flight:** `make migrate` · `make check-crop-catalog-parity` · [`crop-catalog-db-cutover-runbook.md`](crop-catalog-db-cutover-runbook.md) · **Add crops (integrator):** [`catalog-integrator-playbook.md`](catalog-integrator-playbook.md) · **Multi-site promote vs local:** [`enterprise-catalog-promotion-model.md`](enterprise-catalog-promotion-model.md).

**Pre-dev gap index (archived):** [`pre_development_gaps_index.plan.md`](plans/pre_development_gaps_index.plan.md) · **Product backlog:** [`product_backlog_operator_runtime.plan.md`](plans/product_backlog_operator_runtime.plan.md) (**shipped**)

After operator-doc edits, re-ingest Guardian RAG: **`make rag-ingest-platform-docs`** and **`make rag-ingest-field-guides`** (Phase 37 field corpus).

## Quick links

| Topic | Doc |
|--------|-----|
| **Current state (today)** | [`current-state.md`](current-state.md) |
| OpenAPI | [`openapi.yaml`](../openapi.yaml) |
| First session / bootstrap | [`first-session-after-clone.md`](first-session-after-clone.md) · [`local-operator-bootstrap.md`](local-operator-bootstrap.md) |
| Guardian architecture & QA | [`farm-guardian-architecture.md`](farm-guardian-architecture.md) · [`ci-guardian-qa.md`](ci-guardian-qa.md) |
| Pi / edge | [`pi-integration-guide.md`](pi-integration-guide.md) · [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md) |
| Crop catalog | [`crop-catalog-db-cutover-runbook.md`](crop-catalog-db-cutover-runbook.md) · [`crop-knowledge-operator-runbook.md`](crop-knowledge-operator-runbook.md) |
| Backup & upgrade | [`backup-restore-runbook.md`](backup-restore-runbook.md) · [`upgrade-guide.md`](upgrade-guide.md) |
| Security & deps | [`SECURITY.md`](../SECURITY.md) · [`vuln-allowlist.md`](vuln-allowlist.md) |
| Enterprise / commons | [`scripts/enterprise/README.md`](../scripts/enterprise/README.md) · [`commons-catalog-operator-playbook.md`](commons-catalog-operator-playbook.md) |
| **Phase archive** | [`plans/archive/README.md`](plans/archive/README.md) |
| **Full phase history** | Sections below + [`phase-84-110-closure.md`](plans/phase-84-110-closure.md) |

### Legacy per-phase quick links (pre-157)

Older per-phase rows remain in git history; use **Shipped arcs** above and [`current-state.md`](current-state.md) instead of scanning 80+ individual phase links.

## Using this in a new chat

Reference `@docs/phase-14-operator-documentation.md` for Phase 14 deliverables and the **[Phase 31 field validation index](phase-14-operator-documentation.md#phase-31-field-validation-edge)**. For **grow-environment + Guardian doc alignment**, see **[Phases 34–39](#phases-34-39--guardian-polish-grow-environment-plant-needs-ui-edge-fertigation)**. For **current UI nav**, see **[Phases 68–81](#phases-6881--spa-workspace-refactor)** and **[operator tour §7e–§7j](operator-tour.md#7e-workspaces--sidebar-jobs-tabs-inside-phase-68)**. For **first-time local setup**, use **[`local-operator-bootstrap.md`](local-operator-bootstrap.md)** (`make bootstrap-local`).
