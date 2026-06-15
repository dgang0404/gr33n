# Phase 60 — Pi Setup Wizard & UX Design Guide

> **Design companion to** [`phase_60_pi_setup_wizard_ux.plan.md`](phase_60_pi_setup_wizard_ux.plan.md)  
> **For:** Frontend designers / engineers implementing the wizard UI.

---

## Overview

The wizard walks a farmer (zero IT background) from **no Pi** to **live, wired, and ready to pump water**. Every choice is validated, explained, and visual. No terminal, no YAML, no surprises.

---

## User Flow Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│  Settings → Devices → "Set up new Pi"  OR  Hardware → New setup │
└──────────────────────────────┬──────────────────────────────────┘
                               ↓
                    ┌──────────────────────┐
                    │    STEP 1: Welcome   │
                    │  Hardware checklist  │
                    └──────────┬───────────┘
                               ↓
              ┌────────────────────────────────────┐
              │    STEP 2: Register Pi on Farm     │
              │  Name, UID, API key generation     │
              └────────────────┬───────────────────┘
                               ↓
            ┌──────────────────────────────────────────┐
            │  STEP 3: Assign Relay Channels & Wiring  │
            │  Visual grid + actuator dropdown + DIP   │
            └──────────────────┬─────────────────────┘
                               ↓
           ┌───────────────────────────────────────────┐
           │  STEP 4: Network & API Configuration      │
           │  Base URL + network test (connectivity)   │
           └─────────────────┬─────────────────────────┘
                             ↓
          ┌─────────────────────────────────────────────┐
          │  STEP 5: Download Config & Instructions     │
          │  YAML file + copy-to-clipboard + one-liner  │
          └────────────────┬────────────────────────────┘
                           ↓
        ┌─────────────────────────────────────────────────┐
        │  STEP 6: Verify & Complete                       │
        │  Deployment checklist + actuator pulse test      │
        └────────────────┬────────────────────────────────┘
                         ↓
                    ✅ Dashboard / Hardware board
```

---

## Layout & Styling

**Container:**
- Max-width: 800px (comfort for reading)
- Centered, with responsive padding
- Dark theme (zinc-900 / zinc-800 background, text-zinc-200)
- Same design language as current UI

**Progress bar:**
- Horizontal 6-step progress indicator at the top
- Show current step (e.g., "3 of 6: Relay Channels")
- Color: green-500 for completed steps, zinc-500 for pending

**Buttons:**
- "← Back" (left, secondary style)
- "Next →" (right, primary green style, disabled if validation fails)
- "Cancel" (always available, top-right)

**Content area:**
- Main panel (left): 60% — form inputs, visuals
- Help/glossary panel (right): 40% — tooltips, definitions, examples
- On mobile: stack vertically, 100%

---

## Step-by-Step UI Details

### Step 1: Welcome & Hardware Checklist

**Left panel (60%):**

```
┌─────────────────────────────────────────────────────────────┐
│                                                 [Step 1/6]   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  🏠 Welcome to the Pi Setup Wizard                          │
│                                                             │
│  Let's wire up a grow room with one Raspberry Pi + relays.  │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  📋 Before you start — do you have:                         │
│                                                             │
│    ☐ Raspberry Pi 4B or 5 (or Pi Zero 2W)                  │
│      └─ 💡 TIP: Pi 5 recommended; works even on Pi Zero   │
│                                                             │
│    ☐ Sequent Microsystems 8-Relay HAT                       │
│      └─ 💡 Available at: sequentmicrosystems.com/products  │
│                                                             │
│    ☐ 5V / 8A power supply (Pi + HAT)                        │
│                                                             │
│    ☐ 14–16 AWG stranded wire for relays (pumps, lights)    │
│                                                             │
│    ☐ Ferrule crimping kit for connectors                    │
│                                                             │
│    ☐ Ethernet or WiFi connection for Pi                     │
│      └─ 💡 Connected to same network as this computer       │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  Unsure about any of these? Watch the intro video ↗         │
│  (video placeholder, can be added in Phase 61)              │
│                                                             │
│                                           [ Cancel ] Next →  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Right panel (40%):**
```
┌──────────────────────────────────┐
│  ℹ️  What is a Relay HAT?          │
├──────────────────────────────────┤
│                                  │
│  A "HAT" is a hardware add-on    │
│  for the Raspberry Pi. The       │
│  Sequent relay HAT stacks on     │
│  top and adds 8 controllable     │
│  relays (switches) for pumps,    │
│  lights, and fans.               │
│                                  │
│  Key benefits:                   │
│  • Up to 64 relays per Pi        │
│  • Industrial-grade connectors   │
│  • I²C protocol (simple)          │
│  • Stackable (add as you grow)   │
│                                  │
│  [Learn more] →                   │
│  [Parts list] →                   │
│                                  │
└──────────────────────────────────┘
```

