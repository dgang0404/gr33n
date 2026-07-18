---
name: Phase 33 — Guardian polish & enterprise ops
overview: >
  Post-Phase-31 hardening for Guardian read tools (intent guards, smokes, doc parity),
  prompt hygiene with context_ref, optional read audit logging, hardware CI lane, and
  enterprise site-manifest provisioning. Small, high-value slices that do not block
  Phase 32 grow-setup PRs — run WS1 before or in parallel with Phase 32.
todos:
  - id: ws1-read-tools-hardening
    content: "WS1: Read tools hardening — exclude alert write intents from summarize_zone; smoke grounded chat → summarize_zone in prompt; persona mirror + architecture §10 Phase 31 read tools"
    status: completed
  - id: ws2-context-ref-dedup
    content: "WS2: context_ref dedup — skip summarize_zone when zone focus block already injected from Ask Guardian zone entry point; enrich focus block with readings so it is the single zone block"
    status: completed
  - id: ws3-read-tool-audit
    content: "WS3: Read-tool audit — info-level guardian_tool_read log (tool_id, farm_id, user_id, zone_id); document in audit playbook"
    status: completed
  - id: ws4-hardware-ci-lane
    content: "WS4: @hardware CI lane — build-tag GPIO smoke (hardware); GR33N_HARDWARE_TEST=1 runs the bench script; manual workflow_dispatch CI job; INSTALL + phase-14 docs"
    status: completed
  - id: ws5-enterprise-site-manifest
    content: "WS5: Enterprise site manifest — site-manifest.example.yaml + apply-site-manifest.sh (--dry-run + apply: org/farm/zones/recipe pack/Pi hints)"
    status: completed
  - id: ws6-docs-roadmap
    content: "WS6: Docs — README roadmap rows (32/33 shipped, 34-37 planned); phase-14 link; phase-31 Related cross-link"
    status: completed
isProject: false
---

# Phase 33 — Guardian polish & enterprise ops

## Status

**Shipped (WS1–WS6).** WS1 read-tool hardening; WS2 `context_ref` zone dedup (focus block enriched with readings); WS3 read-tool usage info log; WS4 `@hardware` build-tag test lane; WS5 enterprise site manifest stub; WS6 README/phase-14/phase-31 roadmap docs.

**Preconditions (met):**

