---
name: Phase 140 — Guardian QA Settings summary (131 WS7)
overview: >
  Settings card surfaces the latest archived guardian-qa-smoke/regression run from
  data/guardian_qa_runs/ so operators validate Guardian without opening JSON by hand.
  Closes the optional Phase 131 WS7 gap after the 129–139 arc.
todos:
  - id: ws1-latest-loader
    content: "WS1: farmguardian — LoadLatestQARun scans guardian_qa_runs dir; QARunSummary pass/total"
    status: completed
  - id: ws2-api
    content: "WS2: GET /v1/guardian/qa/latest — JWT; 404 when no archives"
    status: completed
  - id: ws3-settings-ui
    content: "WS3: GuardianSettingsQARunCard — suite, model, pass count, step table; data-test settings-guardian-qa"
    status: completed
  - id: ws4-docs-tests
    content: "WS4: Mark 131 WS7 done; phase-140-closure vitest + Go unit test"
    status: completed
isProject: false
---

# Phase 140 — Guardian QA Settings summary

**Status:** **Shipped.** · **Depends on:** [131](phase_131_guardian_qa_harness.plan.md), [139](phase_139_guardian_docs_and_engineering.plan.md)

---

## Problem

Phase 131 archives full smoke answers to `data/guardian_qa_runs/`, but operators had to open JSON manually. Phase 131 WS7 was deferred optional UI.

---

## Acceptance

- [x] After `make guardian-qa-smoke`, Settings shows latest run (suite, model, pass/total, steps)
- [x] `GET /v1/guardian/qa/latest` returns 404 when directory empty
- [x] Card hidden when `AI_ENABLED=false`

---

## Non-goals

- Triggering smoke from UI (CLI/Makefile only)
- Storing QA runs in Postgres