---

### Step 2: Register Pi on Farm

**Left panel (60%):**

```
┌─────────────────────────────────────────────────────────────┐
│                                                 [Step 2/6]   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  📱 Register your Pi on the Farm                            │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  Pi Name / Label                                            │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Flower Room Pi                         [?]          │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  Device UID (unique identifier on network)                 │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ demo-flower-01                        [?] [↻ Auto]  │   │
│  └─────────────────────────────────────────────────────┘   │
│     Auto-detected: rpi4-2a:3b:4c on 192.168.1.42           │
│                                                             │
│  Zone assignment (optional — set later)                    │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ Select zone or create new...           [?]          │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  🔐 API Key (for secure communication)                     │
│                                                             │
│  ⚠️  IMPORTANT: Copy this NOW and save it somewhere safe.  │
│     You won't see it again!                                │
│                                                             │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ gdev_demo-flower-01_xK9mL2p5Q8wNaB1cR3vJ7dT       │   │
│  │  [Copy to clipboard] [Download as file]             │   │
│  └─────────────────────────────────────────────────────┘   │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  ✓ Pi registered                                           │
│                                                             │
│                                           [ ← Back ] Next → │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Right panel (40%):**
```
┌──────────────────────────────────┐
│  💾 About Device UID               │
├──────────────────────────────────┤
│                                  │
│  The UID uniquely identifies     │
│  this Pi on your network.        │
│                                  │
│  You can use:                    │
│  • Hostname (demo-flower-01)     │
│  • MAC address (last 6 chars)    │
│  • Custom name                   │
│                                  │
│  It's just for labeling—pick    │
│  whatever helps you remember.   │
│                                  │
│  🔐 About the API Key             │
│  ───────────────────────────────  │
│  This token proves the Pi is     │
│  trusted. It's like a password:  │
│  • Keep it private               │
│  • Don't share in chat/code      │
│  • You can rotate it later       │
│                                  │
│  [View security guide] →          │
│                                  │
└──────────────────────────────────┘
```

---

### Step 3: Assign Relay Channels & Wiring

**Left panel (60%):**

```
┌─────────────────────────────────────────────────────────────┐
│                                                 [Step 3/6]   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ⚙️  Assign Relay Channels                                  │
│                                                             │
│  What pump or fan goes on each relay?                       │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  STACK 0 (relays 1–8)  [Add more stacks ↓]                 │
│                                                             │
│  ┌─ Relay 1 ──────────────────────────────────────┐        │
│  │ Channel: 0                   [?]                │        │
│  │ ┌──────────────────────────────────────────┐  │        │
│  │ │ Select actuator...                   ▼ │  │        │
│  │ │                                         │  │        │
│  │ │ ☑ Main Irrigation Pump       (pump)  │  │        │
│  │ │   Demo Plant Drain Pump     (pump)   │  │        │
│  │ │   Nutrient Dose A           (pump)   │  │        │
│  │ │   Grow Lights              (lights)  │  │        │
│  │ │   Exhaust Fan              (fan)    │  │        │
│  │ │   Spare (not assigned yet)           │  │        │
│  │ └──────────────────────────────────────────┘  │        │
│  │ ✓ Main Irrigation Pump assigned                │        │
│  └────────────────────────────────────────────────┘        │
│                                                             │
│  ┌─ Relay 2 ──────────────────────────────────────┐        │
│  │ Channel: 1                   [?]                │        │
│  │ ┌──────────────────────────────────────────┐  │        │
│  │ │ Select actuator...                   ▼ │  │        │
│  │ └──────────────────────────────────────────┘  │        │
│  │ ⚠️  No actuator selected yet                   │        │
│  └────────────────────────────────────────────────┘        │
│                                                             │
│  [ Channels 3–8 scrollable below ]                         │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  ✓ 1 actuator assigned  ⚠️  7 channels empty                │
│  ⚠️  Some relays are unused — that's OK                    │
│                                                             │
│                                           [ ← Back ] Next → │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Right panel (40%):**
```
┌──────────────────────────────────┐
│  💡 What is a Channel?            │
├──────────────────────────────────┤
│                                  │
│  A channel is a relay (switch).  │
│  Each relay controls one thing:  │
│  pump, light, fan, etc.          │
│                                  │
│  Numbering:                       │
│  • Stack 0: ch 0–7               │
│  • Stack 1: ch 8–15              │
│  • Stack 2: ch 16–23             │
│                                  │
│  [Learn more] →                   │
│                                  │
│  🔧 DIP Switch Help               │
│  ───────────────────────────────  │
│  You need to set physical DIP     │
│  switches on the relay card.      │
│                                  │
│  For Stack 0:                     │
│  ID0 = OFF, ID1 = OFF, ID2 = OFF │
│                                  │
│  [Use DIP calculator] →            │
│                                  │
└──────────────────────────────────┘
```

