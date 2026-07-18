---
name: Phase 32 — Guardian grow setup PRs
overview: >
  Close the gap between "Guardian explains setup" and "Guardian opens a change request
  you can Confirm." Enable conversational, multi-step grow onboarding PRs — add plant,
  start cycle in zone, create/link fertigation program — without autonomous writes.
  Reuses Phase 30 PR queue; extends tools, snapshot reads, proposal generation, and
  platform doc RAG (WS8) so Guardian can answer how-to questions from the operator corpus.
todos:
  - id: ws1-read-layer
    content: "WS1: Read layer — expand live snapshot (plants, programs per zone); read-only tools list_plants, summarize_zone_fertigation; optional route context from UI"
    status: done
  - id: ws2-create-tools
    content: "WS2: Create tools — create_plant, create_crop_cycle, create_fertigation_program (allowlisted fields); map to existing handlers; medium/high risk tiers"
    status: done
  - id: ws3-setup-pack-pr
    content: "WS3: Setup pack PR — single frozen bundle (plant + cycle + program) or linked proposal group; one Confirm card with step diff; transactional execute"
    status: done
  - id: ws4-intent-generation
    content: "WS4: Intent → proposal — structured extraction / setup templates from chat (beyond Phase 29 regex); house-plant vs commercial zone profiles"
    status: done
  - id: ws5-confirm-ux
    content: "WS5: Confirm UX — SetupPackProposalCard; show created rows preview; dismiss partial bundle rules"
    status: done
  - id: ws6-operator-docs
    content: "WS6: Operator docs — what Guardian can propose vs manual UI; house plant example; link farm-guardian-architecture § blind spots"
    status: done
  - id: ws7-openapi-tests
    content: "WS7: OpenAPI + smokes — setup pack propose→confirm; Vitest bundle card; no silent partial apply on validation failure"
    status: done
  - id: ws8-knowledge-depth
    content: "WS8: Guardian knowledge depth — curated platform docs/ RAG corpus + ingest script; persona cites corpus for how-to; farm-scoped chunks for operator Q&A"
    status: done
isProject: false
---

# Phase 32 — Guardian grow setup PRs

## Status

**Shipped.** All eight workstreams complete. Phase 32 delivered the full Guardian grow-setup PR flow: live snapshot expanded with plants and fertigation programs (WS1), three write tools (`create_plant`, `create_crop_cycle`, `create_fertigation_program`) registered with medium/high risk tiers (WS2), transactional `apply_grow_setup_pack` bundle tool (WS3), rule-assisted intent matching for setup phrases (WS4), `SetupPackProposalCard` confirm UX (WS5), operator docs updated (WS6), smoke tests and OpenAPI coverage (WS7), and platform RAG corpus ingest script with docs indexed (WS8). See also **Phase 33** for Polish & Enterprise Ops built on top of this phase.

**Preconditions:**

- [`guardian_action_proposals`](../../db/migrations/20260521_phase29_guardian_proposals.sql) + Confirm path ([`internal/handler/chat/confirm.go`](../../internal/handler/chat/confirm.go))
- Existing REST handlers (reuse, do not duplicate):
  - `POST /farms/{id}/plants` — [`internal/handler/plants/handler.go`](../../internal/handler/plants/handler.go)
  - `POST /farms/{id}/crop-cycles` — [`internal/handler/cropcycle/handler.go`](../../internal/handler/cropcycle/handler.go)
  - `POST /farms/{id}/fertigation/programs` — [`internal/handler/fertigation/handler.go`](../../internal/handler/fertigation/handler.go)
  - `PATCH /fertigation/programs/{id}` — Phase 30 `patch_fertigation_program` (existing program only)
- Phase 15 farm **bootstrap templates** remain the whole-farm path — this phase is **per-grow conversational setup**, not replacing [`phase_15_farm_onboarding.plan.md`](phase_15_farm_onboarding.plan.md)

---

## Why this phase

Phase 30 answered: *"Can Guardian open a **single** reviewed change (task, alert ack, patch program, Pi enqueue)?"*

Operators still ask: *"Can Guardian **set up this plant** — add it to my zone and wire fertigation — from one conversation?"*

**Today the answer is no:**

