---
name: Phase 15 Farm onboarding & templates
overview: >
  Product-focused onboarding: when a farm is created, operators can choose a blank slate
  or a versioned “starter protocol” pack (zones, JADAM-style inputs/recipes, light + irrigation
  schedules, fertigation baselines, demo tasks) derived from the same patterns as
  `db/seeds/master_seed.sql` — without hard-coding farm_id = 1.
todos:
  - id: farm-template-api
    content: "API + DB: `POST /farms` (or follow-up `POST /farms/{id}/apply-template`) accepts optional `bootstrap_template` (e.g. `none` | `jadam_indoor_photoperiod_v1`); idempotent apply function per farm."
    status: completed
  - id: farm-template-sql
    content: "Refactor seed logic into reusable SQL (e.g. `gr33ncore.apply_farm_bootstrap_template(p_farm_id bigint, p_template text)`) or versioned migration snippets; keep `master_seed.sql` as dev convenience or thin wrapper."
    status: completed
  - id: farm-template-ui
    content: "Dashboard: farm create / settings — ‘Start blank’ vs ‘Indoor photoperiod starter (18/6 veg / 12/12 flower)’; show what will be created; respect RBAC (farm admin)."
    status: completed
  - id: org-default-template
    content: "Optional: org-level default template for new farms under that org (Phase 13 org model)."
    status: completed
  - id: operator-deployment-runner
    content: "Guided bootstrap for non-IT operators: documented one-path setup and/or a small script runner for DB migrate, API, UI, Pi client, env templates, and first admin user — reduce copy-paste and env-var errors."
    status: pending
isProject: false
---

# Phase 15 — Farm onboarding & templates

## Why a separate phase

Phase 14 prioritizes **edge connectivity, insert pipeline, commons, and federation**. **Farm bootstrap** is orthogonal: it is pure **product onboarding** and touches farm creation, templates, and UI — but it reuses the same **data shapes** already proven in [`db/seeds/master_seed.sql`](../../db/seeds/master_seed.sql) (v1.005+).

Coordinate with Phase 14 only where **WS9** (below) overlaps on documentation or operator expectations.

## Goals

| Goal | Detail |
|------|--------|
| **Choice** | Every new farm: **blank** or **template** — never forced demo data. |
| **Not farm 1 only** | Templates apply to **any** `farm_id` via parameterized SQL or service job. |
| **Safe re-run** | Idempotent apply (same template + version → no duplicates / or explicit replace policy). |
| **Versioned** | Template name + schema version so upgrades and migrations stay explicit. |

## Non-goals (for this phase)

- Replacing `master_seed.sql` for **local dev** (keep one-command demo DB story).
- Multi-tenant “marketplace” of user-uploaded templates (defer).

## Follow-on: guided deployment & environment setup (proposed)

**Problem:** A motivated **non-developer** (farm operator, coop IT-light) may need to stand up **Postgres + migrations**, **farm API**, **dashboard UI**, **Pi / edge client**, and dozens of **environment variables** using only docs and shell scripts. That is error-prone and easy to abandon.

**Direction (track under Phase 15 or a thin “Phase 15b”):**

| Track | Outcome |
|-------|---------|
| **Docs** | Single “start here” path: prerequisites, order of operations (DB → API → UI → optional receiver → Pi), `.env.example` pointers, and links to Insert Commons / audit playbooks (including **strict ingest JSON** — no extra top-level keys; see [`insert-commons-pipeline-runbook.md`](../insert-commons-pipeline-runbook.md)). |
| **Script runner (optional)** | One entrypoint (e.g. `make bootstrap-local` or a small shell/Go helper) that checks prerequisites, runs migrations, prints or merges env from templates, and documents how to create the first user — **without** replacing Docker/Kubernetes choices for advanced operators. |
| **Scope guardrails** | Prefer idempotent, readable steps over a black-box installer; keep security-sensitive steps (secrets, TLS) explicit. |

See todo **`operator-deployment-runner`** in the frontmatter.

## Scope: `farm-template-ui` (implemented)

| Area | What ships |
|------|------------|
| **Surface** | Settings → **New farm**: name, timezone, currency, optional org link, starting-content radio (blank / starter pack / org default). |
| **Starter copy** | `ui/src/constants/bootstrapTemplates.js` holds the canonical key (`jadam_indoor_photoperiod_v1`) and a bullet list aligned with `db/migrations/20260423_farm_bootstrap_templates.sql`. Expandable `<details>` when the starter is selected. |
| **Behavior** | Blank sends `bootstrap_template: "none"`. Starter sends the chosen key. Org default omits `bootstrap_template` so the API can apply the org’s stored default. Success switches the session to the new farm. |
| **RBAC** | Farm create remains API-enforced (owner on create, org admin when `organization_id` is set). UI only lists orgs where the user is owner/admin for link + org default controls. |

**Follow-ups (not required for this scope):** dedicated onboarding route, preview API for template contents, non-settings entry points.

## Scope: `org-default-template` (implemented)

| Area | What ships |
|------|------------|
| **Data** | `gr33ncore.organizations.default_bootstrap_template` (nullable `TEXT`); migration `db/migrations/20260424_organization_default_bootstrap_template.sql`. |
| **API** | `PATCH /organizations/{id}` accepts `default_bootstrap_template` (string or JSON `null` to clear). Key presence is detected via raw JSON so `null` clears without clobbering other fields. List/get organization responses include the column. |
| **Farm create** | If `bootstrap_template` is **omitted** and `organization_id` is set, the handler loads the org and applies a non-empty default the same way as an explicit template (wrapped `{ farm, bootstrap }`). `bootstrap_template: "none"` still skips the org default. |
| **UI** | Settings → Organizations: per-org row (owner/admin) with “Default template for new farms” + Save. |
| **Tests** | `TestOrgDefaultBootstrapOnFarmCreate` in `cmd/api/smoke_test.go`. |

## Using this plan in a new chat

Reference `@docs/plans/phase_15_farm_onboarding.plan.md` and `@docs/plans/phase_14_network_and_commons.plan.md` (WS9) together when scoping farm creation work. For Insert Commons integration pitfalls (custom POST bodies), point integrators at `@docs/insert-commons-pipeline-runbook.md` § *Custom senders*.
