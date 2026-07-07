# Guardian QA smoke report — 2026-07-07 (complete)

**Suite:** `make guardian-qa-smoke MODEL=phi3:mini FARM_ID=1`  
**Machine:** CPU laptop (`cpu_laptop` profile)  
**Phase 143 closure run:** **#3** — finished 2026-07-07 12:12 → 13:57 local (~**105 min**)  
**Archive (run #3):** `data/guardian_qa_runs/20260707T175718_smoke_phi3-mini.json`  
**Prior archive (run #2):** `data/guardian_qa_runs/20260707T023729_smoke_phi3-mini.json`  
**Eval index:** `data/guardian_model_eval.json` (local, do not commit)

---

## Executive summary

| Metric | Run #1 (777s) | Run #2 (pre-143) | Run #3 (Phase 143) |
|--------|---------------|------------------|---------------------|
| Heuristic pass | **0/4** | **4/4** | **4/4** |
| HTTP completions | 0 | 4 | 4 |
| Mean latency | — | **~20.4 min/prompt** | **~25.0 min/prompt** |
| Grounded cite rate | 0% | **100%** (3/3 grounded) | **100%** (3/3 grounded) |
| Tools | partial | `walk_farm`, `list_unread_alerts` | `walk_farm` (log scrape) |

**Verdict (run #3):** Phase 143 hygiene ships — **no `gr33n.com` fake URLs**, **no `## Your task` template leak** on morning-walk, **ec-ph includes pH + EC** in the scored answer. Heuristics **4/4**. **[Phase 144](plans/phase_144_guardian_answer_quality_residuals.plan.md)** adds `gr33n-docs/` sanitize, apology-tail trim, and ec-ph drift heuristics for run #3 residuals.

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
| WS6 | Closure | This report + architecture §8.8 + `phase-143-closure.test.js` |

---

## Artifacts

| Path | Purpose |
|------|---------|
| `data/guardian_qa_runs/20260707T175718_smoke_phi3-mini.json` | Run #3 full answers (gitignored) |
| `data/guardian_qa_runs/20260707T023729_smoke_phi3-mini.json` | Run #2 regression fixture (gitignored) |
| `/tmp/guardian-qa-smoke-run3.log` | Runner stdout |
| `/tmp/gr33n-api.log` | API timing + tools |

**Settings:** `GET /v1/guardian/qa/latest` should show run #3 (4 passed).
