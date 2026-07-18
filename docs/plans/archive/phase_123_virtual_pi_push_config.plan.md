---
name: Phase 123 — Virtual Pi (WS4): notify Pi to reload + docs closure
overview: >
  Phase 121 shipped download + drift detection but platform-sync Pis already poll
  config_version — operators needed an explicit "tell the Pi to reload now" after
  hand-editing wiring or replacing a YAML file. This phase adds a one-click bump
  plus closes the 119–122 doc index.
todos:
  - id: ws1-push-config-api
    content: "WS1: POST /devices/{id}/push-config — bump config_version; farm operate auth; returns new version + operator message"
    status: completed
  - id: ws2-virtual-pi-button
    content: "WS2: Virtual Pi 'Notify Pi to reload' button when device uses platform sync (device_uid + config_version > 0)"
    status: completed
  - id: ws3-docs-closure
    content: "WS3: CHANGELOG + README + phase-14 index rows for Phases 119–122; connectivity doc Virtual Pi link"
    status: completed
  - id: ws4-tests
    content: "WS4: Go handler test + Vitest route/UI presence; openapi push-config path"
    status: completed
isProject: false
---

# Phase 123 — Virtual Pi: notify Pi to reload + docs closure

**Status: shipped**

## Why

Platform-sync Pis reload wiring when `config_version` changes (Phase 51). Wiring
PATCHes bump automatically, but after downloading a fresh `config.yaml` or when
the Pi missed a poll, operators had no explicit nudge. This adds `POST
/devices/{id}/push-config` and a Virtual Pi button.

## Out of scope (future)

- Auto-push on every wiring save (already via DB triggers)
- Streaming config body over pending_command
- Multi-Pi fleet orchestration dashboard

## Acceptance

- [x] POST push-config increments config_version and returns 200
- [x] Virtual Pi shows button only for platform-sync devices
- [x] Docs index lists Phases 119–122 as shipped
- [x] Tests cover route registration and UI button
