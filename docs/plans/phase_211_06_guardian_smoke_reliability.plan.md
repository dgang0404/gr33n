---
name: Phase 211.06 — Guardian smoke reliability (laptop loop)
overview: >
  Make laptop Guardian QA honest and recoverable before Phase 212: fail when
  fixtures fail, debug without a 12h marathon, refuse invent/near-miss answers
  from looking like product wins, and ground NF dilutions when the model ignores
  tool payloads.
todos:
  - id: ws1-harness-honesty
    content: "WS1: Fail make/CI paths on fixture fail + timeout; fix smoke-all archive rollup; partial archive on RunQuestion error path"
    status: completed
  - id: ws2-laptop-loop
    content: "WS2: Document + make targets for core / core+NF1 as default debug loop; full smoke-all = certification; optional cool-down / NF timeout knobs"
    status: completed
  - id: ws3-invent-refuse-retry
    content: "WS3: Detect apology / roleplay / instruction-soup openings and regenerate once before scoring (product path)"
    status: completed
  - id: ws4-nf-dilution-grounding
    content: "WS4: Medium first — answer template must quote tool-block ratios; hard catalog short-circuit only for pure dilution intents if phi3 still invents"
    status: completed
  - id: ws5-ecph-write-relevance
    content: "WS5: ec-ph pH near-miss (inject from cite or soften score); keep write proposal pass, add separate answer-relevance grade"
    status: completed
  - id: ws6-closure
    content: "WS6: Docs + unit tests shipped; operator re-runs make guardian-qa-debug vs Jul 29 7/25 baseline"
    status: completed
isProject: false
---

# Phase 211.06 — Guardian smoke reliability (laptop loop)

**Status:** Shipped (WS1–WS6 code + docs) · **Depends on:** [211.05 recipe outcome insights](phase_211_05_recipe_outcome_insights.plan.md) shipped + Jul 29 full smoke-all baseline graded · **After:** [211.07 Pending tab Playwright E2E](phase_211_07_guardian_pending_e2e.plan.md) · **Before:** [212 dual-install federation](phase_212_dual_farm_federation_test.plan.md) (still gated on a trustworthy local QA loop)

**Operator follow-up:** `make guardian-qa-debug` on phi3 laptop — target core ≥4/5 and NF batch1 not all invent/timeout (vs Jul 29 baseline 7/25).

## The one job

> Laptop Guardian QA must **tell the truth** and **fail usefully**: no more celebrating “OK” suites that are 0/N, no 12h marathon as the only debug loop, and invent / near-miss answers must not look like product wins.

## Why now

Jul 28–29 `make guardian-qa-smoke-all` on CPU-only `phi3:mini` finished ~12h and reported **0 suite failures** while fixture reality was **7/25 (28%, letter D)**:

| Batch | Score | Notes |
|-------|-------|--------|
| Core smoke | **3/5** | morning-walk + unread-alerts PASS after prior P0/P1; fail ec-ph, cherry-jlf |
| NF batch 1 | **0/5** | 2× `llm_timeout` @ ~2400s + invent |
| NF batch 2 | **0/7** | 1× timeout + invent / missing recipe-avg language |
| phase127 | **0/4** | instruction-bleed / apology stubs |
| change-requests | **4/4*** | proposals + pending queue real; answer prose garbage / low relevance |

\*Writes pass on **proposal presence**, not answer quality.

Two distinct failure modes showed up:

1. **Laptop reality** — CPU-only phi3 cannot sustain ~25 grounded turns overnight; timeouts and thermal/marathon collapse are expected without harness knobs.
2. **Model collapse / invent** — after ~prompt 4, answers become roleplay, apology, or instruction-soup; tools may fire and the model ignores them.

Partial archives and durable smoke-all logs already land from earlier fixes (`partial_<suite>_<model>.json`, `data/guardian_qa_runs/smoke-all-latest.log`). Remaining gaps: error-path partials, rollup markers, default exit honesty, and product refuse/retry.

## Non-negotiable framing

