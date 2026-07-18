# gr33n roadmap

One page, read top to bottom, no phase-hopping required. gr33n ships in small
numbered **phases** (a phase = one focused change, with its own plan doc and
closure tests) — but you never need to read all 200+ of them to understand
where the project is or where it's going. That's what this page is for.

**Status:** every phase below is **shipped on `main`** unless marked
otherwise. Current phase: **204**.

For what the app actually *does* today, read the [README](../../README.md).
This page is history + direction, not features.

---

## Eras, in order

Each row is a stretch of phases that shipped one coherent piece of the
product. Follow the link only if you want implementation-level detail —
the description here is meant to be enough on its own.

| Era | Phases | What shipped |
|-----|--------|---------------|
| **Foundation** | 10–39 | Core CRUD (farms, zones, sensors, actuators), automation rules & schedules, fertigation programs, cost tracking, the first Pi edge client and offline queue. |
| **Farmer UX** | 40–67 | Whole-app polish pass for a real farmer (not a developer) — mobile layout, plain-language copy, the Pi Setup Wizard, push-to-talk field mode, and Guardian's first LLM-driven proposal cards (propose → Confirm, nothing silent). |
| **SPA workspaces** | 68–81 | Rebuilt the dashboard as single-page workspaces (Water, Light, Climate, Hardware, Money) instead of one long scrolling page. |
| **Crop intelligence** | 82–110 | Crop catalog moved into Postgres (~50 profiles with EC/DLI/photoperiod targets), domain enums served from the API instead of hardcoded in the UI, plant/cycle records bound to the catalog. |
| **Guardian model selection & hardening** | 111–118 | In-app Ollama model picker and pull workflow, security hardening (registration modes, login rate limits, JWT out of query strings), docs refresh. |
| **Virtual Pi wiring arc** | 119–123 | `/virtual-pi` — a graphical 40-pin board for wiring relays without touching a terminal, config export with drift detection, **Notify Pi to reload** for synced devices. |
| **Guardian eval & docs depth** | 122, 116–118 | `make guardian-eval` model-quality scoring, environment variable reference, upgrade/backup runbooks, API quickstart. |
| **Guardian polish & trust infra** | 124–162 | Citation deep links, accessibility pass (skip links, zone tab traps), Guardian Confirm→DB smoke coverage, answer-relevance guard against topic drift. |
| **Today — visual farm cockpit** | 163–177 | Replaced the AI-launcher homepage with **Today**: a spatial farm map (drag zone tiles over a background photo), attention strip, large-farm filters, farm pulse, and first-run coach marks. Guardian demoted to a single ask row instead of the front door. |
| **Online weather** | 178 | Optional live forecast (Open-Meteo, no API key) once a farm sets site coordinates. |
| **2026-07 sit-in arc** | 179–187 | Guardian UX pass, Help Library knowledge surfaces, multi-turn PR conversation smoke tests, and the task **Refine** chain (correct title/zone/due-date in the same session instead of starting over). |
| **Answer-quality audit** | 188–191 | Found and fixed real conversation bugs: off-topic template leak, RAG metadata leaking into chat replies, truncated list intros, and revise requests phrased as questions not being understood. |
| **Post-audit follow-through** | 192–201 | Fixed a due-date/title clobber bug, Help Library sticky-header overlaps, pending-proposal inbox and conversation view, session sidebar labels, accuracy-note persistence across reloads, and unified the knowledge/help surfaces. |
| **Janitorial consolidation** | 202–203 | No new features — paid down test and code duplication. Consolidated repeated UI closure-test assertions into canonical test files, merged duplicated backend helper functions and adjacent handler packages. |
| **Docs & navigation cleanup** | **204** (current) | This page. Product-first README, retired duplicate phase-60 doc pile, moved closed early-era plans into `docs/plans/archive/`. |

---

## What's next

Not phase-gated yet — see the documented backlog:
[`docs/plans/product_backlog_operator_runtime.plan.md`](../plans/product_backlog_operator_runtime.plan.md).

## Where the detail actually lives

You should rarely need these, but if you do:

- **[`docs/phase-14-operator-documentation.md`](../phase-14-operator-documentation.md)** — the exhaustive phase-by-phase index (every phase, every closure test, every migration).
- **`docs/plans/phase_N_*.plan.md`** — one plan per phase: problem statement, workstreams, acceptance criteria.
- **[`docs/plans/archive/`](../plans/archive/)** — plans for phases old enough that nothing else in the repo (code, tests, docs) still points at them directly.
- **[`CHANGELOG.md`](../../CHANGELOG.md)** — operator-visible changes only, newest first.
