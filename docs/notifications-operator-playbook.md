# Farm alert push (FCM) — operator playbook

This document covers **Phase 14 WS5**: server-side hooks from existing **in-app alerts** (`gr33ncore.alerts_notifications`) to **Firebase Cloud Messaging (FCM)** for native apps (Capacitor on Android/iOS). **Web push** is not implemented; `platform=web` is reserved for a future PWA path.

## Behavior

1. **Threshold breaches** on sensors create alerts (as before). After a successful insert, the API dispatches push **asynchronously** when FCM is configured.
2. **Recipients** are users with farm membership role **owner**, **manager**, or **operator** who have:
   - `profiles.preferences.notify.push_enabled === true`, and
   - alert **severity** at or above `profiles.preferences.notify.min_priority` (`low` &lt; `medium` &lt; `high` &lt; `critical`).
3. **Tokens** live in `gr33ncore.user_push_tokens`. The same `fcm_token` is **unique**; re-registering assigns it to the current user (device handoff).
4. If FCM reports an **unregistered** or **invalid-argument** token, that row is **deleted** automatically.

## API (JWT)

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/profile/notification-preferences` | Read effective `push_enabled` + `min_priority` |
| `PATCH` | `/profile/notification-preferences` | Merge notify prefs into `profiles.preferences` |
| `GET` | `/profile/push-tokens` | List the caller’s device tokens |
| `POST` | `/profile/push-tokens` | Body: `platform` (`android` \| `ios` \| `web`), `fcm_token` |
| `DELETE` | `/profile/push-tokens` | Body: `fcm_token` |

OpenAPI: [`openapi.yaml`](../openapi.yaml). Migration: `db/migrations/20260427_user_push_tokens.sql`.

## Server configuration

The API enables FCM when **one** of the following is set:

- **`FCM_SERVICE_ACCOUNT_JSON`** — full JSON of a Google service account key (typical for containers).
- **`GOOGLE_APPLICATION_CREDENTIALS`** — path to a service account JSON file on disk.

If neither is set, push dispatch is a **no-op** (alerts and in-app feed unchanged).

Use a Firebase/Google Cloud service account that can call **Firebase Cloud Messaging**. Point the mobile app at the same Firebase project so registration tokens are valid for that project.

## Mobile (Capacitor)

- Obtain an FCM token via the Capacitor / Firebase plugin for the platform.
- On login (or app resume), call `POST /profile/push-tokens` with the token.
- On logout, call `DELETE /profile/push-tokens` for that token.
- Encourage operators to set **`min_priority`** (e.g. `high`) to limit noise.

See also: [`mobile-distribution.md`](mobile-distribution.md).

## Payload

Notifications include a display **title** and **body** from the alert, plus **data** fields:

- `farm_id`, `alert_id`, `kind=farm_alert`

Clients can use these for deep links into the farm alert UI.
