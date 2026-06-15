/**
 * Phase 60 — Pi Setup Glossary
 * 
 * Structured glossary for tooltips and inline help in the Pi Setup Wizard.
 * Maps term → definition + context + related links.
 * 
 * Usage:
 * - import { glossary } from '@/lib/phase-60-glossary.js'
 * - glossary.find(term) → definition + links
 * - <TermBadge :term="'I2C'" /> → renders with hover
 */

export const PHASE_60_GLOSSARY = [
  {
    id: 'i2c',
    term: 'I²C (I-squared-C)',
    short: 'Serial protocol for connecting multiple devices on 2 wires.',
    long: `
I²C (Inter-Integrated Circuit) is a communication protocol that daisy-chains 
multiple devices on just two wires: SDA (data) and SCL (clock). All Sequent relay 
cards stack on one I²C bus, meaning you can have up to 8 relay cards on the same 
2 GPIO pins without adding more wires. Each card gets a unique address (set via 
DIP switches) so the Pi knows which one is which.

Key benefit: Scalable. Instead of 8 GPIO pins per relay card (64 pins for 8 cards), 
you use just 2 pins total.
    `.trim(),
    context: 'Used in Steps 1 & 3',
    links: [
      {
        label: 'Sequent HAT Overview',
        href: '/docs/pi-sequent-hat-setup.md#why-sequent-microsystems',
      },
      {
        label: 'I²C on Wikipedia',
        href: 'https://en.wikipedia.org/wiki/I%C2%B2C',
        external: true,
      },
    ],
  },

  {
    id: 'gpio',
    term: 'GPIO (General Purpose I/O)',
    short: 'Pins on the Raspberry Pi that turn on/off for sensors and relays.',
    long: `
GPIO stands for General Purpose Input/Output. These are the programmable pins 
on the Raspberry Pi's 40-pin header. Each pin can:
- Output: Turn on/off power to a relay or LED (output)
- Input: Read if a switch is pressed or sensor is active (input)

The Pi 4B / 5 has 26 GPIO pins available. The Sequent relay HAT uses GPIO 2 and 3 
(I²C), leaving 24 pins free for other uses (sensors, future HATs, etc.).

If you wire a single relay directly to a GPIO pin (not using the HAT), you'd specify 
which pin (e.g., GPIO 17).
    `.trim(),
    context: 'Used in Steps 1, 3 & 5',
    links: [
      {
        label: 'Pi GPIO Reference',
        href: '/docs/pi-integration-guide.md#gpio-pins',
      },
      {
        label: 'Raspberry Pi Pinout',
        href: 'https://www.raspberrypi.com/documentation/computers/os.html#gpio',
        external: true,
      },
    ],
  },

  {
    id: 'bcm-pin',
    term: 'BCM Pin Number',
    short: 'The numeric label for a GPIO pin (e.g., BCM 17, BCM 27).',
    long: `
BCM stands for Broadcom (the chip manufacturer). Raspberry Pi pins are numbered 
using Broadcom's naming scheme: GPIO 2, GPIO 3, GPIO 4, …, GPIO 27.

When you see "BCM 17" or "GPIO 17" in docs, they mean the same thing. The physical 
pin position on the 40-pin header is different—that's a separate numbering scheme 
called "physical pin" or "BOARD pin".

In gr33n, we always use BCM numbering. If you need to wire a single relay directly 
to the Pi (without the Sequent HAT), you'd pick a BCM pin and enter its number here.

Example: For a pump directly on one pin, set "BCM GPIO pin: 17"
    `.trim(),
    context: 'Used in Step 3 (direct GPIO mode)',
    links: [
      {
        label: 'Pi GPIO Pinout Diagram',
        href: 'https://www.raspberrypi.com/documentation/computers/os.html#gpio',
        external: true,
      },
    ],
  },

  {
    id: 'relay',
    term: 'Relay (Electric Switch)',
    short: 'An electromagnetic switch that turns high-power devices on/off.',
    long: `
A relay is an electric switch. When you apply power to the relay coil, the internal 
electromagnet pulls a switch contact closed, completing a circuit and turning on the 
load (pump, light, fan, solenoid valve, etc.).

On the Sequent HAT, each relay:
- Can handle up to 4 amps at 120 VAC
- Has NO (normally open) and NC (normally closed) contacts
- Is controlled by the Pi over I²C

Why use a relay instead of direct GPIO? GPIO pins can only source a few milliamps 
and work at 3.3V. A relay lets you switch high-voltage AC devices (pumps, lights, 
solenoids) safely without frying the Pi.

Example in gr33n: Main Irrigation Pump might be on Relay 1 (Channel 0).
    `.trim(),
    context: 'Used in Steps 1 & 3',
    links: [
      {
        label: 'Relay Basics',
        href: 'https://en.wikipedia.org/wiki/Relay',
        external: true,
      },
      {
        label: 'Sequent HAT Specs',
        href: '/docs/pi-sequent-hat-setup.md#parts-list',
      },
    ],
  },

  {
    id: 'channel',
    term: 'Channel (Relay Number)',
    short: 'A numbered slot for one relay; 0–63 depending on how many stacked cards.',
    long: `
A "channel" is gr33n's name for one relay position. Channels are numbered:
- 0–7: Stack level 0 (first relay card)
- 8–15: Stack level 1 (second card, if added)
- 16–23: Stack level 2 (third card, if added)
- …and so on up to 63 (8 cards max)

When you assign an actuator (pump, light, fan) in the UI, you pick a channel. 
That channel maps directly to a physical relay on your HAT stack.

Example: Channel 5 = Stack 0, Relay 6 (the 6th relay on the first card).

Why this numbering? Because it makes it easy to scale. You don't have to renumber 
things when you add a second card—just stack it and the new channels appear.
    `.trim(),
    context: 'Used throughout Steps 3 & 5',
    links: [
      {
        label: 'DIP & Channel Mapping',
        href: '/docs/pi-sequent-hat-setup.md#channel-numbering-in-gr33n',
      },
      {
        label: 'Sequent Hardware Guide',
        href: '/docs/pi-sequent-hat-setup.md',
      },
    ],
  },

  {
    id: 'stack-level',
    term: 'Stack Level',
    short: 'The physical position of a relay card (0 = closest to Pi, 1, 2, …).',
    long: `
A "stack level" is the vertical position of a relay card in the stack.
- Stack level 0: Closest to the Pi (the first card you add)
- Stack level 1: Mounted on top of level 0
- Stack level 2: Mounted on top of level 1
- …up to level 7 (8 cards total)

Each card at a different level needs different DIP switch settings so the Pi knows 
which one is which. The stack level directly determines the I²C address and the 
channel range.

Physical order doesn't matter—what matters is the DIP switch setting. So you could 
have the level 0 card physically on top, but set the DIP switch to "level 2" and 
it would respond to channels 16–23.
    `.trim(),
    context: 'Used in Steps 1 & 3',
    links: [
      {
        label: 'Stack Diagram',
        href: '/docs/pi-sequent-hat-setup.md#stack-diagram',
      },
      {
        label: 'DIP Switch Table',
        href: '/docs/pi-sequent-hat-setup.md#dip-switch-address-table',
      },
    ],
  },

  {
    id: 'dip-switch',
    term: 'DIP Switch',
    short: '3 tiny switches (ID0, ID1, ID2) that set a unique I²C address for each card.',
    long: `
DIP stands for "Dual Inline Package." On the Sequent relay HAT, you'll see 
3 tiny switches labeled ID0, ID1, ID2. Each switch can be ON or OFF.

These 3 switches form a 3-bit binary number that tells the Pi which stack level 
this card is at:

- Level 0: ID0=OFF, ID1=OFF, ID2=OFF (binary 000)
- Level 1: ID0=ON,  ID1=OFF, ID2=OFF (binary 001)
- Level 2: ID0=OFF, ID1=ON,  ID2=OFF (binary 010)
- …and so on

Every card in your stack MUST have a unique setting, or they'll fight on the I²C bus.

Before you stack the cards, set each one's DIP switches to match its intended 
position. Use the interactive calculator in the wizard to figure out which ones.
    `.trim(),
    context: 'Used in Step 3 (crucial for hardware setup)',
    links: [
      {
        label: 'DIP Switch Address Table',
        href: '/docs/pi-sequent-hat-setup.md#dip-switch-address-table',
      },
      {
        label: 'DIP Switch Calculator (in wizard)',
        href: '#ws5-dip-calculator',
      },
    ],
  },

  {
    id: 'i2c-address',
    term: 'I²C Address',
    short: 'A unique hexadecimal code (0x20–0x27) that identifies each relay card.',
    long: `
Every I²C device on the bus needs a unique address so the Pi knows which device 
it's talking to. For the Sequent relay HAT:

- Stack level 0: I²C address 0x27
- Stack level 1: I²C address 0x26
- Stack level 2: I²C address 0x25
- …and so on down to 0x20 for level 7

The address is determined by your DIP switch settings. You don't need to remember 
these numbers—the wizard's DIP calculator shows you what to set.
    `.trim(),
    context: 'Used in Step 3 (reference info)',
    links: [
      {
        label: 'DIP Switch Table',
        href: '/docs/pi-sequent-hat-setup.md#dip-switch-address-table',
      },
    ],
  },

  {
    id: 'api-key',
    term: 'API Key (gdev_*)',
    short: 'A secret token that proves your Pi is trusted to send data to the API.',
    long: `
An API key is like a password. It proves to the gr33n API server that your Pi 
is allowed to:
- Post sensor readings
- Receive pump/light commands
- Store data

gr33n uses device-specific keys starting with "gdev_" followed by a random secret:
  gdev_demo-flower-01_xK9mL2p5Q8wNaB1cR3vJ7dT

⚠️ SECURITY: This is a secret. Treat it like a password:
- Copy it NOW from Step 2 (you won't see it again)
- Don't share it in chat, email, or commit to version control
- If leaked, rotate it immediately (via API key management)

The Pi sends this key in every API request (X-Device-Key header) so the server 
knows it's you.
    `.trim(),
    context: 'Used in Steps 2 & 5 (critical for security)',
    links: [
      {
        label: 'Security Guide',
        href: '/docs/SECURITY.md',
      },
      {
        label: 'API Authentication',
        href: '/docs/pi-integration-guide.md#api-auth',
      },
    ],
  },

  {
    id: 'device-uid',
    term: 'Device UID (Unique ID)',
    short: 'A human-readable name that identifies this specific Pi on your farm.',
    long: `
The Device UID is a unique identifier for your Pi. It's not the IP address or 
hostname—it's a label you create:

Examples:
- demo-flower-01
- greenhouse-room-3
- rpi-nutrient-pump
- flower-zone-intake

Use something descriptive so you can tell your Pis apart. You can use:
- Hostname (if you set one on the Pi)
- MAC address (last 6 characters)
- Custom name (anything you want)

The UID is used when the Pi talks to the API. It's also what you see in the 
Settings → Devices list in gr33n.

Note: This is NOT a secret—it's just a label. The actual security comes from 
the API key.
    `.trim(),
    context: 'Used in Steps 2 & 5',
    links: [
      {
        label: 'Device Registration Guide',
        href: '/docs/pi-integration-guide.md#device-registration',
      },
    ],
  },

  {
    id: 'offline-queue',
    term: 'Offline Queue',
    short: 'Local storage on the Pi for sensor readings when the network is down.',
    long: `
If your Pi loses connection to the API (network outage, server down, etc.), 
it doesn't just give up. Instead, it saves readings to a local SQLite database:
  /var/lib/gr33n/queue.db

When the network comes back, the Pi flushes the queue and uploads all the missed 
readings. This keeps your data safe even if the farm is disconnected for hours 
or days.

You can see pending writes in the dashboard ("X readings queued"). The Pi tries to 
flush every 60 seconds (configurable), so data syncs quickly once the connection 
returns.
    `.trim(),
    context: 'Used in Step 5 (config preview)',
    links: [
      {
        label: 'Offline Resilience',
        href: '/docs/pi-integration-guide.md#offline-resilience',
      },
      {
        label: 'Edge Loop Guide',
        href: '/docs/local-operator-bootstrap.md#edge-loop-in-5-commands-phase-31-ws1',
      },
    ],
  },

  {
    id: 'config-yaml',
    term: 'config.yaml (Bootstrap Config)',
    short: 'A simple text file that tells the Pi where the API is and what farm it belongs to.',
    long: `
config.yaml is the Pi's startup configuration. It's a YAML text file (plain text, 
not a binary file) that lives on the Pi at:
  ~/gr33n/pi_client/config.yaml

In Phase 51 (platform sync), it's minimal:
\`\`\`yaml
api:
  base_url: http://192.168.1.50:8080
  api_key: gdev_demo-flower-01_xK9mL2p5Q8wNaB1cR3vJ7dT
device:
  uid: demo-flower-01
farm:
  farm_id: 1
\`\`\`

That's it. The Pi fetches all other wiring (sensors, actuators, GPIO pins) from the 
API on startup. If you edit actuator/sensor wiring in the UI, the Pi reloads it 
automatically (no manual restart needed).

Note: In Phase 60, you don't edit this file by hand. The wizard generates it for you.
    `.trim(),
    context: 'Used in Step 5 (download & deploy)',
    links: [
      {
        label: 'Platform Sync (Phase 51)',
        href: '/docs/pi-integration-guide.md#2-platform-sync-phase-51--recommended',
      },
      {
        label: 'Bootstrap Example',
        href: '/docs/pi-integration-guide.md#minimal-bootstrapconfig-bootstrap-example-yaml',
      },
    ],
  },

  {
    id: 'network-test',
    term: 'Network Test',
    short: 'A connectivity check to prove the Pi can reach the API server.',
    long: `
The network test is a quick diagnostic:
1. You click "Test now" in Step 4
2. The wizard sends a request to the API
3. The API generates a unique "ping" URL
4. Your Pi (running in the background) calls that URL back
5. The wizard shows: ✓ Connected, latency, or ✗ Failed

This proves that:
- Your Pi can access the network (WiFi/Ethernet is working)
- DNS resolution works (Pi can find the API server by name)
- No firewalls are blocking the connection
- The network latency is acceptable (< 1s is good)

If the test fails, you get friendly error messages:
- "DNS failed" → check the server URL
- "Connection refused" → firewall or wrong port
- "Timeout" → unreachable or very slow network
    `.trim(),
    context: 'Used in Step 4 (connectivity verification)',
    links: [
      {
        label: 'Troubleshooting Network Issues',
        href: '/docs/operator-troubleshooting.md#5-edge-actuator-safety-phase-31-ws3',
      },
    ],
  },

  {
    id: 'farm-id',
    term: 'Farm ID',
    short: 'A number that identifies which farm on the gr33n instance this Pi belongs to.',
    long: `
A Farm ID is a numeric identifier for a farm (yours). If you run gr33n, you might 
manage multiple farms (Farm 1, Farm 2, Farm 3). Each farm is separate:
- Different zones
- Different plants
- Different automation rules
- Different Pis and sensors

Your Pi's config.yaml includes its farm_id so it knows which farm's data to post to:
\`\`\`yaml
farm:
  farm_id: 1
\`\`\`

In the UI, you can only see one farm at a time. Switch farms via the farm selector 
(top-left). Your Pi always posts to its assigned farm_id.
    `.trim(),
    context: 'Used in Steps 2 & 5 (config setup)',
    links: [
      {
        label: 'Multi-farm Architecture',
        href: '/docs/ARCHITECTURE.md#farms-and-zones',
      },
    ],
  },

  {
    id: 'actuator',
    term: 'Actuator',
    short: 'A controllable device: pump, light, fan, solenoid valve, etc.',
    long: `
An actuator is a device that *does* something (as opposed to a sensor that measures 
something).

Common actuators on a farm:
- Main irrigation pump (supplies water)
- Nutrient dosing pump A (nutrients)
- Grow lights (light)
- Exhaust fan (air circulation)
- Solenoid valve (drain)

In gr33n, each actuator is wired to one relay channel. When the automation system 
decides it's time to water, it sends a command to that channel, which turns on the 
pump.

You can manually override any actuator in the UI (Controls → toggle), or let 
automation run them on a schedule.
    `.trim(),
    context: 'Used throughout (Steps 2, 3, 5)',
    links: [
      {
        label: 'Controls (Actuators)',
        href: '/docs/operator-tour.md#3-controls--actuators',
      },
      {
        label: 'Automation & Actuators',
        href: '/docs/workflow-guide.md#actuators',
      },
    ],
  },

  {
    id: 'sensor',
    term: 'Sensor',
    short: 'A device that measures environmental conditions: temperature, humidity, moisture, light, etc.',
    long: `
A sensor is a device that measures something. Unlike an actuator (which does 
something), a sensor just reads and reports.

Common sensors on a farm:
- Temperature (air, water)
- Humidity (air)
- Light (PAR, lux)
- Soil moisture
- pH (water, soil)
- EC (electrical conductivity, salts)
- CO₂

The Pi reads from sensors on a regular schedule (every 5 minutes by default) and 
posts the readings to the API. You see them in real-time on the Dashboard under 
"Live Sensors".

Sensors are wired directly to GPIO pins (or via I²C for advanced sensors). Unlike 
relays (which are discrete on/off), sensors are usually analog (gradual) or digital 
(precise readings).
    `.trim(),
    context: 'Used in Steps 1 & 3',
    links: [
      {
        label: 'Live Sensors in Dashboard',
        href: '/docs/operator-tour.md#1-dashboard--today',
      },
      {
        label: 'Sensor Data Model',
        href: '/docs/database-schema-overview.md#sensors-and-readings',
      },
    ],
  },
]

/**
 * Helper functions
 */

export function findTerm(id) {
  return PHASE_60_GLOSSARY.find(entry => entry.id === id)
}

export function searchTerms(query) {
  const q = query.toLowerCase()
  return PHASE_60_GLOSSARY.filter(
    entry =>
      entry.term.toLowerCase().includes(q) ||
      entry.short.toLowerCase().includes(q) ||
      entry.long.toLowerCase().includes(q)
  )
}

export function getTermsByContext(context) {
  return PHASE_60_GLOSSARY.filter(entry => entry.context.includes(context))
}
