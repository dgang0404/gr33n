---
name: Phase 165 — Farm layout API + background image
overview: >
  Backend + store plumbing for the Phase 166 visual farm canvas: persist
  user-arranged zone positions (drag layout) in zones.meta_data.layout,
  persist an optional farm background image (photo/sketch of the yard or
  floor plan) via the existing file-attachments pipeline, and expose both
  through the UI farm store. No visible UI change ships in this phase beyond
  what tests exercise; the canvas consumes it in 166.
todos:
  - id: ws1-zone-layout-metadata
    content: "WS1: zones.meta_data.layout {x,y,w,h} — validation in zone Update + PATCH round-trip"
    status: pending
  - id: ws2-farm-background
    content: "WS2: Farm background image upload/get/delete via file_attachments (farm entity)"
    status: pending
  - id: ws3-ui-store
    content: "WS3: farm store — saveZoneLayout / loadFarmLayout / background URL helpers"
    status: pending
  - id: ws4-closure
    content: "WS4: Go handler tests + store unit tests + phase-165-closure guard"
    status: pending
isProject: false
---

# Phase 165 — Farm layout API + background image

**Status:** planned · **Depends on:** none (parallel with 164) · **Feeds:** 166, 167

## Design decisions

- **Layout lives in `zones.meta_data.layout`** — not `boundary_gis`. The
  canvas is a user-arranged schematic space (normalized 0–1 coordinates), not
  GPS geometry. `boundary_gis` stays reserved for a future real-map feature.
  This mirrors the shipped `meta_data.greenhouse_climate` pattern
  (`internal/handler/zone/greenhouse.go`), including its extract/validate
  helper style.
- **Background image reuses the file-attachments pipeline** that zone photos
  already use (`POST /zones/{id}/photos` → `internal/handler/files`,
  download via `GET /file-attachments/{id}/content`). We add a farm-scoped
  equivalent and store the attachment id in farm metadata — no new storage
  system.
- **No new table.** Layout is presentational; losing it is cosmetic.

## WS1 — Zone layout in meta_data

Schema (inside `zones.meta_data`):

```json
{
  "layout": { "x": 0.12, "y": 0.40, "w": 0.22, "h": 0.18 }
}
```

- `x/y/w/h` normalized floats 0–1 relative to canvas; `w/h` optional with
  server-applied defaults; clamp/reject out-of-range on write.
- `internal/handler/zone/layout.go` — `ExtractZoneLayout` /
  `ValidateZoneLayout` mirroring `ExtractGreenhouseClimate`. Wire into the
  existing `zone.Update` meta_data path so `PUT /zones/{id}` with a
  `meta_data.layout` key validates and persists; invalid layout → 400 with a
  farmer-readable message.
- Layout must survive updates that touch *other* meta_data keys (greenhouse
  climate, photo ids) — merge, don't clobber. Confirm current Update semantics:
  if the handler replaces meta_data wholesale, add a merge helper for known
  keys; document whichever contract we land on.
- `GET /farms/{id}/zones` already returns meta_data — verify layout comes
  through the list payload (it should; `zones.sql.go` selects meta_data).

## WS2 — Farm background image

- `POST /farms/{id}/layout-background` — multipart upload, same size/type
  guards as zone photos; writes a `file_attachments` row (entity: farm) and
  records `farms.meta_data.layout_background_attachment_id` (verify farms has
  meta_data; if not, a tiny migration adds `meta_data JSONB NOT NULL DEFAULT '{}'`
  — check `db/schema/gr33n-schema-v2-FINAL.sql` first).
- `GET /farms/{id}/layout-background` — returns `{attachment_id}` or 404;
  image bytes come from the existing `GET /file-attachments/{id}/content`.
- `DELETE /farms/{id}/layout-background` — unlink + soft-delete attachment.
- Routes in `cmd/api/routes.go` beside the zone-photo block (line ~482), JWT
  chain, same handler package as zone photos.
- OpenAPI (`openapi.yaml` + `internal/openapiui/openapi.yaml`) entries.

## WS3 — UI store plumbing

`ui/src/stores/farm.js`:

- `saveZoneLayout(zoneId, layout)` — merge-write via existing `updateZone`
  (send meta_data with layout key; respect WS1 merge contract).
- `zoneLayout(zoneId)` getter reading loaded zone meta_data.
- `loadLayoutBackground(fid)` / `uploadLayoutBackground(fid, file)` /
  `clearLayoutBackground(fid)`; content URL helper for `<img>` src (auth
  token handling identical to how zone photos render today — copy that
  pattern).

No Dashboard.vue changes in this phase.

## WS4 — Closure

- Go: handler tests for layout validation (accept, clamp, reject),
  background upload/get/delete round-trip (`internal/handler/zone`,
  files handler tests).
- UI: store unit test — saveZoneLayout merge behavior, background helpers.
- `ui/src/__tests__/phase-165-closure.test.js` bundle guard.

## Acceptance criteria

1. `PUT /zones/{id}` with `meta_data.layout` persists and returns the layout;
   a subsequent greenhouse-climate update does not erase it.
2. Upload → GET → DELETE background round-trips; image renders from
   `/file-attachments/{id}/content` with a JWT.
3. `go test ./internal/handler/... -count=1` and `cd ui && npm test -- --run`
   green.

## Verification

```bash
go test ./internal/handler/zone/... ./internal/handler/... -run 'Layout|Background' -count=1
cd ui && npm test -- --run src/__tests__/phase-165-closure.test.js
```
