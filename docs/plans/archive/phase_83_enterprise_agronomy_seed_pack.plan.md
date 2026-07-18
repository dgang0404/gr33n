---
name: Phase 83 — Enterprise agronomy seed pack & Guardian bootstrap
overview: >
  One-command enterprise bring-up for trustworthy Guardian: commons agronomy seed pack
  (crop profiles + field guides + optional farm EC overrides), site-manifest integration,
  guardian-bootstrap-farm ingest pipeline, scheduled operational RAG refresh, farm-admin
  crop override UI, and Guardian readiness smokes. Validates small-Ollama + seed data;
  scales to 70B without changing pipelines.
todos:
  - id: ws0-deps
    content: "WS0: Hard dependency on Phase 82 crop_library.yaml + field guides shipped"
    status: completed
  - id: ws1-commons-pack
    content: "WS1: Commons catalog gr33n-cultivator-seed-pack-v1 + migration + sample body JSON"
    status: completed
  - id: ws2-override-format
    content: "WS2: Farm agronomy override pack YAML — EC/pH deltas per crop_key; import API/script"
    status: completed
  - id: ws3-bootstrap-script
    content: "WS3: scripts/enterprise/guardian-bootstrap-farm.sh — ingest + verify chunks + smoke prompts"
    status: completed
  - id: ws4-site-manifest
    content: "WS4: site-manifest guardian_seed block; apply-site-manifest.sh calls bootstrap"
    status: completed
  - id: ws5-incremental-ingest
    content: "WS5: Scheduled incremental rag-ingest — cron doc + optional worker hook / Makefile target"
    status: completed
  - id: ws6-override-ui
    content: "WS6: Farm admin UI — view builtin profiles; create/edit farm-specific crop overrides"
    status: completed
  - id: ws7-readiness-smokes
    content: "WS7: Guardian readiness checklist + smoke_phase83 + OC-83 closure"
    status: completed
  - id: ws8-docs
    content: "WS8: enterprise README, operator-tour §6o, architecture §7.0ae, phase-14 index"
    status: completed
isProject: false
---

# Phase 83 — Enterprise agronomy seed pack & Guardian bootstrap

## Status

**Shipped.** Depends on **[Phase 82](phase_82_guardian_crop_grounding_hardening.plan.md)** crop library + field guides and **Phase 84** Postgres catalog defaults. Phase 83 packages, deploys, and keeps fresh what Phase 82 curates, plus enterprise-specific overrides and one-command bring-up.

**Closure:** **OC-83** — [`phase-83-closure.md`](phase-83-closure.md)

---

## The one job

> **A new farm goes from migrate → Guardian-ready in one integrator command — structured crop targets, ingested field guides, operational RAG on a schedule, optional site-specific EC tweaks, and smokes that prove 8B + seed data works before you buy the 70B GPU box.**

---

## Why this phase (not just Phase 82 WS0)

Phase 82 WS0 documents manual ops (`make rag-ingest-field-guides`, model floor). That is enough for a dev laptop. Enterprise integrators need:

| Gap today | Phase 83 closes |
|-----------|-----------------|
| Ingest is **manual** and easy to forget | **`guardian-bootstrap-farm.sh`** + site manifest hook |
| Recipe pack (Phase 31) ≠ agronomy pack | **`gr33n-cultivator-seed-pack-v1`** in commons catalog |
| Farm crop overrides exist in **schema** but no UI/pack format | Override YAML + Settings UI |
| Cycle notes don't reach Guardian until someone runs ingest | **Scheduled incremental ingest** doc + optional automation |
| No proof Guardian is configured correctly | **Readiness smokes** (chunks > 0, mS/cm from tools, no fake `[n]`) |
| Site manifest stops at zones + recipe pack | **`guardian_seed`** block in manifest |

**Validation strategy (operator-facing):** Run readiness smokes on **8B + full seed pack**. If structured answers and citations pass, upgrading `LLM_MODEL` to 70B is a **capacity/quality** decision — not a pipeline rewrite.

---

## Architecture (what gets deployed where)