| Operator ask | Phase 30 capability |
|--------------|---------------------|
| Ack humidity alert | ✅ `ack_alert` (rule-assisted) |
| Create follow-up task | ✅ `create_task` |
| Patch EC on **existing** program | ✅ tool exists; weak chat→proposal matching |
| **Create** plant | ❌ no tool |
| **Create** crop cycle in zone | ❌ no tool |
| **Create** fertigation program | ❌ no tool (`patch_*` only) |
| One Confirm for plant + cycle + program | ❌ one PR = one tool |

Phase 32 makes **grow setup** a first-class **PR bundle** — still **never autonomous**.

---

## Problem statement (blind spots)

Guardian reads a **curated** subset of farm state (see [`farm-guardian-architecture.md`](../farm-guardian-architecture.md)):

- Live snapshot: zones, active cycles (+ EC/pH rollups), top unread alerts
- RAG: indexed operational text (optional)
- `context_ref`: alert / cycle / zone when opened from **Ask Guardian**

It does **not** mirror everything the UI shows (plants list, program catalog, live sensor tiles, current route). Phase 32 WS1 reduces blind spots for **setup** flows before WS2–WS4 propose writes.

---

## Design principles

1. **Human Confirm always** — same Phase 30 gate; no setup autopilot.
2. **Reuse handlers** — tools call the same sqlc/handler logic as dashboard POST/PATCH.
3. **Frozen args at propose time** — bundle JSON stored in `guardian_action_proposals.args`; Confirm replays server copy.
4. **Atomic setup packs (v1 goal)** — Confirm applies **all steps** or **none** (transaction); no silent half-setup.
5. **Allowlisted fields** — create tools expose minimal required columns; no arbitrary JSON blobs from the LLM.
6. **Risk tiers** — single-step creates = **medium**; bundles touching new cycle + new program + schedule link = **high** (warning copy).
7. **Advisory LLM, deterministic execute** — model may draft; server validates IDs, farm scope, zone uniqueness (one active cycle per zone).
8. **Not a replacement for Phase 15 templates** — whole-farm bootstrap stays `apply_bootstrap_template` (farm admin, high tier).

---

## Architecture (setup pack)

```
Operator chat ("add philodendron to Living Room, RO water, light fertigation")
        │
        ▼
  WS4 intent → setup template (house_plant | zone_cycle_program)
        │
        ▼
  WS1 read snapshot (zones, existing plants/programs — avoid duplicates)
        │
        ▼
  Build SetupPack args (frozen JSON)
        │
        ▼
  INSERT guardian_action_proposals  tool=apply_grow_setup_pack  status=pending
        │
        ├──► Chat proposal card + inbox (Phase 30)
        │
        ▼
  Operator Confirm
        │
        ▼
  tools.Execute (transaction):
    1. create_plant (if step present)
    2. create_crop_cycle (zone_id, strain from plant)
    3. create_fertigation_program (target_zone_id, conservative EC/pH defaults)
    4. optional: link primary_program_id on cycle; optional create_task "monitor first week"
        │
        ▼
  audit guardian_tool_executed + return created ids in confirm response
```

**Alternative v1 (simpler, worse UX):** three separate proposals the operator Confirms in sequence — document as fallback if bundle transaction slips schedule.

---

## Scope

| WS | Focus | Primary artifacts |
|----|-------|-------------------|
| **WS1** | Read layer | Snapshot + read tools; optional UI `route` in `context_ref` |
| **WS2** | Create tools | `internal/farmguardian/tools/plants.go`, `cycles.go`, `programs_create.go` |
| **WS3** | Setup pack PR | `apply_grow_setup_pack` tool + migration if `proposal_group_id` needed |
| **WS4** | Intent generation | `matchSetupPackIntent` + templates; optional LLM JSON schema pass |
| **WS5** | Confirm UX | `SetupPackProposalCard.vue`; diff summary |
| **WS6** | Docs | Operator guide + architecture blind-spot table update |
| **WS7** | Tests | Smokes: philodendron bundle; rollback on zone conflict |
| **WS8** | Knowledge depth | Platform `docs/` RAG pack + ingest script; persona/RAG instructions |

---

## Work-stream detail

### WS1 — Read layer (see before proposing)

**Goal:** Guardian knows what already exists so it does not propose duplicate plants or conflicting cycles.

**Tasks:**

- Extend [`BuildSnapshot`](../../internal/farmguardian/snapshot.go):
  - `PlantsSummary` — count + names (cap N)
  - `ProgramsByZone` — active program names per zone (cap N)
