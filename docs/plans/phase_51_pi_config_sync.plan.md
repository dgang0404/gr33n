---
name: Phase 51 — Pi config from platform (edit + sync)
overview: >
  Close "you need SQL or hand-edited YAML to set up a Pi." The Pi client fetches its full
  wiring config (sensors, actuators, GPIO pins, intervals) from the API on startup, caches
  it locally for offline resilience, and re-applies it when the operator edits wiring in
  the UI. The local config.yaml shrinks to a minimal bootstrap (api_url + api_key +
  device_uid only). Old full-YAML setups keep working with no forced migration. Builds
  directly on the wiring model and config generator from Phase 50.
todos:
  - id: ws1-api-config-endpoint
    content: "WS1: GET /devices/by-uid/{uid}/config — return structured wiring config for one device; includes sensors[], actuators[], poll intervals, mix_channels; versioned (ETag/config_version)"
    status: completed
  - id: ws2-pi-bootstrap-rewrite
    content: "WS2: Pi client — minimal-YAML bootstrap (api_url, api_key, device_uid, farm_id only); fetch wiring from API after connect; merge with local defaults; cache to ~/.gr33n/config-cache.json"
    status: completed
  - id: ws3-live-reload
    content: "WS3: Live reload — Pi checks config_version on each schedule poll tick; if changed, reload sensors/actuators in-place without restart; UI shows 'Config pushed to Pi'"
    status: completed
  - id: ws4-offline-safety
    content: "WS4: Offline resilience + safety — if API unreachable on first start, boot from cache with log warning; if no cache and no local wiring, fail loudly; UI badge 'Pi config stale' after configurable age"
    status: completed
  - id: ws5-backward-compat
    content: "WS5: Backward compat — if sensors/actuators present in local config.yaml they take precedence (opt-out of sync); migration guide: run 'import local config to platform' helper once"
    status: completed
  - id: ws6-docs-tests
    content: "WS6: pi-integration-guide §DB-first sync, architecture §7.0p, Pi pytest coverage, Go handler tests, smoke_phase51_test.go, phase-51-closure.test.js, OC-51"
    status: completed
isProject: false
---

# Phase 51 — Pi config from platform (edit + sync)

## Status

**Shipped.** WS1–WS6 complete on `main`. Requires [Phase 50 hardware wiring visibility](phase_50_hardware_wiring_visibility.plan.md) — wiring model + PATCH API consumed by Pi sync.

**WS1 delivered:** `config_version` on `gr33ncore.devices`, bump triggers on sensor/actuator wiring, `GET /devices/by-uid/{uid}/config` + `/config/version` (Pi `X-API-Key`).

**WS2 delivered:** `load_bootstrap` / `fetch_remote_config` / `resolve_config` / `resolve_startup_config` in `pi_client/gr33n_client.py`; cache at `~/.gr33n/config-cache.json` (`CONFIG_CACHE_PATH` override); `config.bootstrap.example.yaml`; local `sensors`/`actuators` in YAML still opt out of sync.

**WS3 delivered:** `_poll_config_version` on each schedule-loop tick; `_reload_config` hot-swaps readers/actuators under `_hw_lock`; rejects empty platform wiring; reuses unchanged hardware handles when wiring keys match.

**WS4 delivered:** Cache-only boot warning; Pi PATCHes `last_config_fetch_at` on live fetch/reload; `ActuatorCard` staleness badge (`deviceConfigSync.js`); handler stores timestamp in `devices.config`.

**WS5 delivered:** Local `sensors`/`actuators` in YAML remain opt-out (unchanged installs); `import_config_to_platform.py` PATCHes wiring via JWT and writes minimal bootstrap YAML; `pi_sensor_entry_to_wiring` / `build_minimal_bootstrap` helpers in `gr33n_client.py`.

**WS6 delivered:** `pi-integration-guide` §2 platform sync + §2b legacy opt-out; architecture §7.0p; `smoke_phase51_test.go`; `phase-51-closure.test.js`; **OC-51** closed.

