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
    content: "WS1: Tutorial + glossary — v1 shipped (/operator-guide, nav, dashboard hints); follow-up: overlay tour + more empty states"
    status: completed
  - id: ws2-obs-log-management
    content: "WS2: Operational logs — runbook + json-file rotation + Promtail/Loki/Grafana compose overlay; optional K8s notes later"
    status: completed
  - id: ws3-rag-education-boundary
    content: "WS3: RAG vs education vs ops logs — v1 in rag-scope §9 + workflow §10.6 + Knowledge HelpTip; follow-up help-library / incident domain"
    status: completed
isProject: false
---

# Phase 26 — Operator tutorial, observability evolution, RAG education layer

## Status

**Phase 26 — WS1–WS3 v1 documented/shipped in-repo** (Guide, logging runbook + Loki overlay, RAG vs education vs logs in **`rag-scope-and-threat-model.md` §9**). Remaining: WS1/WS2 polish per follow-ups; Phase **27** when ready.

## WS1 — Tutorial + glossary (partial)

**Shipped (v1):**

- **`/operator-guide`** — Operator Guide: glossary entries aligned with **`workflow-guide.md`** §11 concepts (schedule vs rule, setpoint vs reading, RAG blurb), suggested **`router-link`** walk matching **`operator-tour.md`** §2.
- **Navigation** — **System → Guide** in sidebar + mobile drawer (`SideNav.vue`, `App.vue`).
- **Dashboard** — HelpTip links to Guide; quick action **Operator guide**; short “why empty” lines on tasks, alerts, schedules, fertigation, sensors, zones, devices widgets.

**Follow-up (still WS1 scope):** richer guided tour (overlay / checklist persistence), empty-state copy on **Schedules**, **Sensors**, **Automation**, etc.; optional deep-link from HelpTips on those pages.

## WS2 — Operational logs (partial)

**Shipped (v1):**

- **[`operator-logging-runbook.md`](../operator-logging-runbook.md)** — **`slog`** baseline recap (**`LOG_FORMAT`**, **`AUTH_DEBUG_LOG`**, request + automation + RAG pointers); **application logs vs Timescale / DB retention**; Docker Compose **json-file** rotation; systemd **journald**; optional **Loki / agents** sketch; **archival** (`docker logs`, `journalctl` export); correlation checklist.
- **`docker-compose.yml`** — **`logging`** (`json-file`, `max-size` / `max-file`) on **`db`**, **`api`**, **`ui`**.
- **`docker-compose.logging.yml`** — merge overlay: **Loki** + **Promtail** + **Grafana**; sets **`LOG_FORMAT=json`** on **`api`**; configs under **`logging/`**.
- **`make compose-logging-up`** / **`compose-logging-down`** — convenience wrappers.
- Cross-links from **[`INSTALL.md`](../../INSTALL.md)** (observability table) and **[`operator-troubleshooting.md`](../operator-troubleshooting.md)**.

**Follow-up:** Kubernetes-specific notes only if we ship K8s manifests later.

## WS3 — RAG education boundary (partial)

**Shipped (v1):**

- **[`rag-scope-and-threat-model.md`](../rag-scope-and-threat-model.md) §9** — **Static education** (Guide, tour, glossary) vs **farm-grounded RAG** (checklist-approved DB domains only) vs **operational / HTTP logs** (out of scope for `rag-ingest` by default); future summarized-incidents criteria.
- **[`workflow-guide.md`](../workflow-guide.md)** §10.6 — short pointer to §9.
- **Knowledge UI** — HelpTip clarifies DB-backed scope + **`rag-scope` §9** path.

**Follow-up:** Product decision if **help-library** chunks ever get embedded; any **curated incident** domain for Knowledge.

## Goals

1. **Tutorial system** — Guided paths in-product (or linked docs) so operators learn zones → sensors/controls → schedules/rules → tasks → fertigation without guessing. Align with **[operator-tour.md](../operator-tour.md)** and expand **“why empty?”** hints (sit-in §1).
2. **Glossary / terminology** — Stable definitions (e.g. **setpoint** vs live reading, **schedule** vs **rule**) so copy and RAG answers stay consistent.
3. **Observability beyond stdout** — Today the API emits structured logs via Go **`log/slog`** to **process stdout** (similar in spirit to structured logging in .NET ecosystems—fields you can parse, optional JSON). Phase 26 tracks **operator-facing** concerns: shipping logs to a **collector**, **retention policies**, **archival** (cold storage, compliance), and runbooks—without confusing them with **database** retention.

## Logs vs database retention

- **Application logs** (`request`, `auth_rejected`, `automation schedule run`, …) are **not** stored in Postgres by default; they follow whatever captures **stdout** (containers, systemd, a log stack).
- **Time-series / hypertable** policies (e.g. trimming old **sensor readings** or event rows to cap DB size) are **separate**: they delete or aggregate **table rows**, not your centralized log archive. If you need a **long audit trail** of operational behavior when DB rows roll off, that belongs in **log retention / archival** (Phase 26 WS2)—not in RAG by default.

## RAG and logs (deliberate boundary)

**Authoritative (Phase 26 WS3 v1):** **[`rag-scope-and-threat-model.md`](../rag-scope-and-threat-model.md) §9** — static education vs DB RAG vs stdout/Loki logs; raw HTTP lines stay **out** of ingestion unless a future product pass adds allowlists + redaction.

## Preconditions

- Sit-in §1–§2 artifacts: [operator-tour.md](../operator-tour.md), [operator-troubleshooting.md](../operator-troubleshooting.md), `cmd/api/request_log.go`, `AUTH_DEBUG_LOG`, automation **`slog`** lines.
- Phase 25 RAG pipeline stable enough that Phase 26 UX does not fight empty Knowledge surfaces.

## References

- [Sit-in workstream](../workstreams/sit-in-operator-experience.md)
- [Operator logging runbook](../operator-logging-runbook.md) — Phase 26 WS2
- [RAG scope and threat model](../rag-scope-and-threat-model.md)
- [Phase 25 — RAG operations and expansion](phase_25_rag_operations_and_expansion.plan.md)

---

*Created as a planning stub; split or rename workstreams when implementation starts.*