- Optional read tools (Confirm N/A — invoked server-side during propose only, or exposed later for WS6 parity with Phase 31):
  - `list_plants` — farm-scoped
  - `summarize_zone` — zone + active cycle + programs + setpoint hints
- Optional UI: send `context_ref: { type: 'route', path: '/fertigation' }` from Vue router (honesty fix; not required for WS3 execute)

**Acceptance:** Snapshot block includes plant names when farm has plants; setup propose skips create when display_name already exists (configurable).

---

### WS2 — Create tools (single-step building blocks)

**Goal:** Each REST create has a Guardian tool with validated args.

| Tool ID | Maps to | Tier | Required args (v1) |
|---------|---------|------|---------------------|
| `create_plant` | `POST /farms/{id}/plants` | medium | `display_name`, optional `variety_or_cultivar`, optional `meta` |
| `create_crop_cycle` | `POST /farms/{id}/crop-cycles` | medium | `zone_id`, `name`, `strain_or_variety`, `current_stage`, `started_at` |
| `create_fertigation_program` | `POST /farms/{id}/fertigation/programs` | medium | `name`, `target_zone_id`, `total_volume_liters`, EC/pH triggers, `is_active` |

**Notes:**

- `gr33ncrops.plants` is **not** FK-linked to crop cycles today — setup pack should document convention (matching names / meta `plant_id`) or add a follow-up migration if product wants hard linkage.
- Respect **one active cycle per zone** — return clear error at propose validation.
- Reuse [`tools/args.go`](../../internal/farmguardian/tools/args.go) farm-scope checks.

**Acceptance:** Each tool passes smoke via Confirm; audit row written.

---

### WS3 — Setup pack PR (`apply_grow_setup_pack`)

**Goal:** One change request covers a conversational "full setup."

**Frozen args shape (illustrative):**

```json
{
  "profile": "house_plant",
  "zone_id": 12,
  "zone_name": "Living Room",
  "plant": {
    "display_name": "Philodendron",
    "variety_or_cultivar": "heartleaf",
    "notes": "RO water only; ice cubes stopped"
  },
  "cycle": {
    "name": "Philodendron — Living Room",
    "current_stage": "vegetative",
    "started_at": "2026-05-27"
  },
  "program": {
    "name": "Philodendron light feed",
    "total_volume_liters": 0.5,
    "ec_trigger_low": 0.8,
    "ph_trigger_low": 5.8,
    "ph_trigger_high": 6.5,
    "is_active": true
  },
  "optional_task": {
    "title": "Monitor new philodendron — first two weeks"
  }
}
```

**Execute:** single DB transaction calling WS2 executors in order; link `primary_program_id` on cycle after program create.

**Tier:** **high** — Confirm shows warning: *creates plant, cycle, and fertigation program*.

**Acceptance:** Confirm → Plants page shows plant; zone has active cycle; Fertigation lists new program; rollback if zone already has active cycle.

---

### WS4 — Intent → proposal generation

**Goal:** Chat message like *"add my philodendron to Living Room with a light fertigation program"* opens a setup pack PR.

**v1 approach (deterministic, like Phase 29):**

- Keyword + snapshot matchers: `add`/`create` + `plant` + zone name from snapshot
- Template defaults by `profile`: `house_plant` (low EC, small volume) vs `commercial_zone` (stricter validation)

**v2 (optional within phase):**

- LLM structured output → validated JSON → same frozen args (never trust raw model for IDs — resolve zone by name server-side)

**Not in scope:** open-ended "do whatever the model thinks" without template validation.

**Acceptance:** Demo phrase produces one pending setup pack; nonsense zone name → no proposal + chat explains available zones.

---

### WS5 — Confirm UX

**Goal:** Operator sees **what will be created** before Confirm.

**Tasks:**

- Extend [`GuardianActionProposal.vue`](../../ui/src/components/GuardianActionProposal.vue) or new `SetupPackProposalCard.vue`
- Render: plant name, zone, cycle stage, program EC/pH/volume summary
- High-tier warning banner (reuse Phase 30 WS2 patterns)

**Acceptance:** Vitest renders bundle diff; Confirm disabled for viewer role.

---

### WS6 — Operator documentation

**Tasks:**

