---
name: Phase 34 — Guardian PR iteration & blind-spot inputs
overview: >
  Turn a Guardian change request from a one-shot propose→Confirm into a short
  conversation: revise a pending proposal until it is correct, supersede the prior
  draft, and let the operator supply facts Guardian cannot sense. Adds richer
  "what happens if I accept this" impact explanations on every card. Never relaxes
  the Confirm gate — iteration happens before Confirm, and every revision is still
  frozen + audited.
todos:
  - id: ws1-supersede-model
    content: "WS1: Supersede model — proposal revision chain (supersedes_proposal_id / status=superseded); one live proposal per thread; TTL refresh on revise"
    status: completed
  - id: ws2-revise-intent
    content: "WS2: Revise intent — detect correction turns against active proposal; rebuild frozen args from prior + delta"
    status: completed
  - id: ws3-operator-facts
    content: "WS3: Operator-supplied facts — accept operator assertions Guardian cannot sense, labeled operator_provided (not measured)"
    status: completed
  - id: ws4-impact-explanation
    content: "WS4: Impact explanation — plain-language 'if you Confirm, this will…' block for every tool/risk tier"
    status: completed
  - id: ws5-revise-ux
    content: "WS5: Revise UX — card shows revision N, Refine affordance, diff vs prior draft, superseded badge"
    status: completed
  - id: ws6-confirm-guard
    content: "WS6: Confirm safety — only the latest live proposal is confirmable; superseded/expired → 410; audit revision lineage"
    status: completed
  - id: ws7-openapi-tests
    content: "WS7: OpenAPI + tests — revision fields + revise flow; Go smokes; Vitest revise card"
    status: completed
  - id: ws8-docs
    content: "WS8: Docs — architecture revise loop + blind-spot facts; operator-tour refine walkthrough; persona note"
    status: completed
isProject: false
---

# Phase 34 — Guardian PR iteration & blind-spot inputs

## Status

**Shipped (WS1–WS8).** Supersede chain + `meta`/`superseded` migration and sqlc (WS1); `tryReviseActiveProposal` delta-merge + revision router (WS2); `operator_provided` blind-spot facts (WS3); backend `impact.go` + `guardianImpact.js` impact explanation (WS4); revise UX — revision badge, Refine, diff vs prior, superseded state (WS5); confirm-safety 410 + `live_proposal_id` + audit lineage (WS6); OpenAPI fields + Go smoke (`smoke_phase34_revise_confirm_test.go`) + Vitest (WS7); architecture §7.7, operator-tour §6c, persona mirror (WS8). Full `go test ./...` and Vitest green.

Depends on **Phase 29/30** (PR queue + `guardian_action_proposals` + Confirm path) and **Phase 32** (setup-pack card / bundle diff UI reused as the impact-explanation base).

**Preconditions:**

- [`guardian_action_proposals`](../../db/migrations/20260521_phase29_guardian_proposals.sql) + Confirm path ([`internal/handler/chat/confirm.go`](../../internal/handler/chat/confirm.go))
- Rule-assisted proposal builder ([`internal/farmguardian/proposals.go`](../../internal/farmguardian/proposals.go), [`proposals_setup_pack.go`](../../internal/farmguardian/proposals_setup_pack.go))
- Proposal cards ([`GuardianActionProposal.vue`](../../ui/src/components/GuardianActionProposal.vue), [`SetupPackProposalCard.vue`](../../ui/src/components/SetupPackProposalCard.vue))
- Proposal inbox + list API ([`internal/handler/chat/proposals.go`](../../internal/handler/chat/proposals.go))

---

## Why this phase

Today a Guardian proposal is **single-shot**: a grounded turn templates one frozen proposal, and the operator can only **Confirm** or **Dismiss** it (5-minute TTL, args frozen at propose time). Real operator conversations are not single-shot:

- *"Close, but use 0.3 L not 0.5."*
- *"Wrong zone — that cycle already ended."*
- *"There's no humidity sensor in that tent; assume it sits around 60%."*

