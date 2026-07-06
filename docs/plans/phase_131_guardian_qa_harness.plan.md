---
name: Phase 131 — Guardian QA harness (smoke prompts, recorded answers, regression tiers)
overview: >
  Extend cmd/guardian-eval into a tiered QA system: operator-authored smoke prompts
  (cherry ungrounded, morning walkthrough, alerts, RAG) run sequentially against a
  live API, full answers recorded to JSON + optional markdown run logs, heuristic
  scoring plus log-correlation hints (walk_farm, citations). Manual UI checklist
  becomes the same fixtures with a --manual flag that prints copy-paste steps.
todos:
  - id: ws1-fixture-tiers
    content: "WS1: eval.Fixtures split — SmokeFixtures (4-step order), RegressionFixtures (existing 24), UngroundedFixtures (cherry); YAML or Go table with id, prompt, grounded, model, expect_tool, pass_criteria"
    status: pending
  - id: ws2-record-answers
    content: "WS2: EvalQuestionScore + report — store answer text, citations, proposals, error, http_status; write data/guardian_qa_runs/{timestamp}_{suite}_{model}.json"
    status: pending
  - id: ws3-sequential-runner
    content: "WS3: RunSuite — one prompt at a time (no parallel); timeout from GUARDIAN_EVAL_TIMEOUT_SECONDS; optional POST /guardian/warmup before grounded block"
    status: pending
  - id: ws4-log-correlation
    content: "WS4: Eval requests send X-Eval-Run-Id header; document grep for walk_farm/summarize in /tmp/gr33n-api.log; optional post-run log scrape for expect_tool"
    status: pending
  - id: ws5-make-targets
    content: "WS5: make guardian-qa-smoke (4 prompts), make guardian-qa-regression (full), make guardian-eval unchanged alias; INSTALL.md tier table"
    status: pending
  - id: ws6-manual-mode
    content: "WS6: guardian-qa --manual prints UI checklist (farm context on/off, model, wait guidance) from same fixture source — single source of truth"
    status: pending
  - id: ws7-ui-optional
    content: "WS7 (optional): Settings → last QA run summary + link to report file; not required for v1"
    status: pending
  - id: ws8-docs
    content: "WS8: Phase 128 checklist points at make guardian-qa-smoke; local-operator-bootstrap QA section"
    status: pending
isProject: false
---

# Phase 131 — Guardian QA harness (smoke prompts, recorded answers)

**Status:** planned

**Depends on:** [Phase 130](phase_130_guardian_runtime_orchestration.plan.md) WS5 (eval timeout), ideally 129 WS2 (warmup before smoke)

**Extends:** [Phase 122](phase_122_guardian_model_eval_and_context_budget.plan.md) `cmd/guardian-eval`

---

## Should we automate this?

**Yes.** Your pasted checklist is exactly what a QA harness should be — but split into **tiers**:

| Tier | Command | When | Duration (CPU laptop) |
|------|---------|------|-------------------------|
| **Smoke** | `make guardian-qa-smoke` | After Guardian changes, before demo | ~30–90 min (4 prompts, sequential) |
| **Regression** | `make guardian-qa-regression` | Weekly / pre-release | Hours (24+ prompts) |
| **Manual** | `make guardian-qa-manual` | Prints same steps for UI validation | Human-paced |

**Not in default CI** — too slow and LLM-dependent. Optional nightly workflow with `//go:build ollama` or self-hosted runner.

---

## What exists today (gaps)

| Have | Missing |
|------|---------|
| 24 grounded fixtures in `eval.Fixtures()` | Your smoke order + cherry **ungrounded** prompt |
| Heuristic pass/fail scores | **Full answer text** in report |
| `data/guardian_model_eval.json` summary | Per-run archive `guardian_qa_runs/` |
| Model quality badges in UI | Log correlation (`walk_farm` fired?) |
| 120s HTTP timeout | Works on GPU; **fails on CPU** (Phase 130) |

---

## Smoke suite (your suggested test order)

| Step | ID | Farm context | Model | Prompt | Tests |
|------|-----|--------------|-------|--------|-------|
| 1 | `smoke-cherry-forest` | **Off** | phi3:mini | Cherry / goldenrod / blackberry forest garden prompt | Ungrounded path, no embed fight |
| 2 | `smoke-morning-walk` | On | phi3:mini | What should I check first on a morning walkthrough of this farm today? | Snapshot + `walk_farm` + trim |
| 3 | `smoke-unread-alerts` | On | phi3:mini | Summarize my unread alerts and what I should do about each one. | Read tools + grounded answer |
| 4 | `smoke-ec-ph` | On | phi3:mini | What does our operational documentation say about EC and pH targets for leafy greens here? | RAG corpus |

**Pass criteria (heuristic v1):**

| ID | Auto pass if |
|----|----------------|
| smoke-cherry-forest | answer len > 80; mentions cherry OR goldenrod OR blackberry; no timeout |
| smoke-morning-walk | answer len > 40; log contains `tool_id=walk_farm` OR answer mentions alert/zone/device |
| smoke-unread-alerts | answer mentions seed alert subject OR "alert"; len > 40 |
| smoke-ec-ph | citation count > 0 OR `[1]` in answer OR mentions EC/pH |

Human can override via recorded JSON review.

---

## Extended regression prompts (from your list)

Add to `RegressionFixtures()` (grounded, farm 1, phi3:mini):

