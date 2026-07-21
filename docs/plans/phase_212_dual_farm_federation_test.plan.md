---
name: Phase 212 — Dual-install federation test (Farm A/Org A + Farm B/Org B + Insert Commons receiver)
overview: >
  Stand up a second, fully independent gr33n clone (own folder, own Postgres,
  own ports) as "Install B" — Farm B under Organization B — alongside the
  existing install ("Install A" — Farm A under Organization A). Stand up the
  real Insert Commons receiver as the third-party service both installs talk
  to. Use the pair to answer, with evidence instead of guesswork, which
  knowledge categories actually move between independent gr33n installs
  (Commons Catalog packs, Insert Commons aggregates) and which never do
  (field guides, platform docs, operational RAG, symptom guides — all
  farm_id-scoped inside one database and never designed to sync).
todos:
  - id: ws1-install-b-bringup
    content: "WS1: Clone repo into sibling folder as Install B; remap Docker Compose ports (DB/API/UI) so both stacks run concurrently without conflict"
    status: pending
  - id: ws2-orgs-and-seed
    content: "WS2: Assign Organization A to existing Farm A (Install A); write farm_b_seed.sql (Organization B + Farm B) for Install B"
    status: pending
  - id: ws3-insert-commons-receiver
    content: "WS3: Run cmd/insert-commons-receiver as the shared third-party service; point both installs' INSERT_COMMONS_INGEST_URL at it; opt-in + sync from both farms; verify /v1/stats shows 2 distinct farm_pseudonyms"
    status: pending
  - id: ws4-commons-catalog-portability
    content: "WS4: Publish a pack from Farm A, export its JSON, hand-import into Farm B's catalog on Install B; confirm cross-install portability works for Commons Catalog packs only"
    status: pending
  - id: ws5-negative-controls
    content: "WS5: Negative controls — confirm field guides, platform docs, operational RAG chunks, and symptom guides on Install A are NOT visible/importable on Install B (documents the real per-farm/per-database boundary)"
    status: pending
  - id: ws6-runbook-and-glossary
    content: "WS6: Write docs/dual-farm-federation-test-runbook.md; glossary table shipped in workflow-guide.md §11a (operator-tour cross-link optional)"
    status: pending
isProject: false
---

# Phase 212 — Dual-install federation test

**Status:** Planned · **Depends on:** existing Install A (this repo) already running · **Motivation:** operator confusion about which knowledge categories (Commons Catalog, Insert Commons, Field guides, Platform docs, Operational RAG, Symptom guides) actually cross a farm boundary vs. an install boundary — settle it with a real second install instead of more docs.

## The one job

> Two independent gr33n installs (Install A / Farm A / Org A, and Install B /
> Farm B / Org B) running side by side, connected only through the one thing
> that's actually designed to connect independent installs — the **Insert
> Commons receiver** — plus a manual Commons Catalog pack hand-off. Everything
> else (field guides, platform docs, operational RAG, symptom guides) is
> proven to stay local to its own database.

## Why this shape (read before building)

Two real architecture facts drove this plan away from "clone the whole thing and expect it all to sync":

1. **`organizations` and `farms` already live in one shared database** (`organization_id` FK on `farms`). Cross-*farm* testing (two farms, two orgs) does **not** require two installs — it only needs two farm rows. We still do two installs here because the receiver test genuinely needs two independent senders.
2. **Only two things are designed to cross an install boundary:**
   - **Commons Catalog packs** — a human exports JSON from one install and imports it on another (no live sync; see `commons-catalog-operator-playbook.md`).
   - **Insert Commons** — a farm POSTs anonymized aggregates to an external receiver over HTTP (`internal/insertcommonsreceiver`, `cmd/insert-commons-receiver`). This is the **only** live, networked, cross-install feature in the platform today.
3. **Field guides, platform docs, operational RAG chunks, and symptom guides are `farm_id`-scoped rows in one Postgres.** They are never exported, never synced, never visible to another install. Proving this (WS5) is as valuable as proving the two things that *do* cross.

## Before you start — memory budget

Phase 212 does **not** need Guardian or Ollama on either install. Keep RAM for two API + two Postgres containers + one receiver.

| Install | Guardian / AI | Notes |
|---------|---------------|-------|
| **A (this repo)** | **Rest now** before WS1 — Settings → Farm Guardian readiness → **Rest now**, or `./scripts/guardian-power.sh sleep` to stop Ollama entirely | `AI_ENABLED` may stay true; we simply don't awaken models during the test |
| **B (clone)** | **`AI_ENABLED=false`** in `.env` | No Ollama, no RAG ingest, no smoke tests |

