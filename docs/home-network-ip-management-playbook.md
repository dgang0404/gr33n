# Home network / IP management playbook

A from-scratch home lab running gr33n typically grows from "one laptop running
everything" into several always-on boxes (API+DB, maybe a split-out DB, an
LLM/Guardian server) plus several Pi edge devices. This page is the one
checklist for keeping that reachable without babysitting it — no new gr33n
code required for any of it.

Related: [connectivity-requirements.md](connectivity-requirements.md) (what
needs a connection and to what), [operator-troubleshooting.md](operator-troubleshooting.md#3b-device-went-dark--check-ip-history-before-psql-phase-214)
(diagnosing a dark Pi).

## Why this matters

Every device on the network — servers and Pis alike — gets its IP address
from the router via **DHCP**, and that address can change on reconnect
(reboot, power blip, lease renewal). Two very different consequences follow:

- **A Pi's own IP changes** — a non-event. The Pi always calls the API (never
  the other way around — see `internal/handler/devicecmd`), so its own
  address doesn't matter for connectivity. `devices.ip_address` self-corrects
  on the Pi's next heartbeat automatically (Phase 213/214); the manual
  override (`PATCH /devices/{id}/ip-address`, Phase 215) is a bookkeeping aid,
  not a reconnection mechanism.
- **A server's IP changes** (the box running the API, DB, or a Pi's
  hardcoded `base_url` target) — this *does* break things, and nothing in
  gr33n's database can fix it remotely, because the device that needs to
  reach the new address can't ask for it without already being able to reach
  the old one. This is solved at the network layer, not the app layer —
  see below.

## Layer 1 — DHCP reservations for every server

In your router's DHCP settings, bind each **server's** MAC address to a fixed
IP ("DHCP reservation" / "static lease" / "address reservation" depending on
brand). This is not the same as configuring a static IP on the machine
itself — the router still hands it out via DHCP, just always the same one.

- **Do this for**: API server, DB server, LLM/Guardian server — anything a
  Pi or your browser needs to find reliably.
- **Optional for**: the Pis themselves — only useful for your own SSH
  convenience, not required for reconnection.

Reservations survive reboots and lease renewals by design, which covers most
of "why did it die while I was on vacation."

## Layer 2 — Names instead of raw IPs in every config

Once you have 3+ boxes, hardcoded IPs anywhere (a Pi's `api.base_url` in
`pi_client/config.yaml`, a DB connection string) are exactly the fragility
above. Two options, cheapest first:

1. **mDNS `.local` hostnames** — zero extra infrastructure. Raspberry Pi OS
   ships `avahi-daemon` by default (`raspberrypi.local` works out of the
   box); Linux boxes can run avahi too. Rename each box (`gr33n-api.local`,
   `gr33n-db.local`, `gr33n-llm.local`) and point every config at the
   hostname instead of the IP. Limits: doesn't cross subnets/VLANs, and
   Windows clients need Bonjour installed.
2. **A local DNS resolver** (Pi-hole or plain `dnsmasq`) — one place that
   defines every hostname, once mDNS starts feeling scattered. Pi-hole also
   gives ad-blocking as a bonus.

Apply it everywhere: DB connection string uses the DB's hostname; every Pi's
`config.yaml` `base_url` uses the API's hostname, not its IP.

If you eventually run more than one API instance, add a lightweight reverse
proxy (Caddy is the easiest — automatic HTTPS too) as one stable front door.
Not needed with a single API box.

## Layer 3 — Monitor from outside gr33n

Don't build alerting into gr33n's own code — if gr33n's process is the thing
that died, gr33n can't be the one telling you. Use an existing self-hosted
monitor instead:

- **Uptime Kuma** (free, one Docker container) already speaks ntfy,
  Pushover, Telegram, email, Discord, and webhooks — no Firebase project,
  no code changes.
- Point a monitor at **`GET /health`** ([routes.go](../cmd/api/routes.go)) —
  public, no auth, pings the DB and returns 503 if it can't connect.
- Optionally add a second monitor on `GET /farms/{id}/alerts` to catch a
  specific device going dark, not just the whole API.

## The one gap reservations don't cover

If the router itself dies or factory-resets, it loses its reservation table.
Mitigate by periodically exporting/backing up the router config, and by
keeping the network map below somewhere durable — not just in the router's
memory.

## Network map

Keep a running record as the lab grows. Fill in as each box is provisioned:

| Device | Role | MAC | Reserved IP | Hostname |
|--------|------|-----|-------------|----------|
| _(laptop, dev)_ | API + DB (dev) | | | |
| | API + DB (prod) | | | |
| | LLM / Guardian | | | |
| | Pi — Veg Relay Controller | | | |
