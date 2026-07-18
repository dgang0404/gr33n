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
    status: completed
  - id: ws4-guardian-eval
    content: "WS4: make guardian-qa-phase127 on phi3:mini after UI passes (optional, slow on CPU)"
    status: completed
isProject: false
---

# Phase 128 — Validate Phase 127 grounding

**Status:** complete

## Automated (repo)

```bash
go test ./internal/farmguardian/... -run Snapshot -count=1
go test ./cmd/api/... -run Phase127 -count=1   # needs test DB + seed
go test ./cmd/api/... -run Phase128 -count=1   # suite wiring + score heuristics
go test ./internal/farmguardian/eval/... -run Phase127 -count=1
```

## Manual UI checklist (farm context ON, phi3:mini, gr33n Demo Farm)

**Single source of truth:**

```bash
make guardian-qa-manual SUITE=phase127
```

| # | Prompt | Pass if |
|---|--------|---------|
| 1 | Are any edge devices offline? | Mentions snapshot device line or `summarize_device_health`; no invented GPIO |
| 2 | Which fertigation programs are manual-only? | Names Outdoor JLF or cites schedule posture from snapshot |
| 3 | Which relay channel is the veg grow light on the demo farm? | Cites demo-farm-pi-layout or `relay_1` / Veg Relay Controller |
| 4 | Program active but no dose — what to check first? | Cites fertigation-troubleshooting (schedule, Pi, reservoir) |

## Automated slow path (after manual passes)

```bash
export GUARDIAN_EVAL_TOKEN="<jwt>"
export GUARDIAN_EVAL_LOG=/tmp/gr33n-api.log
make guardian-qa-phase127 MODEL=phi3:mini FARM_ID=1
# Archives: data/guardian_qa_runs/<timestamp>_phase127_phi3-mini.json

# Broader regression (optional):
make guardian-qa-regression MODEL=phi3:mini
```

Log tail during tests:

```bash
tail -f /tmp/gr33n-api.log | grep -E 'guardian:|summarize_device|summarize_zone_fertigation'
./scripts/guardian-qa-scrape-logs.sh --expect summarize_device_health
```