**Operator:** reboot the laptop first if free RAM is under ~2 GB. Start Install B's UI (`:5174`) only while clicking through a step; `docker compose stop ui` in Install B between sessions.

## Execution strategy — gaps, errors, and when to fix

Do **not** waterfall-fix every surprise and restart from zero — that burns days. Do **not** ignore real bugs and only keep a list — that leaves false conclusions in the runbook. Use three tiers:

### Tier A — Stop, fix, restart **from the affected workstream only**

Use when the issue **invalidates acceptance criteria** or means the product is broken (not a documented limitation).

| Symptom | Action |
|---------|--------|
| Install B won't start (port conflict, bad `.env`) | Fix setup; re-run **WS1** only — Install A untouched |
| Receiver never accepts POST (auth, migration missing) | Fix receiver; re-run **WS3** from opt-in onward |
| Commons pack import on B fails where A succeeded | **Waterfall fix** if it's a real import bug — fix code, reset B's farm/catalog test data, re-run **WS4** |
| Negative control **wrong** (e.g. field guides magically appear on B without local ingest) | **Stop** — that's unintended sync; file Tier A bug, fix, wipe B DB or reinstall B, re-run **WS5** |

Log each Tier A item in `docs/dual-farm-federation-test-runbook.md` under **Incidents** with: WS, symptom, root cause, fix commit, restart point.

### Tier B — Log as **documented finding**, continue the phase

Use when behavior matches **known product limits** — the test is *proving* the limit, not failing.

| Finding | Example | Action |
|---------|---------|--------|
| No live "publish catalog to remote server" API | WS4 requires manual JSON copy | Note in runbook **Expected finding** — not a bug |
| Install B field guides 404 until local migrate + ingest | WS5 negative control | **Pass** — record curl/screenshot |
| Insert Commons sync `skipped_no_receiver` until URL set | WS3 setup step | Fix env only (Tier A if misconfigured after URL set) |

### Tier C — **Bug backlog**, fix after phase 212 closes

Use for issues that don't block understanding or acceptance criteria.

| Example | Action |
|---------|--------|
| UI copy still confusing after glossary shipped | Open follow-up issue / small docs PR |
| Publish from Farm missing summary validation | Backlog — note in runbook |
| Cosmetic Settings layout on Insert Commons | Backlog |

### Decision rule (when unsure)

Ask: *If we document this behavior as-is, would a new dev draw the **wrong** architecture conclusion?*