The third case is the **blind spot**: even with more read tools and platform-doc RAG, Guardian cannot sense hardware that isn't installed. The operator holds ground truth Guardian can't see. Without an iterate loop, the operator must Dismiss and re-ask from scratch, losing context.

| Today | After Phase 34 |
|-------|----------------|
| Propose → Confirm / Dismiss (one shot) | Propose → **Refine ×N** → Confirm the corrected draft |
| Wrong proposal → Dismiss + re-ask | Wrong proposal → "actually, …" → revised proposal supersedes it |
| Guardian only knows sensed + DB facts | Operator can **assert** unsensed facts, labeled `operator_provided` |
| Card shows args / bundle diff | Card shows plain-language **"if you Confirm, this will…"** impact |

**Hard invariant preserved:** nothing writes to the database until **Confirm**, and every confirmable revision is still a frozen, server-replayed, audited proposal. Iteration changes *which* frozen draft is live — never the Confirm gate.

---

## Design principles

1. **Confirm gate is sacred** — revision happens entirely in the pending (pre-Confirm) state. A revised proposal is a *new frozen row*, not an editable live one.
2. **One live draft per thread** — a proposal carries `supersedes_proposal_id`; revising marks the prior `superseded` and creates the successor. Only the latest live proposal is confirmable.
3. **Operator facts are labeled, never invented** — unsensed values the operator supplies are stored as `operator_provided` in proposal meta and surfaced as such ("assumed RH 60%, operator-stated") so the audit trail never confuses an assertion with a measurement.
4. **Reuse, don't fork** — build on `guardian_action_proposals`, the rule-assisted builders, and the Phase 32 card; do not introduce a second proposal store.
5. **Explainability** — every card answers "what happens if I accept this?" in plain language, scaled by risk tier.
6. **Bounded chains** — cap revision depth + refresh TTL on each revise so a thread can't grow unbounded or expire mid-refine.

---

## Architecture

```
Operator: "add basil to Tent A with a light feed"
   └─► grounded /v1/chat ─► propose P1 (pending, rev 1)
        card: "If you Confirm, this will: create plant Basil, start a cycle in Tent A,
               create program 'Basil light feed' (EC 0.8 / pH 5.8–6.5, 0.5 L)…"

Operator: "no humidity sensor in Tent A — assume ~60%, and use 0.3 L"
   └─► grounded /v1/chat ─► revise(P1)
        ├─ P1.status = superseded
        └─ propose P2 (pending, rev 2, supersedes_proposal_id = P1)
            args = P1.args ⊕ {program.total_volume_liters: 0.3}
            meta.operator_provided = [{field: "rh_pct", value: 60, basis: "operator_stated"}]
            card: revision 2 · diff vs rev 1 · refreshed TTL

Operator taps Confirm (P2 only — P1 now 410)
   └─► /v1/chat/confirm {proposal_id: P2} ─► execute ─► audit (records rev lineage)
```

Revision is detected the same rule-assisted way Phase 29/32 detect intent: a correction turn while an **active proposal exists in the thread** routes to a *revise* path instead of a fresh proposal.

---

## Scope

| WS | Focus | Primary artifacts |
|----|-------|-------------------|
| **WS1** | Supersede model | migration: `supersedes_proposal_id`, `status='superseded'`, `revision`; queries in [`db/queries/guardian_proposals.sql`](../../db/queries/guardian_proposals.sql) |
| **WS2** | Revise intent | `internal/farmguardian/proposals_revise.go` (new); delta-merge frozen args |
| **WS3** | Operator-supplied facts | `operator_provided` meta + parser; snapshot/prompt labeling |
| **WS4** | Impact explanation | generalize [`ui/src/lib/guardianSetupPack.js`](../../ui/src/lib/guardianSetupPack.js) → `guardianImpact.js`; backend summary helper |
| **WS5** | Revise UX | [`GuardianActionProposal.vue`](../../ui/src/components/GuardianActionProposal.vue) revision badge, Refine affordance, diff-vs-prior |
| **WS6** | Confirm safety | [`confirm.go`](../../internal/handler/chat/confirm.go) latest-live guard + lineage audit |
| **WS7** | OpenAPI + tests | `openapi.yaml`; `cmd/api/smoke_phase34_*_test.go`; Vitest |
| **WS8** | Docs | [`farm-guardian-architecture.md`](../farm-guardian-architecture.md), [`operator-tour.md`](../operator-tour.md), persona mirror |

