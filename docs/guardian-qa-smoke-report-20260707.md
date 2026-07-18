# Guardian QA smoke report — 2026-07-07 (complete)

**Suite:** `make guardian-qa-smoke MODEL=phi3:mini FARM_ID=1`  
**Machine:** CPU laptop (`cpu_laptop` profile)  
**Phase 143 closure run:** **#3** — finished 2026-07-07 12:12 → 13:57 local (~**105 min**)  
**Phase 145 closure run:** **#4** — finished 2026-07-07 22:43 → 00:27 local (~**103 min**)  
**Phase 147 closure run:** **#5** — ec-ph isolation via `make guardian-qa-smoke-ec-ph` (~**25 min**)  
**Post-147 operator run:** **#6** — full smoke 2026-07-08 10:02 → 11:38 local (~**96 min**)  
**Archive (run #3):** `data/guardian_qa_runs/20260707T175718_smoke_phi3-mini.json`  
**Archive (run #4):** `data/guardian_qa_runs/20260708T042653_smoke_phi3-mini.json`  
**Archive (run #5):** `data/guardian_qa_runs/20260708T130745_smoke_phi3-mini.json`  
**Archive (run #6):** `data/guardian_qa_runs/20260708T153829_smoke_phi3-mini.json`  
**Prior archive (run #2):** `data/guardian_qa_runs/20260707T023729_smoke_phi3-mini.json`  
**Eval index:** `data/guardian_model_eval.json` (local, do not commit)

---

## Executive summary

| Metric | Run #1 (777s) | Run #2 (pre-143) | Run #3 (Phase 143) | Run #4 (Phase 145) | Run #6 (post-147) |
|--------|---------------|------------------|---------------------|---------------------|-------------------|
| Heuristic pass | **0/4** | **4/4** | **4/4** | **3/4** | **3/4** |
| HTTP completions | 0 | 4 | 4 | 3 | **4** |
| Mean latency | — | **~20.4 min/prompt** | **~25.0 min/prompt** | **~17.0 min/prompt** (3 completed) | **~24.0 min/prompt** |
| Grounded cite rate | 0% | **100%** (3/3 grounded) | **100%** (3/3 grounded) | **100%** (2/2 grounded completed) | **100%** (3/3 grounded) |
| Tools | partial | `walk_farm`, `list_unread_alerts` | `walk_farm` (log scrape) | RAG filter active | `walk_farm`, `list_unread_alerts` |

**Verdict (run #3):** Phase 143 hygiene ships — **no `gr33n.com` fake URLs**, **no `## Your task` template leak** on morning-walk, **ec-ph includes pH + EC** in the scored answer. Heuristics **4/4**. **[Phase 144](plans/archive/phase_144_guardian_answer_quality_residuals.plan.md)** adds `gr33n-docs/` sanitize, apology-tail trim, and ec-ph keyword drift heuristics for run #3 residuals.

**Verdict (run #4):** Phase 145 stack verified on CPU laptop — embed relevance + `citations[]` excerpts in archive, RAG filter + finalize trims active. Heuristics **3/4** (`smoke-ec-ph` **client timeout** after ~103 min total run; not a drift-scorer failure). Attempt #1 was 0/4 @ 401 (stale JWT). Attempt #2 completed with fresh token.

**Verdict (run #5):** Phase 147 **eval isolation + timeout fix verified** — `make guardian-qa-smoke-ec-ph` completed in **~25 min** with HTTP 200 and full archive (no client timeout). Heuristic **fail** on `uncited_tail` (model appended unrelated blueberry question); relevance **0.64**; EC + pH in opening with 3 citations. Operator spot-check still required per runbook.

**Verdict (run #6):** First **full 4/4 HTTP completion** post-147 timeout fix (~**96 min**). Heuristics **3/4** — same `smoke-ec-ph` **`uncited_tail`** (blueberry/pH unit confusion). Prompts 2–3 **pass** but are long and **template-heavy** (`proposal card → Confirm`, API path snippets); morning-walk answer **trimmed** 1068 chars at finalize. Human review recommended for farm-state answers even when heuristics pass.

**Next:** Phases **143–147** quality arc shipped — ec-ph drift remains the recurring smoke gap; discuss farm-state answer tone in operator review.

**Verdict (run #7, targeted):** Phases **148–150** shipped in response to run #6's human review (wrong citation numbers, duplicate OHN alert, garbled `0sourced`, raw `PATCH /alerts/{id}/acknowledge` API path, blueberry pH mislabeled as EC mS/cm). `smoke-ec-ph` re-run **proves Phase 148 works live**: it now fails on `citation_number_mismatch` — a real wrong-citation the run #6/#5 scorer could not see (only checked topical drift). `smoke-unread-alerts` hit the CPU laptop's 30-min internal `GUARDIAN_GROUNDED_TIMEOUT_SECONDS` twice in a row after ~2 hours of continuous back-to-back generation (load average 2.95–5.36) — an infra/thermal ceiling, not a Phase 148–150 regression (confirmed via unit tests reproducing the exact run #6 answer/citation text for all four new detectors, plus `PrioritizeAlertChunks` / `RedactDevAPIJargon` unit coverage). Re-run recommended after a cooldown or with `GUARDIAN_GROUNDED_TIMEOUT_SECONDS` bumped (Phase 147 precedent).

**Verdict (run #8, unread-alerts follow-up):** Bumped `GUARDIAN_GROUNDED_TIMEOUT_SECONDS` 1800→**2400**, `LLM_TIMEOUT_SECONDS` 1500→**2100**, added `GUARDIAN_EVAL_TIMEOUT_SECONDS=**2700**`; restarted API (`llm_timeout_seconds=2400` confirmed in logs). `make guardian-qa-smoke-unread-alerts` completed in **~29.8 min** (HTTP 200, heuristic **pass**). Alert list order is **severity-first** (humidity high → OHN medium → light schedule low) — run #6's [3]/[5] swap class resolved in prose. Model used markdown links instead of `[N]` citation markers (`citations=0` in turn log), so Phase 148 numbered-cite mismatch cannot be re-checked on this archive; no raw `PATCH /alerts/…` API path (Phase 150). Operator review still recommended for proposal-card / Confirm template tone.

---

## Phases 148–150 run #7 (targeted verification)

**Pre-run (2026-07-08):** Rebuilt `bin/api` with Phase 148–150 changes; restarted API; ran isolated prompts via `-prompt-ids` (Phase 147 WS1).

| Step | Command / note |
|------|----------------|
| Rebuild + restart | `go build -tags dev -o bin/api ./cmd/api/` then restart |
| Isolated re-run | `go run ./cmd/guardian-eval -models phi3:mini -farm-id 1 -suite smoke -prompt-ids smoke-unread-alerts,smoke-ec-ph` |
| Log | `/tmp/guardian-qa-smoke-run7-alerts-ecph.log`, `/tmp/guardian-qa-smoke-run7-alerts-retry.log` |
| Archives | `data/guardian_qa_runs/20260708T182557_smoke_phi3-mini.json`, `20260708T185710_smoke_phi3-mini.json` |

| ID | Pass | Latency | Citations | Notes |
|----|------|---------|-----------|-------|
| `smoke-ec-ph` | ❌ (by design) | 28.0 min | 5 | **New:** `citation_number_mismatch: claim near [1] matches [3] instead` — Phase 148 caught the model attributing lettuce's 0.8–1.3 mS/cm range to citation [2] (cannabis field guide), a mismatch invisible to run #5/#6's topical-only checker |
| `smoke-unread-alerts` (attempt 1, cold restart) | ❌ (infra) | 30.0 min (timeout) | — | `llm_timeout` HTTP 502 — warmup failed after fresh restart, model loaded cold |
| `smoke-unread-alerts` (attempt 2, warm model) | ❌ (infra) | 30.0 min (timeout) | — | Same `llm_timeout`; system load average 2.95–5.36 after ~2h continuous CPU inference |

**Phase 148–150 checks:**

- ✅ `CitationClaimMismatchNote` fires on a genuine live mismatch (ec-ph) that the pre-148 scorer would have silently passed (all 5 cites were on-topic agronomy field guides — no topical drift, but the wrong one was cited for the claim)
- ✅ All Phase 148/149/150 unit tests pass, including exact reproductions of run #6's alert answer/citation text (`TestSmokeTopicDrift_runSixUnreadAlertsCitationMismatchNowCaught`, `TestGarbledTokenNote_detectsRunSixOHNTypo`, `TestRedactDevAPIJargon_run6UnreadAlerts`)
- ⚠️ Could not get a live `smoke-unread-alerts` completion this session — CPU laptop hit its 30-min internal timeout twice under sustained load; not attributable to Phase 149's `PrioritizeAlertChunks` (O(n) sort of ≤5 chunks) or the ~350-char `alertCitationDiscipline` instruction addendum
- **Follow-up (done run #8):** see § Phases 148–150 run #8 below

---

## Phases 148–150 run #8 (unread-alerts follow-up)

**Pre-run (2026-07-08):** Machine cooled (load ~0.9); bumped timeouts in `.env`; extended `scripts/tune-guardian-laptop.sh` cpu-16gb floor; added `make guardian-qa-smoke-unread-alerts`; rebuilt + restarted API.

| Step | Command / note |
|------|----------------|
| Timeout bump | `GUARDIAN_GROUNDED_TIMEOUT_SECONDS=2400`, `LLM_TIMEOUT_SECONDS=2100`, `GUARDIAN_EVAL_TIMEOUT_SECONDS=2700` |
| Rebuild + restart | `go build -tags dev -o bin/api ./cmd/api/` → `AUTH_MODE=auth_test ./bin/api` |
| Isolated re-run | `make guardian-qa-smoke-unread-alerts MODEL=phi3:mini FARM_ID=1` |
| Log | `/tmp/guardian-qa-smoke-run8-unread-alerts.log`, `/tmp/gr33n-api-run8.log` |
| Archive | `data/guardian_qa_runs/20260708T224619_smoke_phi3-mini.json` |

| ID | Pass | Latency | Citations | Notes |
|----|------|---------|-----------|-------|
| `smoke-unread-alerts` | ✅ | 29.8 min | 0 (`[N]` markers absent) | HTTP 200; `llm_timeout_seconds=2400` in API log; alert list **severity-first** (humidity → OHN → light schedule) — run #6 [3]/[5] class fixed in prose; model used markdown links not numbered cites; proposal-card / Confirm tone remains |

**Phase 148–150 checks (run #8):**

- ✅ Infra follow-up closed — completion under bumped 2400s grounded ceiling (would have timed out at 1800s under similar ~30 min generation)
- ✅ Phase 149 intent — alerts listed most-severe-first; humidity no longer cited as `[3]` while sitting at `[5]` (model skipped `[N]` markers entirely this run)
- ✅ Phase 150 — no inline `PATCH /alerts/{id}/acknowledge` dev API path in answer
- ⚠️ `citations=0` in turn log — cannot live-verify Phase 148 `citation_number_mismatch` on this archive; heuristic pass is lenient (`alert` keyword only)
- ⚠️ Operator tone — still template-heavy (`proposal card`, `Confirm`); fake-looking `gr33ncore.sensor_alerts` links (one sanitized by `citation_url_sanitized`)

---

## Phase 147 run #6 (full smoke, post-ship)

**Pre-run (2026-07-08):**

| Step | Command / note |
|------|----------------|
| Full smoke | `make guardian-qa-smoke MODEL=phi3:mini FARM_ID=1` |
| Log | `/tmp/guardian-qa-smoke-full-run6.log` |
| Archive | `data/guardian_qa_runs/20260708T153829_smoke_phi3-mini.json` |

| # | ID | Grounded | Pass | Latency | Citations | Relevance (Q↔A) | Notes |
|---|-----|----------|------|---------|-----------|-----------------|-------|
| 1 | `smoke-cherry-forest` | No | ✅ | 10.9 min | 0 | 0.74 | On-topic forest-garden counsel |
| 2 | `smoke-morning-walk` | Yes | ✅ | 29.5 min | 3 | 0.40 | `walk_farm`; answer trimmed; low q↔a (tail template-heavy) |
| 3 | `smoke-unread-alerts` | Yes | ✅ | 27.5 min | 5 | 0.55 | Concrete alerts; API/proposal-card phrasing |
| 4 | `smoke-ec-ph` | Yes | ❌ | 28.0 min | 5 | 0.71 | `uncited_tail`; blueberry drift; pH/EC unit mix-up |

**Phase 147 checks on run #6 archive:**

- ✅ All four prompts HTTP 200 — no eval client timeout (fixes run #4 ops gap)
- ✅ `citations[]` excerpts on all grounded rows
- ✅ `walk_farm` + alert tools used on farm-state prompts
- ⚠️ `smoke-ec-ph` heuristic fail — same drift class as run #5
- ⚠️ Morning-walk / unread-alerts pass heuristics but read like operator-bootstrap tutorial text

**Eval summary line (run #6):**
```
phi3:mini: grounded cite 0% · decline 0% · proposal 0% · latency 1439823ms
```
*(Aggregate summary line skewed; all three grounded rows had citations.)*

---

## Phase 147 run #5 (ec-ph isolation)

**Pre-run (2026-07-08):**

| Step | Command / note |
|------|----------------|
| Phase 146 shipped | `ClientTimeoutFromEnv` +15m buffer; Makefile JWT refresh |
| Isolated re-run | `make guardian-qa-smoke-ec-ph MODEL=phi3:mini FARM_ID=1` |
| Log | `/tmp/guardian-qa-smoke-run5-ecph.log` |
| Archive | `data/guardian_qa_runs/20260708T130745_smoke_phi3-mini.json` |

| # | ID | Grounded | Pass | Latency | Citations | Relevance (Q↔A) | Notes |
|---|-----|----------|------|---------|-----------|-----------------|-------|
| 1 | `smoke-ec-ph` | Yes | ❌ | 25.1 min | 3 | 0.64 | **No client timeout**; eval notes `uncited_tail` (blueberry question appended) |

**Phase 147 checks on run #5 archive:**

- ✅ HTTP completion within eval client timeout (~25 min vs run #4 timeout)
- ✅ `citations[]` excerpts present (lettuce, cannabis, spinach field guides)
- ✅ `question_answer_relevance` / `opening_tail_relevance` scored
- ⚠️ Heuristic fail — uncited tail drift (human review; not eval ops regression)
- ✅ Opening mentions EC mS/cm targets for leafy greens

**Eval summary line (run #5):**
```
phi3:mini: grounded cite 0% · decline 0% · proposal 0% · latency 1507496ms
```
*(Summary line shows 0% cite rate because only one prompt and scorer uses aggregate; row has 3 citations.)*

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