- **Harness honesty before heroics.** If make/CI says OK on 0/N, every later counsel tweak is unmeasurable.
- **Full smoke-all is certification, not the debug default.** Default loop = core (± one NF batch).
- **Reuse existing hygiene.** `TrimInstructionLeak`, `TrimMetaCorrection`, topic-drift / invent scoring, and `-fail-on-regression` already exist — extend them; do not rewrite the eval pipeline.
- **No new ML/judge dependency** in this phase. One regenerate on invent stubs is enough; Phase 146-style GPU self-critique stays optional/out of scope.
- Mark deliberate shortcuts (global cool-down, catalog short-circuit, softened near-miss scores) with a `ponytail:` comment naming the ceiling and upgrade path.

## Scope

- `cmd/guardian-eval`, `scripts/guardian-qa-smoke-all.sh`, Makefile Guardian QA targets
- Answer finalize / invent refuse+retry on the chat/counsel path
- NF dilution grounding for pure dilution / application-ratio intents
- Eval scoring: ec-ph near-miss; write-intent **answer relevance** as a separate signal from proposal pass
- Docs: [ci-guardian-qa.md](../ci-guardian-qa.md), [learning/guardian-qa-harness-gaps.md](../learning/guardian-qa-harness-gaps.md), farm-guardian-architecture cross-link

## Out of scope

- Phase 212 dual-install federation
- New NF product features, schema, or Commons packs
- Making full overnight smoke-all the only acceptance proof
- Mandatory PR CI for all Guardian smokes (strict / label-gated stays opt-in)
- Platform-docs RAG ingest reliability (ops note only; not a blocker for WS1–WS5)

## Workstreams

### WS1 — Harness honesty

**Goal:** Exit codes and rollups match fixture reality.

1. Wire laptop / smoke-all paths so fixture heuristic fails and `llm_timeout` / suite errors surface as **non-zero** (reuse `-fail-on-regression`; decide whether default `guardian-qa-smoke` stays artifact-only and smoke-all / a new `*-strict` path fails — prefer: **smoke-all fails the run when any batch fails fixtures**, keep plain smoke as report-only if already documented that way).
2. Fix end-of-run archive rollup in `scripts/guardian-qa-smoke-all.sh` (Jul 29 printed “(no archives captured this run)” despite five timestamped archives under `data/guardian_qa_runs/`).
3. On `RunQuestion` **error** path (timeout, 502, etc.), rewrite the partial archive before `continue` — today success paths save partials; errors can leave the last good prompt only.

**Done when:** A forced-fail fixture or a synthetic timeout makes make exit non-zero, and smoke-all rollup lists the archives from that run with pass/total.

### WS2 — Laptop loop

**Goal:** Debug without a 12h certification run.

1. Document default loop: `guardian-qa-smoke` (core) and/or core + `guardian-qa-smoke-nf-batch1` after counsel changes; `guardian-qa-smoke-all` = certification.
2. Optional knobs (smallest that help): cool-down between prompts; higher / NF-specific `GUARDIAN_EVAL_TIMEOUT_*`; shorter NF prompt text where fixtures allow without changing intent.
3. Make targets or help text that name the debug vs certification distinction (`make guardian-qa-smoke-all-help` and [ci-guardian-qa.md](../ci-guardian-qa.md)).

**Done when:** An operator can re-validate the hot path in one core (+ optional NF1) run without invoking all five batches.

### WS3 — Invent refuse + retry (product)

**Goal:** Stop scoring instruction-soup as a finished answer.

1. Detect apology stubs, roleplay openings (“You are a…”), and substitute-instruction / prompt-echo openings **before** final persist/score (extend existing finalize / leak trim; grep callers of the shared finalize path once).
2. On hit: **one** regenerate (same turn context / tools), then finalize again.
3. If still invent: prefer a grounded refuse / “I don’t have that in farm records” over publishing soup — do not silently pass invent.

**Done when:** Unit/self-check covers the detectors; a mid-suite invent stub would have triggered retry in the Jul 29 phase127 / NF failure shapes.

### WS4 — NF dilution grounding

**Goal:** Dilution answers quote catalog ratios when tools already returned them.

**Hardness ladder (pick in implementation order):**

| Level | Approach | When |
|-------|----------|------|
| Soft | Stronger footers / payloads | Already tried; phi3 still invents — not the primary bet |
| **Medium (default)** | Answer template / skeleton that **must** quote ratios from the tool block | First implementation target |
| Hard | Pure dilution intents: render catalog text **without** LLM (`ponytail:` ceiling — bypasses prose quality; upgrade = constrained decode or stronger model) | Only if medium still fails on phi3 laptop |

