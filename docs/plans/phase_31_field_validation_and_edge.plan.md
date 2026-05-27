---
name: Phase 31 — Field validation & safe edge stories
overview: >
  Close the loop between software-on-a-laptop and real hardware: Pi/breadboard
  validation, documented safe actuator paths, simulated + live sensor readings on
  the dashboard, and optional deployment-script seeds for multi-site integrators.
  Guardian may enqueue Pi commands only via Phase 30 confirmed change requests;
  this phase proves the edge actually executes them safely.
todos:
  - id: ws1-breadboard-smoke
    content: "WS1: Breadboard / stub loop — document + optional smoke: pi_client on laptop posts readings; dashboard Live Sensors shows data; automation simulation off path documented"
    status: completed
  - id: ws2-pi-contract-field
    content: "WS2: Pi field checklist — extend pi-integration-guide with warehouse-room wiring sketch, one-relay-safe-test, offline queue drill; align with TestPiContract* smokes"
    status: completed
  - id: ws3-one-actuator-story
    content: "WS3: Safe actuator story — end-to-end: Phase 30 PR confirm OR rule → pending_command → Pi executes → actuator_events → audit; E-stop checklist (doc)"
    status: pending
  - id: ws4-mqtt-room-scale
    content: "WS4: MQTT room-scale pattern — topic convention for multi-zone warehouse; bridge config example; batch ingest load note (not performance guarantee)"
    status: pending
  - id: ws5-recipe-pack-demo
    content: "WS5: Recipe pack promotion demo — sample commons catalog body for fertigation program v1→v2; script stub in scripts/enterprise/ importing to two farm_ids"
    status: pending
  - id: ws6-guardian-read-tools
    content: "WS6: Guardian read-only edge tools — list unread alerts, summarize zone snapshot (read-only); actuator enqueue remains Phase 30 PR path only"
    status: pending
  - id: ws7-enterprise-doc-link
    content: "WS7: Cross-link enterprise topology doc + phase-14 playbooks; README phase banner"
    status: pending
  - id: ws8-openapi-tests
    content: "WS8: Smokes — live reading on dashboard path; optional tagged hardware test skipped in CI"
    status: pending
isProject: false
---

# Phase 31 — Field validation & safe edge stories

## Status

**In progress (WS1–WS2 shipped).** Phase 29 (Guardian agent layer) should reach **WS6–WS9** ship criteria first. Phase 30 (Guardian change requests) can land before or in parallel with Phase 31 WS1 — field bench work validates that **confirmed PRs** reach real GPIO.

**Preconditions (already met or in progress):**

- Pi HTTP contract + smokes: [`cmd/api/smoke_pi_contract_test.go`](../../cmd/api/smoke_pi_contract_test.go)
- [`pi_client/gr33n_client.py`](../../pi_client/gr33n_client.py) — stub drivers on non-Pi hosts
- MQTT bridge: [`pi_client/mqtt_telemetry_bridge.py`](../../pi_client/mqtt_telemetry_bridge.py)
- Multi-site **thought experiment** doc: [`docs/hypothetical-enterprise-topology.md`](../hypothetical-enterprise-topology.md)
- Enterprise script hook: [`scripts/enterprise/README.md`](../../scripts/enterprise/README.md)
- Guardian **PR queue** (Phase 30): confirmed `enqueue_actuator_command` (or equivalent) writes `pending_command` — this phase proves the Pi side

---

## Why this phase

After Phase 29–30, gr33n is a **credible farm OS** with a safe configuration agent (human-approved change requests). Operators still see **404 on sensor readings** until something edge-shaped posts data. Phase 31 is about **proving the field loop** and **documenting safe physical I/O** — not about becoming a vertical-farm MES.

Parallel activity (README already says this): Pi / MQTT validation can start **before** Phase 29 is fully done; this plan **names and gates** that work.

---

## Design principles

