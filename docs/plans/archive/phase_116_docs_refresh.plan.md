---
name: Phase 116 — Documentation refresh (env reference, upgrade path, API browsing)
overview: >
  The July 2026 docs audit found the README strong on breadth but stale on recency
  (caps at Phase 110), zero screenshots anywhere, an incomplete env-var reference,
  a backup/restore runbook hidden under a cutover doc name, no upgrade guide, no
  CHANGELOG or CONTRIBUTING, and openapi.yaml that can only be browsed by pasting
  into an external site. This phase makes the docs match the product.
todos:
  - id: ws1-env-reference
    content: "WS1: Env reference — extend .env.example with all LLM_*, GUARDIAN_*, STT_*, FCM_*, CHAT_* vars; new docs/environment-variables.md generated/verified against os.Getenv usage; retire stale docs/example-env.md"
    status: completed
  - id: ws2-readme-refresh
    content: "WS2: README — status banner + feature list through Phase 118 plans; model selector + Ollama pull in features; enterprise scripts positioning fixed; screenshots section"
    status: completed
  - id: ws3-upgrade-backup
    content: "WS3: Upgrade + backup — docs/upgrade-guide.md (pull → migrate → restart order, version-specific runbook links); backup/restore extracted from receipt-storage-cutover-runbook.md into docs/backup-restore-runbook.md; linked from README, INSTALL.md, troubleshooting"
    status: completed
  - id: ws4-changelog-contributing
    content: "WS4: CHANGELOG.md (operator-visible changes: API breaks, migrations, Guardian behavior, back-filled from phase closures at coarse grain) + CONTRIBUTING.md (test gates, make targets, plan lifecycle)"
    status: completed
  - id: ws5-api-browsing
    content: "WS5: API browsing — serve Swagger UI (or Redoc) for openapi.yaml at /openapi on the API in dev builds; docs/api-quickstart.md curl cookbook (auth, farm, Pi key, chat)"
    status: completed
  - id: ws6-screenshots
    content: "WS6: Screenshots — docs/images/ with dashboard, Guardian chat + Confirm card, Pi wizard, model selector; referenced from README + operator tour"
    status: completed
  - id: ws7-operator-tour
    content: "WS7: Operator tour — add Guardian model picker + pull workflow section (Phase 111/112); verify tour steps against current SPA routes"
    status: completed
isProject: false
---

# Phase 116 — Documentation refresh (env reference, upgrade path, API browsing)

## Status

**Shipped.** Env reference + CI parity, README/CHANGELOG/CONTRIBUTING, upgrade & backup guides, `/openapi` Redoc UI, api-quickstart, screenshots, operator tour model picker section.

---

## Findings driving this phase

| # | Gap | Evidence |
|---|-----|----------|
| 1 | Guardian env vars undocumented centrally — `.env.example` lacks `LLM_BASE_URL`/`LLM_MODEL`; `GUARDIAN_OLLAMA_*` only in INSTALL.md; `STT_*`/`FCM_*` only in playbooks; `docs/example-env.md` stops at `RAG_SYNTHESIS_*` | audit of `os.Getenv` vs docs |
| 2 | No runtime API browsing — `openapi.yaml` exists, CI-enforced parity, but README says "paste into editor.swagger.io" | no route serves it |
| 3 | Backup/restore buried — the real runbook is `docs/receipt-storage-cutover-runbook.md`, unfindable by name | README mentions it once under S3 |
| 4 | No single upgrade guide — steps split across INSTALL.md §migrations, bootstrap doc, per-phase runbooks | — |
| 5 | README recency — status stops at 110; Phases 111–112 shipped | `README.md` line 10 |
| 6 | No CHANGELOG / CONTRIBUTING at root | — |
| 7 | Zero images in `docs/` — Pi wizard, SPA workspaces, Guardian UI have no visual docs | `find docs -name '*.png'` empty |
| 8 | Operator tour missing model picker / pull workflow | `docs/operator-tour.md` |

Already strong (no action): INSTALL.md, first-session doc, operator troubleshooting,
pi-integration guide, field guides, 10+ playbooks, SECURITY.md, LICENSE, openapi
parity test.

---

## Design notes

### WS1 — Environment reference

One source of truth: `.env.example` gets every operator-relevant var with a one-line
comment and safe default. `docs/environment-variables.md` groups them (core, DB, auth,
Guardian/LLM, RAG/embeddings, push, storage, edge) with links to the deeper doc per
group. Add a small CI check (grep-based) that every `os.Getenv` literal in `cmd/` and
`internal/` appears in the reference — same spirit as the openapi parity test.

### WS5 — API browsing

Embed a static Swagger UI (vendored assets, no CDN — offline-first rule) served at
`/openapi` behind the same dev/prod gating as other diagnostics. Cookbook doc covers
the four flows integrators actually ask about: login + JWT use, farm/zone CRUD,
minting a Pi device key, and a Guardian chat call with model override.

### WS4 — CHANGELOG grain

Not 112 retroactive entries. Back-fill coarse blocks (10–33, 40–67, 68–81, 82–110,
111–112) with operator-visible changes only, then per-phase entries going forward.

### Out of scope

- Docs site generator (mkdocs/docusaurus) — plain markdown stays; revisit if docs
  outgrow GitHub browsing
- Video walkthroughs
- Translated docs

---

## Acceptance

- [ ] Every `os.Getenv` var in `cmd/` + `internal/` documented; CI check green
- [ ] Fresh reader can find backup/restore and upgrade steps from README in ≤2 clicks
- [ ] `/openapi` renders full spec in a browser on a dev install, offline
- [ ] CHANGELOG.md + CONTRIBUTING.md exist at root; README links them
- [ ] README status/features current through the planned-phase ledger; screenshots render
- [ ] Operator tour covers model selector + pull; steps verified against SPA
- [ ] `docs/example-env.md` retired or reduced to a pointer

---

## Files expected to change

| Area | Files |
|------|-------|
| Root | `README.md`, `CHANGELOG.md` (new), `CONTRIBUTING.md` (new), `.env.example` |
| Docs | `docs/environment-variables.md` (new), `docs/upgrade-guide.md` (new), `docs/backup-restore-runbook.md` (new), `docs/api-quickstart.md` (new), `docs/operator-tour.md`, `docs/images/` (new), `docs/example-env.md` |
| API | small static-serve handler for `/openapi` + vendored swagger assets |
| CI | env-reference parity check |
