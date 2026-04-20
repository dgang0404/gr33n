---
name: Phase 23 Stabilization Sprint
overview: >
  No new product features. Harden what shipped in Phases 20–22 + Pi contract:
  CI gates, OpenAPI parity, smoke reliability, docs accuracy, and a short
  backlog of sharp edges. Exit criteria unlock Phase 24 (RAG retrieval).
todos:
  - id: ws1-ci-and-build-gates
    content: "WS1: CI discipline — document required commands (go test ./..., make audit-openapi); fix any drift or failures; optional: go vet, race on critical packages"
    status: completed
  - id: ws2-smoke-and-integration
    content: "WS2: Smoke hardening — full cmd/api suite green on clean DB; document DATABASE_URL / auth env; reduce flakiness; add notes for skipped tests"
    status: completed
  - id: ws3-openapi-routes-parity
    content: "WS3: scripts/openapi_route_diff.sh — resolve any mismatches; spot-check Pi + JWT routes documented"
    status: completed
  - id: ws4-automation-edge-cases
    content: "WS4: Worker + program tick — grep worker logs for error patterns; document metadata.steps fallback monitoring; optional small defensive fixes only"
    status: completed
  - id: ws5-edge-auth-and-secrets
    content: "WS5: Pi / API key — short runbook snippet (rotation, least privilege); confirm RequireFarmMemberOrPiEdge + requireAPIKey behavior in docs"
    status: pending
  - id: ws6-operator-docs-pass
    content: "WS6: workflow-guide + mqtt-edge playbook — accuracy vs current behavior (base64 config, actuator event provenance); add 'Troubleshooting' bullets"
    status: pending
  - id: ws7-exit-checklist
    content: "WS7: Exit checklist — sign-off table in README or here; link next phases phase_21_crop_cycle_analytics.plan.md then phase_24_rag_retrieval_system.plan.md"
    status: pending
isProject: false
---

# Phase 23 — Stabilization sprint

## Why now

Phases **20.x through 22** plus the **Pi ↔ API contract** landed a lot of surface area quickly. Automated tests pass, but **field hardware** and **operator stress** will find issues faster if we skip a deliberate stabilization pass. This phase is **intentionally boring**: tighten gates, fix drift, document reality, and burn down a small list of sharp edges **before** building **Phase 24 — RAG retrieval** on top of Phase **20.95 prep** schema.

## Principles

- **No feature work** unless it is a **P0/P1** correctness or security fix discovered during this sprint.
- Every change should be **small**, **reviewable**, and **test-backed** where possible.
- Prefer **documentation** over code when the code is already correct but operators would hit confusion.

## Work-stream summary

| WS | Focus |
|----|--------|
| **WS1** | CI / build gates (`go test`, `make audit-openapi`, `go vet`) |
| **WS2** | Smoke suite reliability + env documentation (`docs/local-operator-bootstrap.md` § API integration smoke tests; TestMain stderr + CI fail if DB missing) |
| **WS3** | OpenAPI ↔ `routes.go` parity (`openapi_route_diff.sh`) — green; Pi `apiKeyAuth` + `GET /farms/{id}/devices` dual auth documented; see `docs/local-operator-bootstrap.md` § OpenAPI route audit |
| **WS4** | Automation worker / program-tick — log patterns + `metadata.steps` monitoring documented in `docs/workflow-guide.md` (Programs); no code change (audit only) |
| **WS5** | Pi API key trust model + runbook notes |
| **WS6** | Operator docs aligned with shipped behavior |
| **WS7** | Exit checklist → hand off to Phase 21 then 24 |

## Exit criteria (all should be true)

1. `go test ./...` (or project-agreed subset) passes locally and in CI.
2. `make audit-openapi` exits **0** (no undocumented routes).
3. Full `go test ./cmd/api/...` smoke passes against a migrated dev DB.
4. No **undocumented** known P0/P1 issues the team agrees to fix in-sprint (either fixed or explicitly deferred with ticket link).
5. `docs/workflow-guide.md` and `docs/mqtt-edge-operator-playbook.md` reviewed for Pi + automation accuracy.
6. README [Roadmap Status](../README.md#roadmap-status) updated with Phase **23** row and pointer to this plan.

## After this sprint

Recommended order (keeps phase numbers meaningful):

1. **[Phase 21 — Crop cycle analytics](phase_21_crop_cycle_analytics.plan.md)** — reporting surface promised before RAG; richer metrics help retrieval later.
2. **[Phase 24 — RAG retrieval system](phase_24_rag_retrieval_system.plan.md)** — embeddings + API; only after 23 exit (and **after 21** unless product explicitly defers analytics).