- **Yes** → Tier A (fix or prove it's impossible before moving on).
- **No, it's a known gap** → Tier B (expected finding).
- **No, it's polish** → Tier C (backlog).

### Phase 212 is **done** when

Acceptance criteria checkboxes pass **or** every failed checkbox has a Tier A incident with a tracked fix — not when every Tier C item is closed.

## WS1 — Install B bring-up

- `cd ~ && git clone https://github.com/dgang0404/gr33n.git gr33n-platform-b` (or `git clone <local path>` if offline)
- Add `docker-compose.override.yml` in Install B only (gitignored, not committed) remapping host ports to avoid collision with Install A:
  | Service | Install A (this repo) | Install B |
  |---|---|---|
  | Postgres | 5433 | 5434 |
  | API | 8080 | 8081 |
  | UI | 5173 | 5174 |
- `./scripts/setup-first-clone.sh --docker` then `./scripts/bootstrap-local.sh --docker --seed` inside Install B, pointed at its own `.env` (`DATABASE_URL` on 5434, **`AI_ENABLED=false`**).
- See **Before you start — memory budget** above (Install A Rest now, reboot if needed).
- Acceptance: `curl :8081/health` OK and `curl :8080/health` OK at the same time.

## WS2 — Organizations + Farm B seed

- Install A: one-time `UPDATE gr33ncore.farms SET organization_id = (INSERT INTO gr33ncore.organizations... RETURNING id) WHERE id = 1;` — small migration, names it "Org A".
- Install B: new `db/seeds/farm_b_seed.sql` (not committed to shared history — Install-B-local, or committed under a clearly-labeled `db/seeds/` variant if useful to keep for future testers) creating Organization B + Farm B with different names/timezone/currency from Farm A so screenshots and audit trails are unambiguous about which install you're looking at.
- Acceptance: Install A dashboard shows "Farm A" / org shows "Organization A" (Settings); Install B dashboard shows "Farm B" / "Organization B".

## WS3 — Insert Commons receiver as the connector

- Run **one** receiver for both installs (that's the point — it's the neutral third party): `make run-receiver` with its own `DATABASE_URL` (reuse Install A's Postgres with a distinct database name, e.g. `gr33n_insertcommons`, to avoid a third full Postgres container) and `INSERT_COMMONS_RECEIVER_LISTEN=:8765`.
- Apply the two receiver migrations (`20260417_phase13_insert_commons_receiver.sql`, `20260425_insert_commons_receiver_idempotency_stats.sql`) against that database.
- Set the **same** `INSERT_COMMONS_SHARED_SECRET` in Install A's `.env`, Install B's `.env`, and the receiver's env.
- Both installs: `INSERT_COMMONS_INGEST_URL=http://127.0.0.1:8765/v1/ingest`.
- In each install's UI (Settings → Insert Commons): opt in, `Run sync` (or `PATCH .../opt-in` + `POST .../sync` via curl if UI restart is inconvenient).
- Acceptance: `GET :8765/v1/stats` (Bearer secret) shows **2 distinct `farm_pseudonym`s** and non-zero rows — direct proof the receiver is the shared connective tissue, not the installs talking to each other directly.

## WS4 — Commons Catalog portability drill

- Install A: **Help → Import → Publish from Farm** → export a pack (e.g. the JADAM starter pack or a fertigation program bundle) to JSON.
- Copy that JSON to Install B's machine/folder (scp, or just copy-paste since same laptop).
- Install B: hand-insert as a new `commons_catalog_entries` row (SQL insert mirroring the migrations under `db/migrations/2026*_commons*` and `20260527_phase31_commons_recipe_pack_v7.sql` as a template) or via whatever admin path exists — **no live API for "publish this to another server's catalog" exists today**, which is itself a documented finding, not a bug to fix in this phase.
- Install B: **Import to Farm** on the hand-copied entry → confirm it creates the same recipes/inputs it would on Install A.
- Acceptance: the pack works on Install B; the *manual copy step* (no live sync) is captured in WS6 as the documented reality, not glossed over.

## WS5 — Negative controls (the part that resolves the confusion)

On Install B, confirm **without** re-ingesting or copying anything from Install A:

- Field guide `natural-farming-jlf-general` (present on Install A) returns 404 / "not found" on Install B until Install B runs its **own** field-guide migration + `make rag-ingest-field-guides` locally.
- Platform docs indexed on Install A are absent from Install B's Knowledge search.
- Operational RAG chunk counts (Settings → Field memories) start at 0 on Install B regardless of Install A's counts.
- Symptom guide catalog entries only exist on Install B if Install B's own migrations/seed created them.

Acceptance: a short table of "present on A / absent on B until locally ingested" screenshots or `curl` outputs for each of the four categories.

## WS6 — Runbook + glossary

- **Glossary (shipped):** `docs/workflow-guide.md` §11 + **§11a Farm knowledge — how the pieces connect** — layers, "If you want X do Y", install-boundary table, and full Insert Commons coarse-stats explanation.
- New `docs/dual-farm-federation-test-runbook.md`: exact commands for WS1–WS5, **Incidents** section (Tier A/B/C log), expected findings — repeatable without re-deriving.

## Acceptance criteria

- [ ] Install A and Install B run concurrently, distinct ports, distinct databases
- [ ] Farm A/Org A and Farm B/Org B visible and distinguishable in each UI
- [ ] Receiver `/v1/stats` shows 2 farm_pseudonyms after both installs sync
- [ ] One Commons Catalog pack hand-carried from A → B and successfully imported
- [ ] Field guides / platform docs / operational RAG / symptom guides confirmed **not** to cross without local re-ingest
- [ ] Runbook committed (glossary §11a already in workflow-guide)
- [ ] Tier A/B/C incident log filled during execution (even if empty)

## Out of scope

- Building a live "publish catalog pack to remote server" API (WS4 documents the current manual reality; a real feature would be its own future phase)
- Multi-org billing/plan-tier behavior
- Running Guardian/Ollama on Install B (not needed for this phase's acceptance criteria; revisit only if a future phase needs cross-install Guardian testing)
- Production TLS/secrets hardening for the receiver (lab-only `INSERT_COMMONS_RECEIVER_ALLOW_INSECURE_NO_AUTH` is fine here if a shared secret is more friction than value for a same-laptop test)
