---
name: Phase 128 — Validate Phase 127 Guardian grounding in UI and eval
overview: >
  Automated smokes and eval fixtures for device/fertigation snapshot + new field guides;
  manual UI checklist before full regression on the laptop profile.
todos:
  - id: ws1-smoke
    content: "WS1: smoke_phase127_test.go — snapshot devices/schedule + read-tool enrichment"
    status: completed
  - id: ws2-eval-fixtures
    content: "WS2: Four eval fixtures for devices, manual programs, demo Pi, fertigation triage"
    status: completed
  - id: ws3-manual-ui
    content: "WS3: Manual UI checklist — farm context ON, phi3, four Phase 127 prompts"
    status: pending
  - id: ws4-guardian-eval
    content: "WS4: make guardian-eval on phi3:mini after UI passes (optional, slow on CPU)"
    status: pending
isProject: false
---

# Phase 128 — Validate Phase 127 grounding

## Automated (done in repo)

```bash
go test ./internal/farmguardian/... -run Snapshot -count=1
go test ./cmd/api/... -run Phase127 -count=1   # needs test DB + seed
```

## Manual UI checklist (farm context ON, phi3:mini, gr33n Demo Farm)

**Single source of truth:** run `make guardian-qa-manual` (smoke) or `make guardian-qa-manual SUITE=regression` — same prompts as automated `make guardian-qa-smoke`.

| # | Prompt | Pass if |
|---|--------|---------|
| 1 | Are any edge devices offline? | Mentions snapshot device line or `summarize_device_health`; no invented GPIO |
| 2 | Which fertigation programs are manual-only? | Names Outdoor JLF or cites schedule posture from snapshot |
| 3 | Which relay channel is the veg grow light on the demo farm? | Cites demo-farm-pi-layout or `relay_1` / Veg Relay Controller |
| 4 | Program active but no dose — what to check first? | Cites fertigation-troubleshooting (schedule, Pi, reservoir) |

## Optional slow path

```bash
make guardian-qa-smoke MODEL=phi3:mini   # 4-prompt smoke (recommended before full regression)
make guardian-qa-regression              # full fixture set
make guardian-eval                       # alias regression
```

Log tail during manual tests:

```bash
tail -f /tmp/gr33n-api.log | grep -E 'guardian:|summarize_device|summarize_zone_fertigation'
```