| ID | Prompt |
|----|--------|
| `farm-urgent-issue` | What is the most urgent issue I should address on this farm right now? Two sentences max. |
| `farm-zones-setpoints` | Which zones have automation or setpoints I should review before the next grow cycle? |
| `farm-readings-oob` | For the demo farm zones, what environmental readings look out of band or worth a closer look? |
| `farm-fertigation-plain` | Explain the fertigation setup on this farm in plain language — tanks, programs, and what an operator should watch. |
| `farm-active-grows` | What crops or grows are active on this farm, and what stage are they in? |
| `farm-behind-schedule` | Are any plants or beds behind where they should be for this time of year? What would you check? |
| `fg-pi-relay` | What field-guide or platform guidance applies to wiring a Pi relay for lights on this kind of farm? |
| `write-moisture` | If soil moisture in a bed drops below target, what would you recommend I do on this farm — and can you suggest a concrete next step? |
| `farm-restock` | List anything that looks low on stock or restock priority for this farm. |

Overlap with existing fixtures (farm-alerts, farm-low-stock) — dedupe or alias IDs.

---

## WS2 — Recorded responses

Extend `EvalQuestionScore`:

```json
{
  "id": "smoke-morning-walk",
  "prompt": "What should I check first...",
  "grounded": true,
  "model": "phi3:mini",
  "passed": true,
  "latency_ms": 923000,
  "answer": "Start with unread alerts in Flower Room...",
  "citations": [{"title": "..."}],
  "proposals": [],
  "error": "",
  "notes": "walk_farm seen in logs",
  "log_evidence": ["tool_id=walk_farm"]
}
```

Write:

- `data/guardian_model_eval.json` — summary (existing, for UI badges)
- `data/guardian_qa_runs/20260706T153900_smoke_phi3-mini.json` — full run archive
- Optional: `data/guardian_qa_runs/latest_smoke.md` — human-readable for demo postmortems

**Do not commit** run archives to git — `.gitignore` `data/guardian_qa_runs/`; commit fixture definitions only.

---

## WS3 — Sequential runner

```go
// RunSuite runs prompts one at a time — mirrors "one message at a time" UI guidance.
func RunSuite(ctx context.Context, api *APIClient, model string, fixtures []Question) ([]ScoreResult, error)
```

Before steps 2–4 (grounded block):

- Optional `POST /guardian/warmup` with `farm_counsel` (Phase 129)
- Sleep/poll until health `ready` or timeout

Between prompts: no parallel goroutines.

---

## WS4 — Log correlation

Eval client sets header: `X-Guardian-Eval-Id: smoke-morning-walk`

API logs already include `request_id` — eval runner logs `request_id` from response header if exposed, or grep:

```bash
grep "tool_id=walk_farm" /tmp/gr33n-api.log | tail -5
```

Post-run optional: `scripts/guardian-qa-scrape-logs.sh --since 30m --expect walk_farm`

---

## WS5 — Make targets

```makefile
guardian-qa-smoke:    ## 4-prompt smoke (sequential, recorded)
guardian-qa-regression: ## full fixture set
guardian-qa-manual:   ## print UI checklist only
guardian-eval:        ## alias regression (backward compat)
```

Env:

```bash
GUARDIAN_EVAL_TOKEN=<jwt>
GUARDIAN_EVAL_TIMEOUT_SECONDS=1800   # or inherit LLM_TIMEOUT
GUARDIAN_EVAL_FARM_ID=1
GUARDIAN_EVAL_LOG=/tmp/gr33n-api.log
```

---

## WS6 — Manual mode = same fixtures

`go run ./cmd/guardian-eval/ -manual -suite smoke` prints:

```markdown
## Step 1 — Quick chat (farm context OFF)
Model: phi3:mini
Prompt: [cherry text]
Wait: Generating… may take 8–15 min on CPU
Pass if: mentions goldenrod or blackberry; no timeout
```

Eliminates drift between automated and manual docs.

---

## Relationship to other phases

| Phase | Role |
|-------|------|
| 128 WS3 manual UI | Becomes optional after smoke passes in CLI |
| 128 WS4 guardian-eval | Becomes `guardian-qa-regression` |
| 129 awakening | Warmup before smoke grounded block |
| 130 timeout + embed | Smoke completes without false HTTP failures |
| 122 eval badges | Fed by regression summary JSON |

---

## Non-goals

- LLM-as-judge scoring (defer; heuristics + human review of recorded JSON first)
- Running smoke on every `go test ./...`
- Storing QA runs in Postgres
- Replacing browser E2E (Playwright) — complementary

---

## Acceptance

1. `make guardian-qa-smoke MODEL=phi3:mini` writes JSON with **full answers** for 4 steps.
2. `make guardian-qa-manual` output matches smoke fixture prompts exactly.
3. Morning walkthrough step records `log_evidence` containing `walk_farm` when API log available.
4. Cherry step runs with `farm_id` omitted (ungrounded).
5. Phase 128 manual checklist references Phase 131 commands.

---

## Implementation order

1. WS1 + WS2 + WS3 + WS5 (fixtures + recording + smoke command)
2. WS4 (log header + scrape script)
3. WS6 + WS8 (manual mode + docs)
4. WS7 optional UI

---

## Non-goals (Phase 131 + 139)

- **LLM-as-judge** — automated grading of answer quality by a second LLM is **out of scope** for v1. Quality loop instead:
  1. `make guardian-qa-smoke` heuristic pass/fail + archived full answers
  2. Operator thumbs + reasons ([Phase 134](phase_134_guardian_answer_feedback.plan.md))
  3. Human agronomy review via feedback export JSON
- Revisit LLM-as-judge when **self-hosted GPU CI** is stable (`docs/ci-guardian-qa.md`).