```
┌─────────────────────────────────────────────────────────────────┐
│ Platform (repo + migrations)                                     │
│  data/crop_library.yaml (Phase 82)                               │
│  docs/field-guides/crop-*.md                                     │
│  commons: gr33n-cultivator-seed-pack-v1 (manifest + readme)      │
└────────────────────────────┬────────────────────────────────────┘
                             │ migrate + seed SQL (builtins)
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ Per farm (integrator / cron)                                     │
│  1. import agronomy seed pack (audit + optional override apply)  │
│  2. guardian-bootstrap-farm: field_guide + platform_doc ingest   │
│  3. rag-ingest operational domains (tasks, cycles, programs…)    │
│  4. readiness smokes → pass/fail report                          │
└────────────────────────────┬────────────────────────────────────┘
                             │ every grounded chat turn
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ Guardian                                                         │
│  Layer 1: Ollama weights (operator chooses 8B vs 70B)            │
│  Layer 2: pgvector chunks (field_guide + platform_doc + farm)   │
│  Layer 3: crop_profiles + overrides → lookup_crop_targets        │
│  Layer 4: live snapshot (automatic)                                │
└─────────────────────────────────────────────────────────────────┘
```

---

## WS0 — Dependencies (gate)

**Blockers — do not start Phase 83 implementation until:**

- [ ] Phase 82 **WS4a** — `data/crop_library.yaml` merged
- [ ] Phase 82 **WS4b–4d** — Tier A profiles + per-crop field guides in manifest
- [ ] Phase 82 **WS1** — zero-chunk guardrail (smokes assume honest behaviour)
- [ ] Phase 82 **WS3** — multi-crop `lookup_crop_targets` (smoke questions use compare intent)
- [ ] **Sit-in:** [crop catalog enterprise DB](../workstreams/sit-in-crop-catalog-enterprise-db.md) — WS-B migrations (`crop_catalog_*`, `agronomy_field_guides`) before seed pack points at DB not YAML/MD

Phase 83 may **start WS1 catalog body authoring in parallel** if it references stable **`catalog_version`** contracts (DB rows after sit-in WS-B, not file paths). Target schema, later encounters (platform docs, pack upgrades, re-ingest): **[sit-in-crop-catalog-enterprise-db.md](../workstreams/sit-in-crop-catalog-enterprise-db.md)**.

---

## WS1 — Commons agronomy seed pack

**Goal:** Same promotion pattern as Recipe Pack v7 — published catalog entry integrators import per farm; scripts apply structured side effects.

### Catalog entry

| Field | Value |
|-------|--------|
| **slug** | `gr33n-cultivator-seed-pack-v1` |
| **kind** | `agronomy_seed_pack` |
| **title** | gr33n Cultivator Agronomy Seed Pack v1 |
| **license_spdx** | `AGPL-3.0-or-later` (platform corpus) + attribution for extension sources in `readme_md` |

### `body` JSON shape (v1)

```json
{
  "catalog_version": 1,
  "kind": "agronomy_seed_pack",
  "readme_md": "# Cultivator seed pack v1\n…",
  "crop_library_version": 2,
  "field_guide_manifest": "docs/rag/field-guide-manifest.yaml",
  "platform_doc_manifest": "docs/rag/platform-doc-manifest.yaml",
  "builtin_profile_keys": ["cannabis", "tomato", "eggplant", "…"],
  "unsupported_keys": ["ramps", "mushroom", "in_ground_root"],
  "guardian_smoke_prompts": [
    "Compare cannabis and eggplant feed targets in mS/cm",
    "What watering style does phalaenopsis want?"
  ],
  "related_urls": []
}
```

**Migration:** `db/migrations/YYYYMMDD_phase83_cultivator_seed_pack_v1.sql` — publish row (mirror Phase 31 recipe pack pattern).

**Script:** [`scripts/enterprise/import-agronomy-seed-pack.sh`](../../scripts/enterprise/import-agronomy-seed-pack.sh)

- `--dry-run` — print POST bodies + follow-on bootstrap steps
- `--farm-ids 1,2,3` — `POST /farms/{id}/commons/catalog-imports` (farm admin JWT)
- `--apply-overrides path/to/overrides.yaml` — optional WS2 farm profile deltas after import audit

**Idempotency:** catalog import upsert per farm+entry; bootstrap ingest is delete+re-upsert per doc path (existing field-guide ingest behaviour).

**Acceptance:** Dry-run prints slug + farm ids; live run records import row and exits 0 when bootstrap delegated to WS3.

---

## WS2 — Farm agronomy override pack

**Goal:** Enterprise sites tune built-in targets without forking the global `crop_library.yaml`.