- New section in [`docs/farm-guardian-architecture.md`](../farm-guardian-architecture.md) — setup PRs vs bootstrap templates vs manual UI
- [`docs/operator-tour.md`](../operator-tour.md) — walkthrough: house plant setup via Guardian
- Update persona/platform block: Guardian **can** propose multi-step **grow setup** PRs after Confirm (Phase 32)

**Acceptance:** Doc lists exact tools; "Guardian cannot silently add plants" remains true until Confirm.

---

### WS7 — OpenAPI + tests

**Tasks:**

- OpenAPI: document `apply_grow_setup_pack` args schema on proposal objects (extend 0.4.x)
- Go smoke: propose setup pack → Confirm → assert plant + cycle + program rows
- Go smoke: zone with active cycle → propose fails validation
- Vitest: setup pack card snapshot

**Acceptance:** `make test` green; bundle smokes idempotent with cleanup.

---

### WS8 — Guardian knowledge depth (platform doc RAG)

**Goal:** Operators can ask Guardian **how gr33n works** — Pi setup, Guardian PR inbox, fertigation workflows, troubleshooting — and get answers **grounded in your indexed doc corpus**, not only generic Llama weights or the live snapshot.

**Problem today:**

- Persona + platform block (`persona.go`, `platform_context.go`) cover **tone and high-level product facts** — not the full [`docs/`](../) tree.
- RAG is **opt-in per farm** via [`cmd/rag-ingest`](../../cmd/rag-ingest) / [`scripts/rag-ingest-demo.sh`](../../scripts/rag-ingest-demo.sh) — demo ingest indexes **operational farm text**, not platform operator guides.
- Without WS8, questions like *"how do I run the Pi field checklist?"* or *"what does Confirm do on a high-tier actuator PR?"* rely on model guesswork unless the operator reads markdown manually.

**Tasks:**

1. **Curated doc manifest** — `docs/rag/platform-doc-manifest.yaml` (or similar) listing paths to index:
   - Operator-facing: [`operator-tour.md`](../operator-tour.md), [`local-operator-bootstrap.md`](../local-operator-bootstrap.md), [`workflow-guide.md`](../workflow-guide.md), [`pi-integration-guide.md`](../pi-integration-guide.md), [`operator-troubleshooting.md`](../operator-troubleshooting.md), [`farm-guardian-architecture.md`](../farm-guardian-architecture.md), [`farm-guardian-persona-platform-context.md`](../farm-guardian-persona-platform-context.md), [`farm-guardian-ollama-setup.md`](../farm-guardian-ollama-setup.md), phase **operator** plans (14, 26, 30 summaries — not raw dev todos).
   - Exclude: secrets, `.env` examples with placeholders, generated OpenAPI blobs, internal agent plans marked dev-only (configurable `exclude_globs`).
2. **Ingest script** — `scripts/rag-ingest-platform-docs.sh`:
   - Calls `cmd/rag-ingest` with `source_type=platform_doc` (or extend existing enum) per farm_id (default demo farm **1** for dev; production: ingest once per farm or shared template farm — document choice).
   - `--dry-run` prints file list + chunk count estimate.
   - Idempotent re-run (same source_id replaces chunks).
3. **Makefile** — `rag-ingest-platform-docs` + `make dev-stack-fresh-rag` optional hook (after demo ingest).
4. **Persona / prompt** — extend grounded chat instructions: when RAG chunks present with `platform_doc`, prefer citing them for **how-to / troubleshooting**; still never invent live sensor values (snapshot wins for "right now").
5. **Operator doc** — WS6 section: *"What Guardian knows"* = snapshot + RAG corpus + weights; link ingest commands.
6. **Smoke** — with `EMBEDDING_API_KEY` + seeded chunks, grounded `/v1/chat` question *"How do I confirm a Guardian actuator PR?"* returns answer with citation to architecture or operator-tour chunk (assert `context_count > 0` + citation source in smoke).

**Not in scope (WS8):**

- Indexing the entire git repo including `internal/` Go source (use architecture docs instead).
- Auto-ingest on every `git pull` (operator runs ingest deliberately).
- Replacing Phase 32 WS1 **live** reads (plants/programs) — RAG is **documentation**; snapshot/tools are **database truth**.

**Acceptance:**

- `./scripts/rag-ingest-platform-docs.sh --dry-run` lists ≥ N markdown files from manifest.
- After ingest on farm 1, grounded chat cites platform doc for a Pi/Guardian how-to smoke question.
- [`farm-guardian-architecture.md`](../farm-guardian-architecture.md) §3.2 updated with platform doc layer vs farm notes vs snapshot.

