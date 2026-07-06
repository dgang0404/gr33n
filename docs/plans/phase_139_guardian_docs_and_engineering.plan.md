---
name: Phase 139 — Guardian docs, turn debugger & engineering CI
overview: >
  Architecture doc reflects laptop + server profiles; dev turn inspector surfaces
  tools/chunks/trim/logs; optional nightly guardian-qa-smoke workflow documented.
  Closes P0 doc drift and P3 engineering gaps. LLM-as-judge explicitly deferred.
todos:
  - id: ws1-architecture-refresh
    content: "WS1: farm-guardian-architecture.md — Profile A laptop CPU (phi3/tinyllama) + Profile D GPU server; request flow diagram with router + awakening"
    status: completed
  - id: ws2-bootstrap-refresh
    content: "WS2: local-operator-bootstrap + connectivity-requirements — point to 129–138 roadmap; remove ritual steps"
    status: completed
  - id: ws3-turn-debugger
    content: "WS3: Dev-only Guardian turn inspector — last turn meta in AUTH_MODE=dev: tools_planned, chunks, trim_summary, request_id, model; data-test guardian-turn-debug"
    status: completed
  - id: ws4-api-debug-endpoint
    content: "WS4: GET /v1/chat/sessions/{id}/turns/{n}/debug (dev/auth_test only) — prompt budget breakdown, read tool ids"
    status: completed
  - id: ws5-nightly-ci-doc
    content: "WS5: .github or docs/ci-guardian-qa.md — self-hosted runner, make guardian-qa-smoke, artifact upload guardian_qa_runs/"
    status: completed
  - id: ws6-llm-judge-defer
    content: "WS6: Document in phase 131 — LLM-as-judge out of scope; human review of 134 feedback + 131 JSON"
    status: completed
  - id: ws7-roadmap-closure
    content: "WS7: phase-129-139-closure checklist; update phase-14-operator-documentation.md index"
    status: completed
isProject: false
---

# Phase 139 — Guardian docs & engineering CI

**Status:** **Shipped.** · **Depends on:** [131](phase_131_guardian_qa_harness.plan.md), [133](phase_133_guardian_answer_grounding_honesty.plan.md)

**Closure:** [`phase-129-139-closure.md`](phase-129-139-closure.md)

---

## WS1 — Architecture doc refresh

`farm-guardian-architecture.md` §1 table becomes:

| Profile | Hardware | Chat | Farm counsel | Awakening |
|---------|----------|------|--------------|-----------|
| **A — Laptop dev** | 16 GB CPU | tinyllama quick | phi3 counsel | Required |
| **D — Server** | GPU 24GB | 8b+ | 8b/70b counsel | Optional warm |

Remove implicit "70B only" mental model from opening paragraphs.

Add § **Phases 129–139** pointer to [roadmap](phase_129_139_guardian_next_level_roadmap.plan.md).

---

## WS3 — Turn debugger (dev)

`GuardianChatPanel` when `import.meta.env.DEV` or capabilities flag:

```
Last turn debug
  request_id: 3f74000b-…
  tools: walk_farm, summarize_unread_alerts
  rag_chunks: 5 (field_guide×3, platform×2)
  trim: history 20→8
  model: phi3:mini · 4096 effective
```

Source: `done` event debug block (extend 133 trim_summary) + server debug endpoint.

---

## WS5 — Nightly CI (documented pattern)

```yaml
# docs/ci-guardian-qa.md example
on:
  schedule: [{ cron: '0 6 * * 1' }]  # weekly
jobs:
  guardian-smoke:
    runs-on: [self-hosted, ollama]
    steps:
      - run: make guardian-qa-smoke MODEL=phi3:mini
      - uses: actions/upload-artifact@v4
        with:
          name: guardian-qa
          path: data/guardian_qa_runs/
```

Not enabled on GitHub-hosted runners (no Ollama).

---

## WS6 — LLM-as-judge

Explicitly **deferred**. Quality loop v1:

1. `make guardian-qa-smoke` heuristics
2. Operator thumbs (134)
3. Human agronomy review export

Revisit when GPU CI stable.

---

## Acceptance

- [x] New developer reads architecture doc and understands laptop vs server
- [x] Dev mode shows turn debug after chat completes
- [x] Roadmap doc linked from INSTALL.md Guardian section
- [x] phase-129-139 closure checklist all boxes documented

---

## Non-goals

- Production turn debugger for all users (dev/auth_test only)
- Mandatory CI gate on every PR