### Override file format

**New:** `data/agronomy-override-pack.example.yaml`

```yaml
version: 1
farm_slug: north-warehouse-bay3   # informational; script uses --farm-id
source: "North Warehouse agronomy committee 2026-06"
overrides:
  - crop_key: cannabis
    display_name: "Cannabis (North — flower room A)"
    stages:
      - stage: late_flower
        ec_ms_cm_min: 1.0          # builtin 1.2 → site runs lower
        ec_ms_cm_max: 1.4
        notes: "RO water, coco 70/30; never above 1.4 in summer"
  - crop_key: tomato
    stages:
      - stage: fruiting
        vpd_kpa_min: 0.9
        vpd_kpa_max: 1.2
```

**Rules:**

- Units: **mS/cm**, **kPa** VPD, **mol/m²/day** DLI — same as Phase 64/82
- Creates/updates `gr33ncrops.crop_profiles` with **`farm_id` set**, `is_builtin=false`
- `GetCropProfileByKey` already prefers farm row over builtin (existing SQL `ORDER BY is_builtin DESC` — **verify and fix if farm override must win**)

**Apply path:**

- `import-agronomy-seed-pack.sh --apply-overrides file.yaml` **or**
- `scripts/enterprise/apply-agronomy-overrides.sh --farm-id N --file file.yaml`

**API (optional v1.1):** `POST /farms/{id}/agronomy-overrides/import` — JSON body same shape; farm admin only. Defer if script-only ships first.

**Acceptance:** After apply, `lookup_crop_targets` for that farm returns override EC for `cannabis` / `late_flower`; other farms still get builtin.

---

## WS3 — `guardian-bootstrap-farm.sh`

**Goal:** Single command from repo root after API + DB up.

**New:** [`scripts/enterprise/guardian-bootstrap-farm.sh`](../../scripts/enterprise/guardian-bootstrap-farm.sh)

```bash
# Plan
./scripts/enterprise/guardian-bootstrap-farm.sh --dry-run --farm-id 1

# Execute (requires .env: DATABASE_URL, EMBEDDING_API_KEY; API optional for smokes)
./scripts/enterprise/guardian-bootstrap-farm.sh --farm-id 1 [--skip-operational] [--smoke]
```

### Steps (in order)

| Step | Action | Skip flag |
|------|--------|-----------|
| 1 | Assert Phase 82 builtins present (`SELECT count(*) FROM gr33ncrops.crop_profiles WHERE is_builtin`) ≥ 25 | — |
| 2 | `make rag-ingest-field-guides` (or `go run ./cmd/rag-ingest -field-guides -farm-id N`) | `--skip-field-guides` |
| 3 | `make rag-ingest-platform-docs` | `--skip-platform-docs` |
| 4 | Operational domains: tasks, crop-cycles, programs, alerts (reuse `rag-ingest-demo.sh` domains or explicit flags) | `--skip-operational` |
| 5 | Count chunks: `field_guide`, `platform_doc`, operational `source_type`s for farm | — |
| 6 | Optional `--smoke`: HTTP `POST /v1/chat` with farm JWT + fixed prompts from seed pack | `--no-smoke` |

### Verify output (human-readable report)

```
Guardian bootstrap — farm_id=1
  crop_profiles (builtin): 27 OK
  rag chunks field_guide:  18 OK (min 12)
  rag chunks platform_doc: 42 OK
  rag chunks operational:  0 WARN (no cycles yet — expected on greenfield)
  embedding: configured OK
  llm: llama3.1:8b @ http://ollama.farm.local:11434/v1 OK
  smoke "compare cannabis eggplant EC": lookup fired, 0 fake [n], EC in mS/cm PASS
```

**Makefile target:** `make guardian-bootstrap-farm FARM_ID=1` → wraps script.

**Acceptance:** `--dry-run` prints steps; live run with embeddings configured yields `field_guide` chunk count > 0; `--smoke` fails if Phase 82 guardrail detects fake citations at 0 chunks.

---

## WS4 — Site manifest integration

**Extend:** [`scripts/enterprise/site-manifest.example.yaml`](../../scripts/enterprise/site-manifest.example.yaml)

