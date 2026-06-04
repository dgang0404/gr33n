---
name: Product backlog — operator runtime (non-phase)
overview: >
  Smaller product items that do not warrant full phases but must be tracked before
  calling the grow stack "complete." Indexed as Tier B in pre_development_gaps_index.
  Implement opportunistically after Phase 39 WS1 or alongside Phase 41.
todos:
  - id: run-now-api
    content: "POST /farms/{id}/fertigation/programs/{id}/run-now — ad-hoc program tick with idempotency; UI on program row + zone Water"
    status: completed
  - id: metadata-steps-deprecation
    content: "Deprecate programs.metadata.steps — monitor fallback warnings; then harden action_source and drop column"
    status: completed
  - id: guardian-lighting-propose
    content: "Guardian propose tool create_lighting_program — mirror summarize_zone_lighting; Confirm-gated"
    status: completed
  - id: mobile-distribution
    content: "Capacitor / store checklist — docs/mobile-distribution.md execution pass (icons, signing, release notes)"
    status: completed
isProject: false
---

# Product backlog — operator runtime

**Indexed in:** [`pre_development_gaps_index.plan.md`](pre_development_gaps_index.plan.md) (Tier **B1–B4**).

These are **not** blockers for starting Phase 39. They improve day-2 operations and release hygiene.

---

## B1 — Program "run now"

**Problem:** Operators must wait for cron or manually pulse actuators; no explicit "run this program now."

**Proposed:**

- `POST /farms/{farm_id}/fertigation/programs/{program_id}/run-now` (name TBD)
- Reuses program tick / queue enqueue path from Phase 39 WS5
- RBAC: operator+; audit log entry
- UI: program list + zone Water "Run now" when program linked

**Depends on:** Phase 39 WS1 (queue), ideally WS5 pipeline.

**Acceptance:** Run-now enqueues same commands as scheduled tick for demo program; second call within idempotency window does not duplicate.

---

## B2 — Deprecate `programs.metadata.steps`

**Problem:** Legacy `metadata.steps` coexists with `action_source`; worker emits fallback warnings.

**Proposed:**

1. Metric/log: count programs still hitting steps fallback.
2. After N releases with zero warnings: migration drops column or stops reading.
3. Promote `action_source`-only validation to hard errors in program tick.

**Depends on:** None for logging; migration is ops-coordinated.

**Acceptance:** README checklist item closed; no fallback warnings in CI smokes.

---

## B3 — Guardian `create_lighting_program` propose

**Problem:** Phase 35 shipped read `summarize_zone_lighting` but no propose tool for setup from chat.

**Proposed:**

- Add to Guardian tool catalog (Phase 30 pattern): propose lighting program from preset + zone + actuator
- Confirm card with photoperiod summary
- OpenAPI + persona doc update

**Depends on:** Phase 35 APIs (exist).

**Acceptance:** Confirm creates program visible on `/lighting` and zone Light tab.

---

## B4 — Mobile distribution polish

**Problem:** PWA works; store-distributed Capacitor path needs a repeatable checklist.

**Canonical doc:** [`docs/mobile-distribution.md`](../mobile-distribution.md)

**Tasks (checklist, not code phase):**

- [ ] Icons / splash assets per platform — procedure in [`mobile-distribution.md`](../mobile-distribution.md#release-checklist-b4--operator-runtime-backlog)
- [ ] Signing + provisioning profiles documented — same section
- [x] Release notes template — same section
- [ ] Deep link smoke (optional)
- [x] Align with Phase 18 mobile hardening notes — linked in checklist

**Acceptance:** Operator can follow doc end-to-end for one TestFlight/internal track build.

---

## Related

| Doc | Use |
|-----|-----|
| [phase_39_edge_fertigation_execution.plan.md](phase_39_edge_fertigation_execution.plan.md) | Queue for run-now |
| [phase_35_lighting_domain.plan.md](phase_35_lighting_domain.plan.md) | Lighting propose |
| [README.md](../../README.md) | In-flight list |