**DIP Calculator modal (on-click from right panel):**

```
┌─────────────────────────────────────────┐
│  🔲 DIP Switch Calculator                │
├─────────────────────────────────────────┤
│                                         │
│  Want to use channel: [20      ]        │
│                                         │
│  ────────────────────────────────────  │
│                                         │
│  That means:                            │
│  • Stack level: 2                       │
│  • Position: Relay 5 on card 3          │
│  • I²C address: 0x25                    │
│                                         │
│  ────────────────────────────────────  │
│                                         │
│  Set these DIP switches on the card:    │
│                                         │
│    ID0: [🟩 ON ]  ID1: [🟧 OFF]  ID2: [🟩 ON ]  │
│                                         │
│  🖼️ Diagram (visual position help)       │
│  [───────────────────────────────────]  │
│                                         │
│  [Close]                                │
│                                         │
└─────────────────────────────────────────┘
```

---

### Step 4: Network & API Configuration

**Left panel (60%):**

```
┌─────────────────────────────────────────────────────────────┐
│                                                 [Step 4/6]   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  🌐 Network & API Setup                                     │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  API Server Address                                         │
│  ┌─────────────────────────────────────────────────────┐   │
│  │ http://192.168.1.50:8080              [?]          │   │
│  └─────────────────────────────────────────────────────┘   │
│     Using your current server                              │
│     [Edit] if your Pi is on a different network             │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  Test Connectivity                                          │
│                                                             │
│  [ 🔄 Test now ]                                            │
│                                                             │
│  Wait for your Pi to call in (or run manually)…             │
│  ⏱️  Checking… (timeout in 10s)                             │
│                                                             │
│  Once you're ready (after step 5), your Pi's background    │
│  process will automatically prove connectivity.             │
│                                                             │
│  Or manually from the Pi:                                   │
│  ```                                                        │
│  curl -X POST \                                             │
│    https://192.168.1.50:8080/devices/<id>/network-test    │
│  ```                                                        │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  ✓ Network config ready                                    │
│                                                             │
│                                           [ ← Back ] Next → │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Test result examples:**

```
✓ Connected in 120ms                              (green)
⚠️  Slow (3.2s) — may timeout during normal use    (yellow)
✗ No response (10s timeout)                        (red)
⚠️  DNS failed — check your URL                    (red)
```

**Right panel (40%):**
```
┌──────────────────────────────────┐
│  🔗 About the API Server           │
├──────────────────────────────────┤
│                                  │
│  The API is the "brain" of gr33n.│
│  Your Pi talks to it to:          │
│  • Send sensor readings           │
│  • Receive pump/light commands   │
│  • Store offline data if needed  │
│                                  │
│  The address depends on where    │
│  you run gr33n:                   │
│  • Same computer: localhost:8080 │
│  • Local server: 192.168.1.x     │
│  • Cloud: your.domain.com        │
│                                  │
│  [Networking help] →              │
│  [Troubleshooting] →              │
│                                  │
└──────────────────────────────────┘
```

---

### Step 5: Download Config & Instructions

**Left panel (60%):**

```
┌─────────────────────────────────────────────────────────────┐
│                                                 [Step 5/6]   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  📥 Download Config                                         │
│                                                             │
│  Your config.yaml is ready. Copy it to the Pi.              │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  Option A: Download File                                    │
│                                                             │
│  [ ⬇️  Download config.yaml ]                               │
│                                                             │
│  Then copy to Pi:                                           │
│  ```                                                        │
│  scp config.yaml pi@192.168.1.42:~/gr33n/pi_client/       │
│  ```                                                        │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  Option B: Copy to Clipboard                                │
│                                                             │
│  [ 📋 Copy YAML ]                                           │
│                                                             │
│  Then paste into your editor:                               │
│  ```                                                        │
│  ssh pi@192.168.1.42                                        │
│  nano ~/gr33n/pi_client/config.yaml                         │
│  # paste, then Ctrl+X → Y → Enter                           │
│  ```                                                        │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  Preview (read-only):                                       │
│                                                             │
│  [box with syntax-highlighted YAML]                         │
│  api:                                                       │
│    base_url: http://192.168.1.50:8080                       │
│    api_key: gdev_demo-flower-01_xK9mL2p5Q8wNaB1cR3vJ7dT  │
│  device:                                                    │
│    uid: demo-flower-01                                      │
│  farm:                                                      │
│    farm_id: 1                                               │
│  …                                                          │
│                                                             │
│                                           [ ← Back ] Next → │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Right panel (40%):**
```
┌──────────────────────────────────┐
│  ❓ SSH? nano? scp?                │
├──────────────────────────────────┤
│                                  │
│  Don't worry! These are standard │
│  ways to connect to a computer   │
│  over the network.               │
│                                  │
│  • SSH = secure login (like RDP)│
│  • scp = secure copy (like      │
│    drag-and-drop over network)   │
│  • nano = simple text editor     │
│                                  │
│  [Quick SSH guide] →              │
│  [Troubleshooting] →              │
│                                  │
│  📝 The YAML file                 │
│  ───────────────────────────────  │
│  This file tells your Pi:        │
│  • Where the API is              │
│  • What API key to use           │
│  • What farm/device it belongs to│
│  • Polling intervals             │
│                                  │
│  You don't edit it after setup— │
│  updates come from the UI.       │
│                                  │
└──────────────────────────────────┘
```