---

## Work-stream detail

### WS1 — Supersede model

**Goal:** A proposal can point at the one it replaces; only the newest is live.

**Tasks:**

1. Migration adds to `gr33ncore.guardian_action_proposals`: `supersedes_proposal_id UUID NULL` (FK self), `revision INT NOT NULL DEFAULT 1`, and `'superseded'` as a `status` value (alongside `pending`/`confirmed`/`dismissed`/`expired`).
2. Index for "latest live in a thread" lookups (root proposal id or session + tool).
3. sqlc queries: `SupersedeProposal`, `GetLatestLiveInChain`, `ListProposalChain`.
4. On revise, the prior proposal flips to `superseded`; the successor inherits `farm_id`, `user_id`, `session_id`, tool, and a refreshed `expires_at`.

**Acceptance:** Inserting a revision marks the parent `superseded`; chain query returns ordered revisions; only one `pending` row per chain.

### WS2 — Revise intent

**Goal:** A correction turn rebuilds the proposal instead of starting over.

**Tasks:**

1. New `matchReviseIntent` — when the session has an active pending proposal and the turn reads as a correction/refinement ("use 0.3 L", "wrong zone, it's Veg Room", "make the cycle flower not veg"), route to revise.
2. Delta-merge: start from the prior frozen args, apply only the changed fields, re-run the same tool validation used at propose time (zone scope, one-active-cycle, allowlisted fields).
3. Ambiguous correction → Guardian asks a clarifying question instead of guessing (no silent wrong revise).
4. Works for both single-tool proposals and the Phase 32 `apply_grow_setup_pack` bundle (per-section deltas: plant / cycle / program / task).

**Acceptance:** "use 0.3 L" on a setup pack yields rev 2 with only `program.total_volume_liters` changed; invalid delta (zone now busy) returns a clear chat error, no broken proposal.

### WS3 — Operator-supplied facts

**Goal:** Capture ground truth Guardian cannot sense, clearly labeled.

**Tasks:**

1. Parse operator assertions of unsensed values ("assume RH ~60%", "water source is well water", "no EC meter on this line").
2. Store under proposal `meta.operator_provided[]` = `{field, value, basis: "operator_stated", turn_ref}`; never merged into a field that implies a live measurement.
3. Surface in the prompt + card as explicitly operator-stated so neither the model nor the audit confuses it with a sensor reading.
4. Where an operator fact unblocks a proposal (e.g. choosing `water_source=ro|well|plain` for a watering/fertigation decision), feed it into the tool args as a first-class allowlisted field.

**Acceptance:** Operator-stated RH appears in the card as "RH 60% (operator-stated, not measured)"; audit `details` records `operator_provided`; no sensor table is written.

### WS4 — Impact explanation ("if you Confirm, this will…")

**Goal:** Every card explains consequences in plain language, not just raw args.

**Tasks:**

1. Backend summary helper per tool → ordered human steps + reversibility hint (e.g. "creates 1 program (editable later)", "queues a Pi command — relay fires on next poll").
2. Generalize the Phase 32 `formatSetupPackBundle` into `guardianImpact.js` covering all tools/risk tiers.
3. High-tier cards lead with the irreversible/most-impactful line (actuator enqueue, disable rule, setup pack).