- Phase 31 WS6 read tools shipped — [`internal/farmguardian/readtools.go`](../../internal/farmguardian/readtools.go), [`EnrichPromptBlock`](../../internal/farmguardian/readtools.go) in chat handler
- Phase 31 WS7 doc index — [`phase-14-operator-documentation.md`](../phase-14-operator-documentation.md#phase-31-field-validation-edge)
- Phase 30 PR queue + Phase 32 plan ready for grow-setup work — [`phase_32_guardian_grow_setup_prs.plan.md`](phase_32_guardian_grow_setup_prs.plan.md)

---

## Why this phase

Phase 31 shipped **read-only** Guardian tools (`list_unread_alerts`, `summarize_zone`) and proved the edge loop. A post-ship review surfaced **small but real** gaps: intent overlap with alert **write** proposals, missing smokes/docs parity, and enterprise/CI stories teased in topology docs but not planned.

Phase 32 owns **grow-setup PR bundles** and **platform doc RAG (WS8)** — the biggest operator wins. Phase 33 owns **polish and integrator ergonomics** so Phase 32 builds on a clean read-tool foundation.

```
Phase 31 (edge + read tools) ──► Phase 33 WS1 (hardening) ──► Phase 32 (setup packs + platform RAG)
                                      │
                                      └──► Phase 33 WS2–WS5 (parallel / after 32)
```

---

## Design principles

1. **No new autonomous writes** — read tools and audit only; manifests call public API with operator JWT.
2. **Minimal diff** — WS1 is a few guards + one smoke + doc rows; ship fast.
3. **Opt-in hardware** — GPIO/actuator bench tests never block default CI.
4. **Enterprise as scripts** — YAML manifest + API callers under [`scripts/enterprise/`](../../scripts/enterprise/README.md); no core "broadcast" tables.
5. **Phase 32 wins first** — if schedule slips, only WS1 is P0; WS2–WS5 defer without blocking grow-setup.

---

## Scope

| WS | Focus | Primary artifacts |
|----|-------|-------------------|
| **WS1** | Read tools hardening | `readtools.go`, smoke, persona + architecture docs |
| **WS2** | `context_ref` dedup | `EnrichPromptBlock` + handler passes ref |
| **WS3** | Read-tool audit | `slog` or auditlog info event; playbook row |
| **WS4** | Hardware CI lane | `//go:build hardware` or env gate; CI doc |
| **WS5** | Site manifest | `scripts/enterprise/site-manifest.example.yaml` + apply stub |
| **WS6** | Roadmap docs | README, phase-14, enterprise topology cross-links |

---

## Work-stream detail

### WS1 — Read tools hardening (P0 — do before Phase 32)

**Goal:** Close Phase 31 WS6 post-ship review gaps.

**Tasks:**

1. **Intent guard** — In `matchSummarizeZoneIntent` path (or `EnrichPromptBlock`), skip `summarize_zone` when `matchAlertToolIntent(question)` matches (same guard as `list_unread_alerts`). Stops ack/read messages that mention "humidity" + zone from injecting redundant sensor blocks.
2. **Optional:** Narrow `"tell me about "` summarize trigger when question also matches `listAlertsIntent` (single-zone farm false positive).
3. **Smoke** — `cmd/api/smoke_phase33_ws1_test.go`: seeded farm, POST `/v1/chat` grounded with humidity question + zone name; assert response or internal hook that enrichment ran (prefer testing `EnrichPromptBlock` via exported test helper or chat handler integration with mock LLM).
4. **Docs** — Update [`farm-guardian-persona-platform-context.md`](../farm-guardian-persona-platform-context.md) **Reads** row (`list_unread_alerts`, `summarize_zone`, Confirm N/A).
5. **Docs** — Extend [`farm-guardian-architecture.md`](../farm-guardian-architecture.md) §10 phase ledger + new § subsection for Phase 31 read tools (request flow: intent → DB → prompt injection).

**Shipped:** `shouldRunSummarizeZoneReadIntent` blocks alert write + alert-list questions from `summarize_zone`; [`cmd/api/smoke_phase33_ws1_test.go`](../../cmd/api/smoke_phase33_ws1_test.go); persona mirror + architecture §7.0 / code map / request flow / phase ledger updated.

**Acceptance:**

- `"acknowledge the humidity alert in Flower Room"` does **not** inject `summarize_zone` block.
- `"what's the humidity in Flower Room?"` still injects readings.
- Persona mirror matches `platform_context.go` read-tools line.
- Smoke green with `DATABASE_URL`.

---

### WS2 — `context_ref` dedup

**Goal:** Zone **Ask Guardian** already injects [`ContextRefPromptBlock`](../farm-guardian-architecture.md) for `type: zone`. Avoid duplicate `summarize_zone` sensor dump in the same turn.

**Tasks:**

- Pass optional `ContextRef` into `EnrichPromptBlock` (or skip summarize when handler already appended zone focus block).
- Skip `summarize_zone` when `ref.Type == "zone"` and resolved zone matches `ref.ID`.

**As-built (deviation, approved):** The zone focus block (`renderZoneContext`) previously only printed a sensor *count*, so literally skipping `summarize_zone` would have dropped readings and regressed WS1. Instead: extracted `renderZoneSensorReadings` (shared helper), **enriched the focus block to carry the same latest readings**, then skip `summarize_zone` for that zone. Result is exactly **one** zone block that still includes readings.

**Shipped:** `EnrichPromptBlock` takes `*ContextRef`; `zoneContextRefCovers` gates the skip; `renderZoneSensorReadings` shared by `summarize_zone` + zone focus block; handler passes `pb.ContextRef`; smokes [`cmd/api/smoke_phase33_ws2_test.go`](../../cmd/api/smoke_phase33_ws2_test.go) (matching zone ref skips summarize_zone + focus carries readings; non-zone ref keeps summarize_zone). `go build`, `go vet`, and `go test ./internal/farmguardian/...` green.

**Acceptance:** Open Guardian from zone card + ask about humidity → one zone block, not two identical sensor lists.

---

### WS3 — Read-tool audit (info)

**Goal:** Enterprise operators can answer *"who asked Guardian about Flower Room humidity yesterday?"* without full write audit weight.

**Tasks:**

- On successful read-tool enrichment, `slog.Info` with structured fields: `tool_id`, `farm_id`, `user_id` (if auth), `zone_id` / `alert_count`.
- Optional: `auditlog` info event type `guardian_tool_read` (defer if audit enum migration is heavy).
- Row in [`audit-events-operator-playbook.md`](../audit-events-operator-playbook.md) or Guardian architecture § audit.

**Acceptance:** Log line visible in Loki/docker logs on read-tool turn; no Confirm required.

**Shipped:** `logReadToolUse` emits `slog.Info` (`event=guardian_tool_read`, `tool_id`, `farm_id`, `user_id` from `authctx` when present, `zone_id` for zone tools) per fired read tool in [`EnrichPromptBlock`](../../internal/farmguardian/readtools.go); documented in [`audit-events-operator-playbook.md`](../audit-events-operator-playbook.md) § Guardian read-tool usage. No DB audit row / enum migration (deferred per plan).

**Not in scope:** Persisting read history to DB table (v2 if needed).

---

### WS4 — `@hardware` CI lane

**Goal:** Phase 31 WS8 documented hardware tests skipped in CI unless opted in — make that **real**.

**Tasks:**

- Tag or gate tests: `GR33N_HARDWARE_TEST=1` (existing env from Phase 31 plan) or build tag `hardware`.
- Document in [`INSTALL.md`](../../INSTALL.md) + phase-14 index.
- Optional GitHub Actions job `hardware-smoke` — `if: github.event_name == 'workflow_dispatch'` or label-triggered.

**Acceptance:** Default `make test` skips GPIO tests; `GR33N_HARDWARE_TEST=1 make test` runs them when Pi/bench attached.

**Shipped:** Live GPIO E2E moved to [`cmd/api/smoke_hardware_test.go`](../../cmd/api/smoke_hardware_test.go) behind `//go:build hardware` (excluded from `make test` / CI `-tags dev`). With `GR33N_HARDWARE_TEST=1` it actually runs [`scripts/run-edge-actuator-smoke.sh`](../../scripts/run-edge-actuator-smoke.sh) (no longer a no-op double-skip). Manual `hardware-smoke` job in [`.github/workflows/ci.yml`](../../.github/workflows/ci.yml) (`workflow_dispatch`, self-hosted `gr33n-hardware` runner). Run cmd documented in [`INSTALL.md`](../../INSTALL.md) §6 + phase-14 index.

---

### WS5 — Enterprise site manifest

**Goal:** Integrators get one YAML file describing a warehouse site → script creates farm, zones, imports recipe pack, prints smoke commands.

**Tasks:**

- [`scripts/enterprise/site-manifest.example.yaml`](../../scripts/enterprise/site-manifest.example.yaml) — illustrative schema:
  - `org_slug`, `farm_name`, `zones[]` (name, type), `recipe_pack_slug`, `pi_device_hints`
- [`scripts/enterprise/apply-site-manifest.sh`](../../scripts/enterprise/apply-site-manifest.sh) — `--dry-run`, calls public API (like `import-recipe-pack.sh`).
- Cross-link [`hypothetical-enterprise-topology.md`](../hypothetical-enterprise-topology.md) deployment pipeline section.

**Acceptance:** `--dry-run` prints POST/import steps; real run needs local API + admin JWT.

**Shipped:** [`scripts/enterprise/site-manifest.example.yaml`](../../scripts/enterprise/site-manifest.example.yaml) (org_slug, farm_name, zones[], recipe_pack_slug, pi_device_hints) + [`scripts/enterprise/apply-site-manifest.sh`](../../scripts/enterprise/apply-site-manifest.sh) with `--dry-run` (no JWT/HTTP) and apply (POST /farms, /farms/{id}/zones, /farms/{id}/commons/catalog-imports; zones idempotent by name). README WS5 section + topology cross-link.

**Not in scope:** Full 500-site Ansible suite (community PRs welcome).

---

### WS6 — Roadmap docs

**Goal:** README and operator indexes show Phase 33 without confusing Phase 32.

**Tasks:**

- README roadmap row: Phase 33 **Planned** (polish + enterprise ops).
- [`phase-14-operator-documentation.md`](../phase-14-operator-documentation.md) quick link to this plan.
- Phase 31 plan **Related** table → Phase 33 follow-up.

**Acceptance:** New chat prompt `@phase_33_guardian_polish_and_enterprise_ops.plan.md` resolves.

**Shipped:** [`README.md`](../../README.md) roadmap updated (Phase 32 + 33 marked shipped; Phases 34–37 planned rows); [`phase-14-operator-documentation.md`](../phase-14-operator-documentation.md) links this plan + the `@hardware` lane; [`phase_31_field_validation_and_edge.plan.md`](phase_31_field_validation_and_edge.plan.md) Related table points here.

---

## Relationship to other phases

| Phase | Relationship |
|-------|----------------|
| **31** | WS6 read tools — WS1 here hardens them |
| **32** | Grow-setup PRs + platform doc RAG — **main feature track**; run after/beside WS1 |
| **14 / enterprise topology** | WS5 manifest extends `scripts/enterprise/` story |
| **29** | Sketched optional read audit — WS3 implements lightly |

### What stays in Phase 32 (not duplicated here)

| Item | Phase |
|------|-------|
| Setup pack PR (`apply_grow_setup_pack`) | **32 WS3** |
| Platform doc RAG (`rag-ingest-platform-docs.sh`) | **32 WS8** |
| `list_plants`, `summarize_zone_fertigation` read expansion | **32 WS1** |

---

## Suggested implementation order

1. **WS1** — hardening (1 session; unblocks confidence for Phase 32)
2. **Phase 32** — grow-setup + WS8 (parallel mega-track)
3. **WS2** — context_ref dedup (when zone Ask Guardian feels noisy)
4. **WS6** — doc row (can ship with WS1)
5. **WS4** — hardware CI (when bench available)
6. **WS3** — read audit (when enterprise asks)
7. **WS5** — site manifest (integrator-driven)

---

## Definition of done (phase ship)

- [x] WS1 intent guard + smoke + persona/architecture doc parity
- [x] WS2 context_ref dedup (zone focus block enriched with readings; summarize_zone skipped for that zone)
- [x] WS4 hardware gate (build tag `hardware` + env) documented; default CI unchanged; manual `hardware-smoke` lane
- [x] WS6 README + phase-14 cross-links
- [x] WS3 + WS5 both shipped (read audit log **and** site manifest stub)

**Minimal ship:** WS1 + WS6 only — still worth tagging a release note.

---

## Using this plan in a new chat

```text
Implement Phase 33 WS1 per @docs/plans/archive/phase_33_guardian_polish_and_enterprise_ops.plan.md.

Harden Phase 31 read tools: exclude alert write intents from summarize_zone,
add smoke for humidity→summarize_zone, update persona mirror + architecture §10.
Then proceed to Phase 32 grow-setup PRs.
```

---

## Related

| Doc | Role |
|-----|------|
| [`phase_31_field_validation_and_edge.plan.md`](phase_31_field_validation_and_edge.plan.md) | Read tools origin |
| [`phase_32_guardian_grow_setup_prs.plan.md`](phase_32_guardian_grow_setup_prs.plan.md) | Next feature phase |
| [`hypothetical-enterprise-topology.md`](../hypothetical-enterprise-topology.md) | WS5 manifest context |
| [`scripts/enterprise/README.md`](../../scripts/enterprise/README.md) | Script home |
