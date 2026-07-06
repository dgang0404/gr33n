# Phases 129–139 — Guardian next-level closure

**Status:** **Shipped** on `main` (Wave 5 complete with Phase 139).

**Canonical roadmap:** [`phase_129_139_guardian_next_level_roadmap.plan.md`](phase_129_139_guardian_next_level_roadmap.plan.md)

---

## Wave checklist

| Wave | Phases | Exit criteria |
|------|--------|---------------|
| **1 — Foundation** | 129, 130, 131 | Login-and-go awakening; runtime orchestration; `make guardian-qa-smoke` |
| **2 — Truth** | 132, 133, 134 | Read-tool router; source labels + trim banner; thumbs feedback |
| **3 — Intelligence** | 135, 136 | RAG lifecycle Settings; plant context bundle |
| **4 — Integration** | 137, 138 | Nudge → Farm counsel; split-host health; counsel/quick models |
| **5 — Engineering** | 139 | Architecture profiles; dev turn debugger; CI QA doc |

---

## Phase 139 deliverables

| WS | Deliverable | Verify |
|----|-------------|--------|
| **WS1** | Architecture §1 — Profile A/D, router flow, 129–139 link | `docs/farm-guardian-architecture.md` |
| **WS2** | Bootstrap + connectivity — 129–138 pointers, trimmed rituals | `docs/local-operator-bootstrap.md`, `docs/connectivity-requirements.md` |
| **WS3** | Dev turn inspector UI | `GuardianTurnDebug.vue`, `data-test="guardian-turn-debug"` |
| **WS4** | Turn debug API | `GET /v1/chat/sessions/{id}/turns/{n}/debug` (dev/auth_test) |
| **WS5** | Nightly CI pattern doc | `docs/ci-guardian-qa.md` |
| **WS6** | LLM-as-judge deferred in Phase 131 | `docs/plans/phase_131_guardian_qa_harness.plan.md` |
| **WS7** | Index + INSTALL link | `docs/phase-14-operator-documentation.md`, `INSTALL.md` |

---

## Roadmap acceptance (129–139)

- [x] Laptop: login → Farm counsel → morning walkthrough path documented (129–131)
- [x] `make guardian-qa-smoke` documented with archived JSON (`data/guardian_qa_runs/`)
- [x] No manual `ollama stop` ritual in operator bootstrap happy path
- [x] Settings: corpus freshness + readiness + model policy (135, 138)
- [x] Server profile: split embed/chat health (138)
- [x] `farm-guardian-architecture.md` leads with Profile A + D, not 70B-only (139)
- [x] Dev turn debugger after chat completes (139)
- [x] Optional self-hosted nightly QA workflow documented (139)

---

## Automated tests

| Test | Path |
|------|------|
| Turn debug builder | `internal/farmguardian/turn_debug_test.go` |
| Phase 139 closure | `ui/src/__tests__/phase-139-closure.test.js` |

---

## Deferred (explicit)

| Item | Where documented |
|------|------------------|
| LLM-as-judge for smoke/regression | Phase 131 non-goals; `docs/ci-guardian-qa.md` |
| Production turn debugger for all users | Phase 139 non-goals |
| Mandatory CI gate on every PR | Phase 139 non-goals |
