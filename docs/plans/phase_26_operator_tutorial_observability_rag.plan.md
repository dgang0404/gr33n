---
name: Phase 26 Operator tutorial, observability evolution, RAG education layer
overview: >
  Scope the next product slice after sit-in foundations: in-app tutorial and glossary,
  optional consolidation of operational logging (slog/stdout → aggregation/archival),
  and a deliberate education layer that combines static help with farm-grounded RAG.
  Distinct from Phase 25 “RAG ops”—here the focus is operator UX and long-lived narrative,
  not embedding breadth.
todos:
  - id: ws1-tutorial-glossary
    content: "WS1: Tutorial flows + glossary (setpoints, schedules vs rules, empty states)—wire from HelpTip / operator tour"
    status: pending
  - id: ws2-obs-log-management
    content: "WS2: Operational log strategy—aggregation (Docker/journal/Loki), retention vs Timescale row pruning, archival export; document slog baseline"
    status: pending
  - id: ws3-rag-education-boundary
    content: "WS3: Define what may enter RAG (domain DB vs ops logs)—sanitization, farm scope, consent if summarizing HTTP/automation traces"
    status: pending
isProject: false
---

# Phase 26 — Operator tutorial, observability evolution, RAG education layer

## Status

**Planning stub** — refine scope before execution. Complements **[sit-in operator experience](../workstreams/sit-in-operator-experience.md)** (tour, troubleshooting, structured **`slog`** logging already landed in §1–§2).

## Goals

1. **Tutorial system** — Guided paths in-product (or linked docs) so operators learn zones → sensors/controls → schedules/rules → tasks → fertigation without guessing. Align with **[operator-tour.md](../operator-tour.md)** and expand **“why empty?”** hints (sit-in §1).
2. **Glossary / terminology** — Stable definitions (e.g. **setpoint** vs live reading, **schedule** vs **rule**) so copy and RAG answers stay consistent.
3. **Observability beyond stdout** — Today the API emits structured logs via Go **`log/slog`** to **process stdout** (similar in spirit to structured logging in .NET ecosystems—fields you can parse, optional JSON). Phase 26 tracks **operator-facing** concerns: shipping logs to a **collector**, **retention policies**, **archival** (cold storage, compliance), and runbooks—without confusing them with **database** retention.

## Logs vs database retention

- **Application logs** (`request`, `auth_rejected`, `automation schedule run`, …) are **not** stored in Postgres by default; they follow whatever captures **stdout** (containers, systemd, a log stack).
- **Time-series / hypertable** policies (e.g. trimming old **sensor readings** or event rows to cap DB size) are **separate**: they delete or aggregate **table rows**, not your centralized log archive. If you need a **long audit trail** of operational behavior when DB rows roll off, that belongs in **log retention / archival** (Phase 26 WS2)—not in RAG by default.

## RAG and logs (deliberate boundary)

- **Current RAG** ingests **farm-scoped domain text** from approved tables (tasks, cycles, automation narrative fields, etc.) per **[rag-scope-and-threat-model.md](../rag-scope-and-threat-model.md)**.
- **Raw HTTP access logs** are usually a **poor default** for RAG: paths, ids, and errors can touch **PII** or security-sensitive patterns; they are also noisy versus curated domain copy.
- **If** the product later wants “what broke last week?” style answers in **Knowledge**, treat that as **WS3**: explicit **allowlist**, **redaction**, **farm** or **deployment** scope, and possibly **summarized** operational events—not verbatim request logs unless audited for privacy.

Static **tutorial** copy + **glossary** remain the right default for teaching; **RAG** stays strongest on **what your farm’s data says**, with optional synthesis—same split as Phase 25 docs.

## Preconditions

- Sit-in §1–§2 artifacts: [operator-tour.md](../operator-tour.md), [operator-troubleshooting.md](../operator-troubleshooting.md), `cmd/api/request_log.go`, `AUTH_DEBUG_LOG`, automation **`slog`** lines.
- Phase 25 RAG pipeline stable enough that Phase 26 UX does not fight empty Knowledge surfaces.

## References

- [Sit-in workstream](../workstreams/sit-in-operator-experience.md)
- [RAG scope and threat model](../rag-scope-and-threat-model.md)
- [Phase 25 — RAG operations and expansion](phase_25_rag_operations_and_expansion.plan.md)

---

*Created as a planning stub; split or rename workstreams when implementation starts.*
