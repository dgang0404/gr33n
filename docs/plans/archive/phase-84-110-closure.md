# Phases 84–110 — master roadmap closure

**Status:** **Shipped** on `main` (all arcs A–E formally closed).

**Canonical plan:** [`phase_84_100_master_roadmap.plan.md`](phase_84_100_master_roadmap.plan.md)

**All formal roadmaps shipped on `main`.** SPA arc 68–81 closed; crop/intelligence arc 84–110 closed. For new gaps found in testing, use **Phase 111+** — see [`phase_84_100_master_roadmap.plan.md`](phase_84_100_master_roadmap.plan.md#adding-phase-111).

---

## The one job (done)

> **Every blind spot in the 84–110 map has a phase, implementation on main, and a closure doc** — plants-first order honored; no ad-hoc gaps left untracked.

---

## Arc checklist

| Arc | Phases | Closure hub |
|-----|--------|-------------|
| **A** — Plants & crop knowledge | 84–87, 93 | [`phase_84_87_crop_identity_roadmap.plan.md`](phase_84_87_crop_identity_roadmap.plan.md) · [`phase-87-closure.md`](phase-87-closure.md) |
| **B** — UI static → DB/API | 88–92, 99 | [`phase_88_92_platform_data_gaps_roadmap.plan.md`](phase_88_92_platform_data_gaps_roadmap.plan.md) |
| **C** — Blind spots & enterprise | 93–100 | Individual `phase-NN-closure.md` (93–100) |
| **D** — Guardian, programs, analytics | 101–105 | Individual `phase-NN-closure.md` (101–105) |
| **E** — Intelligence & polish | 106–110 | Individual `phase-NN-closure.md` (106–110) · OC-82 via [`phase-110-closure.md`](phase-110-closure.md) |

---

## Blind spot map — all addressed

| # | Blind spot | Phase(s) |
|---|------------|----------|
| 1 | Identity vs label fuzzy | 85, 93 |
| 2 | Guardian alias vs picker diverge | 86, 87 |
| 3 | Farm override vs genetics EC | 87, 94 |
| 4 | Catalog growth cadence | 95 |
| 5 | Picker 404 fallback | 85, 100 |
| 6 | `strain_or_variety` / strains tab | 93 |
| 7 | Feeding program ↔ stage mismatch | 96, 102 |
| 8 | RAG vs structured targets | 97 |
| 9 | Multi-farm / commons promotion | 98 |
| 10 | CI enum drift | 88, 99 |
| 11 | Mobile / offline picker | 100 |
| 12 | Execution order risk | This master doc |

---

## Per-phase closure index

| Phase | Closure |
|-------|---------|
| 84 | [`phase-84-closure.md`](phase-84-closure.md) |
| 87 | [`phase-87-closure.md`](phase-87-closure.md) |
| 88–92 | [`phase-88-closure.md`](phase-88-closure.md) … [`phase-92-closure.md`](phase-92-closure.md) |
| 93–110 | [`phase-93-closure.md`](phase-93-closure.md) … [`phase-110-closure.md`](phase-110-closure.md) |

Phases 85–86 closure artifacts live in their plan docs + smokes; 101–105 each have dedicated closure docs.

---

## Adding Phase 111+

When testing finds a new gap:

1. Add row to blind spot table in [`phase_84_100_master_roadmap.plan.md`](phase_84_100_master_roadmap.plan.md)
2. Create `docs/plans/phase_NNN_<slug>.plan.md`
3. Ship + `phase-NNN-closure.md`
4. One line in [`phase-14-operator-documentation.md`](../phase-14-operator-documentation.md)

---

## OC-84–110

The **84–110 master roadmap is closed** when all five arcs are shipped, every phase in the locked order has a closure artifact, and phase-14 indexes the arc as complete.
