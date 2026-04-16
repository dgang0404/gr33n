# Mobile distribution (Phase 13 WS6)

The **Vue 3 PWA** in `ui/` remains the product UI. This document covers an optional **Capacitor** wrapper for store or sideloaded installs, plus push notification realities.

## When to use what

| Approach | Best for |
|----------|----------|
| **PWA** (install from browser) | Fastest rollout, no app store, same codebase as today; offline queues already supported for tasks/costs. |
| **Capacitor** | Play/App Store presence, deeper OS integration (e.g. file intents, optional push), enterprise MDM sideload. |

Capacitor does **not** replace the PWA: it hosts the same `dist/` assets in a system WebView.

## Prerequisites

- Node 20+ and npm (same as `ui/`).
- For Android: Android Studio, SDK, JDK as required by current Capacitor docs.
- For iOS: Xcode on macOS, Apple Developer account for device/TestFlight/App Store.

## One-time setup (Android example)

From repo root:

```bash
cd ui
npm install
npm run cap:add:android
```

This creates `ui/android/` (gitignored by default). Repeat with `npm run cap:add:ios` on a Mac if you need iOS.

## API URL on real devices

`localhost` is not reachable from a phone. Before `build:cap`, set a reachable API base URL.

1. Create `ui/.env.capacitor.local` (gitignored):

   ```bash
   VITE_API_URL=https://your-api.example.com
   ```

2. Vite loads `.env.capacitor` then `.env.capacitor.local` when you use `--mode capacitor`.

## Build and sync

```bash
cd ui
npm run cap:sync
```

This runs `vite build --mode capacitor` (relative `base` for WebView) and copies `dist/` into the native projects.

Open native IDEs:

```bash
npm run cap:open:android
# or
npm run cap:open:ios
```

## Store and compliance (high level)

- **Play / App Store** listings need privacy policy, data safety forms, and accurate descriptions of farm data handling.
- **Encryption export** (US): review current BIS/self-classification guidance for your distribution.
- **Per-farm data**: the app is a client; your backend privacy posture and DPA still drive compliance.

## Push notifications (FCM — server shipped)

The API can send **farm alert** push when Firebase credentials are configured. See **[`notifications-operator-playbook.md`](notifications-operator-playbook.md)** for env vars (`FCM_SERVICE_ACCOUNT_JSON` or `GOOGLE_APPLICATION_CREDENTIALS`), `/profile/push-tokens`, and `profiles.preferences.notify` volume controls.

- **FCM** (Android) and **APNs** (iOS via FCM) still require Firebase/Google Cloud project setup and Capacitor client integration to obtain tokens.
- **Web push** (PWA) is not implemented yet; `platform=web` is reserved.

Today, **sensor threshold** alerts trigger push for farm roles **owner**, **manager**, and **operator** who opt in and meet **min_priority**.

## Troubleshooting

- **White screen in WebView**: confirm `npm run build:cap` was used (not plain `build`) and `VITE_API_URL` points to HTTPS the device can reach.
- **CORS**: native apps still issue browser-like requests; ensure API CORS or same-origin strategy matches how you deploy the UI.
- **Deep links**: not configured in the scaffold; add Universal Links / App Links when you need them.

## References

- [Capacitor](https://capacitorjs.com/docs)
- Phase 13 doc index: [`phase-13-operator-documentation.md`](phase-13-operator-documentation.md)
- Phase plan: [`phase_13_platform_evolution.plan.md`](plans/phase_13_platform_evolution.plan.md)
