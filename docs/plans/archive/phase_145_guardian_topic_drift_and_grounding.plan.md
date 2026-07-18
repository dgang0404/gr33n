---
name: Phase 145 — Guardian topic drift & grounding depth
overview: >
  Phase 144 closed keyword-level residuals (gr33n-docs, apology tails, ec-ph blocklist).
  Run #3 still shows structural drift: models open on-topic then ramble into unrelated
  RAG chunks, dump raw source metadata, and pass heuristics. Phase 145 adds embedding
  relevance scoring, citation alignment, retrieval guardrails, answer tail hygiene, and
  eval archive enrichment — without full LLM-as-judge (deferred to 146 optional path).
todos:
  - id: ws1-embed-relevance
    content: "WS1: answer_relevance.go — embed cosine question↔answer + paragraph tail drift; turn debug + optional trim"
    status: completed
  - id: ws2-citation-alignment
    content: "WS2: Citation corpus alignment — answer terms vs cited excerpts; smoke/eval fail when tail uncited"
    status: completed
  - id: ws3-rag-retrieval-guard
    content: "WS3: field_guide retrieval guardrails — source_type/doc_path filters for agronomy prompts"
    status: completed
  - id: ws4-answer-tail-hygiene
    content: "WS4: Trim raw Sources: dumps, relative .md plan links, max grounded length by model profile"
    status: completed
  - id: ws5-eval-archive-enrich
    content: "WS5: QA archive stores citation excerpts; generalized smokeTopicDriftNote(category, prompt, answer, cites)"
    status: completed
  - id: ws6-smoke-run4-closure
    content: "WS6: Smoke run #4 post-145; update report; architecture §8.9; phase-145-closure.test.js"
    status: completed
isProject: false
---

# Phase 145 — Guardian topic drift & grounding depth

**Status:** **Shipped.** · **Depends on:** [144](phase_144_guardian_answer_quality_residuals.plan.md) · [131](phase_131_guardian_qa_harness.plan.md)

**Evidence:** Run #3 archive `20260707T175718_smoke_phi3-mini.json` — ec-ph **4174 chars** with endocrine tail; morning-walk **gr33n-docs** + apology (144 trims on *new* turns only).

**Next arc:** [146](phase_146_guardian_quality_loop_and_judge.plan.md) — optional self-critique judge, warmup ops, feedback→fixture loop.

---

## Problem statement (post-144)

| Gap | Run #3 symptom | Why keywords aren't enough |
|-----|----------------|----------------------------|
| Topic drift | ec-ph → endocrine / Lake Erie / Typha | New unrelated topics won't match a static blocklist |
| Citation dishonesty | Answer cites `[6] type=field_guide…endocr` chunks | Model dumps retrieved metadata instead of synthesizing |
| Retrieval pollution | Unrelated `field_guide` chunks in top-K | RAG returns semantically near but agronomically wrong docs |
| Ramble on small ctx | phi3 @ 4096 effective window | Long tails after correct opening paragraph |
| Eval blind spot | Archive has answer text only | Scorer can't compare answer to cited excerpts |

Phase 144 **keyword heuristics** are regression guards for *known* run #3 failures. Phase 145 makes drift detection **generalizable**.

---

## Design principles

1. **Reuse embed stack** — same `internal/rag/embed` client as RAG; no second LLM for v1 scoring.
2. **Prevention before detection** — tighten retrieval for `field_guide` before post-hoc fail.
3. **Fail in eval, warn in prod** — low relevance → `topic_drift_score` on turn debug; smoke/archive can hard-fail.
4. **Keep CPU laptop viable** — embed one short answer is cheap vs another 20 min LLM call.

---

## Workstreams

### WS1 — Embedding relevance scorer ✅

**Shipped:** `internal/farmguardian/answer_relevance.go` — `ScoreAnswerRelevanceFromText`, `GUARDIAN_RELEVANCE_MIN`; wired in chat finalize → turn debug (`question_answer_relevance`, `opening_tail_relevance`, `low_relevance`); `GuardianTurnDebug.vue`.

### WS2 — Citation corpus alignment ✅

**Shipped:** `internal/farmguardian/answer_citation_align.go` — `CitationAlignmentNote`; eval `Score` applies after field_guide pass when citations present; QA archive persists `citations[]` via `eval/runner.go` + `EvalQuestionScore.Citations`.

### WS3 — RAG retrieval guardrails ✅

**Shipped:** `internal/farmguardian/rag_filter.go` — `AgronomyQueryIntent`, `FilterRAGChunks`, `RAGRetrieveLimit`; wired in `retrieveChunks` with over-fetch + post-filter; turn debug `rag_filter_applied`; optional `GUARDIAN_RAG_MAX_CHUNKS_FIELD_GUIDE`.

### WS4 — Answer tail hygiene (structural) ✅

**Shipped:** `TrimSourceDump`, relative `.md` link sanitize in `answer_citation.go`, meta-correction v2 markers, `TrimGroundedAnswerLength` + `GUARDIAN_GROUNDED_ANSWER_MAX_CHARS`; wired in finalize chain with turn debug fields.

### WS5 — Eval harness enrichment ✅

**Shipped:** `SmokeTopicDriftNote` consolidates hygiene, relevance, citation alignment, and keyword regression; eval runner captures turn-debug relevance + citations into QA archives; runbook § Phase 145 drift notes; Settings QA **Relevance** column.

### WS6 — Closure ✅

**Shipped:** API rebuilt `-tags dev`; smoke run **#4** (`make guardian-qa-smoke`); [smoke report](../guardian-qa-smoke-report-20260707.md) run #4 section; architecture [§8.9](../farm-guardian-architecture.md#89-topic-drift--grounding-phase-145); `phase-145-closure.test.js`.

---

## Acceptance

- [x] Run #3 ec-ph archive **fails** `SmokeTopicDriftNote` (citation alignment + keyword regression in `score_smoke_quality_test.go`).
- [x] Run #3-style fixtures covered by `SmokeTopicDriftNote` unit tests (relevance, citation align, hygiene).
- [x] New turns persist without raw `Sources:` chunk dumps (`TrimSourceDump` in finalize chain).
- [x] QA archive JSON includes `citations[]` excerpts for smoke runs (`eval/runner.go`).
- [x] Smoke run #4 executed post-145 (see report § Phase 145 run #4); **3/4** heuristic pass — `smoke-ec-ph` eval client timeout after ~103 min CPU run (documented; not drift regression).

---

## Suggested implementation order

1. WS5 archive citation capture (enables WS2 tests).
2. WS1 embed relevance (mock tests → live embed).
3. WS2 citation alignment.
4. WS4 structural tail trim (quick wins).
5. WS3 RAG guardrails (higher integration risk).
6. WS6 smoke run #4 + docs.

---

## Non-goals (Phase 145)

- Full **LLM-as-judge** second pass (see Phase 146 optional path).
- Mandatory GPU CI gate.
- Re-ingest entire RAG corpus.
- Replacing phi3:mini on CPU.
