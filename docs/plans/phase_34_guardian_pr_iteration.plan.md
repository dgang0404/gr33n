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
    status: pending
  - id: ws2-revise-intent
    content: "WS2: Revise intent — detect correction turns against active proposal; rebuild frozen args from prior + delta"
    status: pending
  - id: ws3-operator-facts
    content: "WS3: Operator-supplied facts — accept operator assertions Guardian cannot sense, labeled operator_provided (not measured)"
    status: pending
  - id: ws4-impact-explanation
    content: "WS4: Impact explanation — plain-language 'if you Confirm, this will…' block for every tool/risk tier"
    status: pending
  - id: ws5-revise-ux
    content: "WS5: Revise UX — card shows revision N, Refine affordance, diff vs prior draft, superseded badge"
    status: pending
  - id: ws6-confirm-guard
    content: "WS6: Confirm safety — only the latest live proposal is confirmable; superseded/expired → 410; audit revision lineage"
    status: pending
  - id: ws7-openapi-tests
    content: "WS7: OpenAPI + tests — revision fields + revise flow; Go smokes; Vitest revise card"
    status: pending
  - id: ws8-docs
    content: "WS8: Docs — architecture revise loop + blind-spot facts; operator-tour refine walkthrough; persona note"
    status: pending
isProject: false
---

# Phase 34 — Guardian PR iteration & blind-spot inputs

## Status

**Not started.** Depends on **Phase 29/30** (PR queue + `guardian_action_proposals` + Confirm path) and **Phase 32** (setup-pack card / bundle diff UI reused as the impact-explanation base).

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

- [ ] Migration adds revision/supersede columns + `superseded` status; sqlc queries generated
- [ ] Correction turns produce a revised, frozen, validated proposal that supersedes the prior
- [ ] Operator-supplied facts stored + displayed as `operator_provided`, never as measurements
- [ ] Every card shows a plain-language "if you Confirm" impact block
- [ ] Confirm rejects superseded (410); audits revision lineage
- [ ] OpenAPI updated; Go + Vitest smokes green
- [ ] Architecture + operator-tour + persona docs updated

---

## Using this plan in a new chat

> Implement Phase 34 from `docs/plans/phase_34_guardian_pr_iteration.plan.md`. Start with WS1 (supersede migration + sqlc), then WS2 revise intent. Reuse `guardian_action_proposals`, the rule-assisted proposal builders, and the Phase 32 setup-pack card. Never relax the Confirm gate — revisions are new frozen rows; only the latest live one is confirmable. Operator-supplied facts must be labeled `operator_provided`, never written as sensor readings.

---

## Related

| Doc | Use |
|-----|-----|
| [farm-guardian-architecture.md](../farm-guardian-architecture.md) | PR flow §7, operator expectations §8 |
| [phase_32_guardian_grow_setup_prs.plan.md](phase_32_guardian_grow_setup_prs.plan.md) | Setup-pack card reused for impact explanation |
| [phase_30_guardian_change_requests.plan.md](phase_30_guardian_change_requests.plan.md) | PR queue, risk tiers, frozen args |
| [phase_35_lighting_domain.plan.md](phase_35_lighting_domain.plan.md) | Lighting (next grow-environment phase) |
| [audit-events-operator-playbook.md](../audit-events-operator-playbook.md) | `guardian_tool_executed` audit (now with revision lineage) |
