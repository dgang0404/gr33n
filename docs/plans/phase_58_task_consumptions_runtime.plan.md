---
name: Phase 58 — Task consumptions & operator runtime
overview: >
  Make stock drawdown visible when tasks complete — consumptions API exists but
  farmer UI is thin. Plus operator-runtime backlog: offline queue hints, task
  templates for refill/mix, and dashboard "what needs doing" coherence.
todos:
  - id: ws1-consumption-ui
    content: "WS1: Task complete sheet — optional batch + qty drawdown; Supplies history line"
    status: completed
  - id: ws2-refill-templates
    content: "WS2: Task templates — Refill input, Log mix, Check sensor (from low-stock / alerts)"
    status: completed
  - id: ws3-runtime-hints
    content: "WS3: Offline actuator queue copy; task due chips on zone strip"
    status: completed
  - id: ws4-docs-tests
    content: "WS4: operator-tour § consumptions; phase-58-closure; OC-58"
    status: completed
isProject: false
---

# Phase 58 — Task consumptions & operator runtime

## Status

**Shipped.** Task complete consumption sheet, templates, zone/dashboard runtime hints, Supplies batch footnotes, `GET /farms/{id}/task-consumptions`.

---

## The one job

> **Finish a task and stock updates — or see why it didn't — without opening Advanced.**

---

## WS1 — Consumption UI

| Surface | Behavior |
|---------|----------|
| Task complete dialog | Optional: pick NF batch, qty, unit → POST consumption |
| Task detail | List consumptions linked to task |
| Supplies batch card | "Used by tasks" footnote with links |
| Validation | Block qty > on-hand with plain message |

Reuse existing store actions if present; else add `recordTaskConsumption`.

---

## WS2 — Task templates

| Trigger | Template |
|---------|----------|
| Low-stock banner (53) | Refill {input} — pre-fill description |
| Alert: sensor offline | Check {sensor} wiring |
| Feed schedule miss | Review feeding plan |

`POST /tasks` with `template_id` or metadata blob.

---

## WS3 — Runtime hints

- Zone strip: overdue task chip → `/tasks?zone=`
- Actuator offline: "Commands queue when back" (if true) or link Pi setup
- Dashboard: merge open tasks + low-stock into one "Do next" strip (read-only aggregate)

---

## WS4 — Docs, tests, OC-58

- operator-tour consumptions paragraph
- Vitest: complete task with consumption mocks API
- Guardian starter: "What did we use on last mix task?" (read — Phase 55 backlog if no tool)

---

## Definition of done

- [x] Complete refill task reduces batch qty in UI
- [x] Templates create tasks from low-stock CTA
- [x] OC-58 closed