---

### Step 6: Verify & Complete

**Left panel (60%):**

```
┌─────────────────────────────────────────────────────────────┐
│                                                 [Step 6/6]   │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ✅ Ready to Deploy                                         │
│                                                             │
│  Deployment Checklist                                       │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  ☑ config.yaml copied to the Pi                            │
│    └─ Placed at: ~/gr33n/pi_client/config.yaml             │
│       [Help] [Troubleshoot]                                │
│                                                             │
│  ☑ Pi service restarted                                    │
│    └─ sudo systemctl restart gr33n-pi-client               │
│       [Help] [Troubleshoot]                                │
│                                                             │
│  ☑ Physical wiring complete                                │
│    └─ All relays connected to pumps/lights/fans             │
│       ⚠️  Do NOT leave any relay floating                   │
│       [Safety guide]                                        │
│                                                             │
│  ☑ All actuators assigned in UI                            │
│    └─ Check: Controls → verify 1 pump + 1 light shown      │
│       [Status]                                              │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  Optional: Test One Actuator                                │
│                                                             │
│  [ 🔄 Pulse Main Pump (1 second) ]                          │
│                                                             │
│  This runs the pump for 1 second to prove it works.         │
│  ⚠️  Make sure the reservoir is full & safe to drain!       │
│                                                             │
│  ────────────────────────────────────────────────────────  │
│                                                             │
│  When ready:                                                │
│                                                             │
│  [ 🎉 Finish Setup ]  →  Go to Hardware board / Dashboard  │
│                                                             │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

**Right panel (40%):**
```
┌──────────────────────────────────┐
│  🎯 You're almost there!           │
├──────────────────────────────────┤
│                                  │
│  After you finish:               │
│                                  │
│  1. Your Pi will phone home      │
│     and start posting readings   │
│                                  │
│  2. You'll see live sensor data  │
│     on the Dashboard            │
│                                  │
│  3. You can now create programs │
│     (watering schedules, etc.)   │
│                                  │
│  4. The Pi will execute your     │
│     commands automatically       │
│                                  │
│  💡 Next Steps                    │
│  ───────────────────────────────  │
│  • Add plants / zones            │
│  • Set up watering program       │
│  • Watch Guardian's suggestions  │
│                                  │
│  [New to gr33n? Tour] →           │
│  [Add your first program] →       │
│                                  │
└──────────────────────────────────┘
```

---

## Glossary Component (Always Available)

**Bottom-right corner, collapsible:**

```
┌──────────────────────────────────┐
│  📚 Glossary         [▲ collapse] │
├──────────────────────────────────┤
│                                  │
│  Search: [ I2C____________     ] │
│                                  │
│  Matches:                        │
│                                  │
│  I²C (I-squared-C)               │
│  ───────────────────────────────  │
│  Serial protocol for connecting  │
│  devices. All Sequent relay      │
│  cards use I²C to daisy-chain    │
│  on just 2 GPIO pins.            │
│  → See: pi-sequent-hat-setup.md  │
│                                  │
│  GPIO (General Purpose I/O)      │
│  ───────────────────────────────  │
│  Pins on the Pi that can turn    │
│  on/off. Used for direct relay   │
│  control or sensors.             │
│  → See: pi-integration-guide.md  │
│                                  │
│  [ More terms… ]                  │
│                                  │
└──────────────────────────────────┘
```

---

## Validation States & Messaging

**Real-time banner at top of wizard:**

| State | Icon | Message | Blocks next? |
|-------|------|---------|--------------|
| ✓ Complete | 🟢 | "Device registered" | No |
| ⚠ Warning | 🟡 | "Network test not run yet" | No |
| ✗ Blocker | 🔴 | "All actuators must be assigned" | **YES** |
| — Pending | ⚫ | "Configure network in step 4" | No |

**Example:**
```
STEP 3 VALIDATION
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ Device registered (demo-flower-01)
✓ DIP switches configured (Stack 0, I²C 0x27)
⚠️  4 of 8 relays assigned (Main Pump, Lights, Exhaust, Spare)
✗ Can't proceed — Nutrient Pump A needs an actuator defined
   → [Go create actuator]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