```yaml
# Phase 83 — optional Guardian agronomy bootstrap (runs after farm + zones exist).
guardian_seed:
  enabled: true
  agronomy_pack_slug: gr33n-cultivator-seed-pack-v1
  overrides_file: null          # or path to WS2 YAML on integrator laptop
  bootstrap:
    ingest_field_guides: true
    ingest_platform_docs: true
    ingest_operational: true
    run_smokes: true
  llm_model_hint: llama3.1:8b-instruct-q4_K_M   # informational; not applied by script
```

**Extend:** [`scripts/enterprise/apply-site-manifest.sh`](../../scripts/enterprise/apply-site-manifest.sh)

After farm + zones + recipe pack:

1. If `guardian_seed.enabled`: call `import-agronomy-seed-pack.sh`
2. Call `guardian-bootstrap-farm.sh --farm-id …` with manifest flags
3. Print readiness summary + link to [`recommended-hardware-and-sizing.md`](../recommended-hardware-and-sizing.md)

**Acceptance:** `--dry-run` on example manifest prints Guardian block; apply path documented even if smokes require manual JWT export in v1.

---

## WS5 — Scheduled operational RAG refresh

**Goal:** Farm admin notes in cycles/tasks reach Guardian without someone remembering ingest.

### v1 (documentation + Makefile — ship first)

**New section in:** [`docs/local-operator-bootstrap.md`](../local-operator-bootstrap.md) + [`scripts/enterprise/README.md`](../../scripts/enterprise/README.md)

Example cron (API host, `/etc/cron.d/gr33n-rag-ingest`):

```cron
# Incremental farm operational RAG — every 6h
0 */6 * * * gr33n cd /opt/gr33n-platform && set -a && . ./.env && set +a && \
  go run ./cmd/rag-ingest -farm-id 1 -tasks -crop-cycles -programs -alerts \
  -updated-after "$(cat /var/lib/gr33n/rag-watermark-farm-1 2>/dev/null || echo 1970-01-01T00:00:00Z)" \
  && date -u +%Y-%m-%dT%H:%M:%SZ > /var/lib/gr33n/rag-watermark-farm-1
```

**Makefile:** `make rag-ingest-farm-operational FARM_ID=1`

Document: multi-farm loops, `EMBEDDING_API_KEY` requirement, watermark file per farm.

### v1.1 (optional code — defer if cron doc sufficient)

- Automation worker tick: `rag_ingest_operational` job type per farm with `farm_active_modules` flag
- Or post-commit hook on `crop_cycles.cycle_notes` update → enqueue ingest for that source_id only

**Acceptance:** Doc merged; cron example copy-paste safe; incremental `-updated-after` verified against existing `cmd/rag-ingest` flags.

---

## WS6 — Farm admin crop override UI

**Goal:** Agronomists edit site EC/VPD/DLI without YAML on integrator laptop.

### UI surface

**Location:** Settings → **Crops & targets** (new card) or Plants workspace → **Crop profiles**

| Feature | Behaviour |
|---------|-----------|
| List | Builtins (read-only badge) + farm overrides (editable) |
| Create override | Pick builtin `crop_key` → clone stages → edit numeric targets + notes |
| Reset | Delete farm override row → revert to builtin |
| Units | Labels enforce mS/cm, kPa, mol/m²/day (Phase 64 parity) |

### API (extend existing crop profile handler)

- `GET /farms/{id}/crop-profiles` — already lists builtins + farm rows
- `PUT /farms/{id}/crop-profiles/{crop_key}` — upsert farm override (farm admin)
- `DELETE /farms/{id}/crop-profiles/{crop_key}` — remove override

OpenAPI + farm authz: **`RequireFarmAdmin`** for mutations; **`RequireFarmMember`** for read.

**Acceptance:** Override visible in UI; Guardian `lookup_crop_targets` returns farm EC on next chat turn (no re-ingest needed — structured DB path).

---

## WS7 — Guardian readiness checklist & smokes

### Operator checklist (printed by bootstrap + operator tour)

- [ ] `AI_ENABLED=true`, `LLM_BASE_URL` + `LLM_MODEL` probe OK
- [ ] `EMBEDDING_API_KEY` set (or documented LAN embedder)
- [ ] `make guardian-bootstrap-farm FARM_ID=N` exit 0
- [ ] Field guide chunk count ≥ 12 for farm
- [ ] Ask: *"Compare cannabis and eggplant EC targets"* → metadata shows **chunks > 0** OR tool block with mS/cm (both acceptable post-WS1)
- [ ] Ask: *"How should I feed ramps?"* → unsupported / cousin (Phase 82 WS4e), not 12/12 cannabis schedule
- [ ] Optional: repeat smokes after switching to 70B — same pass criteria, faster/cleaner prose

