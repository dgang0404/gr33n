---
name: Phases 154–158 — infra & trust gaps arc
overview: >
  Gap-analysis pass (post-153): operator asked for an honest look at what the
  app needs beyond Guardian QA. All five phases now have dedicated plan docs.
  Phase 154 is shipped; 155–158 are written up and ready for Composer to build.
todos:
  - id: verdict-pr-ci-gate
    content: "Verdict on the reverted GitHub PR CI gate — see verdict section below"
    status: completed
  - id: phase-154-test-suite-health
    content: "Phase 154 — test suite health + make test-unit"
    status: completed
  - id: phase-155-automated-backups
    content: "Phase 155 — automated backups"
    status: completed
  - id: phase-156-dependency-scanning
    content: "Phase 156 — dependency/vuln scanning"
    status: completed
  - id: phase-157-docs-consolidation
    content: "Phase 157 — docs consolidation (plan doc ready)"
    status: pending
  - id: phase-158-accessibility-pass
    content: "Phase 158 — accessibility pass (plan doc ready)"
    status: pending
isProject: false
---

# Phases 154–158 — infra & trust gaps arc

**Origin:** operator asked, deliberately open-ended, "is the app missing something, push back if I'm wrong — make phase docs, don't code yet." All five items are now **fully written up** as individual phase plans. Phase 154 was also implemented when you said "yes start"; 155–158 are **plan-only** until you point Composer at them.

| Phase | Status | Plan |
|-------|--------|------|
| **154** | ✅ Shipped | [`phase_154_test_suite_health.plan.md`](phase_154_test_suite_health.plan.md) |
| **155** | ✅ Shipped | [`phase_155_automated_backups.plan.md`](phase_155_automated_backups.plan.md) |
| **156** | ✅ Shipped | [`phase_156_dependency_scanning.plan.md`](phase_156_dependency_scanning.plan.md) |
| **157** | 📋 Planned | [`phase_157_docs_consolidation.plan.md`](phase_157_docs_consolidation.plan.md) |
| **158** | 📋 Planned | [`phase_158_accessibility_pass.plan.md`](phase_158_accessibility_pass.plan.md) |

**Suggested build order:** 155 → 156 → 157 → 158 (154 already done).

---

## Verdict: was the reverted GitHub PR CI gate "bad"?

No. A label-gated, non-blocking CI job for slow/model-dependent smoke tests is **standard practice**. It was reverted because you hadn't asked for it that day — a consent/scope issue, not a quality issue. Phase 156 proposes dependency scanning **explicitly** with clear blocking vs advisory lanes; Guardian slow smokes stay separate (`make guardian-qa-smoke-strict`, not default PR CI).

---

## Phase summaries (detail in each plan doc)

### 154 — Test suite health ✅

`make test-unit` green without Postgres; fixed compile failures and stale unit tests. Full `make test` still needs migrated DB for `cmd/api` smokes.

### 155 — Automated backups

`scripts/backup-gr33n.sh`, retention, scratch-DB verify, `make backup` / `make verify-backup`. Highest-consequence gap for self-hosted operators.

### 156 — Dependency & vuln scanning

Dependabot + `govulncheck` + `npm audit`. SECURITY.md triage process. No surprise CI scope creep.

### 157 — Docs consolidation

`docs/current-state.md` snapshot + `docs/plans/archive/` for closed phases + trimmed phase-14 index.

### 158 — Accessibility pass

Keyboard nav, screen-reader labels, axe audit on Guardian chat + core workspaces. Lower priority than 155–156.

---

## What's NOT a gap (checked)

- **Hosted-LLM lock-in** — `LLM_BASE_URL` is OpenAI-compatible
- **Offline/PWA** — shipped ([`connectivity-requirements.md`](../connectivity-requirements.md))
- **RBAC** — shipped
- **Guardian change-request smoke** — Phase 153 (`make guardian-qa-change-requests`)

---

## Prompt Composer with

```
phase 155 ws1
```

or `phase 155` for the full phase. Same pattern for 156–158.