Reuse `CanonApplicationRecipesForProcess`, application_recipes dilutions, and existing NF process tool footers — do not invent a parallel catalog.

**Done when:** Dilution fixtures in NF batch1 pass or fail for **content** reasons other than invented ratios / empty invent, on a core+NF1 re-run.

### WS5 — ec-ph near-miss + write answer relevance

**Goal:** Grade what operators care about without lying.

1. **ec-ph:** Answer had EC + citations but no “pH” token → scorer fail. Either inject pH from cited excerpt in finalize, or soften score when cite excerpts already contain pH (prefer inject-from-cite if excerpt is authoritative).
2. **Writes:** Keep proposal + pending-queue pass. Add a **separate** answer-relevance / drift grade so junk prose is visible in archives and rollups even when Confirm path works.

**Done when:** ec-ph near-miss no longer fails solely for missing “pH” when cites carry pH; change-requests archives expose proposal pass ≠ answer quality.

### WS6 — Closure

1. Re-run **core + NF batch1 only** on `phi3:mini` after WS1–WS5 (not a full smoke-all unless certification is intentionally scheduled).
2. Grade against Jul 29 baseline (**7/25** overall; core **3/5**; NF1 **0/5**). Target for phase exit: **core ≥ 4/5** and NF batch1 **not** all invent/timeout.
3. Update [guardian-qa-harness-gaps.md](../learning/guardian-qa-harness-gaps.md) (move fixed rows), [ci-guardian-qa.md](../ci-guardian-qa.md), and a short farm-guardian-architecture note; mark this plan **Complete**.

## Acceptance criteria

- [x] Smoke-all (or documented strict path) exits non-zero when any batch has fixture failures or timeouts (`GUARDIAN_QA_FAIL_ON_REGRESSION=1` default in smoke-all)
- [x] Smoke-all end rollup lists archives from the run with pass/total (no false “no archives”) — fixed subshell + skip `partial_*`
- [x] Partial archive rewritten on `RunQuestion` error path
- [x] Docs + make help distinguish debug loop (`guardian-qa-debug`) vs certification (smoke-all)
- [x] Invent/apology/instruction-soup openings trigger at most one regenerate before final answer
- [x] NF dilution path quotes tool/catalog ratios (`EnsureNFDilutionRatiosInAnswer`)
- [x] ec-ph does not fail solely when cites already contain pH and answer has EC
- [x] Write-intent scoring separates proposal pass from answer relevance (`answer_relevant`)
- [ ] Closure re-run: core ≥ 4/5 and NF batch1 not all invent/timeout on phi3 laptop — **operator** (`make guardian-qa-debug`)

## Suggested implementation order

1. WS1 harness honesty  
2. WS3 invent refuse+retry  
3. WS5 ec-ph  
4. WS4 NF dilution grounding  
5. WS2 laptop-loop docs/targets (can land early alongside WS1)  
6. WS6 closure re-run  

## Baseline reference (do not delete)

- Archives: `data/guardian_qa_runs/20260728T202243_smoke_…`, `20260728T231409_smoke-natural-farming_…`, `20260729T020925_smoke-natural-farming_…`, `20260729T041512_phase127_…`, `20260729T062008_change-requests_…`
- Logs: `/tmp/guardian-qa-smoke-all.log`, `data/guardian_qa_runs/smoke-all-latest.log`
- Prior P0/P1 already shipped: morning-walk multi-category cite skip; unread-alert chunk seed; NF dilution routing/payload; phase127 fertigation intent widen; partial archives on success path

## Related

- [211.05 recipe outcome insights](phase_211_05_recipe_outcome_insights.plan.md) — NF history fixtures this phase must keep measurable
- [211 natural farming switchover commons](phase_211_natural_farming_switchover_commons.plan.md) — parent arc; smoke promotion lived in WS5
- [ci-guardian-qa.md](../ci-guardian-qa.md) — make targets and strict vs report-only
- [learning/guardian-qa-harness-gaps.md](../learning/guardian-qa-harness-gaps.md) — harness checklist to update in WS6
- [learning/guardian-pipeline-for-csharp-devs.md](../learning/guardian-pipeline-for-csharp-devs.md) — teaching notes from 211.05 + invent failures
- [212 dual-install federation](phase_212_dual_farm_federation_test.plan.md) — stays deferred until this loop is trustworthy