**Acceptance:** A `patch_fertigation_program` card reads "If you Confirm: EC target 0.8 → 1.0 on 'Basil light feed' (no run triggered now)"; setup pack still shows the numbered bundle.

### WS5 — Revise UX

**Goal:** Operator can see and drive the refine loop.

**Tasks:**

1. Card shows **revision N**, a **Refine** affordance (prompt prefill to push a correction), and a **diff vs previous revision**.
2. Superseded proposals render a muted "superseded by rev N" badge in transcript + inbox; Confirm disabled on them.
3. Viewer role: can see revisions and impact text, **cannot** Confirm or Refine into a write (server-enforced).

**Acceptance:** Vitest renders rev-2 card with diff + superseded badge on rev 1; viewer sees disabled Confirm.

### WS6 — Confirm safety

**Goal:** Only the latest live draft is confirmable; lineage is auditable.

**Tasks:**

1. `confirm.go` rejects Confirm on a `superseded` proposal with **410 Gone** (same as expired), pointing to the live revision id.
2. Confirming the latest live proposal records `revision` + root id in the `guardian_tool_executed` audit details.
3. TTL refresh on revise so a long refine conversation doesn't expire the live draft mid-thread.

**Acceptance:** Confirm on superseded → 410; Confirm on latest → 200 + audit shows `revision: 2`, `root_proposal_id`.

### WS7 — OpenAPI + tests

**Tasks:**

- OpenAPI: add `supersedes_proposal_id`, `revision`, `status: superseded`, `operator_provided[]`, and an impact summary field to `GuardianActionProposal`; document the revise flow note on `/v1/chat`.
- Go smokes: propose → revise → confirm latest (asserts rows reflect the delta); confirm superseded → 410; operator-fact persisted in audit.
- Vitest: revise card snapshot + impact block.

**Acceptance:** `make test` green; smokes idempotent with cleanup; superseded-confirm 410 asserted.

### WS8 — Docs

**Tasks:**

- `farm-guardian-architecture.md` §7.7 — revise/supersede loop + operator-supplied facts + impact explanations; update phase ledger.
- `operator-tour.md` — "Refine a Guardian request" walkthrough (correct a volume, supply an unsensed fact, Confirm the corrected draft).
- Persona / platform mirror: Guardian **may revise a pending request before Confirm** and **may use operator-stated facts** (labeled), still never writes silently.

**Acceptance:** Docs list the exact new fields/flow; "Guardian cannot silently write" remains true; blind-spot handling documented.

---

## Recommended order

WS1 (model) → WS2 (revise) → WS3 (operator facts) → WS4 (impact) → WS5 (UX) → WS6 (confirm safety) → WS7 tests → WS8 docs. WS4 can land in parallel with WS2/WS3 since it is presentation-only.

---

## Definition of done (phase ship)

- [x] Migration adds revision/supersede columns + `superseded` status; sqlc queries generated
- [x] Correction turns produce a revised, frozen, validated proposal that supersedes the prior
- [x] Operator-supplied facts stored + displayed as `operator_provided`, never as measurements
- [x] Every card shows a plain-language "if you Confirm" impact block
- [x] Confirm rejects superseded (410); audits revision lineage
- [x] OpenAPI updated; Go + Vitest smokes green
- [x] Architecture + operator-tour + persona docs updated

---

## Preconditions checklist (verified on `main`)