```

---

## Accessibility & Mobile

**Keyboard navigation:**
- Tab through all form fields
- Enter to submit, Escape to cancel
- Arrow keys for dropdowns
- Help text stays on-page (no small modals)

**Mobile (< 768px):**
- Right panel (40%) moves below left panel on-stack
- Full width, single column
- Glossary becomes a collapsible sidebar drawer
- Buttons are larger (48px height min)

---

## Copy / Messaging Tone

**Tone:** Friendly, encouraging, zero IT jargon without losing clarity.

**Examples:**

❌ **Bad:** "Configure I²C address via DIP switch per HAT stack identity requirements."  
✅ **Good:** "Set the 3 tiny switches on your relay card to match the position you'll stack it (position 0, 1, 2, …)."

❌ **Bad:** "API endpoint unreachable; verify firewall rules and DNS resolution."  
✅ **Good:** "Your Pi couldn't reach the API server. Check: Is the Pi on WiFi? Is the URL right? Are they on the same network?"

---

## Testing Checklist

- [ ] Wizard mounts without errors
- [ ] All 6 steps render correctly
- [ ] Back/Next buttons work, validation blocks advance
- [ ] Tooltips appear and close
- [ ] Glossary search works
- [ ] DIP calculator computes correctly
- [ ] Config download produces valid YAML
- [ ] Network test endpoint responds
- [ ] Mobile layout is readable
- [ ] Keyboard nav works (Tab, Enter, Escape)
- [ ] E2E test covers full flow top-to-bottom

---

## Notes for Designers & Developers

1. **Consistency:** Use same component library as rest of app (HelpTip, Button, Modal, etc.)
2. **Accessibility:** All interactive elements need `aria-label` and keyboard support
3. **State management:** Consider Pinia store for wizard state (step, form data, validation)
4. **API integration:** Step 4 network test is async; handle loading + error states
5. **Responsive:** Test at 375px (mobile), 768px (tablet), 1200px (desktop)
6. **Dark mode:** Verify colors work in both light and dark themes
7. **Localization:** Design with i18n in mind (keep copy in a separate file)
