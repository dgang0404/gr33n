---
name: Phase 32 — Guardian grow setup PRs
overview: >
  Close the gap between "Guardian explains setup" and "Guardian opens a change request
  you can Confirm." Enable conversational, multi-step grow onboarding PRs — add plant,
  start cycle in zone, create/link fertigation program — without autonomous writes.
  Reuses Phase 30 PR queue; extends tools, snapshot reads, and proposal generation.
todos:
  - id: ws1-read-layer
    content: "WS1: Read layer — expand live snapshot (plants, programs per zone); read-only tools list_plants, summarize_zone_fertigation; optional route context from UI"
    status: pending
  - id: ws2-create-tools
    content: "WS2: Create tools — create_plant, create_crop_cycle, create_fertigation_program (allowlisted fields); map to existing handlers; medium/high risk tiers"
    status: pending
  - id: ws3-setup-pack-pr
    content: "WS3: Setup pack PR — single frozen bundle (plant + cycle + program) or linked proposal group; one Confirm card with step diff; transactional execute"
    status: pending
  - id: ws4-intent-generation
    content: "WS4: Intent → proposal — structured extraction / setup templates from chat (beyond Phase 29 regex); house-plant vs commercial zone profiles"
    status: pending
  - id: ws5-confirm-ux
    content: "WS5: Confirm UX — SetupPackProposalCard; show created rows preview; dismiss partial bundle rules"
    status: pending
  - id: ws6-operator-docs
    content: "WS6: Operator docs — what Guardian can propose vs manual UI; house plant example; link farm-guardian-architecture § blind spots"
    status: pending
  - id: ws7-openapi-tests
    content: "WS7: OpenAPI + smokes — setup pack propose→confirm; Vitest bundle card; no silent partial apply on validation failure"
    status: pending
isProject: false
---

# Phase 32 — Guardian grow setup PRs

## Status

**Not started.** Depends on **Phase 30** (PR queue + tools registry) shipped. **Phase 31** (field validation + WS6 read tools) can run in parallel but WS1 here subsumes/expands WS6 read coverage for grow domains.

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

## Relationship to other phases

| Phase | Relationship |
|-------|----------------|
| **29** | Propose→confirm foundation |
| **30** | PR queue, risk tiers, patch tools — **prerequisite** |
| **31** | Edge/Pi validation; WS6 read tools overlap WS1 — merge or implement WS1 here first |
| **15** | Whole-farm templates — complementary, not replaced |

---

## Out of scope (Phase 33+)

- Autonomous recurring fertigation ("feed every Tuesday" without Confirm)
- LLM direct SQL or arbitrary API proxy
- Bulk import / multi-farm plant broadcast
- Certified agronomic prescriptions or guaranteed yields
- Auto-create sensors/actuators/hardware registry from chat
- Full UI parity snapshot (every dashboard widget)
- Replacing human repotting, plumbing, harvest

---

## Suggested implementation order

1. **WS2** — single create tools (testable in isolation via manual proposal insert)
2. **WS1** — read snapshot (unblocks safe WS4 matching)
3. **WS3** — setup pack transaction
4. **WS4** — intent generation + house_plant template
5. **WS5** — Confirm UX
6. **WS7** — smokes
7. **WS6** — doc pass

---

## Definition of done (phase ship)

- [ ] Operator can ask for **house plant + zone + fertigation** setup in chat and receive **one** pending setup pack PR
- [ ] Confirm atomically creates plant, cycle, program (or clear error, no partial state)
- [ ] Snapshot lists existing plants/programs so duplicate proposals are reduced
- [ ] High-tier warning on setup pack Confirm
- [ ] Docs explain setup PR vs Phase 15 bootstrap vs manual Plants/Fertigation UI
- [ ] `make test` green; OpenAPI documents new tool + bundle args

---

## Using this plan in a new chat

```text
Implement Phase 32 per @docs/plans/phase_32_guardian_grow_setup_prs.plan.md.

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
| [`phase_31_field_validation_and_edge.plan.md`](phase_31_field_validation_and_edge.plan.md) | Edge validation; WS6 read overlap |
| [`phase_15_farm_onboarding.plan.md`](phase_15_farm_onboarding.plan.md) | Whole-farm templates |
| [`farm-guardian-architecture.md`](../farm-guardian-architecture.md) | Request flow + blind spots |
| [`domain-modules-operator-playbook.md`](../domain-modules-operator-playbook.md) | Plants module |