| Prerequisite | Status | Anchor |
|--------------|--------|--------|
| `guardian_action_proposals` + TTL + risk tier | ✅ | [`20260521_phase29_guardian_proposals.sql`](../../db/migrations/20260521_phase29_guardian_proposals.sql), [`20260526_phase30_guardian_proposal_risk_tier.sql`](../../db/migrations/20260526_phase30_guardian_proposal_risk_tier.sql) |
| Confirm replay path | ✅ | [`internal/handler/chat/confirm.go`](../../internal/handler/chat/confirm.go) — non-`pending` already → **410 Gone** |
| Rule-assisted propose | ✅ | [`internal/farmguardian/proposals.go`](../../internal/farmguardian/proposals.go) → `attachProposals` in chat handler |
| Setup pack builder + card | ✅ | [`proposals_setup_pack.go`](../../internal/farmguardian/proposals_setup_pack.go), [`SetupPackProposalCard.vue`](../../ui/src/components/SetupPackProposalCard.vue), [`guardianSetupPack.js`](../../ui/src/lib/guardianSetupPack.js) |
| Proposal card + smokes | ✅ | [`GuardianActionProposal.vue`](../../ui/src/components/GuardianActionProposal.vue), `cmd/api/smoke_phase32_*`, `insertGuardianProposalWithRisk` in [`smoke_phase32_ws2_test.go`](../../cmd/api/smoke_phase32_ws2_test.go) |
| Phases 32–33 shipped | ✅ | No `supersedes` / `revision` / `operator_provided` in codebase yet — green field |

**Not present today (Phase 34 delivers):** `meta` JSONB on proposals, `superseded` enum value, session-scoped “latest pending” query, `proposals_revise.go`, `guardianImpact.js`, revise UX on the card.

---

## Implementation decisions (resolve in WS1, do not bikeshed mid-phase)

1. **`meta` column** — add `meta JSONB NOT NULL DEFAULT '{}'::jsonb` on `guardian_action_proposals` in the Phase 34 migration. Store `operator_provided[]` there only; never mirror into `args` as if measured.
2. **Chain identity** — `supersedes_proposal_id` points at the immediate parent; expose `root_proposal_id` in API/audit as the first row in the chain (walk parents or store on insert).
3. **Session gate for revise** — new sqlc: `GetLatestPendingProposalBySession(user_id, session_id)` (or farm+session). WS2 calls this before `matchReviseIntent`; if none, fall through to normal propose.
4. **Revision cap** — `const MaxProposalRevisions = 8` in `farmguardian`; at cap return a chat error (“start a new request”) instead of another supersede.
5. **TTL on revise** — reset `expires_at` to `NOW() + ProposalTTL` on each successor (same 5m constant unless product asks otherwise).
6. **410 body** — on confirm of superseded proposal, include `live_proposal_id` in JSON error detail so the UI can link to rev N.
7. **Enum migration** — `ALTER TYPE gr33ncore.guardian_proposal_status_enum ADD VALUE IF NOT EXISTS 'superseded';` in migration; regenerate sqlc + bump [`db/schema/gr33n-schema-v2-FINAL.sql`](../../db/schema/gr33n-schema-v2-FINAL.sql) if that file is maintained in-repo for this enum.
8. **Impact field** — prefer `impact_summary` on `ActionProposal` / OpenAPI (server-built in WS4) over pushing all copy client-only.

---

## Golden-path scenario (WS7 smoke + manual demo)

Use the Phase 32 house-plant setup pack as the revise fixture (0.5 L default in [`proposals_setup_pack.go`](../../internal/farmguardian/proposals_setup_pack.go)):

1. Grounded chat: *"add philodendron to Tent A with a light feed"* → proposal P1 (`apply_grow_setup_pack`, rev 1).
2. Same `session_id`: *"use 0.3 L not 0.5"* → P1 `superseded`, P2 pending with `program.total_volume_liters: 0.3`, `revision: 2`.
3. Same session: *"no humidity sensor — assume RH around 60%"* → P2 superseded, P3 with `meta.operator_provided` entry; card shows operator-stated RH.
4. `POST /v1/chat/confirm` with P1 → **410** + `live_proposal_id` = P3.
5. Confirm P3 → **200**; audit `guardian_tool_executed` includes `revision`, `root_proposal_id`, `operator_provided`.

Reuse `housePlantSetupPackArgs` / `insertGuardianProposalWithRisk` patterns from Phase 32 smokes; add `smoke_phase34_revise_confirm_test.go` (and Vitest mirror in [`guardian-proposal.test.js`](../../ui/src/__tests__/guardian-proposal.test.js)).

