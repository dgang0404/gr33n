# Hypothetical enterprise topology (multi-site vertical farms)

**Status:** Starter sketch — **not a product commitment**, **not required software changes**.  
**Audience:** Operators, integrators, and investors asking *"Could gr33n run 500 Costco-scale warehouses with frontier sites?"*

This document maps **existing** gr33n nouns (organization, farm, zone, fertigation program, commons catalog, Pi edge) onto a large deployment **without** rewriting the platform. Treat it as a thought experiment and integration guide.

**Companion docs:** [`raspberry-pi-and-deployment-topology.md`](raspberry-pi-and-deployment-topology.md), [`pi-integration-guide.md`](pi-integration-guide.md), [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md), [`commons-catalog-operator-playbook.md`](commons-catalog-operator-playbook.md), [`enterprise-catalog-version-notifications.md`](enterprise-catalog-version-notifications.md), [`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md), [`offline-or-intranet-deployment.md`](offline-or-intranet-deployment.md), [`phase-14-operator-documentation.md`](phase-14-operator-documentation.md#phase-31-field-validation-edge) (Phase 31 operator index), [`scripts/enterprise/README.md`](../scripts/enterprise/README.md).

---

## Disclaimer

- gr33n today targets **homestead → small/multi-site** ops with **operator sovereignty** (on-prem, AGPL, confirm-before-write Guardian).
- A **500-warehouse** operator would need serious **DevOps, networking, provisioning, and change control** on top of this app — the same way Linux in a data center needs Ansible, not just a kernel.
- Nothing here promises performance at that scale without measurement. This doc answers *"what would it look like?"*, not *"we guarantee it."*

---

## Mental model (no new tables required)

| Physical thing | gr33n mapping |
|----------------|---------------|
| Company / holding entity | **`organizations`** |
| One warehouse building (Costco-sized) | **`farms`** row (one `farm_id` per site is the usual convention) |
| Plastic grow room / module | **`zones`** |
| Vertical tier (3 shelves up) | **3 zones** *or* 3 sensor groups inside one zone — pick one naming convention and stick to it |
| Room controller | **`devices`** + edge Pi |
| EC / pH / humidity / PAR | **`sensors`** → `POST /sensors/{id}/readings` or MQTT batch |
| Pump / valve / lights | **`actuators`** → `pending_command` + `POST /actuators/{id}/events` |
| Nutrient recipe + targets | **`fertigation_programs`** (+ linked schedules / rules) |
| Corporate policy change | **Audit** (`GET /farms/{id}/audit-events`, org rollup) + optional **commons pack** import |

Guardian (Phase 29) stays **per-farm**: one chat thread does not silently operate all 500 sites. That is intentional (human-in-the-loop, RBAC, audit).

---

## Topology A — Central HQ (one database, many farms)

Best when sites have **reliable VPN/WAN** to headquarters and you want one operator pane for all warehouses.

```
                    ┌─────────────────────────────┐
                    │  HQ: Postgres + API + UI    │
                    │  One org, 500 farm records  │
                    └──────────────┬──────────────┘
                                   │ VPN / private WAN
         ┌─────────────────────────┼─────────────────────────┐
         ▼                         ▼                         ▼
   Warehouse #1              Warehouse #247           Frontier #500
   Pi(s) + MQTT              Pi(s) + MQTT             Pi(s) + MQTT
   rooms → zones             rooms → zones            offline queue
```

### How it works with today's software

1. **Provision** — create org + 500 farms (script `POST /farms` or SQL seed for pilots).
2. **Per site** — apply a **farm template** (zones, sensor/actuator placeholders, inactive rules) via onboarding patterns in [`plans/phase_15_farm_onboarding.plan.md`](plans/phase_15_farm_onboarding.plan.md).
3. **Edge** — each warehouse runs [`pi_client/gr33n_client.py`](../pi_client/gr33n_client.py) or [`mqtt_telemetry_bridge.py`](../pi_client/mqtt_telemetry_bridge.py) pointing at `api.base_url` on the LAN/VPN; auth via shared **`PI_API_KEY`** (or split keys + multiple API deployments if you outgrow one secret).
4. **Operate** — managers use the dashboard **farm selector**; alerts, tasks, fertigation, and Guardian are **scoped to the selected farm**.
5. **Telemetry volume** — Timescale hypertables for readings; retention, partitioning, and read replicas are **operator infrastructure**, not app features.

### Recipe / program updates ("push v7 everywhere")

There is **no built-in "broadcast to all farms" button**. Hypothetical promotion paths **using existing APIs**:

| Method | When to use |
|--------|-------------|
| **Golden template farm** | Maintain master programs on `farm_id=template`; copy via script calling fertigation/schedule APIs per target farm |
| **Commons catalog import** | Publish **`Recipe Pack v7`** as a catalog entry; each farm admin runs `POST /farms/{id}/commons/catalog-imports` (records provenance; body is JSON for tools — see [`commons-catalog-operator-playbook.md`](commons-catalog-operator-playbook.md)) |
| **Insert Commons export bundle** | Federation-style package for reviewed aggregates or config snapshots — [`insert-commons-pipeline-runbook.md`](insert-commons-pipeline-runbook.md) |
| **Custom deployment pipeline** | Ansible/Terraform + your scripts under [`scripts/enterprise/`](../scripts/enterprise/README.md) (community contributions welcome via PR) |

Change control: every write should leave an **audit trail**; rollouts are staged (canary farms → region → global).

---

## Topology B — Frontier autonomy (local stack per site)

Best when **frontier sites** must run **offline** or cannot depend on HQ uptime. Aligns with gr33n's *"don't call home"* posture.

```
HQ ── publishes "Recipe Pack v7" ──► Commons catalog / export bundle
                                              │
                    each site imports when ready (dashboard or script)
                                              ▼
         ┌──────────────────┐     ┌──────────────────┐
         │ Site A           │     │ Frontier Site B  │
         │ local API+DB+UI  │     │ local API+DB+UI  │
         │ works offline    │     │ syncs when link  │
         └──────────────────┘     └──────────────────┘
```

### How it works with today's software

1. **Each site** — full stack (Postgres + API + built UI) on a NUC or small server; Pis stay **edge-only** ([`raspberry-pi-and-deployment-topology.md`](raspberry-pi-and-deployment-topology.md)).
2. **HQ** — curates **commons catalog entries** or export bundles; does **not** need live RPC into every DB.
3. **Update** — when a link exists, site operator or script **imports** the pack; automation picks up new program references on the next schedule/rule evaluation cycle.
4. **Guardian** — local Ollama per site (or regional inference host on VLAN); no cloud LLM required.

Multi-master **live sync** between 500 Postgres instances is **not** in scope today. Eventual consistency via **packages + import** matches the codebase as shipped.

---

## What already works vs what you bring

| Capability | In gr33n today | At 500× scale you add |
|------------|----------------|------------------------|
| Multi-farm UI + org audit | ✅ | SSO, RBAC roles per region |
| Zones / sensors / actuators / rules | ✅ | Device provisioning at volume |
| Fertigation programs | ✅ | Promotion pipeline (scripts) |
| Pi HTTP + offline queue | ✅ | Fleet monitoring, key rotation |
| MQTT batch ingest | ✅ | Broker HA, topic conventions — [room-scale pattern](mqtt-edge-operator-playbook.md#room-scale-warehouse-pattern-phase-31-ws4) (Phase 31 WS4) |
| Commons catalog import | ✅ | Curator workflow for packs |
| Guardian confirm actions | ✅ (Phase 30 PR queue) | Per-site LLM capacity planning |
| Guardian live zone/alert reads | ✅ (Phase 31 WS6) | Same per-farm scope as chat |
| Instant global recipe push | ❌ | Scripts / MES layer |
| Guardian → actuators | Phase 30 PR (`enqueue_actuator_command` → Confirm) | Phase 31 bench proves Pi executes — [`pi-integration-guide.md` §9](pi-integration-guide.md#9-safe-actuator-e2e--pending_command-round-trip-phase-31-ws3) |

---

## Deployment pipeline scripts (community extension point)

Large integrators will eventually want **repeatable** site bring-up:

- Create farm + zones from a YAML manifest  
- Register devices/sensors/actuator IDs  
- Deploy `pi_client/config.yaml` from template  
- Import commons pack version pin  
- Smoke: `GET /health`, one reading POST, one pending_command round-trip  

**Repository convention:** optional helpers live under [`scripts/enterprise/`](../scripts/enterprise/README.md). Phase 33 WS5 adds a **`site-manifest.yaml`** bring-up stub (farm + zones + recipe pack pin). **Phase 83 (shipped)** adds **`guardian_seed`** — commons agronomy pack import, **`guardian-bootstrap-farm.sh`** RAG ingest, farm crop overrides (YAML or **Settings → Crops & targets**), scheduled operational ingest, and readiness smokes ([`phase_83_enterprise_agronomy_seed_pack.plan.md`](plans/phase_83_enterprise_agronomy_seed_pack.plan.md) · [`phase-83-closure.md`](plans/phase-83-closure.md)). The core team does not need to ship a full 500-site suite for the platform to be valid.

### AGPL and pull requests (why this matters)

gr33n is **[AGPL v3](../LICENSE)**. If an integrator modifies the **platform software** and runs it as a network service for users, copyleft obligations apply. In practice:

- **Config, YAML manifests, Ansible, and deployment scripts** that only *call* the public API are usually **your ops artifacts** — contribute them back if you want, but they are not necessarily "derived work" of the Go/Vue codebase. (Not legal advice; counsel for your jurisdiction.)
- **Forks of `cmd/api`, UI patches, or embedded proprietary modules linked into gr33n** — those trigger AGPL sharing requirements when users interact over a network.

A Fortune-scale deployment that **customizes the platform** without publishing sources is a **compliance risk for them**, not a feature request for us. Conversely, integrators who **upstream** deployment tooling via pull request strengthen the commons — good advertising for ethical open source at enterprise scale.

---

## Comparison snapshot (not a sales pitch)

| | Big Ag / OEM cloud (FieldView, Operations Center, vertical MES vendors) | gr33n at hypothetical 500 sites |
|--|--|--|
| Control | Vendor cloud, prescriptions, dealer lock-in | Your DB, your LAN, your packs |
| Recipe rollout | Central push, often proprietary | Catalog import + [`scripts/enterprise/import-recipe-pack.sh`](../scripts/enterprise/import-recipe-pack.sh) (Phase 31 WS5 demo) |
| Edge | Certified controllers | Pi + your wiring |
| Scale polish | Decades of enterprise SE | Integrator + [`scripts/enterprise/`](../scripts/enterprise/README.md) |
| Guardian | N/A or black-box automation | Propose → confirm → audit |

gr33n wins on **sovereignty and transparency**; it does not try to out-Deere Deere on fleet telematics.

---

## Suggested reading order for integrators

1. [`local-operator-bootstrap.md`](local-operator-bootstrap.md) — one laptop demo  
2. [`pi-integration-guide.md`](pi-integration-guide.md) — close the Pi loop (§8 field checklist, §9 actuator bench)  
3. [`plans/phase_30_guardian_change_requests.plan.md`](plans/phase_30_guardian_change_requests.plan.md) — **Guardian PR queue** (config + Pi via confirm, not autonomous)  
4. [`plans/phase_31_field_validation_and_edge.plan.md`](plans/phase_31_field_validation_and_edge.plan.md) — **field / Pi validation** (shipped artifacts indexed in [`phase-14-operator-documentation.md` § Phase 31](phase-14-operator-documentation.md#phase-31-field-validation-edge))  
5. This doc — scale-out thought experiment  
6. [`commons-catalog-operator-playbook.md`](commons-catalog-operator-playbook.md) — recipe pack provenance; Phase 31 WS5 [`import-recipe-pack.sh`](../scripts/enterprise/import-recipe-pack.sh) demo  
7. [`mqtt-edge-operator-playbook.md`](mqtt-edge-operator-playbook.md#room-scale-warehouse-pattern-phase-31-ws4) — room-scale MQTT ingest for multi-zone warehouses  
8. [`plans/phase_33_guardian_polish_and_enterprise_ops.plan.md`](plans/phase_33_guardian_polish_and_enterprise_ops.plan.md) — site manifest + read-tool polish  
9. [`plans/phase_83_enterprise_agronomy_seed_pack.plan.md`](plans/phase_83_enterprise_agronomy_seed_pack.plan.md) — **shipped** — Guardian bootstrap + agronomy pack on new warehouse ([`scripts/enterprise/README.md`](../scripts/enterprise/README.md))

**README entry point:** [Phase 31 row](../README.md) in the roadmap table links here for multi-site integrators.

---

## Changelog

| Date | Note |
|------|------|
| 2026-06-12 | Phase 83 shipped — `guardian_seed`, bootstrap script, crop override UI |
| 2026-05-27 | Phase 33 plan — site manifest WS5, read-tool hardening; reading order step 8 |
| 2026-05-27 | Phase 31 WS7 — cross-links to phase-14 index, README roadmap, Guardian read tools row |
| 2026-05-26 | Starter sketch — central HQ vs frontier autonomy, commons promotion, AGPL/PR note, `scripts/enterprise/` hook |