1. **Software reuse** — edge paths call the same routes the smokes assert; no parallel GPIO API.
2. **Safety first** — one relay, one pump, manual E-stop story before "automate the warehouse."
3. **Actuator writes go through Phase 30 PRs** — Guardian never bypasses confirm; automation **rules/alerts** remain the autonomous layer (by design).
4. **Offline is real** — exercise `offline_queue_path` flush at least once in docs/smoke.
5. **Scale is ops** — multi-site promotion uses commons packs + [`scripts/enterprise/`](../../scripts/enterprise/README.md), not new core tables (see enterprise topology doc).

---

## Scope

| WS | Focus | Primary artifacts |
|----|-------|-------------------|
| **WS1** | Laptop/breadboard loop | Docs + optional `make edge-smoke-help`; dashboard shows non-404 readings |
| **WS2** | Pi field checklist | [`docs/pi-integration-guide.md`](../pi-integration-guide.md) §8, wiring annex |
| **WS3** | One safe actuator E2E | Phase 30 PR → pending_command → Pi → event; safety checklist doc |
| **WS4** | MQTT multi-zone pattern | [`docs/mqtt-edge-operator-playbook.md`](../mqtt-edge-operator-playbook.md) extension |
| **WS5** | Recipe pack demo | Sample catalog JSON + import script stub |
| **WS6** | Guardian read tools | Extend tool registry (list alerts, zone summary) — read-only |
| **WS7** | Docs cross-link | README, enterprise topology, Phase 14 index |
| **WS8** | Tests | Smokes; hardware tests tagged `hardware` skipped in CI |

---

## Work-stream detail

### WS1 — Breadboard / stub loop

**Goal:** Any developer can run `pi_client` on a laptop, post readings, see **Live Sensors** update.

**Tasks:**

- Document in [`docs/local-operator-bootstrap.md`](../local-operator-bootstrap.md): "Edge loop in 5 commands" (API up, seed, run client with stubs, verify SSE or polling).
- Confirm [`pi_client/gr33n_client.py`](../../pi_client/gr33n_client.py) stub path matches seeded sensor IDs for farm 1.
- Optional Makefile target `edge-smoke-help` (prints commands only — no new binary required for v1).

**Acceptance:** Dashboard sensor card shows a value (not NO DATA / 404) after client run.

---

### WS2 — Pi field checklist

**Goal:** Operator with a real Pi knows **exactly** what to wire first.

**Tasks:**

- Annex in [`pi-integration-guide.md`](../pi-integration-guide.md): power, relay module, **NO mains on breadboard**, `PI_API_KEY`, LAN firewall.
- Reference [`scripts/install-pi-edge-deps.sh`](../../scripts/install-pi-edge-deps.sh) + [`raspberry-pi-and-deployment-topology.md`](../raspberry-pi-and-deployment-topology.md).
- Map **one plastic room, three tiers** → three zones (naming example only).

**Acceptance:** Checklist is copy-pasteable; links to existing smokes.

---

### WS3 — Safe actuator story

**Goal:** One LED or relay proves **`pending_command`** round-trip — including from a **Phase 30 confirmed Guardian PR** when available.

**Tasks:**

- Manual test script: enqueue command from API, dashboard, **or confirmed Guardian PR** → Pi polls `GET /farms/{id}/devices` → GPIO → `POST /actuators/{id}/events` → clear pending.
- **Safety doc** (new section or [`docs/operator-troubleshooting.md`](../operator-troubleshooting.md)): fail-safe defaults, flood risk, de-energize on comms loss (operator responsibility — gr33n documents, does not enforce in software v1).

**Acceptance:** Reproduce [`TestPiContract*`](../../cmd/api/smoke_pi_contract_test.go) on a bench; audit row or actuator_events row visible.

---

### WS4 — MQTT room-scale pattern

**Goal:** Hypothetical warehouse room publishes telemetry without one HTTP POST per sensor per second from custom code.

**Tasks:**

- Topic layout example: `gr33n/farm/{farm_id}/zone/{zone_id}/sensor/{sensor_id}`.
- Bridge env vars documented; batch endpoint limits noted in playbook.

**Acceptance:** One markdown section + example env block; no broker vendor lock-in.

---

### WS5 — Recipe pack promotion demo

**Goal:** Show how **Recipe Pack v7** might propagate without a core "broadcast" feature.