---

## Agent kickoff (Opus / high-reasoning model)

**Model:** Claude Opus (or equivalent) with full repo access. Work on a feature branch off `main`; run `make test` after each WS.

```text
Implement Phase 34 from @docs/plans/archive/phase_34_guardian_pr_iteration.plan.md.

Order: WS1 → WS2 → WS3 → WS4 (may parallel with WS2/3) → WS5 → WS6 → WS7 → WS8.
Follow "Implementation decisions" in the plan — especially meta JSONB, session pending lookup, MaxProposalRevisions, 410 live_proposal_id.

WS1: Migration db/migrations/20260602_phase34_guardian_proposal_revision.sql (date may shift):
  supersedes_proposal_id, revision, meta, status superseded; sqlc SupersedeProposal,
  GetLatestPendingProposalBySession, GetLatestLiveInChain, ListProposalChain; make sqlc.

WS2: internal/farmguardian/proposals_revise.go — matchReviseIntent before fresh propose in
  BuildRuleAssistedProposals; delta-merge frozen args; re-validate with existing tool validators;
  setup-pack section deltas (plant/cycle/program/task). Ambiguous → clarifying chat, no guess.

WS3: operator_provided in meta only; label in summary/prompt/card.

WS4: backend impact helper per tool; ui/src/lib/guardianImpact.js generalized from guardianSetupPack.js.

WS5–6: GuardianActionProposal.vue revision badge, Refine prefill, diff vs prior, superseded state;
  confirm.go latest-live + audit lineage.

WS7: openapi GuardianActionProposal fields; smoke_phase34_*; Vitest revise card.

WS8: farm-guardian-architecture §7.7, operator-tour refine walkthrough, persona mirror.

Invariants: never write DB except via Confirm; revisions are new INSERT rows; only latest pending confirmable.
Do not start Phase 35. Commit per WS or one squashed commit at end — ask operator preference if unclear.
```

**Touch list (expected diffs):**

| Area | Files |
|------|--------|
| DB | `db/migrations/20260602_phase34_*.sql`, `db/queries/guardian_proposals.sql`, `internal/db/*.go` |
| Revise | `internal/farmguardian/proposals.go`, `proposals_revise.go`, `proposals_revise_test.go` |
| Chat | `internal/handler/chat/confirm.go`, `proposals.go` |
| UI | `GuardianActionProposal.vue`, `guardianImpact.js`, `guardianProposals.js`, tests |
| Contract | `openapi.yaml` |
| Docs | `docs/farm-guardian-architecture.md`, `docs/operator-tour.md`, persona doc if present |

---

## Using this plan in a new chat

> Implement Phase 34 from `docs/plans/archive/phase_34_guardian_pr_iteration.plan.md`. Start with WS1 (supersede migration + sqlc), then WS2 revise intent. Reuse `guardian_action_proposals`, the rule-assisted proposal builders, and the Phase 32 setup-pack card. Never relax the Confirm gate — revisions are new frozen rows; only the latest live one is confirmable. Operator-supplied facts must be labeled `operator_provided`, never written as sensor readings. Use the **Agent kickoff** and **Implementation decisions** sections above.

---

## Related

| Doc | Use |
|-----|-----|
| [farm-guardian-architecture.md](../farm-guardian-architecture.md) | PR flow §7, operator expectations §8 |
| [phase_32_guardian_grow_setup_prs.plan.md](phase_32_guardian_grow_setup_prs.plan.md) | Setup-pack card reused for impact explanation |
| [phase_30_guardian_change_requests.plan.md](phase_30_guardian_change_requests.plan.md) | PR queue, risk tiers, frozen args |
| [phase_35_lighting_domain.plan.md](phase_35_lighting_domain.plan.md) | Lighting (next grow-environment phase) |
| [audit-events-operator-playbook.md](../audit-events-operator-playbook.md) | `guardian_tool_executed` audit (now with revision lineage) |