**Roadmap:** [farmer_ux_roadmap_40_plus.plan.md](farmer_ux_roadmap_40_plus.plan.md) (edge/Pi track).

**Closure:** **OC-51** in [phase_35_37_operational_closure.plan.md](phase_35_37_operational_closure.plan.md).

---

## Problem

After Phase 50 the platform *knows* how every sensor and actuator is wired. But the Pi still reads its wiring from a local YAML file. These two sources of truth drift:

| Source | Today |
|--------|-------|
| `pi_client/config.yaml` | Pin numbers, sources, intervals, channel map — operator hand-edits on the Pi |
| `gr33ncore.sensors.config.wiring`, `actuators.config.wiring` | Structured wiring set in the UI (Phase 50) |

The Pi doesn't know the platform updated; the platform doesn't know if the Pi ever applied the config. If an operator changes a GPIO pin in the UI, the Pi keeps running on the old YAML until someone SSHes in and edits it manually. This is the same friction as before Phase 50.

**Phase 51** closes the loop: the Pi pulls its wiring from the API, the operator's edit in the UI propagates automatically on the next poll.

---

## Current Pi startup flow (before Phase 51)

```
pi_client/gr33n_client.py
  └── load_config("config.yaml")
        ├── read DEFAULT_CONFIG (hardcoded in client)
        ├── merge local config.yaml (sensors[], actuators[], api, farm)
        └── build SensorReader / ActuatorController from that merged dict
```

The entire wiring (pin, source, channel, interval) comes from the local file. Changing it requires an SSH + file edit + service restart.

---

## Target flow (after Phase 51)

```
pi_client/gr33n_client.py
  └── load_bootstrap("config.yaml")          # api_url + api_key + device_uid + farm_id only
        └── Gr33nApiClient.fetch_device_config(device_uid)
              ├── GET /devices/by-uid/{uid}/config  (X-Api-Key)
              ├── returns {sensors:[], actuators:[], config_version, ...}
              ├── merge into runtime config
              └── write ~/.gr33n/config-cache.json   (offline resilience)
  └── build SensorReader / ActuatorController from merged runtime config

  schedule_loop (every N seconds):
    └── poll config_version → if changed, hot-reload sensors/actuators in-place
```

The `config.yaml` shrinks to 4–5 lines. GPIO/wiring lives in the platform.

---

## Design decisions

1. **Bootstrap YAML stays.** The Pi still needs `api_url`, `api_key`, and `device_uid` locally — you can't fetch from the API before you know where it is. The file becomes a **bootstrap only**, not the wiring spec.

2. **Local wiring takes precedence (opt-out).** If `sensors[]` or `actuators[]` are present in `config.yaml`, the Pi skips the platform fetch for that section. This means Phase 50 deployments and all existing Pi setups keep working unchanged; adoption is incremental.

3. **Local cache for offline resilience.** On first successful fetch, the full config JSON is written to a local cache file. On subsequent starts, if the API is unreachable, the Pi boots from cache and logs a warning. If neither API nor cache is available and local YAML has no wiring, the Pi fails loud and fast instead of running on stale defaults.

4. **Config versioning, not polling for the whole doc.** The API returns a `config_version` field (an incrementing integer on the device row). The Pi checks this on each `schedule_loop` tick without re-fetching the full payload; only on version change does it call the full endpoint. This keeps poll overhead negligible.

5. **In-place hot-reload, no restart required.** When a version change is detected, the Pi rebuilds `SensorReader` / `ActuatorController` objects and swaps them into the running daemon's maps. The sensor-loop and schedule-loop continue without a process restart. A SIGTERM-based restart fallback is provided as a CLI flag for operators who prefer explicit control.