### Automated tests

**New:** `cmd/api/smoke_phase83_test.go` (build tag `smoke` or default integration)

- Seed farm with builtins + mock embedder if CI lacks key
- Assert bootstrap SQL paths / chunk counts / handler guardrail invariants

**Vitest (optional):** Settings crop override form smoke.

**Closure artifact:** `docs/plans/phase-83-closure.md` + **OC-83** row in phase-14 index.

---

## WS8 — Documentation

| Doc | Change |
|-----|--------|
| [`scripts/enterprise/README.md`](../../scripts/enterprise/README.md) | WS1–WS5 tools, quick start, idempotency |
| [`docs/hypothetical-enterprise-topology.md`](../hypothetical-enterprise-topology.md) | § Guardian seed pack on new warehouse |
| [`docs/farm-guardian-architecture.md`](../farm-guardian-architecture.md) | §7.0ae Enterprise agronomy bootstrap |
| [`docs/operator-tour.md`](../operator-tour.md) | §6 callout + §6o bootstrap walkthrough |
| [`docs/guardian-real-grow-readiness.md`](../guardian-real-grow-readiness.md) | **New** — live-plant checklist, public-demo honesty |
| [`docs/commons-catalog-operator-playbook.md`](../commons-catalog-operator-playbook.md) | Agronomy pack kind + import semantics |
| [`docs/phase-14-operator-documentation.md`](../phase-14-operator-documentation.md) | Phase 83 row |
| [`docs/plans/archive/phase_82_guardian_crop_grounding_hardening.plan.md`](phase_82_guardian_crop_grounding_hardening.plan.md) | Out-of-scope → link Phase 83 |

---

## Implementation order

1. **WS0** — confirm Phase 82 gates
2. **WS1** — catalog migration + import script stub
3. **WS3** — bootstrap script (highest integrator value)
4. **WS7** — smokes + checklist (locks behaviour)
5. **WS4** — site manifest hook
6. **WS2** — override YAML + apply script
7. **WS5** — cron doc + Makefile
8. **WS6** — override UI
9. **WS8** — docs sweep

---

## Out of scope (Phase 84+)

| Topic | Why defer |
|-------|-----------|
| Per-genetics ML / auto-tune from harvest | needs data volume + training |
| Operator-uploaded PDF plant notes RAG | Phase 53 roadmap |
| 500-site Ansible/Terraform fleet provisioner | community extension beyond manifest stub |
| Certified pest ID / pesticide prescriptions | regulatory |
| Auto-ingest on every DB write | WS5 v1.1 worker — only if cron insufficient |

---

## Related

| Doc | Use |
|-----|-----|
| [phase_82_guardian_crop_grounding_hardening.plan.md](phase_82_guardian_crop_grounding_hardening.plan.md) | Crop library + field guides source |
| [phase_33_guardian_polish_and_enterprise_ops.plan.md](phase_33_guardian_polish_and_enterprise_ops.plan.md) | Site manifest pattern |
| [phase_31_field_validation_and_edge.plan.md](phase_31_field_validation_and_edge.plan.md) | Recipe pack v7 precedent |
| [phase_64_crop_knowledge_base.plan.md](phase_64_crop_knowledge_base.plan.md) | crop_profiles schema |
| [recommended-hardware-and-sizing.md](../recommended-hardware-and-sizing.md) | 8B vs 70B sizing |
| [farm-guardian-architecture.md](../farm-guardian-architecture.md) | Three knowledge layers |
| [rag-scope-and-threat-model.md](../rag-scope-and-threat-model.md) | What may be embedded |
| [scripts/enterprise/README.md](../../scripts/enterprise/README.md) | Enterprise script home |

---

## Using this in a new chat

> Read `docs/plans/archive/phase_83_enterprise_agronomy_seed_pack.plan.md`. Build enterprise Guardian bootstrap: commons `gr33n-cultivator-seed-pack-v1`, `guardian-bootstrap-farm.sh`, site-manifest `guardian_seed` block, agronomy override YAML, scheduled ingest doc, crop override UI. Depends on Phase 82 crop library. Smokes prove 8B + seed data before 70B upgrade.