**Relationship to WS1:** WS1 = **live DB reads** for setup; WS8 = **static operator knowledge**. Both reduce blind spots; implement **WS8 early** if persona/how-to testing is the priority (parallel with WS2).

---

## Relationship to other phases

| Phase | Relationship |
|-------|----------------|
| **29** | Propose→confirm foundation |
| **30** | PR queue, risk tiers, patch tools — **prerequisite** |
| **31** | Edge/Pi validation; WS6 `list_unread_alerts` + `summarize_zone` — Phase 33 WS1 hardens; WS1 here adds grow-domain reads |
| **33** | Read-tool polish, hardware CI, site manifest — **WS1 optional preamble** |
| **15** | Whole-farm templates — complementary, not replaced |

---

## Out of scope (defer to Phase 33 or later)

- Read-tool intent guards, context_ref dedup, read audit — **Phase 33 WS1–WS3**
- Hardware CI gate (`GR33N_HARDWARE_TEST=1`) — **Phase 33 WS4**
- Enterprise `site-manifest.yaml` provisioner — **Phase 33 WS5**
- Autonomous recurring fertigation ("feed every Tuesday" without Confirm)
- LLM direct SQL or arbitrary API proxy
- Bulk import / multi-farm plant broadcast
- Certified agronomic prescriptions or guaranteed yields
- Auto-create sensors/actuators/hardware registry from chat
- Full UI parity snapshot (every dashboard widget)
- Replacing human repotting, plumbing, harvest
- **Whole-repo code indexing** — Phase 32 WS8 covers **curated operator docs**, not every Go/SQL file

---

## Suggested implementation order

0. **Phase 33 WS1** (optional, ~1 session) — read-tool hardening if Phase 31 WS6 is on `main`
1. **WS8** (optional first) — platform doc RAG if persona/how-to testing is blocked on knowledge gaps
2. **WS2** — single create tools (testable in isolation via manual proposal insert)
3. **WS1** — read snapshot (unblocks safe WS4 matching)
4. **WS3** — setup pack transaction
5. **WS4** — intent generation + house_plant template
6. **WS5** — Confirm UX
7. **WS7** — smokes
8. **WS6** — doc pass (includes WS8 ingest runbook)

---

## Definition of done (phase ship)

- [x] Operator can ask for **house plant + zone + fertigation** setup in chat and receive **one** pending setup pack PR
- [x] Confirm atomically creates plant, cycle, program (or clear error, no partial state)
- [x] Snapshot lists existing plants/programs so duplicate proposals are reduced
- [x] High-tier warning on setup pack Confirm
- [x] Docs explain setup PR vs Phase 15 bootstrap vs manual Plants/Fertigation UI
- [x] Platform doc RAG pack ingestible; grounded chat cites operator docs for how-to (WS8)
- [x] `make test` green; OpenAPI documents new tool + bundle args

**Phase shipped.** All criteria met. See Phase 33 for post-ship polish and Phase 34 for iterative PR refinement.

---

## Using this plan in a new chat

```text
Implement Phase 32 per @docs/plans/archive/phase_32_guardian_grow_setup_prs.plan.md.

Start with WS2 create_plant + create_crop_cycle tools and smokes, then WS3
apply_grow_setup_pack transaction. Read Phase 30 tools registry and plants/cropcycle/
fertigation handlers. WS4 house_plant template for demo phrase "philodendron".
Do not bypass Confirm; no autonomous writes.
```

---

## Related

| Doc | Role |
|-----|------|
| [`phase_30_guardian_change_requests.plan.md`](phase_30_guardian_change_requests.plan.md) | PR queue prerequisite |
| [`phase_31_field_validation_and_edge.plan.md`](phase_31_field_validation_and_edge.plan.md) | Edge validation; read tools origin |
| [`phase_33_guardian_polish_and_enterprise_ops.plan.md`](phase_33_guardian_polish_and_enterprise_ops.plan.md) | WS1 hardening preamble; enterprise manifest |
| [`phase_15_farm_onboarding.plan.md`](phase_15_farm_onboarding.plan.md) | Whole-farm templates |
| [`farm-guardian-architecture.md`](../farm-guardian-architecture.md) | Request flow + blind spots |
| [`domain-modules-operator-playbook.md`](../domain-modules-operator-playbook.md) | Plants module |