6. **No per-Pi credentials yet.** Phase 50/51 still uses the farm-wide `PI_API_KEY` as auth for the config endpoint (same `requireAPIKey` middleware). Per-device API keys are **[Phase 57](phase_57_pi_device_api_keys.plan.md)** security hardening — see [Out of scope](#out-of-scope).

---

## WS1 — API config endpoint

New route (Go handler):

```
GET /devices/by-uid/{device_uid}/config
Auth: requireAPIKey  (X-Api-Key = PI_API_KEY)
```

Response body (JSON — same shape the Pi's `load_config` already expects for `sensors[]` and `actuators[]`):

```jsonc
{
  "device_uid": "veg-room-pi-01",
  "device_id": 1,
  "farm_id": 1,
  "config_version": 7,
  "sensors": [
    {
      "sensor_id": 3,
      "sensor_type": "temperature",
      "source": "dht22",
      "pin": 4,
      "interval_seconds": 60
    }
    // ... one entry per sensor whose wiring.device_id matches this device
  ],
  "actuators": [
    {
      "actuator_id": 1,
      "device_id": 1,
      "device_type": "light",
      "gpio_pin": 17
    }
    // ...
  ],
  "mix_channels": [1, 2],     // optional; actuator_id per channel index
  "schedule_poll_interval_seconds": 30,
  "offline_queue_path": "/var/lib/gr33n/queue.db",
  "offline_flush_interval_seconds": 60
}
```

The handler:
- Looks up `device_uid` in `gr33ncore.devices` (new index if not already present).
- Queries `gr33ncore.sensors WHERE config->'wiring'->>'device_id' = device.id AND deleted_at IS NULL`.
- Same for `gr33ncore.actuators`.
- Serializes the `wiring` sub-object fields into the flat shape the Pi client already consumes — so the client code change is minimal (merge-in, not a schema redesign).
- Returns `config_version` from a new `config_version INTEGER DEFAULT 0` column on `gr33ncore.devices` (migration; increment via DB trigger on sensor/actuator wiring PATCH).

**Lightweight version-check endpoint** (reduces chatter):

```
GET /devices/by-uid/{device_uid}/config/version
Response: {"config_version": 7}
```

Called by the Pi on every schedule poll tick (~30s). Full config only fetched on version change.

---

## WS2 — Pi client: minimal-YAML bootstrap + API fetch

### Minimal `config.yaml` (post-Phase 51)

```yaml
api:
  base_url: "http://192.168.1.100:8080"
  api_key: "replace-with-PI_API_KEY"
device:
  uid: "veg-room-pi-01"   # must match gr33ncore.devices.device_uid
farm:
  farm_id: 1
```

That is the entire file for a DB-synced Pi. No sensors or actuators sections.

### `load_config` changes

Refactor `load_config` into two phases:

1. **`load_bootstrap(path)`** — loads only `api`, `device`, `farm`, and the optional local `sensors`/`actuators` overrides. Returns a bootstrap dict. No network calls.
2. **`fetch_remote_config(client, device_uid)`** — calls `GET /devices/by-uid/{uid}/config`, parses the response, and returns a full config dict.
3. **`resolve_config(bootstrap, remote)`** — merges: local `sensors[]`/`actuators[]` win if present (opt-out); otherwise use remote. Non-wiring keys (api, farm, offline paths) always come from local bootstrap.

`Gr33nPiClient.__init__` calls these in order:
1. `load_bootstrap`
2. Try `fetch_remote_config` → if offline, load `config-cache.json` → if no cache and no local wiring, raise `RuntimeError("no wiring config available")`
3. `resolve_config` → store as `self.cfg`
4. Build `SensorReader` / `ActuatorController` as today
5. Write successful remote fetch to `config-cache.json`

### Cache location

Default: `~/.gr33n/config-cache.json` (or `CONFIG_CACHE_PATH` env override). Human-readable; operator can inspect or delete.

---

## WS3 — Live reload

### Version polling

Add `_config_version` tracking to `Gr33nPiClient`. On each `_schedule_loop` tick (after draining commands):

```python
remote_version = self.api.get_config_version(self.device_uid)
if remote_version is not None and remote_version != self._config_version:
    self._reload_config()
```

`get_config_version` calls `GET /devices/by-uid/{uid}/config/version` → returns int or None on error.

### `_reload_config()`

1. Fetch full config from `GET /devices/by-uid/{uid}/config`.
2. `resolve_config(bootstrap, remote)` → new config dict.
3. For sensors that are new or changed: instantiate new `SensorReader`; close old hardware if applicable.
4. For actuators that are new or changed: instantiate new `ActuatorController` (turn off any in-progress state first).
5. Swap `self._readers` and `self._actuators` atomically under a `threading.Lock`.
6. Update `self._config_version`.
7. Write new config to cache.
8. Log: `[config-reload] config_version=%d sensors=%d actuators=%d`.

**Safety:** if the new config has zero sensors and zero actuators (empty platform wiring), the reload is rejected and the old config stays. Log an error.

### UI signal: "Config pushed to Pi"

- When operator PATCHes wiring in the UI (Phase 50 WS2), the API bumps `config_version`.
- The Pi picks up the new version within one `schedule_poll_interval_seconds` (default 30s).
- UI can show a lightweight status: "Config sent — Pi will apply within ~30s" (optimistic, no socket needed) or a `last_config_fetch_at` timestamp surfaced on the device card (requires Pi to PATCH that on each successful fetch — one extra small field).

---

## WS4 — Offline resilience and safety

### Startup decision tree

```
load_bootstrap("config.yaml")
    └── local sensors/actuators present? → YES → use local (opt-out), skip fetch
                                          → NO  → try fetch_remote_config()
                                                     ├── success → use + cache
                                                     └── failure → try config-cache.json
                                                                      ├── cache exists → use + log WARNING("running on cached config, version may be stale")
                                                                      └── no cache → RuntimeError: "Cannot start: no wiring config. Connect to API or add sensors/actuators to config.yaml"
```

### Staleness badge in the UI

Store `last_config_fetch_at` on the device row (Pi PATCHes it on each successful fetch, reusing `PATCH /devices/{id}/status` extended body, or a new lightweight field). The device card in the UI shows:

- Green: fetched within 2× `schedule_poll_interval_seconds`
- Yellow "Config stale": fetched more than N minutes ago (configurable, default 10 min)
- Grey "Never fetched": `last_config_fetch_at` null (Pi hasn't synced yet — still on local YAML)

This is read-only in the UI; no action required from the operator unless something is broken.

---

## WS5 — Backward compatibility and migration

### Existing deployments: no forced migration

Any `config.yaml` that still includes `sensors[]` or `actuators[]` sections is treated as an **opt-out** from platform sync for those sections. The Pi logs: `"Using local wiring for sensors/actuators (platform sync disabled for this device)"`. Zero changes required to existing Pi installs to upgrade to this version of the client.

### Migration helper: "import local config to platform"

One-time import script (Python or CLI) for operators who want to adopt platform sync:

```bash
cd pi_client
python3 import_config_to_platform.py --config config.yaml --api-url http://... --api-key ...
```

For each `sensor_id` / `actuator_id` in the local YAML:
1. PATCHes `config.wiring` on the corresponding DB row (using the Phase 50 API endpoint).
2. Removes the `sensors[]` / `actuators[]` sections from `config.yaml` (or writes a new minimal YAML).
3. Prints a summary of what was imported.

After import, `config.yaml` is minimal and the Pi uses platform sync on next restart.

The script is idempotent — re-running it overwrites the same wiring fields without duplicating.

---

## WS6 — Docs, tests, closure (OC-51)

| Artifact | Content |
|----------|---------|
| [pi-integration-guide.md](../pi-integration-guide.md) | New §2 "Platform sync (Phase 51)": minimal YAML, startup flow, live reload, offline fallback; keep old §2 as "Legacy local YAML (opt-out)" |
| [farm-guardian-architecture.md](../farm-guardian-architecture.md) | §7.0p — platform ↔ Pi config sync loop diagram |
| `pi_client/test_gr33n_client.py` | `test_fetch_remote_config_success`, `test_fetch_remote_config_offline_falls_back_to_cache`, `test_reload_config_swaps_readers_atomically`, `test_local_wiring_takes_precedence` |
| `cmd/api/smoke_phase51_test.go` | Go handler: `GET /devices/by-uid/{uid}/config` returns correct wiring; version bump on PATCH |
| `ui/src/__tests__/phase-51-closure.test.js` | Closure bundle + device card staleness badge renders |
| **OC-51** | Row added to closure plan; closed when WS1–WS6 ship |

---

## Relationship to other phases

| Phase | Relationship |
|-------|-------------|
| **50** | **Hard prerequisite.** Wiring model (WS1) + API PATCH (WS2) must ship first. The Phase 51 endpoint reads the same `config.wiring` rows Phase 50 writes. |
| **44** | Device wizard generates and applies wiring; Phase 51 means the Pi picks it up without SSH. |
| **39** | Command queue contract unchanged — Phase 51 adds config sync only; command execution is untouched. |
| **[57](phase_57_pi_device_api_keys.plan.md)** | Per-device API keys (replacing shared `PI_API_KEY`) — security hardening that builds on Phase 51's per-device identity path. |

---

## Out of scope

- Per-device API keys / RBAC for edge clients — the shared `PI_API_KEY` stays this phase; see [Phase 57](phase_57_pi_device_api_keys.plan.md).
- OTA firmware or Python package updates pushed from the platform.
- Auto-discovery of connected I2C/GPIO hardware (the platform still relies on the operator to declare wiring; the Pi does not scan).
- MQTT bridge config sync — MQTT bridge (`mqtt_telemetry_bridge.py`) uses env vars, not `config.yaml`; out of scope.
- Multi-Pi coordination / load balancing (each Pi self-identifies by `device_uid`; farm-wide orchestration is a Tier D feature).
- Rollback UI (config version history) — the current `config_version` integer is enough for sync; full version history is a future audit feature.

---

## Definition of done

- [x] `GET /devices/by-uid/{uid}/config` returns structured wiring that round-trips with the Pi
- [x] `GET /devices/by-uid/{uid}/config/version` is lightweight (Pi can call it every 30s)
- [x] Pi client starts from a 5-line `config.yaml` (no sensors/actuators) and reads all wiring from API
- [x] Pi caches config; survives API restart / network blip on the next boot
- [x] Wiring change in UI → Pi hot-reloads within one `schedule_poll_interval_seconds` (~30s); no manual restart
- [x] Existing full-YAML `config.yaml` deployments work unchanged (opt-out path)
- [x] `import_config_to_platform.py` migrates a local YAML to the platform in one command
- [x] Staleness badge on device card in UI
- [x] All tests green; OC-51 closed

---

## Suggested implementation order

1. WS1 — API endpoint (the contract both sides depend on)
2. WS4 — offline resilience model (the safety net; implement alongside WS2 so tests are realistic)
3. WS2 — Pi client bootstrap rewrite + pytest coverage
4. WS5 — backward-compat check + `import_config_to_platform.py`
5. WS3 — live reload + UI staleness badge
6. WS6 — docs + smokes + closure

---

## Related

| Doc | Use |
|-----|-----|
| [phase_50_hardware_wiring_visibility.plan.md](phase_50_hardware_wiring_visibility.plan.md) | Hard prerequisite — wiring model + API PATCH |
| [pi-integration-guide.md](../pi-integration-guide.md) | Current guide this phase rewrites §2 of |
| [pi_client/gr33n_client.py](../pi_client/gr33n_client.py) | Client code: `load_config`, `Gr33nPiClient.__init__`, `_schedule_loop` |
| [pi_client/config.yaml](../pi_client/config.yaml) | Full-YAML format that becomes the opt-out / legacy path |
| [docs/field-guides/pi-wiring-basics.md](../docs/field-guides/pi-wiring-basics.md) | GPIO safety + wiring reference |

---

## Using this in a new chat

> Read `docs/plans/phase_51_pi_config_sync.plan.md` and `docs/plans/phase_50_hardware_wiring_visibility.plan.md`. Implement one workstream (WS1–WS6). Phase 50 WS1 wiring model must exist first. Do not change the `PI_API_KEY` auth contract. Existing full-YAML config.yaml deployments must keep working unchanged.
