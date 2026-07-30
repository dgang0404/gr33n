---
name: Phase 211.07 — Guardian Pending tab Playwright E2E + Stagehand spike
overview: >
  After 211.06 closes the laptop counsel loop, expand e2e/ so Confirm / Dismiss
  on Guardian pending proposals is covered by deterministic Playwright, then
  spike Stagehand on one flaky flow — without replacing Vitest or the existing
  e2e suite.
todos:
  - id: ws0-gate
    content: "WS0: Gate on 211.06 debug exit (core ≥4/5, NF1 not all invent/timeout) + list Pending UI selectors / data-test hooks"
    status: pending
  - id: ws1-seed-path
    content: "WS1: Deterministic pending proposal seed for E2E — API leave-pending or fixture POST; no live phi3 marathon"
    status: pending
  - id: ws2-confirm-dismiss
    content: "WS2: Playwright specs — open /chat?tab=pending, Confirm one proposal, Dismiss another; assert queue + toast/side-effect"
    status: pending
  - id: ws3-guardian-pending-shell
    content: "WS3: Extend guardian-chat.spec — Pending tab visible, empty state, card chrome (still no live LLM required for shell)"
    status: pending
  - id: ws4-stagehand-spike
    content: "WS4: Optional spike — Stagehand on one flaky flow only; document keep/kill; do not make Stagehand the default suite"
    status: pending
  - id: ws5-closure
    content: "WS5: e2e README + make e2e-browser green locally; optional browser-e2e CI note; mark Complete"
    status: pending
isProject: false
---

# Phase 211.07 — Guardian Pending tab Playwright E2E + Stagehand spike

**Status:** Planned · **Depends on:** [211.06 Guardian smoke reliability](phase_211_06_guardian_smoke_reliability.plan.md) closed (debug loop trustworthy) · **Before:** [212 dual-install federation](phase_212_dual_farm_federation_test.plan.md)

## The one job

> Prove in a **browser** that an operator can **Confirm** and **Dismiss** Guardian change-request cards on the Pending tab — without a 4-hour phi3 marathon — and only then try AI-assisted UI automation on one brittle flow.

## Why now

211.06 hardens counsel **eval** (`guardian-eval` / smoke archives). Write-intent fixtures already verify proposals land in the pending queue via API, but Phase 184 deliberately left the **browser Confirm/Dismiss** path as a manual UI step. You already have Playwright (`e2e/`, `make e2e-browser`) for login, tasks, and Guardian **shell** — not for Pending cards.

UI testing after counsel reliability: stop treating Confirm as “trust the API smoke + human click.”

## Non-negotiable framing

- **Deterministic first.** Playwright with stable `data-test` hooks and a seeded pending proposal. Live LLM optional and out of the default journey.
- **Do not replace Vitest.** Component/closure tests stay; E2E is the journey layer.
- **Stagehand is a spike, not a rewrite.** One flaky flow; keep/kill written down. No SaaS lock-in required for the phase to ship.
- **No Gitee / Chinese forum work in this phase** — deferred until after 212 / blinking-light loose ends.

## Scope

- `e2e/*.spec.js`, `e2e/README.md`, Playwright config as needed
- Minimal UI `data-test` attributes on Pending / proposal cards if missing
- Seed path: reuse `guardian-qa-change-requests-pending-quick` output **or** a small authenticated API helper in e2e that creates a pending proposal without Ollama
- Optional: Stagehand (or Midscene) prototype behind an env flag / separate npm script

## Out of scope

- Full multi-turn `change-requests-ui` parity in Playwright (API suite already covers dialogues)
- Replacing `make guardian-qa-change-requests*` 
- Mandatory Stagehand in CI
- Phase 212 federation
- Expanding AI UI tools beyond a single documented spike

## Workstreams

### WS0 — Gate + selector audit

1. Confirm 211.06 operator exit: `make guardian-qa-debug` meets core ≥4/5 and NF1 not all invent/timeout (or document accepted residual).
2. Inventory Pending UI: `GuardianActionProposal.vue`, `/chat?tab=pending`, Confirm / Dismiss / Refine controls — note existing `data-test` gaps.

### WS1 — Seed without a marathon

Prefer one of:

- **A (ponytail):** E2E setup calls API with eval JWT to insert or leave a short-TTL pending proposal (mirror leave-pending), **or**
- **B:** Document “run `make guardian-qa-change-requests-pending-quick` once, then Playwright” as a manual precondition for the Confirm journey only.

Default to **A** if a stable create-proposal test endpoint or admin path already exists; otherwise B with a clear README.

### WS2 — Confirm + Dismiss journeys

Playwright specs (names flexible):

1. Login → `/chat?tab=pending` → see ≥1 proposal card  
2. **Confirm** → card leaves pending; optional side-effect assertion (task/ack/schedule already covered by API Confirm smoke — UI can assert “gone from pending” + success chrome)  
3. **Dismiss** (separate card or re-seed) → removed without applying  

Keep timeouts laptop-friendly; no chat completion wait.

### WS3 — Shell hardening

Extend `guardian-chat.spec.js` (or sibling): Pending tab switch, empty-state copy, model selector still present. Still **no** live LLM.

### WS4 — Stagehand spike (optional, after WS2 green)

1. Add optional dep / script (e.g. `npm run test:stagehand` in `e2e/`) behind clear docs.  
2. Pick **one** flaky or selector-churny flow (candidate: Pending tab discovery after layout tweaks).  
3. Write keep/kill note in `e2e/README.md`: cost, flakiness, whether it earns a permanent home.  
4. Default CI remains plain Playwright.

### WS5 — Closure

- `make e2e-browser` green on `dev-auth-test`  
- README journeys list updated  
- Cross-link from [ci-guardian-qa.md](../ci-guardian-qa.md) or operator bootstrap if Pending UI check is part of smoke checklist  
- Mark this plan **Complete**

## Acceptance criteria

- [ ] At least one Playwright journey Confirms a pending proposal from the UI  
- [ ] At least one Playwright journey Dismisses a pending proposal from the UI  
- [ ] Default e2e path does not require a live local LLM  
- [ ] Stagehand either documented as spike-only with keep/kill, or explicitly deferred in README with reason  
- [ ] Phase 206 allowlist includes this plan file  

## Suggested order

1. Finish / grade 211.06 debug re-run  
2. WS0 → WS1 → WS2 → WS3  
3. WS4 only if Confirm/Dismiss is green and a real flaky flow exists  
4. WS5  

## Related

- [211.06 Guardian smoke reliability](phase_211_06_guardian_smoke_reliability.plan.md) — counsel loop gate  
- [e2e/README.md](../../e2e/README.md) — existing Playwright journeys  
- [ci-guardian-qa.md](../ci-guardian-qa.md) — change-requests / leave-pending make targets  
- Phase 184 (archive) — multi-turn pending prep; browser Confirm left manual on purpose then  
- [212 dual-install federation](phase_212_dual_farm_federation_test.plan.md) — still after this QA arc  
