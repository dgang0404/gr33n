---
name: Phase 127 — Snapshot device/fertigation hints + curated field guides
overview: >
  Morning walkthrough and grounded Guardian get compact edge-device and fertigation-schedule
  lines in the live snapshot (cycles alone do not cover Pi or feed posture). Add reviewed
  field guides for fertigation troubleshooting and gr33n Demo Farm Pi layout; expand
  field-troubleshooting symptom table.
todos:
  - id: ws1-field-guides
    content: "WS1: fertigation-troubleshooting.md, demo-farm-pi-layout.md, field-troubleshooting rows + DB upsert"
    status: completed
  - id: ws2-snapshot
    content: "WS2: BuildSnapshot device counts + fertigation schedule summary in Render()"
    status: completed
  - id: ws3-ingest-docs
    content: "WS3: field-guide-manifest + README + local-operator-bootstrap ingest note"
    status: completed
isProject: false
---

# Phase 127 — Snapshot device/fertigation hints + curated field guides

**Status: shipped (local)**

## Problem

- Live snapshot listed cycle **stage** and fertigation program **names** but not **Pi online/offline** or **which active programs lack a schedule**.
- Field guides covered Pi wiring generically but not the **demo farm device map** or **fertigation failure triage**.

## Acceptance

- [x] `BuildSnapshot` renders edge device online/offline counts (+ offline names when present)
- [x] `BuildSnapshot` renders scheduled vs manual-only active program counts
- [x] New guides in `docs/field-guides/` and `agronomy_field_guides` via migration
- [x] `make migrate` then `make rag-ingest-field-guides` picks up new bodies

## Operator steps after pull

```bash
make migrate
make rag-ingest-field-guides   # needs EMBEDDING_API_KEY in .env
```

Test grounded questions (farm context **on**):

- "Are any Pis offline on this farm?"
- "Why didn't the veg fertigation program run?"
- "Which GPIO is the veg grow light relay on the demo farm?"