**Tasks:**

- Sample `commons_catalog_entries.body` JSON (fertigation program definitions as opaque payload + readme).
- Stub script `scripts/enterprise/import-recipe-pack.sh` (two farm IDs, idempotent, calls public API).
- Cross-link [`hypothetical-enterprise-topology.md`](../hypothetical-enterprise-topology.md).

**Acceptance:** `import-recipe-pack.sh --dry-run` prints actions; real run requires local API + JWT.

---

### WS6 — Guardian read-only edge tools

**Goal:** Guardian can **answer** "what's the humidity in Flower Room?" from live snapshot + sensor latest without new GPIO paths.

**Tasks:**

- Tools: `list_unread_alerts`, `summarize_zone` (read-only) — propose only if message asks; confirm N/A.
- Reuse [`internal/farmguardian/snapshot.go`](../../internal/farmguardian/snapshot.go) + readings queries.

**Out of scope:** Direct actuator enqueue without Phase 30 PR table + confirm.

---

### WS7 — Docs

- README phase banner: Phase 31 link.
- [`phase-14-operator-documentation.md`](../phase-14-operator-documentation.md) row for enterprise topology + Phase 31.

---

### WS8 — Tests

- Smoke: after WS1 path, assert `GET /sensors/{id}/readings/latest` ≠ 404 for seeded sensor when client posted.
- Tag live-GPIO tests `@hardware` — skipped unless `GR33N_HARDWARE_TEST=1`.

---

## Out of scope (Phase 32+)

- **Conversational grow setup PRs** (plant + cycle + fertigation bundle) — see [`phase_32_guardian_grow_setup_prs.plan.md`](phase_32_guardian_grow_setup_prs.plan.md)
- Automatic multi-farm recipe broadcast in core API
- Hardware certification, UL listings, proprietary controller marketplace
- 500-site performance guarantees
- 3D-printed enclosure library (ops choice — mention in enterprise doc only)
- **Autonomous Guardian** — never; see Phase 30 principles

---

## Suggested implementation order

1. **WS1** — dashboard shows real (stub) readings (unblocks morale + demos)
2. **WS2 + WS3** — Pi checklist + one relay story (after Phase 30 actuator PR tool if testing Guardian path)
3. **WS8** — smokes for WS1/WS3
4. **WS5** — recipe pack stub (integrator story)
5. **WS4** — MQTT pattern
6. **WS6** — Guardian read tools
7. **WS7** — doc pass

Phase 29 **WS6–WS9** can run **in parallel** with WS1–WS3.

---

## Definition of done (phase ship)

- [ ] Operator doc path: laptop stub readings → dashboard live
- [ ] Pi checklist + one actuator bench test documented
- [ ] `TestPiContract*` narrative matches field checklist
- [ ] Sample recipe pack + `scripts/enterprise/` import stub (dry-run OK)
- [ ] Enterprise topology doc linked from README
- [ ] Confirmed Phase 30 actuator PR → Pi execution demonstrated on bench (when Phase 30 shipped)
- [ ] `make test` green; hardware tests opt-in only

---

## Using this plan in a new chat

```text
Implement Phase 31 per @docs/plans/phase_31_field_validation_and_edge.plan.md.

Start with WS1 (pi_client stub → dashboard readings). Read @docs/pi-integration-guide.md
and @cmd/api/smoke_pi_contract_test.go. Actuator path may use Phase 30 Guardian PR confirm.
Cross-link @docs/hypothetical-enterprise-topology.md in WS7.
```

---

## Related

| Doc | Role |
|-----|------|
| [`hypothetical-enterprise-topology.md`](../hypothetical-enterprise-topology.md) | 500-site **hypothetical** without core changes |
| [`phase_29_guardian_agent_layer.md`](phase_29_guardian_agent_layer.md) | Alert ack PRs (v1) |
| [`phase_30_guardian_change_requests.plan.md`](phase_30_guardian_change_requests.plan.md) | PR queue + config/actuator proposals |
| [`scripts/enterprise/README.md`](../../scripts/enterprise/README.md) | PR-friendly deployment script home |
