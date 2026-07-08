# Guardian QA smoke report — 2026-07-07 (complete)

**Suite:** `make guardian-qa-smoke MODEL=phi3:mini FARM_ID=1`  
**Machine:** CPU laptop (`cpu_laptop` profile)  
**Phase 143 closure run:** **#3** — finished 2026-07-07 12:12 → 13:57 local (~**105 min**)  
**Phase 145 closure run:** **#4** — finished 2026-07-07 22:43 → 00:27 local (~**103 min**)  
**Archive (run #3):** `data/guardian_qa_runs/20260707T175718_smoke_phi3-mini.json`  
**Archive (run #4):** `data/guardian_qa_runs/20260708T042653_smoke_phi3-mini.json`  
**Prior archive (run #2):** `data/guardian_qa_runs/20260707T023729_smoke_phi3-mini.json`  
**Eval index:** `data/guardian_model_eval.json` (local, do not commit)

---

## Executive summary

| Metric | Run #1 (777s) | Run #2 (pre-143) | Run #3 (Phase 143) | Run #4 (Phase 145) |
|--------|---------------|------------------|---------------------|---------------------|
| Heuristic pass | **0/4** | **4/4** | **4/4** | **3/4** |
| HTTP completions | 0 | 4 | 4 | 3 |
| Mean latency | — | **~20.4 min/prompt** | **~25.0 min/prompt** | **~17.0 min/prompt** (3 completed) |
| Grounded cite rate | 0% | **100%** (3/3 grounded) | **100%** (3/3 grounded) | **100%** (2/2 grounded completed) |
| Tools | partial | `walk_farm`, `list_unread_alerts` | `walk_farm` (log scrape) | RAG filter active |

**Verdict (run #3):** Phase 143 hygiene ships — **no `gr33n.com` fake URLs**, **no `## Your task` template leak** on morning-walk, **ec-ph includes pH + EC** in the scored answer. Heuristics **4/4**. **[Phase 144](plans/phase_144_guardian_answer_quality_residuals.plan.md)** adds `gr33n-docs/` sanitize, apology-tail trim, and ec-ph keyword drift heuristics for run #3 residuals.

**Verdict (run #4):** Phase 145 stack verified on CPU laptop — embed relevance + `citations[]` excerpts in archive, RAG filter + finalize trims active. Heuristics **3/4** (`smoke-ec-ph` **client timeout** after ~103 min total run; not a drift-scorer failure). Attempt #1 was 0/4 @ 401 (stale JWT). Attempt #2 completed with fresh token.

**Next:** **[Phase 146](plans/phase_146_guardian_quality_loop_and_judge.plan.md)** — optional GPU self-critique, eval ops, feedback→fixtures; consider longer eval client timeout for 4th grounded prompt on CPU.

---

## Phase 145 run #4

**Pre-run (2026-07-07):**

| Step | Command / note |
|------|----------------|
| Rebuild API | `go build -tags dev -o bin/api ./cmd/api/` |
| Restart API | `./bin/api` (or `make run`) |
| Refresh JWT | `source scripts/source-local-env.sh --refresh-eval-token` |
| Smoke | `make guardian-qa-smoke MODEL=phi3:mini FARM_ID=1` |
| Log | `/tmp/guardian-qa-smoke-run4.log` |

**Attempt #1 (19:54):** 0/4 — stale `GUARDIAN_EVAL_TOKEN` (401 on all prompts). Archive: `20260707T235458_smoke_phi3-mini.json`.

**Attempt #2 (22:43 → 00:27):** **3/4 pass** — Phase 145 finalize + RAG filter + `SmokeTopicDriftNote` active. Archive: `20260708T042653_smoke_phi3-mini.json`.

| # | ID | Grounded | Pass | Latency | Citations | Relevance (Q↔A) | Notes |
|---|-----|----------|------|---------|-----------|-----------------|-------|
| 1 | `smoke-cherry-forest` | No | ✅ | 12.4 min | 0 | 0.69 | On-topic forest-garden counsel |
| 2 | `smoke-morning-walk` | Yes | ✅ | 27.9 min | 2 | 0.51 | `citations[]` excerpts; no raw `Sources:` dump |
| 3 | `smoke-unread-alerts` | Yes | ✅ | 27.8 min | 5 | 0.61 | Concrete alerts; archive has alert excerpts |
| 4 | `smoke-ec-ph` | Yes | ❌ | — | — | — | Eval **client timeout** (4th prompt after ~103 min wall clock) |

**Phase 145 checks on run #4 archive:**

- ✅ `citations[]` excerpts present on grounded rows (morning-walk, unread-alerts)
- ✅ `question_answer_relevance` / `opening_tail_relevance` on completed rows
- ✅ No raw `Sources:` chunk dumps in persisted answers (finalize trim)
- ⚠️ `smoke-ec-ph` not scored — client timeout; re-run ec-ph alone or extend eval timeout (Phase 146 ops)
- ✅ Completed grounded rows: no `low_relevance` / `uncited_tail` / `citation_misaligned` in eval notes

**Eval summary line (run #4):**
```
phi3:mini: grounded cite 0% · decline 0% · proposal 0% · latency 1022070ms
```
*(Summary line skewed by ec-ph timeout with 0 ms latency; 2/2 completed grounded rows had citations.)*

---

## Per-prompt results (run #3 — Phase 143 closure)

| # | ID | Grounded | Pass | Latency | Citations | Tools / notes |
|---|-----|----------|------|---------|-----------|---------------|
| 1 | `smoke-cherry-forest` | No | ✅ | 15.7 min | 0 | On-topic forest-garden counsel |
| 2 | `smoke-morning-walk` | Yes | ✅ | 29.2 min | 0 | `walk_farm` log evidence; no `gr33n.com` / `## Your task` |
| 3 | `smoke-unread-alerts` | Yes | ✅ | 27.1 min | 4 | Concrete humidity, OHN, photoperiod alerts |
| 4 | `smoke-ec-ph` | Yes | ✅ | 27.9 min | 5 | EC + pH in opening; long off-topic tail (human review) |

**Eval summary line (run #3):**
```
phi3:mini: grounded cite 100% · decline 0% · proposal 0% · latency 1498128ms
```

---

## Phase 143 acceptance vs run #3

| Criterion | Run #2 | Run #3 |
|-----------|--------|--------|
| No instruction template leak (`## Your task`) | ❌ | ✅ trimmed / not present in archive |
| No `gr33n.com` fake URLs | ❌ | ✅ none in any answer |
| `smoke-ec-ph` mentions pH | ⚠️ EC-only pass | ✅ pH + EC in answer text |
| Eval warmup with `chat_model` | 503 (env tinyllama) | `POST /guardian/warmup` **200**; eval block warmup **timed out 5m** → inline warmup on send |
| WS4 regression on run #2 archive | — | ✅ fails in `score_smoke_quality_test.go` |

---

## Answer quality notes (run #3 human review)

### 1. `smoke-cherry-forest` — Good ✅
Cherry tree, goldenrod removal, blackberry/thorns, forest-garden framing. Sensible ungrounded counsel.

### 2. `smoke-morning-walk` — Heuristic pass; residual issues ⚠️
- **WS1/WS2 wins:** No `## Your task` leak; no `gr33n.com` links.
- **Residual:** Markdown links to `gr33n-docs/…` paths; meta apology paragraph (“I apologize for misunderstanding…”) — not in v1 heuristics; flag via [feedback runbook](guardian-feedback-review-runbook.md).
- **Farm signal:** Veg EC 1.2–2.0 mS/cm, flower room humidity / powdery mildew.
- **Tools:** `walk_farm` log evidence.

### 3. `smoke-unread-alerts` — Strong ✅
Four concrete alerts with actionable steps. Citations present.

### 4. `smoke-ec-ph` — Heuristic pass; severe tail hallucination ⚠️
- Opening cites lettuce EC **1.0–1.3 mS/cm** and pH context (meets WS4 `ph` + EC rule).
- Answer then diverges into unrelated endocrine-disruptor / Lake Erie content — **human review required**; consider thumbs-down `invented_data` / `other`.

---

## Infrastructure observations (run #3)

| Issue | Severity | Notes |
|-------|----------|-------|
| Eval grounded-block warmup **5m timeout** | Low | Inline warmup on first grounded send succeeded; `POST /guardian/warmup` returned **200** (WS3) |
| Stale `GUARDIAN_EVAL_TOKEN` in `.env.local` | Ops | First attempt 0/4 @ 401; re-run with fresh login JWT |
| phi3 latency ~25 min/prompt | Operational | Expected on CPU; `LLM_TIMEOUT_SECONDS=1500` / `1800` grounded |
| `guardian_counsel_model` column | Medium | Background tick noise if Phase 138 migration not applied locally |

---

## Run comparison

| Run | Started | Timeouts | API | Result |
|-----|---------|----------|-----|--------|
| **#1** | 20:12 | 777s | Stale binary (warmup 404) | 0/4 `llm_timeout` |
| **#2** | 21:15 | 1500 / 1800 | `./bin/api` + phi3 pre-warm | **4/4 pass** (quality gaps) |
| **#3** | 12:12 | 1500 / 1800 | `./bin/api` rebuilt `-tags dev` + fresh JWT | **4/4 pass** (Phase 143 closure) |

**Pre-run fixes (run #3):** `go build -tags dev -o bin/api`, restart API, fresh JWT (expired token in `.env.local`), Phase 143 WS1–5 on `main`.

---

## Phase 143 — shipped workstreams

| WS | Title | Outcome |
|----|-------|---------|
| WS1 | Instruction leak guard | `TrimInstructionLeak` before persist |
| WS2 | Citation URL hygiene | `SanitizeCitationURLs`; no `gr33n.com` in run #3 |
| WS3 | Warmup + eval model | `chat_model` on warmup; 200 on `/guardian/warmup` |
| WS4 | Smoke heuristic hardening | Leak/URL/pH gates in `score.go` |
| WS5 | Feedback checklist | [Runbook § Smoke quality](guardian-feedback-review-runbook.md#smoke-quality-checklist-phase-143) |
| WS6 | Closure | Architecture §8.8 + `phase-143-closure.test.js` |

## Phase 145 — shipped workstreams (run #4)

| WS | Title | Outcome |
|----|-------|---------|
| WS1 | Embed relevance | `answer_relevance.go`; turn debug `low_relevance` |
| WS2 | Citation alignment | `CitationAlignmentNote`; archive `citations[]` |
| WS3 | RAG guardrails | `FilterRAGChunks`; `rag_filter_applied` |
| WS4 | Tail hygiene | `TrimSourceDump`, length cap, relative `.md` sanitize |
| WS5 | Eval drift scorer | `SmokeTopicDriftNote`; QA Relevance column |
| WS6 | Closure | Run #4 + architecture §8.9 + `phase-145-closure.test.js` |

---

## Artifacts

| Path | Purpose |
|------|---------|
| `data/guardian_qa_runs/20260707T175718_smoke_phi3-mini.json` | Run #3 full answers (gitignored) |
| `data/guardian_qa_runs/20260707T023729_smoke_phi3-mini.json` | Run #2 regression fixture (gitignored) |
| `/tmp/guardian-qa-smoke-run3.log` | Runner stdout |
| `/tmp/guardian-qa-smoke-run4.log` | Runner stdout (attempt #2 complete) |
| `data/guardian_qa_runs/20260708T042653_smoke_phi3-mini.json` | Run #4 full answers (gitignored) |
| `/tmp/gr33n-api-ws6.log` | API log (run #4) |
| `/tmp/gr33n-api.log` | API timing + tools |

**Settings:** `GET /v1/guardian/qa/latest` should show run #4 (3 passed, 1 timeout).
