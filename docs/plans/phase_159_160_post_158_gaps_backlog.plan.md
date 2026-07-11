---
name: Phases 159–160 — post-158 gaps arc
overview: >
  Honest follow-up after the 154–158 infra & trust arc. Two cohesive phases
  remain: Guardian citation completeness + turn persistence (159), and a11y
  residuals + advisory CI guard (160). Doc drift in current-state.md ships
  with 159.
todos:
  - id: phase-159-guardian-citations
    content: "Phase 159 — Guardian citation completeness + accuracy_note persistence"
    status: completed
  - id: phase-160-a11y-residuals
    content: "Phase 160 — a11y residuals + advisory axe smoke"
    status: completed
isProject: false
---

# Phases 159–160 — post-158 gaps arc

**Origin:** operator pushed 154–158 and asked what gaps remain.

**Verdict:** No new “whole-app missing piece” like backups or vuln scanning. What’s left is **targeted follow-through** on shipped arcs (152 citations, 158 a11y) plus **doc drift** in `current-state.md`.

| Phase | Status | Plan | Priority |
|-------|--------|------|----------|
| **159** | ✅ Shipped | [`phase_159_guardian_citation_completeness.plan.md`](phase_159_guardian_citation_completeness.plan.md) | **P0** — completes Guardian trust loop |
| **160** | ✅ Shipped | [`phase_160_a11y_residuals.plan.md`](phase_160_a11y_residuals.plan.md) | **P1** — field UX polish |

**Suggested build order:** 159 → 160.

---

## What’s closed (don’t re-open)

| Area | Status |
|------|--------|
| **154–158** infra & trust | Shipped (test-unit, backup, vuln-check, docs, a11y core path) |
| **115** schema utilization | Shipped — `current-state.md` wrongly still lists it as planned |
| **150** dev-jargon hygiene | Shipped |
| **153** change-request smoke | Shipped (`make guardian-qa-change-requests`) |
| **Product backlog B1–B4** | Shipped |
| **Pre-dev gaps index** | Archived — historical only |

---

## Remaining gaps (honest)

### P0 — Guardian trust residuals (Phase 159)

1. **Phase 152 WS2b** — citation chips still plain text for `schedule`, `alert_notification`, `field_guide`, `platform_doc` ([plan](phase_152_guardian_live_accuracy_guardrails.plan.md) § WS2b).
2. **`accuracy_note` not persisted** — accuracy banners disappear on session reload ([plan](phase_152_guardian_live_accuracy_guardrails.plan.md) § WS1c limitation).
3. **`current-state.md` stale** — still says 157/158 planned and 158 “not started”.

### P1 — A11y residuals (Phase 160)

From [`a11y-audit-2026-07-11.md`](../a11y-audit-2026-07-11.md) § Deferred:

- `ZoneLightingEditor` modal — no `role="dialog"` / focus trap
- `LightingProgramForm` — many fields lack `for`/`id`
- Mobile hamburger drawer — no focus trap (Guardian drawer has one)
- Advisory axe / eslint a11y in CI (non-blocking)

### P2 — Ongoing / not one phase

| Item | Why defer |
|------|-----------|
| **ec-ph smoke drift** | Recurring Guardian *quality* issue — tune detectors/prompts in a 143–147-style pass, not infra |
| **LLM-as-judge in CI** | Optional GPU lane; Phase 146 marked advisory |
| **Full `make test` without DB** | Phase 154 scope — `cmd/api` smokes need migrated Postgres by design |
| **Insert Commons federation** | Opt-in; not required for single-farm LAN |
| **Sidebar roving tabindex** | Nice-to-have; no operator request yet |
| **WCAG certification** | Explicit non-goal for 158 |

---

## Prompt Composer with

```
phase 159
```

or `phase 159 ws1` for a single workstream. Same pattern for 160.
