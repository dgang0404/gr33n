# Sit-in workstream: operator experience, observability, tasks-first

**Sit-in** means this backlog **stays named as-is** even if calendar phases (e.g. Phase 25 RAG) advance. New work that does not belong here should show up as **scope creep** if it lands in this file — keep product phases and this stream separate.

**Goal:** Make the live product **understandable** (docs + UI cues), **debuggable** (logging), and **usable day-to-day** with **tasks** as the spine — before leaning harder into RAG or net-new features.

---

## 1. Documentation / onboarding

| Item | Notes |
|------|--------|
| **Single-page operator tour** | Under `docs/` — narrative walk: Farm → Zones → Sensors/Controls → Schedules/Rules → Tasks → Fertigation. Include **one data-flow diagram** (mermaid or ASCII). |
| **“Why empty?” UX** | Per major UI area, future inline hints (telemetry vs setpoints vs automation inactive). Track implementation as **separate UX tickets**; link from tour. |

**Suggested artifact:** `docs/operator-tour.md` (name flexible).

---

## 2. Logging / observability (“logging phase” — can align with Phase 26 docs)

| Item | Notes |
|------|--------|
| **API structured logs** | Request correlation id; farm_id / user_id where applicable; route; status. |
| **Auth debug** | Optional flag for auth failure **reasons** without printing secrets or full JWTs. |
| **Automation worker** | Schedule tick: **info** for outcomes; **warn** on failure with **schedule_id / rule_id**. |
| **Runbook doc** | “Where to look” for login loops, 401, empty farms — short checklist linking middleware + farm membership + `ADMIN_BIND_*`. |

---

## 3. Tasks-first / “tasks domination”

| Item | Notes |
|------|--------|
| **Primary journey** | Define one golden path (e.g. **Morning ops** = Tasks board + Alerts + one schedule status). Defer polishing secondary surfaces until that path works end-to-end. |
| **Tasks ↔ automation** | Document when a **rule** creates a **task** vs only touches **actuators**; list **gaps in product copy** (help text, empty states). |
| **Offline / queue** | Reconcile **docs** with actual **`taskWriteQueue` / sync failed** behavior so operators trust the UI. |

---

## 4. Multi-device hardening

| Item | Notes |
|------|--------|
| **Machine checklist** | [`docs/machine-setup-checklist.md`](../machine-setup-checklist.md) — use on every new machine; extend as failure modes appear. |

---

## 5. Relationship to Phase 25 (RAG)

Phase 25 plans should **assume** this sit-in stream has at least **operator tour + troubleshooting doc + minimal API/worker logging** underway; avoid stacking RAG UX on top of an opaque dashboard.

---

## Changelog

| Date | Note |
|------|------|
| 2026-04-21 | Stream created from operator bootstrap learnings (Compose, auth_test, seed, env-admin JWT binding). |
