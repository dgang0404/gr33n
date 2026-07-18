---
name: Phase 45 WS4 — mobile sit-in path (PWA + optional Capacitor)
overview: >
  End-to-end path for farmer-sit-in-protocol.md Session C without App Store
  submission. Store signing (TestFlight / Play internal track) deferred until
  Apple/Google accounts and release assets are ready.
status: shipped
parent_plan: phase_45_farmer_validation_whole_app_polish.plan.md
---

# Phase 45 WS4 — mobile sit-in path

**Protocol:** [farmer-sit-in-protocol.md](farmer-sit-in-protocol.md) Session C · **Checklist source:** [mobile-distribution.md](../mobile-distribution.md)

## Decision

| Path | Sit-in Session C | Store release |
|------|------------------|---------------|
| **PWA** (Add to Home Screen) | ✅ **Primary** — documented below | N/A |
| **Capacitor debug / sideload** | ✅ Optional — `./scripts/cap-lan-build.sh` | ⏳ Deferred |
| **TestFlight / Play internal** | Not required for Phase 45 validation | ⏳ Deferred — needs signing assets + store accounts |

**Deferral reason (store track):** Release keystore, distribution certificates, and 1024×1024 marketing icon set are operator-specific and not committed to the repo. Follow [mobile-distribution.md § Release checklist](../mobile-distribution.md#release-checklist-b4--operator-runtime-backlog) when accounts are ready.

---

## Path A — PWA on phone (recommended for sit-in)

### Prerequisites

- Laptop running API + UI on the same LAN as the tester’s phone.
- `ui/public/icons/icon-192.png` and `icon-512.png` present (WS4 — generated from `icon.svg`).
- API `CORS_ORIGIN` matches the URL the phone uses to load the UI.

### Steps

1. **Prep URLs** (repo root):

   ```bash
   ./scripts/mobile-sit-in-prep.sh
   ```

2. **CORS** — in `.env` on the API host:

   ```bash
   CORS_ORIGIN=http://<LAN-IP>:5173
   ```

   Restart the API after editing.

3. **Start stack** — e.g. `./scripts/restart-local.sh --serve` or `make dev-auth-test`.

4. **Expose UI on LAN** (if not already):

   ```bash
   cd ui && npm run dev -- --host 0.0.0.0 --port 5173
   ```

5. **Phone** — open `http://<LAN-IP>:5173`, log in, then **Add to Home Screen** (iOS Safari) or **Install app** (Android Chrome).

6. **Session C** — run blocks from [farmer-sit-in-protocol.md § Session C](farmer-sit-in-protocol.md#session-c--mobile-webview-optional): Dashboard alert ack, Guardian Confirm/Dismiss tap targets (WS6 a11y).

7. **Log** — copy scores into [sit-in-45-session-log-template.md](sit-in-45-session-log-template.md).

### Pass criteria (WS4)

- [ ] Tester completes Session C on a physical phone without facilitator typing URLs for them after step 5.
- [ ] Confirm and Dismiss are reachable with one thumb (WS6 ~44px targets).
- [ ] No white screen after PWA cold start on the same LAN.

---

## Path B — Capacitor WebView (optional)

For facilitators who already ran `npm run cap:add:android` (or iOS on macOS):

```bash
./scripts/cap-lan-build.sh
cd ui && npm run cap:open:android
```

`cap-lan-build.sh` writes `ui/.env.capacitor.local` with `VITE_API_URL=http://<LAN-IP>:8080` and runs `npm run cap:sync`.

**CORS note:** Capacitor Android WebView may use `https://localhost` as origin — set `CORS_ORIGIN` accordingly or use a reverse proxy; see [mobile-distribution.md § Troubleshooting](../mobile-distribution.md#troubleshooting).

---

## B4 checklist status (Phase 45 execution)

| Item | WS4 status |
|------|------------|
| PWA icons 192 / 512 / maskable | ✅ `ui/public/icons/` |
| PWA sit-in path documented | ✅ This doc + `scripts/mobile-sit-in-prep.sh` |
| Capacitor LAN build script | ✅ `scripts/cap-lan-build.sh` |
| Release notes template | ✅ [mobile-distribution.md](../mobile-distribution.md) |
| Android release keystore | ⏳ Operator-owned — not in repo |
| iOS distribution cert / TestFlight | ⏳ Requires macOS + Apple Developer |
| Deep link smoke | ⏳ Optional — not configured |
| Push token smoke | ⏳ Optional — FCM client not wired in Capacitor v1 |

---

## Related

| Doc | Use |
|-----|-----|
| [operator-tour.md §10c](../operator-tour.md#10c-mobile-distribution-phase-45-ws4--shipped) | Operator summary |
| [phase_45_farmer_validation_whole_app_polish.plan.md](../plans/archive/phase_45_farmer_validation_whole_app_polish.plan.md) | Parent WS4 |
| [phase_18_platform_polish.plan.md](../plans/archive/phase_18_platform_polish.plan.md) | Mobile nav / drawer hardening |
